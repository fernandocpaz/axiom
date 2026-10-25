// Package host contains explicit adapters for external AI systems. Adapters
// are opt-in: Axiom programs still need the matching capability declaration.
package host

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fernandocpaz/axiom/internal/parser"
)

// OpenAICompatible talks to /v1/chat/completions. Ollama exposes this same
// endpoint, so the adapter works for both cloud OpenAI-compatible gateways and
// a local Ollama server.
type OpenAICompatible struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

// NewOllama creates an adapter for Ollama's OpenAI-compatible endpoint.
func NewOllama(model string) OpenAICompatible {
	return OpenAICompatible{BaseURL: "http://127.0.0.1:11434/v1", Model: model}
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    int             `json:"temperature"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Infer asks the model for a JSON object matching schema and validates the
// decoded response before returning it to the Axiom runtime.
func (h OpenAICompatible) Infer(prompt string, schema parser.Value) (parser.Value, error) {
	if schema.Kind != "record" {
		return parser.Value{}, fmt.Errorf("inference schema must be a record")
	}
	if h.BaseURL == "" {
		return parser.Value{}, fmt.Errorf("model host URL is empty")
	}
	if h.Model == "" {
		return parser.Value{}, fmt.Errorf("model name is empty")
	}
	schemaJSON, err := parser.ValueToJSON(schema)
	if err != nil {
		return parser.Value{}, fmt.Errorf("encode inference schema: %w", err)
	}
	system := "Return only one JSON object. It must satisfy this Axiom schema: " + string(schemaJSON)
	payload := chatRequest{
		Model: h.Model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: prompt},
		},
		Temperature:    0,
		ResponseFormat: &responseFormat{Type: "json_object"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return parser.Value{}, err
	}
	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}
	url := strings.TrimRight(h.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return parser.Value{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if h.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.APIKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return parser.Value{}, err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return parser.Value{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parser.Value{}, fmt.Errorf("model host returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	var decoded chatResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return parser.Value{}, fmt.Errorf("decode model response: %w", err)
	}
	if decoded.Error != nil {
		return parser.Value{}, fmt.Errorf("model host: %s", decoded.Error.Message)
	}
	if len(decoded.Choices) == 0 {
		return parser.Value{}, fmt.Errorf("model host returned no choices")
	}
	value, err := parser.ValueFromJSON([]byte(strings.TrimSpace(decoded.Choices[0].Message.Content)))
	if err != nil {
		return parser.Value{}, fmt.Errorf("decode model JSON output: %w", err)
	}
	if err := ValidateSchema(value, schema); err != nil {
		return parser.Value{}, fmt.Errorf("model output failed schema validation: %w", err)
	}
	return value, nil
}

// Generate asks the same endpoint for source text. The caller remains
// responsible for parsing and checking the result; generated text never
// enters the runtime directly.
func (h OpenAICompatible) Generate(prompt string) (string, error) {
	if h.BaseURL == "" {
		return "", fmt.Errorf("model host URL is empty")
	}
	if h.Model == "" {
		return "", fmt.Errorf("model name is empty")
	}
	payload := chatRequest{Model: h.Model, Messages: []chatMessage{
		{Role: "system", Content: "Return only canonical Axiom source. Do not use markdown fences or explanations."},
		{Role: "user", Content: prompt},
	}}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}
	url := strings.TrimRight(h.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if h.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.APIKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("model host returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	var decoded chatResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return "", fmt.Errorf("decode model response: %w", err)
	}
	if decoded.Error != nil {
		return "", fmt.Errorf("model host: %s", decoded.Error.Message)
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("model host returned no choices")
	}
	return stripFences(strings.TrimSpace(decoded.Choices[0].Message.Content)), nil
}

func stripFences(source string) string {
	if strings.HasPrefix(source, "```") && strings.HasSuffix(source, "```") {
		lines := strings.Split(source, "\n")
		if len(lines) >= 3 {
			return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
		}
	}
	return source
}

// ValidateSchema checks the small structural schema used by infer. A schema
// is a record whose values are type names (string, int, bool, none, list,
// record) or nested record schemas. Fields are required by default.
func ValidateSchema(value parser.Value, schema parser.Value) error {
	if schema.Kind != "record" || value.Kind != "record" {
		return fmt.Errorf("expected record value and record schema")
	}
	for name, expected := range schema.Fields {
		actual, ok := value.Fields[name]
		if !ok {
			return fmt.Errorf("missing field %q", name)
		}
		if expected.Kind == "string" {
			if actual.Kind != expected.Str {
				return fmt.Errorf("field %q requires %s, got %s", name, expected.Str, actual.Kind)
			}
			continue
		}
		if expected.Kind == "record" {
			if err := ValidateSchema(actual, expected); err != nil {
				return fmt.Errorf("field %q: %w", name, err)
			}
			continue
		}
		return fmt.Errorf("schema field %q must be a type name or nested record", name)
	}
	return nil
}
