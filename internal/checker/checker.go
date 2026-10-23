package checker

import (
	"fmt"
	"github.com/fernandocpaz/axiom/internal/diagnostic"
	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
)

const (
	Int     = "int"
	Bool    = "bool"
	String  = "string"
	Unknown = "unknown"
)

// Check validates a parsed program without executing it.
func Check(program parser.Program) []diagnostic.Diagnostic {
	c := &checker{env: map[string]string{}}
	for _, stmt := range program.Body {
		if stmt.Let != "" {
			t := c.expr(stmt.Value)
			c.env[stmt.Let] = t
			continue
		}
		c.expr(stmt.Print)
	}
	return c.diagnostics
}

type checker struct {
	env         map[string]string
	diagnostics []diagnostic.Diagnostic
}

func (c *checker) report(code, message string) {
	c.diagnostics = append(c.diagnostics, diagnostic.Diagnostic{Code: code, Message: message})
}

func (c *checker) expr(expr parser.Expr) string {
	switch e := expr.(type) {
	case parser.Number:
		return Int
	case parser.String:
		return String
	case parser.Boolean:
		return Bool
	case parser.Variable:
		t, ok := c.env[string(e)]
		if !ok {
			c.report("undefined-name", fmt.Sprintf("undefined variable %q", string(e)))
			return Unknown
		}
		return t
	case parser.Binary:
		return c.binary(e)
	case parser.Conditional:
		condition := c.expr(e.Test)
		if condition != Bool && condition != Unknown {
			c.report("condition-type", fmt.Sprintf("if condition must be bool, got %s", condition))
		}
		yes := c.expr(e.Yes)
		no := c.expr(e.No)
		if yes != Unknown && no != Unknown && yes != no {
			c.report("branch-type", fmt.Sprintf("if branches must have the same type, got %s and %s", yes, no))
			return Unknown
		}
		if yes == Unknown {
			return no
		}
		return yes
	default:
		c.report("internal-error", "unknown expression")
		return Unknown
	}
}

func (c *checker) binary(expr parser.Binary) string {
	left, right := c.expr(expr.Left), c.expr(expr.Right)
	if left == Unknown || right == Unknown {
		return Unknown
	}
	switch expr.Op {
	case lexer.Plus, lexer.Minus, lexer.Star, lexer.Slash:
		if left != Int || right != Int {
			c.report("operator-type", fmt.Sprintf("operator %v requires int operands, got %s and %s", expr.Op, left, right))
			return Unknown
		}
		return Int
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		if left != Int || right != Int {
			c.report("comparison-type", fmt.Sprintf("ordered comparison requires int operands, got %s and %s", left, right))
		}
		return Bool
	case lexer.EqualEqual, lexer.NotEqual:
		if left != right {
			c.report("equality-type", fmt.Sprintf("cannot compare %s and %s", left, right))
		}
		return Bool
	default:
		c.report("unknown-operator", fmt.Sprintf("unknown operator %v", expr.Op))
		return Unknown
	}
}
