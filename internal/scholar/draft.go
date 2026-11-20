package scholar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
)

type DraftSpec struct {
	Title           string              `json:"title"`
	Method          string              `json:"method"`
	TopicID         string              `json:"topic_id"`
	Exam            string              `json:"exam"`
	StartsAt        time.Time           `json:"starts_at"`
	DurationMinutes int                 `json:"duration_minutes"`
	Instructions    string              `json:"instructions"`
	Learners        []string            `json:"learners"`
	Questions       []academic.Question `json:"questions"`
	ImportID        string              `json:"import_id"`
	JobIDs          []string            `json:"job_ids"`
	AppliedJobs     []string            `json:"applied_jobs"`
	SourceApplied   bool                `json:"source_applied"`
}
type AssessmentDraft struct {
	ID       string    `json:"id"`
	CampusID string    `json:"campus_id"`
	CourseID string    `json:"course_id"`
	Version  int       `json:"version"`
	State    string    `json:"state"`
	PaperID  string    `json:"paper_id"`
	Spec     DraftSpec `json:"spec"`
}
type DraftWrite struct {
	CampusID string    `json:"campus_id"`
	CourseID string    `json:"course_id"`
	Version  int       `json:"version"`
	Spec     DraftSpec `json:"spec"`
}

func (s *Service) Draft(ctx context.Context, tx pgx.Tx, actor, org, id string) (AssessmentDraft, error) {
	var out AssessmentDraft
	var data []byte
	err := tx.QueryRow(ctx, "SELECT id,campus_id,course_id,version,state,coalesce(paper_id,''),spec FROM assessment_draft WHERE institution_id=$1 AND id=$2", org, id).Scan(&out.ID, &out.CampusID, &out.CourseID, &out.Version, &out.State, &out.PaperID, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, Problem(404, "draft_not_found", "Assessment draft unavailable.")
	}
	if err != nil {
		return out, err
	}
	if err = Scope(ctx, tx, actor, org, out.CampusID, out.CourseID, ""); err != nil {
		return out, err
	}
	if err = json.Unmarshal(data, &out.Spec); err != nil {
		return out, err
	}
	if out.Spec.AppliedJobs == nil {
		out.Spec.AppliedJobs = []string{}
	}
	if err = draftTopicScope(ctx, tx, actor, org, out.CampusID, out.CourseID, out.Spec.TopicID); err != nil {
		return out, err
	}
	for _, q := range out.Spec.Questions {
		if err = draftTopicScope(ctx, tx, actor, org, out.CampusID, out.CourseID, q.TopicID); err != nil {
			return out, err
		}
	}
	return out, nil
}
func draftTopicScope(ctx context.Context, tx pgx.Tx, actor, org, campus, course, topic string) error {
	t, err := GetTopic(ctx, tx, org, topic)
	if err != nil {
		return err
	}
	if t.CourseID != course || t.CampusID != campus {
		return Problem(400, "draft_topic", "Each topic must belong to the selected student batch.")
	}
	return Scope(ctx, tx, actor, org, campus, course, t.Subject)
}
func (s *Service) Drafts(ctx context.Context, tx pgx.Tx, actor, org, campus string) ([]AssessmentDraft, error) {
	type item struct{ ID string }
	ids, err := rows[item](ctx, tx, "SELECT id FROM assessment_draft WHERE institution_id=$1 AND ($2='' OR campus_id=$2) ORDER BY updated_at DESC LIMIT 100", org, campus)
	if err != nil {
		return nil, err
	}
	out := []AssessmentDraft{}
	for _, id := range ids {
		d, e := s.Draft(ctx, tx, actor, org, id.ID)
		if e == nil {
			out = append(out, d)
		} else if _, known := e.(*Error); !known {
			return nil, e
		}
	}
	return out, nil
}
func validateDraft(in DraftWrite) error {
	if !validString(in.Spec.Title, 200) || !slices.Contains([]string{"manual", "upload", "ai"}, in.Spec.Method) || in.Spec.StartsAt.IsZero() || in.Spec.DurationMinutes < 1 || in.Spec.DurationMinutes > 480 || len(in.Spec.Questions) > 200 || len(in.Spec.Learners) > 1000 || len(in.Spec.JobIDs) > 100 || len(in.Spec.Instructions) > 10000 {
		return Problem(400, "invalid_draft", "Enter a name, creation method, assessment date, duration (1–480 minutes), and at most 200 questions.")
	}
	return nil
}
func (s *Service) SaveDraft(ctx context.Context, tx pgx.Tx, actor, org, id string, in DraftWrite) (AssessmentDraft, error) {
	if in.Spec.Learners == nil {
		in.Spec.Learners = []string{}
	}
	if in.Spec.Questions == nil {
		in.Spec.Questions = []academic.Question{}
	}
	if in.Spec.JobIDs == nil {
		in.Spec.JobIDs = []string{}
	}
	if in.Spec.AppliedJobs == nil {
		in.Spec.AppliedJobs = []string{}
	}
	if len(in.Spec.AppliedJobs) > 100 {
		return AssessmentDraft{}, Problem(400, "draft_jobs", "Too many applied generation jobs.")
	}
	for _, applied := range in.Spec.AppliedJobs {
		if !slices.Contains(in.Spec.JobIDs, applied) {
			return AssessmentDraft{}, Problem(400, "draft_jobs", "An applied job must belong to this draft's generation history.")
		}
	}

	if err := validateDraft(in); err != nil {
		return AssessmentDraft{}, err
	}
	if err := Scope(ctx, tx, actor, org, in.CampusID, in.CourseID, ""); err != nil {
		return AssessmentDraft{}, err
	}
	if err := draftTopicScope(ctx, tx, actor, org, in.CampusID, in.CourseID, in.Spec.TopicID); err != nil {
		return AssessmentDraft{}, err
	}
	for _, q := range in.Spec.Questions {
		if err := draftTopicScope(ctx, tx, actor, org, in.CampusID, in.CourseID, q.TopicID); err != nil {
			return AssessmentDraft{}, err
		}
	}
	roster, err := s.CourseLearners(ctx, tx, actor, org, in.CourseID)
	if err != nil {
		return AssessmentDraft{}, err
	}
	for _, learner := range in.Spec.Learners {
		if !slices.ContainsFunc(roster, func(r Recipient) bool { return r.ID == learner }) {
			return AssessmentDraft{}, Problem(400, "draft_student", "Select active students from the selected batch.")
		}
	}
	if in.Spec.ImportID != "" {
		imp, e := s.Import(ctx, tx, actor, org, in.Spec.ImportID)
		if e != nil {
			return AssessmentDraft{}, e
		}
		if imp.TopicID != in.Spec.TopicID {
			return AssessmentDraft{}, Problem(400, "draft_import", "This import belongs to another topic.")
		}
	}
	for _, jobID := range in.Spec.JobIDs {
		job, e := s.Job(ctx, tx, actor, org, jobID)
		if e != nil {
			return AssessmentDraft{}, e
		}
		if job.Course != in.CourseID || job.Kind != "generate_questions" {
			return AssessmentDraft{}, Problem(400, "draft_job", "Generation job belongs to another batch.")
		}
	}
	if id == "" {
		id = access.Random()
		_, err = tx.Exec(ctx, "INSERT INTO assessment_draft(id,institution_id,campus_id,course_id,spec,created_by) VALUES($1,$2,$3,$4,$5,$6)", id, org, in.CampusID, in.CourseID, Encode(in.Spec), actor)
	} else {
		old, e := s.Draft(ctx, tx, actor, org, id)
		if e != nil {
			return old, e
		}
		if old.CourseID != in.CourseID || old.CampusID != in.CampusID {
			return old, Problem(409, "draft_batch", "Create a new draft to change the student batch.")
		}
		tag, e := tx.Exec(ctx, "UPDATE assessment_draft SET spec=$3,version=version+1,updated_at=now() WHERE institution_id=$1 AND id=$2 AND version=$4 AND state='draft'", org, id, Encode(in.Spec), in.Version)
		err = e
		if e == nil && tag.RowsAffected() == 0 {
			return old, Problem(409, "draft_changed", "Another edit or publication changed this draft. Reload before saving.")
		}
	}
	if err != nil {
		return AssessmentDraft{}, err
	}
	if err = Event(ctx, tx, org, actor, "assessment.draft_saved", id, "Saved assessment setup and editable questions"); err != nil {
		return AssessmentDraft{}, err
	}
	return s.Draft(ctx, tx, actor, org, id)
}
func (s *Service) PublishDraft(ctx context.Context, tx pgx.Tx, actor, org, id string, version int) (AssessmentDraft, error) {
	if _, err := tx.Exec(ctx, "SELECT id FROM assessment_draft WHERE institution_id=$1 AND id=$2 FOR UPDATE", org, id); err != nil {
		return AssessmentDraft{}, err
	}
	d, err := s.Draft(ctx, tx, actor, org, id)
	if err != nil {
		return d, err
	}
	if d.State == "published" {
		return d, nil
	}
	if d.Version != version {
		return d, Problem(409, "draft_changed", "Reload the latest draft before publishing.")
	}
	if len(d.Spec.Questions) == 0 {
		return d, Problem(400, "draft_incomplete", "Add and review questions before publishing.")
	}
	ids := []string{}
	seen := map[string]bool{}
	for i, q := range d.Spec.Questions {
		fp := fingerprint(q)
		if seen[fp] {
			return d, Problem(400, "duplicate_question", fmt.Sprintf("Question %d duplicates an earlier question.", i+1))
		}
		seen[fp] = true
		q.ID = ""
		q.State = ""
		q.Version = 0
		if err := academic.ValidateQuestion(q); err != nil {
			return d, Problem(400, "draft_question", fmt.Sprintf("Question %d: %s", i+1, err))
		}
		if strings.TrimSpace(q.Solution) == "" {
			q.Solution = "Faculty-approved marking guidance: apply the selected answer rule or rubric; correct response earns " + q.Marks + ", incorrect response deducts " + q.Penalty + ", and a blank earns zero."
		}
		desired := q
		origin := "authored"
		if q.Origin == "generated" {
			origin = "generated"
		} else if q.Origin == "imported" {
			origin = "imported"
		}
		var duplicate bool
		q, duplicate, err = s.CreateQuestion(ctx, tx, actor, org, q, origin)
		if err != nil {
			return d, err
		}
		if duplicate {
			var current int
			if err = tx.QueryRow(ctx, "SELECT r.version FROM question_record r JOIN question_revision v ON v.record_id=r.id AND v.institution_id=r.institution_id WHERE v.institution_id=$1 AND v.id=$2", org, q.ID).Scan(&current); err != nil {
				return d, err
			}
			q, err = s.ReviseQuestion(ctx, tx, actor, org, q.ID, desired, current)
			if err != nil {
				return d, err
			}
		}
		if q.State != "approved" {
			if err = s.ReviewQuestion(ctx, tx, actor, org, q.ID, "approved", "Faculty reviewed draft wording, answer and marking scheme", q.Version); err != nil {
				return d, err
			}
		}
		ids = append(ids, q.ID)
	}
	p, err := s.Compose(ctx, tx, actor, org, Compose{CampusID: d.CampusID, CourseID: d.CourseID, Title: d.Spec.Title, Purpose: "mock", Instructions: d.Spec.Instructions, DurationMinutes: d.Spec.DurationMinutes, ReleasePolicy: "staff", Sections: []ComposeSection{{ID: "main", Title: d.Spec.Exam, Policy: "all", Count: len(ids), QuestionIDs: ids}}})
	if err != nil {
		return d, err
	}
	if err = s.Publish(ctx, tx, actor, org, p.ID, p.Version); err != nil {
		return d, err
	}
	if len(d.Spec.Learners) > 0 {
		if _, err = s.Assign(ctx, tx, actor, org, p.ID, d.Spec.Learners, d.Spec.StartsAt, d.Spec.StartsAt.Add(time.Duration(d.Spec.DurationMinutes)*time.Minute), 0); err != nil {
			return d, err
		}
	}
	_, err = tx.Exec(ctx, "UPDATE assessment_draft SET state='published',paper_id=$3,version=version+1,updated_at=now() WHERE institution_id=$1 AND id=$2", org, id, p.ID)
	if err != nil {
		return d, err
	}
	return s.Draft(ctx, tx, actor, org, id)
}
