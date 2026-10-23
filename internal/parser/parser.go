package parser

import (
	"encoding/json"
	"fmt"
	"github.com/fernandocpaz/axiom/internal/diagnostic"
	"github.com/fernandocpaz/axiom/internal/lexer"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Value struct {
	Kind   string
	Int    int
	Bool   bool
	Str    string
	List   []Value
	Fields map[string]Value
}

func (v Value) String() string {
	if v.Kind == "bool" {
		if v.Bool {
			return "true"
		}
		return "false"
	}
	if v.Kind == "string" {
		return v.Str
	}
	if v.Kind == "none" {
		return "none"
	}
	if v.Kind == "list" {
		parts := make([]string, len(v.List))
		for i, item := range v.List {
			parts[i] = item.String()
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	if v.Kind == "record" {
		parts := make([]string, 0, len(v.Fields))
		for key, item := range v.Fields {
			parts = append(parts, key+": "+item.String())
		}
		sort.Strings(parts)
		return "record {" + strings.Join(parts, ", ") + "}"
	}
	return strconv.Itoa(v.Int)
}

type Expr interface {
	eval(*Runtime) (Value, error)
}

// Located preserves source information while keeping expression evaluation small.
type Located struct {
	Inner Expr
	Span  diagnostic.Span
}

func (e Located) eval(rt *Runtime) (Value, error) { return e.Inner.eval(rt) }

type Number int
type String string
type Boolean bool
type None struct{}
type Variable string
type Binary struct {
	Op          lexer.Kind
	Left, Right Expr
}
type Unary struct {
	Op    lexer.Kind
	Value Expr
}
type Conditional struct{ Test, Yes, No Expr }
type Call struct {
	Name string
	Args []Expr
}
type ListLiteral []Expr
type RecordLiteral map[string]Expr
type Field struct {
	Object Expr
	Name   string
}
type Index struct {
	Object Expr
	Key    Expr
}

func (n Number) eval(*Runtime) (Value, error) { return Value{Kind: "int", Int: int(n)}, nil }
func (s String) eval(*Runtime) (Value, error) {
	return Value{Kind: "string", Str: string(s)}, nil
}
func (b Boolean) eval(*Runtime) (Value, error) {
	return Value{Kind: "bool", Bool: bool(b)}, nil
}
func (None) eval(*Runtime) (Value, error) { return Value{Kind: "none"}, nil }
func (v Variable) eval(rt *Runtime) (Value, error) {
	n, ok := rt.env[string(v)]
	if !ok {
		return Value{}, fmt.Errorf("undefined variable %q", v)
	}
	return n, nil
}
func (b Binary) eval(rt *Runtime) (Value, error) {
	l, e := b.Left.eval(rt)
	if e != nil {
		return Value{}, e
	}
	if b.Op == lexer.AndAnd || b.Op == lexer.OrOr {
		if l.Kind != "bool" {
			return Value{}, fmt.Errorf("logical operator requires boolean operands")
		}
		if b.Op == lexer.AndAnd && !l.Bool {
			return Value{Kind: "bool", Bool: false}, nil
		}
		if b.Op == lexer.OrOr && l.Bool {
			return Value{Kind: "bool", Bool: true}, nil
		}
	}
	r, e := b.Right.eval(rt)
	if e != nil {
		return Value{}, e
	}
	if b.Op == lexer.EqualEqual || b.Op == lexer.NotEqual {
		if l.Kind != r.Kind {
			return Value{}, fmt.Errorf("cannot compare %s and %s", l.Kind, r.Kind)
		}
		eq := (l.Kind == "int" && l.Int == r.Int) || (l.Kind == "bool" && l.Bool == r.Bool) || (l.Kind == "string" && l.Str == r.Str)
		if b.Op == lexer.NotEqual {
			eq = !eq
		}
		return Value{Kind: "bool", Bool: eq}, nil
	}
	if b.Op == lexer.AndAnd || b.Op == lexer.OrOr {
		if r.Kind != "bool" {
			return Value{}, fmt.Errorf("logical operator requires boolean operands")
		}
		if b.Op == lexer.AndAnd {
			return Value{Kind: "bool", Bool: l.Bool && r.Bool}, nil
		}
		return Value{Kind: "bool", Bool: l.Bool || r.Bool}, nil
	}
	if b.Op == lexer.Plus && l.Kind == "string" && r.Kind == "string" {
		return Value{Kind: "string", Str: l.Str + r.Str}, nil
	}
	if l.Kind != "int" || r.Kind != "int" {
		return Value{}, fmt.Errorf("operator requires integers")
	}
	switch b.Op {
	case lexer.AndAnd:
		if r.Kind != "bool" {
			return Value{}, fmt.Errorf("logical operator requires boolean operands")
		}
		return Value{Kind: "bool", Bool: l.Bool && r.Bool}, nil
	case lexer.OrOr:
		if r.Kind != "bool" {
			return Value{}, fmt.Errorf("logical operator requires boolean operands")
		}
		return Value{Kind: "bool", Bool: l.Bool || r.Bool}, nil
	case lexer.Plus:
		return Value{Kind: "int", Int: l.Int + r.Int}, nil
	case lexer.Minus:
		return Value{Kind: "int", Int: l.Int - r.Int}, nil
	case lexer.Star:
		return Value{Kind: "int", Int: l.Int * r.Int}, nil
	case lexer.Slash:
		if r.Int == 0 {
			return Value{}, fmt.Errorf("division by zero")
		}
		return Value{Kind: "int", Int: l.Int / r.Int}, nil
	case lexer.Less:
		return Value{Kind: "bool", Bool: l.Int < r.Int}, nil
	case lexer.LessEqual:
		return Value{Kind: "bool", Bool: l.Int <= r.Int}, nil
	case lexer.Greater:
		return Value{Kind: "bool", Bool: l.Int > r.Int}, nil
	case lexer.GreaterEqual:
		return Value{Kind: "bool", Bool: l.Int >= r.Int}, nil
	}
	return Value{}, fmt.Errorf("unknown operator")
}

func (u Unary) eval(rt *Runtime) (Value, error) {
	value, err := u.Value.eval(rt)
	if err != nil {
		return Value{}, err
	}
	switch u.Op {
	case lexer.Not:
		if value.Kind != "bool" {
			return Value{}, fmt.Errorf("operator ! requires a boolean")
		}
		return Value{Kind: "bool", Bool: !value.Bool}, nil
	case lexer.Minus:
		if value.Kind != "int" {
			return Value{}, fmt.Errorf("unary - requires an integer")
		}
		return Value{Kind: "int", Int: -value.Int}, nil
	case lexer.Plus:
		if value.Kind != "int" {
			return Value{}, fmt.Errorf("unary + requires an integer")
		}
		return value, nil
	default:
		return Value{}, fmt.Errorf("unknown unary operator")
	}
}
func (c Conditional) eval(rt *Runtime) (Value, error) {
	t, e := c.Test.eval(rt)
	if e != nil {
		return Value{}, e
	}
	if t.Kind != "bool" {
		return Value{}, fmt.Errorf("if condition requires a boolean")
	}
	if t.Bool {
		return c.Yes.eval(rt)
	}
	return c.No.eval(rt)
}

func (c Call) eval(rt *Runtime) (Value, error) {
	args := make([]Value, len(c.Args))
	for i, arg := range c.Args {
		value, err := arg.eval(rt)
		if err != nil {
			return Value{}, err
		}
		args[i] = value
	}
	return rt.call(c.Name, args)
}

func (l ListLiteral) eval(rt *Runtime) (Value, error) {
	values := make([]Value, len(l))
	for i, item := range l {
		value, err := item.eval(rt)
		if err != nil {
			return Value{}, err
		}
		values[i] = value
	}
	return Value{Kind: "list", List: values}, nil
}

func (r RecordLiteral) eval(rt *Runtime) (Value, error) {
	values := make(map[string]Value, len(r))
	for name, item := range r {
		value, err := item.eval(rt)
		if err != nil {
			return Value{}, err
		}
		values[name] = value
	}
	return Value{Kind: "record", Fields: values}, nil
}

func (f Field) eval(rt *Runtime) (Value, error) {
	object, err := f.Object.eval(rt)
	if err != nil {
		return Value{}, err
	}
	if object.Kind != "record" {
		return Value{}, fmt.Errorf("cannot access field %q on %s", f.Name, object.Kind)
	}
	value, ok := object.Fields[f.Name]
	if !ok {
		return Value{}, fmt.Errorf("record has no field %q", f.Name)
	}
	return value, nil
}

func (i Index) eval(rt *Runtime) (Value, error) {
	object, err := i.Object.eval(rt)
	if err != nil {
		return Value{}, err
	}
	key, err := i.Key.eval(rt)
	if err != nil {
		return Value{}, err
	}
	if object.Kind == "list" && key.Kind == "int" {
		if key.Int < 0 || key.Int >= len(object.List) {
			return Value{}, fmt.Errorf("list index %d out of range", key.Int)
		}
		return object.List[key.Int], nil
	}
	if object.Kind == "record" && key.Kind == "string" {
		value, ok := object.Fields[key.Str]
		if !ok {
			return Value{}, fmt.Errorf("record has no field %q", key.Str)
		}
		return value, nil
	}
	return Value{}, fmt.Errorf("invalid index on %s", object.Kind)
}

type Stmt struct {
	Let       string
	Value     Expr
	Set       string
	Print     Expr
	Return    bool
	Result    Expr
	While     Expr
	Body      []Stmt
	Break     bool
	Continue  bool
	ForVar    string
	ForExpr   Expr
	TryBody   []Stmt
	CatchName string
	CatchBody []Stmt
}
type Function struct {
	Name   string
	Params []string
	Body   []Stmt
}
type Program struct {
	Body      []Stmt
	Functions []Function
	Imports   []string
}
type Parser struct {
	tokens []lexer.Token
	pos    int
}

func Parse(ts []lexer.Token) (Program, error) {
	p := &Parser{tokens: ts}
	var functions []Function
	var imports []string
	for p.peek().Kind != lexer.EOF {
		if p.peek().Kind == lexer.Import {
			p.pos++
			path := p.peek()
			if path.Kind != lexer.String {
				return Program{}, p.error("expected import path")
			}
			imports = append(imports, path.Text)
			p.pos++
			continue
		}
		f, e := p.function()
		if e != nil {
			return Program{}, e
		}
		functions = append(functions, f)
	}
	if e := p.expect(lexer.EOF); e != nil {
		return Program{}, e
	}
	for _, f := range functions {
		if f.Name == "main" {
			return Program{Body: f.Body, Functions: functions, Imports: imports}, nil
		}
	}
	return Program{Functions: functions, Imports: imports}, nil
}

func (p *Parser) function() (Function, error) {
	if e := p.expect(lexer.Fn); e != nil {
		return Function{}, e
	}
	n := p.peek()
	if n.Kind != lexer.Ident {
		return Function{}, p.error("expected function name")
	}
	p.pos++
	if e := p.expect(lexer.LParen); e != nil {
		return Function{}, e
	}
	var params []string
	if p.peek().Kind != lexer.RParen {
		for {
			param := p.peek()
			if param.Kind != lexer.Ident {
				return Function{}, p.error("expected parameter name")
			}
			params = append(params, param.Text)
			p.pos++
			if p.peek().Kind != lexer.Comma {
				break
			}
			p.pos++
		}
	}
	if e := p.expect(lexer.RParen); e != nil {
		return Function{}, e
	}
	if e := p.expect(lexer.LBrace); e != nil {
		return Function{}, e
	}
	body, e := p.statements()
	if e != nil {
		return Function{}, e
	}
	return Function{Name: n.Text, Params: params, Body: body}, nil
}

func (p *Parser) statements() ([]Stmt, error) {
	var body []Stmt
	for p.peek().Kind != lexer.RBrace && p.peek().Kind != lexer.EOF {
		s, e := p.stmt()
		if e != nil {
			return nil, e
		}
		body = append(body, s)
	}
	if e := p.expect(lexer.RBrace); e != nil {
		return nil, e
	}
	return body, nil
}
func (p *Parser) peek() lexer.Token { return p.tokens[p.pos] }
func (p *Parser) error(msg string) error {
	t := p.peek()
	return fmt.Errorf("%d:%d: %s", t.Line, t.Col, msg)
}
func (p *Parser) expect(k lexer.Kind) error {
	if p.peek().Kind != k {
		return p.error(fmt.Sprintf("expected token %v, got %v", k, p.peek().Kind))
	}
	p.pos++
	return nil
}
func (p *Parser) stmt() (Stmt, error) {
	t := p.peek()
	switch t.Kind {
	case lexer.Let:
		p.pos++
		n := p.peek()
		if n.Kind != lexer.Ident {
			return Stmt{}, p.error("expected binding name")
		}
		p.pos++
		if e := p.expect(lexer.Equal); e != nil {
			return Stmt{}, e
		}
		x, e := p.expr(0)
		return Stmt{Let: n.Text, Value: x}, e
	case lexer.Set:
		p.pos++
		name := p.peek()
		if name.Kind != lexer.Ident {
			return Stmt{}, p.error("expected binding name")
		}
		p.pos++
		if e := p.expect(lexer.Equal); e != nil {
			return Stmt{}, e
		}
		x, e := p.expr(0)
		return Stmt{Set: name.Text, Value: x}, e
	case lexer.Print:
		p.pos++
		x, e := p.expr(0)
		return Stmt{Print: x}, e
	case lexer.Return:
		p.pos++
		x, e := p.expr(0)
		return Stmt{Return: true, Result: x}, e
	case lexer.While:
		p.pos++
		condition, e := p.expr(0)
		if e != nil {
			return Stmt{}, e
		}
		if e = p.expect(lexer.LBrace); e != nil {
			return Stmt{}, e
		}
		body, e := p.statements()
		return Stmt{While: condition, Body: body}, e
	case lexer.For:
		p.pos++
		name := p.peek()
		if name.Kind != lexer.Ident {
			return Stmt{}, p.error("expected loop variable")
		}
		p.pos++
		if e := p.expect(lexer.In); e != nil {
			return Stmt{}, e
		}
		collection, e := p.expr(0)
		if e != nil {
			return Stmt{}, e
		}
		if e = p.expect(lexer.LBrace); e != nil {
			return Stmt{}, e
		}
		body, e := p.statements()
		return Stmt{ForVar: name.Text, ForExpr: collection, Body: body}, e
	case lexer.Try:
		p.pos++
		if e := p.expect(lexer.LBrace); e != nil {
			return Stmt{}, e
		}
		tryBody, e := p.statements()
		if e != nil {
			return Stmt{}, e
		}
		if e = p.expect(lexer.Catch); e != nil {
			return Stmt{}, e
		}
		name := p.peek()
		if name.Kind != lexer.Ident {
			return Stmt{}, p.error("expected catch binding")
		}
		p.pos++
		if e = p.expect(lexer.LBrace); e != nil {
			return Stmt{}, e
		}
		catchBody, e := p.statements()
		return Stmt{TryBody: tryBody, CatchName: name.Text, CatchBody: catchBody}, e
	case lexer.Break:
		p.pos++
		return Stmt{Break: true}, nil
	case lexer.Continue:
		p.pos++
		return Stmt{Continue: true}, nil
	}
	return Stmt{}, p.error("expected let or print")
}
func precedence(k lexer.Kind) int {
	switch k {
	case lexer.OrOr:
		return 0
	case lexer.AndAnd:
		return 1
	case lexer.EqualEqual, lexer.NotEqual:
		return 2
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return 3
	case lexer.Plus, lexer.Minus:
		return 4
	case lexer.Star, lexer.Slash:
		return 5
	}
	return -1
}
func (p *Parser) expr(min int) (Expr, error) {
	t := p.peek()
	var left Expr
	switch t.Kind {
	case lexer.Plus, lexer.Minus, lexer.Not:
		p.pos++
		value, e := p.expr(6)
		if e != nil {
			return nil, e
		}
		left = Located{Inner: Unary{Op: t.Kind, Value: value}, Span: spanBetween(tokenSpan(t), exprSpan(value))}
	case lexer.Number:
		n, _ := strconv.Atoi(t.Text)
		left = Located{Inner: Number(n), Span: tokenSpan(t)}
		p.pos++
	case lexer.String:
		left = Located{Inner: String(t.Text), Span: tokenSpan(t)}
		p.pos++
	case lexer.True:
		left = Located{Inner: Boolean(true), Span: tokenSpan(t)}
		p.pos++
	case lexer.False:
		left = Located{Inner: Boolean(false), Span: tokenSpan(t)}
		p.pos++
	case lexer.None:
		left = Located{Inner: None{}, Span: tokenSpan(t)}
		p.pos++
	case lexer.Ident:
		p.pos++
		if p.peek().Kind != lexer.LParen {
			left = Located{Inner: Variable(t.Text), Span: tokenSpan(t)}
			break
		}
		p.pos++
		var args []Expr
		if p.peek().Kind != lexer.RParen {
			for {
				arg, e := p.expr(0)
				if e != nil {
					return nil, e
				}
				args = append(args, arg)
				if p.peek().Kind != lexer.Comma {
					break
				}
				p.pos++
			}
		}
		if e := p.expect(lexer.RParen); e != nil {
			return nil, e
		}
		left = Located{Inner: Call{Name: t.Text, Args: args}, Span: spanFrom(t, p.tokens[p.pos-1])}
	case lexer.LParen:
		p.pos++
		var e error
		left, e = p.expr(0)
		if e != nil {
			return nil, e
		}
		if e = p.expect(lexer.RParen); e != nil {
			return nil, e
		}
	case lexer.LBracket:
		p.pos++
		var items []Expr
		if p.peek().Kind != lexer.RBracket {
			for {
				item, e := p.expr(0)
				if e != nil {
					return nil, e
				}
				items = append(items, item)
				if p.peek().Kind != lexer.Comma {
					break
				}
				p.pos++
			}
		}
		if e := p.expect(lexer.RBracket); e != nil {
			return nil, e
		}
		left = Located{Inner: ListLiteral(items), Span: spanFrom(t, p.tokens[p.pos-1])}
	case lexer.Record:
		p.pos++
		if e := p.expect(lexer.LBrace); e != nil {
			return nil, e
		}
		fields := map[string]Expr{}
		if p.peek().Kind != lexer.RBrace {
			for {
				name := p.peek()
				if name.Kind != lexer.Ident && name.Kind != lexer.String {
					return nil, p.error("expected record field name")
				}
				p.pos++
				if e := p.expect(lexer.Colon); e != nil {
					return nil, e
				}
				value, e := p.expr(0)
				if e != nil {
					return nil, e
				}
				fields[name.Text] = value
				if p.peek().Kind != lexer.Comma {
					break
				}
				p.pos++
			}
		}
		if e := p.expect(lexer.RBrace); e != nil {
			return nil, e
		}
		left = Located{Inner: RecordLiteral(fields), Span: spanFrom(t, p.tokens[p.pos-1])}
	case lexer.If:
		p.pos++
		test, e := p.expr(0)
		if e != nil {
			return nil, e
		}
		if e = p.expect(lexer.LBrace); e != nil {
			return nil, e
		}
		yes, e := p.expr(0)
		if e != nil {
			return nil, e
		}
		if e = p.expect(lexer.RBrace); e != nil {
			return nil, e
		}
		if e = p.expect(lexer.Else); e != nil {
			return nil, e
		}
		if e = p.expect(lexer.LBrace); e != nil {
			return nil, e
		}
		no, e := p.expr(0)
		if e != nil {
			return nil, e
		}
		if e = p.expect(lexer.RBrace); e != nil {
			return nil, e
		}
		left = Located{Inner: Conditional{Test: test, Yes: yes, No: no}, Span: spanFrom(t, p.tokens[p.pos-1])}
	default:
		return nil, p.error("expected expression")
	}
	for {
		switch p.peek().Kind {
		case lexer.Dot:
			p.pos++
			name := p.peek()
			if name.Kind != lexer.Ident {
				return nil, p.error("expected field name")
			}
			p.pos++
			left = Located{Inner: Field{Object: left, Name: name.Text}, Span: spanBetween(exprSpan(left), tokenSpan(name))}
		case lexer.LBracket:
			p.pos++
			key, e := p.expr(0)
			if e != nil {
				return nil, e
			}
			if e = p.expect(lexer.RBracket); e != nil {
				return nil, e
			}
			left = Located{Inner: Index{Object: left, Key: key}, Span: spanBetween(exprSpan(left), tokenSpan(p.tokens[p.pos-1]))}
		default:
			goto postfixDone
		}
	}
postfixDone:
	for {
		op := p.peek().Kind
		prec := precedence(op)
		if prec < min {
			break
		}
		p.pos++
		right, e := p.expr(prec + 1)
		if e != nil {
			return nil, e
		}
		left = Located{Inner: Binary{Op: op, Left: left, Right: right}, Span: spanBetween(exprSpan(left), exprSpan(right))}
	}
	return left, nil
}

func tokenSpan(t lexer.Token) diagnostic.Span {
	return diagnostic.Span{Line: t.Line, Column: t.Col, EndLine: t.Line, EndColumn: t.Col + len(t.Text)}
}

func exprSpan(expr Expr) diagnostic.Span {
	if located, ok := expr.(Located); ok {
		return located.Span
	}
	return diagnostic.Span{}
}

func spanFrom(start, end lexer.Token) diagnostic.Span {
	span := tokenSpan(start)
	endSpan := tokenSpan(end)
	span.EndLine, span.EndColumn = endSpan.EndLine, endSpan.EndColumn
	return span
}

func spanBetween(start, end diagnostic.Span) diagnostic.Span {
	if start.Line == 0 {
		return end
	}
	start.EndLine, start.EndColumn = end.EndLine, end.EndColumn
	return start
}

type Runtime struct {
	env   map[string]Value
	funcs map[string]Function
	out   func(Value)
	depth int
}

type flowKind int

const (
	flowNormal flowKind = iota
	flowReturn
	flowBreak
	flowContinue
)

func (rt *Runtime) call(name string, args []Value) (Value, error) {
	if value, ok, err := builtin(name, args); ok {
		return value, err
	}
	function, ok := rt.funcs[name]
	if !ok {
		return Value{}, fmt.Errorf("undefined function %q", name)
	}
	if len(args) != len(function.Params) {
		return Value{}, fmt.Errorf("function %q expects %d arguments, got %d", name, len(function.Params), len(args))
	}
	if rt.depth >= 256 {
		return Value{}, fmt.Errorf("maximum function call depth exceeded")
	}
	previousEnv := rt.env
	rt.env = map[string]Value{}
	for i, param := range function.Params {
		rt.env[param] = args[i]
	}
	rt.depth++
	defer func() {
		rt.depth--
		rt.env = previousEnv
	}()
	value, flow, err := rt.execute(function.Body)
	if err != nil {
		return Value{}, err
	}
	if flow == flowReturn {
		return value, nil
	}
	if flow == flowBreak || flow == flowContinue {
		return Value{}, fmt.Errorf("loop control used outside loop")
	}
	return Value{}, nil
}

func builtin(name string, args []Value) (Value, bool, error) {
	need := func(count int) error {
		if len(args) != count {
			return fmt.Errorf("builtin %q expects %d arguments, got %d", name, count, len(args))
		}
		return nil
	}
	switch name {
	case "len":
		if err := need(1); err != nil {
			return Value{}, true, err
		}
		switch args[0].Kind {
		case "string":
			return Value{Kind: "int", Int: len([]rune(args[0].Str))}, true, nil
		case "list":
			return Value{Kind: "int", Int: len(args[0].List)}, true, nil
		case "record":
			return Value{Kind: "int", Int: len(args[0].Fields)}, true, nil
		default:
			return Value{}, true, fmt.Errorf("len does not support %s", args[0].Kind)
		}
	case "text":
		if err := need(1); err != nil {
			return Value{}, true, err
		}
		return Value{Kind: "string", Str: args[0].String()}, true, nil
	case "number":
		if err := need(1); err != nil {
			return Value{}, true, err
		}
		if args[0].Kind == "int" {
			return args[0], true, nil
		}
		if args[0].Kind != "string" {
			return Value{}, true, fmt.Errorf("number requires text or int")
		}
		value, err := strconv.Atoi(strings.TrimSpace(args[0].Str))
		if err != nil {
			return Value{}, true, fmt.Errorf("invalid number %q", args[0].Str)
		}
		return Value{Kind: "int", Int: value}, true, nil
	case "contains", "starts_with", "ends_with":
		if err := need(2); err != nil {
			return Value{}, true, err
		}
		if args[0].Kind != "string" || args[1].Kind != "string" {
			return Value{}, true, fmt.Errorf("%s requires text arguments", name)
		}
		matched := false
		switch name {
		case "contains":
			matched = strings.Contains(args[0].Str, args[1].Str)
		case "starts_with":
			matched = strings.HasPrefix(args[0].Str, args[1].Str)
		case "ends_with":
			matched = strings.HasSuffix(args[0].Str, args[1].Str)
		}
		return Value{Kind: "bool", Bool: matched}, true, nil
	case "matches":
		if err := need(2); err != nil {
			return Value{}, true, err
		}
		if args[0].Kind != "string" || args[1].Kind != "string" {
			return Value{}, true, fmt.Errorf("matches requires text arguments")
		}
		matched, err := regexp.MatchString(args[1].Str, args[0].Str)
		if err != nil {
			return Value{}, true, fmt.Errorf("invalid regular expression: %w", err)
		}
		return Value{Kind: "bool", Bool: matched}, true, nil
	case "json_encode":
		if err := need(1); err != nil {
			return Value{}, true, err
		}
		encoded, err := json.Marshal(toJSON(args[0]))
		if err != nil {
			return Value{}, true, err
		}
		return Value{Kind: "string", Str: string(encoded)}, true, nil
	case "json_decode":
		if err := need(1); err != nil {
			return Value{}, true, err
		}
		if args[0].Kind != "string" {
			return Value{}, true, fmt.Errorf("json_decode requires text")
		}
		var decoded interface{}
		if err := json.Unmarshal([]byte(args[0].Str), &decoded); err != nil {
			return Value{}, true, fmt.Errorf("invalid JSON: %w", err)
		}
		value, err := fromJSON(decoded)
		return value, true, err
	}
	return Value{}, false, nil
}

func toJSON(value Value) interface{} {
	switch value.Kind {
	case "none":
		return nil
	case "int":
		return value.Int
	case "bool":
		return value.Bool
	case "string":
		return value.Str
	case "list":
		items := make([]interface{}, len(value.List))
		for i, item := range value.List {
			items[i] = toJSON(item)
		}
		return items
	case "record":
		fields := make(map[string]interface{}, len(value.Fields))
		for name, item := range value.Fields {
			fields[name] = toJSON(item)
		}
		return fields
	default:
		return nil
	}
}

func fromJSON(value interface{}) (Value, error) {
	switch value := value.(type) {
	case nil:
		return Value{Kind: "none"}, nil
	case bool:
		return Value{Kind: "bool", Bool: value}, nil
	case string:
		return Value{Kind: "string", Str: value}, nil
	case float64:
		if value != float64(int(value)) {
			return Value{}, fmt.Errorf("JSON number %v is not an integer", value)
		}
		return Value{Kind: "int", Int: int(value)}, nil
	case []interface{}:
		items := make([]Value, len(value))
		for i, item := range value {
			converted, err := fromJSON(item)
			if err != nil {
				return Value{}, err
			}
			items[i] = converted
		}
		return Value{Kind: "list", List: items}, nil
	case map[string]interface{}:
		fields := make(map[string]Value, len(value))
		for name, item := range value {
			converted, err := fromJSON(item)
			if err != nil {
				return Value{}, err
			}
			fields[name] = converted
		}
		return Value{Kind: "record", Fields: fields}, nil
	default:
		return Value{}, fmt.Errorf("unsupported JSON value")
	}
}

func (rt *Runtime) execute(body []Stmt) (Value, flowKind, error) {
	for _, stmt := range body {
		if stmt.Set != "" {
			value, err := stmt.Value.eval(rt)
			if err != nil {
				return Value{}, flowNormal, err
			}
			if _, ok := rt.env[stmt.Set]; !ok {
				return Value{}, flowNormal, fmt.Errorf("undefined variable %q", stmt.Set)
			}
			rt.env[stmt.Set] = value
			continue
		}
		if stmt.Let != "" {
			value, err := stmt.Value.eval(rt)
			if err != nil {
				return Value{}, flowNormal, err
			}
			rt.env[stmt.Let] = value
			continue
		}
		if stmt.While != nil {
			for {
				condition, err := stmt.While.eval(rt)
				if err != nil {
					return Value{}, flowNormal, err
				}
				if condition.Kind != "bool" {
					return Value{}, flowNormal, fmt.Errorf("while condition requires a boolean")
				}
				if !condition.Bool {
					break
				}
				value, flow, err := rt.execute(stmt.Body)
				if err != nil {
					return Value{}, flowNormal, err
				}
				if flow == flowReturn {
					return value, flowReturn, nil
				}
				if flow == flowBreak {
					break
				}
			}
			continue
		}
		if stmt.ForVar != "" {
			collection, err := stmt.ForExpr.eval(rt)
			if err != nil {
				return Value{}, flowNormal, err
			}
			var items []Value
			switch collection.Kind {
			case "list":
				items = collection.List
			case "string":
				for _, char := range []rune(collection.Str) {
					items = append(items, Value{Kind: "string", Str: string(char)})
				}
			default:
				return Value{}, flowNormal, fmt.Errorf("cannot iterate over %s", collection.Kind)
			}
			for _, item := range items {
				rt.env[stmt.ForVar] = item
				value, flow, err := rt.execute(stmt.Body)
				if err != nil {
					return Value{}, flowNormal, err
				}
				if flow == flowReturn {
					return value, flowReturn, nil
				}
				if flow == flowBreak {
					break
				}
			}
			continue
		}
		if stmt.TryBody != nil {
			_, flow, err := rt.execute(stmt.TryBody)
			if err == nil {
				if flow != flowNormal {
					return Value{}, flow, nil
				}
				continue
			}
			rt.env[stmt.CatchName] = Value{Kind: "string", Str: err.Error()}
			value, flow, catchErr := rt.execute(stmt.CatchBody)
			if catchErr != nil {
				return Value{}, flowNormal, catchErr
			}
			if flow != flowNormal {
				return value, flow, nil
			}
			continue
		}
		if stmt.Print != nil {
			value, err := stmt.Print.eval(rt)
			if err != nil {
				return Value{}, flowNormal, err
			}
			rt.out(value)
			continue
		}
		if stmt.Return {
			value, err := stmt.Result.eval(rt)
			if err != nil {
				return Value{}, flowNormal, err
			}
			return value, flowReturn, nil
		}
		if stmt.Break {
			return Value{}, flowBreak, nil
		}
		if stmt.Continue {
			return Value{}, flowContinue, nil
		}
	}
	return Value{}, flowNormal, nil
}

func (pr Program) Run(out func(Value)) error {
	functions := map[string]Function{}
	for _, function := range pr.Functions {
		functions[function.Name] = function
	}
	if _, ok := functions["main"]; !ok {
		functions["main"] = Function{Name: "main", Body: pr.Body}
	}
	rt := &Runtime{env: map[string]Value{}, funcs: functions, out: out}
	_, err := rt.call("main", nil)
	return err
}
