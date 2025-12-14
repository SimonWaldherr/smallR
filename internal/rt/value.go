package rt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"simonwaldherr.de/go/smallr/internal/ast"
)

type Value interface {
	Type() string
	Len() int
	String() string

	Attrs() map[string]Value
	GetAttr(name string) (Value, bool)
	SetAttr(name string, v Value)
}

type Base struct {
	attrs map[string]Value
}

func (b *Base) Attrs() map[string]Value {
	if b.attrs == nil {
		b.attrs = map[string]Value{}
	}
	return b.attrs
}

func (b *Base) GetAttr(name string) (Value, bool) {
	if b.attrs == nil {
		return nil, false
	}
	v, ok := b.attrs[name]
	return v, ok
}

func (b *Base) SetAttr(name string, v Value) {
	if b.attrs == nil {
		b.attrs = map[string]Value{}
	}
	if v == nil {
		delete(b.attrs, name)
		return
	}
	b.attrs[name] = v
}

type Null struct{ Base }

func (n *Null) Type() string { return "null" }
func (n *Null) Len() int     { return 0 }
func (n *Null) String() string {
	return "NULL"
}

var NullValue = &Null{}

type Missing struct{ Base }

func (m *Missing) Type() string { return "missing" }
func (m *Missing) Len() int     { return 0 }
func (m *Missing) String() string {
	return "<missing>"
}

var MissingValue = &Missing{}

type LogicalElem struct {
	Val bool
	NA  bool
}

type IntElem struct {
	Val int64
	NA  bool
}

type FloatElem struct {
	Val float64
	NA  bool
}

type StringElem struct {
	Val string
	NA  bool
}

type LogicalVec struct {
	Base
	Data []LogicalElem
}

func (v *LogicalVec) Type() string { return "logical" }
func (v *LogicalVec) Len() int     { return len(v.Data) }
func (v *LogicalVec) String() string {
	return formatAtomic(func(i int) (string, bool) {
		e := v.Data[i]
		if e.NA {
			return "NA", true
		}
		if e.Val {
			return "TRUE", true
		}
		return "FALSE", true
	}, len(v.Data))
}

type IntVec struct {
	Base
	Data []IntElem
}

func (v *IntVec) Type() string { return "integer" }
func (v *IntVec) Len() int     { return len(v.Data) }
func (v *IntVec) String() string {
	return formatAtomic(func(i int) (string, bool) {
		e := v.Data[i]
		if e.NA {
			return "NA", true
		}
		return strconv.FormatInt(e.Val, 10), true
	}, len(v.Data))
}

type DoubleVec struct {
	Base
	Data []FloatElem
}

func (v *DoubleVec) Type() string { return "double" }
func (v *DoubleVec) Len() int     { return len(v.Data) }
func (v *DoubleVec) String() string {
	return formatAtomic(func(i int) (string, bool) {
		e := v.Data[i]
		if e.NA {
			return "NA", true
		}
		if math.IsNaN(e.Val) {
			return "NaN", true
		}
		if math.IsInf(e.Val, 1) {
			return "Inf", true
		}
		if math.IsInf(e.Val, -1) {
			return "-Inf", true
		}
		// R prints without trailing .0 sometimes; keep compact
		s := strconv.FormatFloat(e.Val, 'g', -1, 64)
		return s, true
	}, len(v.Data))
}

type CharVec struct {
	Base
	Data []StringElem
}

func (v *CharVec) Type() string { return "character" }
func (v *CharVec) Len() int     { return len(v.Data) }
func (v *CharVec) String() string {
	return formatAtomic(func(i int) (string, bool) {
		e := v.Data[i]
		if e.NA {
			return "NA", true
		}
		return fmt.Sprintf("%q", e.Val), true
	}, len(v.Data))
}

type ListVec struct {
	Base
	Data []Value
}

func (v *ListVec) Type() string { return "list" }
func (v *ListVec) Len() int     { return len(v.Data) }
func (v *ListVec) String() string {
	// minimal R-like display
	var parts []string
	for i, e := range v.Data {
		parts = append(parts, fmt.Sprintf("[[%d]] %s", i+1, formatValueForPrint(e)))
	}
	if len(parts) == 0 {
		return "list()"
	}
	return "list(\n  " + strings.Join(parts, ",\n  ") + "\n)"
}

type ExprValue struct {
	Base
	Expr ast.Expr
}

func (e *ExprValue) Type() string { return "expr" }
func (e *ExprValue) Len() int     { return 1 }
func (e *ExprValue) String() string {
	if e.Expr == nil {
		return "expression()"
	}
	return "expression(" + e.Expr.String() + ")"
}

type Promise struct {
	Base
	Expr   ast.Expr
	Env    *Env
	forced bool
	val    Value
	err    error
}

func (p *Promise) Type() string { return "promise" }
func (p *Promise) Len() int     { return 1 }
func (p *Promise) String() string {
	if p.forced {
		return p.val.String()
	}
	return "<promise>"
}

func (p *Promise) Force(ctx *Context) (Value, error) {
	if p.forced {
		return p.val, p.err
	}
	p.forced = true
	v, err := Eval(ctx, p.Env, p.Expr)
	p.val, p.err = v, err
	return v, err
}

type ArgValue struct {
	Name string
	Val  Value
	// Original expression is not kept except via Promise.
}

type Dots struct {
	Base
	Args []ArgValue
}

func (d *Dots) Type() string { return "dots" }
func (d *Dots) Len() int     { return len(d.Args) }
func (d *Dots) String() string {
	return "<...>"
}

type Callable interface {
	Value
	Call(ctx *Context, caller *Env, args []ArgValue) (Value, error)
	Name() string
}

type BuiltinFunc struct {
	Base
	FnName string
	Impl   func(ctx *Context, args []ArgValue) (Value, error)
}

func (b *BuiltinFunc) Type() string { return "function" }
func (b *BuiltinFunc) Len() int     { return 1 }
func (b *BuiltinFunc) String() string {
	return fmt.Sprintf("<builtin:%s>", b.FnName)
}
func (b *BuiltinFunc) Name() string { return b.FnName }
func (b *BuiltinFunc) Call(ctx *Context, caller *Env, args []ArgValue) (Value, error) {
	_ = caller
	return b.Impl(ctx, args)
}

type Param struct {
	Name    string
	Default ast.Expr // may be nil
	Dots    bool
}

type ClosureFunc struct {
	Base
	FnName string // optional
	Params []Param
	Body   ast.Expr
	Env    *Env
}

func (c *ClosureFunc) Type() string { return "function" }
func (c *ClosureFunc) Len() int     { return 1 }
func (c *ClosureFunc) String() string {
	if c.FnName != "" {
		return fmt.Sprintf("<function:%s>", c.FnName)
	}
	return "<function>"
}
func (c *ClosureFunc) Name() string { return c.FnName }
func (c *ClosureFunc) Call(ctx *Context, caller *Env, args []ArgValue) (Value, error) {
	_ = caller
	return callClosure(ctx, c, args)
}

// --- Constructors ---

func LogicalScalar(v bool) *LogicalVec { return &LogicalVec{Data: []LogicalElem{{Val: v}}} }
func LogicalNA() *LogicalVec           { return &LogicalVec{Data: []LogicalElem{{NA: true}}} }

func IntScalar(v int64) *IntVec { return &IntVec{Data: []IntElem{{Val: v}}} }
func IntNA() *IntVec            { return &IntVec{Data: []IntElem{{NA: true}}} }

func DoubleScalar(v float64) *DoubleVec { return &DoubleVec{Data: []FloatElem{{Val: v}}} }
func DoubleNA() *DoubleVec              { return &DoubleVec{Data: []FloatElem{{NA: true}}} }

func CharScalar(v string) *CharVec { return &CharVec{Data: []StringElem{{Val: v}}} }
func CharNA() *CharVec             { return &CharVec{Data: []StringElem{{NA: true}}} }

func List(values ...Value) *ListVec { return &ListVec{Data: values} }

// --- Formatting helpers ---

func formatAtomic(get func(i int) (string, bool), n int) string {
	if n == 0 {
		return "c()"
	}
	if n == 1 {
		s, _ := get(0)
		return s
	}
	var parts []string
	for i := 0; i < n; i++ {
		s, _ := get(i)
		parts = append(parts, s)
	}
	return strings.Join(parts, " ")
}

func formatValueForPrint(v Value) string {
	if v == nil {
		return "<nil>"
	}
	// force promises for printing
	if p, ok := v.(*Promise); ok {
		// cannot force without ctx here; show promise
		_ = p
	}
	return v.String()
}

func listNames(l *ListVec) ([]string, bool) {
	nv, ok := l.GetAttr("names")
	if !ok || nv == nil {
		return nil, false
	}
	cv, ok := nv.(*CharVec)
	if !ok {
		return nil, false
	}
	names := make([]string, 0, len(cv.Data))
	for _, e := range cv.Data {
		if e.NA {
			names = append(names, "")
		} else {
			names = append(names, e.Val)
		}
	}
	return names, true
}

func ToJSON(v Value) string {
	obj := toJSONValue(v)
	b, _ := json.Marshal(obj)
	return string(b)
}

func toJSONValue(v Value) any {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case *Null:
		return nil
	case *LogicalVec:
		if t.Len() == 1 {
			e := t.Data[0]
			if e.NA {
				return nil
			}
			return e.Val
		}
		arr := make([]any, 0, t.Len())
		for _, e := range t.Data {
			if e.NA {
				arr = append(arr, nil)
			} else {
				arr = append(arr, e.Val)
			}
		}
		return arr
	case *IntVec:
		if t.Len() == 1 {
			e := t.Data[0]
			if e.NA {
				return nil
			}
			return e.Val
		}
		arr := make([]any, 0, t.Len())
		for _, e := range t.Data {
			if e.NA {
				arr = append(arr, nil)
			} else {
				arr = append(arr, e.Val)
			}
		}
		return arr
	case *DoubleVec:
		if t.Len() == 1 {
			e := t.Data[0]
			if e.NA {
				return nil
			}
			return e.Val
		}
		arr := make([]any, 0, t.Len())
		for _, e := range t.Data {
			if e.NA {
				arr = append(arr, nil)
			} else {
				arr = append(arr, e.Val)
			}
		}
		return arr
	case *CharVec:
		if t.Len() == 1 {
			e := t.Data[0]
			if e.NA {
				return nil
			}
			return e.Val
		}
		arr := make([]any, 0, t.Len())
		for _, e := range t.Data {
			if e.NA {
				arr = append(arr, nil)
			} else {
				arr = append(arr, e.Val)
			}
		}
		return arr
	case *ListVec:
		// If list has non-empty, unique names, encode as a JSON object for convenient JS interop.
		if names, ok := listNames(t); ok {
			allNonEmpty := true
			seen := map[string]bool{}
			unique := true
			for _, n := range names {
				if n == "" {
					allNonEmpty = false
					break
				}
				if seen[n] {
					unique = false
					break
				}
				seen[n] = true
			}
			if allNonEmpty && unique && len(names) == t.Len() {
				obj := map[string]any{}
				for i, n := range names {
					obj[n] = toJSONValue(t.Data[i])
				}
				return obj
			}
			// otherwise fall back to an array
		}
		arr := make([]any, 0, t.Len())
		for _, e := range t.Data {
			arr = append(arr, toJSONValue(e))
		}
		return arr
	case *ExprValue:
		if t.Expr == nil {
			return "expression()"
		}
		return t.Expr.String()
	case *Promise:
		// do not force for JSON
		if t.forced {
			return toJSONValue(t.val)
		}
		return "<promise>"
	case *BuiltinFunc, *ClosureFunc:
		return t.String()
	case *Dots:
		m := make(map[string]any)
		for i, a := range t.Args {
			key := a.Name
			if key == "" {
				key = fmt.Sprintf("..%d", i+1)
			}
			m[key] = toJSONValue(a.Val)
		}
		return m
	default:
		return v.String()
	}
}

func DebugValue(v Value) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Type=%s Len=%d Value=%s", v.Type(), v.Len(), v.String())
	if attrs := v.Attrs(); len(attrs) > 0 {
		fmt.Fprintf(&buf, " Attrs=[")
		first := true
		for k, vv := range attrs {
			if !first {
				fmt.Fprintf(&buf, ", ")
			}
			first = false
			fmt.Fprintf(&buf, "%s=%s", k, vv.String())
		}
		fmt.Fprintf(&buf, "]")
	}
	return buf.String()
}
