package scholar

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
)

type ComposeSection struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Policy      string   `json:"policy"`
	Count       int      `json:"count"`
	Cap         string   `json:"cap"`
	QuestionIDs []string `json:"question_ids"`
}
type Compose struct {
	CampusID        string           `json:"campus_id"`
	CourseID        string           `json:"course_id"`
	Title           string           `json:"title"`
	Purpose         string           `json:"purpose"`
	Instructions    string           `json:"instructions"`
	DurationMinutes int              `json:"duration_minutes"`
	ReleasePolicy   string           `json:"release_policy"`
	Sections        []ComposeSection `json:"sections"`
}
type PaperRecord struct {
	ID        string         `json:"id"`
	CampusID  string         `json:"campus_id"`
	CourseID  string         `json:"course_id"`
	Reference string         `json:"reference"`
	State     string         `json:"state"`
	Version   int            `json:"version"`
	Spec      academic.Paper `json:"spec"`
	Maximum   string         `json:"maximum"`
}

func loadPaper(ctx context.Context, tx pgx.Tx, org, id string) (PaperRecord, error) {
	var out PaperRecord
	var data []byte
	err := tx.QueryRow(ctx, "SELECT id,campus_id,course_id,reference,state,version,spec FROM paper WHERE institution_id=$1 AND id=$2", org, id).Scan(&out.ID, &out.CampusID, &out.CourseID, &out.Reference, &out.State, &out.Version, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, Problem(404, "paper_not_found", "Paper unavailable.")
	}
	if err != nil {
		return out, err
	}
	if err = json.Unmarshal(data, &out.Spec); err != nil {
		return out, err
	}
	marked, err := academic.Mark(out.Spec, map[string]academic.Answer{}, nil, "paper")
	if err != nil {
		return out, err
	}
	out.Maximum = marked.Maximum
	return out, nil
}
func paperStaff(ctx context.Context, tx pgx.Tx, actor, org string, p PaperRecord) error {
	for _, sec := range p.Spec.Sections {
		for _, q := range sec.Questions {
			t, err := GetTopic(ctx, tx, org, q.TopicID)
			if err != nil {
				return err
			}
			if err = Scope(ctx, tx, actor, org, p.CampusID, p.CourseID, t.Subject); err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *Service) Paper(ctx context.Context, tx pgx.Tx, actor, org, id string) (PaperRecord, error) {
	p, err := loadPaper(ctx, tx, org, id)
	if err != nil {
		return p, err
	}
	return p, paperStaff(ctx, tx, actor, org, p)
}
func (s *Service) Papers(ctx context.Context, tx pgx.Tx, actor, org, campus string) ([]PaperRecord, error) {
	type item struct{ ID string }
	ids, err := rows[item](ctx, tx, "SELECT id FROM paper WHERE institution_id=$1 AND ($2='' OR campus_id=$2) ORDER BY created_at DESC LIMIT 100", org, campus)
	if err != nil {
		return nil, err
	}
	out := []PaperRecord{}
	for _, id := range ids {
		p, e := s.Paper(ctx, tx, actor, org, id.ID)
		if e == nil {
			out = append(out, p)
		} else if _, known := e.(*Error); !known {
			return nil, e
		}
	}
	return out, nil
}
func (s *Service) Compose(ctx context.Context, tx pgx.Tx, actor, org string, in Compose) (PaperRecord, error) {
	if len(in.Sections) < 1 || len(in.Sections) > 20 {
		return PaperRecord{}, Problem(400, "invalid_sections", "Use 1–20 sections.")
	}
	count := 0
	for _, sec := range in.Sections {
		count += len(sec.QuestionIDs)
	}
	if count > 200 {
		return PaperRecord{}, Problem(400, "too_many_questions", "Use at most 200 questions per paper.")
	}
	if err := Scope(ctx, tx, actor, org, in.CampusID, in.CourseID, ""); err != nil {
		return PaperRecord{}, err
	}
	p := academic.Paper{Title: in.Title, Purpose: in.Purpose, Instructions: in.Instructions, DurationMinutes: in.DurationMinutes, ReleasePolicy: in.ReleasePolicy}
	for _, input := range in.Sections {
		sec := academic.Section{ID: input.ID, Title: input.Title, Policy: input.Policy, Count: input.Count, Cap: input.Cap, Questions: []academic.Question{}}
		for _, id := range input.QuestionIDs {
			var data []byte
			var state, subject, topicCourse string
			err := tx.QueryRow(ctx, "SELECT v.content,v.state,t.subject,t.course_id FROM question_revision v JOIN question_record r ON r.id=v.record_id AND r.institution_id=v.institution_id JOIN topic t ON t.id=r.topic_id AND t.institution_id=r.institution_id WHERE v.institution_id=$1 AND v.id=$2", org, id).Scan(&data, &state, &subject, &topicCourse)
			if errors.Is(err, pgx.ErrNoRows) {
				return PaperRecord{}, Problem(400, "question_unavailable", "A selected question is unavailable.")
			}
			if err != nil {
				return PaperRecord{}, err
			}
			if state != "approved" || topicCourse != in.CourseID {
				return PaperRecord{}, Problem(400, "question_not_approved", "Every selected question must be approved for this course.")
			}
			if err = Scope(ctx, tx, actor, org, in.CampusID, in.CourseID, subject); err != nil {
				return PaperRecord{}, err
			}
			var q academic.Question
			if err = json.Unmarshal(data, &q); err != nil {
				return PaperRecord{}, err
			}
			q.ID = id
			sec.Questions = append(sec.Questions, q)
		}
		p.Sections = append(p.Sections, sec)
	}
	if err := academic.ValidatePaper(p); err != nil {
		return PaperRecord{}, Invalid(err)
	}
	id := access.Random()
	_, err := tx.Exec(ctx, "INSERT INTO paper(id,institution_id,campus_id,course_id,spec,created_by) VALUES($1,$2,$3,$4,$5,$6)", id, org, in.CampusID, in.CourseID, Encode(p), actor)
	if err != nil {
		return PaperRecord{}, err
	}
	if err = Event(ctx, tx, org, actor, "paper.composed", id, "Pinned selected question revisions in explicit order"); err != nil {
		return PaperRecord{}, err
	}
	return s.Paper(ctx, tx, actor, org, id)
}
func (s *Service) Publish(ctx context.Context, tx pgx.Tx, actor, org, id string, version int) error {
	p, err := s.Paper(ctx, tx, actor, org, id)
	if err != nil {
		return err
	}
	if p.State == "published" {
		return nil
	}
	if p.State != "draft" || p.Version != version {
		return Problem(409, "paper_changed", "Refresh the draft before publication.")
	}
	for _, sec := range p.Spec.Sections {
		for _, q := range sec.Questions {
			var state string
			if err = tx.QueryRow(ctx, "SELECT state FROM question_revision WHERE institution_id=$1 AND id=$2 FOR SHARE", org, q.ID).Scan(&state); err != nil {
				return err
			}
			if state != "approved" {
				return Problem(409, "question_not_approved", "A pinned question is no longer approved.")
			}
			t, e := GetTopic(ctx, tx, org, q.TopicID)
			if e != nil {
				return e
			}
			if err = s.validateQuestionAssets(ctx, tx, org, t, q); err != nil {
				return err
			}
		}
	}
	tag, err := tx.Exec(ctx, "UPDATE paper SET state='published',published_by=$3,published_at=now(),version=version+1 WHERE institution_id=$1 AND id=$2 AND state='draft' AND version=$4", org, id, actor, version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return Problem(409, "version_conflict", "Paper changed during publication.")
	}
	return Event(ctx, tx, org, actor, "paper.published", id, "Immutable content, key, scoring and release snapshot")
}

type Recipient struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Reference string `json:"reference"`
}

func (s *Service) Recipients(ctx context.Context, tx pgx.Tx, actor, org, paper string) ([]Recipient, error) {
	p, err := s.Paper(ctx, tx, actor, org, paper)
	if err != nil {
		return nil, err
	}
	return rows[Recipient](ctx, tx, `SELECT p.id,p.name,e.reference FROM enrollment e JOIN person p ON p.id=e.person_id WHERE e.institution_id=$1 AND e.course_id=$2 AND e.campus_id=$3 AND e.state='active' AND p.state IN ('active','invited') ORDER BY p.name LIMIT 1000`, org, p.CourseID, p.CampusID)
}
func (s *Service) Assign(ctx context.Context, tx pgx.Tx, actor, org, id string, learners []string, opens, closes time.Time, extra int) ([]string, error) {
	p, err := s.Paper(ctx, tx, actor, org, id)
	if err != nil {
		return nil, err
	}
	if p.State != "published" || closes.Before(opens) || !closes.After(time.Now()) || extra < 0 || extra > 480 || len(learners) == 0 || len(learners) > 1000 {
		return nil, Problem(400, "invalid_assignment", "Select recipients, a published paper, a valid future window and approved extra time.")
	}
	out := []string{}
	for _, learner := range slices.Compact(slices.Sorted(slices.Values(learners))) {
		if _, err = tx.Exec(ctx, "SELECT id FROM person WHERE id=$1 FOR UPDATE", learner); err != nil {
			return nil, err
		}
		if err = LearnerScope(ctx, tx, learner, org, p.CampusID, p.CourseID); err != nil {
			return nil, err
		}
		if p.Spec.Purpose == "formal" {
			for _, sec := range p.Spec.Sections {
				for _, q := range sec.Questions {
					var exposed bool
					if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM answer_exposure e JOIN question_revision old ON old.id=e.question_id AND old.institution_id=e.institution_id JOIN question_revision current ON current.record_id=old.record_id AND current.institution_id=old.institution_id WHERE e.institution_id=$1 AND e.learner_id=$2 AND current.id=$3)`, org, learner, q.ID).Scan(&exposed); err != nil {
						return nil, err
					}
					if exposed {
						return nil, Problem(409, "answer_already_exposed", "A recipient has seen a selected question's solution. Use a different item or an explicitly formative paper.")
					}
				}
			}
		}
		assignment := access.Random()
		err = tx.QueryRow(ctx, "INSERT INTO assignment(id,institution_id,paper_id,learner_id,opens_at,closes_at,extra_minutes,assigned_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(paper_id,learner_id) DO UPDATE SET id=assignment.id RETURNING id", assignment, org, id, learner, opens, closes, extra, actor).Scan(&assignment)
		if err != nil {
			return nil, err
		}
		out = append(out, assignment)
	}
	return out, Event(ctx, tx, org, actor, "paper.assigned", id, "Recipient membership frozen; duplicate assignments retain the original window")
}

type Assigned struct {
	ID           string    `json:"id"`
	PaperID      string    `json:"paper_id"`
	Title        string    `json:"title"`
	Reference    string    `json:"reference"`
	LearnerID    string    `json:"learner_id"`
	Learner      string    `json:"learner"`
	OpensAt      time.Time `json:"opens_at"`
	ClosesAt     time.Time `json:"closes_at"`
	ExtraMinutes int       `json:"extra_minutes"`
	State        string    `json:"state"`
	SittingID    string    `json:"sitting_id"`
}

func (s *Service) Assignments(ctx context.Context, tx pgx.Tx, actor, org, paper string) ([]Assigned, error) {
	var list []Assigned
	var err error
	query := `SELECT a.id,p.id,p.spec->>'title',p.reference,a.learner_id,u.name,a.opens_at,a.closes_at,a.extra_minutes,coalesce(s.state,'not_started'),coalesce(s.id,'') FROM assignment a JOIN paper p ON p.id=a.paper_id AND p.institution_id=a.institution_id JOIN person u ON u.id=a.learner_id LEFT JOIN sitting s ON s.assignment_id=a.id WHERE a.institution_id=$1`
	if paper != "" {
		if _, err = s.Paper(ctx, tx, actor, org, paper); err != nil {
			return nil, err
		}
		list, err = rows[Assigned](ctx, tx, query+" AND p.id=$2 ORDER BY u.name LIMIT 1000", org, paper)
	} else {
		list, err = rows[Assigned](ctx, tx, query+" AND a.learner_id=$2 ORDER BY a.opens_at DESC LIMIT 200", org, actor)
		if err != nil {
			return nil, err
		}
		valid := []Assigned{}
		for _, a := range list {
			p, e := loadPaper(ctx, tx, org, a.PaperID)
			if e != nil {
				return nil, e
			}
			if LearnerScope(ctx, tx, actor, org, p.CampusID, p.CourseID) == nil {
				valid = append(valid, a)
			}
		}
		list = valid
	}
	return list, err
}
func (s *Service) Blueprint(ctx context.Context, tx pgx.Tx, actor, org, topic string, count int, difficulty string) ([]string, error) {
	bank, err := s.Bank(ctx, tx, actor, org, topic)
	if err != nil {
		return nil, err
	}
	if count < 1 || count > 100 {
		return nil, Problem(400, "invalid_count", "Choose 1–100 questions.")
	}
	ids := []string{}
	for _, q := range bank {
		if q.State == "approved" && (difficulty == "" || q.Difficulty == difficulty) {
			ids = append(ids, q.ID)
		}
	}
	if len(ids) < count {
		return nil, Problem(409, "thin_question_bank", "The approved bank cannot meet this topic/difficulty count. Author or approve more questions.")
	}
	slices.Sort(ids)
	return ids[:count], nil
}
func validString(s string, n int) bool { return len(strings.TrimSpace(s)) > 0 && len(s) <= n }
