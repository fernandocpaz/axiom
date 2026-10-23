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

func TestDiagnosticsCarrySourceSpans(t *testing.T) {
	ts, err := lexer.Scan("fn main() {\n  print true + 1\n}")
	if err != nil {
		t.Fatal(err)
	}
	p, err := parser.Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := Check(p)
	if len(diagnostics) != 1 {
		t.Fatalf("got %d diagnostics, want 1", len(diagnostics))
	}
	span := diagnostics[0].Span
	if span.Line != 2 || span.Column != 9 || span.EndLine != 2 || span.EndColumn != 17 {
		t.Fatalf("got span %#v", span)
	}
}

func TestAcceptsFunctionCalls(t *testing.T) {
	if got := checkSource(t, `fn add(a, b) { return a + b } fn main() { print add(20, 22) }`); len(got) != 0 {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}

func TestAcceptsStructuredValues(t *testing.T) {
	if got := checkSource(t, `fn main() { let x = record { values: [1, 2], active: true } print x.values[0] }`); len(got) != 0 {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}

func TestChecksWhileConditions(t *testing.T) {
	got := checkSource(t, `fn main() { while 1 { break } }`)
	if len(got) != 1 || got[0] != "condition-type" {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}

func TestAcceptsLogicalExpressions(t *testing.T) {
	if got := checkSource(t, `fn main() { print !false && (1 < 2 || false) }`); len(got) != 0 {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}

func TestChecksAssignmentTypes(t *testing.T) {
	got := checkSource(t, `fn main() { let x = 1 set x = "wrong" }`)
	if len(got) != 1 || got[0] != "assignment-type" {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}

func TestAcceptsTextBuiltins(t *testing.T) {
	if got := checkSource(t, `fn main() { print matches("api", "a.*") print len([1, 2]) }`); len(got) != 0 {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}

func TestAcceptsForInIteration(t *testing.T) {
	if got := checkSource(t, `fn main() { for item in [1, 2] { print item } }`); len(got) != 0 {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}

func TestAcceptsTryCatch(t *testing.T) {
	if got := checkSource(t, `fn main() { try { print 1 / 0 } catch error { print error } }`); len(got) != 0 {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}

func TestAcceptsJSONBuiltins(t *testing.T) {
	if got := checkSource(t, `fn main() { print json_encode(record { ok: true }) }`); len(got) != 0 {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}
