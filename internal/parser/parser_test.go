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
	e = p.Run(func(v Value) { got = v.Int })
	if e != nil || got != 42 {
		t.Fatalf("got %d, err %v", got, e)
	}
}

func TestConditionalAndBoolean(t *testing.T) {
	ts, err := lexer.Scan("fn main() { let x = if 4 >= 2 { 7 } else { 0 } print x print x == 7 }")
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	var got []Value
	if err = p.Run(func(v Value) { got = append(got, v) }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Int != 7 || got[1].Kind != "bool" || !got[1].Bool {
		t.Fatalf("unexpected output: %#v", got)
	}
}
