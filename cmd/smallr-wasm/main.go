//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"

	"simonwaldherr.de/go/smallr/internal/rt"
)

func main() {
	ctx := rt.NewContext()

	js.Global().Set("smallrEval", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return map[string]any{"error": "smallrEval(code) missing code string"}
		}
		code := args[0].String()
		res, err := ctx.EvalString(code)
		if err != nil {
			return map[string]any{
				"error":  err.Error(),
				"output": res.Output,
			}
		}
		return map[string]any{
			"value":  res.Value.String(),
			"json":   rt.ToJSON(res.Value),
			"output": res.Output,
		}
	}))

	// Keep running
	select {}
}
