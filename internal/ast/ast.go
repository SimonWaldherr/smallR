package ast

import (
	"fmt"
	"strings"

	"simonwaldherr.de/go/smallr/internal/token"
)

type Node interface {
	Pos() token.Pos
	String() string
}

type Expr interface {
	Node
	exprNode()
}

type Program struct {
	Exprs []Expr
}

func (p *Program) Pos() token.Pos {
	if len(p.Exprs) == 0 {
		return token.Pos{Line: 1, Col: 1}
	}
	return p.Exprs[0].Pos()
}

func (p *Program) String() string {
	var sb strings.Builder
	for i, e := range p.Exprs {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(e.String())
	}
	return sb.String()
}

// --- Literals ---

type Ident struct {
	P    token.Pos
	Name string
}

func (i *Ident) Pos() token.Pos { return i.P }
func (i *Ident) exprNode()      {}
func (i *Ident) String() string { return i.Name }

type NumberLit struct {
	P     token.Pos
	Text  string
	Value float64
	IsInt bool
}

func (n *NumberLit) Pos() token.Pos { return n.P }
func (n *NumberLit) exprNode()      {}
func (n *NumberLit) String() string { return n.Text }

type StringLit struct {
	P     token.Pos
	Value string
}

func (s *StringLit) Pos() token.Pos { return s.P }
func (s *StringLit) exprNode()      {}
func (s *StringLit) String() string { return fmt.Sprintf("%q", s.Value) }

type BoolLit struct {
	P     token.Pos
	Value bool
}

func (b *BoolLit) Pos() token.Pos { return b.P }
func (b *BoolLit) exprNode()      {}
func (b *BoolLit) String() string {
	if b.Value {
		return "TRUE"
	}
	return "FALSE"
}

type NullLit struct{ P token.Pos }

func (n *NullLit) Pos() token.Pos { return n.P }
func (n *NullLit) exprNode()      {}
func (n *NullLit) String() string { return "NULL" }

type NALit struct{ P token.Pos }

func (n *NALit) Pos() token.Pos { return n.P }
func (n *NALit) exprNode()      {}
func (n *NALit) String() string { return "NA" }

// --- Expressions ---

type UnaryExpr struct {
	P  token.Pos
	Op token.Type
	X  Expr
}

func (u *UnaryExpr) Pos() token.Pos { return u.P }
func (u *UnaryExpr) exprNode()      {}
func (u *UnaryExpr) String() string { return fmt.Sprintf("(%s%s)", u.Op, u.X.String()) }

type BinaryExpr struct {
	P     token.Pos
	Op    token.Type
	Left  Expr
	Right Expr
}

func (b *BinaryExpr) Pos() token.Pos { return b.P }
func (b *BinaryExpr) exprNode()      {}
func (b *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Op, b.Right.String())
}

type AssignExpr struct {
	P     token.Pos
	Op    token.Type // <-, =, <<-
	Left  Expr
	Right Expr
}

func (a *AssignExpr) Pos() token.Pos { return a.P }
func (a *AssignExpr) exprNode()      {}
func (a *AssignExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", a.Left.String(), a.Op, a.Right.String())
}

type BlockExpr struct {
	P     token.Pos
	Exprs []Expr
}

func (b *BlockExpr) Pos() token.Pos { return b.P }
func (b *BlockExpr) exprNode()      {}
func (b *BlockExpr) String() string {
	var parts []string
	for _, e := range b.Exprs {
		parts = append(parts, e.String())
	}
	return "{ " + strings.Join(parts, "; ") + " }"
}

type IfExpr struct {
	P    token.Pos
	Cond Expr
	Then Expr
	Else Expr // may be nil
}

func (i *IfExpr) Pos() token.Pos { return i.P }
func (i *IfExpr) exprNode()      {}
func (i *IfExpr) String() string {
	if i.Else != nil {
		return fmt.Sprintf("(if %s %s else %s)", i.Cond.String(), i.Then.String(), i.Else.String())
	}
	return fmt.Sprintf("(if %s %s)", i.Cond.String(), i.Then.String())
}

type ForExpr struct {
	P    token.Pos
	Var  string
	Seq  Expr
	Body Expr
}

func (f *ForExpr) Pos() token.Pos { return f.P }
func (f *ForExpr) exprNode()      {}
func (f *ForExpr) String() string {
	return fmt.Sprintf("(for %s in %s %s)", f.Var, f.Seq.String(), f.Body.String())
}

type WhileExpr struct {
	P    token.Pos
	Cond Expr
	Body Expr
}

func (w *WhileExpr) Pos() token.Pos { return w.P }
func (w *WhileExpr) exprNode()      {}
func (w *WhileExpr) String() string {
	return fmt.Sprintf("(while %s %s)", w.Cond.String(), w.Body.String())
}

type RepeatExpr struct {
	P    token.Pos
	Body Expr
}

func (r *RepeatExpr) Pos() token.Pos { return r.P }
func (r *RepeatExpr) exprNode()      {}
func (r *RepeatExpr) String() string { return fmt.Sprintf("(repeat %s)", r.Body.String()) }

type BreakExpr struct{ P token.Pos }

func (b *BreakExpr) Pos() token.Pos { return b.P }
func (b *BreakExpr) exprNode()      {}
func (b *BreakExpr) String() string { return "break" }

type NextExpr struct{ P token.Pos }

func (n *NextExpr) Pos() token.Pos { return n.P }
func (n *NextExpr) exprNode()      {}
func (n *NextExpr) String() string { return "next" }

type ReturnExpr struct {
	P token.Pos
	X Expr // may be nil
}

func (r *ReturnExpr) Pos() token.Pos { return r.P }
func (r *ReturnExpr) exprNode()      {}
func (r *ReturnExpr) String() string {
	if r.X == nil {
		return "return()"
	}
	return "return(" + r.X.String() + ")"
}

type Param struct {
	Name    string
	Default Expr // may be nil
	Dots    bool // ...
}

type FuncExpr struct {
	P      token.Pos
	Params []Param
	Body   Expr // usually *BlockExpr
}

func (f *FuncExpr) Pos() token.Pos { return f.P }
func (f *FuncExpr) exprNode()      {}
func (f *FuncExpr) String() string { return "function(...)" }

type Arg struct {
	Name  string // optional
	Value Expr
}

type CallExpr struct {
	P    token.Pos
	Fun  Expr
	Args []Arg
}

func (c *CallExpr) Pos() token.Pos { return c.P }
func (c *CallExpr) exprNode()      {}
func (c *CallExpr) String() string { return c.Fun.String() + "(...)" }

type IndexExpr struct {
	P      token.Pos
	X      Expr
	Index  Expr
	Double bool // [[ ]]
}

func (i *IndexExpr) Pos() token.Pos { return i.P }
func (i *IndexExpr) exprNode()      {}
func (i *IndexExpr) String() string {
	if i.Double {
		return fmt.Sprintf("%s[[%s]]", i.X.String(), i.Index.String())
	}
	return fmt.Sprintf("%s[%s]", i.X.String(), i.Index.String())
}

type DollarExpr struct {
	P    token.Pos
	X    Expr
	Name string
}

func (d *DollarExpr) Pos() token.Pos { return d.P }
func (d *DollarExpr) exprNode()      {}
func (d *DollarExpr) String() string { return fmt.Sprintf("%s$%s", d.X.String(), d.Name) }
