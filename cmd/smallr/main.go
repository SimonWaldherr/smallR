package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"simonwaldherr.de/go/smallr"
)

// printResult prints a script's captured stdout (if any), then its value —
// except in the non-interactive (-e / file) modes, where the value is only
// shown when nothing was already printed (mirroring plain Rscript behavior).
func printResult(res smallr.EvalResult, alwaysPrintValue bool) {
	hasOutput := strings.TrimSpace(res.Output) != ""
	if hasOutput {
		fmt.Print(res.Output)
	}
	if !hasOutput || alwaysPrintValue {
		fmt.Println(res.Value.String())
	}
}

func main() {
	var expr string
	flag.StringVar(&expr, "e", "", "evaluate expression")
	flag.Parse()

	ctx := smallr.NewContext()

	if expr != "" {
		res, err := ctx.EvalString(expr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		printResult(res, false)
		return
	}

	if flag.NArg() > 0 {
		path := flag.Arg(0)
		b, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		res, err := ctx.EvalString(string(b))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		printResult(res, false)
		return
	}

	// REPL
	fmt.Println("smallR (minimal R-compatible interpreter, Go) — type 'quit' to exit")
	sc := bufio.NewScanner(os.Stdin)
	var buf strings.Builder
	for {
		fmt.Print("> ")
		if !sc.Scan() {
			break
		}
		line := sc.Text()
		if strings.TrimSpace(line) == "quit" {
			break
		}
		// naive multi-line support: continue if braces/parens not balanced
		buf.WriteString(line)
		buf.WriteString("\n")
		src := buf.String()
		if !looksComplete(src) {
			continue
		}
		res, err := ctx.EvalString(src)
		if err != nil {
			fmt.Println("Error:", err)
			buf.Reset()
			continue
		}
		printResult(res, true)
		buf.Reset()
	}
}

func looksComplete(src string) bool {
	// Heuristic: balanced (), {}, []
	var p, b, s int
	inStr := false
	quote := byte(0)
	esc := false
	for i := 0; i < len(src); i++ {
		ch := src[i]
		if inStr {
			if esc {
				esc = false
				continue
			}
			if ch == '\\' {
				esc = true
				continue
			}
			if ch == quote {
				inStr = false
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inStr = true
			quote = ch
			continue
		}
		switch ch {
		case '(':
			p++
		case ')':
			if p > 0 {
				p--
			}
		case '{':
			b++
		case '}':
			if b > 0 {
				b--
			}
		case '[':
			s++
		case ']':
			if s > 0 {
				s--
			}
		}
	}
	return p == 0 && b == 0 && s == 0
}
