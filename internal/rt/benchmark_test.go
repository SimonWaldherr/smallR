package rt

import "testing"

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
