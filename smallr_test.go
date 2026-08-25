package smallr

import (
	"sync"
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

func TestCompiledProgramCanBeReused(t *testing.T) {
	program, err := Compile("x <- x + 1; x")
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	ctx := NewContext()
	if _, err := EvalString(ctx, "x <- 0"); err != nil {
		t.Fatalf("initialization failed: %v", err)
	}
	for _, want := range []string{"1", "2", "3"} {
		res, err := EvalProgram(ctx, program)
		if err != nil {
			t.Fatalf("EvalProgram error: %v", err)
		}
		if got := res.Value.String(); got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	}
}

func TestProgramCacheCachesAndEvictsPrograms(t *testing.T) {
	cache := NewProgramCache(2)
	first, err := cache.Compile("1 + 2")
	if err != nil {
		t.Fatalf("first Compile error: %v", err)
	}
	again, err := cache.Compile("1 + 2")
	if err != nil {
		t.Fatalf("cached Compile error: %v", err)
	}
	if first != again {
		t.Fatal("expected cached program to be reused")
	}
	if _, err := cache.Compile("2 + 3"); err != nil {
		t.Fatalf("second Compile error: %v", err)
	}
	if _, err := cache.Compile("3 + 4"); err != nil {
		t.Fatalf("third Compile error: %v", err)
	}
	if got := cache.Len(); got != 2 {
		t.Fatalf("expected cache capacity of 2, got %d", got)
	}
}

func TestProgramCacheConcurrentCompile(t *testing.T) {
	cache := NewProgramCache(4)
	const workers = 16
	programs := make(chan *Program, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			program, err := cache.Compile("sum(1:100)")
			if err != nil {
				errs <- err
				return
			}
			programs <- program
		}()
	}
	wg.Wait()
	close(programs)
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent Compile error: %v", err)
	}
	var first *Program
	for program := range programs {
		if first == nil {
			first = program
			continue
		}
		if program != first {
			t.Fatal("expected concurrent callers to receive the cached program")
		}
	}
}
