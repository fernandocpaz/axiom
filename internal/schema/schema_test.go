package schema

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
)

func TestEncodeDeclaredTypes(t *testing.T) {
	tokens, err := lexer.Scan(`type Label=record{value:string,confidence:int} fn main(){print 1}`)
	if err != nil {
		t.Fatal(err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	data, err := Encode(program)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]interface{}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"Label"`) || !strings.Contains(string(data), `"integer"`) {
		t.Fatalf("unexpected schema: %s", data)
	}
}
