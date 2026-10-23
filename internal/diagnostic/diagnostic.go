package diagnostic

import "fmt"

// Diagnostic is a machine-readable compiler or interpreter message.
type Diagnostic struct {
	Code    string
	Message string
	Line    int
	Column  int
}

func (d Diagnostic) Error() string {
	if d.Line > 0 {
		return fmt.Sprintf("%d:%d: %s: %s", d.Line, d.Column, d.Code, d.Message)
	}
	return fmt.Sprintf("%s: %s", d.Code, d.Message)
}
