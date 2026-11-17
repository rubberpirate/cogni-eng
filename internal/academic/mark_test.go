package academic

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"
	"testing"
)

func question(id string) Question {
	return Question{ID: id, TopicID: "topic", Type: "single", Stem: "Which answer is correct?", Options: []Option{{ID: "A", Text: "One"}, {ID: "B", Text: "Two"}}, Key: []string{"A"}, Solution: "One is the pinned answer.", Marks: "3", Penalty: "1", Difficulty: "standard", Language: "en", Provenance: "Original test fixture", License: "CC0-1.0"}
}
func paper(qs ...Question) Paper {
	return Paper{Title: "Original assessment", Purpose: "practice", DurationMinutes: 30, ReleasePolicy: "immediate", Sections: []Section{{ID: "one", Title: "Core", Policy: "all", Count: len(qs), Questions: qs}}}
}
func TestOptionalAccountingAndExactDecimals(t *testing.T) {
	qs := []Question{}
	answers := map[string]Answer{}
	for i := 0; i < 26; i++ {
		q := question(fmt.Sprint(i))
		if i >= 18 {
			q.Marks = "2"
		}
		qs = append(qs, q)
		answers[q.ID] = Answer{Selection: []string{"A"}}
	}
	p := paper(qs...)
	got, err := Mark(p, answers, nil, "online")
	if err != nil || got.Maximum != "70" || got.Score != "70" {
		t.Fatalf("18×3+8×2: %+v %v", got, err)
	}
	p = paper(question("a"), question("b"), question("c"))
	p.Sections[0].Policy = "first"
	p.Sections[0].Count = 2
	answers = map[string]Answer{"a": {Selection: []string{"B"}, Order: 2}, "b": {Selection: []string{"A"}, Order: 3}, "c": {Selection: []string{"A"}, Order: 1}}
	got, err = Mark(p, answers, nil, "online")
	if err != nil || got.Score != "2" || got.Maximum != "6" || got.Items[1].Counted {
		t.Fatalf("first answer order: %+v %v", got, err)
	}
	p.Sections[0].Policy = "best"
	got, err = Mark(p, answers, nil, "paper")
	if err != nil || got.Score != "6" || got.Items[0].Counted {
		t.Fatalf("best contribution: %+v %v", got, err)
	}
	delete(answers, "a")
	delete(answers, "b")
	got, err = Mark(p, answers, nil, "paper")
	if err != nil || got.Wrong != 0 || got.Items[1].Status != "not_required" || got.Required != 2 {
		t.Fatalf("optional blank: %+v %v", got, err)
	}
}
func TestNumericalAndManualRules(t *testing.T) {
	q := question("n")
	q.Type = "numerical"
	q.Options = nil
	q.Key = nil
	q.Numeric = &NumericRule{Values: []string{"0.3"}, Absolute: "0.01"}
	q.Marks = "1.125"
	q.Penalty = "0.125"
	for _, tc := range []struct{ answer, score string }{{"0.31", "1.125"}, {".300000000000000000001", "1.125"}, {"0.310000000000000000001", "-0.125"}, {"NaN", "-0.125"}, {"1e-1", "-0.125"}} {
		got, err := Mark(paper(q), map[string]Answer{"n": {Number: tc.answer}}, nil, "online")
		if err != nil || got.Score != tc.score {
			t.Fatalf("%s: %+v %v", tc.answer, got, err)
		}
	}
	q.Type = "short"
	q.Numeric = nil
	q.Rubric = []Criterion{{ID: "reason", Description: "Valid reasoning", Marks: "1.125"}}
	got, err := Mark(paper(q), map[string]Answer{"n": {Text: "A reason"}}, map[string]ManualMark{"n": {Approved: false, Criteria: map[string]string{"reason": "1.125"}}}, "paper")
	if err != nil || got.Pending != 1 {
		t.Fatalf("proposal released: %+v %v", got, err)
	}
	_, err = Mark(paper(q), map[string]Answer{"n": {Text: "A reason"}}, map[string]ManualMark{"n": {Approved: true, Criteria: map[string]string{"reason": "1.126"}}}, "paper")
	if err == nil {
		t.Fatal("rubric overflow accepted")
	}
}
func FuzzNumber(f *testing.F) {
	for _, v := range []string{"0", "-0.125", "NaN", "1e5", ".5", "99999999999999999999999"} {
		f.Add(v)
	}
	f.Fuzz(func(t *testing.T, s string) {
		n, err := Number(s)
		if err == nil && n == nil {
			t.Fatal("accepted number has no exact value")
		}
	})
}
func TestImportPreviewRetainsInvalidAndNumerical(t *testing.T) {
	data := []byte("type,question,A,B,answer,marks,penalty\nsingle,Choose one,One,Two,A,2,0.5\nsingle,Bad key,One,Two,C,2,0\nnumerical,Calculate a decimal,,,0.3,1,0\n")
	rows, err := ParseImport("questions.csv", data, "topic", "Original teacher material", "CC0-1.0")
	if err != nil || len(rows) != 3 || rows[0].Error != "" || rows[1].Error == "" || rows[2].Question.Numeric.Values[0] != "0.3" {
		t.Fatalf("preview %+v %v", rows, err)
	}
}
func TestRegisteredOMRRealPixelsAndAmbiguity(t *testing.T) {
	p := paper(question(strings.Repeat("a", 64)), question(strings.Repeat("b", 64)))
	for _, columns := range []int{1, 2} {
		l, err := NewLayout(p, columns)
		if err != nil {
			t.Fatal(err)
		}
		id, fp := strings.Repeat("a", 64), strings.Repeat("b", 64)
		data, err := SheetImage(l, id, fp, "Test paper", "Synthetic learner", map[string]Answer{p.Sections[0].Questions[0].ID: {Selection: []string{"B"}}})
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		read := Recognize(img, l, fp)
		if read.Error != "" || read.AssignmentID != id || read.Rows[l.Rows[0].QuestionID].Answer.Selection[0] != "B" || len(read.Rows[l.Rows[1].QuestionID].Answer.Selection) != 0 {
			t.Fatalf("reading %+v", read)
		}
		rotated := image.NewRGBA(image.Rect(0, 0, 1400, 1000))
		for y := 0; y < 1400; y++ {
			for x := 0; x < 1000; x++ {
				rotated.Set(1399-y, x, img.At(x, y))
			}
		}
		read = Recognize(rotated, l, fp)
		if read.Error != "" || read.AssignmentID != id {
			t.Fatalf("rotation %+v", read)
		}
		partial := image.NewRGBA(img.Bounds())
		draw.Draw(partial, partial.Bounds(), img, image.Point{}, draw.Src)
		c := l.Rows[1].Centers[0]
		draw.Draw(partial, image.Rect(int(c.X)-5, int(c.Y)-5, int(c.X), int(c.Y)+5), image.NewUniform(color.Black), image.Point{}, draw.Src)
		read = Recognize(partial, l, fp)
		if !read.Rows[l.Rows[1].QuestionID].Uncertain {
			t.Fatal("faint fill silently accepted")
		}
		read = Recognize(img, l, strings.Repeat("c", 64))
		if read.Error == "" {
			t.Fatal("wrong layout accepted")
		}
	}
}

func TestOMRPerspectiveAndExposure(t *testing.T) {
	q := question(strings.Repeat("d", 64))
	l, err := NewLayout(paper(q), 1)
	if err != nil {
		t.Fatal(err)
	}
	id, fp := strings.Repeat("e", 64), strings.Repeat("f", 64)
	data, err := SheetImage(l, id, fp, "Perspective fixture", "Synthetic learner", map[string]Answer{q.ID: {Selection: []string{"A"}}})
	if err != nil {
		t.Fatal(err)
	}
	original, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	destination := []Point{{75, 85}, {930, 45}, {970, 1350}, {35, 1320}}
	inverse, err := projectionBetween(destination, []Point{{40, 40}, {960, 40}, {960, 1360}, {40, 1360}})
	if err != nil {
		t.Fatal(err)
	}
	warped := image.NewGray(image.Rect(0, 0, 1000, 1400))
	for y := 0; y < 1400; y++ {
		for x := 0; x < 1000; x++ {
			p := inverse(Point{float64(x), float64(y)})
			value := uint8(235)
			if dark(original, int(p.X), int(p.Y)) {
				value = 45
			}
			warped.SetGray(x, y, color.Gray{Y: value})
		}
	}
	read := Recognize(warped, l, fp)
	if read.Error != "" || read.AssignmentID != id || read.Rows[q.ID].Uncertain || len(read.Rows[q.ID].Answer.Selection) != 1 || read.Rows[q.ID].Answer.Selection[0] != "A" {
		t.Fatalf("perspective/exposure fixture failed: %+v", read)
	}
}
