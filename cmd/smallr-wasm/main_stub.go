//go:build !js || !wasm
// +build !js !wasm

package main

import "fmt"

func main() {
	fmt.Println("smallr-wasm is a WebAssembly build target. Build with: GOOS=js GOARCH=wasm go build ./cmd/smallr-wasm")
}
