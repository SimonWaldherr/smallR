package rt

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"

	"simonwaldherr.de/go/smallr/internal/ast"
	"simonwaldherr.de/go/smallr/internal/token"
)

var ErrBreakOutsideLoop = errors.New("break used outside of a loop")
var ErrNextOutsideLoop = errors.New("next used outside of a loop")

type ControlKind int

const (
	ctrlNone ControlKind = iota
	ctrlBreak
	ctrlNext
	ctrlReturn
)

type ControlError struct {
	Kind  ControlKind
	Value Value // only for return
}

func (c *ControlError) Error() string {
	switch c.Kind {
	case ctrlBreak:
		return "break"
	case ctrlNext:
		return "next"
	case ctrlReturn:
		return "return"
	default:
		return "control"
	}
}

func isControl(err error, kind ControlKind) (*ControlError, bool) {
	var ce *ControlError
	if errors.As(err, &ce) && ce.Kind == kind {
		return ce, true
	}
	return nil, false
}

func Eval(ctx *Context, env *Env, expr ast.Expr) (Value, error) {
	switch e := expr.(type) {
	case *ast.Ident:
		v, ok := env.Get(e.Name)
		if !ok {
			return nil, fmt.Errorf("object '%s' not found", e.Name)
		}
		return v, nil
	case *ast.NumberLit:
		if e.IsInt {
			// In R, integer literal uses 1L. We accept plain numbers and treat ints when no '.'/'e'.
			return IntScalar(int64(e.Value)), nil
		}
		return DoubleScalar(e.Value), nil
	case *ast.StringLit:
		return CharScalar(e.Value), nil
	case *ast.BoolLit:
		return LogicalScalar(e.Value), nil
	case *ast.NullLit:
		return NullValue, nil
	case *ast.NALit:
		return LogicalNA(), nil

	case *ast.UnaryExpr:
		x, err := Eval(ctx, env, e.X)
		if err != nil {
			return nil, err
		}
		x, err = Force(ctx, x)
		if err != nil {
			return nil, err
		}
		switch e.Op {
		case token.BANG:
			return unaryNot(ctx, x)
		case token.PLUS:
			return unaryPlus(ctx, x)
		case token.MINUS:
			return unaryMinus(ctx, x)
		default:
			return nil, fmt.Errorf("unsupported unary op %s", e.Op)
		}

	case *ast.BinaryExpr:
		// short-circuit for && and ||
		if e.Op == token.ANDAND || e.Op == token.OROR {
			left, err := Eval(ctx, env, e.Left)
			if err != nil {
				return nil, err
			}
			left, err = Force(ctx, left)
			if err != nil {
				return nil, err
			}
			lb, lna, err := asLogicalScalar(ctx, left)
			if err != nil {
				return nil, err
			}
			if e.Op == token.ANDAND {
				if lna {
					// NA && anything -> NA unless right is FALSE? R does tri-logic; keep NA if right not forced
					// We'll follow a simple rule: if NA and right is FALSE => FALSE else NA.
					right, err := Eval(ctx, env, e.Right)
					if err != nil {
						return nil, err
					}
					right, err = Force(ctx, right)
					if err != nil {
						return nil, err
					}
					rb, rna, err := asLogicalScalar(ctx, right)
					if err != nil {
						return nil, err
					}
					if !rna && rb == false {
						return LogicalScalar(false), nil
					}
					return LogicalNA(), nil
				}
				if !lb {
					return LogicalScalar(false), nil
				}
				right, err := Eval(ctx, env, e.Right)
				if err != nil {
					return nil, err
				}
				right, err = Force(ctx, right)
				if err != nil {
					return nil, err
				}
				rb, rna, err := asLogicalScalar(ctx, right)
				if err != nil {
					return nil, err
				}
				if rna {
					return LogicalNA(), nil
				}
				return LogicalScalar(rb), nil
			}
			// OROR
			if lna {
				right, err := Eval(ctx, env, e.Right)
				if err != nil {
					return nil, err
				}
				right, err = Force(ctx, right)
				if err != nil {
					return nil, err
				}
				rb, rna, err := asLogicalScalar(ctx, right)
				if err != nil {
					return nil, err
				}
				if !rna && rb == true {
					return LogicalScalar(true), nil
				}
				return LogicalNA(), nil
			}
			if lb {
				return LogicalScalar(true), nil
			}
			right, err := Eval(ctx, env, e.Right)
			if err != nil {
				return nil, err
			}
			right, err = Force(ctx, right)
			if err != nil {
				return nil, err
			}
			rb, rna, err := asLogicalScalar(ctx, right)
			if err != nil {
				return nil, err
			}
			if rna {
				return LogicalNA(), nil
			}
			return LogicalScalar(rb), nil
		}

		left, err := Eval(ctx, env, e.Left)
		if err != nil {
			return nil, err
		}
		right, err := Eval(ctx, env, e.Right)
		if err != nil {
			return nil, err
		}
		left, err = Force(ctx, left)
		if err != nil {
			return nil, err
		}
		right, err = Force(ctx, right)
		if err != nil {
			return nil, err
		}
		return evalBinary(ctx, e.Op, left, right)

	case *ast.AssignExpr:
		return evalAssign(ctx, env, e)

	case *ast.BlockExpr:
		var last Value = NullValue
		for _, ex := range e.Exprs {
			v, err := Eval(ctx, env, ex)
			if err != nil {
				return nil, err
			}
			last = v
		}
		return last, nil

	case *ast.IfExpr:
		condV, err := Eval(ctx, env, e.Cond)
		if err != nil {
			return nil, err
		}
		condV, err = Force(ctx, condV)
		if err != nil {
			return nil, err
		}
		b, na, err := asLogicalScalar(ctx, condV)
		if err != nil {
			return nil, err
		}
		if na {
			return nil, fmt.Errorf("missing value where TRUE/FALSE needed")
		}
		if b {
			return Eval(ctx, env, e.Then)
		}
		if e.Else != nil {
			return Eval(ctx, env, e.Else)
		}
		return NullValue, nil

	case *ast.ForExpr:
		seqV, err := Eval(ctx, env, e.Seq)
		if err != nil {
			return nil, err
		}
		seqV, err = Force(ctx, seqV)
		if err != nil {
			return nil, err
		}
		n := seqV.Len()
		var last Value = NullValue
		for i := 0; i < n; i++ {
			// assign loop var as scalar element
			elem, err := vectorElement(ctx, seqV, i)
			if err != nil {
				return nil, err
			}
			env.Assign(e.Var, elem)
			v, err := Eval(ctx, env, e.Body)
			if err != nil {
				if _, ok := isControl(err, ctrlNext); ok {
					continue
				}
				if _, ok := isControl(err, ctrlBreak); ok {
					break
				}
				return nil, err
			}
			last = v
		}
		return last, nil

	case *ast.WhileExpr:
		var last Value = NullValue
		for {
			condV, err := Eval(ctx, env, e.Cond)
			if err != nil {
				return nil, err
			}
			condV, err = Force(ctx, condV)
			if err != nil {
				return nil, err
			}
			b, na, err := asLogicalScalar(ctx, condV)
			if err != nil {
				return nil, err
			}
			if na {
				return nil, fmt.Errorf("missing value where TRUE/FALSE needed")
			}
			if !b {
				break
			}
			v, err := Eval(ctx, env, e.Body)
			if err != nil {
				if _, ok := isControl(err, ctrlNext); ok {
					continue
				}
				if _, ok := isControl(err, ctrlBreak); ok {
					break
				}
				return nil, err
			}
			last = v
		}
		return last, nil

	case *ast.RepeatExpr:
		var last Value = NullValue
		for {
			v, err := Eval(ctx, env, e.Body)
			if err != nil {
				if _, ok := isControl(err, ctrlNext); ok {
					continue
				}
				if _, ok := isControl(err, ctrlBreak); ok {
					break
				}
				return nil, err
			}
			last = v
		}
		return last, nil

	case *ast.BreakExpr:
		return nil, &ControlError{Kind: ctrlBreak}
	case *ast.NextExpr:
		return nil, &ControlError{Kind: ctrlNext}
	case *ast.ReturnExpr:
		var v Value = NullValue
		if e.X != nil {
			var err error
			v, err = Eval(ctx, env, e.X)
			if err != nil {
				return nil, err
			}
			v, err = Force(ctx, v)
			if err != nil {
				return nil, err
			}
		}
		return nil, &ControlError{Kind: ctrlReturn, Value: v}

	case *ast.FuncExpr:
		params := make([]Param, 0, len(e.Params))
		for _, ap := range e.Params {
			params = append(params, Param{Name: ap.Name, Default: ap.Default, Dots: ap.Dots})
		}
		return &ClosureFunc{Params: params, Body: e.Body, Env: env}, nil

	case *ast.CallExpr:
		return evalCall(ctx, env, e)

	case *ast.IndexExpr:
		x, err := Eval(ctx, env, e.X)
		if err != nil {
			return nil, err
		}
		x, err = Force(ctx, x)
		if err != nil {
			return nil, err
		}
		idx, err := Eval(ctx, env, e.Index)
		if err != nil {
			return nil, err
		}
		idx, err = Force(ctx, idx)
		if err != nil {
			return nil, err
		}
		return subset(ctx, x, idx, e.Double)

	case *ast.DollarExpr:
		x, err := Eval(ctx, env, e.X)
		if err != nil {
			return nil, err
		}
		x, err = Force(ctx, x)
		if err != nil {
			return nil, err
		}
		return dollar(ctx, x, e.Name)

	default:
		return nil, fmt.Errorf("unhandled AST node %T", expr)
	}
}

// Force resolves a promise if needed.
func Force(ctx *Context, v Value) (Value, error) {
	if p, ok := v.(*Promise); ok {
		return p.Force(ctx)
	}
	return v, nil
}

func evalAssign(ctx *Context, env *Env, a *ast.AssignExpr) (Value, error) {
	// Special case: right assignment "->"
	if a.Op == token.ASSIGN_RIGHT {
		// value -> name
		val, err := Eval(ctx, env, a.Left)
		if err != nil {
			return nil, err
		}
		val, err = Force(ctx, val)
		if err != nil {
			return nil, err
		}
		id, ok := a.Right.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("invalid right-assignment target")
		}
		env.Assign(id.Name, val)
		return val, nil
	}

	val, err := Eval(ctx, env, a.Right)
	if err != nil {
		return nil, err
	}
	val, err = Force(ctx, val)
	if err != nil {
		return nil, err
	}

	// Ident assignment
	if id, ok := a.Left.(*ast.Ident); ok {
		switch a.Op {
		case token.ASSIGN_LEFT, token.ASSIGN_EQ:
			env.Assign(id.Name, val)
		case token.ASSIGN_SUPER:
			env.AssignSuper(id.Name, val)
		default:
			env.Assign(id.Name, val)
		}
		return val, nil
	}

	// Subset assignment x[i] <- v (only when x is ident)
	if ix, ok := a.Left.(*ast.IndexExpr); ok {
		if xid, ok := ix.X.(*ast.Ident); ok {
			cur, ok := env.Get(xid.Name)
			if !ok {
				return nil, fmt.Errorf("object '%s' not found", xid.Name)
			}
			cur, err = Force(ctx, cur)
			if err != nil {
				return nil, err
			}
			idxV, err := Eval(ctx, env, ix.Index)
			if err != nil {
				return nil, err
			}
			idxV, err = Force(ctx, idxV)
			if err != nil {
				return nil, err
			}
			updated, err := setSubset(ctx, cur, idxV, val, ix.Double)
			if err != nil {
				return nil, err
			}
			env.Assign(xid.Name, updated)
			return val, nil
		}
	}

	// Dollar assignment x$name <- v (only when x is ident)
	if dx, ok := a.Left.(*ast.DollarExpr); ok {
		if xid, ok := dx.X.(*ast.Ident); ok {
			cur, ok := env.Get(xid.Name)
			if !ok {
				return nil, fmt.Errorf("object '%s' not found", xid.Name)
			}
			cur, err = Force(ctx, cur)
			if err != nil {
				return nil, err
			}
			updated, err := setDollar(ctx, cur, dx.Name, val)
			if err != nil {
				return nil, err
			}
			env.Assign(xid.Name, updated)
			return val, nil
		}
	}

	return nil, fmt.Errorf("invalid assignment target")
}

func evalCall(ctx *Context, env *Env, c *ast.CallExpr) (Value, error) {
	// Special forms implemented here (non-evaluating)
	if id, ok := c.Fun.(*ast.Ident); ok {
		switch id.Name {
		case "quote":
			if len(c.Args) != 1 {
				return nil, fmt.Errorf("quote() expects 1 argument")
			}
			return &ExprValue{Expr: c.Args[0].Value}, nil
		case "missing":
			// missing(x) inspects whether argument is missing in current function env.
			if len(c.Args) != 1 {
				return nil, fmt.Errorf("missing() expects 1 argument")
			}
			argExpr := c.Args[0].Value
			argId, ok := argExpr.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("missing() expects a symbol")
			}
			v, ok := env.GetLocal(argId.Name)
			if !ok {
				// not bound => missing
				return LogicalScalar(true), nil
			}
			if v == MissingValue {
				return LogicalScalar(true), nil
			}
			return LogicalScalar(false), nil
		}
	}

	fv, err := Eval(ctx, env, c.Fun)
	if err != nil {
		return nil, err
	}
	fv, err = Force(ctx, fv)
	if err != nil {
		return nil, err
	}
	callable, ok := fv.(Callable)
	if !ok {
		return nil, fmt.Errorf("attempt to apply non-function")
	}

	// Build argument list as promises (lazy), expanding ... if present.
	args := make([]ArgValue, 0, len(c.Args))
	for _, a := range c.Args {
		// Expand ...
		if a.Name == "" {
			if id, ok := a.Value.(*ast.Ident); ok && id.Name == "..." {
				dv, ok := env.Get("...")
				if ok {
					dv, err = Force(ctx, dv)
					if err != nil {
						return nil, err
					}
					if dots, ok := dv.(*Dots); ok {
						args = append(args, dots.Args...)
						continue
					}
				}
			}
		}
		args = append(args, ArgValue{Name: a.Name, Val: &Promise{Expr: a.Value, Env: env}})
	}

	return callable.Call(ctx, env, args)
}

func callClosure(ctx *Context, fn *ClosureFunc, args []ArgValue) (Value, error) {
	callEnv := NewEnv(fn.Env)

	// Parameter table
	type bind struct {
		set bool
		val Value
	}
	binds := make([]bind, len(fn.Params))
	paramIndex := map[string]int{}
	dotsIndex := -1
	for i, p := range fn.Params {
		if p.Dots {
			dotsIndex = i
		} else {
			paramIndex[p.Name] = i
		}
	}

	// Capture dots
	var dotsArgs []ArgValue

	// 1) positional matching for unnamed args
	paramPos := 0
	for _, a := range args {
		if a.Name != "" {
			continue
		}
		// find next non-dots param
		for paramPos < len(fn.Params) && fn.Params[paramPos].Dots {
			paramPos++
		}
		if paramPos < len(fn.Params) {
			binds[paramPos] = bind{set: true, val: a.Val}
			paramPos++
		} else {
			if dotsIndex >= 0 {
				dotsArgs = append(dotsArgs, a)
			} else {
				return nil, fmt.Errorf("unused argument (positional)")
			}
		}
	}

	// 2) named matching
	for _, a := range args {
		if a.Name == "" {
			continue
		}
		if idx, ok := paramIndex[a.Name]; ok {
			if binds[idx].set {
				return nil, fmt.Errorf("formal argument '%s' matched by multiple actual arguments", a.Name)
			}
			binds[idx] = bind{set: true, val: a.Val}
		} else {
			if dotsIndex >= 0 {
				dotsArgs = append(dotsArgs, a)
			} else {
				return nil, fmt.Errorf("unused argument '%s'", a.Name)
			}
		}
	}

	// 3) bind parameters into callEnv
	for i, p := range fn.Params {
		if p.Dots {
			continue
		}
		if binds[i].set {
			callEnv.SetLocal(p.Name, binds[i].val)
			continue
		}
		if p.Default != nil {
			// default is a promise evaluated in callEnv
			callEnv.SetLocal(p.Name, &Promise{Expr: p.Default, Env: callEnv})
			continue
		}
		callEnv.SetLocal(p.Name, MissingValue)
	}

	// 4) bind dots
	if dotsIndex >= 0 {
		callEnv.SetLocal("...", &Dots{Args: dotsArgs})
	}

	// Execute body
	v, err := Eval(ctx, callEnv, fn.Body)
	if err != nil {
		if ce, ok := isControl(err, ctrlReturn); ok {
			return ce.Value, nil
		}
		if _, ok := isControl(err, ctrlBreak); ok {
			return nil, ErrBreakOutsideLoop
		}
		if _, ok := isControl(err, ctrlNext); ok {
			return nil, ErrNextOutsideLoop
		}
		return nil, err
	}
	return v, nil
}

// --- Value helpers ---

func vectorElement(ctx *Context, v Value, i int) (Value, error) {
	v, err := Force(ctx, v)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case *LogicalVec:
		e := t.Data[i]
		if e.NA {
			return LogicalNA(), nil
		}
		return LogicalScalar(e.Val), nil
	case *IntVec:
		e := t.Data[i]
		if e.NA {
			return IntNA(), nil
		}
		return IntScalar(e.Val), nil
	case *DoubleVec:
		e := t.Data[i]
		if e.NA {
			return DoubleNA(), nil
		}
		return DoubleScalar(e.Val), nil
	case *CharVec:
		e := t.Data[i]
		if e.NA {
			return CharNA(), nil
		}
		return CharScalar(e.Val), nil
	case *ListVec:
		return t.Data[i], nil
	default:
		return nil, fmt.Errorf("cannot index type %s", v.Type())
	}
}

func asLogicalScalar(ctx *Context, v Value) (bool, bool, error) {
	v, err := Force(ctx, v)
	if err != nil {
		return false, false, err
	}
	if v.Len() != 1 {
		return false, false, fmt.Errorf("expected scalar logical, got length %d", v.Len())
	}
	switch t := v.(type) {
	case *LogicalVec:
		e := t.Data[0]
		return e.Val, e.NA, nil
	case *IntVec:
		e := t.Data[0]
		if e.NA {
			return false, true, nil
		}
		return e.Val != 0, false, nil
	case *DoubleVec:
		e := t.Data[0]
		if e.NA {
			return false, true, nil
		}
		return e.Val != 0, false, nil
	case *CharVec:
		e := t.Data[0]
		if e.NA {
			return false, true, nil
		}
		switch e.Val {
		case "TRUE", "T", "true", "1":
			return true, false, nil
		case "FALSE", "F", "false", "0":
			return false, false, nil
		default:
			return false, false, fmt.Errorf("cannot coerce '%s' to logical", e.Val)
		}
	default:
		return false, false, fmt.Errorf("cannot coerce %s to logical", v.Type())
	}
}

func asFloatElem(ctx *Context, v Value) (FloatElem, error) {
	v, err := Force(ctx, v)
	if err != nil {
		return FloatElem{}, err
	}
	if v.Len() != 1 {
		return FloatElem{}, fmt.Errorf("expected scalar, got length %d", v.Len())
	}
	switch t := v.(type) {
	case *DoubleVec:
		return t.Data[0], nil
	case *IntVec:
		e := t.Data[0]
		if e.NA {
			return FloatElem{NA: true}, nil
		}
		return FloatElem{Val: float64(e.Val)}, nil
	case *LogicalVec:
		e := t.Data[0]
		if e.NA {
			return FloatElem{NA: true}, nil
		}
		if e.Val {
			return FloatElem{Val: 1}, nil
		}
		return FloatElem{Val: 0}, nil
	case *CharVec:
		e := t.Data[0]
		if e.NA {
			return FloatElem{NA: true}, nil
		}
		f, err := strconv.ParseFloat(e.Val, 64)
		if err != nil {
			return FloatElem{}, fmt.Errorf("cannot coerce '%s' to double", e.Val)
		}
		return FloatElem{Val: f}, nil
	default:
		return FloatElem{}, fmt.Errorf("cannot coerce %s to double", v.Type())
	}
}

func unaryNot(ctx *Context, v Value) (Value, error) {
	if v.Len() != 1 {
		// vectorized not
		lv, err := asLogicalVec(ctx, v)
		if err != nil {
			return nil, err
		}
		out := make([]LogicalElem, len(lv))
		for i, e := range lv {
			if e.NA {
				out[i] = LogicalElem{NA: true}
			} else {
				out[i] = LogicalElem{Val: !e.Val}
			}
		}
		return &LogicalVec{Data: out}, nil
	}
	b, na, err := asLogicalScalar(ctx, v)
	if err != nil {
		return nil, err
	}
	if na {
		return LogicalNA(), nil
	}
	return LogicalScalar(!b), nil
}

func unaryPlus(ctx *Context, v Value) (Value, error) {
	// no-op but forces numeric coercion
	f, err := asFloatElem(ctx, v)
	if err != nil {
		return nil, err
	}
	if f.NA {
		return DoubleNA(), nil
	}
	return DoubleScalar(f.Val), nil
}

func unaryMinus(ctx *Context, v Value) (Value, error) {
	// numeric
	if v.Len() == 1 {
		f, err := asFloatElem(ctx, v)
		if err != nil {
			return nil, err
		}
		if f.NA {
			return DoubleNA(), nil
		}
		return DoubleScalar(-f.Val), nil
	}
	// vectorized
	fv, err := asDoubleVec(ctx, v)
	if err != nil {
		return nil, err
	}
	out := make([]FloatElem, len(fv))
	for i, e := range fv {
		if e.NA {
			out[i] = FloatElem{NA: true}
		} else {
			out[i] = FloatElem{Val: -e.Val}
		}
	}
	return &DoubleVec{Data: out}, nil
}

func asLogicalVec(ctx *Context, v Value) ([]LogicalElem, error) {
	v, err := Force(ctx, v)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case *LogicalVec:
		return t.Data, nil
	case *IntVec:
		out := make([]LogicalElem, len(t.Data))
		for i, e := range t.Data {
			if e.NA {
				out[i] = LogicalElem{NA: true}
			} else {
				out[i] = LogicalElem{Val: e.Val != 0}
			}
		}
		return out, nil
	case *DoubleVec:
		out := make([]LogicalElem, len(t.Data))
		for i, e := range t.Data {
			if e.NA {
				out[i] = LogicalElem{NA: true}
			} else {
				out[i] = LogicalElem{Val: e.Val != 0}
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("cannot coerce %s to logical", v.Type())
	}
}

func asDoubleVec(ctx *Context, v Value) ([]FloatElem, error) {
	v, err := Force(ctx, v)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case *DoubleVec:
		return t.Data, nil
	case *IntVec:
		out := make([]FloatElem, len(t.Data))
		for i, e := range t.Data {
			if e.NA {
				out[i] = FloatElem{NA: true}
			} else {
				out[i] = FloatElem{Val: float64(e.Val)}
			}
		}
		return out, nil
	case *LogicalVec:
		out := make([]FloatElem, len(t.Data))
		for i, e := range t.Data {
			if e.NA {
				out[i] = FloatElem{NA: true}
			} else if e.Val {
				out[i] = FloatElem{Val: 1}
			} else {
				out[i] = FloatElem{Val: 0}
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("cannot coerce %s to double", v.Type())
	}
}

func evalBinary(ctx *Context, op token.Type, a, b Value) (Value, error) {
	switch op {
	case token.PLUS, token.MINUS, token.STAR, token.SLASH, token.CARET, token.MOD, token.INTDIV, token.COLON:
		return evalNumericBinary(ctx, op, a, b)
	case token.LT, token.LTE, token.GT, token.GTE, token.EQ, token.NEQ:
		return evalCompare(ctx, op, a, b)
	case token.AND, token.OR:
		return evalLogicalVector(ctx, op, a, b)
	default:
		return nil, fmt.Errorf("unsupported binary op %s", op)
	}
}

func evalNumericBinary(ctx *Context, op token.Type, a, b Value) (Value, error) {
	// colon operator creates sequence
	if op == token.COLON {
		af, err := asFloatElem(ctx, a)
		if err != nil {
			return nil, err
		}
		bf, err := asFloatElem(ctx, b)
		if err != nil {
			return nil, err
		}
		if af.NA || bf.NA {
			return &DoubleVec{Data: []FloatElem{{NA: true}}}, nil
		}
		start := int64(af.Val)
		end := int64(bf.Val)
		step := int64(1)
		if end < start {
			step = -1
		}
		n := int((end-start)/step) + 1
		out := make([]IntElem, 0, n)
		for x := start; ; x += step {
			out = append(out, IntElem{Val: x})
			if x == end {
				break
			}
		}
		return &IntVec{Data: out}, nil
	}

	// vectorize with recycling
	av, err := asDoubleVec(ctx, a)
	if err != nil {
		return nil, err
	}
	bv, err := asDoubleVec(ctx, b)
	if err != nil {
		return nil, err
	}
	n := max(len(av), len(bv))
	if n == 0 {
		return &DoubleVec{Data: nil}, nil
	}
	out := make([]FloatElem, n)
	for i := 0; i < n; i++ {
		ae := av[i%len(av)]
		be := bv[i%len(bv)]
		if ae.NA || be.NA {
			out[i] = FloatElem{NA: true}
			continue
		}
		switch op {
		case token.PLUS:
			out[i] = FloatElem{Val: ae.Val + be.Val}
		case token.MINUS:
			out[i] = FloatElem{Val: ae.Val - be.Val}
		case token.STAR:
			out[i] = FloatElem{Val: ae.Val * be.Val}
		case token.SLASH:
			out[i] = FloatElem{Val: ae.Val / be.Val}
		case token.CARET:
			out[i] = FloatElem{Val: math.Pow(ae.Val, be.Val)}
		case token.MOD:
			out[i] = FloatElem{Val: math.Mod(ae.Val, be.Val)}
		case token.INTDIV:
			out[i] = FloatElem{Val: math.Floor(ae.Val / be.Val)}
		default:
			return nil, fmt.Errorf("unsupported numeric op %s", op)
		}
	}
	return &DoubleVec{Data: out}, nil
}

func evalCompare(ctx *Context, op token.Type, a, b Value) (Value, error) {
	// For now, compare as doubles if numeric, else strings if character, else logical.
	// Vectorized with recycling.
	// Character comparisons in R are lexicographic.
	// Equality for lists is not implemented.
	// Determine common type.
	a, _ = Force(ctx, a)
	b, _ = Force(ctx, b)

	// If either is character, coerce both to character.
	if a.Type() == "character" || b.Type() == "character" {
		ac, err := asCharVec(ctx, a)
		if err != nil {
			return nil, err
		}
		bc, err := asCharVec(ctx, b)
		if err != nil {
			return nil, err
		}
		n := max(len(ac), len(bc))
		out := make([]LogicalElem, n)
		for i := 0; i < n; i++ {
			ae := ac[i%len(ac)]
			be := bc[i%len(bc)]
			if ae.NA || be.NA {
				out[i] = LogicalElem{NA: true}
				continue
			}
			cmp := 0
			if ae.Val < be.Val {
				cmp = -1
			} else if ae.Val > be.Val {
				cmp = 1
			}
			out[i] = LogicalElem{Val: compareOp(op, cmp, ae.Val == be.Val)}
		}
		return &LogicalVec{Data: out}, nil
	}

	// Numeric
	av, err := asDoubleVec(ctx, a)
	if err != nil {
		// fallback to logical
		lv, err2 := asLogicalVec(ctx, a)
		if err2 != nil {
			return nil, err
		}
		rv, err2 := asLogicalVec(ctx, b)
		if err2 != nil {
			return nil, err2
		}
		n := max(len(lv), len(rv))
		out := make([]LogicalElem, n)
		for i := 0; i < n; i++ {
			ae := lv[i%len(lv)]
			be := rv[i%len(rv)]
			if ae.NA || be.NA {
				out[i] = LogicalElem{NA: true}
				continue
			}
			cmp := 0
			if !ae.Val && be.Val {
				cmp = -1
			} else if ae.Val && !be.Val {
				cmp = 1
			}
			out[i] = LogicalElem{Val: compareOp(op, cmp, ae.Val == be.Val)}
		}
		return &LogicalVec{Data: out}, nil
	}
	bv, err := asDoubleVec(ctx, b)
	if err != nil {
		return nil, err
	}
	n := max(len(av), len(bv))
	out := make([]LogicalElem, n)
	for i := 0; i < n; i++ {
		ae := av[i%len(av)]
		be := bv[i%len(bv)]
		if ae.NA || be.NA {
			out[i] = LogicalElem{NA: true}
			continue
		}
		cmp := 0
		if ae.Val < be.Val {
			cmp = -1
		} else if ae.Val > be.Val {
			cmp = 1
		}
		out[i] = LogicalElem{Val: compareOp(op, cmp, ae.Val == be.Val)}
	}
	return &LogicalVec{Data: out}, nil
}

func compareOp(op token.Type, cmp int, eq bool) bool {
	switch op {
	case token.LT:
		return cmp < 0
	case token.LTE:
		return cmp <= 0
	case token.GT:
		return cmp > 0
	case token.GTE:
		return cmp >= 0
	case token.EQ:
		return eq
	case token.NEQ:
		return !eq
	default:
		return false
	}
}

func evalLogicalVector(ctx *Context, op token.Type, a, b Value) (Value, error) {
	av, err := asLogicalVec(ctx, a)
	if err != nil {
		return nil, err
	}
	bv, err := asLogicalVec(ctx, b)
	if err != nil {
		return nil, err
	}
	n := max(len(av), len(bv))
	out := make([]LogicalElem, n)
	for i := 0; i < n; i++ {
		ae := av[i%len(av)]
		be := bv[i%len(bv)]
		if ae.NA || be.NA {
			// tri-logic: NA & FALSE => FALSE, NA & TRUE => NA, NA | TRUE => TRUE, NA | FALSE => NA
			if op == token.AND {
				if (!ae.NA && ae.Val == false) || (!be.NA && be.Val == false) {
					out[i] = LogicalElem{Val: false}
				} else {
					out[i] = LogicalElem{NA: true}
				}
			} else {
				// OR
				if (!ae.NA && ae.Val == true) || (!be.NA && be.Val == true) {
					out[i] = LogicalElem{Val: true}
				} else {
					out[i] = LogicalElem{NA: true}
				}
			}
			continue
		}
		switch op {
		case token.AND:
			out[i] = LogicalElem{Val: ae.Val && be.Val}
		case token.OR:
			out[i] = LogicalElem{Val: ae.Val || be.Val}
		}
	}
	return &LogicalVec{Data: out}, nil
}

func asCharVec(ctx *Context, v Value) ([]StringElem, error) {
	v, err := Force(ctx, v)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case *CharVec:
		return t.Data, nil
	case *DoubleVec:
		out := make([]StringElem, len(t.Data))
		for i, e := range t.Data {
			if e.NA {
				out[i] = StringElem{NA: true}
			} else {
				out[i] = StringElem{Val: strconv.FormatFloat(e.Val, 'g', -1, 64)}
			}
		}
		return out, nil
	case *IntVec:
		out := make([]StringElem, len(t.Data))
		for i, e := range t.Data {
			if e.NA {
				out[i] = StringElem{NA: true}
			} else {
				out[i] = StringElem{Val: strconv.FormatInt(e.Val, 10)}
			}
		}
		return out, nil
	case *LogicalVec:
		out := make([]StringElem, len(t.Data))
		for i, e := range t.Data {
			if e.NA {
				out[i] = StringElem{NA: true}
			} else if e.Val {
				out[i] = StringElem{Val: "TRUE"}
			} else {
				out[i] = StringElem{Val: "FALSE"}
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("cannot coerce %s to character", v.Type())
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// --- Subsetting ---

func subset(ctx *Context, x Value, idx Value, dbl bool) (Value, error) {
	x, err := Force(ctx, x)
	if err != nil {
		return nil, err
	}
	idx, err = Force(ctx, idx)
	if err != nil {
		return nil, err
	}

	// [[ expects scalar index
	if dbl {
		switch t := idx.(type) {
		case *IntVec:
			if t.Len() != 1 {
				return nil, fmt.Errorf("[[ expects a single index")
			}
			e := t.Data[0]
			if e.NA {
				return nil, fmt.Errorf("subscript out of bounds")
			}
			i := int(e.Val) - 1
			if i < 0 {
				return nil, fmt.Errorf("subscript out of bounds")
			}
			switch xv := x.(type) {
			case *ListVec:
				if i >= len(xv.Data) {
					return nil, fmt.Errorf("subscript out of bounds")
				}
				return xv.Data[i], nil
			default:
				// atomic: return scalar element
				if i >= x.Len() {
					return nil, fmt.Errorf("subscript out of bounds")
				}
				return vectorElement(ctx, x, i)
			}
		case *CharVec:
			if t.Len() != 1 {
				return nil, fmt.Errorf("[[ expects a single index")
			}
			e := t.Data[0]
			if e.NA {
				return nil, fmt.Errorf("subscript out of bounds")
			}
			name := e.Val
			return dollar(ctx, x, name)
		default:
			return nil, fmt.Errorf("invalid subscript type %s", idx.Type())
		}
	}

	// [ vectorized
	switch xv := x.(type) {
	case *ListVec:
		indices, naMask, err := normalizeIndex(ctx, idx, xv.Len())
		if err != nil {
			return nil, err
		}
		out := make([]Value, 0, len(indices))
		for j, i := range indices {
			if naMask[j] {
				out = append(out, NullValue)
				continue
			}
			if i < 0 || i >= xv.Len() {
				out = append(out, NullValue)
			} else {
				out = append(out, xv.Data[i])
			}
		}
		return &ListVec{Data: out}, nil

	case *LogicalVec:
		return subsetAtomicLogical(ctx, xv, idx)
	case *IntVec:
		return subsetAtomicInt(ctx, xv, idx)
	case *DoubleVec:
		return subsetAtomicDouble(ctx, xv, idx)
	case *CharVec:
		return subsetAtomicChar(ctx, xv, idx)
	default:
		return nil, fmt.Errorf("object of type '%s' is not subsettable", x.Type())
	}
}

func dollar(ctx *Context, x Value, name string) (Value, error) {
	x, err := Force(ctx, x)
	if err != nil {
		return nil, err
	}
	// For lists, use names attribute if present.
	switch xv := x.(type) {
	case *ListVec:
		if nm, ok := xv.GetAttr("names"); ok {
			nmv, err := Force(ctx, nm)
			if err != nil {
				return nil, err
			}
			if cn, ok := nmv.(*CharVec); ok && cn.Len() == xv.Len() {
				for i, e := range cn.Data {
					if e.NA {
						continue
					}
					if e.Val == name {
						return xv.Data[i], nil
					}
				}
			}
		}
		// fallback: treat as not found
		return NullValue, nil
	default:
		return nil, fmt.Errorf("$ operator is invalid for atomic vectors")
	}
}

func normalizeIndex(ctx *Context, idx Value, n int) ([]int, []bool, error) {
	idx, err := Force(ctx, idx)
	if err != nil {
		return nil, nil, err
	}
	switch t := idx.(type) {
	case *IntVec:
		// Handle negative indices (exclude) and positive (include)
		hasNeg := false
		hasPos := false
		for _, e := range t.Data {
			if e.NA {
				continue
			}
			if e.Val < 0 {
				hasNeg = true
			}
			if e.Val > 0 {
				hasPos = true
			}
		}
		if hasNeg && hasPos {
			return nil, nil, fmt.Errorf("only 0's may be mixed with negative subscripts")
		}
		if hasNeg {
			exclude := map[int]bool{}
			for _, e := range t.Data {
				if e.NA {
					continue
				}
				if e.Val < 0 {
					exclude[int(-e.Val)-1] = true
				}
			}
			var out []int
			var na []bool
			for i := 0; i < n; i++ {
				if !exclude[i] {
					out = append(out, i)
					na = append(na, false)
				}
			}
			return out, na, nil
		}
		// positive selection, keep NA
		out := make([]int, 0, len(t.Data))
		na := make([]bool, 0, len(t.Data))
		for _, e := range t.Data {
			if e.NA {
				out = append(out, 0)
				na = append(na, true)
				continue
			}
			if e.Val == 0 {
				continue
			}
			out = append(out, int(e.Val)-1)
			na = append(na, false)
		}
		return out, na, nil
	case *DoubleVec:
		// coerce to int
		iv := make([]IntElem, len(t.Data))
		for i, e := range t.Data {
			if e.NA {
				iv[i] = IntElem{NA: true}
			} else {
				iv[i] = IntElem{Val: int64(e.Val)}
			}
		}
		return normalizeIndex(ctx, &IntVec{Data: iv}, n)
	case *LogicalVec:
		// recycle logical index
		if t.Len() == 0 {
			return []int{}, []bool{}, nil
		}
		out := make([]int, 0, n)
		na := make([]bool, 0, n)
		for i := 0; i < n; i++ {
			e := t.Data[i%t.Len()]
			if e.NA {
				out = append(out, i)
				na = append(na, true)
				continue
			}
			if e.Val {
				out = append(out, i)
				na = append(na, false)
			}
		}
		return out, na, nil
	case *CharVec:
		// match against names attr is handled elsewhere; here we return empty
		return nil, nil, fmt.Errorf("character subscripts not supported here")
	default:
		return nil, nil, fmt.Errorf("invalid subscript type %s", idx.Type())
	}
}

func subsetAtomicLogical(ctx *Context, x *LogicalVec, idx Value) (Value, error) {
	indices, naMask, err := normalizeIndex(ctx, idx, x.Len())
	if err != nil {
		// character indexing via names
		if idx.Type() == "character" {
			return subsetByName(ctx, x, idx)
		}
		return nil, err
	}
	out := make([]LogicalElem, 0, len(indices))
	for j, i := range indices {
		if naMask[j] {
			out = append(out, LogicalElem{NA: true})
			continue
		}
		if i < 0 || i >= x.Len() {
			out = append(out, LogicalElem{NA: true})
		} else {
			out = append(out, x.Data[i])
		}
	}
	return &LogicalVec{Data: out}, nil
}

func subsetAtomicInt(ctx *Context, x *IntVec, idx Value) (Value, error) {
	indices, naMask, err := normalizeIndex(ctx, idx, x.Len())
	if err != nil {
		if idx.Type() == "character" {
			return subsetByName(ctx, x, idx)
		}
		return nil, err
	}
	out := make([]IntElem, 0, len(indices))
	for j, i := range indices {
		if naMask[j] {
			out = append(out, IntElem{NA: true})
			continue
		}
		if i < 0 || i >= x.Len() {
			out = append(out, IntElem{NA: true})
		} else {
			out = append(out, x.Data[i])
		}
	}
	return &IntVec{Data: out}, nil
}

func subsetAtomicDouble(ctx *Context, x *DoubleVec, idx Value) (Value, error) {
	indices, naMask, err := normalizeIndex(ctx, idx, x.Len())
	if err != nil {
		if idx.Type() == "character" {
			return subsetByName(ctx, x, idx)
		}
		return nil, err
	}
	out := make([]FloatElem, 0, len(indices))
	for j, i := range indices {
		if naMask[j] {
			out = append(out, FloatElem{NA: true})
			continue
		}
		if i < 0 || i >= x.Len() {
			out = append(out, FloatElem{NA: true})
		} else {
			out = append(out, x.Data[i])
		}
	}
	return &DoubleVec{Data: out}, nil
}

func subsetAtomicChar(ctx *Context, x *CharVec, idx Value) (Value, error) {
	indices, naMask, err := normalizeIndex(ctx, idx, x.Len())
	if err != nil {
		if idx.Type() == "character" {
			return subsetByName(ctx, x, idx)
		}
		return nil, err
	}
	out := make([]StringElem, 0, len(indices))
	for j, i := range indices {
		if naMask[j] {
			out = append(out, StringElem{NA: true})
			continue
		}
		if i < 0 || i >= x.Len() {
			out = append(out, StringElem{NA: true})
		} else {
			out = append(out, x.Data[i])
		}
	}
	return &CharVec{Data: out}, nil
}

func subsetByName(ctx *Context, x Value, idx Value) (Value, error) {
	// idx is character vector; match against names attr.
	nmVal, ok := x.GetAttr("names")
	if !ok {
		// no names => NA result
		cidx, _ := idx.(*CharVec)
		n := cidx.Len()
		return makeNAOfType(x.Type(), n), nil
	}
	nmVal, err := Force(ctx, nmVal)
	if err != nil {
		return nil, err
	}
	nm, ok := nmVal.(*CharVec)
	if !ok {
		return nil, fmt.Errorf("names attribute is not character")
	}
	cidx, ok := idx.(*CharVec)
	if !ok {
		return nil, fmt.Errorf("character subscript expected")
	}
	outIdx := make([]int, cidx.Len())
	naMask := make([]bool, cidx.Len())
	for i, e := range cidx.Data {
		if e.NA {
			naMask[i] = true
			continue
		}
		found := -1
		for j, nme := range nm.Data {
			if nme.NA {
				continue
			}
			if nme.Val == e.Val {
				found = j
				break
			}
		}
		outIdx[i] = found
		if found < 0 {
			naMask[i] = true
		}
	}
	// Build result by mapping indices
	switch xv := x.(type) {
	case *LogicalVec:
		out := make([]LogicalElem, cidx.Len())
		for i := range out {
			if naMask[i] {
				out[i] = LogicalElem{NA: true}
			} else {
				out[i] = xv.Data[outIdx[i]]
			}
		}
		return &LogicalVec{Data: out}, nil
	case *IntVec:
		out := make([]IntElem, cidx.Len())
		for i := range out {
			if naMask[i] {
				out[i] = IntElem{NA: true}
			} else {
				out[i] = xv.Data[outIdx[i]]
			}
		}
		return &IntVec{Data: out}, nil
	case *DoubleVec:
		out := make([]FloatElem, cidx.Len())
		for i := range out {
			if naMask[i] {
				out[i] = FloatElem{NA: true}
			} else {
				out[i] = xv.Data[outIdx[i]]
			}
		}
		return &DoubleVec{Data: out}, nil
	case *CharVec:
		out := make([]StringElem, cidx.Len())
		for i := range out {
			if naMask[i] {
				out[i] = StringElem{NA: true}
			} else {
				out[i] = xv.Data[outIdx[i]]
			}
		}
		return &CharVec{Data: out}, nil
	default:
		return nil, fmt.Errorf("unsupported type for name subset: %s", x.Type())
	}
}

func makeNAOfType(typ string, n int) Value {
	switch typ {
	case "logical":
		out := make([]LogicalElem, n)
		for i := range out {
			out[i] = LogicalElem{NA: true}
		}
		return &LogicalVec{Data: out}
	case "integer":
		out := make([]IntElem, n)
		for i := range out {
			out[i] = IntElem{NA: true}
		}
		return &IntVec{Data: out}
	case "double":
		out := make([]FloatElem, n)
		for i := range out {
			out[i] = FloatElem{NA: true}
		}
		return &DoubleVec{Data: out}
	case "character":
		out := make([]StringElem, n)
		for i := range out {
			out[i] = StringElem{NA: true}
		}
		return &CharVec{Data: out}
	default:
		return &ListVec{Data: make([]Value, n)}
	}
}

func setSubset(ctx *Context, x Value, idx Value, rhs Value, dbl bool) (Value, error) {
	// only support atomic vectors and lists. idx must be integer scalar for [[ and [ for now.
	x, err := Force(ctx, x)
	if err != nil {
		return nil, err
	}
	idx, err = Force(ctx, idx)
	if err != nil {
		return nil, err
	}
	rhs, err = Force(ctx, rhs)
	if err != nil {
		return nil, err
	}

	if dbl {
		// [[ scalar integer
		iv, ok := idx.(*IntVec)
		if !ok || iv.Len() != 1 || iv.Data[0].NA {
			return nil, fmt.Errorf("invalid subscript in [[<-")
		}
		i := int(iv.Data[0].Val) - 1
		if i < 0 {
			return nil, fmt.Errorf("subscript out of bounds")
		}
		switch xv := x.(type) {
		case *ListVec:
			out := cloneList(xv)
			// extend if needed
			for len(out.Data) <= i {
				out.Data = append(out.Data, NullValue)
			}
			out.Data[i] = rhs
			return out, nil
		default:
			return nil, fmt.Errorf("[[<- not implemented for %s", x.Type())
		}
	}

	// [ assignment: support integer positions (no negative) and scalar rhs recycling
	iv, ok := idx.(*IntVec)
	if !ok {
		return nil, fmt.Errorf("[<- only supports integer indices in smallR MVP")
	}
	// Determine positions
	var pos []int
	for _, e := range iv.Data {
		if e.NA {
			continue
		}
		if e.Val <= 0 {
			continue
		}
		pos = append(pos, int(e.Val)-1)
	}
	switch xv := x.(type) {
	case *DoubleVec:
		out := cloneDouble(xv)
		rv, err := asDoubleVec(ctx, rhs)
		if err != nil {
			return nil, err
		}
		if len(rv) == 0 {
			return out, nil
		}
		for i, p := range pos {
			for len(out.Data) <= p {
				out.Data = append(out.Data, FloatElem{NA: true})
			}
			out.Data[p] = rv[i%len(rv)]
		}
		return out, nil
	case *IntVec:
		out := cloneInt(xv)
		// coerce rhs to int
		rv, err := coerceToIntVec(ctx, rhs)
		if err != nil {
			return nil, err
		}
		if len(rv) == 0 {
			return out, nil
		}
		for i, p := range pos {
			for len(out.Data) <= p {
				out.Data = append(out.Data, IntElem{NA: true})
			}
			out.Data[p] = rv[i%len(rv)]
		}
		return out, nil
	case *LogicalVec:
		out := cloneLogical(xv)
		rv, err := asLogicalVec(ctx, rhs)
		if err != nil {
			return nil, err
		}
		if len(rv) == 0 {
			return out, nil
		}
		for i, p := range pos {
			for len(out.Data) <= p {
				out.Data = append(out.Data, LogicalElem{NA: true})
			}
			out.Data[p] = rv[i%len(rv)]
		}
		return out, nil
	case *CharVec:
		out := cloneChar(xv)
		rv, err := asCharVec(ctx, rhs)
		if err != nil {
			return nil, err
		}
		if len(rv) == 0 {
			return out, nil
		}
		for i, p := range pos {
			for len(out.Data) <= p {
				out.Data = append(out.Data, StringElem{NA: true})
			}
			out.Data[p] = rv[i%len(rv)]
		}
		return out, nil
	case *ListVec:
		out := cloneList(xv)
		// rhs can be list or scalar
		var rlist []Value
		if rl, ok := rhs.(*ListVec); ok {
			rlist = rl.Data
		} else {
			rlist = []Value{rhs}
		}
		for i, p := range pos {
			for len(out.Data) <= p {
				out.Data = append(out.Data, NullValue)
			}
			out.Data[p] = rlist[i%len(rlist)]
		}
		return out, nil
	default:
		return nil, fmt.Errorf("[<- not implemented for %s", x.Type())
	}
}

func setDollar(ctx *Context, x Value, name string, rhs Value) (Value, error) {
	x, err := Force(ctx, x)
	if err != nil {
		return nil, err
	}
	rhs, err = Force(ctx, rhs)
	if err != nil {
		return nil, err
	}
	switch xv := x.(type) {
	case *ListVec:
		out := cloneList(xv)
		// get names
		var names []StringElem
		if nm, ok := out.GetAttr("names"); ok {
			nmv, _ := Force(ctx, nm)
			if cn, ok := nmv.(*CharVec); ok {
				names = cn.Data
			}
		}
		if names == nil || len(names) != len(out.Data) {
			names = make([]StringElem, len(out.Data))
			for i := range names {
				names[i] = StringElem{NA: true}
			}
		}
		// find
		found := -1
		for i, e := range names {
			if e.NA {
				continue
			}
			if e.Val == name {
				found = i
				break
			}
		}
		if found >= 0 {
			out.Data[found] = rhs
			out.SetAttr("names", &CharVec{Data: names})
			return out, nil
		}
		// append
		out.Data = append(out.Data, rhs)
		names = append(names, StringElem{Val: name})
		out.SetAttr("names", &CharVec{Data: names})
		return out, nil
	default:
		return nil, fmt.Errorf("$<- not implemented for %s", x.Type())
	}
}

func cloneList(v *ListVec) *ListVec {
	out := &ListVec{Data: append([]Value(nil), v.Data...)}
	// shallow copy attrs
	for k, a := range v.Attrs() {
		out.SetAttr(k, a)
	}
	return out
}
func cloneDouble(v *DoubleVec) *DoubleVec {
	out := &DoubleVec{Data: append([]FloatElem(nil), v.Data...)}
	for k, a := range v.Attrs() {
		out.SetAttr(k, a)
	}
	return out
}
func cloneInt(v *IntVec) *IntVec {
	out := &IntVec{Data: append([]IntElem(nil), v.Data...)}
	for k, a := range v.Attrs() {
		out.SetAttr(k, a)
	}
	return out
}
func cloneLogical(v *LogicalVec) *LogicalVec {
	out := &LogicalVec{Data: append([]LogicalElem(nil), v.Data...)}
	for k, a := range v.Attrs() {
		out.SetAttr(k, a)
	}
	return out
}
func cloneChar(v *CharVec) *CharVec {
	out := &CharVec{Data: append([]StringElem(nil), v.Data...)}
	for k, a := range v.Attrs() {
		out.SetAttr(k, a)
	}
	return out
}

func coerceToIntVec(ctx *Context, v Value) ([]IntElem, error) {
	v, err := Force(ctx, v)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case *IntVec:
		return t.Data, nil
	case *DoubleVec:
		out := make([]IntElem, len(t.Data))
		for i, e := range t.Data {
			if e.NA {
				out[i] = IntElem{NA: true}
			} else {
				out[i] = IntElem{Val: int64(e.Val)}
			}
		}
		return out, nil
	case *LogicalVec:
		out := make([]IntElem, len(t.Data))
		for i, e := range t.Data {
			if e.NA {
				out[i] = IntElem{NA: true}
			} else if e.Val {
				out[i] = IntElem{Val: 1}
			} else {
				out[i] = IntElem{Val: 0}
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("cannot coerce %s to integer", v.Type())
	}
}

// --- IO helpers for print/cat ---

func write(ctx *Context, s string) error {
	if ctx.Output == nil {
		return nil
	}
	_, err := io.WriteString(ctx.Output, s)
	return err
}
