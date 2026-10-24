package ir

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
)

func TestEncodeIsDeterministic(t *testing.T) {
	src := `allow "z" allow "a" fn main() { print record { z: 2, a: 1 } }`
	program := parse(t, src)
	one, err := Encode(program)
	if err != nil {
		t.Fatal(err)
	}
	two, err := Encode(program)
	if err != nil {
		t.Fatal(err)
	}
	if string(one) != string(two) {
		t.Fatalf("encoding is not deterministic\nfirst: %s\nsecond: %s", one, two)
	}
	if !strings.Contains(string(one), `"schema":"`+Schema+`"`) {
		t.Fatalf("missing schema: %s", one)
	}
	if !strings.Contains(string(one), `"fields":[{"name":"a"`) {
		t.Fatalf("record fields are not sorted: %s", one)
	}
}

func TestEncodeContainsSourceSpans(t *testing.T) {
	program := parse(t, "fn main(){print 42}")
	encoded, err := Encode(program)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]interface{}
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"span":{"column":17`) {
		t.Fatalf("expected expression span: %s", encoded)
	}
}

func parse(t *testing.T, src string) parser.Program {
	t.Helper()
	tokens, err := lexer.Scan(src)
	if err != nil {
		t.Fatal(err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	return program
}
