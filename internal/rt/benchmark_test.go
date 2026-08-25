package rt

import (
	"testing"

	"simonwaldherr.de/go/smallr/internal/parser"
)

var benchmarkValue Value

func BenchmarkEvalVectorArithmetic(b *testing.B) {

	ctx := NewContext()
	const src = "x <- 1:100000; x + 2"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := ctx.EvalString(src)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkValue = result.Value
	}
}

func BenchmarkEvalIntegerSum(b *testing.B) {

	ctx := NewContext()
	const src = "sum(1:100000)"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := ctx.EvalString(src)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkValue = result.Value
	}
}

func BenchmarkEvalForLoop(b *testing.B) {
	ctx := NewContext()
	const src = `
total <- 0
for (i in 1:10000) { total <- total + i }
total
`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := ctx.EvalString(src)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkValue = result.Value
	}
}

func BenchmarkEvalCompiledForLoop(b *testing.B) {
	ctx := NewContext()
	const src = `
total <- 0
for (i in 1:10000) { total <- total + i }
total
`
	program, err := parser.New(src).ParseProgram()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := ctx.EvalProgram(program)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkValue = result.Value
	}
}
