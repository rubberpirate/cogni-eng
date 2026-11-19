//go:build integration

package transport

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
	"learning.local/workspace/internal/scholar"
)

func TestAssessmentDraftPersistsScopesVersionsAndPublishesNewSnapshot(t *testing.T) {
	h, svc, course, learner := academicFixture(t)
	original, _, q := academicPaper(t, h, svc, course, learner)
	ctx := context.Background()
	var d scholar.AssessmentDraft
	in := scholar.DraftWrite{CampusID: h.campus, CourseID: course, Spec: scholar.DraftSpec{Title: "Faculty draft workflow", Method: "manual", TopicID: q.TopicID, Exam: "IIT-JEE", StartsAt: time.Now().Add(time.Hour), DurationMinutes: 60, Instructions: "Choose the correct answer.", Learners: []string{learner}, Questions: []academic.Question{}, JobIDs: []string{}}}
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		d, err = svc.SaveDraft(ctx, tx, h.owner, h.org, "", in)
		return err
	})
	academicTx(t, h, func(tx pgx.Tx) error {
		loaded, err := svc.Draft(ctx, tx, h.owner, h.org, d.ID)
		if err == nil && (loaded.Spec.Title != in.Spec.Title || len(loaded.Spec.Questions) != 0) {
			t.Fatal("incomplete setup not retained")
		}
		return err
	})
	instructor := access.Random()
	h.exec("INSERT INTO person(id,email,name,password_hash,state,verified_at) VALUES($1,$2,'Scoped faculty','unused','active',now())", instructor, instructor+"@example.invalid")
	h.exec("INSERT INTO membership(id,institution_id,person_id,campus_id,role,state) VALUES($1,$2,$3,$4,'instructor','active')", access.Random(), h.org, instructor, h.campus)
	responsibility := access.Random()
	h.exec("INSERT INTO responsibility(id,institution_id,staff_id,course_id,subject,starts_on) VALUES($1,$2,$3,$4,'Mathematics',current_date)", responsibility, h.org, instructor, course)
	academicTx(t, h, func(tx pgx.Tx) error { _, err := svc.Draft(ctx, tx, instructor, h.org, d.ID); return err })
	h.exec("UPDATE responsibility SET starts_on=current_date+1 WHERE id=$1", responsibility)
	academicTx(t, h, func(tx pgx.Tx) error {
		if _, err := svc.Draft(ctx, tx, instructor, h.org, d.ID); err == nil {
			t.Fatal("empty draft exposed after faculty scope revoked")
		}
		return nil
	})
	academicTx(t, h, func(tx pgx.Tx) error {
		if _, err := svc.Draft(ctx, tx, h.owner, h.otherOrg, d.ID); err == nil {
			t.Fatal("cross-institution draft exposed")
		}
		return nil
	})
	q.Solution = "" // Faculty-selected answers can publish with truthful marking guidance.
	q.Marks = "4"
	q.Penalty = "1"
	in.Version = d.Version
	in.Spec.Questions = []academic.Question{q}
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		d, err = svc.SaveDraft(ctx, tx, h.owner, h.org, d.ID, in)
		return err
	})
	academicTx(t, h, func(tx pgx.Tx) error {
		if _, err := svc.SaveDraft(ctx, tx, h.owner, h.org, d.ID, in); err == nil {
			t.Fatal("stale edit overwrote draft")
		}
		return nil
	})
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		d, err = svc.PublishDraft(ctx, tx, h.owner, h.org, d.ID, d.Version)
		return err
	})
	if d.State != "published" || d.PaperID == "" {
		t.Fatal("draft not published")
	}
	academicTx(t, h, func(tx pgx.Tx) error {
		old, err := svc.Paper(ctx, tx, h.owner, h.org, original.ID)
		if err != nil {
			return err
		}
		if old.Maximum != "2" {
			t.Fatal("published original changed")
		}
		p, err := svc.Paper(ctx, tx, h.owner, h.org, d.PaperID)
		if err != nil {
			return err
		}
		if p.Maximum != "4" {
			t.Fatal("new marking scheme not pinned")
		}
		if !strings.Contains(p.Spec.Sections[0].Questions[0].Solution, "Faculty-approved marking guidance") {
			t.Fatal("missing truthful marking guidance")
		}
		assigned, err := svc.Assignments(ctx, tx, h.owner, h.org, p.ID)
		if err == nil && len(assigned) != 1 {
			t.Fatal("students not assigned")
		}
		return err
	})
	academicTx(t, h, func(tx pgx.Tx) error {
		again, err := svc.PublishDraft(ctx, tx, h.owner, h.org, d.ID, d.Version)
		if err == nil && again.PaperID != d.PaperID {
			t.Fatal("repeat publish duplicated paper")
		}
		return err
	})
}
