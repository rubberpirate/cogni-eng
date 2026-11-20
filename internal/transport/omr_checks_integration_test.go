//go:build integration

package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
	"learning.local/workspace/internal/scholar"
)

func TestSavedOMRTemplatesGradeWithoutAccountsAndProtectReview(t *testing.T) {
	if os.Getenv("RUN_OMRCHECKER_TESTS") != "1" {
		t.Skip("installed native engine and private object store required")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("OMRCHECKER_PYTHON", filepath.Join(root, "omr/.venv/bin/python"))
	t.Setenv("OMRCHECKER_ADAPTER", filepath.Join(root, "omr/bridge.py"))
	h, svc, course, learner := academicFixture(t)
	ctx := context.Background()
	sample := filepath.Join(root, "apps/web/public/samples/omrchecker")
	image, err := os.ReadFile(filepath.Join(sample, "filled-sheet.png"))
	if err != nil {
		t.Fatal(err)
	}
	var expected struct {
		Responses map[string]string `json:"responses"`
	}
	raw, err := os.ReadFile(filepath.Join(sample, "expected-answers.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &expected); err != nil {
		t.Fatal(err)
	}
	var paper scholar.PaperRecord
	academicTx(t, h, func(tx pgx.Tx) error {
		topic, e := svc.CreateTopic(ctx, tx, h.owner, h.org, scholar.Topic{CampusID: h.campus, CourseID: course, Subject: "Recognition testing", Name: "Original recognition test", Outcome: "Read a documented marked response accurately", Language: "en"})
		if e != nil {
			return e
		}
		questions := []academic.Question{}
		for i := 1; i <= 20; i++ {
			key := expected.Responses[fmt.Sprintf("q%d", i)]
			if key == "" {
				key = "B"
			}
			q := academic.Question{TopicID: topic.ID, Type: "single", Stem: fmt.Sprintf("Original recognition exercise %d: choose %s as the documented practice response.", i, key), Options: []academic.Option{{ID: "A", Text: "First option"}, {ID: "B", Text: "Second option"}, {ID: "C", Text: "Third option"}, {ID: "D", Text: "Fourth option"}}, Key: strings.Split(key, ""), Solution: "This synthetic exercise tests the documented marked response.", Marks: "4", Penalty: "1", Difficulty: "standard", Language: "en", Provenance: "Original synthetic recognition exercise", License: "CC0-1.0"}
			if i == 12 {
				q.Type = "multiple"
			}
			questions = append(questions, q)
		}
		draft, e := svc.SaveDraft(ctx, tx, h.owner, h.org, "", scholar.DraftWrite{CampusID: h.campus, CourseID: course, Spec: scholar.DraftSpec{Title: "Account-free native recognition", Method: "manual", TopicID: topic.ID, Exam: "Original recognition", StartsAt: time.Now(), DurationMinutes: 30, Instructions: "Original synthetic recognition test", Learners: []string{}, Questions: questions}})
		if e != nil {
			return e
		}
		draft, e = svc.PublishDraft(ctx, tx, h.owner, h.org, draft.ID, draft.Version)
		if e != nil {
			return e
		}
		paper, e = svc.Paper(ctx, tx, h.owner, h.org, draft.PaperID)
		return e
	})
	blocks := map[string]any{}
	for i := 1; i <= 20; i++ {
		col, row := (i-1)/10, (i-1)%10
		blocks[fmt.Sprintf("answer_%d", i)] = map[string]any{"fieldLabels": []string{fmt.Sprintf("q%d", i)}, "bubbleValues": []string{"A", "B", "C", "D"}, "origin": []int{167 + col*560, 737 + row*70}, "bubbleDimensions": []int{26, 26}, "bubblesGap": 70, "labelsGap": 0, "direction": "horizontal"}
	}
	encode := func(v any) json.RawMessage {
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	definition := scholar.OMRTemplateWrite{Name: "Faculty recognition layout", Template: encode(map[string]any{"pageDimensions": []int{1200, 1600}, "bubbleDimensions": []int{26, 26}, "fieldBlocks": blocks}), Config: json.RawMessage(`{}`), Layout: json.RawMessage(`{}`), Identity: []scholar.OMRIdentityRegion{{Kind: "student_name", Rect: academic.GridRect{X: .06, Y: .1, Width: .66, Height: .028}}, {Kind: "roll_number", Rect: academic.GridRect{X: .06, Y: .13, Width: .34, Height: .027}}}}
	var template scholar.OMRSavedTemplate
	academicTx(t, h, func(tx pgx.Tx) error {
		var e error
		template, e = svc.SaveOMRTemplate(ctx, tx, h.owner, h.org, course, definition, image)
		return e
	})
	academicTx(t, h, func(tx pgx.Tx) error {
		list, e := svc.OMRTemplates(ctx, tx, h.owner, h.org)
		if e != nil {
			return e
		}
		if len(list) != 1 || list[0].Name != definition.Name {
			t.Fatal("template not persistent")
		}
		return nil
	})
	var check scholar.OMRCheck
	academicTx(t, h, func(tx pgx.Tx) error {
		var e error
		check, e = svc.StageOMRCheck(ctx, tx, h.owner, h.org, paper.ID, template.ID, "original-filled.png", image)
		return e
	})
	academicTx(t, h, func(tx pgx.Tx) error {
		repeat, e := svc.StageOMRCheck(ctx, tx, h.owner, h.org, paper.ID, template.ID, "renamed.png", image)
		if e == nil && repeat.ID != check.ID {
			t.Fatal("duplicate upload created a second check")
		}
		return e
	})
	for _, actor := range []string{learner} {
		if e := scholar.Tx(ctx, h.db, func(tx pgx.Tx) error { _, e := svc.OMRTemplates(ctx, tx, actor, h.org); return e }); e == nil {
			t.Fatal("learner read templates")
		}
	}
	if e := scholar.Tx(ctx, h.db, func(tx pgx.Tx) error { _, e := svc.OMRCheck(ctx, tx, h.owner, h.otherOrg, check.ID); return e }); e == nil {
		t.Fatal("cross-institution check exposed")
	}
	worker, e := svc.WorkerClient()
	if e != nil {
		t.Fatal(e)
	}
	if e = worker.Start(ctx); e != nil {
		t.Fatal(e)
	}
	defer func() {
		stop, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if e := worker.Stop(stop); e != nil {
			t.Error(e)
		}
	}()
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		academicTx(t, h, func(tx pgx.Tx) error {
			var e error
			check, e = svc.OMRCheck(ctx, tx, h.owner, h.org, check.ID)
			return e
		})
		if check.State == "completed" {
			break
		}
		if check.State == "failed" || check.State == "configuration_required" {
			t.Fatalf("native job failed: %s", check.Error)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if check.State != "completed" || len(check.Sheets) != 1 {
		t.Fatalf("native job not completed: %+v", check)
	}
	for _, action := range []string{"read", "review", "asset"} {
		if e := scholar.Tx(ctx, h.db, func(tx pgx.Tx) error {
			v, e := svc.OMRSheet(ctx, tx, h.owner, h.org, check.Sheets[0].ID)
			if e != nil {
				return e
			}
			switch action {
			case "read":
				_, e = svc.OMRSheet(ctx, tx, learner, h.org, v.Summary.ID)
			case "review":
				_, e = svc.ReviewOMRSheet(ctx, tx, learner, h.org, v.Summary.ID, scholar.OMRReview{})
			case "asset":
				_, e = svc.Asset(ctx, tx, learner, h.org, v.AssetID)
			}
			return e
		}); e == nil {
			t.Fatalf("learner could %s private OMR grade", action)
		}
	}
	summary := check.Sheets[0]
	if summary.Registration != "123456" || summary.Name != "Synthetic Test Learner" || summary.Score != "76" || summary.Maximum != "80" || summary.State != "needs_review" {
		t.Fatalf("unexpected persisted grade: %+v", summary)
	}
	var sheet scholar.OMRCheckedSheet
	academicTx(t, h, func(tx pgx.Tx) error {
		var e error
		sheet, e = svc.OMRSheet(ctx, tx, h.owner, h.org, summary.ID)
		return e
	})
	review := scholar.OMRReview{Version: summary.Version, Registration: "00123456", Name: summary.Name, Confirmed: true, Reason: "Checked registration and confirmed the intentionally blank seventh response", Answers: sheet.Answers}
	first := paper.Spec.Sections[0].Questions[0].ID
	correction := review.Answers[first]
	correction.Selection = []string{"B"}
	review.Answers[first] = correction
	review.Reason = "Original test of a versioned answer correction and registration review"
	academicTx(t, h, func(tx pgx.Tx) error {
		var e error
		sheet, e = svc.ReviewOMRSheet(ctx, tx, h.owner, h.org, summary.ID, review)
		return e
	})
	if sheet.Summary.State != "graded" || sheet.Summary.Registration != "00123456" || len(sheet.Summary.Correct) != 18 || sheet.Summary.Score != "71" {
		t.Fatalf("review grade not saved: %+v", sheet.Summary)
	}
	if e := scholar.Tx(ctx, h.db, func(tx pgx.Tx) error {
		_, e := svc.ReviewOMRSheet(ctx, tx, h.owner, h.org, summary.ID, review)
		return e
	}); e == nil {
		t.Fatal("stale review accepted")
	}
	command := exec.CommandContext(ctx, filepath.Join(root, "omr/.venv/bin/python"), "-c", "from PIL import Image;import sys,io;im=Image.open(sys.argv[1]).convert('RGB');out=io.BytesIO();im.save(out,format='PDF',resolution=144,save_all=True,append_images=[im]);sys.stdout.buffer.write(out.getvalue())", filepath.Join(sample, "filled-sheet.png"))
	pdf, e := command.Output()
	if e != nil {
		t.Fatal(e)
	}
	var batch scholar.OMRCheck
	academicTx(t, h, func(tx pgx.Tx) error {
		var e error
		batch, e = svc.StageOMRCheck(ctx, tx, h.owner, h.org, paper.ID, template.ID, "two-original-sheets.pdf", pdf)
		return e
	})
	deadline = time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		academicTx(t, h, func(tx pgx.Tx) error {
			var e error
			batch, e = svc.OMRCheck(ctx, tx, h.owner, h.org, batch.ID)
			return e
		})
		if batch.State == "completed" {
			break
		}
		if batch.State == "failed" {
			t.Fatalf("PDF failed: %s", batch.Error)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if len(batch.Sheets) != 2 || batch.State != "completed" {
		t.Fatalf("PDF batch did not retain both pages: %+v", batch)
	}
	for _, v := range batch.Sheets {
		if v.Score != "76" || v.Registration != "123456" {
			t.Fatalf("PDF native reading changed: %+v", v)
		}
	}
	pdfReview := scholar.OMRReview{Version: 1, Registration: "123456", Name: "Synthetic Test Learner", Confirmed: true, Reason: "Confirmed page one and the intentionally blank response", Answers: map[string]academic.Answer{}}
	academicTx(t, h, func(tx pgx.Tx) error {
		_, e := svc.ReviewOMRSheet(ctx, tx, h.owner, h.org, batch.Sheets[0].ID, pdfReview)
		return e
	})
	if e := scholar.Tx(ctx, h.db, func(tx pgx.Tx) error {
		_, e := svc.ReviewOMRSheet(ctx, tx, h.owner, h.org, batch.Sheets[1].ID, pdfReview)
		return e
	}); e == nil {
		t.Fatal("duplicate registration overwritten or merged")
	}
	pdfReview.Exclude = true
	pdfReview.Reason = "Excluded the duplicate PDF page; original evidence retained"
	academicTx(t, h, func(tx pgx.Tx) error {
		out, e := svc.ReviewOMRSheet(ctx, tx, h.owner, h.org, batch.Sheets[1].ID, pdfReview)
		if e == nil && out.Summary.State != "excluded" {
			t.Fatal("duplicate page not excluded")
		}
		return e
	})
	var accounts, assignments, results int
	if e = h.db.QueryRow(ctx, "SELECT (SELECT count(*) FROM person),(SELECT count(*) FROM assignment),(SELECT count(*) FROM result_revision)").Scan(&accounts, &assignments, &results); e != nil {
		t.Fatal(e)
	}
	if accounts != 2 || assignments != 0 || results != 0 {
		t.Fatalf("scan manufactured account or linked results: %d %d %d", accounts, assignments, results)
	}
	var revisions int
	if e = h.db.QueryRow(ctx, "SELECT jsonb_array_length(reviews) FROM omr_checked_sheet WHERE id=$1", summary.ID).Scan(&revisions); e != nil || revisions != 1 {
		t.Fatalf("review evidence lost: %d %v", revisions, e)
	}
	var previousScore string
	if e = h.db.QueryRow(ctx, "SELECT reviews->0->'result'->>'score' FROM omr_checked_sheet WHERE id=$1", summary.ID).Scan(&previousScore); e != nil || previousScore != "76" {
		t.Fatalf("previous grade evidence changed: %s %v", previousScore, e)
	}
	var previousAnswer string
	if e = h.db.QueryRow(ctx, "SELECT reviews->0->'answers'->$2::text->'selection'->>0 FROM omr_checked_sheet WHERE id=$1", summary.ID, first).Scan(&previousAnswer); e != nil || previousAnswer != "A" {
		t.Fatalf("previous answer evidence changed: %s %v", previousAnswer, e)
	}
	configuredRoot := os.Getenv("OMRCHECKER_ROOT")
	t.Setenv("OMRCHECKER_ROOT", filepath.Join(root, ".local/nonexistent-omr-runtime"))
	var retry scholar.OMRCheck
	academicTx(t, h, func(tx pgx.Tx) error {
		var e error
		retry, e = svc.StageOMRCheck(ctx, tx, h.owner, h.org, paper.ID, template.ID, "retained-retry.png", append(append([]byte{}, image...), '\n'))
		return e
	})
	deadline = time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		academicTx(t, h, func(tx pgx.Tx) error {
			var e error
			retry, e = svc.OMRCheck(ctx, tx, h.owner, h.org, retry.ID)
			return e
		})
		if retry.State == "configuration_required" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if retry.State != "configuration_required" {
		t.Fatalf("missing runtime did not expose recovery: %+v", retry)
	}
	t.Setenv("OMRCHECKER_ROOT", configuredRoot)
	deadline = time.Now().Add(5 * time.Second)
	for {
		e := scholar.Tx(ctx, h.db, func(tx pgx.Tx) error { return svc.ControlJob(ctx, tx, h.owner, h.org, retry.JobID, "retry") })
		if e == nil {
			break
		}
		var problem *scholar.Error
		if !errors.As(e, &problem) || problem.Code != "job_still_stopping" || time.Now().After(deadline) {
			t.Fatal(e)
		}
		time.Sleep(100 * time.Millisecond)
	}
	deadline = time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		academicTx(t, h, func(tx pgx.Tx) error {
			var e error
			retry, e = svc.OMRCheck(ctx, tx, h.owner, h.org, retry.ID)
			return e
		})
		if retry.State == "completed" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if retry.State != "completed" || len(retry.Sheets) != 1 || retry.Sheets[0].Score != "76" {
		t.Fatalf("retained upload did not recover: %+v", retry)
	}
	academicTx(t, h, func(tx pgx.Tx) error {
		older, e := svc.OMRChecks(ctx, tx, h.owner, h.org, paper.ID, retry.ID)
		if e != nil {
			return e
		}
		if len(older) != 2 {
			t.Fatalf("cursor lost older uploads: %+v", older)
		}
		for _, v := range older {
			if v.ID == retry.ID {
				t.Fatal("cursor duplicated its boundary upload")
			}
		}
		last, e := svc.OMRChecks(ctx, tx, h.owner, h.org, paper.ID, check.ID)
		if e == nil && len(last) != 0 {
			t.Fatal("oldest cursor did not end the list")
		}
		return e
	})
	if e := scholar.Tx(ctx, h.db, func(tx pgx.Tx) error {
		_, e := svc.OMRChecks(ctx, tx, h.owner, h.org, paper.ID, "unavailable-cursor")
		return e
	}); e == nil {
		t.Fatal("invalid results cursor accepted")
	}
	// Layout geometry is shared across assessments and teaching batches; keys are not.
	otherCourse, instructor := access.Random(), access.Random()
	h.exec("INSERT INTO course(id,institution_id,campus_id,name,academic_year) VALUES($1,$2,$3,'Second teaching batch','2026')", otherCourse, h.org, h.campus)
	h.exec("INSERT INTO person(id,email,name,password_hash,state,verified_at) VALUES($1,'second-instructor@example.invalid','Synthetic second instructor','test-only-not-a-login-hash','active',now())", instructor)
	h.exec("INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,'instructor')", access.Random(), h.org, instructor, h.campus)
	h.exec("INSERT INTO responsibility(id,institution_id,staff_id,course_id,subject,starts_on) VALUES($1,$2,$3,$4,'Recognition reuse',current_date)", access.Random(), h.org, instructor, otherCourse)
	var shorter scholar.PaperRecord
	academicTx(t, h, func(tx pgx.Tx) error {
		topic, e := svc.CreateTopic(ctx, tx, h.owner, h.org, scholar.Topic{CampusID: h.campus, CourseID: otherCourse, Subject: "Recognition reuse", Name: "Reuse a generic OMR layout", Outcome: "Grade a shorter assessment using its own key", Language: "en"})
		if e != nil {
			return e
		}
		qs := []academic.Question{}
		for _, q := range paper.Spec.Sections[0].Questions[:3] {
			q.ID = ""
			q.TopicID = topic.ID
			q.Stem += " Original second-assessment reuse."
			qs = append(qs, q)
		}
		// A different key verifies that the template never supplies the answers.
		qs[0].Key = []string{"B"}
		d, e := svc.SaveDraft(ctx, tx, h.owner, h.org, "", scholar.DraftWrite{CampusID: h.campus, CourseID: otherCourse, Spec: scholar.DraftSpec{Title: "Three-question reuse", Method: "manual", TopicID: topic.ID, Exam: "Original reuse", StartsAt: time.Now(), DurationMinutes: 30, Instructions: "Independent assessment key", Questions: qs, Learners: []string{}}})
		if e != nil {
			return e
		}
		d, e = svc.PublishDraft(ctx, tx, h.owner, h.org, d.ID, d.Version)
		if e != nil {
			return e
		}
		shorter, e = svc.Paper(ctx, tx, h.owner, h.org, d.PaperID)
		return e
	})
	var shared scholar.OMRCheck
	academicTx(t, h, func(tx pgx.Tx) error {
		items, e := svc.OMRTemplates(ctx, tx, instructor, h.org)
		if e != nil {
			return e
		}
		found := false
		for _, v := range items {
			if v.ID == template.ID {
				found = true
			}
		}
		if !found {
			t.Fatal("shared template missing for another batch's instructor")
		}
		shared, e = svc.StageOMRCheck(ctx, tx, instructor, h.org, shorter.ID, template.ID, "same-sheet-different-assessment.png", image)
		return e
	})
	deadline = time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		academicTx(t, h, func(tx pgx.Tx) error {
			var e error
			shared, e = svc.OMRCheck(ctx, tx, instructor, h.org, shared.ID)
			return e
		})
		if shared.State == "completed" {
			break
		}
		if shared.State == "failed" {
			t.Fatal(shared.Error)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if len(shared.Sheets) != 1 || shared.TemplateID != template.ID || shared.Sheets[0].Maximum != "12" || shared.Sheets[0].Score != "7" || shared.Sheets[0].State != "graded" {
		t.Fatalf("generic template or independent key failed: %+v", shared)
	}
	if e := scholar.Tx(ctx, h.db, func(tx pgx.Tx) error {
		var reference string
		if e := tx.QueryRow(ctx, "SELECT reference_asset_id FROM omr_template WHERE id=$1", template.ID).Scan(&reference); e != nil {
			return e
		}
		_, e := svc.Asset(ctx, tx, instructor, h.org, reference)
		return e
	}); e == nil {
		t.Fatal("shared metadata exposed another course's private reference image")
	}
}
