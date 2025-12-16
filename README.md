
# smallR

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

