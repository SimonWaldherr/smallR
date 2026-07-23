package rt

type Env struct {
	parent *Env
	vars   map[string]Value
}

func NewEnv(parent *Env) *Env {
	// vars is left nil and allocated lazily on first write: reading a nil map
	// is safe in Go, and most call frames (e.g. zero-argument functions with
	// no locals) never assign anything, so this skips a map allocation on
	// every single function call.
	return &Env{parent: parent}
}

func (e *Env) Parent() *Env { return e.parent }

func (e *Env) Get(name string) (Value, bool) {
	v, ok := e.vars[name]
	if ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

func (e *Env) GetLocal(name string) (Value, bool) {
	v, ok := e.vars[name]
	return v, ok
}

func (e *Env) SetLocal(name string, v Value) {
	if e.vars == nil {
		e.vars = map[string]Value{}
	}
	e.vars[name] = v
}

func (e *Env) Assign(name string, v Value) {
	// R '<-' assigns in current environment by default.
	e.SetLocal(name, v)
}

func (e *Env) AssignSuper(name string, v Value) {
	// R '<<-' assigns in the parent chain where the symbol exists; otherwise in the global env.
	for env := e.parent; env != nil; env = env.parent {
		if _, ok := env.vars[name]; ok {
			env.vars[name] = v
			return
		}
	}
	// fallback: assign in the topmost env
	top := e
	for top.parent != nil {
		top = top.parent
	}
	top.SetLocal(name, v)
}
