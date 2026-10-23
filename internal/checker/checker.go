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
	List    = "list"
	Record  = "record"
	None    = "none"
	Unknown = "unknown"
)

// Check validates a parsed program without executing it.
func Check(program parser.Program) []diagnostic.Diagnostic {
	functions := map[string]parser.Function{}
	for _, function := range program.Functions {
		functions[function.Name] = function
	}
	if _, ok := functions["main"]; !ok {
		functions["main"] = parser.Function{Name: "main", Body: program.Body}
	}
	c := &checker{env: map[string]string{}, functions: functions}
	for _, function := range functions {
		c.env = map[string]string{}
		for _, param := range function.Params {
			c.env[param] = Unknown
		}
		c.checkBody(function.Body)
	}
	return c.diagnostics
}

func (c *checker) checkBody(body []parser.Stmt) {
	for _, stmt := range body {
		if stmt.Set != "" {
			valueType := c.expr(stmt.Value)
			previous, ok := c.env[stmt.Set]
			if !ok {
				c.report("undefined-name", fmt.Sprintf("undefined variable %q", stmt.Set), exprSpan(stmt.Value))
			} else if previous != Unknown && valueType != Unknown && previous != valueType {
				c.report("assignment-type", fmt.Sprintf("cannot assign %s to %s variable %q", valueType, previous, stmt.Set), exprSpan(stmt.Value))
			}
			if ok {
				c.env[stmt.Set] = valueType
			}
			continue
		}
		if stmt.Let != "" {
			t := c.expr(stmt.Value)
			c.env[stmt.Let] = t
			continue
		}
		if stmt.Print != nil {
			c.expr(stmt.Print)
		}
		if stmt.Return {
			c.expr(stmt.Result)
		}
		if stmt.While != nil {
			condition := c.expr(stmt.While)
			if condition != Bool && condition != Unknown {
				c.report("condition-type", fmt.Sprintf("while condition must be bool, got %s", condition), exprSpan(stmt.While))
			}
			c.checkBody(stmt.Body)
		}
		if stmt.ForVar != "" {
			c.expr(stmt.ForExpr)
			c.env[stmt.ForVar] = Unknown
			c.checkBody(stmt.Body)
		}
		if stmt.TryBody != nil {
			c.checkBody(stmt.TryBody)
			c.env[stmt.CatchName] = String
			c.checkBody(stmt.CatchBody)
		}
	}
}

type checker struct {
	env         map[string]string
	functions   map[string]parser.Function
	diagnostics []diagnostic.Diagnostic
}

func exprSpan(expr parser.Expr) diagnostic.Span {
	if located, ok := expr.(parser.Located); ok {
		return located.Span
	}
	return diagnostic.Span{}
}

func (c *checker) report(code, message string, span diagnostic.Span) {
	c.diagnostics = append(c.diagnostics, diagnostic.Diagnostic{Code: code, Message: message, Span: span})
}

func (c *checker) expr(expr parser.Expr) string {
	span := diagnostic.Span{}
	if located, ok := expr.(parser.Located); ok {
		span = located.Span
		expr = located.Inner
	}
	switch e := expr.(type) {
	case parser.Number:
		return Int
	case parser.String:
		return String
	case parser.Boolean:
		return Bool
	case parser.None:
		return None
	case parser.ListLiteral:
		for _, item := range e {
			c.expr(item)
		}
		return List
	case parser.RecordLiteral:
		for _, item := range e {
			c.expr(item)
		}
		return Record
	case parser.Field:
		c.expr(e.Object)
		return Unknown
	case parser.Index:
		c.expr(e.Object)
		c.expr(e.Key)
		return Unknown
	case parser.Variable:
		t, ok := c.env[string(e)]
		if !ok {
			c.report("undefined-name", fmt.Sprintf("undefined variable %q", string(e)), span)
			return Unknown
		}
		return t
	case parser.Binary:
		return c.binary(e, span)
	case parser.Unary:
		operand := c.expr(e.Value)
		if operand == Unknown {
			return Unknown
		}
		if e.Op == lexer.Not && operand != Bool {
			c.report("operator-type", fmt.Sprintf("operator ! requires bool, got %s", operand), span)
			return Unknown
		}
		if (e.Op == lexer.Plus || e.Op == lexer.Minus) && operand != Int {
			c.report("operator-type", fmt.Sprintf("unary operator requires int, got %s", operand), span)
			return Unknown
		}
		if e.Op == lexer.Not {
			return Bool
		}
		return Int
	case parser.Call:
		if _, ok := c.functions[e.Name]; !ok && !isBuiltin(e.Name) {
			c.report("undefined-function", fmt.Sprintf("undefined function %q", e.Name), span)
		}
		for _, arg := range e.Args {
			c.expr(arg)
		}
		switch e.Name {
		case "len", "number":
			return Int
		case "text":
			return String
		case "contains", "starts_with", "ends_with", "matches":
			return Bool
		}
		return Unknown
	case parser.Conditional:
		condition := c.expr(e.Test)
		if condition != Bool && condition != Unknown {
			c.report("condition-type", fmt.Sprintf("if condition must be bool, got %s", condition), exprSpan(e.Test))
		}
		yes := c.expr(e.Yes)
		no := c.expr(e.No)
		if yes != Unknown && no != Unknown && yes != no {
			c.report("branch-type", fmt.Sprintf("if branches must have the same type, got %s and %s", yes, no), span)
			return Unknown
		}
		if yes == Unknown {
			return no
		}
		return yes
	default:
		c.report("internal-error", "unknown expression", span)
		return Unknown
	}
}

func isBuiltin(name string) bool {
	switch name {
	case "len", "text", "number", "contains", "starts_with", "ends_with", "matches":
		return true
	default:
		return false
	}
}

func (c *checker) binary(expr parser.Binary, span diagnostic.Span) string {
	left, right := c.expr(expr.Left), c.expr(expr.Right)
	if left == Unknown || right == Unknown {
		return Unknown
	}
	switch expr.Op {
	case lexer.AndAnd, lexer.OrOr:
		if left != Bool || right != Bool {
			c.report("operator-type", fmt.Sprintf("logical operator requires bool operands, got %s and %s", left, right), span)
			return Unknown
		}
		return Bool
	case lexer.Plus, lexer.Minus, lexer.Star, lexer.Slash:
		if expr.Op == lexer.Plus && left == String && right == String {
			return String
		}
		if left != Int || right != Int {
			c.report("operator-type", fmt.Sprintf("operator %v requires int operands, got %s and %s", expr.Op, left, right), span)
			return Unknown
		}
		return Int
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		if left != Int || right != Int {
			c.report("comparison-type", fmt.Sprintf("ordered comparison requires int operands, got %s and %s", left, right), span)
		}
		return Bool
	case lexer.EqualEqual, lexer.NotEqual:
		if left != right {
			c.report("equality-type", fmt.Sprintf("cannot compare %s and %s", left, right), span)
		}
		return Bool
	default:
		c.report("unknown-operator", fmt.Sprintf("unknown operator %v", expr.Op), span)
		return Unknown
	}
}
