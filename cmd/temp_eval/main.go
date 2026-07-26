package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"simonwaldherr.de/go/smallr/internal/rt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: temp_eval <file>")
		os.Exit(2)
	}
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(2)
	}
	ctx := rt.NewContext()
	res, err := ctx.EvalString(string(b))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if res.Output != "" {
		fmt.Print(res.Output)
	}
	fmt.Println(res.Value.String())
}
