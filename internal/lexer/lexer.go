package lexer

import "fmt"

type Kind int

const (
	EOF Kind = iota
	Ident
	Number
	Fn
	Let
	Print
	LBrace
	RBrace
	LParen
	RParen
	Equal
	Plus
	Minus
	Star
	Slash
)

type Token struct {
	Kind Kind
	Text string
	Line int
	Col  int
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q) at %d:%d", name(t.Kind), t.Text, t.Line, t.Col)
}

func name(k Kind) string {
	for n, v := range map[Kind]string{EOF: "eof", Ident: "ident", Number: "number", Fn: "fn", Let: "let", Print: "print", LBrace: "{", RBrace: "}", LParen: "(", RParen: ")", Equal: "=", Plus: "+", Minus: "-", Star: "*", Slash: "/"} {
		if v != "" && k == n {
			return v
		}
	}
	return "token"
}

func Scan(src string) ([]Token, error) {
	var out []Token
	line, col := 1, 1
	for i := 0; i < len(src); {
		c := src[i]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			if c == '\n' {
				line++
				col = 1
			} else {
				col++
			}
			i++
			continue
		}
		startLine, startCol := line, col
		if c >= '0' && c <= '9' {
			j := i
			for j < len(src) && src[j] >= '0' && src[j] <= '9' {
				j++
			}
			out = append(out, Token{Number, src[i:j], startLine, startCol})
			col += j - i
			i = j
			continue
		}
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
			j := i + 1
			for j < len(src) && ((src[j] >= 'a' && src[j] <= 'z') || (src[j] >= 'A' && src[j] <= 'Z') || (src[j] >= '0' && src[j] <= '9') || src[j] == '_') {
				j++
			}
			text := src[i:j]
			k := Ident
			if text == "fn" {
				k = Fn
			}
			if text == "let" {
				k = Let
			}
			if text == "print" {
				k = Print
			}
			out = append(out, Token{k, text, startLine, startCol})
			col += j - i
			i = j
			continue
		}
		k := map[byte]Kind{'{': LBrace, '}': RBrace, '(': LParen, ')': RParen, '=': Equal, '+': Plus, '-': Minus, '*': Star, '/': Slash}[c]
		if k == 0 {
			return nil, fmt.Errorf("%d:%d: unexpected character %q", line, col, c)
		}
		out = append(out, Token{k, string(c), line, col})
		i++
		col++
	}
	return append(out, Token{Kind: EOF, Line: line, Col: col}), nil
}
