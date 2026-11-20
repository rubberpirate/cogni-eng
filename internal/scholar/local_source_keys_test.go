package scholar

import (
	"reflect"
	"testing"

	"learning.local/workspace/internal/academic"
)

func TestReconcileSourceKeysPreservesExplicitAnswers(t *testing.T) {
	for _, tc := range []struct {
		name, kind, answer string
		keys               []string
		numeric            *academic.NumericRule
		invalid            bool
	}{
		{name: "single", kind: "single", answer: "C", keys: []string{"C"}},
		{name: "multiple", kind: "multiple", answer: "A, D", keys: []string{"A", "D"}},
		{name: "integer", kind: "numerical", answer: "105", numeric: &academic.NumericRule{Values: []string{"105"}, Integer: true}},
		{name: "decimal", kind: "numerical", answer: "−2.3", numeric: &academic.NumericRule{Values: []string{"-2.3"}}},
		{name: "alternative values", kind: "numerical", answer: "0.5 OR 0.75", numeric: &academic.NumericRule{Values: []string{"0.5", "0.75"}}},
		{name: "range", kind: "numerical", answer: "[-7.2 to -7]", numeric: &academic.NumericRule{Values: []string{}, Min: "-7.2", Max: "-7"}},
		{name: "disjoint ranges require review", kind: "numerical", answer: "[-30 to -29] OR [29 to 30]", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := questionExample("topic")
			q.Type, q.Key, q.Solution = tc.kind, nil, tc.answer
			if tc.kind == "numerical" {
				q.Options = nil
				q.Numeric = &academic.NumericRule{Values: []string{"999"}}
			}
			rows := []academic.ImportRow{{SourceNumber: 7, Question: q, Error: "missing key"}}
			reconcileSourceKeys("Q.7 Printed wording\nAnswer: "+tc.answer+"\n", rows)
			got := rows[0]
			if !reflect.DeepEqual(got.Question.Key, tc.keys) || !reflect.DeepEqual(got.Question.Numeric, tc.numeric) {
				t.Fatalf("source answer changed: %+v", got.Question)
			}
			if (got.Error != "") != tc.invalid {
				t.Fatalf("review state: %s", got.Error)
			}
			if got.Question.Solution != "" {
				t.Fatal("printed answer retained as a worked solution")
			}
		})
	}
}

func TestReconcileSourceKeysUsesIdentityAndRejectsAmbiguity(t *testing.T) {
	q := questionExample("topic")
	rows := []academic.ImportRow{{SourceNumber: 2, Question: q}, {SourceNumber: 1, Question: q}}
	reconcileSourceKeys("Q1 First\nAnswer: B\nQ2 Second\nAnswer: D\n", rows)
	if rows[0].Question.Key[0] != "D" || rows[1].Question.Key[0] != "B" {
		t.Fatal("answers matched by position instead of source number")
	}
	reconcileSourceKeys("Q1 First\nAnswer: B\nQ1 Another subject\nAnswer: C\n", rows)
	if rows[1].Error == "" || len(rows[1].Question.Key) != 0 {
		t.Fatal("duplicate source numbers silently assigned a key")
	}
	rows = []academic.ImportRow{{SourceNumber: 1, Question: q}}
	reconcileSourceKeys("Q1 First without a supplied answer\n", rows)
	if !reflect.DeepEqual(rows[0].Question, q) {
		t.Fatal("unmarked source was rewritten")
	}
}
