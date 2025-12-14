package lexer

import (
	"testing"

	"simonwaldherr.de/go/smallr/internal/token"
)

func TestBasicTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Type
	}{
		{
			input:    "1 + 2",
			expected: []token.Type{token.NUMBER, token.PLUS, token.NUMBER, token.EOF},
		},
		{
			input:    "x <- 10",
			expected: []token.Type{token.IDENT, token.ASSIGN_LEFT, token.NUMBER, token.EOF},
		},
		{
			input:    "function(a, b) { a + b }",
			expected: []token.Type{token.FUNCTION, token.LPAREN, token.IDENT, token.COMMA, token.IDENT, token.RPAREN, token.LBRACE, token.IDENT, token.PLUS, token.IDENT, token.RBRACE, token.EOF},
		},
	}

	for _, tt := range tests {
		l := New(tt.input)
		for i, expectedType := range tt.expected {
			tok := l.Next()
			if tok.Type != expectedType {
				t.Errorf("test %q token %d: expected %s, got %s", tt.input, i, expectedType, tok.Type)
			}
		}
	}
}

func TestNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"3.14", "3.14"},
		{".5", ".5"},
		{"1e10", "1e10"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.Next()
		if tok.Type != token.NUMBER {
			t.Errorf("expected NUMBER, got %s", tok.Type)
		}
		if tok.Lit != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tok.Lit)
		}
	}
}

func TestStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'world'`, "world"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.Next()
		if tok.Type != token.STRING {
			t.Errorf("expected STRING, got %s", tok.Type)
		}
		if tok.Lit != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tok.Lit)
		}
	}
}

func TestOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected token.Type
	}{
		{"<-", token.ASSIGN_LEFT},
		{"=", token.ASSIGN_EQ},
		{"==", token.EQ},
		{"!=", token.NEQ},
		{"<", token.LT},
		{"<=", token.LTE},
		{">", token.GT},
		{">=", token.GTE},
		{"%%", token.MOD},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.Next()
		if tok.Type != tt.expected {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expected, tok.Type)
		}
	}
}

func TestKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected token.Type
	}{
		{"if", token.IF},
		{"else", token.ELSE},
		{"for", token.FOR},
		{"function", token.FUNCTION},
		{"TRUE", token.TRUE},
		{"FALSE", token.FALSE},
		{"NULL", token.NULL},
		{"NA", token.NA},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.Next()
		if tok.Type != tt.expected {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expected, tok.Type)
		}
	}
}
