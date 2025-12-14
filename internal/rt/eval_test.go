package rt

import (
	"strings"
	"testing"
)

func TestBasicArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 2", "3"},
		{"10 - 3", "7"},
		{"4 * 5", "20"},
		{"20 / 4", "5"},
		{"2 ^ 3", "8"},
	}

	for _, tt := range tests {
		ctx := NewContext()
		res, err := ctx.EvalString(tt.input)
		if err != nil {
			t.Errorf("input %q: unexpected error: %v", tt.input, err)
			continue
		}
		if res.Value.String() != tt.expected {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expected, res.Value.String())
		}
	}
}

func TestComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 < 2", "TRUE"},
		{"2 > 1", "TRUE"},
		{"1 == 1", "TRUE"},
		{"1 != 2", "TRUE"},
		{"1 == 2", "FALSE"},
	}

	for _, tt := range tests {
		ctx := NewContext()
		res, err := ctx.EvalString(tt.input)
		if err != nil {
			t.Errorf("input %q: unexpected error: %v", tt.input, err)
			continue
		}
		if res.Value.String() != tt.expected {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expected, res.Value.String())
		}
	}
}

func TestAssignment(t *testing.T) {
	tests := []struct {
		code     string
		varName  string
		expected string
	}{
		{"x <- 10", "x", "10"},
		{"y = 20", "y", "20"},
	}

	for _, tt := range tests {
		ctx := NewContext()
		_, err := ctx.EvalString(tt.code)
		if err != nil {
			t.Errorf("code %q: unexpected error: %v", tt.code, err)
			continue
		}
		val, ok := ctx.Global.Get(tt.varName)
		if !ok {
			t.Errorf("code %q: variable %q not found", tt.code, tt.varName)
			continue
		}
		if val.String() != tt.expected {
			t.Errorf("code %q: expected %s, got %s", tt.code, tt.expected, val.String())
		}
	}
}

func TestVectorCreation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"c(1, 2, 3)", "1 2 3"},
		{"1:5", "1 2 3 4 5"},
	}

	for _, tt := range tests {
		ctx := NewContext()
		res, err := ctx.EvalString(tt.input)
		if err != nil {
			t.Errorf("input %q: unexpected error: %v", tt.input, err)
			continue
		}
		if res.Value.String() != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, res.Value.String())
		}
	}
}

func TestFunction(t *testing.T) {
	code := `
		f <- function(x) { x * 2 }
		f(5)
	`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "10" {
		t.Errorf("expected 10, got %s", res.Value.String())
	}
}

func TestIfElse(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"if (TRUE) 1 else 2", "1"},
		{"if (FALSE) 1 else 2", "2"},
	}

	for _, tt := range tests {
		ctx := NewContext()
		res, err := ctx.EvalString(tt.code)
		if err != nil {
			t.Errorf("code %q: unexpected error: %v", tt.code, err)
			continue
		}
		if res.Value.String() != tt.expected {
			t.Errorf("code %q: expected %s, got %s", tt.code, tt.expected, res.Value.String())
		}
	}
}

func TestBuiltinSum(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sum(1, 2, 3)", "6"},
		{"sum(c(10, 20, 30))", "60"},
	}

	for _, tt := range tests {
		ctx := NewContext()
		res, err := ctx.EvalString(tt.input)
		if err != nil {
			t.Errorf("input %q: unexpected error: %v", tt.input, err)
			continue
		}
		if res.Value.String() != tt.expected {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expected, res.Value.String())
		}
	}
}

func TestBuiltinMean(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"mean(c(1, 2, 3))", "2"},
		{"mean(c(10, 20, 30))", "20"},
	}

	for _, tt := range tests {
		ctx := NewContext()
		res, err := ctx.EvalString(tt.input)
		if err != nil {
			t.Errorf("input %q: unexpected error: %v", tt.input, err)
			continue
		}
		if res.Value.String() != tt.expected {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expected, res.Value.String())
		}
	}
}

func TestBuiltinLength(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"length(c(1, 2, 3))", "3"},
		{"length(1:10)", "10"},
	}

	for _, tt := range tests {
		ctx := NewContext()
		res, err := ctx.EvalString(tt.input)
		if err != nil {
			t.Errorf("input %q: unexpected error: %v", tt.input, err)
			continue
		}
		if res.Value.String() != tt.expected {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expected, res.Value.String())
		}
	}
}

func TestBuiltinIsNA(t *testing.T) {
	code := "is.na(c(1, NA, 3))"
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "FALSE TRUE FALSE" {
		t.Errorf("expected 'FALSE TRUE FALSE', got %s", res.Value.String())
	}
}

func TestPrint(t *testing.T) {
	code := `print("hello")`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "hello") {
		t.Errorf("expected output to contain 'hello', got %q", res.Output)
	}
}

func TestNAHandling(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"is.na(NA)", "TRUE"},
		{"is.na(1)", "FALSE"},
	}

	for _, tt := range tests {
		ctx := NewContext()
		res, err := ctx.EvalString(tt.code)
		if err != nil {
			t.Errorf("code %q: unexpected error: %v", tt.code, err)
			continue
		}
		if res.Value.String() != tt.expected {
			t.Errorf("code %q: expected %s, got %s", tt.code, tt.expected, res.Value.String())
		}
	}
}
