package token

import "fmt"

type Type string

const (
	// Special
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"
	NL      Type = "NL" // newline (statement separator)
	SEMI    Type = ";"  // semicolon (statement separator)

	// Literals
	IDENT  Type = "IDENT"
	NUMBER Type = "NUMBER"
	STRING Type = "STRING"

	// Keywords
	IF       Type = "if"
	ELSE     Type = "else"
	FOR      Type = "for"
	IN       Type = "in"
	WHILE    Type = "while"
	REPEAT   Type = "repeat"
	BREAK    Type = "break"
	NEXT     Type = "next"
	FUNCTION Type = "function"
	RETURN   Type = "return"

	TRUE  Type = "TRUE"
	FALSE Type = "FALSE"
	NULL  Type = "NULL"
	NA    Type = "NA"

	// Operators / Delimiters
	ASSIGN_LEFT  Type = "<-"
	ASSIGN_RIGHT Type = "->"
	ASSIGN_EQ    Type = "="
	ASSIGN_SUPER Type = "<<-"

	PLUS   Type = "+"
	MINUS  Type = "-"
	STAR   Type = "*"
	SLASH  Type = "/"
	CARET  Type = "^"
	MOD    Type = "%%"
	INTDIV Type = "%/%"
	INOP   Type = "%in%"
	COLON  Type = ":"
	BANG   Type = "!"
	AND    Type = "&"
	ANDAND Type = "&&"
	OR     Type = "|"
	OROR   Type = "||"
	PIPE   Type = "|>"
	LT     Type = "<"
	LTE    Type = "<="
	GT     Type = ">"
	GTE    Type = ">="
	EQ     Type = "=="
	NEQ    Type = "!="

	DOLLAR Type = "$"

	COMMA   Type = ","
	LPAREN  Type = "("
	RPAREN  Type = ")"
	LBRACE  Type = "{"
	RBRACE  Type = "}"
	LBRACK  Type = "["
	RBRACK  Type = "]"
	LDBRACK Type = "[["
	RDBRACK Type = "]]"
)

type Pos struct {
	Offset int
	Line   int
	Col    int
}

func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

type Token struct {
	Type Type
	Lit  string
	Pos  Pos
}

func (t Token) String() string {
	if t.Lit == "" {
		return string(t.Type)
	}
	return fmt.Sprintf("%s(%q)", t.Type, t.Lit)
}

var keywords = map[string]Type{
	"if":       IF,
	"else":     ELSE,
	"for":      FOR,
	"in":       IN,
	"while":    WHILE,
	"repeat":   REPEAT,
	"break":    BREAK,
	"next":     NEXT,
	"function": FUNCTION,
	"return":   RETURN,
	"TRUE":     TRUE,
	"FALSE":    FALSE,
	"NULL":     NULL,
	"NA":       NA,
}

func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
