// Package schema turns Axiom's declared record types into deterministic JSON
// Schema documents. Hosts can use the result as a constrained-decoding schema
// or include it in a repair prompt without inspecting compiler internals.
package schema

import (
	"encoding/json"
	"sort"

	"github.com/fernandocpaz/axiom/internal/parser"
)

const Schema = "axiom.schema/v1"

// Encode returns the stable JSON schema catalogue for a parsed program.
func Encode(program parser.Program) ([]byte, error) {
	types := map[string]interface{}{}
	names := make([]string, 0, len(program.Types))
	for _, declaration := range program.Types {
		names = append(names, declaration.Name)
	}
	sort.Strings(names)
	for _, name := range names {
		for _, declaration := range program.Types {
			if declaration.Name == name {
				types[name] = recordSchema(declaration.Fields)
				break
			}
		}
	}
	document := map[string]interface{}{
		"schema": "https://json-schema.org/draft/2020-12/schema",
		"id":     Schema,
		"types":  types,
	}
	return json.Marshal(document)
}

func recordSchema(fields map[string]parser.TypeRef) map[string]interface{} {
	properties := map[string]interface{}{}
	required := make([]string, 0, len(fields))
	for name, typ := range fields {
		properties[name] = refSchema(typ)
		required = append(required, name)
	}
	sort.Strings(required)
	return map[string]interface{}{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

func refSchema(typ parser.TypeRef) interface{} {
	if len(typ.Args) == 0 {
		switch typ.Name {
		case "int":
			return map[string]interface{}{"type": "integer"}
		case "string":
			return map[string]interface{}{"type": "string"}
		case "bool":
			return map[string]interface{}{"type": "boolean"}
		case "none":
			return map[string]interface{}{"type": "null"}
		case "record":
			return map[string]interface{}{"type": "object"}
		default:
			return map[string]interface{}{"$ref": "#/types/" + typ.Name}
		}
	}
	if typ.Name == "Result" && len(typ.Args) == 2 {
		return map[string]interface{}{"oneOf": []interface{}{
			map[string]interface{}{"type": "object", "properties": map[string]interface{}{"ok": map[string]interface{}{"const": true}, "value": refSchema(typ.Args[0])}, "required": []string{"ok", "value"}, "additionalProperties": false},
			map[string]interface{}{"type": "object", "properties": map[string]interface{}{"ok": map[string]interface{}{"const": false}, "error": refSchema(typ.Args[1])}, "required": []string{"ok", "error"}, "additionalProperties": false},
		}}
	}
	if typ.Name == "list" && len(typ.Args) == 1 {
		return map[string]interface{}{"type": "array", "items": refSchema(typ.Args[0])}
	}
	return map[string]interface{}{"$ref": "#/types/" + typ.Name}
}
