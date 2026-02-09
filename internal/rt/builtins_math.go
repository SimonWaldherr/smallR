package rt

import (
	"fmt"
	"math"
)

func installMathBuiltins(env *Env) {
	builtins := map[string]*BuiltinFunc{
		"abs":     {FnName: "abs", Impl: builtinAbs},
		"sqrt":    {FnName: "sqrt", Impl: builtinSqrt},
		"floor":   {FnName: "floor", Impl: builtinFloor},
		"ceiling": {FnName: "ceiling", Impl: builtinCeiling},
		"round":   {FnName: "round", Impl: builtinRound},
		"trunc":   {FnName: "trunc", Impl: builtinTrunc},
		"log":     {FnName: "log", Impl: builtinLog},
		"log2":    {FnName: "log2", Impl: builtinLog2},
		"log10":   {FnName: "log10", Impl: builtinLog10},
		"exp":     {FnName: "exp", Impl: builtinExp},
		"sin":     {FnName: "sin", Impl: builtinSin},
		"cos":     {FnName: "cos", Impl: builtinCos},
		"tan":     {FnName: "tan", Impl: builtinTan},
		"asin":    {FnName: "asin", Impl: builtinAsin},
		"acos":    {FnName: "acos", Impl: builtinAcos},
		"atan":    {FnName: "atan", Impl: builtinAtan},
		"atan2":   {FnName: "atan2", Impl: builtinAtan2},
		"sign":    {FnName: "sign", Impl: builtinSign},
		"max":     {FnName: "max", Impl: builtinMax},
		"min":     {FnName: "min", Impl: builtinMin},
		"range":   {FnName: "range", Impl: builtinRange},
		"cumsum":  {FnName: "cumsum", Impl: builtinCumsum},
		"cumprod": {FnName: "cumprod", Impl: builtinCumprod},
		"cummax":  {FnName: "cummax", Impl: builtinCummax},
		"cummin":  {FnName: "cummin", Impl: builtinCummin},
		"prod":    {FnName: "prod", Impl: builtinProd},
		"diff":    {FnName: "diff", Impl: builtinDiff},
	}
	for name, fn := range builtins {
		env.SetLocal(name, fn)
	}

	// Constants
	env.SetLocal("pi", DoubleScalar(math.Pi))
	env.SetLocal("Inf", DoubleScalar(math.Inf(1)))
	env.SetLocal("NaN", DoubleScalar(math.NaN()))
	env.SetLocal("T", LogicalScalar(true))
	env.SetLocal("F", LogicalScalar(false))
	env.SetLocal("LETTERS", makeLetters(true))
	env.SetLocal("letters", makeLetters(false))
}

func makeLetters(upper bool) *CharVec {
	data := make([]StringElem, 26)
	base := byte('a')
	if upper {
		base = byte('A')
	}
	for i := 0; i < 26; i++ {
		data[i] = StringElem{Val: string(rune(base + byte(i)))}
	}
	return &CharVec{Data: data}
}

// --- Vectorized unary math functions ---

func vecMathUnary(ctx *Context, args []ArgValue, name string, fn func(float64) float64) (Value, error) {
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
	out := make([]FloatElem, len(dv))
	for i, e := range dv {
		if e.NA {
			out[i] = FloatElem{NA: true}
		} else {
			out[i] = FloatElem{Val: fn(e.Val)}
		}
	}
	return &DoubleVec{Data: out}, nil
}

func builtinAbs(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "abs", math.Abs)
}

func builtinSqrt(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "sqrt", math.Sqrt)
}

func builtinFloor(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "floor", math.Floor)
}

func builtinCeiling(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "ceiling", math.Ceil)
}

func builtinTrunc(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "trunc", math.Trunc)
}

func builtinRound(ctx *Context, args []ArgValue) (Value, error) {
	// round(x, digits=0)
	if len(args) < 1 {
		return nil, fmt.Errorf("round(x) expects at least 1 argument")
	}
	digits := 0
	if v, ok := getNamed(args, "digits"); ok {
		fv, err := Force(ctx, v)
		if err != nil {
			return nil, err
		}
		fe, err := asFloatElem(ctx, fv)
		if err != nil {
			return nil, err
		}
		if !fe.NA {
			digits = int(fe.Val)
		}
	} else if len(args) >= 2 && args[1].Name == "" {
		fv, err := Force(ctx, args[1].Val)
		if err != nil {
			return nil, err
		}
		fe, err := asFloatElem(ctx, fv)
		if err != nil {
			return nil, err
		}
		if !fe.NA {
			digits = int(fe.Val)
		}
	}

	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	mult := math.Pow(10, float64(digits))
	out := make([]FloatElem, len(dv))
	for i, e := range dv {
		if e.NA {
			out[i] = FloatElem{NA: true}
		} else {
			out[i] = FloatElem{Val: math.Round(e.Val*mult) / mult}
		}
	}
	return &DoubleVec{Data: out}, nil
}

func builtinLog(ctx *Context, args []ArgValue) (Value, error) {
	// log(x, base=exp(1))
	if len(args) < 1 {
		return nil, fmt.Errorf("log(x) expects at least 1 argument")
	}
	base := math.E
	if v, ok := getNamed(args, "base"); ok {
		fv, err := Force(ctx, v)
		if err != nil {
			return nil, err
		}
		fe, err := asFloatElem(ctx, fv)
		if err != nil {
			return nil, err
		}
		if !fe.NA {
			base = fe.Val
		}
	} else if len(args) >= 2 && args[1].Name == "" {
		fv, err := Force(ctx, args[1].Val)
		if err != nil {
			return nil, err
		}
		fe, err := asFloatElem(ctx, fv)
		if err != nil {
			return nil, err
		}
		if !fe.NA {
			base = fe.Val
		}
	}

	fn := math.Log
	if base != math.E {
		logBase := math.Log(base)
		fn = func(x float64) float64 { return math.Log(x) / logBase }
	}
	return vecMathUnary(ctx, args[:1], "log", fn)
}

func builtinLog2(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "log2", math.Log2)
}

func builtinLog10(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "log10", math.Log10)
}

func builtinExp(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "exp", math.Exp)
}

func builtinSin(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "sin", math.Sin)
}

func builtinCos(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "cos", math.Cos)
}

func builtinTan(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "tan", math.Tan)
}

func builtinAsin(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "asin", math.Asin)
}

func builtinAcos(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "acos", math.Acos)
}

func builtinAtan(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "atan", math.Atan)
}

func builtinAtan2(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("atan2(y, x) expects 2 arguments")
	}
	y, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	x, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	yv, err := asDoubleVec(ctx, y)
	if err != nil {
		return nil, err
	}
	xv, err := asDoubleVec(ctx, x)
	if err != nil {
		return nil, err
	}
	n := max(len(yv), len(xv))
	out := make([]FloatElem, n)
	for i := 0; i < n; i++ {
		ye := yv[i%len(yv)]
		xe := xv[i%len(xv)]
		if ye.NA || xe.NA {
			out[i] = FloatElem{NA: true}
		} else {
			out[i] = FloatElem{Val: math.Atan2(ye.Val, xe.Val)}
		}
	}
	return &DoubleVec{Data: out}, nil
}

func builtinSign(ctx *Context, args []ArgValue) (Value, error) {
	return vecMathUnary(ctx, args, "sign", func(x float64) float64 {
		if x > 0 {
			return 1
		}
		if x < 0 {
			return -1
		}
		return 0
	})
}

func builtinMax(ctx *Context, args []ArgValue) (Value, error) {
	naRm := false
	if v, ok := getNamed(args, "na.rm"); ok {
		fv, _ := Force(ctx, v)
		b, na, _ := asLogicalScalar(ctx, fv)
		if !na {
			naRm = b
		}
	}
	result := math.Inf(-1)
	anyNA := false
	any := false
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
			any = true
			if e.Val > result {
				result = e.Val
			}
		}
	}
	if anyNA && !naRm {
		return DoubleNA(), nil
	}
	if !any {
		return DoubleScalar(math.Inf(-1)), nil
	}
	return DoubleScalar(result), nil
}

func builtinMin(ctx *Context, args []ArgValue) (Value, error) {
	naRm := false
	if v, ok := getNamed(args, "na.rm"); ok {
		fv, _ := Force(ctx, v)
		b, na, _ := asLogicalScalar(ctx, fv)
		if !na {
			naRm = b
		}
	}
	result := math.Inf(1)
	anyNA := false
	any := false
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
			any = true
			if e.Val < result {
				result = e.Val
			}
		}
	}
	if anyNA && !naRm {
		return DoubleNA(), nil
	}
	if !any {
		return DoubleScalar(math.Inf(1)), nil
	}
	return DoubleScalar(result), nil
}

func builtinRange(ctx *Context, args []ArgValue) (Value, error) {
	naRm := false
	if v, ok := getNamed(args, "na.rm"); ok {
		fv, _ := Force(ctx, v)
		b, na, _ := asLogicalScalar(ctx, fv)
		if !na {
			naRm = b
		}
	}
	minVal := math.Inf(1)
	maxVal := math.Inf(-1)
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
			if e.Val < minVal {
				minVal = e.Val
			}
			if e.Val > maxVal {
				maxVal = e.Val
			}
		}
	}
	if anyNA && !naRm {
		return &DoubleVec{Data: []FloatElem{{NA: true}, {NA: true}}}, nil
	}
	return &DoubleVec{Data: []FloatElem{{Val: minVal}, {Val: maxVal}}}, nil
}

func builtinCumsum(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("cumsum(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	out := make([]FloatElem, len(dv))
	var sum float64
	for i, e := range dv {
		if e.NA {
			for j := i; j < len(dv); j++ {
				out[j] = FloatElem{NA: true}
			}
			break
		}
		sum += e.Val
		out[i] = FloatElem{Val: sum}
	}
	return &DoubleVec{Data: out}, nil
}

func builtinCumprod(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("cumprod(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	out := make([]FloatElem, len(dv))
	prod := 1.0
	for i, e := range dv {
		if e.NA {
			for j := i; j < len(dv); j++ {
				out[j] = FloatElem{NA: true}
			}
			break
		}
		prod *= e.Val
		out[i] = FloatElem{Val: prod}
	}
	return &DoubleVec{Data: out}, nil
}

func builtinCummax(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("cummax(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	out := make([]FloatElem, len(dv))
	curMax := math.Inf(-1)
	for i, e := range dv {
		if e.NA {
			for j := i; j < len(dv); j++ {
				out[j] = FloatElem{NA: true}
			}
			break
		}
		if e.Val > curMax {
			curMax = e.Val
		}
		out[i] = FloatElem{Val: curMax}
	}
	return &DoubleVec{Data: out}, nil
}

func builtinCummin(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("cummin(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	out := make([]FloatElem, len(dv))
	curMin := math.Inf(1)
	for i, e := range dv {
		if e.NA {
			for j := i; j < len(dv); j++ {
				out[j] = FloatElem{NA: true}
			}
			break
		}
		if e.Val < curMin {
			curMin = e.Val
		}
		out[i] = FloatElem{Val: curMin}
	}
	return &DoubleVec{Data: out}, nil
}

func builtinProd(ctx *Context, args []ArgValue) (Value, error) {
	naRm := false
	if v, ok := getNamed(args, "na.rm"); ok {
		fv, _ := Force(ctx, v)
		b, na, _ := asLogicalScalar(ctx, fv)
		if !na {
			naRm = b
		}
	}
	result := 1.0
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
			result *= e.Val
		}
	}
	if anyNA && !naRm {
		return DoubleNA(), nil
	}
	return DoubleScalar(result), nil
}

func builtinDiff(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("diff(x) expects at least 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	dv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	lag := 1
	if len(args) >= 2 {
		lv, err := Force(ctx, args[1].Val)
		if err != nil {
			return nil, err
		}
		fe, err := asFloatElem(ctx, lv)
		if err != nil {
			return nil, err
		}
		if !fe.NA {
			lag = int(fe.Val)
		}
	}
	if lag < 1 || lag >= len(dv) {
		return &DoubleVec{Data: nil}, nil
	}
	out := make([]FloatElem, len(dv)-lag)
	for i := 0; i < len(out); i++ {
		a := dv[i]
		b := dv[i+lag]
		if a.NA || b.NA {
			out[i] = FloatElem{NA: true}
		} else {
			out[i] = FloatElem{Val: b.Val - a.Val}
		}
	}
	return &DoubleVec{Data: out}, nil
}
