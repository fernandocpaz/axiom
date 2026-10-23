package main

import (
	"fmt"
	"github.com/fernandocpaz/axiom/internal/checker"
	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: axiom <file.ax>")
		os.Exit(2)
	}
	src, e := ioutil.ReadFile(os.Args[1])
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	ts, e := lexer.Scan(string(src))
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	p, e := parser.Parse(ts)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	diagnostics := checker.Check(p)
	for _, d := range diagnostics {
		fmt.Fprintln(os.Stderr, d.Error())
	}
	if len(diagnostics) > 0 {
		os.Exit(1)
	}
	if err := p.Run(func(v parser.Value) { fmt.Println(v) }); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
