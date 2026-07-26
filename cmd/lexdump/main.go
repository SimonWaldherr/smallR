package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"simonwaldherr.de/go/smallr/internal/lexer"
	"simonwaldherr.de/go/smallr/internal/token"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: lexdump <file>")
		os.Exit(2)
	}
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(2)
	}
	l := lexer.New(string(b))
	for {
		tok := l.Next()
		fmt.Printf("%3d: %-12s %q\n", tok.Pos.Line, tok.Type, tok.Lit)
		if tok.Type == token.EOF {
			break
		}
	}
}
