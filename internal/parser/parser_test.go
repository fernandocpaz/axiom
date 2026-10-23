package parser

import (
	"github.com/fernandocpaz/axiom/internal/lexer"
	"testing"
)

func TestArithmetic(t *testing.T) {
	ts, e := lexer.Scan("fn main() { let x = 2 * (20 + 1) print x }")
	if e != nil {
		t.Fatal(e)
	}
	p, e := Parse(ts)
	if e != nil {
		t.Fatal(e)
	}
	got := -1
	p.Run(func(v int) { got = v })
	if got != 42 {
		t.Fatalf("got %d", got)
	}
}
