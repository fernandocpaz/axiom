package repair

import "testing"

func TestAnalyzeSuggestsCanonicalReplacement(t *testing.T) {
	report := Analyze(`fn main() { print 1 }`)
	if report.Valid || len(report.Actions) != 1 || report.Actions[0].Kind != "replace_source" {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestAnalyzeAcceptsCanonicalProgram(t *testing.T) {
	report := Analyze("fn main(){print 1}\n")
	if !report.Valid || len(report.Diagnostics) != 0 {
		t.Fatalf("unexpected report: %#v", report)
	}
}
