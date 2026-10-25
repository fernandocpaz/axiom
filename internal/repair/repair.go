// Package repair produces deterministic, machine-readable feedback for an
// Axiom generation or edit attempt. It deliberately returns the canonical
// replacement as a whole-file action: models can apply it without guessing
// whitespace or source locations.
package repair

import (
	"github.com/fernandocpaz/axiom/internal/canon"
	"github.com/fernandocpaz/axiom/internal/checker"
	"github.com/fernandocpaz/axiom/internal/diagnostic"
	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
)

const Schema = "axiom.repair/v1"

type Action struct {
	Kind        string `json:"kind"`
	Replacement string `json:"replacement,omitempty"`
}

type Report struct {
	Schema      string                  `json:"schema"`
	Valid       bool                    `json:"valid"`
	Canonical   string                  `json:"canonical,omitempty"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics,omitempty"`
	Actions     []Action                `json:"actions,omitempty"`
}

func Analyze(source string) Report {
	report := Report{Schema: Schema, Valid: false}
	tokens, err := lexer.Scan(source)
	if err != nil {
		report.Diagnostics = []diagnostic.Diagnostic{{Code: "lex-error", Message: err.Error()}}
		return report
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		report.Diagnostics = []diagnostic.Diagnostic{{Code: "parse-error", Message: err.Error()}}
		return report
	}
	report.Diagnostics = checker.Check(program)
	canonical, already, err := canon.Canonical(source)
	if err == nil {
		report.Canonical = canonical
		if !already {
			report.Actions = append(report.Actions, Action{Kind: "replace_source", Replacement: canonical})
		}
	}
	if len(report.Diagnostics) == 0 && already {
		report.Valid = true
	} else if len(report.Diagnostics) == 0 && !already {
		report.Diagnostics = append(report.Diagnostics, diagnostic.Diagnostic{Code: "non-canonical", Message: "source does not match canonical form"})
	}
	return report
}
