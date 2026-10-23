package diagnostic

import "fmt"

type Span struct {
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}

// Diagnostic is a machine-readable compiler or interpreter message.
type Diagnostic struct {
	Code    string
	Message string
	Span    Span
}

func (d Diagnostic) Error() string {
	if d.Span.Line > 0 {
		return fmt.Sprintf("%d:%d: %s: %s", d.Span.Line, d.Span.Column, d.Code, d.Message)
	}
	return fmt.Sprintf("%s: %s", d.Code, d.Message)
}
