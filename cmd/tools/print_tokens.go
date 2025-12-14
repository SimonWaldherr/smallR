package main

import (
	"fmt"
	"os"

	"simonwaldherr.de/go/smallr/internal/lexer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: print_tokens <file>")
		os.Exit(1)
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("read error:", err)
		os.Exit(1)
	}
	l := lexer.New(string(b))
	for {
		tok := l.Next()
		fmt.Printf("%s %q %d:%d\n", tok.Type, tok.Lit, tok.Pos.Line, tok.Pos.Col)
		if tok.Type == "EOF" {
			break
		}
	}
}
