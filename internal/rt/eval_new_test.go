package rt

import (
	"strings"
	"testing"
)

// --- Math builtins ---

func TestMathFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abs(-5)", "5"},
		{"abs(3)", "3"},
		{"sqrt(9)", "3"},
		{"sqrt(16)", "4"},
		{"floor(3.7)", "3"},
		{"ceiling(3.2)", "4"},
		{"round(3.456, 2)", "3.46"},
		{"round(3.5)", "4"},
		{"trunc(3.9)", "3"},
		{"trunc(-3.9)", "-3"},
		{"exp(0)", "1"},
		{"log(1)", "0"},
		{"log2(8)", "3"},
		{"log10(100)", "2"},
		{"sign(-5)", "-1"},
		{"sign(0)", "0"},
		{"sign(3)", "1"},
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

func TestMinMax(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"max(1, 5, 3)", "5"},
		{"min(1, 5, 3)", "1"},
		{"max(c(10, 20, 30))", "30"},
		{"min(c(10, 20, 30))", "10"},
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

func TestCumulative(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"cumsum(c(1, 2, 3, 4))", "1 3 6 10"},
		{"cumprod(c(1, 2, 3, 4))", "1 2 6 24"},
		{"cummax(c(1, 3, 2, 5))", "1 3 3 5"},
		{"cummin(c(5, 3, 4, 1))", "5 3 3 1"},
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

func TestProd(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("prod(c(2, 3, 4))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "24" {
		t.Errorf("expected 24, got %s", res.Value.String())
	}
}

func TestDiff(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("diff(c(1, 3, 6, 10))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "2 3 4" {
		t.Errorf("expected '2 3 4', got %s", res.Value.String())
	}
}

func TestRange(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("range(c(3, 1, 5, 2))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "1 5" {
		t.Errorf("expected '1 5', got %s", res.Value.String())
	}
}

// --- String builtins ---

func TestPastePaste0(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`paste("hello", "world")`, `"hello world"`},
		{`paste0("hello", "world")`, `"helloworld"`},
		{`paste("a", "b", sep="-")`, `"a-b"`},
		{`paste(c("x", "y"), collapse=",")`, `"x,y"`},
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

func TestNchar(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString(`nchar("hello")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "5" {
		t.Errorf("expected 5, got %s", res.Value.String())
	}
}

func TestSubstr(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString(`substr("hello world", 1, 5)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != `"hello"` {
		t.Errorf("expected '\"hello\"', got %s", res.Value.String())
	}
}

func TestToupperTolower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`toupper("hello")`, `"HELLO"`},
		{`tolower("WORLD")`, `"world"`},
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

func TestTrimws(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString(`trimws("  hello  ")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != `"hello"` {
		t.Errorf("expected '\"hello\"', got %s", res.Value.String())
	}
}

func TestStartsEndsWith(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`startsWith("hello", "hel")`, "TRUE"},
		{`startsWith("hello", "world")`, "FALSE"},
		{`endsWith("hello", "llo")`, "TRUE"},
		{`endsWith("hello", "world")`, "FALSE"},
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

func TestGrepl(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString(`grepl("lo", c("hello", "world", "below"))`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "TRUE FALSE TRUE" {
		t.Errorf("expected 'TRUE FALSE TRUE', got %s", res.Value.String())
	}
}

func TestGsub(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString(`gsub("o", "0", "hello world")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != `"hell0 w0rld"` {
		t.Errorf("expected '\"hell0 w0rld\"', got %s", res.Value.String())
	}
}

func TestStrsplit(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString(`strsplit("a,b,c", ",")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Result should be a list containing one character vector
	lv, ok := res.Value.(*ListVec)
	if !ok {
		t.Fatalf("expected list, got %T", res.Value)
	}
	if lv.Len() != 1 {
		t.Fatalf("expected list of length 1, got %d", lv.Len())
	}
	inner, ok := lv.Data[0].(*CharVec)
	if !ok {
		t.Fatalf("expected CharVec inside list, got %T", lv.Data[0])
	}
	if inner.Len() != 3 {
		t.Errorf("expected 3 elements, got %d", inner.Len())
	}
}

func TestSprintf(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString(`sprintf("Hello %s, you are %d", "world", 42)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != `"Hello world, you are 42"` {
		t.Errorf("expected '\"Hello world, you are 42\"', got %s", res.Value.String())
	}
}

func TestStrrep(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString(`strrep("ab", 3)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != `"ababab"` {
		t.Errorf("expected '\"ababab\"', got %s", res.Value.String())
	}
}

// --- Vector utilities ---

func TestWhich(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("which(c(FALSE, TRUE, FALSE, TRUE))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "2 4" {
		t.Errorf("expected '2 4', got %s", res.Value.String())
	}
}

func TestWhichMinMax(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"which.min(c(3, 1, 2))", "2"},
		{"which.max(c(3, 1, 2))", "1"},
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

func TestAnyAll(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"any(c(FALSE, TRUE, FALSE))", "TRUE"},
		{"any(c(FALSE, FALSE, FALSE))", "FALSE"},
		{"all(c(TRUE, TRUE, TRUE))", "TRUE"},
		{"all(c(TRUE, FALSE, TRUE))", "FALSE"},
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

func TestRev(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("rev(c(1, 2, 3))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "3 2 1" {
		t.Errorf("expected '3 2 1', got %s", res.Value.String())
	}
}

func TestSort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sort(c(3, 1, 2))", "1 2 3"},
		{"sort(c(3, 1, 2), decreasing=TRUE)", "3 2 1"},
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

func TestUnique(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("unique(c(1, 2, 2, 3, 1))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "1 2 3" {
		t.Errorf("expected '1 2 3', got %s", res.Value.String())
	}
}

func TestDuplicated(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("duplicated(c(1, 2, 2, 3, 1))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "FALSE FALSE TRUE FALSE TRUE" {
		t.Errorf("expected 'FALSE FALSE TRUE FALSE TRUE', got %s", res.Value.String())
	}
}

func TestTable(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString(`table(c("a", "b", "a", "c", "b", "a"))`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should produce named int vector with counts
	iv, ok := res.Value.(*IntVec)
	if !ok {
		t.Fatalf("expected IntVec, got %T", res.Value)
	}
	if iv.Len() != 3 {
		t.Errorf("expected 3 entries, got %d", iv.Len())
	}
}

func TestMatch(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString(`match(c("b", "d", "a"), c("a", "b", "c"))`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "2 NA 1" {
		t.Errorf("expected '2 NA 1', got %s", res.Value.String())
	}
}

func TestSeqLenSeqAlong(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"seq_len(5)", "1 2 3 4 5"},
		{"seq_along(c(10, 20, 30))", "1 2 3"},
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

func TestSetOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"union(c(1, 2, 3), c(3, 4, 5))", "1 2 3 4 5"},
		{"intersect(c(1, 2, 3), c(2, 3, 4))", "2 3"},
		{"setdiff(c(1, 2, 3), c(2, 3, 4))", "1"},
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

// --- Type checking ---

func TestTypeChecking(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"is.numeric(1)", "TRUE"},
		{"is.numeric(1:5)", "TRUE"},
		{`is.numeric("a")`, "FALSE"},
		{`is.character("hello")`, "TRUE"},
		{"is.character(1)", "FALSE"},
		{"is.logical(TRUE)", "TRUE"},
		{"is.logical(1)", "FALSE"},
		{"is.null(NULL)", "TRUE"},
		{"is.null(1)", "FALSE"},
		{"is.list(list(1, 2))", "TRUE"},
		{"is.list(c(1, 2))", "FALSE"},
		{"is.function(print)", "TRUE"},
		{"is.function(1)", "FALSE"},
		{"is.vector(c(1, 2, 3))", "TRUE"},
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

func TestIsFiniteNaN(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"is.finite(1)", "TRUE"},
		{"is.finite(Inf)", "FALSE"},
		{"is.nan(NaN)", "TRUE"},
		{"is.nan(1)", "FALSE"},
		{"is.infinite(Inf)", "TRUE"},
		{"is.infinite(1)", "FALSE"},
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

func TestIdentical(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"identical(1, 1)", "TRUE"},
		{"identical(1, 2)", "FALSE"},
		{`identical("a", "a")`, "TRUE"},
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

// --- Apply family ---

func TestLapply(t *testing.T) {
	code := `
		x <- list(1, 4, 9)
		lapply(x, sqrt)
	`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lv, ok := res.Value.(*ListVec)
	if !ok {
		t.Fatalf("expected ListVec, got %T", res.Value)
	}
	if lv.Len() != 3 {
		t.Errorf("expected 3 elements, got %d", lv.Len())
	}
}

func TestSapply(t *testing.T) {
	code := `sapply(c(1, 4, 9), sqrt)`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "1 2 3" {
		t.Errorf("expected '1 2 3', got %s", res.Value.String())
	}
}

func TestReduce(t *testing.T) {
	code := `Reduce(function(a, b) a + b, c(1, 2, 3, 4), 0)`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "10" {
		t.Errorf("expected '10', got %s", res.Value.String())
	}
}

func TestFilter(t *testing.T) {
	code := `Filter(function(x) x > 2, c(1, 2, 3, 4, 5))`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lv, ok := res.Value.(*ListVec)
	if !ok {
		t.Fatalf("expected ListVec, got %T", res.Value)
	}
	if lv.Len() != 3 {
		t.Errorf("expected 3 elements, got %d", lv.Len())
	}
}

func TestDoCall(t *testing.T) {
	code := `do.call(paste, list("a", "b", "c"))`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != `"a b c"` {
		t.Errorf("expected '\"a b c\"', got %s", res.Value.String())
	}
}

// --- Control / error handling ---

func TestIfelse(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("ifelse(c(TRUE, FALSE, TRUE), c(1, 2, 3), c(10, 20, 30))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "1 20 3" {
		t.Errorf("expected '1 20 3', got %s", res.Value.String())
	}
}

func TestSwitch(t *testing.T) {
	code := `switch("b", a=1, b=2, c=3)`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "2" {
		t.Errorf("expected 2, got %s", res.Value.String())
	}
}

func TestTryCatch(t *testing.T) {
	code := `tryCatch(stop("oops"), error=function(e) paste("caught:", e))`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Value.String(), "caught:") {
		t.Errorf("expected caught message, got %s", res.Value.String())
	}
}

// --- Constants ---

func TestConstants(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"T", "TRUE"},
		{"F", "FALSE"},
		{"is.infinite(Inf)", "TRUE"},
		{"is.nan(NaN)", "TRUE"},
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

func TestPi(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("round(pi, 4)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "3.1416" {
		t.Errorf("expected 3.1416, got %s", res.Value.String())
	}
}

func TestLetters(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("length(letters)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "26" {
		t.Errorf("expected 26, got %s", res.Value.String())
	}
}

func TestLETTERS(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("LETTERS[1]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != `"A"` {
		t.Errorf("expected '\"A\"', got %s", res.Value.String())
	}
}

// --- %in% operator ---

func TestInOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`2 %in% c(1, 2, 3)`, "TRUE"},
		{`4 %in% c(1, 2, 3)`, "FALSE"},
		{`c(1, 4, 2) %in% c(1, 2, 3)`, "TRUE FALSE TRUE"},
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

// --- Pipe operator ---

func TestPipeOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"c(3, 1, 2) |> sort()", "1 2 3"},
		{"c(1, 2, 3) |> sum()", "6"},
		{"c(1, 2, 3) |> length()", "3"},
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

// --- Environment ---

func TestExists(t *testing.T) {
	code := `
		x <- 5
		exists("x")
	`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "TRUE" {
		t.Errorf("expected TRUE, got %s", res.Value.String())
	}
}

func TestVectorizedMath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abs(c(-1, -2, 3))", "1 2 3"},
		{"sqrt(c(1, 4, 9))", "1 2 3"},
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

// --- Order ---

func TestOrder(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("order(c(30, 10, 20))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "2 3 1" {
		t.Errorf("expected '2 3 1', got %s", res.Value.String())
	}
}

// --- is.data.frame ---

func TestIsDataFrame(t *testing.T) {
	code := `
		df <- data.frame(x = 1:3, y = c(4, 5, 6))
		is.data.frame(df)
	`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "TRUE" {
		t.Errorf("expected TRUE, got %s", res.Value.String())
	}
}

// --- Trig functions ---

func TestTrigFunctions(t *testing.T) {
	code := `round(sin(pi / 2), 1)`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "1" {
		t.Errorf("expected 1, got %s", res.Value.String())
	}
}

// --- Append ---

func TestAppend(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("append(c(1, 2, 3), c(4, 5))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "1 2 3 4 5" {
		t.Errorf("expected '1 2 3 4 5', got %s", res.Value.String())
	}
}

// --- Message ---

func TestMessage(t *testing.T) {
	code := `message("test message")`
	ctx := NewContext()
	res, err := ctx.EvalString(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "test message") {
		t.Errorf("expected output to contain 'test message', got %q", res.Output)
	}
}

// --- Tabulate ---

func TestTabulate(t *testing.T) {
	ctx := NewContext()
	res, err := ctx.EvalString("tabulate(c(2, 3, 3, 5))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Value.String() != "0 1 2 0 1" {
		t.Errorf("expected '0 1 2 0 1', got %s", res.Value.String())
	}
}
