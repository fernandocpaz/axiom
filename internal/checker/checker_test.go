package checker

import (
	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
	"testing"
)

func checkSource(t *testing.T, source string) []string {
	t.Helper()
	ts, err := lexer.Scan(source)
	if err != nil {
		t.Fatal(err)
	}
	p, err := parser.Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := Check(p)
	codes := make([]string, len(diagnostics))
	for i, d := range diagnostics {
		codes[i] = d.Code
	}
	return codes
}

func TestValidProgram(t *testing.T) {
	if got := checkSource(t, `fn main() { let x = if true { 1 } else { 2 } print x == 1 }`); len(got) != 0 {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}

func TestReportsStaticTypeErrors(t *testing.T) {
	got := checkSource(t, `fn main() { let x = true + 1 let y = if 1 { "a" } else { 2 } print missing }`)
	want := []string{"operator-type", "condition-type", "branch-type", "undefined-name"}
	if len(got) != len(want) {
		t.Fatalf("got codes %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got codes %v, want %v", got, want)
		}
	}
}
