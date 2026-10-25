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
	types := map[string]parser.TypeDecl{}
	for _, typ := range program.Types {
		if _, exists := types[typ.Name]; exists {
			// Keep checking the program and report the duplicate through the same
			// stable diagnostic channel as other static errors.
			continue
		}
		types[typ.Name] = typ
	}
	c := &checker{env: map[string]string{}, functions: functions, types: types, capabilities: capabilities, inferred: map[string]string{}, inferring: map[string]bool{}}
	seenTypes := map[string]bool{}
	for _, typ := range program.Types {
		if seenTypes[typ.Name] {
			c.report("duplicate-type", fmt.Sprintf("type %q is declared more than once", typ.Name), diagnostic.Span{})
		}
		seenTypes[typ.Name] = true
		for field, fieldType := range typ.Fields {
			if !knownType(fieldType, types) {
				c.report("unknown-type", fmt.Sprintf("field %q.%s uses unknown type %s", typ.Name, field, typeName(fieldType)), diagnostic.Span{})
			}
		}
	}
	for _, function := range functions {
		c.env = map[string]string{}
		c.current = function
		c.declaredReturn = typeName(function.ReturnType)
		c.declaredEffects = map[string]bool{}
		for _, effect := range function.Effects {
			c.declaredEffects[effect] = true
		}
		for i, param := range function.Params {
			typ := Unknown
			if i < len(function.ParamTypes) && function.ParamTypes[i].Name != "" {
				typ = typeName(function.ParamTypes[i])
			}
			c.env[param] = typ
			if i < len(function.ParamTypes) && function.ParamTypes[i].Name != "" && !knownType(function.ParamTypes[i], types) {
				c.report("unknown-type", fmt.Sprintf("parameter %q uses unknown type %s", param, typeName(function.ParamTypes[i])), diagnostic.Span{})
			}
		}
		if function.ReturnType.Name != "" && !knownType(function.ReturnType, types) {
			c.report("unknown-type", fmt.Sprintf("function %q uses unknown return type %s", function.Name, typeName(function.ReturnType)), diagnostic.Span{})
		}
		c.returnTypes = nil
		c.checkBody(function.Body)
		c.reportInconsistentReturns(c.returnTypes)
		if c.declaredReturn != "" {
			for _, returnType := range c.returnTypes {
				if !compatible(returnType, c.declaredReturn) {
					c.report("return-type", fmt.Sprintf("function %q returns %s, declared %s", function.Name, returnType, c.declaredReturn), diagnostic.Span{})
					break
				}
			}
		}
	}
	return c.diagnostics
}

func knownType(typ parser.TypeRef, declarations map[string]parser.TypeDecl) bool {
	if typ.Name == "int" || typ.Name == "bool" || typ.Name == "string" || typ.Name == "none" || typ.Name == "record" {
		return len(typ.Args) == 0
	}
	if typ.Name == "Result" || typ.Name == "list" {
		if typ.Name == "Result" && len(typ.Args) != 2 || typ.Name == "list" && len(typ.Args) != 1 {
			return false
		}
		for _, arg := range typ.Args {
			if !knownType(arg, declarations) {
				return false
			}
		}
		return true
	}
	return len(typ.Args) == 0 && declarations[typ.Name].Name != ""
}

func typeName(typ parser.TypeRef) string {
	if typ.Name == "" {
		return ""
	}
	if len(typ.Args) == 0 {
		return typ.Name
	}
	args := make([]string, len(typ.Args))
	for i, arg := range typ.Args {
		args[i] = typeName(arg)
	}
	return typ.Name + "<" + strings.Join(args, ",") + ">"
}

func compatible(actual, expected string) bool {
	return actual == Unknown || expected == Unknown || actual == expected
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
	env             map[string]string
	functions       map[string]parser.Function
	types           map[string]parser.TypeDecl
	capabilities    map[string]bool
	diagnostics     []diagnostic.Diagnostic
	returnTypes     []string
	inferred        map[string]string
	inferring       map[string]bool
	silent          bool
	current         parser.Function
	declaredReturn  string
	declaredEffects map[string]bool
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
	if function.ReturnType.Name != "" {
		if len(function.ParamTypes) == len(function.Params) {
			for i, paramType := range function.ParamTypes {
				if paramType.Name != "" && i < len(args) && !compatible(args[i], typeName(paramType)) {
					c.report("argument-type", fmt.Sprintf("argument %d to %s requires %s, got %s", i+1, name, typeName(paramType), args[i]), diagnostic.Span{})
				}
			}
		}
		return typeName(function.ReturnType)
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
		objectType := c.expr(e.Object)
		if decl, ok := c.types[objectType]; ok {
			field, exists := decl.Fields[e.Name]
			if !exists {
				c.report("unknown-field", fmt.Sprintf("type %q has no field %q", objectType, e.Name), span)
				return Unknown
			}
			return typeName(field)
		}
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
		if requiredCapability != "" && c.current.Effects != nil && !c.declaredEffects[requiredCapability] {
			c.report("effect", fmt.Sprintf("function %q must declare effects(%s) for %s", c.current.Name, requiredCapability, e.Name), span)
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
		case "infer", "invoke", "ok", "err", "is_ok":
			if (e.Name == "infer" || e.Name == "ok" || e.Name == "err") && c.declaredReturn != "" {
				if strings.HasPrefix(c.declaredReturn, "Result<") {
					return c.declaredReturn
				}
				if e.Name == "infer" {
					if _, named := c.types[c.declaredReturn]; named {
						return c.declaredReturn
					}
				}
			}
			if e.Name == "is_ok" {
				return Bool
			}
			return Record
		case "unwrap":
			return Unknown
		case "recall":
			return List
		case "json_decode":
			return Unknown
		case "contains", "starts_with", "ends_with", "matches":
			return Bool
		}
		if _, ok := c.functions[e.Name]; ok {
			for _, effect := range c.functions[e.Name].Effects {
				if c.current.Effects != nil && !c.declaredEffects[effect] {
					c.report("effect", fmt.Sprintf("function %q must declare effects(%s) for call to %s", c.current.Name, effect, e.Name), span)
				}
			}
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
	case "len", "text", "number", "contains", "starts_with", "ends_with", "matches", "json_encode", "json_decode", "read_file", "write_file", "run", "now", "sleep", "map", "filter", "infer", "recall", "invoke", "ok", "err", "is_ok", "unwrap":
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
