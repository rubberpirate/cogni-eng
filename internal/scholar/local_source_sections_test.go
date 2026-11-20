package scholar

import (
	"testing"

	"learning.local/workspace/internal/academic"
)

func TestSourceSectionsRestoreTypesMarksAndKeysAcrossNumberResets(t *testing.T) {
	text := "SECTION 1\nThe answer is a NUMERICAL VALUE. Full Marks: +4\nZero Marks: 0 In all other cases.\nQ1 First\nAnswer: 105\nQ2 Second\n" +
		"SECTION 1\nONE OR MORE options are correct. Full Marks: +4\nNegative Marks: −2\nQ1 New subject\nAnswer: A, D\n" +
		"SECTION 2\nONLY ONE option is correct. Full Marks: +3\nNegative Marks: −1\nQ2 Last\nAnswer: C\n"
	rows := []academic.ImportRow{}
	for _, n := range []int{1, 2, 1, 2} {
		q := questionExample("topic")
		q.Marks, q.Penalty = "1", "-2"
		rows = append(rows, academic.ImportRow{SourceNumber: n, Question: q})
	}
	if err := reconcileSourceSections(text, rows); err != nil {
		t.Fatal(err)
	}
	if q := rows[0].Question; q.Type != "numerical" || len(q.Options) != 0 || q.Numeric.Values[0] != "105" || q.Marks != "4" || q.Penalty != "0" || rows[0].Error != "" {
		t.Fatalf("numerical section lost: %+v", rows[0])
	}
	if q := rows[1].Question; q.Numeric != nil || rows[1].Error == "" {
		t.Fatal("a missing numerical key was fabricated")
	}
	if q := rows[2].Question; q.Type != "multiple" || len(q.Key) != 2 || q.Key[1] != "D" || q.Penalty != "2" || rows[2].Error != "" {
		t.Fatalf("subject reset lost: %+v", rows[2])
	}
	if q := rows[3].Question; q.Type != "single" || q.Key[0] != "C" || q.Marks != "3" || q.Penalty != "1" || rows[3].Error != "" {
		t.Fatalf("single section lost: %+v", rows[3])
	}
	if err := reconcileSourceSections(text, rows[:3]); err == nil {
		t.Fatal("missing question accepted")
	}
	rows[0].SourceNumber = 2
	if err := reconcileSourceSections(text, rows); err == nil {
		t.Fatal("reordered question accepted")
	}
}
