package canon

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
)

func TestFormat(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "hello",
			src:  "fn main() {\n  print \"Hello, world!\"\n}\n",
			want: "fn main(){print\"Hello, world!\"}\n",
		},
		{
			name: "let and call",
			src:  "fn main() {\n  let x = 2 * (20 + 1)\n  print x\n}\n",
			want: "fn main(){let x=2*(20+1)print x}\n",
		},
		{
			name: "right nested subtraction keeps parentheses",
			src:  "fn main() { print 1 - (2 - 3) }",
			want: "fn main(){print 1-(2-3)}\n",
		},
		{
			name: "left nested subtraction drops parentheses",
			src:  "fn main() { print (1 - 2) - 3 }",
			want: "fn main(){print 1-2-3}\n",
		},
		{
			name: "unary as product operand is parenthesized",
			src:  "fn main() { print (-x) * y }",
			want: "fn main(){print(-x)*y}\n",
		},
		{
			name: "unary absorbs product operand",
			src:  "fn main() { print -x * y }",
			want: "fn main(){print-x*y}\n",
		},
		{
			name: "record fields are sorted",
			src:  "fn main() { print record { b: 1, a: \"x\" } }",
			want: "fn main(){print record{a:\"x\",b:1}}\n",
		},
		{
			name: "record field that is a keyword is quoted",
			src:  "fn main() { print record { \"if\": 1 } }",
			want: "fn main(){print record{\"if\":1}}\n",
		},
		{
			name: "conditional expression",
			src:  "fn main() { print if true { 1 } else { 2 } }",
			want: "fn main(){print if true{1}else{2}}\n",
		},
		{
			name: "control flow",
			src:  "fn main() { while true { break } for x in [1, 2] { continue } try { print 1 } catch e { print e } }",
			want: "fn main(){while true{break}for x in[1,2]{continue}try{print 1}catch e{print e}}\n",
		},
		{
			name: "strings are escaped once",
			src:  "fn main() { print \"a\\nb\\t\\\"c\\\"\" }",
			want: "fn main(){print\"a\\nb\\t\\\"c\\\"\"}\n",
		},
		{
			name: "imports and capabilities lead",
			src:  "allow \"x\" import \"m.ax\" fn main() { }",
			want: "import\"m.ax\"\nallow\"x\"\nfn main(){}\n",
		},
		{
			name: "pipeline",
			src:  "fn main() { let xs = [1, 2] |> map(double) |> filter(keep) }",
			want: "fn main(){let xs=[1,2]|>map(double)|>filter(keep)}\n",
		},
		{
			name: "field and index bind tighter than calls",
			src:  "fn main() { print f(1).a[0] }",
			want: "fn main(){print f(1).a[0]}\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Format(tc.src)
			if err != nil {
				t.Fatalf("format: %v", err)
			}
			if got != tc.want {
				t.Fatalf("canonical form mismatch\n got: %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestFormatIsIdempotent(t *testing.T) {
	paths, err := filepath.Glob("../../examples/*.ax")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no example programs found")
	}
	for _, path := range paths {
		src, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		once, err := Format(string(src))
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		twice, err := Format(once)
		if err != nil {
			t.Fatalf("%s: reformat: %v", path, err)
		}
		if once != twice {
			t.Fatalf("%s: formatter is not idempotent", path)
		}
	}
}

func TestFormatPreservesStructure(t *testing.T) {
	paths, err := filepath.Glob("../../examples/*.ax")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no example programs found")
	}
	for _, path := range paths {
		src, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		formatted, err := Format(string(src))
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		before := mustParse(t, path, string(src))
		after := mustParse(t, path, formatted)
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("%s: canonical form changed the program structure", path)
		}
	}
}

func mustParse(t *testing.T, path, src string) parser.Program {
	t.Helper()
	tokens, err := lexer.Scan(src)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return stripProgram(program)
}

// stripProgram removes source spans so two parses can be compared structurally.
func stripProgram(p parser.Program) parser.Program {
	out := parser.Program{
		Functions:    make([]parser.Function, len(p.Functions)),
		Imports:      append([]string(nil), p.Imports...),
		Capabilities: append([]string(nil), p.Capabilities...),
	}
	for i, function := range p.Functions {
		out.Functions[i] = parser.Function{
			Name:   function.Name,
			Params: append([]string(nil), function.Params...),
			Body:   stripStmts(function.Body),
		}
	}
	out.Body = stripStmts(p.Body)
	return out
}

func stripStmts(body []parser.Stmt) []parser.Stmt {
	out := make([]parser.Stmt, len(body))
	for i, stmt := range body {
		out[i] = parser.Stmt{
			Let:       stmt.Let,
			Value:     stripExpr(stmt.Value),
			Set:       stmt.Set,
			Print:     stripExpr(stmt.Print),
			Return:    stmt.Return,
			Result:    stripExpr(stmt.Result),
			While:     stripExpr(stmt.While),
			Body:      stripStmts(stmt.Body),
			Break:     stmt.Break,
			Continue:  stmt.Continue,
			ForVar:    stmt.ForVar,
			ForExpr:   stripExpr(stmt.ForExpr),
			TryBody:   stripStmts(stmt.TryBody),
			CatchName: stmt.CatchName,
			CatchBody: stripStmts(stmt.CatchBody),
		}
	}
	return out
}

func stripExpr(e parser.Expr) parser.Expr {
	switch x := e.(type) {
	case nil:
		return nil
	case parser.Located:
		return stripExpr(x.Inner)
	case parser.Binary:
		return parser.Binary{Op: x.Op, Left: stripExpr(x.Left), Right: stripExpr(x.Right)}
	case parser.Unary:
		return parser.Unary{Op: x.Op, Value: stripExpr(x.Value)}
	case parser.Conditional:
		return parser.Conditional{Test: stripExpr(x.Test), Yes: stripExpr(x.Yes), No: stripExpr(x.No)}
	case parser.Call:
		return parser.Call{Name: x.Name, Args: stripExprs(x.Args)}
	case parser.Pipeline:
		return parser.Pipeline{Left: stripExpr(x.Left), Name: x.Name, Args: stripExprs(x.Args)}
	case parser.ListLiteral:
		return parser.ListLiteral(stripExprs(x))
	case parser.RecordLiteral:
		fields := make(map[string]parser.Expr, len(x))
		for name, value := range x {
			fields[name] = stripExpr(value)
		}
		return parser.RecordLiteral(fields)
	case parser.Field:
		return parser.Field{Object: stripExpr(x.Object), Name: x.Name}
	case parser.Index:
		return parser.Index{Object: stripExpr(x.Object), Key: stripExpr(x.Key)}
	}
	return e
}

func stripExprs(items []parser.Expr) []parser.Expr {
	out := make([]parser.Expr, len(items))
	for i, item := range items {
		out[i] = stripExpr(item)
	}
	return out
}
