package diagnostic

import "fmt"

type Span struct {
	Line      int `json:"line"`
	Column    int `json:"column"`
	EndLine   int `json:"end_line"`
	EndColumn int `json:"end_column"`
}

// Diagnostic is a machine-readable compiler or interpreter message.
type Diagnostic struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Span    Span   `json:"span"`
}

func (d Diagnostic) Error() string {
	if d.Span.Line > 0 {
		return fmt.Sprintf("%d:%d: %s: %s", d.Span.Line, d.Span.Column, d.Code, d.Message)
	}
	return fmt.Sprintf("%s: %s", d.Code, d.Message)
}
