package rt

import "testing"

func TestScalarConstructorsUseIndependentStorage(t *testing.T) {
	integers := []*IntVec{IntScalar(1), IntNA()}
	doubles := []*DoubleVec{DoubleScalar(1.5), DoubleNA()}

	integers[1].Data[0] = IntElem{Val: 9}
	doubles[1].Data[0] = FloatElem{Val: 9.5}

	if got := integers[0].String(); got != "1" {
		t.Fatalf("IntScalar storage was overwritten: got %q", got)
	}
	if got := doubles[0].String(); got != "1.5" {
		t.Fatalf("DoubleScalar storage was overwritten: got %q", got)
	}
}
