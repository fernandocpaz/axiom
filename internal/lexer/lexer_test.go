package lexer

import "testing"

func TestScanKeywords(t *testing.T) {
	ts, e := Scan("fn main() { let x = 42 print x }")
	if e != nil {
		t.Fatal(e)
	}
	if ts[0].Kind != Fn || ts[4].Kind != LBrace || ts[7].Kind != Equal {
		t.Fatalf("unexpected tokens: %#v", ts)
	}
}

func TestScanComparisons(t *testing.T) {
	ts, err := Scan("if x >= 2 == true { print x }")
	if err != nil {
		t.Fatal(err)
	}
	if ts[2].Kind != GreaterEqual || ts[4].Kind != EqualEqual || ts[5].Kind != True {
		t.Fatalf("unexpected tokens: %#v", ts)
	}
}
