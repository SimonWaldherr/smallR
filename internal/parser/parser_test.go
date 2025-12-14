package parser

import (
	"testing"

	"simonwaldherr.de/go/smallr/internal/ast"
)

func TestParseNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"42", 42},
		{"3.14", 3.14},
	}

	for _, tt := range tests {
		p := New(tt.input)
		prog, err := p.ParseProgram()
		if err != nil {
			t.Fatalf("ParseProgram() error: %v", err)
		}
		if len(prog.Exprs) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(prog.Exprs))
		}
		lit, ok := prog.Exprs[0].(*ast.NumberLit)
		if !ok {
			t.Fatalf("expected NumberLit, got %T", prog.Exprs[0])
		}
		if lit.Value != tt.expected {
			t.Errorf("expected %f, got %f", tt.expected, lit.Value)
		}
	}
}

func TestParseBinaryExpr(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"1 + 2", true},
		{"x * y", true},
		{"1:10", true},
	}

	for _, tt := range tests {
		p := New(tt.input)
		prog, err := p.ParseProgram()
		if (err == nil) != tt.valid {
			t.Errorf("input %q: expected valid=%v, got error=%v", tt.input, tt.valid, err)
		}
		if tt.valid && len(prog.Exprs) != 1 {
			t.Errorf("input %q: expected 1 expression, got %d", tt.input, len(prog.Exprs))
		}
	}
}

func TestParseAssignment(t *testing.T) {
	tests := []string{
		"x <- 10",
		"y = 20",
	}

	for _, input := range tests {
		p := New(input)
		prog, err := p.ParseProgram()
		if err != nil {
			t.Fatalf("input %q: ParseProgram() error: %v", input, err)
		}
		if len(prog.Exprs) != 1 {
			t.Fatalf("input %q: expected 1 expression, got %d", input, len(prog.Exprs))
		}
		_, ok := prog.Exprs[0].(*ast.AssignExpr)
		if !ok {
			t.Errorf("input %q: expected AssignExpr, got %T", input, prog.Exprs[0])
		}
	}
}

func TestParseFunction(t *testing.T) {
	tests := []struct {
		input      string
		paramCount int
	}{
		{"function() { 1 }", 0},
		{"function(x) { x + 1 }", 1},
		{"function(a, b) { a + b }", 2},
	}

	for _, tt := range tests {
		p := New(tt.input)
		prog, err := p.ParseProgram()
		if err != nil {
			t.Fatalf("input %q: ParseProgram() error: %v", tt.input, err)
		}
		if len(prog.Exprs) != 1 {
			t.Fatalf("input %q: expected 1 expression, got %d", tt.input, len(prog.Exprs))
		}
		fn, ok := prog.Exprs[0].(*ast.FuncExpr)
		if !ok {
			t.Fatalf("input %q: expected FuncExpr, got %T", tt.input, prog.Exprs[0])
		}
		if len(fn.Params) != tt.paramCount {
			t.Errorf("input %q: expected %d params, got %d", tt.input, tt.paramCount, len(fn.Params))
		}
	}
}

func TestParseCall(t *testing.T) {
	tests := []struct {
		input    string
		argCount int
	}{
		{"f()", 0},
		{"f(1)", 1},
		{"f(1, 2)", 2},
	}

	for _, tt := range tests {
		p := New(tt.input)
		prog, err := p.ParseProgram()
		if err != nil {
			t.Fatalf("input %q: ParseProgram() error: %v", tt.input, err)
		}
		if len(prog.Exprs) != 1 {
			t.Fatalf("input %q: expected 1 expression, got %d", tt.input, len(prog.Exprs))
		}
		call, ok := prog.Exprs[0].(*ast.CallExpr)
		if !ok {
			t.Fatalf("input %q: expected CallExpr, got %T", tt.input, prog.Exprs[0])
		}
		if len(call.Args) != tt.argCount {
			t.Errorf("input %q: expected %d args, got %d", tt.input, tt.argCount, len(call.Args))
		}
	}
}

func TestParseIf(t *testing.T) {
	tests := []struct {
		input   string
		hasElse bool
	}{
		{"if (TRUE) 1", false},
		{"if (x > 0) 1 else 2", true},
	}

	for _, tt := range tests {
		p := New(tt.input)
		prog, err := p.ParseProgram()
		if err != nil {
			t.Fatalf("input %q: ParseProgram() error: %v", tt.input, err)
		}
		if len(prog.Exprs) != 1 {
			t.Fatalf("input %q: expected 1 expression, got %d", tt.input, len(prog.Exprs))
		}
		ifExpr, ok := prog.Exprs[0].(*ast.IfExpr)
		if !ok {
			t.Fatalf("input %q: expected IfExpr, got %T", tt.input, prog.Exprs[0])
		}
		if (ifExpr.Else != nil) != tt.hasElse {
			t.Errorf("input %q: expected hasElse=%v, got %v", tt.input, tt.hasElse, ifExpr.Else != nil)
		}
	}
}
