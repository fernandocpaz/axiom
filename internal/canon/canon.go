// Package canon defines the single canonical surface form of Axiom source.
//
// Axiom source is machine-owned: the canonical form is the only accepted shape
// of a .ax file. It is not designed to be read by people. The formatter is
// deterministic and idempotent, so any file can be compared byte for byte
// against Format to detect hand editing.
package canon

import (
	"sort"
	"strconv"
	"strings"

	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
)

// Format parses source and returns its canonical form.
func Format(src string) (string, error) {
	tokens, err := lexer.Scan(src)
	if err != nil {
		return "", err
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		return "", err
	}
	return Program(program), nil
}

// Program returns the canonical source of an already parsed program.
func Program(program parser.Program) string {
	w := &writer{}
	for _, path := range program.Imports {
		w.write("import")
		w.write(quote(path))
		w.write("\n")
	}
	for _, capability := range program.Capabilities {
		w.write("allow")
		w.write(quote(capability))
		w.write("\n")
	}
	functions := program.Functions
	if len(functions) == 0 && len(program.Body) > 0 {
		functions = []parser.Function{{Name: "main", Body: program.Body}}
	}
	for _, function := range functions {
		w.function(function)
		w.write("\n")
	}
	return w.b.String()
}

// Canonical reports the canonical form of src and whether src already matches it.
func Canonical(src string) (string, bool, error) {
	formatted, err := Format(src)
	if err != nil {
		return "", false, err
	}
	return formatted, formatted == src, nil
}

// writer emits tokens with the minimum separating whitespace the lexer needs.
// Two word characters would merge into one token, so only those boundaries get
// a space. Everything else is adjacent by construction.
type writer struct {
	b    strings.Builder
	last byte
}

func (w *writer) write(text string) {
	if text == "" {
		return
	}
	if isWord(w.last) && isWord(text[0]) {
		w.b.WriteByte(' ')
	}
	w.b.WriteString(text)
	w.last = text[len(text)-1]
}

func (w *writer) function(f parser.Function) {
	w.write("fn")
	w.write(f.Name)
	w.write("(")
	for i, param := range f.Params {
		if i > 0 {
			w.write(",")
		}
		w.write(param)
	}
	w.write(")")
	w.block(f.Body)
}

func (w *writer) block(body []parser.Stmt) {
	w.write("{")
	for _, stmt := range body {
		w.stmt(stmt)
	}
	w.write("}")
}

func (w *writer) stmt(s parser.Stmt) {
	switch {
	case s.Let != "":
		w.write("let")
		w.write(s.Let)
		w.write("=")
		w.expr(s.Value, 0)
	case s.Set != "":
		w.write("set")
		w.write(s.Set)
		w.write("=")
		w.expr(s.Value, 0)
	case s.Print != nil:
		w.write("print")
		w.expr(s.Print, 0)
	case s.Return:
		w.write("return")
		w.expr(s.Result, 0)
	case s.While != nil:
		w.write("while")
		w.expr(s.While, 0)
		w.block(s.Body)
	case s.ForVar != "":
		w.write("for")
		w.write(s.ForVar)
		w.write("in")
		w.expr(s.ForExpr, 0)
		w.block(s.Body)
	case s.TryBody != nil:
		w.write("try")
		w.block(s.TryBody)
		w.write("catch")
		w.write(s.CatchName)
		w.block(s.CatchBody)
	case s.Break:
		w.write("break")
	case s.Continue:
		w.write("continue")
	}
}

func (w *writer) expr(e parser.Expr, min int) {
	e = unwrap(e)
	p := precedenceOf(e)
	wrap := p < min
	if wrap {
		w.write("(")
	}
	switch x := e.(type) {
	case parser.Number:
		w.write(strconv.Itoa(int(x)))
	case parser.String:
		w.write(quote(string(x)))
	case parser.Boolean:
		if x {
			w.write("true")
		} else {
			w.write("false")
		}
	case parser.None:
		w.write("none")
	case parser.Variable:
		w.write(string(x))
	case parser.Unary:
		w.write(operator(x.Op))
		w.expr(x.Value, precProduct)
	case parser.Binary:
		op := precedence(x.Op)
		w.expr(x.Left, op)
		w.write(operator(x.Op))
		w.expr(x.Right, op+1)
	case parser.Pipeline:
		w.expr(x.Left, precPipe)
		w.write("|>")
		w.call(x.Name, x.Args)
	case parser.Conditional:
		w.write("if")
		w.expr(x.Test, 0)
		w.write("{")
		w.expr(x.Yes, 0)
		w.write("}")
		w.write("else")
		w.write("{")
		w.expr(x.No, 0)
		w.write("}")
	case parser.Call:
		w.call(x.Name, x.Args)
	case parser.Field:
		w.expr(x.Object, precPostfix)
		w.write(".")
		w.write(x.Name)
	case parser.Index:
		w.expr(x.Object, precPostfix)
		w.write("[")
		w.expr(x.Key, 0)
		w.write("]")
	case parser.ListLiteral:
		w.write("[")
		for i, item := range x {
			if i > 0 {
				w.write(",")
			}
			w.expr(item, 0)
		}
		w.write("]")
	case parser.RecordLiteral:
		names := make([]string, 0, len(x))
		for name := range x {
			names = append(names, name)
		}
		sort.Strings(names)
		w.write("record")
		w.write("{")
		for i, name := range names {
			if i > 0 {
				w.write(",")
			}
			w.write(fieldName(name))
			w.write(":")
			w.expr(x[name], 0)
		}
		w.write("}")
	}
	if wrap {
		w.write(")")
	}
}

func (w *writer) call(name string, args []parser.Expr) {
	w.write(name)
	w.write("(")
	for i, arg := range args {
		if i > 0 {
			w.write(",")
		}
		w.expr(arg, 0)
	}
	w.write(")")
}

const (
	precOr          = 0
	precPipe        = 1
	precAnd         = 2
	precEquality    = 3
	precComparison  = 4
	precSum         = 5
	precProduct     = 6
	precUnary       = 5
	precConditional = 7
	precPostfix     = 8
)

// precedence mirrors the expression precedence table in the parser.
func precedence(k lexer.Kind) int {
	switch k {
	case lexer.OrOr:
		return precOr
	case lexer.Pipe:
		return precPipe
	case lexer.AndAnd:
		return precAnd
	case lexer.EqualEqual, lexer.NotEqual:
		return precEquality
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return precComparison
	case lexer.Plus, lexer.Minus:
		return precSum
	case lexer.Star, lexer.Slash:
		return precProduct
	}
	return -1
}

func precedenceOf(e parser.Expr) int {
	switch x := e.(type) {
	case parser.Binary:
		return precedence(x.Op)
	case parser.Pipeline:
		return precPipe
	case parser.Unary:
		// A unary node swallows a product-level operand: `-a * b` parses as
		// -(a * b). It therefore binds looser than `*` when printed, and is
		// parenthesized anywhere a product is expected.
		return precUnary
	case parser.Conditional:
		return precConditional
	}
	return precPostfix
}

func unwrap(e parser.Expr) parser.Expr {
	for {
		located, ok := e.(parser.Located)
		if !ok {
			return e
		}
		e = located.Inner
	}
}

func operator(k lexer.Kind) string {
	switch k {
	case lexer.Plus:
		return "+"
	case lexer.Minus:
		return "-"
	case lexer.Star:
		return "*"
	case lexer.Slash:
		return "/"
	case lexer.EqualEqual:
		return "=="
	case lexer.NotEqual:
		return "!="
	case lexer.Less:
		return "<"
	case lexer.LessEqual:
		return "<="
	case lexer.Greater:
		return ">"
	case lexer.GreaterEqual:
		return ">="
	case lexer.AndAnd:
		return "&&"
	case lexer.OrOr:
		return "||"
	case lexer.Not:
		return "!"
	}
	return ""
}

func fieldName(name string) string {
	if isIdentifier(name) && !isKeyword(name) {
		return name
	}
	return quote(name)
}

func isIdentifier(name string) bool {
	if name == "" || !isLetter(name[0]) && name[0] != '_' {
		return false
	}
	for i := 1; i < len(name); i++ {
		if !isWord(name[i]) {
			return false
		}
	}
	return true
}

func isKeyword(name string) bool {
	switch name {
	case "fn", "let", "print", "true", "false", "if", "else", "return", "none",
		"record", "while", "break", "continue", "set", "for", "in", "try",
		"catch", "import", "allow":
		return true
	}
	return false
}

func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isWord(c byte) bool {
	return isLetter(c) || c >= '0' && c <= '9' || c == '_'
}

func quote(text string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '\\':
			b.WriteString("\\\\")
		case '"':
			b.WriteString("\\\"")
		case '\n':
			b.WriteString("\\n")
		case '\t':
			b.WriteString("\\t")
		default:
			b.WriteByte(text[i])
		}
	}
	b.WriteByte('"')
	return b.String()
}
