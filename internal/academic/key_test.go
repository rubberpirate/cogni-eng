package academic

import (
	"testing"
)

func TestSeparateAnswerKeyValidation(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		want       bool
	}{{"markdown", "# Answer key\n|Question|Answer|\n|---|---|\n|1|B|\n|2|4|", true}, {"csv", "question,answer\n1,B\n2,4\n", true}, {"json", `{"2":"C","1":"A|B"}`, true}, {"duplicate", "1 A\n1 B", false}, {"unknown", "1 Z", false}, {"empty", "Answer key", false}, {"garbled", "1,B\n2 unclear", false}, {"repeated label", "1 A|A", false}} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := ParseAnswerKey(tc.text)
			if (err == nil) != tc.want {
				t.Fatalf("%+v %v", out, err)
			}
			if err == nil && out.Entries[0].Number != 1 {
				t.Fatal("unordered key")
			}
		})
	}
}
