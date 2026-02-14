package smallr

import (
	"io"

	"simonwaldherr.de/go/smallr/internal/ast"
	"simonwaldherr.de/go/smallr/internal/lexer"
	"simonwaldherr.de/go/smallr/internal/parser"
	"simonwaldherr.de/go/smallr/internal/rt"
	"simonwaldherr.de/go/smallr/internal/token"
)

// Token ist ein Alias für internal/token.Token
type Token = token.Token

// TokenType ist ein Alias für internal/token.Type
type TokenType = token.Type

// Pos ist ein Alias für internal/token.Pos
type Pos = token.Pos

// Lexer ist ein Alias für internal/lexer.Lexer
type Lexer = lexer.Lexer

// NewLexer erstellt einen neuen Lexer für den gegebenen Quelltext.
// Beispiel:
//
//	l := smallr.NewLexer("1 + 2")
func NewLexer(src string) *Lexer {
	return lexer.New(src)
}

// DebugTokens liest alle Tokens aus dem Quelltext (Hilfsfunktion).
func DebugTokens(src string) ([]Token, error) {
	l := lexer.New(src)
	return l.DebugTokens()
}

// Parser ist ein Alias für internal/parser.Parser
type Parser = parser.Parser

// NewParser erstellt einen Parser aus Quelltext.
func NewParser(src string) *Parser {
	return parser.New(src)
}

// ParseProgram parst ein komplettes Programm und gibt das AST zurück.
func ParseProgram(p *Parser) (*ast.Program, error) {
	return p.ParseProgram()
}

// Program und Expr sind Aliase für die AST-Typen
type Program = ast.Program
type Expr = ast.Expr

// Context ist ein Alias für internal/rt.Context
type Context = rt.Context

// NewContext erstellt einen neuen Auswertungskontext mit Standard-Builtins.
func NewContext() *Context { return rt.NewContext() }

// NewContextWithOutput erstellt einen Kontext mit einem benutzerdefinierten Writer.
func NewContextWithOutput(w io.Writer) *Context { return rt.NewContextWithOutput(w) }

// EvalResult ist ein Alias für internal/rt.EvalResult
type EvalResult = rt.EvalResult

// Env, Value sind Aliase für die entsprechenden Laufzeit-Typen
type Env = rt.Env
type Value = rt.Value

// Eval wertet einen AST-Knoten im gegebenen Kontext und Environment aus.
func Eval(ctx *Context, env *Env, expr Expr) (Value, error) {
	return rt.Eval(ctx, env, expr)
}

// EvalString ist eine praktische Kombination aus Parser + Eval,
// die einen Quelltext direkt auswertet (Ausgabe wird getee'd).
func EvalString(ctx *Context, src string) (EvalResult, error) {
	return ctx.EvalString(src)
}
