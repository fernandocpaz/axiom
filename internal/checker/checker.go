package checker

import (
	"fmt"
	"github.com/fernandocpaz/axiom/internal/diagnostic"
	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
	"strings"
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
	capabilities := map[string]bool{}
	for _, capability := range program.Capabilities {
		capabilities[capability] = true
	}
	c := &checker{env: map[string]string{}, functions: functions, capabilities: capabilities, inferred: map[string]string{}, inferring: map[string]bool{}}
	for _, function := range functions {
		c.env = map[string]string{}
		for _, param := range function.Params {
			c.env[param] = Unknown
		}
		c.returnTypes = nil
		c.checkBody(function.Body)
		c.reportInconsistentReturns(c.returnTypes)
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
			c.returnTypes = append(c.returnTypes, c.expr(stmt.Result))
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
	env          map[string]string
	functions    map[string]parser.Function
	capabilities map[string]bool
	diagnostics  []diagnostic.Diagnostic
	returnTypes  []string
	inferred     map[string]string
	inferring    map[string]bool
	silent       bool
}

func exprSpan(expr parser.Expr) diagnostic.Span {
	if located, ok := expr.(parser.Located); ok {
		return located.Span
	}
	return diagnostic.Span{}
}

func (c *checker) report(code, message string, span diagnostic.Span) {
	if c.silent {
		return
	}
	c.diagnostics = append(c.diagnostics, diagnostic.Diagnostic{Code: code, Message: message, Span: span})
}

func (c *checker) reportInconsistentReturns(types []string) {
	if len(types) < 2 {
		return
	}
	first := types[0]
	for _, typ := range types[1:] {
		if first != Unknown && typ != Unknown && first != typ {
			c.report("return-type", fmt.Sprintf("function returns both %s and %s", first, typ), diagnostic.Span{})
			return
		}
	}
}

func inferredKey(name string, args []string) string {
	return name + "(" + strings.Join(args, ",") + ")"
}

// functionType specializes a function at a call site so functions such as
// identity(value) preserve the argument's static type without type syntax.
func (c *checker) functionType(name string, args []string) string {
	function, ok := c.functions[name]
	if !ok {
		return Unknown
	}
	key := inferredKey(name, args)
	if typ, ok := c.inferred[key]; ok {
		return typ
	}
	if c.inferring[key] {
		return Unknown
	}
	c.inferring[key] = true

	savedEnv, savedReturns, savedSilent := c.env, c.returnTypes, c.silent
	c.env = map[string]string{}
	for i, param := range function.Params {
		typ := Unknown
		if i < len(args) {
			typ = args[i]
		}
		c.env[param] = typ
	}
	c.returnTypes = nil
	c.silent = true
	c.checkBody(function.Body)
	typ := None
	if len(c.returnTypes) > 0 {
		typ = c.returnTypes[0]
		for _, returnType := range c.returnTypes[1:] {
			if returnType != typ {
				typ = Unknown
				break
			}
		}
	}
	c.env, c.returnTypes, c.silent = savedEnv, savedReturns, savedSilent
	delete(c.inferring, key)
	c.inferred[key] = typ
	return typ
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
			if _, isFunction := c.functions[string(e)]; isFunction {
				return "function"
			}
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
		argTypes := make([]string, len(e.Args))
		for i, arg := range e.Args {
			argTypes[i] = c.expr(arg)
		}
		requiredCapability := map[string]string{
			"read_file": "filesystem.read", "write_file": "filesystem.write",
			"run": "process.run", "sleep": "time.sleep",
			"infer": "model.infer", "recall": "retrieval.read", "invoke": "tool.invoke",
		}[e.Name]
		if requiredCapability != "" && !c.capabilities[requiredCapability] {
			c.report("capability", fmt.Sprintf("%s requires allow %q", e.Name, requiredCapability), span)
		}
		switch e.Name {
		case "len", "number":
			return Int
		case "map", "filter":
			return List
		case "text", "json_encode", "read_file":
			return String
		case "now":
			return String
		case "write_file", "sleep":
			return None
		case "run":
			return Record
		case "infer", "invoke":
			return Record
		case "recall":
			return List
		case "json_decode":
			return Unknown
		case "contains", "starts_with", "ends_with", "matches":
			return Bool
		}
		if _, ok := c.functions[e.Name]; ok {
			return c.functionType(e.Name, argTypes)
		}
		return Unknown
	case parser.Pipeline:
		argTypes := []string{c.expr(e.Left)}
		for _, arg := range e.Args {
			argTypes = append(argTypes, c.expr(arg))
		}
		if _, ok := c.functions[e.Name]; !ok && !isBuiltin(e.Name) {
			c.report("undefined-function", fmt.Sprintf("undefined function %q", e.Name), span)
		}
		if e.Name == "map" || e.Name == "filter" {
			return List
		}
		if _, ok := c.functions[e.Name]; ok {
			return c.functionType(e.Name, argTypes)
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
	case "len", "text", "number", "contains", "starts_with", "ends_with", "matches", "json_encode", "json_decode", "read_file", "write_file", "run", "now", "sleep", "map", "filter", "infer", "recall", "invoke":
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
