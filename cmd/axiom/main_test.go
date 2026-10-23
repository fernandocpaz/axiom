package main

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/fernandocpaz/axiom/internal/parser"
)

func TestLoadProgramImportsRelativeModules(t *testing.T) {
	dir := t.TempDir()
	if err := ioutil.WriteFile(filepath.Join(dir, "format.ax"), []byte(`fn decorate(text) { return "[" + text + "]" }`), 0600); err != nil {
		t.Fatal(err)
	}
	mainPath := filepath.Join(dir, "main.ax")
	if err := ioutil.WriteFile(mainPath, []byte(`import "format.ax" fn main() { print decorate("ok") }`), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := loadProgram(mainPath, map[string]bool{}, map[string]parser.Program{})
	if err != nil {
		t.Fatal(err)
	}
	var output []parser.Value
	if err := program.Run(func(value parser.Value) { output = append(output, value) }); err != nil {
		t.Fatal(err)
	}
	if len(output) != 1 || output[0].Str != "[ok]" {
		t.Fatalf("unexpected output: %#v", output)
	}
}

func TestLoadProgramRejectsImportCycles(t *testing.T) {
	dir := t.TempDir()
	if err := ioutil.WriteFile(filepath.Join(dir, "a.ax"), []byte(`import "b.ax" fn main() {}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "b.ax"), []byte(`import "a.ax" fn helper() {}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadProgram(filepath.Join(dir, "a.ax"), map[string]bool{}, map[string]parser.Program{}); err == nil {
		t.Fatal("expected import cycle error")
	}
}
