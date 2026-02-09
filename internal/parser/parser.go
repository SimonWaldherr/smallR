package parser

import (
	"fmt"
	"strconv"

	"simonwaldherr.de/go/smallr/internal/ast"
	"simonwaldherr.de/go/smallr/internal/lexer"
	"simonwaldherr.de/go/smallr/internal/token"
)

type Parser struct {
	l    *lexer.Lexer
	cur  token.Token
	peek token.Token

	errors []error
}

func New(src string) *Parser {
	l := lexer.New(src)
	p := &Parser{l: l}
	// prime
	p.cur = l.Next()
	p.peek = l.Next()
	return p
}

func (p *Parser) Errors() []error { return p.errors }

func (p *Parser) next() {
	p.cur = p.peek
	p.peek = p.l.Next()
}

func (p *Parser) curIs(t token.Type) bool  { return p.cur.Type == t }
func (p *Parser) peekIs(t token.Type) bool { return p.peek.Type == t }

func (p *Parser) skipSeparatorsCur() {
	for p.cur.Type == token.NL || p.cur.Type == token.SEMI {
		p.next()
	}
}

func (p *Parser) skipSeparatorsPeek() {
	for p.peek.Type == token.NL || p.peek.Type == token.SEMI {
		p.next()
	}
}

func (p *Parser) expectPeek(t token.Type) bool {
	p.skipSeparatorsPeek()
	if p.peekIs(t) {
		p.next()
		return true
	}
	p.errorf(p.peek.Pos, "expected next token %s, got %s", t, p.peek.Type)
	return false
}

func (p *Parser) errorf(pos token.Pos, format string, args ...any) {
	p.errors = append(p.errors, fmt.Errorf("%s: %s", pos.String(), fmt.Sprintf(format, args...)))
}

func (p *Parser) ParseProgram() (*ast.Program, error) {
	prog := &ast.Program{}
	// consume leading separators
	p.skipSeparatorsCur()
	for !p.curIs(token.EOF) {
		expr := p.parseExpression(precLowest)
		if expr != nil {
			prog.Exprs = append(prog.Exprs, expr)
		}
		// advance to next statement separator or EOF
		for p.cur.Type != token.EOF && p.cur.Type != token.NL && p.cur.Type != token.SEMI {
			p.next()
		}
		p.skipSeparatorsCur()
	}
	if len(p.errors) > 0 {
		return prog, p.errors[0]
	}
	return prog, nil
}

// --- Pratt precedence ---

const (
	precLowest = iota
	precAssign
	precPipe
	precOrOr
	precAndAnd
	precOr
	precAnd
	precCompare
	precAdd
	precMul
	precColon
	precCaret
	precUnary
	precCall
)

func (p *Parser) peekPrecedence() int { return precedence(p.peek.Type) }
func (p *Parser) curPrecedence() int  { return precedence(p.cur.Type) }

func precedence(t token.Type) int {
	switch t {
	case token.ASSIGN_LEFT, token.ASSIGN_EQ, token.ASSIGN_SUPER, token.ASSIGN_RIGHT:
		return precAssign
	case token.PIPE:
		return precPipe
	case token.OROR:
		return precOrOr
	case token.ANDAND:
		return precAndAnd
	case token.OR:
		return precOr
	case token.AND:
		return precAnd
	case token.LT, token.LTE, token.GT, token.GTE, token.EQ, token.NEQ:
		return precCompare
	case token.PLUS, token.MINUS:
		return precAdd
	case token.STAR, token.SLASH, token.MOD, token.INTDIV, token.INOP:
		return precMul
	case token.COLON:
		return precColon
	case token.CARET:
		return precCaret
	case token.LPAREN, token.LBRACK, token.LDBRACK, token.DOLLAR:
		return precCall
	default:
		return precLowest
	}
}

func (p *Parser) parseExpression(precedence int) ast.Expr {
	p.skipSeparatorsCur()

	left := p.parsePrefix()
	if left == nil {
		return nil
	}

	for !p.peekIs(token.EOF) && !p.peekIs(token.NL) && !p.peekIs(token.SEMI) && precedence < p.peekPrecedence() {
		p.next() // advance to infix token
		left = p.parseInfix(left)
		if left == nil {
			return nil
		}
	}
	return left
}

func (p *Parser) parsePrefix() ast.Expr {
	switch p.cur.Type {
	case token.IDENT:
		return &ast.Ident{P: p.cur.Pos, Name: p.cur.Lit}
	case token.NUMBER:
		return p.parseNumber()
	case token.STRING:
		return &ast.StringLit{P: p.cur.Pos, Value: p.cur.Lit}
	case token.TRUE:
		return &ast.BoolLit{P: p.cur.Pos, Value: true}
	case token.FALSE:
		return &ast.BoolLit{P: p.cur.Pos, Value: false}
	case token.NULL:
		return &ast.NullLit{P: p.cur.Pos}
	case token.NA:
		return &ast.NALit{P: p.cur.Pos}
	case token.LPAREN:
		return p.parseGrouped()
	case token.LBRACE:
		return p.parseBlock()
	case token.IF:
		return p.parseIf()
	case token.FOR:
		return p.parseFor()
	case token.WHILE:
		return p.parseWhile()
	case token.REPEAT:
		return p.parseRepeat()
	case token.BREAK:
		return &ast.BreakExpr{P: p.cur.Pos}
	case token.NEXT:
		return &ast.NextExpr{P: p.cur.Pos}
	case token.RETURN:
		return p.parseReturn()
	case token.FUNCTION:
		return p.parseFunction()
	case token.PLUS, token.MINUS, token.BANG:
		op := p.cur.Type
		pos := p.cur.Pos
		p.next()
		p.skipSeparatorsCur()
		x := p.parseExpression(precUnary)
		return &ast.UnaryExpr{P: pos, Op: op, X: x}
	default:
		p.errorf(p.cur.Pos, "unexpected token: %s", p.cur.Type)
		return nil
	}
}

func (p *Parser) parseInfix(left ast.Expr) ast.Expr {
	switch p.cur.Type {
	case token.PLUS, token.MINUS, token.STAR, token.SLASH, token.CARET, token.COLON,
		token.MOD, token.INTDIV, token.INOP,
		token.LT, token.LTE, token.GT, token.GTE, token.EQ, token.NEQ,
		token.AND, token.ANDAND, token.OR, token.OROR:
		op := p.cur.Type
		pos := p.cur.Pos
		prec := p.curPrecedence()
		// right associative for ^ only
		rightPrec := prec
		if op == token.CARET {
			rightPrec = prec - 1
		}
		p.next()
		p.skipSeparatorsCur()
		right := p.parseExpression(rightPrec)
		return &ast.BinaryExpr{P: pos, Op: op, Left: left, Right: right}

	case token.PIPE:
		// |> pipe operator: x |> f becomes f(x), x |> f(y) becomes f(x, y)
		pos := p.cur.Pos
		p.next()
		p.skipSeparatorsCur()
		right := p.parseExpression(precPipe)
		// Transform: if right is a CallExpr, prepend left as first arg
		if call, ok := right.(*ast.CallExpr); ok {
			newArgs := make([]ast.Arg, 0, 1+len(call.Args))
			newArgs = append(newArgs, ast.Arg{Value: left})
			newArgs = append(newArgs, call.Args...)
			return &ast.CallExpr{P: pos, Fun: call.Fun, Args: newArgs}
		}
		// Otherwise, treat right as a function and call it with left
		return &ast.CallExpr{P: pos, Fun: right, Args: []ast.Arg{{Value: left}}}

	case token.ASSIGN_LEFT, token.ASSIGN_EQ, token.ASSIGN_SUPER, token.ASSIGN_RIGHT:
		op := p.cur.Type
		pos := p.cur.Pos
		prec := p.curPrecedence()
		// assignment right associative
		rightPrec := prec - 1
		p.next()
		p.skipSeparatorsCur()
		right := p.parseExpression(rightPrec)
		return &ast.AssignExpr{P: pos, Op: op, Left: left, Right: right}

	case token.LPAREN:
		return p.parseCall(left)
	case token.LBRACK:
		return p.parseIndex(left, false)
	case token.LDBRACK:
		return p.parseIndex(left, true)
	case token.DOLLAR:
		return p.parseDollar(left)
	default:
		p.errorf(p.cur.Pos, "unexpected infix token: %s", p.cur.Type)
		return nil
	}
}

func (p *Parser) parseGrouped() ast.Expr {
	// cur is (
	pos := p.cur.Pos
	p.next()
	p.skipSeparatorsCur()
	exp := p.parseExpression(precLowest)
	if !p.expectPeek(token.RPAREN) {
		return exp
	}
	_ = pos
	// current is )
	return exp
}

func (p *Parser) parseBlock() ast.Expr {
	pos := p.cur.Pos
	var exprs []ast.Expr
	// consume '{'
	p.next()
	p.skipSeparatorsCur()
	for !p.curIs(token.RBRACE) && !p.curIs(token.EOF) {
		e := p.parseExpression(precLowest)
		if e != nil {
			exprs = append(exprs, e)
		}
		// move to separator or }
		for p.cur.Type != token.EOF && p.cur.Type != token.RBRACE && p.cur.Type != token.NL && p.cur.Type != token.SEMI {
			p.next()
		}
		p.skipSeparatorsCur()
	}
	if !p.curIs(token.RBRACE) {
		p.errorf(p.cur.Pos, "expected }")
	}
	// current is '}'
	return &ast.BlockExpr{P: pos, Exprs: exprs}
}

func (p *Parser) parseIf() ast.Expr {
	pos := p.cur.Pos
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	// now cur is '('
	p.next()
	p.skipSeparatorsCur()
	cond := p.parseExpression(precLowest)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	// consume ')'
	p.next()
	p.skipSeparatorsCur()
	thenExpr := p.parseExpression(precLowest)

	var elseExpr ast.Expr
	// allow separators before else
	p.skipSeparatorsCur()
	if p.curIs(token.NL) || p.curIs(token.SEMI) {
		p.skipSeparatorsCur()
	}
	if p.curIs(token.ELSE) || p.peekIs(token.ELSE) {
		// else may be in peek if thenExpr ended without consuming
		if !p.curIs(token.ELSE) {
			p.next()
		}
		// cur is else
		p.next()
		p.skipSeparatorsCur()
		elseExpr = p.parseExpression(precLowest)
	}
	return &ast.IfExpr{P: pos, Cond: cond, Then: thenExpr, Else: elseExpr}
}

func (p *Parser) parseFor() ast.Expr {
	pos := p.cur.Pos
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.next() // at '('
	p.skipSeparatorsCur()
	if p.cur.Type != token.IDENT {
		p.errorf(p.cur.Pos, "expected identifier in for()")
		return nil
	}
	varName := p.cur.Lit
	if !p.expectPeek(token.IN) {
		return nil
	}
	p.next()
	p.skipSeparatorsCur()
	seq := p.parseExpression(precLowest)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	p.next()
	p.skipSeparatorsCur()
	body := p.parseExpression(precLowest)
	return &ast.ForExpr{P: pos, Var: varName, Seq: seq, Body: body}
}

func (p *Parser) parseWhile() ast.Expr {
	pos := p.cur.Pos
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.next()
	p.skipSeparatorsCur()
	cond := p.parseExpression(precLowest)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	p.next()
	p.skipSeparatorsCur()
	body := p.parseExpression(precLowest)
	return &ast.WhileExpr{P: pos, Cond: cond, Body: body}
}

func (p *Parser) parseRepeat() ast.Expr {
	pos := p.cur.Pos
	p.next()
	p.skipSeparatorsCur()
	body := p.parseExpression(precLowest)
	return &ast.RepeatExpr{P: pos, Body: body}
}

func (p *Parser) parseReturn() ast.Expr {
	pos := p.cur.Pos
	// return may have ()
	if p.peekIs(token.LPAREN) {
		p.next() // move to '('
		p.next() // first token inside
		p.skipSeparatorsCur()
		if p.curIs(token.RPAREN) {
			// current is ')'
			return &ast.ReturnExpr{P: pos, X: nil}
		}
		x := p.parseExpression(precLowest)
		if !p.expectPeek(token.RPAREN) {
			return &ast.ReturnExpr{P: pos, X: x}
		}
		// current is ')'
		return &ast.ReturnExpr{P: pos, X: x}
	}
	return &ast.ReturnExpr{P: pos, X: nil}
}

func (p *Parser) parseFunction() ast.Expr {
	pos := p.cur.Pos
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	// cur is '('
	var params []ast.Param
	p.next()
	p.skipSeparatorsCur()
	// parse params until ')'
	for !p.curIs(token.RPAREN) && !p.curIs(token.EOF) {
		if p.cur.Type == token.IDENT && p.cur.Lit == "..." {
			params = append(params, ast.Param{Name: "...", Dots: true})
			p.next()
		} else if p.cur.Type == token.IDENT {
			name := p.cur.Lit
			var def ast.Expr
			// default via '='
			if p.peekIs(token.ASSIGN_EQ) {
				p.next() // cur becomes '='
				p.next() // first token of default
				p.skipSeparatorsCur()
				def = p.parseExpression(precLowest)
				// advance past the end of the default expression so cur points
				// at the separator (comma or ')') for the next loop iteration
				p.next()
			} else {
				p.next()
			}
			params = append(params, ast.Param{Name: name, Default: def})
		} else if p.cur.Type == token.COMMA {
			p.next()
			p.skipSeparatorsCur()
			continue
		} else {
			p.errorf(p.cur.Pos, "unexpected token in parameter list: %s", p.cur.Type)
			return nil
		}
		// after param, expect comma or ')'
		if p.curIs(token.COMMA) {
			p.next()
			p.skipSeparatorsCur()
		} else if p.curIs(token.RPAREN) {
			break
		} else if p.peekIs(token.COMMA) {
			p.next()
			p.next()
		}
	}
	if !p.curIs(token.RPAREN) {
		p.errorf(p.cur.Pos, "expected ) to close parameter list")
		return nil
	}
	// consume ')'
	p.next()
	p.skipSeparatorsCur()
	body := p.parseExpression(precLowest)
	return &ast.FuncExpr{P: pos, Params: params, Body: body}
}

func (p *Parser) parseCall(fun ast.Expr) ast.Expr {
	pos := p.cur.Pos
	// current token is '('

	var args []ast.Arg

	// Move to first token inside the call
	p.next()
	p.skipSeparatorsCur()

	// Empty call: f()
	if p.curIs(token.RPAREN) {
		// current is ')'
		return &ast.CallExpr{P: pos, Fun: fun, Args: args}
	}

	for !p.curIs(token.RPAREN) && !p.curIs(token.EOF) {
		p.skipSeparatorsCur()
		// allow stray commas
		if p.curIs(token.COMMA) {
			p.next()
			p.skipSeparatorsCur()
			continue
		}

		// Named arg: IDENT '=' expr
		if p.cur.Type == token.IDENT && p.peekIs(token.ASSIGN_EQ) {
			name := p.cur.Lit
			p.next() // cur becomes '='
			p.next() // first token of value
			p.skipSeparatorsCur()
			val := p.parseExpression(precLowest)
			args = append(args, ast.Arg{Name: name, Value: val})
		} else {
			val := p.parseExpression(precLowest)
			args = append(args, ast.Arg{Name: "", Value: val})
		}

		// After parsing an argument, we expect either ',' or ')'
		if p.peekIs(token.COMMA) {
			p.next() // cur becomes ','
			p.next() // move to token after comma
			continue
		}
		if p.peekIs(token.RPAREN) {
			p.next() // cur becomes ')'
			break
		}
		if p.peekIs(token.EOF) {
			break
		}

		// Unexpected token; attempt to recover by advancing.
		p.errorf(p.peek.Pos, "expected ',' or ')', got %s", p.peek.Type)
		p.next()
	}

	if !p.curIs(token.RPAREN) {
		// try to sync at ')'
		if p.peekIs(token.RPAREN) {
			p.next()
		} else {
			p.errorf(p.cur.Pos, "expected ) to close call")
		}
	}
	// current is ')'
	return &ast.CallExpr{P: pos, Fun: fun, Args: args}
}

func (p *Parser) parseIndex(x ast.Expr, dbl bool) ast.Expr {
	pos := p.cur.Pos
	// current is '[' or '[['
	// move to first token inside
	p.next()
	p.skipSeparatorsCur()
	var idx ast.Expr
	if dbl {
		idx = p.parseExpression(precLowest)
		if !p.expectPeek(token.RDBRACK) {
			return &ast.IndexExpr{P: pos, X: x, Index: idx, Double: true}
		}
		// current is ']]'
	} else {
		idx = p.parseExpression(precLowest)
		if !p.expectPeek(token.RBRACK) {
			return &ast.IndexExpr{P: pos, X: x, Index: idx, Double: false}
		}
		// current is ']'
	}
	return &ast.IndexExpr{P: pos, X: x, Index: idx, Double: dbl}
}

func (p *Parser) parseDollar(x ast.Expr) ast.Expr {
	pos := p.cur.Pos
	// current is $
	p.next()
	p.skipSeparatorsCur()
	if p.cur.Type != token.IDENT && p.cur.Type != token.STRING {
		p.errorf(p.cur.Pos, "expected name after $")
		return &ast.DollarExpr{P: pos, X: x, Name: ""}
	}
	name := p.cur.Lit
	// current is name
	return &ast.DollarExpr{P: pos, X: x, Name: name}
}

func (p *Parser) parseNumber() ast.Expr {
	pos := p.cur.Pos
	txt := p.cur.Lit
	// heuristics for int vs double
	isInt := true
	for _, c := range txt {
		if c == '.' || c == 'e' || c == 'E' {
			isInt = false
			break
		}
	}
	v, err := strconv.ParseFloat(txt, 64)
	if err != nil {
		p.errorf(pos, "invalid number: %s", txt)
		v = 0
	}
	return &ast.NumberLit{P: pos, Text: txt, Value: v, IsInt: isInt}
}
