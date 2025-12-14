package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"simonwaldherr.de/go/smallr/internal/rt"
)

func main() {
	var expr string
	flag.StringVar(&expr, "e", "", "evaluate expression")
	flag.Parse()

	ctx := rt.NewContext()

	if expr != "" {
		res, err := ctx.EvalString(expr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		if strings.TrimSpace(res.Output) != "" {
			fmt.Print(res.Output)
		} else {
			fmt.Println(res.Value.String())
		}
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
		if strings.TrimSpace(res.Output) != "" {
			fmt.Print(res.Output)
		} else {
			// print last value only if there was no printed output
			fmt.Println(res.Value.String())
		}
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
		if strings.TrimSpace(res.Output) != "" {
			fmt.Print(res.Output)
		}
		fmt.Println(res.Value.String())
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
