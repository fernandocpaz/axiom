package main

import (
	"fmt"
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
	p.Run(func(v int) { fmt.Println(v) })
}
