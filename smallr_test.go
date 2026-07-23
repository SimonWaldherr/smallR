package smallr

import (
	"testing"
)

func TestDebugTokens(t *testing.T) {
	toks, err := DebugTokens("1 + 2")
	if err != nil {
		t.Fatalf("DebugTokens error: %v", err)
	}
	if len(toks) == 0 {
		t.Fatal("expected tokens, got none")
	}
}

func TestEvalStringSimpleNumber(t *testing.T) {
	ctx := NewContext()
	res, err := EvalString(ctx, "1")
	if err != nil {
		t.Fatalf("EvalString error: %v", err)
	}
	if res.Value == nil {
		t.Fatalf("expected non-nil value")
	}
	if res.Value.String() != "1" {
		t.Fatalf("expected "+"1"+", got %s", res.Value.String())
	}
}
