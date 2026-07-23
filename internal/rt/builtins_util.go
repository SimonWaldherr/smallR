package rt

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

func installUtilBuiltins(env *Env) {
	builtins := map[string]*BuiltinFunc{
		// Vector utilities
		"which":      {FnName: "which", Impl: builtinWhich},
		"which.min":  {FnName: "which.min", Impl: builtinWhichMin},
		"which.max":  {FnName: "which.max", Impl: builtinWhichMax},
		"any":        {FnName: "any", Impl: builtinAny},
		"all":        {FnName: "all", Impl: builtinAll},
		"rev":        {FnName: "rev", Impl: builtinRev},
		"sort":       {FnName: "sort", Impl: builtinSort},
		"order":      {FnName: "order", Impl: builtinOrder},
		"unique":     {FnName: "unique", Impl: builtinUnique},
		"duplicated": {FnName: "duplicated", Impl: builtinDuplicated},
		"table":      {FnName: "table", Impl: builtinTable},
		"match":      {FnName: "match", Impl: builtinMatch},
		"append":     {FnName: "append", Impl: builtinAppend},
		"seq_len":    {FnName: "seq_len", Impl: builtinSeqLen},
		"seq_along":  {FnName: "seq_along", Impl: builtinSeqAlong},
		"setdiff":    {FnName: "setdiff", Impl: builtinSetdiff},
		"intersect":  {FnName: "intersect", Impl: builtinIntersect},
		"union":      {FnName: "union", Impl: builtinUnion},

		// Type checking
		"is.numeric":    {FnName: "is.numeric", Impl: builtinIsNumeric},
		"is.integer":    {FnName: "is.integer", Impl: builtinIsInteger},
		"is.double":     {FnName: "is.double", Impl: builtinIsDouble},
		"is.character":  {FnName: "is.character", Impl: builtinIsCharacter},
		"is.logical":    {FnName: "is.logical", Impl: builtinIsLogical},
		"is.null":       {FnName: "is.null", Impl: builtinIsNull},
		"is.list":       {FnName: "is.list", Impl: builtinIsList},
		"is.vector":     {FnName: "is.vector", Impl: builtinIsVector},
		"is.function":   {FnName: "is.function", Impl: builtinIsFunction},
		"is.finite":     {FnName: "is.finite", Impl: builtinIsFinite},
		"is.nan":        {FnName: "is.nan", Impl: builtinIsNaN},
		"is.infinite":   {FnName: "is.infinite", Impl: builtinIsInfinite},
		"is.data.frame": {FnName: "is.data.frame", Impl: builtinIsDataFrame},
		"identical":     {FnName: "identical", Impl: builtinIdentical},

		// Apply family
		"sapply":  {FnName: "sapply", Impl: builtinSapply},
		"lapply":  {FnName: "lapply", Impl: builtinLapply},
		"vapply":  {FnName: "vapply", Impl: builtinVapply},
		"Map":     {FnName: "Map", Impl: builtinMapFunc},
		"Reduce":  {FnName: "Reduce", Impl: builtinReduce},
		"Filter":  {FnName: "Filter", Impl: builtinFilter},
		"do.call": {FnName: "do.call", Impl: builtinDoCall},

		// Control / error handling
		"ifelse":   {FnName: "ifelse", Impl: builtinIfelse},
		"switch":   {FnName: "switch", Impl: builtinSwitch},
		"tryCatch": {FnName: "tryCatch", Impl: builtinTryCatch},
		"message":  {FnName: "message", Impl: builtinMessage},
		"nargs":    {FnName: "nargs", Impl: builtinNargs},

		// Environment
		"exists":      {FnName: "exists", Impl: builtinExists},
		"environment": {FnName: "environment", Impl: builtinEnvironment},
		"Sys.time":    {FnName: "Sys.time", Impl: builtinSysTime},

		// Numeric utilities
		"which.na": {FnName: "which.na", Impl: builtinWhichNA},
		"tabulate": {FnName: "tabulate", Impl: builtinTabulate},
	}
	for name, fn := range builtins {
		env.SetLocal(name, fn)
	}
}

// --- Vector utilities ---

func builtinWhich(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("which(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	lv, err := asLogicalVec(ctx, v)
	if err != nil {
		return nil, err
	}
	var out []IntElem
	for i, e := range lv {
		if !e.NA && e.Val {
			out = append(out, IntElem{Val: int64(i + 1)})
		}
	}
	if out == nil {
		out = []IntElem{}
	}
	return &IntVec{Data: out}, nil
}

// whichExtreme implements which.min/which.max: the 1-based index of the
// first element for which better(candidate, current-best) is true.
func whichExtreme(ctx *Context, args []ArgValue, name string, better func(candidate, best float64) bool) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s(x) expects 1 argument", name)
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	bestIdx := -1
	var best float64
	for i, e := range dv {
		if e.NA {
			continue
		}
		if bestIdx < 0 || better(e.Val, best) {
			best = e.Val
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return &IntVec{Data: nil}, nil
	}
	return IntScalar(int64(bestIdx + 1)), nil
}

func builtinWhichMin(ctx *Context, args []ArgValue) (Value, error) {
	return whichExtreme(ctx, args, "which.min", func(candidate, best float64) bool { return candidate < best })
}

func builtinWhichMax(ctx *Context, args []ArgValue) (Value, error) {
	return whichExtreme(ctx, args, "which.max", func(candidate, best float64) bool { return candidate > best })
}

func builtinAny(ctx *Context, args []ArgValue) (Value, error) {
	naRm, err := naRmArg(ctx, args)
	if err != nil {
		return nil, err
	}
	anyNA := false
	for _, a := range args {
		if a.Name == "na.rm" {
			continue
		}
		v, err := Force(ctx, a.Val)
		if err != nil {
			return nil, err
		}
		lv, err := asLogicalVec(ctx, v)
		if err != nil {
			return nil, err
		}
		for _, e := range lv {
			if e.NA {
				if !naRm {
					anyNA = true
				}
				continue
			}
			if e.Val {
				return LogicalScalar(true), nil
			}
		}
	}
	if anyNA {
		return LogicalNA(), nil
	}
	return LogicalScalar(false), nil
}

func builtinAll(ctx *Context, args []ArgValue) (Value, error) {
	naRm, err := naRmArg(ctx, args)
	if err != nil {
		return nil, err
	}
	anyNA := false
	for _, a := range args {
		if a.Name == "na.rm" {
			continue
		}
		v, err := Force(ctx, a.Val)
		if err != nil {
			return nil, err
		}
		lv, err := asLogicalVec(ctx, v)
		if err != nil {
			return nil, err
		}
		for _, e := range lv {
			if e.NA {
				if !naRm {
					anyNA = true
				}
				continue
			}
			if !e.Val {
				return LogicalScalar(false), nil
			}
		}
	}
	if anyNA {
		return LogicalNA(), nil
	}
	return LogicalScalar(true), nil
}

func builtinRev(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("rev(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case *DoubleVec:
		out := make([]FloatElem, len(t.Data))
		for i, e := range t.Data {
			out[len(out)-1-i] = e
		}
		return &DoubleVec{Data: out}, nil
	case *IntVec:
		out := make([]IntElem, len(t.Data))
		for i, e := range t.Data {
			out[len(out)-1-i] = e
		}
		return &IntVec{Data: out}, nil
	case *LogicalVec:
		out := make([]LogicalElem, len(t.Data))
		for i, e := range t.Data {
			out[len(out)-1-i] = e
		}
		return &LogicalVec{Data: out}, nil
	case *CharVec:
		out := make([]StringElem, len(t.Data))
		for i, e := range t.Data {
			out[len(out)-1-i] = e
		}
		return &CharVec{Data: out}, nil
	case *ListVec:
		out := make([]Value, len(t.Data))
		for i, e := range t.Data {
			out[len(out)-1-i] = e
		}
		return &ListVec{Data: out}, nil
	default:
		return nil, fmt.Errorf("rev: unsupported type %s", v.Type())
	}
}

func builtinSort(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("sort(x) expects at least 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	decreasing := false
	if dv, ok := getNamed(args, "decreasing"); ok {
		fv, _ := Force(ctx, dv)
		b, na, _ := asLogicalScalar(ctx, fv)
		if !na {
			decreasing = b
		}
	}

	switch t := v.(type) {
	case *DoubleVec:
		out := make([]FloatElem, 0, len(t.Data))
		for _, e := range t.Data {
			if !e.NA {
				out = append(out, e)
			}
		}
		sort.Slice(out, func(i, j int) bool {
			if decreasing {
				return out[i].Val > out[j].Val
			}
			return out[i].Val < out[j].Val
		})
		return &DoubleVec{Data: out}, nil
	case *IntVec:
		out := make([]IntElem, 0, len(t.Data))
		for _, e := range t.Data {
			if !e.NA {
				out = append(out, e)
			}
		}
		sort.Slice(out, func(i, j int) bool {
			if decreasing {
				return out[i].Val > out[j].Val
			}
			return out[i].Val < out[j].Val
		})
		return &IntVec{Data: out}, nil
	case *CharVec:
		out := make([]StringElem, 0, len(t.Data))
		for _, e := range t.Data {
			if !e.NA {
				out = append(out, e)
			}
		}
		sort.Slice(out, func(i, j int) bool {
			if decreasing {
				return out[i].Val > out[j].Val
			}
			return out[i].Val < out[j].Val
		})
		return &CharVec{Data: out}, nil
	default:
		return nil, fmt.Errorf("sort: unsupported type %s", v.Type())
	}
}

func builtinOrder(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("order(x) expects at least 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	decreasing := false
	if dv, ok := getNamed(args, "decreasing"); ok {
		fv, _ := Force(ctx, dv)
		b, na, _ := asLogicalScalar(ctx, fv)
		if !na {
			decreasing = b
		}
	}

	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	indices := make([]int, len(dv))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		ai := dv[indices[i]]
		aj := dv[indices[j]]
		if ai.NA && aj.NA {
			return false
		}
		if ai.NA {
			return false
		}
		if aj.NA {
			return true
		}
		if decreasing {
			return ai.Val > aj.Val
		}
		return ai.Val < aj.Val
	})
	out := make([]IntElem, len(indices))
	for i, idx := range indices {
		out[i] = IntElem{Val: int64(idx + 1)}
	}
	return &IntVec{Data: out}, nil
}

func builtinUnique(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("unique(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case *DoubleVec:
		seen := map[float64]bool{}
		var out []FloatElem
		for _, e := range t.Data {
			if e.NA {
				if !seen[0] { // use 0 as NA sentinel
					out = append(out, e)
				}
				continue
			}
			if !seen[e.Val] {
				seen[e.Val] = true
				out = append(out, e)
			}
		}
		return &DoubleVec{Data: out}, nil
	case *IntVec:
		seen := map[int64]bool{}
		var out []IntElem
		for _, e := range t.Data {
			if e.NA {
				continue
			}
			if !seen[e.Val] {
				seen[e.Val] = true
				out = append(out, e)
			}
		}
		return &IntVec{Data: out}, nil
	case *CharVec:
		seen := map[string]bool{}
		var out []StringElem
		for _, e := range t.Data {
			if e.NA {
				continue
			}
			if !seen[e.Val] {
				seen[e.Val] = true
				out = append(out, e)
			}
		}
		return &CharVec{Data: out}, nil
	case *LogicalVec:
		seen := map[bool]bool{}
		var out []LogicalElem
		for _, e := range t.Data {
			if e.NA {
				continue
			}
			if !seen[e.Val] {
				seen[e.Val] = true
				out = append(out, e)
			}
		}
		return &LogicalVec{Data: out}, nil
	default:
		return nil, fmt.Errorf("unique: unsupported type %s", v.Type())
	}
}

func builtinDuplicated(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("duplicated(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	strs := toPlainStrings(v)
	seen := map[string]bool{}
	out := make([]LogicalElem, len(strs))
	for i, s := range strs {
		if seen[s] {
			out[i] = LogicalElem{Val: true}
		} else {
			seen[s] = true
			out[i] = LogicalElem{Val: false}
		}
	}
	return &LogicalVec{Data: out}, nil
}

func builtinTable(ctx *Context, args []ArgValue) (Value, error) {
	// table(x) — returns named integer vector of counts
	if len(args) < 1 {
		return nil, fmt.Errorf("table(x) expects at least 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	strs := toPlainStrings(v)

	// Ordered unique keys
	var keys []string
	counts := map[string]int{}
	for _, s := range strs {
		if _, ok := counts[s]; !ok {
			keys = append(keys, s)
		}
		counts[s]++
	}
	sort.Strings(keys)

	data := make([]IntElem, len(keys))
	names := make([]StringElem, len(keys))
	for i, k := range keys {
		data[i] = IntElem{Val: int64(counts[k])}
		names[i] = StringElem{Val: k}
	}
	result := &IntVec{Data: data}
	result.SetAttr("names", &CharVec{Data: names})
	return result, nil
}

func builtinMatch(ctx *Context, args []ArgValue) (Value, error) {
	// match(x, table) — returns integer vector of first positions
	if len(args) < 2 {
		return nil, fmt.Errorf("match(x, table) expects 2 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	table, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	xStrs := toPlainStrings(x)
	tStrs := toPlainStrings(table)

	// Build lookup
	lookup := map[string]int{}
	for i, s := range tStrs {
		if _, ok := lookup[s]; !ok {
			lookup[s] = i + 1 // 1-indexed
		}
	}
	out := make([]IntElem, len(xStrs))
	for i, s := range xStrs {
		if pos, ok := lookup[s]; ok {
			out[i] = IntElem{Val: int64(pos)}
		} else {
			out[i] = IntElem{NA: true}
		}
	}
	return &IntVec{Data: out}, nil
}

func builtinAppend(ctx *Context, args []ArgValue) (Value, error) {
	// append(x, values, after=length(x))
	if len(args) < 2 {
		return nil, fmt.Errorf("append(x, values) expects at least 2 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	values, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	// Simple: concatenate x and values using c()
	return builtinC(ctx, []ArgValue{{Val: x}, {Val: values}})
}

func builtinSeqLen(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("seq_len(n) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	fe, err := asFloatElem(ctx, v)
	if err != nil {
		return nil, err
	}
	if fe.NA {
		return IntNA(), nil
	}
	n := int(fe.Val)
	if n < 0 {
		return nil, fmt.Errorf("seq_len: argument must be non-negative")
	}
	out := make([]IntElem, n)
	for i := 0; i < n; i++ {
		out[i] = IntElem{Val: int64(i + 1)}
	}
	return &IntVec{Data: out}, nil
}

func builtinSeqAlong(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("seq_along(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	n := v.Len()
	out := make([]IntElem, n)
	for i := 0; i < n; i++ {
		out[i] = IntElem{Val: int64(i + 1)}
	}
	return &IntVec{Data: out}, nil
}

// isNumericType reports whether setdiff/intersect/union should operate on
// the double/integer fast path rather than falling back to string keys.
func isNumericType(t string) bool {
	return t == "double" || t == "integer"
}

// filterBySetF keeps the unique, non-NA elements of xs for which keep(present
// in ySet) is true, preserving first-occurrence order. Shared by the numeric
// path of setdiff (keep = !present) and intersect (keep = present).
func filterBySetF(xs []FloatElem, ySet map[float64]bool, keep func(inY bool) bool) []FloatElem {
	var out []FloatElem
	seen := map[float64]bool{}
	for _, e := range xs {
		if e.NA || seen[e.Val] || !keep(ySet[e.Val]) {
			continue
		}
		seen[e.Val] = true
		out = append(out, e)
	}
	return out
}

func toFloatSet(dv []FloatElem) map[float64]bool {
	set := make(map[float64]bool, len(dv))
	for _, e := range dv {
		if !e.NA {
			set[e.Val] = true
		}
	}
	return set
}

// filterBySetS is the string-keyed counterpart of filterBySetF.
func filterBySetS(xs []string, ySet map[string]bool, keep func(inY bool) bool) []StringElem {
	var out []StringElem
	seen := map[string]bool{}
	for _, s := range xs {
		if seen[s] || !keep(ySet[s]) {
			continue
		}
		seen[s] = true
		out = append(out, StringElem{Val: s})
	}
	return out
}

func toStringSet(strs []string) map[string]bool {
	set := make(map[string]bool, len(strs))
	for _, s := range strs {
		set[s] = true
	}
	return set
}

func notIn(inY bool) bool { return !inY }
func isIn(inY bool) bool  { return inY }

func builtinSetdiff(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("setdiff(x, y) expects 2 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	y, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	if isNumericType(x.Type()) {
		xDv, _ := asDoubleVec(ctx, x)
		yDv, _ := asDoubleVec(ctx, y)
		return &DoubleVec{Data: filterBySetF(xDv, toFloatSet(yDv), notIn)}, nil
	}
	xStrs := toPlainStrings(x)
	yStrs := toPlainStrings(y)
	return &CharVec{Data: filterBySetS(xStrs, toStringSet(yStrs), notIn)}, nil
}

func builtinIntersect(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("intersect(x, y) expects 2 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	y, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	if isNumericType(x.Type()) {
		xDv, _ := asDoubleVec(ctx, x)
		yDv, _ := asDoubleVec(ctx, y)
		return &DoubleVec{Data: filterBySetF(xDv, toFloatSet(yDv), isIn)}, nil
	}
	xStrs := toPlainStrings(x)
	yStrs := toPlainStrings(y)
	return &CharVec{Data: filterBySetS(xStrs, toStringSet(yStrs), isIn)}, nil
}

func builtinUnion(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("union(x, y) expects 2 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	y, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	if isNumericType(x.Type()) {
		xDv, _ := asDoubleVec(ctx, x)
		yDv, _ := asDoubleVec(ctx, y)
		seen := map[float64]bool{}
		out := make([]FloatElem, 0, len(xDv)+len(yDv))
		for _, e := range xDv {
			if e.NA || seen[e.Val] {
				continue
			}
			seen[e.Val] = true
			out = append(out, e)
		}
		for _, e := range yDv {
			if e.NA || seen[e.Val] {
				continue
			}
			seen[e.Val] = true
			out = append(out, e)
		}
		return &DoubleVec{Data: out}, nil
	}
	xStrs := toPlainStrings(x)
	yStrs := toPlainStrings(y)
	seen := map[string]bool{}
	out := make([]StringElem, 0, len(xStrs)+len(yStrs))
	for _, s := range xStrs {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, StringElem{Val: s})
	}
	for _, s := range yStrs {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, StringElem{Val: s})
	}
	return &CharVec{Data: out}, nil
}

// --- Type checking ---

func typeCheck(ctx *Context, args []ArgValue, name string, check func(Value) bool) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s(x) expects 1 argument", name)
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	return LogicalScalar(check(v)), nil
}

func builtinIsNumeric(ctx *Context, args []ArgValue) (Value, error) {
	return typeCheck(ctx, args, "is.numeric", func(v Value) bool {
		return v.Type() == "double" || v.Type() == "integer"
	})
}

func builtinIsInteger(ctx *Context, args []ArgValue) (Value, error) {
	return typeCheck(ctx, args, "is.integer", func(v Value) bool {
		return v.Type() == "integer"
	})
}

func builtinIsDouble(ctx *Context, args []ArgValue) (Value, error) {
	return typeCheck(ctx, args, "is.double", func(v Value) bool {
		return v.Type() == "double"
	})
}

func builtinIsCharacter(ctx *Context, args []ArgValue) (Value, error) {
	return typeCheck(ctx, args, "is.character", func(v Value) bool {
		return v.Type() == "character"
	})
}

func builtinIsLogical(ctx *Context, args []ArgValue) (Value, error) {
	return typeCheck(ctx, args, "is.logical", func(v Value) bool {
		return v.Type() == "logical"
	})
}

func builtinIsNull(ctx *Context, args []ArgValue) (Value, error) {
	return typeCheck(ctx, args, "is.null", func(v Value) bool {
		return v.Type() == "null"
	})
}

func builtinIsList(ctx *Context, args []ArgValue) (Value, error) {
	return typeCheck(ctx, args, "is.list", func(v Value) bool {
		return v.Type() == "list"
	})
}

func builtinIsVector(ctx *Context, args []ArgValue) (Value, error) {
	return typeCheck(ctx, args, "is.vector", func(v Value) bool {
		switch v.Type() {
		case "logical", "integer", "double", "character", "list":
			return true
		}
		return false
	})
}

func builtinIsFunction(ctx *Context, args []ArgValue) (Value, error) {
	return typeCheck(ctx, args, "is.function", func(v Value) bool {
		return v.Type() == "function"
	})
}

// vecBoolUnary is the LogicalVec counterpart to vecMathUnary, shared by
// is.finite/is.nan/is.infinite: NA always maps to FALSE, otherwise fn decides.
func vecBoolUnary(ctx *Context, args []ArgValue, name string, fn func(float64) bool) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s(x) expects 1 argument", name)
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	out := make([]LogicalElem, len(dv))
	for i, e := range dv {
		if e.NA {
			out[i] = LogicalElem{Val: false}
		} else {
			out[i] = LogicalElem{Val: fn(e.Val)}
		}
	}
	return &LogicalVec{Data: out}, nil
}

func builtinIsFinite(ctx *Context, args []ArgValue) (Value, error) {
	return vecBoolUnary(ctx, args, "is.finite", func(v float64) bool { return !math.IsInf(v, 0) && !math.IsNaN(v) })
}

func builtinIsNaN(ctx *Context, args []ArgValue) (Value, error) {
	return vecBoolUnary(ctx, args, "is.nan", math.IsNaN)
}

func builtinIsInfinite(ctx *Context, args []ArgValue) (Value, error) {
	return vecBoolUnary(ctx, args, "is.infinite", func(v float64) bool { return math.IsInf(v, 0) })
}

func builtinIsDataFrame(ctx *Context, args []ArgValue) (Value, error) {
	return typeCheck(ctx, args, "is.data.frame", func(v Value) bool {
		return isDataFrame(v)
	})
}

func builtinIdentical(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("identical(x, y) expects 2 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	y, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	return LogicalScalar(x.String() == y.String() && x.Type() == y.Type()), nil
}

// --- Apply family ---

func builtinLapply(ctx *Context, args []ArgValue) (Value, error) {
	// lapply(X, FUN, ...)
	if len(args) < 2 {
		return nil, fmt.Errorf("lapply(X, FUN) expects at least 2 arguments")
	}
	xV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	funV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	callable, ok := funV.(Callable)
	if !ok {
		return nil, fmt.Errorf("lapply: FUN is not a function")
	}

	// Extra args
	extraArgs := args[2:]

	n := xV.Len()
	out := make([]Value, n)
	for i := 0; i < n; i++ {
		elem, err := vectorElement(ctx, xV, i)
		if err != nil {
			return nil, err
		}
		callArgs := make([]ArgValue, 0, 1+len(extraArgs))
		callArgs = append(callArgs, ArgValue{Val: elem})
		callArgs = append(callArgs, extraArgs...)
		res, err := callable.Call(ctx, nil, callArgs)
		if err != nil {
			return nil, err
		}
		out[i] = res
	}
	return &ListVec{Data: out}, nil
}

func builtinSapply(ctx *Context, args []ArgValue) (Value, error) {
	// sapply = lapply + simplify
	result, err := builtinLapply(ctx, args)
	if err != nil {
		return nil, err
	}
	lv, ok := result.(*ListVec)
	if !ok {
		return result, nil
	}
	return simplifyList(ctx, lv)
}

func builtinVapply(ctx *Context, args []ArgValue) (Value, error) {
	// vapply(X, FUN, FUN.VALUE, ...) — same as sapply for now
	if len(args) < 3 {
		return nil, fmt.Errorf("vapply(X, FUN, FUN.VALUE) expects at least 3 arguments")
	}
	// Drop FUN.VALUE (args[2]) and call sapply-like
	sapplyArgs := make([]ArgValue, 0, len(args)-1)
	sapplyArgs = append(sapplyArgs, args[0], args[1])
	sapplyArgs = append(sapplyArgs, args[3:]...)
	return builtinSapply(ctx, sapplyArgs)
}

func simplifyList(ctx *Context, lv *ListVec) (Value, error) {
	if lv.Len() == 0 {
		return lv, nil
	}
	// Check if all elements are scalar of the same type
	firstType := ""
	allScalar := true
	for _, v := range lv.Data {
		if v.Len() != 1 {
			allScalar = false
			break
		}
		if firstType == "" {
			firstType = v.Type()
		} else if v.Type() != firstType {
			// mixed types — try coercion
			allScalar = false
			break
		}
	}
	if !allScalar {
		return lv, nil
	}
	switch firstType {
	case "double":
		out := make([]FloatElem, lv.Len())
		for i, v := range lv.Data {
			dv, _ := asDoubleVec(ctx, v)
			out[i] = dv[0]
		}
		return &DoubleVec{Data: out}, nil
	case "integer":
		out := make([]IntElem, lv.Len())
		for i, v := range lv.Data {
			iv := v.(*IntVec)
			out[i] = iv.Data[0]
		}
		return &IntVec{Data: out}, nil
	case "logical":
		out := make([]LogicalElem, lv.Len())
		for i, v := range lv.Data {
			bv := v.(*LogicalVec)
			out[i] = bv.Data[0]
		}
		return &LogicalVec{Data: out}, nil
	case "character":
		out := make([]StringElem, lv.Len())
		for i, v := range lv.Data {
			cv := v.(*CharVec)
			out[i] = cv.Data[0]
		}
		return &CharVec{Data: out}, nil
	default:
		return lv, nil
	}
}

func builtinMapFunc(ctx *Context, args []ArgValue) (Value, error) {
	// Map(f, ...) — apply f to corresponding elements
	if len(args) < 2 {
		return nil, fmt.Errorf("Map(f, ...) expects at least 2 arguments")
	}
	funV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	callable, ok := funV.(Callable)
	if !ok {
		return nil, fmt.Errorf("Map: first argument must be a function")
	}

	// Collect the vectors
	var vecs []Value
	for _, a := range args[1:] {
		v, err := Force(ctx, a.Val)
		if err != nil {
			return nil, err
		}
		vecs = append(vecs, v)
	}

	// Find common length
	n := 0
	for _, v := range vecs {
		if v.Len() > n {
			n = v.Len()
		}
	}

	out := make([]Value, n)
	for i := 0; i < n; i++ {
		callArgs := make([]ArgValue, len(vecs))
		for j, v := range vecs {
			idx := i % v.Len()
			elem, err := vectorElement(ctx, v, idx)
			if err != nil {
				return nil, err
			}
			callArgs[j] = ArgValue{Val: elem}
		}
		res, err := callable.Call(ctx, nil, callArgs)
		if err != nil {
			return nil, err
		}
		out[i] = res
	}
	return &ListVec{Data: out}, nil
}

func builtinReduce(ctx *Context, args []ArgValue) (Value, error) {
	// Reduce(f, x, init)
	if len(args) < 2 {
		return nil, fmt.Errorf("Reduce(f, x) expects at least 2 arguments")
	}
	funV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	callable, ok := funV.(Callable)
	if !ok {
		return nil, fmt.Errorf("Reduce: first argument must be a function")
	}
	xV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}

	var acc Value
	start := 0
	if v, ok := getNamed(args, "init"); ok {
		acc, err = Force(ctx, v)
		if err != nil {
			return nil, err
		}
	} else if len(args) >= 3 && args[2].Name == "" {
		acc, err = Force(ctx, args[2].Val)
		if err != nil {
			return nil, err
		}
	} else {
		if xV.Len() == 0 {
			return nil, fmt.Errorf("Reduce: empty sequence with no init")
		}
		acc, err = vectorElement(ctx, xV, 0)
		if err != nil {
			return nil, err
		}
		start = 1
	}

	for i := start; i < xV.Len(); i++ {
		elem, err := vectorElement(ctx, xV, i)
		if err != nil {
			return nil, err
		}
		acc, err = callable.Call(ctx, nil, []ArgValue{{Val: acc}, {Val: elem}})
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

func builtinFilter(ctx *Context, args []ArgValue) (Value, error) {
	// Filter(f, x)
	if len(args) != 2 {
		return nil, fmt.Errorf("Filter(f, x) expects 2 arguments")
	}
	funV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	callable, ok := funV.(Callable)
	if !ok {
		return nil, fmt.Errorf("Filter: first argument must be a function")
	}
	xV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}

	var out []Value
	for i := 0; i < xV.Len(); i++ {
		elem, err := vectorElement(ctx, xV, i)
		if err != nil {
			return nil, err
		}
		res, err := callable.Call(ctx, nil, []ArgValue{{Val: elem}})
		if err != nil {
			return nil, err
		}
		b, na, err := asLogicalScalar(ctx, res)
		if err != nil {
			return nil, err
		}
		if !na && b {
			out = append(out, elem)
		}
	}
	if out == nil {
		out = []Value{}
	}
	return &ListVec{Data: out}, nil
}

func builtinDoCall(ctx *Context, args []ArgValue) (Value, error) {
	// do.call(fun, args)
	if len(args) < 2 {
		return nil, fmt.Errorf("do.call(fun, args) expects 2 arguments")
	}
	funV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	callable, ok := funV.(Callable)
	if !ok {
		return nil, fmt.Errorf("do.call: first argument must be a function")
	}
	argsV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	lv, ok := argsV.(*ListVec)
	if !ok {
		// If not a list, call with single arg
		return callable.Call(ctx, nil, []ArgValue{{Val: argsV}})
	}

	callArgs := make([]ArgValue, lv.Len())
	names, hasNames := listNames(lv)
	for i, v := range lv.Data {
		name := ""
		if hasNames && i < len(names) {
			name = names[i]
		}
		callArgs[i] = ArgValue{Name: name, Val: v}
	}
	return callable.Call(ctx, nil, callArgs)
}

// --- Control / error handling ---

func builtinIfelse(ctx *Context, args []ArgValue) (Value, error) {
	// ifelse(test, yes, no)
	if len(args) != 3 {
		return nil, fmt.Errorf("ifelse(test, yes, no) expects 3 arguments")
	}
	test, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	yes, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	no, err := Force(ctx, args[2].Val)
	if err != nil {
		return nil, err
	}

	lv, err := asLogicalVec(ctx, test)
	if err != nil {
		return nil, err
	}

	n := len(lv)
	// Result type follows yes
	switch yv := yes.(type) {
	case *DoubleVec:
		nv, _ := asDoubleVec(ctx, no)
		out := make([]FloatElem, n)
		for i, e := range lv {
			if e.NA {
				out[i] = FloatElem{NA: true}
			} else if e.Val {
				out[i] = yv.Data[i%yv.Len()]
			} else {
				out[i] = nv[i%len(nv)]
			}
		}
		return &DoubleVec{Data: out}, nil
	case *IntVec:
		nv, _ := coerceToIntVec(ctx, no)
		out := make([]IntElem, n)
		for i, e := range lv {
			if e.NA {
				out[i] = IntElem{NA: true}
			} else if e.Val {
				out[i] = yv.Data[i%yv.Len()]
			} else {
				out[i] = nv[i%len(nv)]
			}
		}
		return &IntVec{Data: out}, nil
	case *CharVec:
		nv, _ := asCharVec(ctx, no)
		out := make([]StringElem, n)
		for i, e := range lv {
			if e.NA {
				out[i] = StringElem{NA: true}
			} else if e.Val {
				out[i] = yv.Data[i%yv.Len()]
			} else {
				out[i] = nv[i%len(nv)]
			}
		}
		return &CharVec{Data: out}, nil
	default:
		// Generic: use list
		out := make([]Value, n)
		for i, e := range lv {
			if e.NA {
				out[i] = NullValue
			} else if e.Val {
				elem, _ := vectorElement(ctx, yes, i%yes.Len())
				out[i] = elem
			} else {
				elem, _ := vectorElement(ctx, no, i%no.Len())
				out[i] = elem
			}
		}
		return &ListVec{Data: out}, nil
	}
}

func builtinSwitch(ctx *Context, args []ArgValue) (Value, error) {
	// switch(EXPR, case1=val1, case2=val2, ...)
	if len(args) < 2 {
		return nil, fmt.Errorf("switch() expects at least 2 arguments")
	}
	exprV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	cv, err := asCharVec(ctx, exprV)
	if err != nil {
		return nil, err
	}
	if len(cv) == 0 || cv[0].NA {
		return NullValue, nil
	}
	key := cv[0].Val

	var defaultVal Value
	for _, a := range args[1:] {
		if a.Name == key {
			v, err := Force(ctx, a.Val)
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		if a.Name == "" {
			v, err := Force(ctx, a.Val)
			if err != nil {
				return nil, err
			}
			defaultVal = v
		}
	}
	if defaultVal != nil {
		return defaultVal, nil
	}
	return NullValue, nil
}

func builtinTryCatch(ctx *Context, args []ArgValue) (Value, error) {
	// tryCatch(expr, error=function(e) {...}, warning=function(w) {...})
	if len(args) < 1 {
		return nil, fmt.Errorf("tryCatch() expects at least 1 argument")
	}

	// Evaluate the expression
	result, err := Force(ctx, args[0].Val)
	if err != nil {
		// Check if there's an error handler
		if handler, ok := getNamed(args, "error"); ok {
			handlerV, herr := Force(ctx, handler)
			if herr != nil {
				return nil, herr
			}
			callable, ok := handlerV.(Callable)
			if !ok {
				return nil, fmt.Errorf("tryCatch: error handler is not a function")
			}
			// Create error message value
			errMsg := CharScalar(err.Error())
			return callable.Call(ctx, nil, []ArgValue{{Val: errMsg}})
		}
		return nil, err
	}

	return result, nil
}

func builtinMessage(ctx *Context, args []ArgValue) (Value, error) {
	fargs, err := forceArgs(ctx, args)
	if err != nil {
		return nil, err
	}
	var parts []string
	for _, a := range fargs {
		ps := toPlainStrings(a.Val)
		parts = append(parts, ps...)
	}
	ctx.Println(strings.Join(parts, ""))
	return NullValue, nil
}

func builtinNargs(ctx *Context, args []ArgValue) (Value, error) {
	return IntScalar(int64(len(args))), nil
}

// --- Environment ---

func builtinExists(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("exists(x) expects at least 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	cv, ok := v.(*CharVec)
	if !ok || cv.Len() < 1 || cv.Data[0].NA {
		return nil, fmt.Errorf("exists: argument must be a character string")
	}
	name := cv.Data[0].Val
	_, found := ctx.Global.Get(name)
	return LogicalScalar(found), nil
}

func builtinEnvironment(ctx *Context, args []ArgValue) (Value, error) {
	// Simplified: just returns a description
	return CharScalar("<environment>"), nil
}

func builtinSysTime(ctx *Context, args []ArgValue) (Value, error) {
	// Return current time as a numeric (not using time package to keep it simple)
	return CharScalar("Sys.time() not implemented in smallR"), nil
}

func builtinWhichNA(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("which.na(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	// find indices of NA values
	var out []IntElem
	switch t := v.(type) {
	case *LogicalVec:
		for i, e := range t.Data {
			if e.NA {
				out = append(out, IntElem{Val: int64(i + 1)})
			}
		}
	case *IntVec:
		for i, e := range t.Data {
			if e.NA {
				out = append(out, IntElem{Val: int64(i + 1)})
			}
		}
	case *DoubleVec:
		for i, e := range t.Data {
			if e.NA {
				out = append(out, IntElem{Val: int64(i + 1)})
			}
		}
	case *CharVec:
		for i, e := range t.Data {
			if e.NA {
				out = append(out, IntElem{Val: int64(i + 1)})
			}
		}
	}
	if out == nil {
		out = []IntElem{}
	}
	return &IntVec{Data: out}, nil
}

func builtinTabulate(ctx *Context, args []ArgValue) (Value, error) {
	// tabulate(bin, nbins = max(1, bin, na.rm=TRUE))
	if len(args) < 1 {
		return nil, fmt.Errorf("tabulate(bin) expects at least 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	iv, err := coerceToIntVec(ctx, v)
	if err != nil {
		return nil, err
	}
	nbins := 0
	for _, e := range iv {
		if !e.NA && int(e.Val) > nbins {
			nbins = int(e.Val)
		}
	}
	if len(args) >= 2 {
		nbV, err := Force(ctx, args[1].Val)
		if err != nil {
			return nil, err
		}
		fe, err := asFloatElem(ctx, nbV)
		if err != nil {
			return nil, err
		}
		if !fe.NA {
			nbins = int(fe.Val)
		}
	}
	out := make([]IntElem, nbins)
	for _, e := range iv {
		if e.NA || e.Val < 1 || int(e.Val) > nbins {
			continue
		}
		out[e.Val-1].Val++
	}
	return &IntVec{Data: out}, nil
}
