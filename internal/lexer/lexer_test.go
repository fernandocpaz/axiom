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
