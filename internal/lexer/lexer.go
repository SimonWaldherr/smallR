package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"simonwaldherr.de/go/smallr/internal/token"
)

type Lexer struct {
	src   string
	start int
	pos   int
	width int

	line int
	col  int

	parenDepth int
	brackDepth int
}

func New(src string) *Lexer {
	return &Lexer{src: src, line: 1, col: 1}
}

func (l *Lexer) Next() token.Token {
	l.skipWhitespaceAndComments()

	if l.pos >= len(l.src) {
		return token.Token{Type: token.EOF, Pos: l.curPos()}
	}

	ch := l.peek()

	// Newline as statement separator unless inside parens/brackets
	if ch == '\n' {
		l.read()
		if l.parenDepth > 0 || l.brackDepth > 0 {
			// treat as whitespace
			return l.Next()
		}
		return token.Token{Type: token.NL, Lit: "\n", Pos: l.curPosWithOffset(l.pos - 1)}
	}

	// Delimiters affecting depth
	if ch == '(' {
		p := l.curPos()
		l.read()
		l.parenDepth++
		return token.Token{Type: token.LPAREN, Lit: "(", Pos: p}
	}
	if ch == ')' {
		p := l.curPos()
		l.read()
		if l.parenDepth > 0 {
			l.parenDepth--
		}
		return token.Token{Type: token.RPAREN, Lit: ")", Pos: p}
	}
	if ch == '[' {
		p := l.curPos()
		l.read()
		if l.match('[') {
			l.brackDepth++
			return token.Token{Type: token.LDBRACK, Lit: "[[", Pos: p}
		}
		l.brackDepth++
		return token.Token{Type: token.LBRACK, Lit: "[", Pos: p}
	}
	if ch == ']' {
		p := l.curPos()
		l.read()
		if l.match(']') {
			if l.brackDepth > 0 {
				l.brackDepth--
			}
			return token.Token{Type: token.RDBRACK, Lit: "]]", Pos: p}
		}
		if l.brackDepth > 0 {
			l.brackDepth--
		}
		return token.Token{Type: token.RBRACK, Lit: "]", Pos: p}
	}
	if ch == '{' {
		p := l.curPos()
		l.read()
		return token.Token{Type: token.LBRACE, Lit: "{", Pos: p}
	}
	if ch == '}' {
		p := l.curPos()
		l.read()
		return token.Token{Type: token.RBRACE, Lit: "}", Pos: p}
	}
	if ch == ',' {
		p := l.curPos()
		l.read()
		return token.Token{Type: token.COMMA, Lit: ",", Pos: p}
	}
	if ch == ';' {
		p := l.curPos()
		l.read()
		return token.Token{Type: token.SEMI, Lit: ";", Pos: p}
	}
	if ch == '$' {
		p := l.curPos()
		l.read()
		return token.Token{Type: token.DOLLAR, Lit: "$", Pos: p}
	}

	// Strings
	if ch == '"' || ch == '\'' {
		return l.readString()
	}

	// Backtick identifiers
	if ch == '`' {
		return l.readBacktickIdent()
	}

	// Numbers: digit or dot followed by digit
	if isDigit(ch) || (ch == '.' && l.pos+1 < len(l.src) && isDigit(rune(l.src[l.pos+1]))) {
		return l.readNumber()
	}

	// Identifiers (including ...)
	if isIdentStart(ch) {
		return l.readIdent()
	}

	// Operators (longest first)
	p := l.curPos()
	switch ch {
	case '<':
		l.read()
		if l.match('<') {
			if l.match('-') {
				return token.Token{Type: token.ASSIGN_SUPER, Lit: "<<-", Pos: p}
			}
			return token.Token{Type: token.ILLEGAL, Lit: "<<", Pos: p}
		}
		if l.match('=') {
			return token.Token{Type: token.LTE, Lit: "<=", Pos: p}
		}
		if l.match('-') {
			return token.Token{Type: token.ASSIGN_LEFT, Lit: "<-", Pos: p}
		}
		return token.Token{Type: token.LT, Lit: "<", Pos: p}
	case '-':
		l.read()
		if l.match('>') {
			return token.Token{Type: token.ASSIGN_RIGHT, Lit: "->", Pos: p}
		}
		return token.Token{Type: token.MINUS, Lit: "-", Pos: p}
	case '=':
		l.read()
		if l.match('=') {
			return token.Token{Type: token.EQ, Lit: "==", Pos: p}
		}
		return token.Token{Type: token.ASSIGN_EQ, Lit: "=", Pos: p}
	case '!':
		l.read()
		if l.match('=') {
			return token.Token{Type: token.NEQ, Lit: "!=", Pos: p}
		}
		return token.Token{Type: token.BANG, Lit: "!", Pos: p}
	case '>':
		l.read()
		if l.match('=') {
			return token.Token{Type: token.GTE, Lit: ">=", Pos: p}
		}
		return token.Token{Type: token.GT, Lit: ">", Pos: p}
	case '+':
		l.read()
		return token.Token{Type: token.PLUS, Lit: "+", Pos: p}
	case '*':
		l.read()
		return token.Token{Type: token.STAR, Lit: "*", Pos: p}
	case '/':
		l.read()
		return token.Token{Type: token.SLASH, Lit: "/", Pos: p}
	case '^':
		l.read()
		return token.Token{Type: token.CARET, Lit: "^", Pos: p}
	case ':':
		l.read()
		return token.Token{Type: token.COLON, Lit: ":", Pos: p}
	case '&':
		l.read()
		if l.match('&') {
			return token.Token{Type: token.ANDAND, Lit: "&&", Pos: p}
		}
		return token.Token{Type: token.AND, Lit: "&", Pos: p}
	case '|':
		l.read()
		if l.match('|') {
			return token.Token{Type: token.OROR, Lit: "||", Pos: p}
		}
		if l.match('>') {
			return token.Token{Type: token.PIPE, Lit: "|>", Pos: p}
		}
		return token.Token{Type: token.OR, Lit: "|", Pos: p}
	case '%':
		l.read()
		if l.match('%') {
			return token.Token{Type: token.MOD, Lit: "%%", Pos: p}
		}
		if l.match('/') && l.match('%') {
			return token.Token{Type: token.INTDIV, Lit: "%/%", Pos: p}
		}
		// %in%
		if l.pos+3 <= len(l.src) && l.src[l.pos:l.pos+3] == "in%" {
			l.read() // i
			l.read() // n
			l.read() // %
			return token.Token{Type: token.INOP, Lit: "%in%", Pos: p}
		}
		return token.Token{Type: token.ILLEGAL, Lit: "%", Pos: p}
	}

	// Unknown
	l.read()
	return token.Token{Type: token.ILLEGAL, Lit: string(ch), Pos: p}
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		if l.pos >= len(l.src) {
			return
		}
		ch := l.peek()
		// whitespace except newline
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.read()
			continue
		}
		// comments start with #
		if ch == '#' {
			l.read()
			for l.pos < len(l.src) && l.peek() != '\n' {
				l.read()
			}
			// do not consume newline here, Next() handles it
			continue
		}
		break
	}
}

func (l *Lexer) readString() token.Token {
	quote := l.peek()
	p := l.curPos()
	l.read() // consume quote
	var out []rune
	for {
		if l.pos >= len(l.src) {
			return token.Token{Type: token.ILLEGAL, Lit: "unterminated string", Pos: p}
		}
		ch := l.peek()
		if ch == quote {
			lit := string(out)
			l.read()
			return token.Token{Type: token.STRING, Lit: lit, Pos: p}
		}
		if ch == '\\' {
			l.read()
			if l.pos >= len(l.src) {
				return token.Token{Type: token.ILLEGAL, Lit: "unterminated escape", Pos: p}
			}
			esc := l.peek()
			l.read()
			switch esc {
			case 'n':
				out = append(out, '\n')
			case 't':
				out = append(out, '\t')
			case 'r':
				out = append(out, '\r')
			case '\\':
				out = append(out, '\\')
			case '"':
				out = append(out, '"')
			case '\'':
				out = append(out, '\'')
			default:
				out = append(out, esc)
			}
			continue
		}
		out = append(out, ch)
		l.read()
	}
}

func (l *Lexer) readBacktickIdent() token.Token {
	p := l.curPos()
	l.read() // consume `
	var out []rune
	for {
		if l.pos >= len(l.src) {
			return token.Token{Type: token.ILLEGAL, Lit: "unterminated backtick identifier", Pos: p}
		}
		ch := l.peek()
		if ch == '`' {
			l.read()
			return token.Token{Type: token.IDENT, Lit: string(out), Pos: p}
		}
		out = append(out, ch)
		l.read()
	}
}

func (l *Lexer) readNumber() token.Token {
	p := l.curPos()
	start := l.pos

	// leading dot
	if l.peek() == '.' {
		l.read()
	}
	for l.pos < len(l.src) && isDigit(l.peek()) {
		l.read()
	}

	// decimal
	if l.pos < len(l.src) && l.peek() == '.' {
		l.read()
		for l.pos < len(l.src) && isDigit(l.peek()) {
			l.read()
		}
	}

	// exponent
	if l.pos < len(l.src) && (l.peek() == 'e' || l.peek() == 'E') {
		l.read()
		if l.pos < len(l.src) && (l.peek() == '+' || l.peek() == '-') {
			l.read()
		}
		for l.pos < len(l.src) && isDigit(l.peek()) {
			l.read()
		}
	}

	lit := l.src[start:l.pos]
	return token.Token{Type: token.NUMBER, Lit: lit, Pos: p}
}

func (l *Lexer) readIdent() token.Token {
	p := l.curPos()
	start := l.pos

	// special ...
	if l.peek() == '.' && l.pos+2 < len(l.src) && l.src[l.pos:l.pos+3] == "..." {
		l.pos += 3
		l.col += 3
		return token.Token{Type: token.IDENT, Lit: "...", Pos: p}
	}

	l.read() // consume first
	for l.pos < len(l.src) {
		ch := l.peek()
		if isIdentPart(ch) || isDigit(ch) {
			l.read()
			continue
		}
		break
	}
	lit := l.src[start:l.pos]
	tt := token.LookupIdent(lit)
	return token.Token{Type: tt, Lit: lit, Pos: p}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[l.pos:])
	return r
}

func (l *Lexer) read() rune {
	if l.pos >= len(l.src) {
		l.width = 0
		return 0
	}
	r, w := utf8.DecodeRuneInString(l.src[l.pos:])
	l.width = w
	l.pos += w
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *Lexer) match(r rune) bool {
	if l.pos >= len(l.src) {
		return false
	}
	ch := l.peek()
	if ch == r {
		l.read()
		return true
	}
	return false
}

func (l *Lexer) curPos() token.Pos {
	return token.Pos{Offset: l.pos, Line: l.line, Col: l.col}
}

func (l *Lexer) curPosWithOffset(off int) token.Pos {
	// best effort: we don't backtrack line/col precisely here;
	// used only for newline token.
	return token.Pos{Offset: off, Line: l.line, Col: l.col}
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isIdentStart(r rune) bool {
	return r == '.' || r == '_' || unicode.IsLetter(r)
}

func isIdentPart(r rune) bool {
	return r == '.' || r == '_' || unicode.IsLetter(r)
}

func (l *Lexer) DebugTokens() ([]token.Token, error) {
	var toks []token.Token
	for {
		tok := l.Next()
		toks = append(toks, tok)
		if tok.Type == token.ILLEGAL {
			return toks, fmt.Errorf("illegal token at %s: %s", tok.Pos.String(), tok.Lit)
		}
		if tok.Type == token.EOF {
			return toks, nil
		}
	}
}
