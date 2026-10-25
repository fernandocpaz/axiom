package bench

import "testing"

func TestRunReportsPassAtOne(t *testing.T) {
	summary := Run([]Task{{ID: "good", Prompt: "x"}, {ID: "bad", Prompt: "y"}}, 1, "fake", func(prompt string) (string, error) {
		if prompt == "x" {
			return "fn main(){print 1}\n", nil
		}
		return "fn main(){print true+1}\n", nil
	})
	if summary.PassAt1 != 1 || len(summary.Results) != 2 || !summary.Results[0].PassAt1 || summary.Results[1].PassAt1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
}
