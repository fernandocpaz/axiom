package parser

import (
	"fmt"
	"github.com/fernandocpaz/axiom/internal/lexer"
	"strconv"
)

type Value struct {
	Kind string
	Int  int
	Bool bool
}

func (v Value) String() string {
	if v.Kind == "bool" {
		if v.Bool {
			return "true"
		}
		return "false"
	}
	return strconv.Itoa(v.Int)
}

type Expr interface {
	eval(map[string]Value) (Value, error)
}
type Number int
type Boolean bool
type Variable string
type Binary struct {
	Op          lexer.Kind
	Left, Right Expr
}
type Conditional struct{ Test, Yes, No Expr }

func (n Number) eval(map[string]Value) (Value, error) { return Value{Kind: "int", Int: int(n)}, nil }
func (b Boolean) eval(map[string]Value) (Value, error) {
	return Value{Kind: "bool", Bool: bool(b)}, nil
}
func (v Variable) eval(env map[string]Value) (Value, error) {
	n, ok := env[string(v)]
	if !ok {
		return Value{}, fmt.Errorf("undefined variable %q", v)
	}
	return n, nil
}
func (b Binary) eval(env map[string]Value) (Value, error) {
	l, e := b.Left.eval(env)
	if e != nil {
		return Value{}, e
	}
	r, e := b.Right.eval(env)
	if e != nil {
		return Value{}, e
	}
	if b.Op == lexer.EqualEqual || b.Op == lexer.NotEqual {
		if l.Kind != r.Kind {
			return Value{}, fmt.Errorf("cannot compare %s and %s", l.Kind, r.Kind)
		}
		eq := (l.Kind == "int" && l.Int == r.Int) || (l.Kind == "bool" && l.Bool == r.Bool)
		if b.Op == lexer.NotEqual {
			eq = !eq
		}
		return Value{Kind: "bool", Bool: eq}, nil
	}
	if l.Kind != "int" || r.Kind != "int" {
		return Value{}, fmt.Errorf("operator requires integers")
	}
	switch b.Op {
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
func (c Conditional) eval(env map[string]Value) (Value, error) {
	t, e := c.Test.eval(env)
	if e != nil {
		return Value{}, e
	}
	if t.Kind != "bool" {
		return Value{}, fmt.Errorf("if condition requires a boolean")
	}
	if t.Bool {
		return c.Yes.eval(env)
	}
	return c.No.eval(env)
}

type Stmt struct {
	Let   string
	Value Expr
	Print Expr
}
type Program struct{ Body []Stmt }
type Parser struct {
	tokens []lexer.Token
	pos    int
}

func Parse(ts []lexer.Token) (Program, error) {
	p := &Parser{tokens: ts}
	if e := p.expect(lexer.Fn); e != nil {
		return Program{}, e
	}
	n := p.peek()
	if n.Kind != lexer.Ident || n.Text != "main" {
		return Program{}, p.error("expected main")
	}
	p.pos++
	if e := p.expect(lexer.LParen); e != nil {
		return Program{}, e
	}
	if e := p.expect(lexer.RParen); e != nil {
		return Program{}, e
	}
	if e := p.expect(lexer.LBrace); e != nil {
		return Program{}, e
	}
	var body []Stmt
	for p.peek().Kind != lexer.RBrace && p.peek().Kind != lexer.EOF {
		s, e := p.stmt()
		if e != nil {
			return Program{}, e
		}
		body = append(body, s)
	}
	if e := p.expect(lexer.RBrace); e != nil {
		return Program{}, e
	}
	return Program{Body: body}, nil
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
	case lexer.Print:
		p.pos++
		x, e := p.expr(0)
		return Stmt{Print: x}, e
	}
	return Stmt{}, p.error("expected let or print")
}
func precedence(k lexer.Kind) int {
	switch k {
	case lexer.EqualEqual, lexer.NotEqual:
		return 1
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return 2
	case lexer.Plus, lexer.Minus:
		return 3
	case lexer.Star, lexer.Slash:
		return 4
	}
	return -1
}
func (p *Parser) expr(min int) (Expr, error) {
	t := p.peek()
	var left Expr
	switch t.Kind {
	case lexer.Number:
		n, _ := strconv.Atoi(t.Text)
		left = Number(n)
		p.pos++
	case lexer.True:
		left = Boolean(true)
		p.pos++
	case lexer.False:
		left = Boolean(false)
		p.pos++
	case lexer.Ident:
		left = Variable(t.Text)
		p.pos++
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
		left = Conditional{Test: test, Yes: yes, No: no}
	default:
		return nil, p.error("expected expression")
	}
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
		left = Binary{Op: op, Left: left, Right: right}
	}
	return left, nil
}
func (pr Program) Run(out func(Value)) error {
	env := map[string]Value{}
	for _, s := range pr.Body {
		if s.Let != "" {
			v, e := s.Value.eval(env)
			if e != nil {
				return e
			}
			env[s.Let] = v
		} else {
			v, e := s.Print.eval(env)
			if e != nil {
				return e
			}
			out(v)
		}
	}
	return nil
}
