package parser

import (
	"fmt"
	"github.com/fernandocpaz/axiom/internal/lexer"
	"strconv"
)

type Expr interface {
	eval(map[string]int) (int, error)
}
type Number int
type Variable string
type Binary struct {
	Op          lexer.Kind
	Left, Right Expr
}

func (n Number) eval(map[string]int) (int, error) { return int(n), nil }
func (v Variable) eval(env map[string]int) (int, error) {
	n, ok := env[string(v)]
	if !ok {
		return 0, fmt.Errorf("undefined variable %q", v)
	}
	return n, nil
}
func (b Binary) eval(env map[string]int) (int, error) {
	l, e := b.Left.eval(env)
	if e != nil {
		return 0, e
	}
	r, e := b.Right.eval(env)
	if e != nil {
		return 0, e
	}
	switch b.Op {
	case lexer.Plus:
		return l + r, nil
	case lexer.Minus:
		return l - r, nil
	case lexer.Star:
		return l * r, nil
	case lexer.Slash:
		if r == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return l / r, nil
	}
	return 0, fmt.Errorf("unknown operator")
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
	if err := p.expect(lexer.Fn); err != nil {
		return Program{}, err
	}
	if err := p.expect(lexer.Ident); err != nil {
		return Program{}, err
	}
	if p.tokens[p.pos-1].Text != "main" {
		return Program{}, fmt.Errorf("expected main")
	}
	if err := p.expect(lexer.LParen); err != nil {
		return Program{}, err
	}
	if err := p.expect(lexer.RParen); err != nil {
		return Program{}, err
	}
	if err := p.expect(lexer.LBrace); err != nil {
		return Program{}, err
	}
	var body []Stmt
	for p.peek().Kind != lexer.RBrace && p.peek().Kind != lexer.EOF {
		s, e := p.stmt()
		if e != nil {
			return Program{}, e
		}
		body = append(body, s)
	}
	if err := p.expect(lexer.RBrace); err != nil {
		return Program{}, err
	}
	return Program{body}, nil
}
func (p *Parser) peek() lexer.Token { return p.tokens[p.pos] }
func (p *Parser) expect(k lexer.Kind) error {
	t := p.peek()
	if t.Kind != k {
		return fmt.Errorf("%d:%d: expected token %v, got %v", t.Line, t.Col, k, t.Kind)
	}
	p.pos++
	return nil
}
func (p *Parser) stmt() (Stmt, error) {
	t := p.peek()
	switch t.Kind {
	case lexer.Let:
		p.pos++
		name := p.peek()
		if name.Kind != lexer.Ident {
			return Stmt{}, fmt.Errorf("%d:%d: expected binding name", name.Line, name.Col)
		}
		p.pos++
		if e := p.expect(lexer.Equal); e != nil {
			return Stmt{}, e
		}
		x, e := p.expr(0)
		return Stmt{Let: name.Text, Value: x}, e
	case lexer.Print:
		p.pos++
		x, e := p.expr(0)
		return Stmt{Print: x}, e
	}
	return Stmt{}, fmt.Errorf("%d:%d: expected let or print", t.Line, t.Col)
}
func (p *Parser) expr(min int) (Expr, error) {
	t := p.peek()
	var left Expr
	switch t.Kind {
	case lexer.Number:
		n, _ := strconv.Atoi(t.Text)
		left = Number(n)
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
	default:
		return nil, fmt.Errorf("%d:%d: expected expression", t.Line, t.Col)
	}
	for {
		op := p.peek().Kind
		prec := -1
		if op == lexer.Plus || op == lexer.Minus {
			prec = 1
		}
		if op == lexer.Star || op == lexer.Slash {
			prec = 2
		}
		if prec < min {
			break
		}
		p.pos++
		right, e := p.expr(prec + 1)
		if e != nil {
			return nil, e
		}
		left = Binary{op, left, right}
	}
	return left, nil
}
func (pr Program) Run(out func(int)) {
	env := map[string]int{}
	for _, s := range pr.Body {
		if s.Let != "" {
			v, e := s.Value.eval(env)
			if e != nil {
				panic(e)
			}
			env[s.Let] = v
		} else {
			v, e := s.Print.eval(env)
			if e != nil {
				panic(e)
			}
			out(v)
		}
	}
}
