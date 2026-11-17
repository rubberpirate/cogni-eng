package scholar

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
)

type SampleKit struct {
	PaperID      string `json:"paper_id"`
	AssignmentID string `json:"assignment_id"`
	LayoutID     string `json:"layout_id"`
	Title        string `json:"title"`
}

func (s *Service) CourseLearners(ctx context.Context, tx pgx.Tx, actor, org, course string) ([]Recipient, error) {
	var campus string
	err := tx.QueryRow(ctx, "SELECT campus_id FROM course WHERE institution_id=$1 AND id=$2 AND state='active'", org, course).Scan(&campus)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, Problem(404, "course_unavailable", "Choose an active course.")
	}
	if err != nil {
		return nil, err
	}
	if err = Scope(ctx, tx, actor, org, campus, course, ""); err != nil {
		return nil, err
	}
	return rows[Recipient](ctx, tx, `SELECT p.id,p.name,e.reference FROM enrollment e JOIN person p ON p.id=e.person_id WHERE e.institution_id=$1 AND e.course_id=$2 AND e.state='active' AND p.state='active' AND e.start_date<=current_date AND (e.end_date IS NULL OR e.end_date>=current_date) ORDER BY p.name LIMIT 1000`, org, course)
}
func (s *Service) CreateSampleKit(ctx context.Context, tx pgx.Tx, actor, org, course, learner string) (SampleKit, error) {
	var campus string
	err := tx.QueryRow(ctx, "SELECT campus_id FROM course WHERE institution_id=$1 AND id=$2 AND state='active'", org, course).Scan(&campus)
	if errors.Is(err, pgx.ErrNoRows) {
		return SampleKit{}, Problem(404, "course_unavailable", "Choose an active course.")
	}
	if err != nil {
		return SampleKit{}, err
	}
	if err = Scope(ctx, tx, actor, org, campus, course, "Mathematics"); err != nil {
		return SampleKit{}, err
	}
	if err = LearnerScope(ctx, tx, learner, org, campus, course); err != nil {
		return SampleKit{}, err
	}
	topicName := "Practice samples - arithmetic and science"
	var topicID string
	err = tx.QueryRow(ctx, "SELECT id FROM topic WHERE institution_id=$1 AND course_id=$2 AND subject='Mathematics' AND name=$3", org, course, topicName).Scan(&topicID)
	if errors.Is(err, pgx.ErrNoRows) {
		t, e := s.CreateTopic(ctx, tx, actor, org, Topic{CampusID: campus, CourseID: course, Subject: "Mathematics", Name: topicName, Outcome: "Use original sample questions to test imports, OMR, review and exact marking; not a validated curriculum assessment.", Language: "en"})
		if e != nil {
			return SampleKit{}, e
		}
		topicID = t.ID
	} else if err != nil {
		return SampleKit{}, err
	}
	qs, err := academic.Samples(topicID)
	if err != nil {
		return SampleKit{}, err
	}
	ids := []string{}
	for _, q := range qs {
		created, duplicate, e := s.CreateQuestion(ctx, tx, actor, org, q, "development_fixture")
		if e != nil {
			return SampleKit{}, e
		}
		if duplicate {
			var state string
			var content []byte
			if e = tx.QueryRow(ctx, "SELECT state,content FROM question_revision WHERE institution_id=$1 AND id=$2", org, created.ID).Scan(&state, &content); e != nil {
				return SampleKit{}, e
			}
			if state != "approved" {
				return SampleKit{}, Problem(409, "sample_changed", "The sample questions were revised. Approve them in Question bank before creating another sample test.")
			}
			var existing academic.Question
			if e = json.Unmarshal(content, &existing); e != nil {
				return SampleKit{}, e
			}
			actual, expected := slices.Clone(existing.Key), slices.Clone(q.Key)
			slices.Sort(actual)
			slices.Sort(expected)
			if existing.Marks != q.Marks || existing.Penalty != q.Penalty || !slices.Equal(actual, expected) {
				return SampleKit{}, Problem(409, "sample_changed", "A sample answer key or marking rule was changed. Restore the original sample in Question bank or choose another course before creating a new kit.")
			}

		} else {
			if e = s.ReviewQuestion(ctx, tx, actor, org, created.ID, "approved", "Original CC0 demonstration fixture; approved for sample practice only", created.Version); e != nil {
				return SampleKit{}, e
			}
		}
		ids = append(ids, created.ID)
	}
	p, err := s.Compose(ctx, tx, actor, org, Compose{CampusID: campus, CourseID: course, Title: "Sample: practice, scan and review - " + time.Now().Format("02 Jan 15:04:05"), Purpose: "practice", Instructions: "Original sample questions. +2 for a correct answer, -0.5 for a wrong answer, 0 for a blank. Question 3 requires every correct choice.", DurationMinutes: 15, ReleasePolicy: "immediate", Sections: []ComposeSection{{ID: "samples", Title: "Six sample questions", Policy: "all", Count: 6, QuestionIDs: ids}}})
	if err != nil {
		return SampleKit{}, err
	}
	if err = s.Publish(ctx, tx, actor, org, p.ID, p.Version); err != nil {
		return SampleKit{}, err
	}
	assignments, err := s.Assign(ctx, tx, actor, org, p.ID, []string{learner}, time.Now().Add(-time.Minute), time.Now().AddDate(0, 0, 30), 0)
	if err != nil {
		return SampleKit{}, err
	}
	layout, err := s.CreateLayout(ctx, tx, actor, org, p.ID, 1)
	if err != nil {
		return SampleKit{}, err
	}
	if err = Event(ctx, tx, org, actor, "sample_kit.created", p.ID, "Original practice-only kit assigned to the selected learner; no result created"); err != nil {
		return SampleKit{}, err
	}
	return SampleKit{p.ID, assignments[0], layout.ID, p.Spec.Title}, nil
}
func (s *Service) SampleFiles(ctx context.Context, tx pgx.Tx, actor, org, paper, assignment string) ([]byte, error) {
	p, err := s.Paper(ctx, tx, actor, org, paper)
	if err != nil {
		return nil, err
	}
	if p.Spec.Purpose != "practice" || !strings.HasPrefix(p.Spec.Title, "Sample: practice, scan and review") {
		return nil, Problem(400, "sample_only", "Choose a sample practice test. Completed example sheets are only available for original practice samples.")
	}
	var learner string
	if err = tx.QueryRow(ctx, "SELECT learner_id FROM assignment WHERE institution_id=$1 AND id=$2 AND paper_id=$3", org, assignment, paper).Scan(&learner); errors.Is(err, pgx.ErrNoRows) {
		return nil, Problem(404, "assignment_unavailable", "Choose this sample test's assigned learner.")
	}
	if err != nil {
		return nil, err
	}
	layouts, err := s.Layouts(ctx, tx, actor, org, paper)
	if err != nil {
		return nil, err
	}
	if len(layouts) == 0 {
		return nil, Problem(409, "layout_missing", "Create the sample sheet layout first.")
	}
	layout := layouts[0]
	qs := []academic.Question{}
	correct, mixed, ambiguous := map[string]academic.Answer{}, map[string]academic.Answer{}, map[string]academic.Answer{}
	for _, sec := range p.Spec.Sections {
		qs = append(qs, sec.Questions...)
	}
	if len(qs) != 6 {
		return nil, Problem(409, "sample_changed", "This sample test no longer matches the six-question sample pack.")
	}
	mix := [][]string{{"B"}, {"A"}, {"A", "C"}, {}, {"B"}, {"D"}}
	for i, q := range qs {
		correct[q.ID] = academic.Answer{Selection: q.Key}
		mixed[q.ID] = academic.Answer{Selection: mix[i]}
		ambiguous[q.ID] = academic.Answer{Selection: q.Key}
	}
	ambiguous[qs[0].ID] = academic.Answer{Selection: []string{"B", "C"}}
	score, err := academic.Mark(p.Spec, mixed, nil, "paper")
	if err != nil {
		return nil, err
	}
	if score.Score != "5" || score.Maximum != "12" {
		return nil, Problem(409, "sample_changed", "This paper no longer matches the documented sample answers.")
	}
	files := map[string][]byte{"question-paper.pdf": academic.SamplePaperPDF, "marking-guide.pdf": academic.SampleKeyPDF, "questions.json": Encode(qs), "questions.csv": academic.SampleCSV, "expected-results.json": Encode(map[string]any{"correct_score": "12", "mixed_score": score.Score, "maximum": score.Maximum, "mixed_correct": score.Correct, "mixed_wrong": score.Wrong, "mixed_blank": 1, "ambiguous_instruction": "Question 1 has two marks. Review it and keep B for a score of 12.", "assignment_id": assignment, "layout_id": layout.ID})}
	var first image.Image
	for _, item := range []struct {
		Name    string
		Answers map[string]academic.Answer
	}{{"blank", nil}, {"correct", correct}, {"mixed", mixed}, {"needs-review", ambiguous}} {
		data, e := academic.SheetImage(layout.Spec, assignment, layout.Fingerprint, p.Spec.Title, "Sample learner - "+assignment[:8], item.Answers)
		if e != nil {
			return nil, e
		}
		files["omr-"+item.Name+".png"] = data
		img, e := png.Decode(bytes.NewReader(data))
		if e != nil {
			return nil, e
		}
		if item.Name == "correct" {
			first = img
		}
		pdf, e := academic.ImagesPDF(img)
		if e != nil {
			return nil, e
		}
		files["omr-"+item.Name+".pdf"] = pdf
	}
	files["duplicate-pages.pdf"], err = academic.ImagesPDF(first, first)
	if err != nil {
		return nil, err
	}
	files["START-HERE.txt"] = []byte(fmt.Sprintf("ORIGINAL SAMPLE KIT - CC0-1.0\n\nThis kit matches paper %s and assignment %s. It contains no real learner's answers.\n\n1. Read question-paper.pdf and marking-guide.pdf. You may also import questions.json or the public sample CSV into a topic.\n2. Open Scan & check and choose this sample test. Choose its assigned learner and default layout.\n3. Upload ONE example to begin: omr-correct.png (12/12), omr-mixed.png (5/12: 3 correct, 2 wrong, 1 blank), or omr-needs-review.png (question 1 must be corrected to B). PDF versions test PDF intake.\n4. Wait for processing, check identity, review any uncertainty and commit once. Your saved result must agree with expected-results.json.\n5. duplicate-pages.pdf contains the same page twice to test page deduplication. Re-upload the same source to test request deduplication.\n6. A learner assignment may have only one sitting. Create another sample kit in Try sample files before committing a different example. Never use a real learner's formal test to experiment.\n\nThe blank PDF can be printed and scanned or photographed. Keep all four black corner squares and the identity grid visible. These are original registered sheets, not universal commercial OMR forms.\n", p.Reference, assignment))
	var out bytes.Buffer
	archive := zip.NewWriter(&out)
	for name, data := range files {
		writer, e := archive.Create(name)
		if e != nil {
			return nil, e
		}
		if _, e = writer.Write(data); e != nil {
			return nil, e
		}
	}
	if err = archive.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
