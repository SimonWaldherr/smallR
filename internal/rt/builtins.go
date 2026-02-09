package rt

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func InstallBuiltins(env *Env) {
	installMathBuiltins(env)
	installStringBuiltins(env)
	installUtilBuiltins(env)

	builtins := map[string]*BuiltinFunc{
		"print":        {FnName: "print", Impl: builtinPrint},
		"cat":          {FnName: "cat", Impl: builtinCat},
		"c":            {FnName: "c", Impl: builtinC},
		"list":         {FnName: "list", Impl: builtinList},
		"data.frame":   {FnName: "data.frame", Impl: builtinDataFrame},
		"nrow":         {FnName: "nrow", Impl: builtinNRow},
		"ncol":         {FnName: "ncol", Impl: builtinNCol},
		"dim":          {FnName: "dim", Impl: builtinDim},
		"head":         {FnName: "head", Impl: builtinHead},
		"tail":         {FnName: "tail", Impl: builtinTail},
		"length":       {FnName: "length", Impl: builtinLength},
		"sum":          {FnName: "sum", Impl: builtinSum},
		"mean":         {FnName: "mean", Impl: builtinMean},
		"sd":           {FnName: "sd", Impl: builtinSD},
		"seq":          {FnName: "seq", Impl: builtinSeq},
		"rep":          {FnName: "rep", Impl: builtinRep},
		"typeof":       {FnName: "typeof", Impl: builtinTypeof},
		"class":        {FnName: "class", Impl: builtinClass},
		"attr":         {FnName: "attr", Impl: builtinAttr},
		"attributes":   {FnName: "attributes", Impl: builtinAttributes},
		"names":        {FnName: "names", Impl: builtinNames},
		"is.na":        {FnName: "is.na", Impl: builtinIsNA},
		"as.integer":   {FnName: "as.integer", Impl: builtinAsInteger},
		"as.numeric":   {FnName: "as.numeric", Impl: builtinAsNumeric},
		"as.character": {FnName: "as.character", Impl: builtinAsCharacter},
		"as.logical":   {FnName: "as.logical", Impl: builtinAsLogical},
		"stop":         {FnName: "stop", Impl: builtinStop},
		"warning":      {FnName: "warning", Impl: builtinWarning},
		"str":          {FnName: "str", Impl: builtinStr},
	}
	for name, fn := range builtins {
		env.SetLocal(name, fn)
	}
}

func forceArgs(ctx *Context, args []ArgValue) ([]ArgValue, error) {
	out := make([]ArgValue, len(args))
	for i, a := range args {
		v, err := Force(ctx, a.Val)
		if err != nil {
			return nil, err
		}
		out[i] = ArgValue{Name: a.Name, Val: v}
	}
	return out, nil
}

func getNamed(args []ArgValue, name string) (Value, bool) {
	for _, a := range args {
		if a.Name == name {
			return a.Val, true
		}
	}
	return nil, false
}

func builtinPrint(ctx *Context, args []ArgValue) (Value, error) {
	fargs, err := forceArgs(ctx, args)
	if err != nil {
		return nil, err
	}
	if len(fargs) == 0 {
		ctx.Println("NULL")
		return NullValue, nil
	}
	for _, a := range fargs {
		ctx.Println(a.Val.String())
	}
	return fargs[0].Val, nil
}

func builtinCat(ctx *Context, args []ArgValue) (Value, error) {
	fargs, err := forceArgs(ctx, args)
	if err != nil {
		return nil, err
	}
	sep := " "
	if v, ok := getNamed(fargs, "sep"); ok {
		if s, ok := v.(*CharVec); ok && s.Len() >= 1 && !s.Data[0].NA {
			sep = s.Data[0].Val
		}
	}
	end := ""
	if v, ok := getNamed(fargs, "end"); ok {
		if s, ok := v.(*CharVec); ok && s.Len() >= 1 && !s.Data[0].NA {
			end = s.Data[0].Val
		}
	}

	var parts []string
	for _, a := range fargs {
		if a.Name == "sep" || a.Name == "end" {
			continue
		}
		ps := toPlainStrings(a.Val)
		parts = append(parts, ps...)
	}
	out := strings.Join(parts, sep) + end
	_ = write(ctx, out)
	return NullValue, nil
}

func toPlainStrings(v Value) []string {
	switch t := v.(type) {
	case *CharVec:
		out := make([]string, 0, t.Len())
		for _, e := range t.Data {
			if e.NA {
				out = append(out, "NA")
			} else {
				out = append(out, e.Val)
			}
		}
		return out
	case *DoubleVec:
		out := make([]string, 0, t.Len())
		for _, e := range t.Data {
			if e.NA {
				out = append(out, "NA")
			} else {
				out = append(out, strconv.FormatFloat(e.Val, 'g', -1, 64))
			}
		}
		return out
	case *IntVec:
		out := make([]string, 0, t.Len())
		for _, e := range t.Data {
			if e.NA {
				out = append(out, "NA")
			} else {
				out = append(out, strconv.FormatInt(e.Val, 10))
			}
		}
		return out
	case *LogicalVec:
		out := make([]string, 0, t.Len())
		for _, e := range t.Data {
			if e.NA {
				out = append(out, "NA")
			} else if e.Val {
				out = append(out, "TRUE")
			} else {
				out = append(out, "FALSE")
			}
		}
		return out
	default:
		return []string{v.String()}
	}
}

func builtinC(ctx *Context, args []ArgValue) (Value, error) {
	fargs, err := forceArgs(ctx, args)
	if err != nil {
		return nil, err
	}
	// Determine target type
	target := "logical"
	hasList := false
	for _, a := range fargs {
		switch a.Val.Type() {
		case "list":
			hasList = true
		case "character":
			target = "character"
		case "double":
			if target != "character" {
				target = "double"
			}
		case "integer":
			if target != "character" && target != "double" {
				target = "integer"
			}
		case "logical":
			// keep
		default:
			hasList = true
		}
	}
	if hasList {
		var out []Value
		var names []StringElem
		for _, a := range fargs {
			if lv, ok := a.Val.(*ListVec); ok {
				out = append(out, lv.Data...)
				// ignore names for now
				_ = names
			} else {
				out = append(out, a.Val)
			}
		}
		return &ListVec{Data: out}, nil
	}

	switch target {
	case "character":
		var out []StringElem
		for _, a := range fargs {
			cv, err := asCharVec(ctx, a.Val)
			if err != nil {
				return nil, err
			}
			out = append(out, cv...)
		}
		return &CharVec{Data: out}, nil
	case "double":
		var out []FloatElem
		for _, a := range fargs {
			dv, err := asDoubleVec(ctx, a.Val)
			if err != nil {
				return nil, err
			}
			out = append(out, dv...)
		}
		return &DoubleVec{Data: out}, nil
	case "integer":
		var out []IntElem
		for _, a := range fargs {
			iv, err := coerceToIntVec(ctx, a.Val)
			if err != nil {
				return nil, err
			}
			out = append(out, iv...)
		}
		return &IntVec{Data: out}, nil
	default:
		var out []LogicalElem
		for _, a := range fargs {
			lv, err := asLogicalVec(ctx, a.Val)
			if err != nil {
				return nil, err
			}
			out = append(out, lv...)
		}
		return &LogicalVec{Data: out}, nil
	}
}

func builtinList(ctx *Context, args []ArgValue) (Value, error) {
	fargs, err := forceArgs(ctx, args)
	if err != nil {
		return nil, err
	}
	out := make([]Value, 0, len(fargs))
	names := make([]StringElem, 0, len(fargs))
	for i, a := range fargs {
		_ = i
		out = append(out, a.Val)
		if a.Name == "" {
			names = append(names, StringElem{Val: ""})
		} else {
			names = append(names, StringElem{Val: a.Name})
		}
	}
	l := &ListVec{Data: out}
	l.SetAttr("names", &CharVec{Data: names})
	return l, nil
}

func builtinLength(ctx *Context, args []ArgValue) (Value, error) {
	fargs, err := forceArgs(ctx, args)
	if err != nil {
		return nil, err
	}
	if len(fargs) != 1 {
		return nil, fmt.Errorf("length() expects 1 argument")
	}
	return IntScalar(int64(fargs[0].Val.Len())), nil
}

func builtinSum(ctx *Context, args []ArgValue) (Value, error) {
	// sum(..., na.rm=FALSE)
	naRm := false
	// Do not force until we parse na.rm
	for _, a := range args {
		if a.Name == "na.rm" {
			v, err := Force(ctx, a.Val)
			if err != nil {
				return nil, err
			}
			b, na, err := asLogicalScalar(ctx, v)
			if err != nil {
				return nil, err
			}
			if na {
				naRm = false
			} else {
				naRm = b
			}
		}
	}
	var sum float64
	anyNA := false
	for _, a := range args {
		if a.Name == "na.rm" {
			continue
		}
		v, err := Force(ctx, a.Val)
		if err != nil {
			return nil, err
		}
		dv, err := asDoubleVec(ctx, v)
		if err != nil {
			return nil, err
		}
		for _, e := range dv {
			if e.NA {
				if naRm {
					continue
				}
				anyNA = true
				continue
			}
			sum += e.Val
		}
	}
	if anyNA && !naRm {
		return DoubleNA(), nil
	}
	return DoubleScalar(sum), nil
}

func builtinMean(ctx *Context, args []ArgValue) (Value, error) {
	naRm := false
	for _, a := range args {
		if a.Name == "na.rm" {
			v, err := Force(ctx, a.Val)
			if err != nil {
				return nil, err
			}
			b, na, err := asLogicalScalar(ctx, v)
			if err != nil {
				return nil, err
			}
			if na {
				naRm = false
			} else {
				naRm = b
			}
		}
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("mean() expects at least 1 argument")
	}
	// mean(x) only
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	var sum float64
	var n int
	for _, e := range dv {
		if e.NA {
			if naRm {
				continue
			}
			return DoubleNA(), nil
		}
		sum += e.Val
		n++
	}
	if n == 0 {
		return DoubleNA(), nil
	}
	return DoubleScalar(sum / float64(n)), nil
}

func builtinSD(ctx *Context, args []ArgValue) (Value, error) {
	naRm := false
	for _, a := range args {
		if a.Name == "na.rm" {
			v, err := Force(ctx, a.Val)
			if err != nil {
				return nil, err
			}
			b, na, err := asLogicalScalar(ctx, v)
			if err != nil {
				return nil, err
			}
			if na {
				naRm = false
			} else {
				naRm = b
			}
		}
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("sd() expects at least 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}

	// First pass: calculate mean
	var sum float64
	var n int
	for _, e := range dv {
		if e.NA {
			if naRm {
				continue
			}
			return DoubleNA(), nil
		}
		sum += e.Val
		n++
	}
	if n <= 1 {
		return DoubleNA(), nil
	}
	mean := sum / float64(n)

	// Second pass: calculate variance
	var variance float64
	for _, e := range dv {
		if e.NA && naRm {
			continue
		}
		diff := e.Val - mean
		variance += diff * diff
	}
	variance /= float64(n - 1) // sample standard deviation

	return DoubleScalar(mathSqrt(variance)), nil
}

func mathSqrt(x float64) float64 {
	return math.Sqrt(x)
}

func builtinSeq(ctx *Context, args []ArgValue) (Value, error) {
	// seq(to) or seq(from=, to=, by=)
	var from float64 = 1
	var to float64
	var by float64 = 1
	hasTo := false

	// positional
	if len(args) == 1 && args[0].Name == "" {
		v, err := Force(ctx, args[0].Val)
		if err != nil {
			return nil, err
		}
		fe, err := asFloatElem(ctx, v)
		if err != nil {
			return nil, err
		}
		if fe.NA {
			return &DoubleVec{Data: []FloatElem{{NA: true}}}, nil
		}
		to = fe.Val
		hasTo = true
	} else {
		for _, a := range args {
			switch a.Name {
			case "from":
				v, _ := Force(ctx, a.Val)
				fe, err := asFloatElem(ctx, v)
				if err != nil {
					return nil, err
				}
				if fe.NA {
					return DoubleNA(), nil
				}
				from = fe.Val
			case "to":
				v, _ := Force(ctx, a.Val)
				fe, err := asFloatElem(ctx, v)
				if err != nil {
					return nil, err
				}
				if fe.NA {
					return DoubleNA(), nil
				}
				to = fe.Val
				hasTo = true
			case "by":
				v, _ := Force(ctx, a.Val)
				fe, err := asFloatElem(ctx, v)
				if err != nil {
					return nil, err
				}
				if fe.NA {
					return DoubleNA(), nil
				}
				by = fe.Val
			}
		}
	}
	if !hasTo {
		return nil, fmt.Errorf("seq() missing 'to'")
	}
	if by == 0 {
		return nil, fmt.Errorf("seq() by must be non-zero")
	}
	// generate
	n := int(((to - from) / by) + 1)
	if n < 0 {
		n = -n
	}
	if n > 1000000 {
		return nil, fmt.Errorf("seq() too long")
	}
	out := make([]FloatElem, 0, n)
	cur := from
	if (by > 0 && from > to) || (by < 0 && from < to) {
		// empty
		return &DoubleVec{Data: out}, nil
	}
	for {
		if (by > 0 && cur > to) || (by < 0 && cur < to) {
			break
		}
		out = append(out, FloatElem{Val: cur})
		cur += by
	}
	return &DoubleVec{Data: out}, nil
}

func builtinRep(ctx *Context, args []ArgValue) (Value, error) {
	// rep(x, times)
	if len(args) < 2 {
		return nil, fmt.Errorf("rep() expects x and times")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	timesV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	te, err := asFloatElem(ctx, timesV)
	if err != nil {
		return nil, err
	}
	if te.NA {
		return DoubleNA(), nil
	}
	times := int(te.Val)
	if times < 0 {
		return nil, fmt.Errorf("invalid 'times' argument")
	}
	switch xv := x.(type) {
	case *DoubleVec:
		out := make([]FloatElem, 0, xv.Len()*times)
		for i := 0; i < times; i++ {
			out = append(out, xv.Data...)
		}
		return &DoubleVec{Data: out}, nil
	case *IntVec:
		out := make([]IntElem, 0, xv.Len()*times)
		for i := 0; i < times; i++ {
			out = append(out, xv.Data...)
		}
		return &IntVec{Data: out}, nil
	case *LogicalVec:
		out := make([]LogicalElem, 0, xv.Len()*times)
		for i := 0; i < times; i++ {
			out = append(out, xv.Data...)
		}
		return &LogicalVec{Data: out}, nil
	case *CharVec:
		out := make([]StringElem, 0, xv.Len()*times)
		for i := 0; i < times; i++ {
			out = append(out, xv.Data...)
		}
		return &CharVec{Data: out}, nil
	default:
		if lv, ok := x.(*ListVec); ok {
			out := make([]Value, 0, lv.Len()*times)
			for i := 0; i < times; i++ {
				out = append(out, lv.Data...)
			}
			return &ListVec{Data: out}, nil
		}
		return nil, fmt.Errorf("rep() unsupported type %s", x.Type())
	}
}

func builtinTypeof(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("typeof() expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	return CharScalar(v.Type()), nil
}

func builtinClass(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("class() expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	if cls, ok := v.GetAttr("class"); ok {
		return cls, nil
	}
	return CharScalar(v.Type()), nil
}

func builtinAttr(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("attr(x, which) expects 2 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	whichV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	cv, ok := whichV.(*CharVec)
	if !ok || cv.Len() < 1 || cv.Data[0].NA {
		return nil, fmt.Errorf("attr: 'which' must be character")
	}
	name := cv.Data[0].Val
	if v, ok := x.GetAttr(name); ok {
		return v, nil
	}
	return NullValue, nil
}

func builtinAttributes(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("attributes(x) expects 1 argument")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	attrs := x.Attrs()
	l := &ListVec{}
	if len(attrs) == 0 {
		return NullValue, nil
	}
	names := make([]StringElem, 0, len(attrs))
	data := make([]Value, 0, len(attrs))
	for k, v := range attrs {
		names = append(names, StringElem{Val: k})
		data = append(data, v)
	}
	l.Data = data
	l.SetAttr("names", &CharVec{Data: names})
	return l, nil
}

func builtinNames(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("names(x) expects 1 argument")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	if nm, ok := x.GetAttr("names"); ok {
		return nm, nil
	}
	return NullValue, nil
}

func builtinIsNA(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is.na(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case *LogicalVec:
		out := make([]LogicalElem, t.Len())
		for i, e := range t.Data {
			out[i] = LogicalElem{Val: e.NA}
		}
		return &LogicalVec{Data: out}, nil
	case *IntVec:
		out := make([]LogicalElem, t.Len())
		for i, e := range t.Data {
			out[i] = LogicalElem{Val: e.NA}
		}
		return &LogicalVec{Data: out}, nil
	case *DoubleVec:
		out := make([]LogicalElem, t.Len())
		for i, e := range t.Data {
			out[i] = LogicalElem{Val: e.NA}
		}
		return &LogicalVec{Data: out}, nil
	case *CharVec:
		out := make([]LogicalElem, t.Len())
		for i, e := range t.Data {
			out[i] = LogicalElem{Val: e.NA}
		}
		return &LogicalVec{Data: out}, nil
	default:
		return &LogicalVec{Data: []LogicalElem{{Val: false}}}, nil
	}
}

func builtinAsInteger(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("as.integer(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	iv, err := coerceToIntVec(ctx, v)
	if err != nil {
		return nil, err
	}
	return &IntVec{Data: iv}, nil
}

func builtinAsNumeric(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("as.numeric(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	return &DoubleVec{Data: dv}, nil
}

func builtinAsCharacter(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("as.character(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	cv, err := asCharVec(ctx, v)
	if err != nil {
		return nil, err
	}
	return &CharVec{Data: cv}, nil
}

func builtinAsLogical(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("as.logical(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	lv, err := asLogicalVec(ctx, v)
	if err != nil {
		return nil, err
	}
	return &LogicalVec{Data: lv}, nil
}

func builtinStop(ctx *Context, args []ArgValue) (Value, error) {
	fargs, err := forceArgs(ctx, args)
	if err != nil {
		return nil, err
	}
	msg := "stopped"
	if len(fargs) > 0 {
		msg = strings.Join(toPlainStrings(fargs[0].Val), " ")
	}
	return nil, fmt.Errorf("%s", msg)
}

func builtinWarning(ctx *Context, args []ArgValue) (Value, error) {
	fargs, err := forceArgs(ctx, args)
	if err != nil {
		return nil, err
	}
	msg := "warning"
	if len(fargs) > 0 {
		msg = strings.Join(toPlainStrings(fargs[0].Val), " ")
	}
	ctx.Println("Warning:", msg)
	return NullValue, nil
}

func builtinStr(ctx *Context, args []ArgValue) (Value, error) {
	fargs, err := forceArgs(ctx, args)
	if err != nil {
		return nil, err
	}
	if len(fargs) != 1 {
		return nil, fmt.Errorf("str(x) expects 1 argument")
	}
	v := fargs[0].Val
	ctx.Println(DebugValue(v))
	return v, nil
}

// --- Data frame helpers ---

func isDataFrame(v Value) bool {
	cls, ok := v.GetAttr("class")
	if !ok {
		return false
	}
	cv, ok := cls.(*CharVec)
	if !ok || cv.Len() == 0 {
		return false
	}
	for _, e := range cv.Data {
		if e.NA {
			continue
		}
		if e.Val == "data.frame" {
			return true
		}
	}
	return false
}

func builtinDataFrame(ctx *Context, args []ArgValue) (Value, error) {
	// data.frame(..., stringsAsFactors=FALSE, check.names=TRUE, row.names=...)
	// We implement a minimal version: columns are vectors/lists; lengths are recycled to max.
	fargs, err := forceArgs(ctx, args)
	if err != nil {
		return nil, err
	}

	var cols []Value
	var colNames []StringElem

	// Optional row.names
	var rowNames Value = nil

	colAuto := 1
	for _, a := range fargs {
		switch a.Name {
		case "stringsAsFactors", "check.names":
			continue
		case "row.names":
			rowNames = a.Val
			continue
		}
		v := a.Val
		if v == NullValue {
			// In R, NULL columns are dropped; mimic that.
			continue
		}
		cols = append(cols, v)
		if a.Name != "" {
			colNames = append(colNames, StringElem{Val: a.Name})
		} else {
			colNames = append(colNames, StringElem{Val: fmt.Sprintf("V%d", colAuto)})
			colAuto++
		}
	}

	// Determine nrow
	nrow := 0
	for _, c := range cols {
		if c.Len() > nrow {
			nrow = c.Len()
		}
	}
	// Recycle columns
	for i, c := range cols {
		if c.Len() == nrow {
			continue
		}
		rc, err := recycleTo(ctx, c, nrow)
		if err != nil {
			return nil, err
		}
		cols[i] = rc
	}

	df := &ListVec{Data: cols}
	df.SetAttr("names", &CharVec{Data: colNames})
	df.SetAttr("class", &CharVec{Data: []StringElem{{Val: "data.frame"}}})

	// row.names
	if rowNames != nil {
		df.SetAttr("row.names", rowNames)
	} else {
		// default 1:nrow
		rn := make([]IntElem, nrow)
		for i := 0; i < nrow; i++ {
			rn[i] = IntElem{Val: int64(i + 1)}
		}
		df.SetAttr("row.names", &IntVec{Data: rn})
	}
	return df, nil
}

func recycleTo(ctx *Context, v Value, n int) (Value, error) {
	_ = ctx
	if n < 0 {
		return nil, fmt.Errorf("invalid target length")
	}
	if n == 0 {
		// empty of same type
		switch t := v.(type) {
		case *DoubleVec:
			return &DoubleVec{Data: nil}, nil
		case *IntVec:
			return &IntVec{Data: nil}, nil
		case *LogicalVec:
			return &LogicalVec{Data: nil}, nil
		case *CharVec:
			return &CharVec{Data: nil}, nil
		case *ListVec:
			return &ListVec{Data: nil}, nil
		default:
			_ = t
			return &ListVec{Data: nil}, nil
		}
	}
	if v.Len() == 0 {
		return nil, fmt.Errorf("cannot recycle length-0 vector")
	}
	switch t := v.(type) {
	case *DoubleVec:
		out := make([]FloatElem, n)
		for i := 0; i < n; i++ {
			out[i] = t.Data[i%t.Len()]
		}
		return &DoubleVec{Data: out}, nil
	case *IntVec:
		out := make([]IntElem, n)
		for i := 0; i < n; i++ {
			out[i] = t.Data[i%t.Len()]
		}
		return &IntVec{Data: out}, nil
	case *LogicalVec:
		out := make([]LogicalElem, n)
		for i := 0; i < n; i++ {
			out[i] = t.Data[i%t.Len()]
		}
		return &LogicalVec{Data: out}, nil
	case *CharVec:
		out := make([]StringElem, n)
		for i := 0; i < n; i++ {
			out[i] = t.Data[i%t.Len()]
		}
		return &CharVec{Data: out}, nil
	case *ListVec:
		out := make([]Value, n)
		for i := 0; i < n; i++ {
			out[i] = t.Data[i%t.Len()]
		}
		return &ListVec{Data: out}, nil
	default:
		return nil, fmt.Errorf("cannot recycle type %s", v.Type())
	}
}

func builtinNRow(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("nrow(x) expects 1 argument")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	if isDataFrame(x) {
		if lv, ok := x.(*ListVec); ok && lv.Len() > 0 {
			return IntScalar(int64(lv.Data[0].Len())), nil
		}
		return IntScalar(0), nil
	}
	return NullValue, nil
}

func builtinNCol(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ncol(x) expects 1 argument")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	if isDataFrame(x) {
		return IntScalar(int64(x.Len())), nil
	}
	return NullValue, nil
}

func builtinDim(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dim(x) expects 1 argument")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	if isDataFrame(x) {
		nr := int64(0)
		if lv, ok := x.(*ListVec); ok && lv.Len() > 0 {
			nr = int64(lv.Data[0].Len())
		}
		nc := int64(x.Len())
		return &IntVec{Data: []IntElem{{Val: nr}, {Val: nc}}}, nil
	}
	return NullValue, nil
}

func builtinHead(ctx *Context, args []ArgValue) (Value, error) {
	// head(x, n=6)
	if len(args) < 1 {
		return nil, fmt.Errorf("head() expects at least 1 argument")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	n := 6
	if len(args) >= 2 {
		nv, err := Force(ctx, args[1].Val)
		if err != nil {
			return nil, err
		}
		fe, err := asFloatElem(ctx, nv)
		if err != nil {
			return nil, err
		}
		if !fe.NA {
			n = int(fe.Val)
		}
	}
	if n < 0 {
		n = 0
	}
	if isDataFrame(x) {
		return dfHeadTail(ctx, x, n, true)
	}
	// generic vector/list head using integer indices
	if x.Len() == 0 {
		return x, nil
	}
	if n > x.Len() {
		n = x.Len()
	}
	ind := make([]IntElem, n)
	for i := 0; i < n; i++ {
		ind[i] = IntElem{Val: int64(i + 1)}
	}
	return subset(ctx, x, &IntVec{Data: ind}, false)
}

func builtinTail(ctx *Context, args []ArgValue) (Value, error) {
	// tail(x, n=6)
	if len(args) < 1 {
		return nil, fmt.Errorf("tail() expects at least 1 argument")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	n := 6
	if len(args) >= 2 {
		nv, err := Force(ctx, args[1].Val)
		if err != nil {
			return nil, err
		}
		fe, err := asFloatElem(ctx, nv)
		if err != nil {
			return nil, err
		}
		if !fe.NA {
			n = int(fe.Val)
		}
	}
	if n < 0 {
		n = 0
	}
	if isDataFrame(x) {
		return dfHeadTail(ctx, x, n, false)
	}
	if x.Len() == 0 {
		return x, nil
	}
	if n > x.Len() {
		n = x.Len()
	}
	ind := make([]IntElem, n)
	start := x.Len() - n
	for i := 0; i < n; i++ {
		ind[i] = IntElem{Val: int64(start + i + 1)}
	}
	return subset(ctx, x, &IntVec{Data: ind}, false)
}

func dfHeadTail(ctx *Context, x Value, n int, head bool) (Value, error) {
	lv, ok := x.(*ListVec)
	if !ok {
		return nil, fmt.Errorf("expected data.frame to be a list")
	}
	nrow := 0
	if lv.Len() > 0 {
		nrow = lv.Data[0].Len()
	}
	if n > nrow {
		n = nrow
	}
	start := 0
	if !head {
		start = nrow - n
		if start < 0 {
			start = 0
		}
	}
	ind := make([]IntElem, n)
	for i := 0; i < n; i++ {
		ind[i] = IntElem{Val: int64(start + i + 1)}
	}

	newCols := make([]Value, len(lv.Data))
	for i, col := range lv.Data {
		v, err := subset(ctx, col, &IntVec{Data: ind}, false)
		if err != nil {
			return nil, err
		}
		newCols[i] = v
	}
	out := &ListVec{Data: newCols}
	// copy attrs (names/class/row.names)
	for k, a := range lv.Attrs() {
		out.SetAttr(k, a)
	}
	return out, nil
}
