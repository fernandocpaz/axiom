package host

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fernandocpaz/axiom/internal/parser"
)

func TestOpenAICompatibleInferValidatesStructuredOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Method != "POST" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := ioutil.ReadAll(r.Body)
		if !strings.Contains(string(body), `"model":"test-model"`) || !strings.Contains(string(body), `"json_object"`) {
			t.Fatalf("unexpected request body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"label\":\"infra\",\"confidence\":3}"}}]}`))
	}))
	defer server.Close()
	adapter := OpenAICompatible{BaseURL: server.URL + "/v1", Model: "test-model", Client: server.Client()}
	schema := parser.Value{Kind: "record", Fields: map[string]parser.Value{
		"label":      {Kind: "string", Str: "string"},
		"confidence": {Kind: "string", Str: "int"},
	}}
	value, err := adapter.Infer("classify", schema)
	if err != nil {
		t.Fatal(err)
	}
	if value.Fields["label"].Str != "infra" || value.Fields["confidence"].Int != 3 {
		t.Fatalf("unexpected value: %#v", value)
	}
}

func TestValidateSchemaRejectsMissingOrWrongFields(t *testing.T) {
	schema := parser.Value{Kind: "record", Fields: map[string]parser.Value{"label": {Kind: "string", Str: "string"}}}
	if err := ValidateSchema(parser.Value{Kind: "record", Fields: map[string]parser.Value{}}, schema); err == nil {
		t.Fatal("expected missing field error")
	}
	if err := ValidateSchema(parser.Value{Kind: "record", Fields: map[string]parser.Value{"label": {Kind: "int", Int: 1}}}, schema); err == nil {
		t.Fatal("expected type error")
	}
}
