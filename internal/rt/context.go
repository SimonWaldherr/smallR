package rt

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"simonwaldherr.de/go/smallr/internal/parser"
)

type Context struct {
	Global *Env
	Output io.Writer
}

func NewContext() *Context {
	ctx := &Context{
		Global: NewEnv(nil),
		Output: os.Stdout,
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

func (ctx *Context) EvalString(src string) (EvalResult, error) {
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
		v, err := Eval(ctx, env, e)
		if err != nil {
			return EvalResult{Value: last, Output: buf.String()}, err
		}
		last = v
	}
	return EvalResult{Value: last, Output: buf.String()}, nil
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
