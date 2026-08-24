package rt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"simonwaldherr.de/go/smallr/internal/parser"
)

type Context struct {
	Global *Env
	Output io.Writer

	mu     sync.Mutex
	limits ExecutionLimits
	budget executionBudget
}

// ExecutionLimits bounds a single evaluation. Zero disables the respective
// limit; keep both limits enabled when evaluating untrusted code.
type ExecutionLimits struct {
	MaxSteps     uint64
	MaxCallDepth uint
	Timeout      time.Duration
}

// DefaultExecutionLimits make the standard Context safe for interactive and
// embedded use while still leaving ample room for ordinary analyses.
var DefaultExecutionLimits = ExecutionLimits{
	MaxSteps: 1_000_000,
	// A separate call-depth cap prevents recursive code from exhausting the Go
	// stack before it reaches the step limit.
	MaxCallDepth: 1_000,
	Timeout:      2 * time.Second,
}

var (
	ErrExecutionStepLimit = errors.New("smallr: execution step limit exceeded")
	ErrExecutionCallDepth = errors.New("smallr: execution call-depth limit exceeded")
	ErrExecutionTimeout   = errors.New("smallr: execution timed out")
)

type executionBudget struct {
	limits    ExecutionLimits
	steps     uint64
	callDepth uint
	deadline  time.Time
}

func NewContext() *Context {
	ctx := &Context{
		Global: NewEnv(nil),
		Output: os.Stdout,
		limits: DefaultExecutionLimits,
	}
	InstallBuiltins(ctx.Global)
	return ctx
}

func NewContextWithOutput(w io.Writer) *Context {
	ctx := NewContext()
	ctx.Output = w
	return ctx
}

type EvalResult struct {
	Value  Value
	Output string
}

// SetExecutionLimits changes the default limits used by Eval and EvalString.
// It serializes with an active evaluation, so a shared Context cannot have its
// policy changed half way through a program.
func (ctx *Context) SetExecutionLimits(limits ExecutionLimits) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.limits = limits
}

// ExecutionLimits returns the Context's default evaluation policy.
func (ctx *Context) ExecutionLimits() ExecutionLimits {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.limits
}

func (ctx *Context) EvalString(src string) (EvalResult, error) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.evalString(src, ctx.limits)
}

// EvalStringWithLimits evaluates one source string under a caller-supplied
// policy without changing the Context's default limits.
func (ctx *Context) EvalStringWithLimits(src string, limits ExecutionLimits) (EvalResult, error) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.evalString(src, limits)
}

func (ctx *Context) evalString(src string, limits ExecutionLimits) (EvalResult, error) {
	ctx.beginExecution(limits)
	defer ctx.endExecution()

	var buf bytes.Buffer
	// tee output: simple approach
	out := ctx.Output
	ctx.Output = &buf
	defer func() { ctx.Output = out }()

	p := parser.New(src)
	prog, err := p.ParseProgram()
	if err != nil {
		return EvalResult{}, err
	}
	env := ctx.Global
	var last Value = NullValue
	for _, e := range prog.Exprs {
		v, err := eval(ctx, env, e)
		if err != nil {
			return EvalResult{Value: last, Output: buf.String()}, err
		}
		last = v
	}
	return EvalResult{Value: last, Output: buf.String()}, nil
}

func (ctx *Context) beginExecution(limits ExecutionLimits) {
	ctx.budget = executionBudget{limits: limits}
	if limits.Timeout > 0 {
		ctx.budget.deadline = time.Now().Add(limits.Timeout)
	}
}

func (ctx *Context) endExecution() {
	ctx.budget = executionBudget{}
}

// consumeExecutionStep is called for every AST expression. Checking the
// timeout periodically keeps the fast path inexpensive while MaxSteps gives
// a deterministic escape hatch even on runtimes without a useful wall clock.
func (ctx *Context) consumeExecutionStep() error {
	ctx.budget.steps++
	if max := ctx.budget.limits.MaxSteps; max > 0 && ctx.budget.steps > max {
		return fmt.Errorf("%w: maximum is %d", ErrExecutionStepLimit, max)
	}
	if !ctx.budget.deadline.IsZero() && ctx.budget.steps%1024 == 0 && time.Now().After(ctx.budget.deadline) {
		return fmt.Errorf("%w: limit is %s", ErrExecutionTimeout, ctx.budget.limits.Timeout)
	}
	return nil
}

func (ctx *Context) enterCall() error {
	if max := ctx.budget.limits.MaxCallDepth; max > 0 && ctx.budget.callDepth >= max {
		return fmt.Errorf("%w: maximum is %d", ErrExecutionCallDepth, max)
	}
	ctx.budget.callDepth++
	return nil
}

func (ctx *Context) leaveCall() {
	ctx.budget.callDepth--
}

func (ctx *Context) SprintValue(v Value) string {
	if v == nil {
		return "<nil>"
	}
	return v.String()
}

func (ctx *Context) Println(v ...any) {
	fmt.Fprintln(ctx.Output, v...)
}
