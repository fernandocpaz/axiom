package main

import (
	"fmt"
	"github.com/fernandocpaz/axiom/internal/checker"
	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: axiom <file.ax>")
		os.Exit(2)
	}
	p, e := loadProgram(os.Args[1], map[string]bool{}, map[string]parser.Program{})
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

func loadProgram(path string, active map[string]bool, cache map[string]parser.Program) (parser.Program, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return parser.Program{}, err
	}
	if active[absolute] {
		return parser.Program{}, fmt.Errorf("import cycle involving %q", path)
	}
	if program, ok := cache[absolute]; ok {
		return program, nil
	}
	src, err := ioutil.ReadFile(absolute)
	if err != nil {
		return parser.Program{}, err
	}
	ts, err := lexer.Scan(string(src))
	if err != nil {
		return parser.Program{}, fmt.Errorf("%s: %w", path, err)
	}
	program, err := parser.Parse(ts)
	if err != nil {
		return parser.Program{}, fmt.Errorf("%s: %w", path, err)
	}
	active[absolute] = true
	defer delete(active, absolute)
	merged := parser.Program{}
	for _, imported := range program.Imports {
		importedPath := imported
		if !filepath.IsAbs(importedPath) {
			importedPath = filepath.Join(filepath.Dir(absolute), importedPath)
		}
		dependency, err := loadProgram(importedPath, active, cache)
		if err != nil {
			return parser.Program{}, err
		}
		merged.Functions = append(merged.Functions, dependency.Functions...)
	}
	merged.Functions = append(merged.Functions, program.Functions...)
	for _, function := range merged.Functions {
		if function.Name == "main" {
			merged.Body = function.Body
		}
	}
	cache[absolute] = merged
	return merged, nil
}
