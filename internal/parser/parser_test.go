package parser

import (
	"github.com/fernandocpaz/axiom/internal/lexer"
	"testing"
)

func TestArithmetic(t *testing.T) {
	ts, e := lexer.Scan("fn main() { let x = 2 * (20 + 1) print x }")
	if e != nil {
		t.Fatal(e)
	}
	p, e := Parse(ts)
	if e != nil {
		t.Fatal(e)
	}
	got := -1
	e = p.Run(func(v Value) { got = v.Int })
	if e != nil || got != 42 {
		t.Fatalf("got %d, err %v", got, e)
	}
}

func TestConditionalAndBoolean(t *testing.T) {
	ts, err := lexer.Scan("fn main() { let x = if 4 >= 2 { 7 } else { 0 } print x print x == 7 }")
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	var got []Value
	if err = p.Run(func(v Value) { got = append(got, v) }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Int != 7 || got[1].Kind != "bool" || !got[1].Bool {
		t.Fatalf("unexpected output: %#v", got)
	}
}

func TestStringOutput(t *testing.T) {
	ts, err := lexer.Scan(`fn main() { print "Hello, world!" }`)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	var got Value
	if err = p.Run(func(v Value) { got = v }); err != nil {
		t.Fatal(err)
	}
	if got.Kind != "string" || got.Str != "Hello, world!" {
		t.Fatalf("unexpected output: %#v", got)
	}
}

func TestRejectsTrailingSource(t *testing.T) {
	ts, err := lexer.Scan("fn main() {} print 1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Parse(ts); err == nil {
		t.Fatal("expected trailing source to be rejected")
	}
}

func TestFunctionsParametersAndReturns(t *testing.T) {
	source := `fn add(a, b) { return a + b }
fn main() { print add(20, 22) }`
	ts, err := lexer.Scan(source)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	var got []Value
	if err = p.Run(func(v Value) { got = append(got, v) }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Kind != "int" || got[0].Int != 42 {
		t.Fatalf("unexpected output: %#v", got)
	}
}

func TestStructuredValues(t *testing.T) {
	source := `fn main() {
  let items = [10, 20, 12]
  let pod = record { name: "api", restarts: items[2] }
  print pod.name
  print pod["restarts"]
}`
	ts, err := lexer.Scan(source)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	var got []Value
	if err = p.Run(func(v Value) { got = append(got, v) }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Kind != "string" || got[0].Str != "api" || got[1].Int != 12 {
		t.Fatalf("unexpected output: %#v", got)
	}
}

func TestWhileBreakAndContinue(t *testing.T) {
	source := `fn main() {
  let n = 0
  while n < 5 {
    let n = n + 1
    continue
  }
  print n
  let m = 0
  while m < 5 {
    let m = m + 1
    print m
    break
  }
}`
	ts, err := lexer.Scan(source)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	var got []Value
	if err = p.Run(func(v Value) { got = append(got, v) }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Int != 5 || got[1].Int != 1 {
		t.Fatalf("unexpected output: %#v", got)
	}
}

func TestLogicalAndUnaryExpressions(t *testing.T) {
	source := `fn main() {
  print !false && (2 < 3 || -1 > 0)
}`
	ts, err := lexer.Scan(source)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	var got Value
	if err = p.Run(func(v Value) { got = v }); err != nil {
		t.Fatal(err)
	}
	if got.Kind != "bool" || !got.Bool {
		t.Fatalf("unexpected output: %#v", got)
	}
}

func TestAssignmentAndStringConcatenation(t *testing.T) {
	source := `fn main() {
  let greeting = "Hello"
  set greeting = greeting + ", world!"
  print greeting
}`
	ts, err := lexer.Scan(source)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	var got Value
	if err = p.Run(func(v Value) { got = v }); err != nil {
		t.Fatal(err)
	}
	if got.Kind != "string" || got.Str != "Hello, world!" {
		t.Fatalf("unexpected output: %#v", got)
	}
}

func TestPureTextBuiltins(t *testing.T) {
	source := `fn main() {
  print len("hello")
  print matches("pod-api-1", "^pod-.*")
  print text(number("42"))
}`
	ts, err := lexer.Scan(source)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	var got []Value
	if err = p.Run(func(v Value) { got = append(got, v) }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Int != 5 || !got[1].Bool || got[2].Str != "42" {
		t.Fatalf("unexpected output: %#v", got)
	}
}

func TestForInIteration(t *testing.T) {
	source := `fn main() { for item in [1, 2, 3] { print item } }`
	ts, err := lexer.Scan(source)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	var got []Value
	if err = p.Run(func(v Value) { got = append(got, v) }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Int != 1 || got[1].Int != 2 || got[2].Int != 3 {
		t.Fatalf("unexpected output: %#v", got)
	}
}

func TestTryCatchRecoversRuntimeErrors(t *testing.T) {
	source := `fn main() {
  try {
    print 1 / 0
  } catch error {
    print "recovered: " + error
  }
}`
	ts, err := lexer.Scan(source)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	var got []Value
	if err = p.Run(func(v Value) { got = append(got, v) }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Str != "recovered: division by zero" {
		t.Fatalf("unexpected output: %#v", got)
	}
}
