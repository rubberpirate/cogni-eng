package scholar

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"testing"

	"learning.local/workspace/internal/academic"
)

func testOMRPaper() academic.Paper {
	questions := []academic.Question{}
	for i, id := range []string{"one", "two", "three"} {
		q := academic.Question{ID: id, TopicID: "topic", Type: "single", Stem: "Original selection test " + id, Options: []academic.Option{{ID: "A", Text: "First"}, {ID: "B", Text: "Second"}, {ID: "C", Text: "Third"}, {ID: "D", Text: "Fourth"}}, Key: []string{"B"}, Solution: "Use the explicitly stated selection rule.", Marks: "4", Penalty: "1.25", Difficulty: "standard", Language: "en", Provenance: "Original synthetic test", License: "CC0-1.0"}
		if i == 1 {
			q.Type = "multiple"
			q.Key = []string{"A", "C"}
		}
		questions = append(questions, q)
	}
	return academic.Paper{Title: "Original OMR grading test", Purpose: "practice", DurationMinutes: 30, ReleasePolicy: "staff", Sections: []academic.Section{{ID: "all", Title: "All", Policy: "all", Count: 3, Questions: questions}}}
}
func TestNativeOMRUsesPinnedGraderAndKeepsUncertainty(t *testing.T) {
	p := testOMRPaper()
	reading := OMRCheckerReading{Fields: []OMRCheckerField{{Label: "q1", Value: "2", Selection: []string{"2"}}, {Label: "q2", Value: "AC", Selection: []string{"A", "C"}, Uncertain: true}, {Label: "q3", Selection: []string{}, Uncertain: true}}}
	answers, uncertain := nativeAnswers(p, reading)
	if len(uncertain) != 1 || uncertain[0] != 3 || answers["one"].Selection[0] != "B" {
		t.Fatalf("bad native mapping: %+v %+v", answers, uncertain)
	}
	result, err := academic.Mark(p, answers, nil, "paper")
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != "8" || result.Maximum != "12" {
		t.Fatalf("wrong score: %+v", result)
	}
	reading.Fields[0].Selection = []string{"A", "C"}
	reading.Fields[0].Uncertain = true
	answers, uncertain = nativeAnswers(p, reading)
	result, err = academic.Mark(p, answers, nil, "paper")
	if err != nil {
		t.Fatal(err)
	}
	if !answers["one"].Invalid || result.Score != "2.75" || len(uncertain) != 2 {
		t.Fatalf("invalid marks lost: %+v %+v", result, uncertain)
	}
}
func TestOMRTemplateRejectsInconsistentGeometryAndKeepsRegistrationZeros(t *testing.T) {
	var imageData bytes.Buffer
	if err := png.Encode(&imageData, image.NewGray(image.Rect(0, 0, 200, 300))); err != nil {
		t.Fatal(err)
	}
	template := map[string]any{"pageDimensions": []int{200, 300}, "bubbleDimensions": []int{10, 10}, "fieldBlocks": map[string]any{"q1": map[string]any{"fieldLabels": []string{"q1"}, "bubbleValues": []string{"1", "2", "3", "4"}, "origin": []int{20, 20}, "bubblesGap": 20, "direction": "horizontal"}}}
	in := OMRTemplateWrite{Name: "Four options", Template: Encode(template), Config: json.RawMessage(`{}`), Layout: json.RawMessage(`{}`)}
	if _, err := validateOMRTemplate(in, imageData.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := matchOMRPaper(PaperRecord{State: "published", Spec: testOMRPaper()}, in); err == nil {
		t.Fatal("mismatched assessment accepted")
	}
	template["pageDimensions"] = []int{201, 300}
	in.Template = Encode(template)
	if _, err := validateOMRTemplate(in, imageData.Bytes()); err == nil {
		t.Fatal("mismatched reference accepted")
	}
	registration, name, uncertain := checkedIdentity(OMRCheckerReading{Identity: []OMRIdentityReading{{Kind: "roll_number", Text: " 00 1234 "}, {Kind: "student_name", Text: "Synthetic Learner"}}})
	if registration != "001234" || name != "Synthetic Learner" || uncertain {
		t.Fatalf("identity changed: %s %s %v", registration, name, uncertain)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := rasterPagesLimit(ctx, "application/pdf", []byte("invalid"), 20); err == nil {
		t.Fatal("cancelled PDF continued")
	}
}
