// Package ir exposes the stable machine-facing representation of an Axiom
// program. It intentionally uses tagged JSON nodes instead of Go's internal
// parser layout so agents can inspect and patch programs without depending on
// compiler implementation details.
package ir

import (
	"encoding/json"
	"sort"

	"github.com/fernandocpaz/axiom/internal/diagnostic"
	"github.com/fernandocpaz/axiom/internal/parser"
)

const Schema = "axiom.ast/v1"

// Encode returns deterministic JSON for program. encoding/json sorts map keys;
// record fields are represented as sorted arrays as well, so output is stable
// across runs and suitable for model context, caching, and exact diffs.
func Encode(program parser.Program) ([]byte, error) {
	document := map[string]interface{}{
		"schema":       Schema,
		"imports":      stringList(program.Imports),
		"capabilities": stringList(program.Capabilities),
		"functions":    functions(program.Functions),
		"body":         statements(program.Body),
	}
	return json.Marshal(document)
}

func functions(items []parser.Function) []interface{} {
	result := make([]interface{}, len(items))
	for i, item := range items {
		result[i] = map[string]interface{}{
			"kind":   "function",
			"name":   item.Name,
			"params": stringList(item.Params),
			"body":   statements(item.Body),
		}
	}
	return result
}

func stringList(items []string) []string {
	if items == nil {
		return []string{}
	}
	return append([]string{}, items...)
}

func statements(items []parser.Stmt) []interface{} {
	if items == nil {
		return []interface{}{}
	}
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		var node map[string]interface{}
		switch {
		case item.Let != "":
			node = map[string]interface{}{"kind": "let", "name": item.Let, "value": expr(item.Value)}
		case item.Set != "":
			node = map[string]interface{}{"kind": "set", "name": item.Set, "value": expr(item.Value)}
		case item.Print != nil:
			node = map[string]interface{}{"kind": "print", "value": expr(item.Print)}
		case item.Return:
			node = map[string]interface{}{"kind": "return", "value": expr(item.Result)}
		case item.While != nil:
			node = map[string]interface{}{"kind": "while", "condition": expr(item.While), "body": statements(item.Body)}
		case item.ForVar != "":
			node = map[string]interface{}{"kind": "for", "name": item.ForVar, "collection": expr(item.ForExpr), "body": statements(item.Body)}
		case item.TryBody != nil:
			node = map[string]interface{}{"kind": "try", "body": statements(item.TryBody), "catch": item.CatchName, "catch_body": statements(item.CatchBody)}
		case item.Break:
			node = map[string]interface{}{"kind": "break"}
		case item.Continue:
			node = map[string]interface{}{"kind": "continue"}
		default:
			node = map[string]interface{}{"kind": "unknown"}
		}
		result = append(result, node)
	}
	return result
}

func expr(value parser.Expr) interface{} {
	if value == nil {
		return nil
	}
	var span diagnostic.Span
	if located, ok := value.(parser.Located); ok {
		span = located.Span
		value = located.Inner
	}
	var node map[string]interface{}
	switch item := value.(type) {
	case parser.Number:
		node = map[string]interface{}{"kind": "number", "value": int(item)}
	case parser.String:
		node = map[string]interface{}{"kind": "string", "value": string(item)}
	case parser.Boolean:
		node = map[string]interface{}{"kind": "boolean", "value": bool(item)}
	case parser.None:
		node = map[string]interface{}{"kind": "none"}
	case parser.Variable:
		node = map[string]interface{}{"kind": "variable", "name": string(item)}
	case parser.Unary:
		node = map[string]interface{}{"kind": "unary", "operator": item.Op.String(), "value": expr(item.Value)}
	case parser.Binary:
		node = map[string]interface{}{"kind": "binary", "operator": item.Op.String(), "left": expr(item.Left), "right": expr(item.Right)}
	case parser.Conditional:
		node = map[string]interface{}{"kind": "conditional", "test": expr(item.Test), "yes": expr(item.Yes), "no": expr(item.No)}
	case parser.Call:
		node = map[string]interface{}{"kind": "call", "name": item.Name, "args": expressions(item.Args)}
	case parser.Pipeline:
		node = map[string]interface{}{"kind": "pipeline", "name": item.Name, "left": expr(item.Left), "args": expressions(item.Args)}
	case parser.ListLiteral:
		node = map[string]interface{}{"kind": "list", "items": expressions(item)}
	case parser.RecordLiteral:
		keys := make([]string, 0, len(item))
		for key := range item {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fields := make([]interface{}, 0, len(keys))
		for _, key := range keys {
			fields = append(fields, map[string]interface{}{"name": key, "value": expr(item[key])})
		}
		node = map[string]interface{}{"kind": "record", "fields": fields}
	case parser.Field:
		node = map[string]interface{}{"kind": "field", "object": expr(item.Object), "name": item.Name}
	case parser.Index:
		node = map[string]interface{}{"kind": "index", "object": expr(item.Object), "key": expr(item.Key)}
	default:
		node = map[string]interface{}{"kind": "unknown"}
	}
	if span.Line > 0 {
		node["span"] = map[string]interface{}{
			"line": span.Line, "column": span.Column,
			"end_line": span.EndLine, "end_column": span.EndColumn,
		}
	}
	return node
}

func expressions(items []parser.Expr) []interface{} {
	if items == nil {
		return []interface{}{}
	}
	result := make([]interface{}, len(items))
	for i, item := range items {
		result[i] = expr(item)
	}
	return result
}
