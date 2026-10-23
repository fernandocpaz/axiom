package lexer

import "fmt"

type Kind int

const (
	EOF Kind = iota
	Ident
	Number
	String
	Fn
	Let
	Print
	True
	False
	If
	Else
	LBrace
	RBrace
	LParen
	RParen
	Equal
	Plus
	Minus
	Star
	Slash
	EqualEqual
	NotEqual
	Less
	LessEqual
	Greater
	GreaterEqual
	Comma
	Return
	None
	Record
	Colon
	Dot
	LBracket
	RBracket
	While
	Break
	Continue
	AndAnd
	OrOr
	Not
	Set
	For
	In
	Try
	Catch
	Import
	Allow
	Pipe
)

type Token struct {
	Kind Kind
	Text string
	Line int
	Col  int
}

func (k Kind) String() string { return name(k) }

func (t Token) String() string {
	return fmt.Sprintf("%s(%q) at %d:%d", name(t.Kind), t.Text, t.Line, t.Col)
}

func name(k Kind) string {
	for n, v := range map[Kind]string{EOF: "eof", Ident: "ident", Number: "number", String: "string", Fn: "fn", Let: "let", Print: "print", True: "true", False: "false", If: "if", Else: "else", LBrace: "{", RBrace: "}", LParen: "(", RParen: ")", Equal: "=", Plus: "+", Minus: "-", Star: "*", Slash: "/", EqualEqual: "==", NotEqual: "!=", Less: "<", LessEqual: "<=", Greater: ">", GreaterEqual: ">=", Comma: ",", Return: "return", None: "none", Record: "record", Colon: ":", Dot: ".", LBracket: "[", RBracket: "]", While: "while", Break: "break", Continue: "continue", AndAnd: "&&", OrOr: "||", Not: "!", Set: "set", For: "for", In: "in", Try: "try", Catch: "catch", Import: "import", Allow: "allow", Pipe: "|>"} {
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
		if c == '"' {
			j := i + 1
			text := ""
			closed := false
			col++ // opening quote
			for j < len(src) {
				if src[j] == '"' {
					closed = true
					j++
					col++
					break
				}
				if src[j] == '\\' {
					if j+1 >= len(src) {
						return nil, fmt.Errorf("%d:%d: unterminated string", startLine, startCol)
					}
					switch src[j+1] {
					case 'n':
						text += "\n"
					case 't':
						text += "\t"
					case '"':
						text += "\""
					case '\\':
						text += "\\"
					default:
						return nil, fmt.Errorf("%d:%d: unknown string escape \\%c", line, col, src[j+1])
					}
					j += 2
					col += 2
					continue
				}
				if src[j] == '\n' {
					line++
					col = 1
				} else {
					col++
				}
				text += string(src[j])
				j++
			}
			if !closed {
				return nil, fmt.Errorf("%d:%d: unterminated string", startLine, startCol)
			}
			out = append(out, Token{String, text, startLine, startCol})
			i = j
			continue
		}
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
			if text == "true" {
				k = True
			}
			if text == "false" {
				k = False
			}
			if text == "if" {
				k = If
			}
			if text == "else" {
				k = Else
			}
			if text == "return" {
				k = Return
			}
			if text == "none" {
				k = None
			}
			if text == "record" {
				k = Record
			}
			if text == "while" {
				k = While
			}
			if text == "break" {
				k = Break
			}
			if text == "continue" {
				k = Continue
			}
			if text == "set" {
				k = Set
			}
			if text == "for" {
				k = For
			}
			if text == "in" {
				k = In
			}
			if text == "try" {
				k = Try
			}
			if text == "catch" {
				k = Catch
			}
			if text == "import" {
				k = Import
			}
			if text == "allow" {
				k = Allow
			}
			out = append(out, Token{k, text, startLine, startCol})
			col += j - i
			i = j
			continue
		}
		if i+1 < len(src) {
			pair := src[i : i+2]
			if pair == "==" {
				out = append(out, Token{EqualEqual, pair, line, col})
				i += 2
				col += 2
				continue
			}
			if pair == "!=" {
				out = append(out, Token{NotEqual, pair, line, col})
				i += 2
				col += 2
				continue
			}
			if pair == "<=" {
				out = append(out, Token{LessEqual, pair, line, col})
				i += 2
				col += 2
				continue
			}
			if pair == ">=" {
				out = append(out, Token{GreaterEqual, pair, line, col})
				i += 2
				col += 2
				continue
			}
			if pair == "&&" {
				out = append(out, Token{AndAnd, pair, line, col})
				i += 2
				col += 2
				continue
			}
			if pair == "||" {
				out = append(out, Token{OrOr, pair, line, col})
				i += 2
				col += 2
				continue
			}
			if pair == "|>" {
				out = append(out, Token{Pipe, pair, line, col})
				i += 2
				col += 2
				continue
			}
		}
		k := map[byte]Kind{'{': LBrace, '}': RBrace, '(': LParen, ')': RParen, ',': Comma, ':': Colon, '.': Dot, '[': LBracket, ']': RBracket, '=': Equal, '+': Plus, '-': Minus, '*': Star, '/': Slash, '<': Less, '>': Greater, '!': Not}[c]
		if k == 0 {
			return nil, fmt.Errorf("%d:%d: unexpected character %q", line, col, c)
		}
		out = append(out, Token{k, string(c), line, col})
		i++
		col++
	}
	return append(out, Token{Kind: EOF, Line: line, Col: col}), nil
}
