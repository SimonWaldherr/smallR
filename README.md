# smallR

[![DOI](https://zenodo.org/badge/1116464179.svg)](https://doi.org/10.5281/zenodo.18652409)
[![Go Version](https://img.shields.io/github/go-mod/go-version/SimonWaldherr/smallR)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/SimonWaldherr/smallR?label=release)](https://github.com/SimonWaldherr/smallR/releases)
[![Stars](https://img.shields.io/github/stars/SimonWaldherr/smallR?style=social)](https://github.com/SimonWaldherr/smallR/stargazers)

**smallR** is a compact, R-inspired interpreter implemented in **Go**, designed to be compilable to **WebAssembly**.

![](https://simonwaldherr.de/gh-pages/smallR.png)

This repository is a working foundation: lexer → parser → AST → evaluator with environments, closures, lazy arguments (promises), vector semantics, subsetting, and a small base of built-in functions.

[play with it online](https://simonwaldherr.github.io/smallR/)

> Note: Full R compatibility is a very large target. This code focuses on a pragmatic core that you can extend.

## Quick start (CLI)

```bash
go run ./cmd/smallr -e "x <- c(1,2,3); sum(x)"
```

Run a script:

```bash
go run ./cmd/smallr examples/intro.R
```

Start the REPL:

```bash
go run ./cmd/smallr
```

## Safe embedding

Every evaluation has safe defaults: at most 2,000,000 AST steps, 2,000 nested
function calls, and four seconds of wall-clock time. This stops endless
`while`/`repeat` loops and runaway recursion. Limit errors can be matched with
`errors.Is`.

Tools that evaluate untrusted snippets should set limits appropriate to their
request budget:

```go
ctx := smallr.NewContext()
res, err := smallr.EvalStringWithLimits(ctx, code, smallr.ExecutionLimits{
    MaxSteps:     50_000,
    MaxCallDepth: 100,
    Timeout:      250 * time.Millisecond,
})
if errors.Is(err, smallr.ErrExecutionStepLimit) ||
   errors.Is(err, smallr.ErrExecutionCallDepth) ||
   errors.Is(err, smallr.ErrExecutionTimeout) {
    // Treat the input as over budget.
}
```

`Context` serializes its public evaluation calls. Do not mutate its public
`Global` or `Output` fields while an evaluation is active. Use a separate
context when evaluations should run in parallel or need isolated variables.

### Fast repeated embedding

Parse stable scripts once with `Compile`, then reuse the resulting program.
This skips lexing and parsing on every request; create one context per
concurrent or isolated request, and set a narrow budget for user input.

```go
program, err := smallr.Compile(`sum(values) / length(values)`)
if err != nil { /* reject invalid script */ }

limits := smallr.ExecutionLimits{
    MaxSteps: 50_000, MaxCallDepth: 100, Timeout: 250 * time.Millisecond,
}
ctx := smallr.NewContextWithLimits(limits)
// Set values in ctx.Global, then reuse program for later requests.
result, err := smallr.EvalProgram(ctx, program)
```

For many distinct scripts, use a bounded cache:

```go
scripts := smallr.NewProgramCache(256)
program, err := scripts.Compile(userRule)
```

`ProgramCache` caches only valid, compiled programs by source. Cache result
values only when the script and all input data are part of the cache key.
Concurrent cache misses for the same source are coalesced into one parse;
`Stats` exposes hits, misses, waiters, and evictions for monitoring.

## WebAssembly build

Build:

```bash
GOOS=js GOARCH=wasm go build -o smallr.wasm ./cmd/smallr-wasm
```

Copy `wasm_exec.js` from your Go installation:

```bash
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

Minimal HTML to run:

```html
<script src="wasm_exec.js"></script>
<script>
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("smallr.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
    console.log(smallrEval("1 + 2"));
  });
</script>
```

## Implemented language features (subset)

- Literals: numbers, strings, TRUE/FALSE, NULL, NA
- Assignment: `<-`, `=`, `<<-`, `->`
- Control flow: `if`, `for`, `while`, `repeat`, `break`, `next`, `return`
- Functions: `function(...) { ... }` with closures + **lazy arguments** (Promises)
- Operators: arithmetic, comparisons, `:` sequence, `&&`/`||` short-circuit, `&`/`|` vectorized
- Subsetting: `[]`, `[[ ]]`, `$` (minimal; list names supported)

Built-ins: `print`, `cat`, `c`, `list`, `length`, `sum`, `mean`, `seq`, `rep`, `typeof`, `class`, `attr`, `attributes`, `names`, `is.na`, `as.*`, `stop`, `warning`, `str`.

## Examples

See `examples/`.
