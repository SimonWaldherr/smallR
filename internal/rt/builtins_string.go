package rt

import (
	"fmt"
	"strings"
	"unicode"
)

func installStringBuiltins(env *Env) {
	builtins := map[string]*BuiltinFunc{
		"paste":      {FnName: "paste", Impl: builtinPaste},
		"paste0":     {FnName: "paste0", Impl: builtinPaste0},
		"nchar":      {FnName: "nchar", Impl: builtinNchar},
		"substr":     {FnName: "substr", Impl: builtinSubstr},
		"substring":  {FnName: "substring", Impl: builtinSubstr},
		"toupper":    {FnName: "toupper", Impl: builtinToupper},
		"tolower":    {FnName: "tolower", Impl: builtinTolower},
		"trimws":     {FnName: "trimws", Impl: builtinTrimws},
		"startsWith": {FnName: "startsWith", Impl: builtinStartsWith},
		"endsWith":   {FnName: "endsWith", Impl: builtinEndsWith},
		"grep":       {FnName: "grep", Impl: builtinGrep},
		"grepl":      {FnName: "grepl", Impl: builtinGrepl},
		"sub":        {FnName: "sub", Impl: builtinSub},
		"gsub":       {FnName: "gsub", Impl: builtinGsub},
		"strsplit":   {FnName: "strsplit", Impl: builtinStrsplit},
		"sprintf":    {FnName: "sprintf", Impl: builtinSprintf},
		"format":     {FnName: "format", Impl: builtinFormat},
		"chartr":     {FnName: "chartr", Impl: builtinChartr},
		"strrep":     {FnName: "strrep", Impl: builtinStrrep},
	}
	for name, fn := range builtins {
		env.SetLocal(name, fn)
	}
}

func builtinPaste(ctx *Context, args []ArgValue) (Value, error) {
	return pasteImpl(ctx, args, " ")
}

func builtinPaste0(ctx *Context, args []ArgValue) (Value, error) {
	return pasteImpl(ctx, args, "")
}

func pasteImpl(ctx *Context, args []ArgValue, defaultSep string) (Value, error) {
	sep := defaultSep
	collapse := ""
	hasCollapse := false

	if v, ok := getNamed(args, "sep"); ok {
		fv, err := Force(ctx, v)
		if err != nil {
			return nil, err
		}
		if cv, ok := fv.(*CharVec); ok && cv.Len() >= 1 && !cv.Data[0].NA {
			sep = cv.Data[0].Val
		}
	}
	if v, ok := getNamed(args, "collapse"); ok {
		fv, err := Force(ctx, v)
		if err != nil {
			return nil, err
		}
		if cv, ok := fv.(*CharVec); ok && cv.Len() >= 1 && !cv.Data[0].NA {
			collapse = cv.Data[0].Val
			hasCollapse = true
		}
	}

	// Collect all non-named args as string vectors
	var vecs [][]string
	maxLen := 0
	for _, a := range args {
		if a.Name == "sep" || a.Name == "collapse" {
			continue
		}
		v, err := Force(ctx, a.Val)
		if err != nil {
			return nil, err
		}
		strs := toPlainStrings(v)
		vecs = append(vecs, strs)
		if len(strs) > maxLen {
			maxLen = len(strs)
		}
	}

	if len(vecs) == 0 {
		return CharScalar(""), nil
	}

	// Recycle and paste element-wise
	result := make([]StringElem, maxLen)
	for i := 0; i < maxLen; i++ {
		parts := make([]string, len(vecs))
		for j, vec := range vecs {
			if len(vec) == 0 {
				parts[j] = ""
			} else {
				parts[j] = vec[i%len(vec)]
			}
		}
		result[i] = StringElem{Val: strings.Join(parts, sep)}
	}

	if hasCollapse {
		parts := make([]string, len(result))
		for i, e := range result {
			parts[i] = e.Val
		}
		return CharScalar(strings.Join(parts, collapse)), nil
	}
	return &CharVec{Data: result}, nil
}

func builtinNchar(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("nchar(x) expects at least 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	cv, err := asCharVec(ctx, v)
	if err != nil {
		return nil, err
	}
	out := make([]IntElem, len(cv))
	for i, e := range cv {
		if e.NA {
			out[i] = IntElem{NA: true}
		} else {
			out[i] = IntElem{Val: int64(len([]rune(e.Val)))}
		}
	}
	return &IntVec{Data: out}, nil
}

func builtinSubstr(ctx *Context, args []ArgValue) (Value, error) {
	// substr(x, start, stop)
	if len(args) < 3 {
		return nil, fmt.Errorf("substr(x, start, stop) expects 3 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	cv, err := asCharVec(ctx, x)
	if err != nil {
		return nil, err
	}
	startV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	stopV, err := Force(ctx, args[2].Val)
	if err != nil {
		return nil, err
	}
	startF, err := asFloatElem(ctx, startV)
	if err != nil {
		return nil, err
	}
	stopF, err := asFloatElem(ctx, stopV)
	if err != nil {
		return nil, err
	}
	if startF.NA || stopF.NA {
		return CharNA(), nil
	}
	start := int(startF.Val) - 1 // R is 1-indexed
	stop := int(stopF.Val)

	out := make([]StringElem, len(cv))
	for i, e := range cv {
		if e.NA {
			out[i] = StringElem{NA: true}
			continue
		}
		runes := []rune(e.Val)
		s := start
		t := stop
		if s < 0 {
			s = 0
		}
		if t > len(runes) {
			t = len(runes)
		}
		if s >= t {
			out[i] = StringElem{Val: ""}
		} else {
			out[i] = StringElem{Val: string(runes[s:t])}
		}
	}
	return &CharVec{Data: out}, nil
}

func builtinToupper(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("toupper(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	cv, err := asCharVec(ctx, v)
	if err != nil {
		return nil, err
	}
	out := make([]StringElem, len(cv))
	for i, e := range cv {
		if e.NA {
			out[i] = StringElem{NA: true}
		} else {
			out[i] = StringElem{Val: strings.ToUpper(e.Val)}
		}
	}
	return &CharVec{Data: out}, nil
}

func builtinTolower(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tolower(x) expects 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	cv, err := asCharVec(ctx, v)
	if err != nil {
		return nil, err
	}
	out := make([]StringElem, len(cv))
	for i, e := range cv {
		if e.NA {
			out[i] = StringElem{NA: true}
		} else {
			out[i] = StringElem{Val: strings.ToLower(e.Val)}
		}
	}
	return &CharVec{Data: out}, nil
}

func builtinTrimws(ctx *Context, args []ArgValue) (Value, error) {
	// trimws(x, which = "both")
	if len(args) < 1 {
		return nil, fmt.Errorf("trimws(x) expects at least 1 argument")
	}
	which := "both"
	if v, ok := getNamed(args, "which"); ok {
		fv, err := Force(ctx, v)
		if err != nil {
			return nil, err
		}
		if cv, ok := fv.(*CharVec); ok && cv.Len() >= 1 && !cv.Data[0].NA {
			which = cv.Data[0].Val
		}
	}

	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	cv, err := asCharVec(ctx, v)
	if err != nil {
		return nil, err
	}
	out := make([]StringElem, len(cv))
	for i, e := range cv {
		if e.NA {
			out[i] = StringElem{NA: true}
			continue
		}
		switch which {
		case "left":
			out[i] = StringElem{Val: strings.TrimLeftFunc(e.Val, unicode.IsSpace)}
		case "right":
			out[i] = StringElem{Val: strings.TrimRightFunc(e.Val, unicode.IsSpace)}
		default:
			out[i] = StringElem{Val: strings.TrimSpace(e.Val)}
		}
	}
	return &CharVec{Data: out}, nil
}

func builtinStartsWith(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("startsWith(x, prefix) expects 2 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	prefix, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	xv, err := asCharVec(ctx, x)
	if err != nil {
		return nil, err
	}
	pv, err := asCharVec(ctx, prefix)
	if err != nil {
		return nil, err
	}
	n := max(len(xv), len(pv))
	out := make([]LogicalElem, n)
	for i := 0; i < n; i++ {
		xe := xv[i%len(xv)]
		pe := pv[i%len(pv)]
		if xe.NA || pe.NA {
			out[i] = LogicalElem{NA: true}
		} else {
			out[i] = LogicalElem{Val: strings.HasPrefix(xe.Val, pe.Val)}
		}
	}
	return &LogicalVec{Data: out}, nil
}

func builtinEndsWith(ctx *Context, args []ArgValue) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("endsWith(x, suffix) expects 2 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	suffix, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	xv, err := asCharVec(ctx, x)
	if err != nil {
		return nil, err
	}
	sv, err := asCharVec(ctx, suffix)
	if err != nil {
		return nil, err
	}
	n := max(len(xv), len(sv))
	out := make([]LogicalElem, n)
	for i := 0; i < n; i++ {
		xe := xv[i%len(xv)]
		se := sv[i%len(sv)]
		if xe.NA || se.NA {
			out[i] = LogicalElem{NA: true}
		} else {
			out[i] = LogicalElem{Val: strings.HasSuffix(xe.Val, se.Val)}
		}
	}
	return &LogicalVec{Data: out}, nil
}

func builtinGrep(ctx *Context, args []ArgValue) (Value, error) {
	// grep(pattern, x) — returns indices of matches (simple substring search)
	if len(args) < 2 {
		return nil, fmt.Errorf("grep(pattern, x) expects at least 2 arguments")
	}
	patV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	xV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	patCV, ok := patV.(*CharVec)
	if !ok || patCV.Len() < 1 || patCV.Data[0].NA {
		return nil, fmt.Errorf("grep: invalid pattern")
	}
	pattern := patCV.Data[0].Val
	xCV, err := asCharVec(ctx, xV)
	if err != nil {
		return nil, err
	}

	valueMode := false
	if v, ok := getNamed(args, "value"); ok {
		fv, _ := Force(ctx, v)
		b, na, _ := asLogicalScalar(ctx, fv)
		if !na {
			valueMode = b
		}
	}

	if valueMode {
		var out []StringElem
		for _, e := range xCV {
			if e.NA {
				continue
			}
			if strings.Contains(e.Val, pattern) {
				out = append(out, e)
			}
		}
		if len(out) == 0 {
			return &CharVec{Data: nil}, nil
		}
		return &CharVec{Data: out}, nil
	}

	var out []IntElem
	for i, e := range xCV {
		if e.NA {
			continue
		}
		if strings.Contains(e.Val, pattern) {
			out = append(out, IntElem{Val: int64(i + 1)})
		}
	}
	if len(out) == 0 {
		return &IntVec{Data: nil}, nil
	}
	return &IntVec{Data: out}, nil
}

func builtinGrepl(ctx *Context, args []ArgValue) (Value, error) {
	// grepl(pattern, x) — returns logical vector
	if len(args) < 2 {
		return nil, fmt.Errorf("grepl(pattern, x) expects at least 2 arguments")
	}
	patV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	xV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	patCV, ok := patV.(*CharVec)
	if !ok || patCV.Len() < 1 || patCV.Data[0].NA {
		return nil, fmt.Errorf("grepl: invalid pattern")
	}
	pattern := patCV.Data[0].Val
	xCV, err := asCharVec(ctx, xV)
	if err != nil {
		return nil, err
	}
	out := make([]LogicalElem, len(xCV))
	for i, e := range xCV {
		if e.NA {
			out[i] = LogicalElem{NA: true}
		} else {
			out[i] = LogicalElem{Val: strings.Contains(e.Val, pattern)}
		}
	}
	return &LogicalVec{Data: out}, nil
}

func builtinSub(ctx *Context, args []ArgValue) (Value, error) {
	// sub(pattern, replacement, x) — replace first occurrence
	if len(args) < 3 {
		return nil, fmt.Errorf("sub(pattern, replacement, x) expects 3 arguments")
	}
	patV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	replV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	xV, err := Force(ctx, args[2].Val)
	if err != nil {
		return nil, err
	}
	patS := toPlainStrings(patV)
	replS := toPlainStrings(replV)
	if len(patS) == 0 || len(replS) == 0 {
		return nil, fmt.Errorf("sub: invalid arguments")
	}
	pattern := patS[0]
	replacement := replS[0]

	cv, err := asCharVec(ctx, xV)
	if err != nil {
		return nil, err
	}
	out := make([]StringElem, len(cv))
	for i, e := range cv {
		if e.NA {
			out[i] = StringElem{NA: true}
		} else {
			out[i] = StringElem{Val: strings.Replace(e.Val, pattern, replacement, 1)}
		}
	}
	return &CharVec{Data: out}, nil
}

func builtinGsub(ctx *Context, args []ArgValue) (Value, error) {
	// gsub(pattern, replacement, x) — replace all occurrences
	if len(args) < 3 {
		return nil, fmt.Errorf("gsub(pattern, replacement, x) expects 3 arguments")
	}
	patV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	replV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	xV, err := Force(ctx, args[2].Val)
	if err != nil {
		return nil, err
	}
	patS := toPlainStrings(patV)
	replS := toPlainStrings(replV)
	if len(patS) == 0 || len(replS) == 0 {
		return nil, fmt.Errorf("gsub: invalid arguments")
	}
	pattern := patS[0]
	replacement := replS[0]

	cv, err := asCharVec(ctx, xV)
	if err != nil {
		return nil, err
	}
	out := make([]StringElem, len(cv))
	for i, e := range cv {
		if e.NA {
			out[i] = StringElem{NA: true}
		} else {
			out[i] = StringElem{Val: strings.ReplaceAll(e.Val, pattern, replacement)}
		}
	}
	return &CharVec{Data: out}, nil
}

func builtinStrsplit(ctx *Context, args []ArgValue) (Value, error) {
	// strsplit(x, split) — returns a list of character vectors
	if len(args) < 2 {
		return nil, fmt.Errorf("strsplit(x, split) expects 2 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	splitV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	cv, err := asCharVec(ctx, x)
	if err != nil {
		return nil, err
	}
	splitS := toPlainStrings(splitV)
	if len(splitS) == 0 {
		return nil, fmt.Errorf("strsplit: invalid split")
	}
	split := splitS[0]

	out := make([]Value, len(cv))
	for i, e := range cv {
		if e.NA {
			out[i] = CharNA()
		} else {
			parts := strings.Split(e.Val, split)
			elems := make([]StringElem, len(parts))
			for j, p := range parts {
				elems[j] = StringElem{Val: p}
			}
			out[i] = &CharVec{Data: elems}
		}
	}
	return &ListVec{Data: out}, nil
}

func builtinSprintf(ctx *Context, args []ArgValue) (Value, error) {
	// sprintf(fmt, ...) — simplified implementation
	if len(args) < 1 {
		return nil, fmt.Errorf("sprintf(fmt, ...) expects at least 1 argument")
	}
	fmtV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	fmtCV, ok := fmtV.(*CharVec)
	if !ok || fmtCV.Len() < 1 || fmtCV.Data[0].NA {
		return nil, fmt.Errorf("sprintf: invalid format string")
	}
	fmtStr := fmtCV.Data[0].Val

	// Collect remaining args as interface values for fmt.Sprintf
	var fmtArgs []interface{}
	for _, a := range args[1:] {
		v, err := Force(ctx, a.Val)
		if err != nil {
			return nil, err
		}
		switch t := v.(type) {
		case *DoubleVec:
			if t.Len() == 1 && !t.Data[0].NA {
				fmtArgs = append(fmtArgs, t.Data[0].Val)
			} else {
				fmtArgs = append(fmtArgs, v.String())
			}
		case *IntVec:
			if t.Len() == 1 && !t.Data[0].NA {
				fmtArgs = append(fmtArgs, t.Data[0].Val)
			} else {
				fmtArgs = append(fmtArgs, v.String())
			}
		case *CharVec:
			if t.Len() == 1 && !t.Data[0].NA {
				fmtArgs = append(fmtArgs, t.Data[0].Val)
			} else {
				fmtArgs = append(fmtArgs, v.String())
			}
		case *LogicalVec:
			if t.Len() == 1 && !t.Data[0].NA {
				if t.Data[0].Val {
					fmtArgs = append(fmtArgs, "TRUE")
				} else {
					fmtArgs = append(fmtArgs, "FALSE")
				}
			} else {
				fmtArgs = append(fmtArgs, v.String())
			}
		default:
			fmtArgs = append(fmtArgs, v.String())
		}
	}

	result := fmt.Sprintf(fmtStr, fmtArgs...)
	return CharScalar(result), nil
}

func builtinFormat(ctx *Context, args []ArgValue) (Value, error) {
	// format(x) — convert to character representation
	if len(args) < 1 {
		return nil, fmt.Errorf("format(x) expects at least 1 argument")
	}
	v, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	cv, err := asCharVec(ctx, v)
	if err != nil {
		// fallback
		return CharScalar(v.String()), nil
	}
	return &CharVec{Data: cv}, nil
}

func builtinChartr(ctx *Context, args []ArgValue) (Value, error) {
	// chartr(old, new, x)
	if len(args) != 3 {
		return nil, fmt.Errorf("chartr(old, new, x) expects 3 arguments")
	}
	oldV, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	newV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	xV, err := Force(ctx, args[2].Val)
	if err != nil {
		return nil, err
	}
	oldS := toPlainStrings(oldV)
	newS := toPlainStrings(newV)
	if len(oldS) == 0 || len(newS) == 0 {
		return nil, fmt.Errorf("chartr: invalid arguments")
	}
	oldRunes := []rune(oldS[0])
	newRunes := []rune(newS[0])
	if len(oldRunes) != len(newRunes) {
		return nil, fmt.Errorf("chartr: old and new must have the same length")
	}
	tr := make(map[rune]rune, len(oldRunes))
	for i, r := range oldRunes {
		tr[r] = newRunes[i]
	}

	cv, err := asCharVec(ctx, xV)
	if err != nil {
		return nil, err
	}
	out := make([]StringElem, len(cv))
	for i, e := range cv {
		if e.NA {
			out[i] = StringElem{NA: true}
		} else {
			runes := []rune(e.Val)
			for j, r := range runes {
				if rep, ok := tr[r]; ok {
					runes[j] = rep
				}
			}
			out[i] = StringElem{Val: string(runes)}
		}
	}
	return &CharVec{Data: out}, nil
}

func builtinStrrep(ctx *Context, args []ArgValue) (Value, error) {
	// strrep(x, times)
	if len(args) != 2 {
		return nil, fmt.Errorf("strrep(x, times) expects 2 arguments")
	}
	x, err := Force(ctx, args[0].Val)
	if err != nil {
		return nil, err
	}
	timesV, err := Force(ctx, args[1].Val)
	if err != nil {
		return nil, err
	}
	cv, err := asCharVec(ctx, x)
	if err != nil {
		return nil, err
	}
	tf, err := asFloatElem(ctx, timesV)
	if err != nil {
		return nil, err
	}
	if tf.NA {
		return CharNA(), nil
	}
	times := int(tf.Val)
	if times < 0 {
		return nil, fmt.Errorf("strrep: invalid times argument")
	}
	out := make([]StringElem, len(cv))
	for i, e := range cv {
		if e.NA {
			out[i] = StringElem{NA: true}
		} else {
			out[i] = StringElem{Val: strings.Repeat(e.Val, times)}
		}
	}
	return &CharVec{Data: out}, nil
}
