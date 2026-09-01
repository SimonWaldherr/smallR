package smallr

import (
	"container/list"
	"io"
	"sync"

	"simonwaldherr.de/go/smallr/internal/ast"
	"simonwaldherr.de/go/smallr/internal/lexer"
	"simonwaldherr.de/go/smallr/internal/parser"
	"simonwaldherr.de/go/smallr/internal/rt"
	"simonwaldherr.de/go/smallr/internal/token"
)

// Token ist ein Alias für internal/token.Token
type Token = token.Token

// TokenType ist ein Alias für internal/token.Type
type TokenType = token.Type

// Pos ist ein Alias für internal/token.Pos
type Pos = token.Pos

// Lexer ist ein Alias für internal/lexer.Lexer
type Lexer = lexer.Lexer

// NewLexer erstellt einen neuen Lexer für den gegebenen Quelltext.
// Beispiel:
//
//	l := smallr.NewLexer("1 + 2")
func NewLexer(src string) *Lexer {
	return lexer.New(src)
}

// DebugTokens liest alle Tokens aus dem Quelltext (Hilfsfunktion).
func DebugTokens(src string) ([]Token, error) {
	l := lexer.New(src)
	return l.DebugTokens()
}

// Parser ist ein Alias für internal/parser.Parser
type Parser = parser.Parser

// NewParser erstellt einen Parser aus Quelltext.
func NewParser(src string) *Parser {
	return parser.New(src)
}

// ParseProgram parst ein komplettes Programm und gibt das AST zurück.
func ParseProgram(p *Parser) (*ast.Program, error) {
	return p.ParseProgram()
}

// Compile parses src once into a reusable program. Reuse the returned program
// with EvalProgram when the same script is evaluated repeatedly.
func Compile(src string) (*ast.Program, error) {
	return parser.New(src).ParseProgram()
}

// ProgramCache is a bounded, concurrency-safe LRU cache for compiled source.
// It caches only valid programs; cache result values in the embedding
// application, where all data dependencies can be included in the cache key.
type ProgramCache struct {
	mu       sync.Mutex
	capacity int
	entries  map[string]*list.Element
	inflight map[string]*programCacheCall
	order    list.List
	stats    ProgramCacheStats
}

type programCacheEntry struct {
	source  string
	program *Program
}

type programCacheCall struct {
	done    chan struct{}
	program *Program
	err     error
}

// ProgramCacheStats describes cache activity since construction or Clear.
// A waiter is a concurrent caller that reused an in-progress compilation.
type ProgramCacheStats struct {
	Hits      uint64
	Misses    uint64
	Waiters   uint64
	Evictions uint64
}

// NewProgramCache creates a cache that retains up to capacity compiled
// programs. A non-positive capacity disables retention while keeping the same
// Compile API.
func NewProgramCache(capacity int) *ProgramCache {
	return &ProgramCache{
		capacity: capacity,
		entries:  make(map[string]*list.Element),
		inflight: make(map[string]*programCacheCall),
	}
}

// Compile returns a cached program for src when available, otherwise parses
// it and stores the successful result. Returned programs are safe to evaluate
// concurrently with separate contexts; callers must not mutate the AST.
func (c *ProgramCache) Compile(src string) (*Program, error) {
	c.mu.Lock()
	if entry, ok := c.entries[src]; ok {
		c.order.MoveToFront(entry)
		c.stats.Hits++
		program := entry.Value.(*programCacheEntry).program
		c.mu.Unlock()
		return program, nil
	}
	if call, ok := c.inflight[src]; ok {
		c.stats.Waiters++
		c.mu.Unlock()
		<-call.done
		return call.program, call.err
	}
	call := &programCacheCall{done: make(chan struct{})}
	c.inflight[src] = call
	c.stats.Misses++
	c.mu.Unlock()

	program, err := Compile(src)

	c.mu.Lock()
	if err == nil && c.capacity > 0 {
		entry := c.order.PushFront(&programCacheEntry{source: src, program: program})
		c.entries[src] = entry
		if c.order.Len() > c.capacity {
			last := c.order.Back()
			delete(c.entries, last.Value.(*programCacheEntry).source)
			c.order.Remove(last)
			c.stats.Evictions++
		}
	}
	delete(c.inflight, src)
	call.program = program
	call.err = err
	close(call.done)
	c.mu.Unlock()
	return program, err
}

// Len reports the number of programs currently retained in the cache.
func (c *ProgramCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}

// Stats returns a consistent snapshot of cache activity.
func (c *ProgramCache) Stats() ProgramCacheStats {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stats
}

// Clear removes retained programs and resets the cache statistics. Ongoing
// compilations are allowed to finish and may populate the cache afterwards.
func (c *ProgramCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*list.Element)
	c.order.Init()
	c.stats = ProgramCacheStats{}
}

// Engine provides isolated, concurrent-safe evaluations for embedding in
// services. It reuses contexts internally and compiles source through its
// bounded program cache.
//
// Values passed as inputs must not be mutated by the caller until Eval
// returns. Engine itself does not retain the input map after evaluation.
type Engine struct {
	limits   ExecutionLimits
	programs *ProgramCache
	contexts sync.Pool
}

// NewEngine creates an isolated evaluator with a bounded source cache. The
// supplied limits apply to every evaluation; choose explicit, tight limits
// when source originates from users. A non-positive cache capacity disables
// retention while still coalescing simultaneous compiles.
func NewEngine(limits ExecutionLimits, programCacheCapacity int) *Engine {
	e := &Engine{
		limits:   limits,
		programs: NewProgramCache(programCacheCapacity),
	}
	e.contexts.New = func() any { return NewContextWithLimits(limits) }
	return e
}

// Eval compiles src through the engine's cache and evaluates it with isolated
// input variables. It is safe to call concurrently.
func (e *Engine) Eval(src string, inputs map[string]Value) (EvalResult, error) {
	program, err := e.programs.Compile(src)
	if err != nil {
		return EvalResult{}, err
	}
	return e.EvalProgram(program, inputs)
}

// EvalProgram evaluates a precompiled program with isolated input variables.
// It is safe to call concurrently.
func (e *Engine) EvalProgram(program *Program, inputs map[string]Value) (EvalResult, error) {
	ctx := e.contexts.Get().(*Context)
	defer func() {
		ctx.Reset()
		e.contexts.Put(ctx)
	}()
	for name, value := range inputs {
		ctx.Global.SetLocal(name, value)
	}
	return ctx.EvalProgram(program)
}

// Limits reports the fixed execution policy of this engine.
func (e *Engine) Limits() ExecutionLimits { return e.limits }

// ProgramCacheStats reports activity of the engine's program cache.
func (e *Engine) ProgramCacheStats() ProgramCacheStats { return e.programs.Stats() }

// ClearProgramCache removes compiled source retained by this engine.
func (e *Engine) ClearProgramCache() { e.programs.Clear() }

// Program und Expr sind Aliase für die AST-Typen
type Program = ast.Program
type Expr = ast.Expr

// Context ist ein Alias für internal/rt.Context
type Context = rt.Context

// NewContext erstellt einen neuen Auswertungskontext mit Standard-Builtins.
func NewContext() *Context { return rt.NewContext() }

// NewContextWithOutput erstellt einen Kontext mit einem benutzerdefinierten Writer.
func NewContextWithOutput(w io.Writer) *Context { return rt.NewContextWithOutput(w) }

// NewContextWithLimits erstellt einen Kontext mit einer passenden
// Standard-Ausführungsrichtlinie für eingebetteten Code.
func NewContextWithLimits(limits ExecutionLimits) *Context {
	return rt.NewContextWithLimits(limits)
}

// EvalResult ist ein Alias für internal/rt.EvalResult
type EvalResult = rt.EvalResult

// Env, Value sind Aliase für die entsprechenden Laufzeit-Typen
type Env = rt.Env
type Value = rt.Value

// Concrete values and element types are exposed for integrations that need
// direct access to a smallR result. Prefer the constructors below for inputs.
type LogicalVec = rt.LogicalVec
type IntVec = rt.IntVec
type DoubleVec = rt.DoubleVec
type CharVec = rt.CharVec
type ListVec = rt.ListVec
type LogicalElem = rt.LogicalElem
type IntElem = rt.IntElem
type FloatElem = rt.FloatElem
type StringElem = rt.StringElem

var NullValue Value = rt.NullValue

// Logical creates a logical vector from Go booleans.
func Logical(values ...bool) *LogicalVec { return rt.Logical(values...) }

// Int creates a scalar integer value.
func Int(value int64) *IntVec { return rt.IntScalar(value) }

// Ints creates an integer vector from Go int64 values.
func Ints(values ...int64) *IntVec { return rt.Integers(values...) }

// Float creates a scalar double value.
func Float(value float64) *DoubleVec { return rt.DoubleScalar(value) }

// Floats creates a double vector from Go float64 values.
func Floats(values ...float64) *DoubleVec { return rt.Doubles(values...) }

// Char creates a scalar character value.
func Char(value string) *CharVec { return rt.CharScalar(value) }

// Chars creates a character vector from Go strings.
func Chars(values ...string) *CharVec { return rt.Chars(values...) }

// List creates a smallR list from existing values.
func List(values ...Value) *ListVec { return rt.List(values...) }

// ToJSON converts a smallR value to a JSON document suitable for HTTP or JS
// responses. Missing values and NULL become JSON null.
func ToJSON(value Value) string { return rt.ToJSON(value) }

// Eval wertet einen AST-Knoten im gegebenen Kontext und Environment aus.
func Eval(ctx *Context, env *Env, expr Expr) (Value, error) {
	return rt.Eval(ctx, env, expr)
}

// ExecutionLimits bounds one evaluation. Keep the default limits enabled for
// untrusted input; zero disables the respective limit.
type ExecutionLimits = rt.ExecutionLimits

var (
	DefaultExecutionLimits = rt.DefaultExecutionLimits
	ErrExecutionStepLimit  = rt.ErrExecutionStepLimit
	ErrExecutionCallDepth  = rt.ErrExecutionCallDepth
	ErrExecutionTimeout    = rt.ErrExecutionTimeout
)

// EvalWithLimits evaluates an already-parsed expression under the supplied
// limits without modifying the Context's defaults.
func EvalWithLimits(ctx *Context, env *Env, expr Expr, limits ExecutionLimits) (Value, error) {
	return rt.EvalWithLimits(ctx, env, expr, limits)
}

// EvalString ist eine praktische Kombination aus Parser + Eval,
// die einen Quelltext direkt auswertet (Ausgabe wird getee'd).
func EvalString(ctx *Context, src string) (EvalResult, error) {
	return ctx.EvalString(src)
}

// EvalStringWithLimits evaluates source under the supplied limits without
// modifying the Context's defaults.
func EvalStringWithLimits(ctx *Context, src string, limits ExecutionLimits) (EvalResult, error) {
	return ctx.EvalStringWithLimits(src, limits)
}

// EvalProgram wertet ein vorab kompiliertes Programm mit den Standardlimits
// des Contexts aus.
func EvalProgram(ctx *Context, program *Program) (EvalResult, error) {
	return ctx.EvalProgram(program)
}

// EvalProgramWithLimits wertet ein vorab kompiliertes Programm mit einer
// Ausführungsrichtlinie nur für diesen Lauf aus.
func EvalProgramWithLimits(ctx *Context, program *Program, limits ExecutionLimits) (EvalResult, error) {
	return ctx.EvalProgramWithLimits(program, limits)
}
