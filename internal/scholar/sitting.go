package scholar

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
)

type Sitting struct {
	ID            string                         `json:"id"`
	AssignmentID  string                         `json:"assignment_id"`
	LearnerID     string                         `json:"learner_id"`
	Channel       string                         `json:"channel"`
	State         string                         `json:"state"`
	StartedAt     time.Time                      `json:"started_at"`
	Deadline      time.Time                      `json:"deadline"`
	Writer        string                         `json:"-"`
	Answers       map[string]academic.Answer     `json:"answers"`
	Manual        map[string]academic.ManualMark `json:"manual_marks,omitempty"`
	AnswerVersion int                            `json:"answer_version"`
	Result        *academic.Result               `json:"result,omitempty"`
	Revision      int                            `json:"revision"`
	Paper         PaperRecord                    `json:"paper"`
	ServerTime    time.Time                      `json:"server_time"`
}

func loadSitting(ctx context.Context, tx pgx.Tx, org, id string) (Sitting, error) {
	var out Sitting
	var answers, manual, result []byte
	var paper string
	err := tx.QueryRow(ctx, `SELECT s.id,s.assignment_id,s.learner_id,s.channel,s.state,s.started_at,s.deadline,coalesce(s.writer_session,''),s.answers,s.manual_marks,s.answer_version,s.result,s.marking_revision,a.paper_id FROM sitting s JOIN assignment a ON a.id=s.assignment_id AND a.institution_id=s.institution_id WHERE s.institution_id=$1 AND s.id=$2 FOR UPDATE OF s`, org, id).Scan(&out.ID, &out.AssignmentID, &out.LearnerID, &out.Channel, &out.State, &out.StartedAt, &out.Deadline, &out.Writer, &answers, &manual, &out.AnswerVersion, &result, &out.Revision, &paper)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, Problem(404, "sitting_unavailable", "Sitting unavailable.")
	}
	if err != nil {
		return out, err
	}
	if err = json.Unmarshal(answers, &out.Answers); err != nil {
		return out, err
	}
	if err = json.Unmarshal(manual, &out.Manual); err != nil {
		return out, err
	}
	if len(result) > 0 {
		if err = json.Unmarshal(result, &out.Result); err != nil {
			return out, err
		}
	}
	out.Paper, err = loadPaper(ctx, tx, org, paper)
	out.ServerTime = time.Now().UTC()
	return out, err
}
func (s *Service) Start(ctx context.Context, tx pgx.Tx, actor, session, org, assignment string, takeover bool) (Sitting, error) {
	var learner, paper string
	var opens, closes time.Time
	var extra int
	err := tx.QueryRow(ctx, "SELECT learner_id,paper_id,opens_at,closes_at,extra_minutes FROM assignment WHERE institution_id=$1 AND id=$2 FOR UPDATE", org, assignment).Scan(&learner, &paper, &opens, &closes, &extra)
	if errors.Is(err, pgx.ErrNoRows) {
		return Sitting{}, Problem(404, "assignment_unavailable", "Assignment unavailable.")
	}
	if err != nil {
		return Sitting{}, err
	}
	if learner != actor {
		return Sitting{}, Problem(403, "sitting_forbidden", "Only the assigned learner may start this sitting.")
	}
	p, err := loadPaper(ctx, tx, org, paper)
	if err != nil {
		return Sitting{}, err
	}
	if err = LearnerScope(ctx, tx, actor, org, p.CampusID, p.CourseID); err != nil {
		return Sitting{}, err
	}
	var id string
	err = tx.QueryRow(ctx, "SELECT id FROM sitting WHERE assignment_id=$1", assignment).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		now := time.Now().UTC()
		if now.Before(opens) || !now.Before(closes) || p.State != "published" {
			return Sitting{}, Problem(409, "outside_window", "This assessment is outside its assignment window.")
		}
		deadline := now.Add(time.Duration(p.Spec.DurationMinutes+extra) * time.Minute)
		if deadline.After(closes) {
			deadline = closes
		}
		id = access.Random()
		_, err = tx.Exec(ctx, "INSERT INTO sitting(id,institution_id,assignment_id,learner_id,channel,state,started_at,deadline,writer_session) VALUES($1,$2,$3,$4,'online','active',$5,$6,$7)", id, org, assignment, actor, now, deadline, session)
		if err != nil {
			return Sitting{}, err
		}
		if err = s.expireAt(ctx, tx, org, id, deadline); err != nil {
			return Sitting{}, err
		}
		if err = Event(ctx, tx, org, actor, "sitting.started", id, ""); err != nil {
			return Sitting{}, err
		}
	} else if err != nil {
		return Sitting{}, err
	}
	out, err := loadSitting(ctx, tx, org, id)
	if err != nil {
		return out, err
	}
	if out.State == "active" && out.Channel == "online" && out.Writer != session {
		if !takeover {
			return out, Problem(409, "writer_conflict", "Another signed-in session owns this sitting. Explicitly resume here to transfer control.")
		}
		if _, err = tx.Exec(ctx, "UPDATE sitting SET writer_session=$2 WHERE id=$1", id, session); err != nil {
			return out, err
		}
		if err = Event(ctx, tx, org, actor, "sitting.writer_transferred", id, "Learner explicitly resumed on this session"); err != nil {
			return out, err
		}
	}
	return s.Sitting(ctx, tx, actor, org, id)
}
func (s *Service) Sitting(ctx context.Context, tx pgx.Tx, actor, org, id string) (Sitting, error) {
	out, err := loadSitting(ctx, tx, org, id)
	if err != nil {
		return out, err
	}
	learner := out.LearnerID == actor
	if learner {
		err = LearnerScope(ctx, tx, actor, org, out.Paper.CampusID, out.Paper.CourseID)
	} else {
		err = paperStaff(ctx, tx, actor, org, out.Paper)
	}
	if err != nil {
		return Sitting{}, err
	}
	if out.State == "active" && !time.Now().Before(out.Deadline) {
		if err = s.finalize(ctx, tx, org, &out, "", "server_expiry"); err != nil {
			return out, err
		}
	}
	if learner && out.State != "released" {
		out.Result = nil
		out.Manual = nil
		out.Paper.Spec = academic.PublicPaper(out.Paper.Spec)
	}
	if learner {
		out.Manual = nil
	}
	return out, nil
}
func (s *Service) Save(ctx context.Context, tx pgx.Tx, actor, session, org, id, operation string, version int, answers map[string]academic.Answer) (Sitting, error) {
	out, err := loadSitting(ctx, tx, org, id)
	if err != nil {
		return out, err
	}
	if actor != out.LearnerID {
		return out, Problem(403, "sitting_forbidden", "Only the assigned learner may save answers.")
	}
	if err = LearnerScope(ctx, tx, actor, org, out.Paper.CampusID, out.Paper.CourseID); err != nil {
		return out, err
	}
	if len(operation) < 8 || len(operation) > 100 {
		return out, Problem(400, "operation_required", "A unique save operation reference is required.")
	}
	var old int
	err = tx.QueryRow(ctx, "SELECT version FROM response_operation WHERE sitting_id=$1 AND operation_id=$2", id, operation).Scan(&old)
	if err == nil {
		return s.Sitting(ctx, tx, actor, org, id)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return out, err
	}
	if out.State != "active" || out.Channel != "online" || !time.Now().Before(out.Deadline) {
		if out.State == "active" {
			if err = s.finalize(ctx, tx, org, &out, "", "server_expiry"); err != nil {
				return out, err
			}
		}
		return s.Sitting(ctx, tx, actor, org, id)
	}
	if out.Writer != session {
		return out, Problem(409, "writer_conflict", "Resume this sitting here before saving.")
	}
	if version != out.AnswerVersion {
		return out, Problem(409, "answer_conflict", "Saved answers changed. Reload before merging your local work.")
	}
	if len(answers) > 200 {
		return out, Problem(400, "invalid_answers", "Too many answer entries.")
	}
	maxOrder := 0
	for _, a := range out.Answers {
		if a.Order > maxOrder {
			maxOrder = a.Order
		}
	}
	for question, a := range answers {
		q, ok := FindQuestion(out.Paper.Spec, question)
		if !ok {
			return out, Problem(400, "unknown_question", "Answer references a question outside this sitting.")
		}
		if err = academic.ValidateAnswer(q, a); err != nil {
			return out, Invalid(err)
		}
		a.Order = out.Answers[question].Order
		if a.Order == 0 && !academic.Blank(a) {
			maxOrder++
			a.Order = maxOrder
		}
		out.Answers[question] = a
	}
	_, err = tx.Exec(ctx, "UPDATE sitting SET answers=$2,answer_version=answer_version+1 WHERE id=$1", id, Encode(out.Answers))
	if err != nil {
		return out, err
	}
	_, err = tx.Exec(ctx, "INSERT INTO response_operation(sitting_id,operation_id,version) VALUES($1,$2,$3)", id, operation, out.AnswerVersion+1)
	if err != nil {
		return out, err
	}
	return s.Sitting(ctx, tx, actor, org, id)
}
func (s *Service) Submit(ctx context.Context, tx pgx.Tx, actor, session, org, id string) (Sitting, error) {
	out, err := loadSitting(ctx, tx, org, id)
	if err != nil {
		return out, err
	}
	if out.LearnerID != actor {
		return out, Problem(403, "sitting_forbidden", "Only the assigned learner may submit.")
	}
	if err = LearnerScope(ctx, tx, actor, org, out.Paper.CampusID, out.Paper.CourseID); err != nil {
		return out, err
	}
	if out.State == "active" {
		if out.Writer != session {
			return out, Problem(409, "writer_conflict", "Resume this sitting here before submitting.")
		}
		reason := "submitted"
		if !time.Now().Before(out.Deadline) {
			reason = "server_expiry"
		}
		if err = s.finalize(ctx, tx, org, &out, actor, reason); err != nil {
			return out, err
		}
	}
	return s.Sitting(ctx, tx, actor, org, id)
}
func (s *Service) finalize(ctx context.Context, tx pgx.Tx, org string, out *Sitting, actor, reason string) error {
	result, err := academic.Mark(out.Paper.Spec, out.Answers, out.Manual, out.Channel)
	if err != nil {
		return err
	}
	state := "marked"
	if result.Pending > 0 {
		state = "awaiting_review"
	}
	if result.Pending == 0 && out.Paper.Spec.ReleasePolicy == "immediate" {
		allowed, e := safeRelease(ctx, tx, org, *out)
		if e != nil {
			return e
		}
		if allowed {
			state = "released"
		}
	}
	out.State = state
	out.Result = &result
	out.Revision++
	_, err = tx.Exec(ctx, "UPDATE sitting SET state=$2,result=$3,marking_revision=$4,finished_at=coalesce(finished_at,now()),finish_reason=$5,manual_marks=$6,release_at=CASE WHEN $2='released' THEN now() ELSE release_at END,version=version+1 WHERE id=$1", out.ID, state, Encode(result), out.Revision, reason, Encode(out.Manual))
	if err != nil {
		return err
	}
	var person any
	if actor != "" {
		person = actor
	}
	_, err = tx.Exec(ctx, "INSERT INTO result_revision(id,institution_id,sitting_id,revision,result,reason,actor_id) VALUES($1,$2,$3,$4,$5,$6,$7)", access.Random(), org, out.ID, out.Revision, Encode(result), reason, person)
	if err != nil {
		return err
	}
	if state == "released" {
		if err = expose(ctx, tx, org, *out); err != nil {
			return err
		}
	}
	return Event(ctx, tx, org, actor, "sitting."+state, out.ID, reason)
}
func expose(ctx context.Context, tx pgx.Tx, org string, out Sitting) error {
	for _, sec := range out.Paper.Spec.Sections {
		for _, q := range sec.Questions {
			if _, err := tx.Exec(ctx, "INSERT INTO answer_exposure(institution_id,learner_id,question_id) VALUES($1,$2,$3) ON CONFLICT DO NOTHING", org, out.LearnerID, q.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *Service) Release(ctx context.Context, tx pgx.Tx, actor, org, id string) error {
	out, err := loadSitting(ctx, tx, org, id)
	if err != nil {
		return err
	}
	if err = paperStaff(ctx, tx, actor, org, out.Paper); err != nil {
		return err
	}
	if out.State == "released" {
		return nil
	}
	if out.State != "marked" {
		return Problem(409, "marking_incomplete", "Complete submission and rubric review before release.")
	}
	var closes time.Time
	if err = tx.QueryRow(ctx, "SELECT max(closes_at) FROM assignment WHERE institution_id=$1 AND paper_id=$2", org, out.Paper.ID).Scan(&closes); err != nil {
		return err
	}
	if time.Now().Before(closes) {
		return Problem(409, "release_window_open", "The assignment window is still open. Release after all recipients' deadlines.")
	}
	allowed, e := safeRelease(ctx, tx, org, out)
	if e != nil {
		return e
	}
	if !allowed {
		return Problem(409, "formal_exposure_conflict", "A current formal assignment contains this item. Release after its window, or replace that assignment under an explicit correction policy.")
	}
	if _, err = tx.Exec(ctx, "UPDATE sitting SET state='released',release_at=now(),version=version+1 WHERE id=$1", id); err != nil {
		return err
	}
	if err = expose(ctx, tx, org, out); err != nil {
		return err
	}
	return Event(ctx, tx, org, actor, "result.released", id, "")
}
func (s *Service) ReviewMarks(ctx context.Context, tx pgx.Tx, actor, org, id, reason string, revision int, manual map[string]academic.ManualMark) (Sitting, error) {
	out, err := loadSitting(ctx, tx, org, id)
	if err != nil {
		return out, err
	}
	if err = paperStaff(ctx, tx, actor, org, out.Paper); err != nil {
		return out, err
	}
	if out.State == "active" || out.Revision != revision || len(reason) < 5 || len(reason) > 2000 {
		return out, Problem(409, "review_conflict", "Refresh completed work and provide a reason for this marking revision.")
	}
	for id, m := range manual {
		q, ok := FindQuestion(out.Paper.Spec, id)
		if !ok || (q.Type != "short" && q.Type != "extended") || !validString(m.Feedback, 4000) {
			return out, Problem(400, "invalid_rubric_review", "Choose written questions and provide criterion marks and feedback.")
		}
		m.Approved = true
		out.Manual[id] = m
	}
	wasReleased := out.State == "released"
	if err = s.finalize(ctx, tx, org, &out, actor, reason); err != nil {
		return out, err
	}
	if wasReleased && out.Result.Pending == 0 {
		if _, err = tx.Exec(ctx, "UPDATE sitting SET state='released' WHERE id=$1", id); err != nil {
			return out, err
		}
	}
	return s.Sitting(ctx, tx, actor, org, id)
}

// Serialize exposure and assignment decisions on the learner. A future formal sitting must never inherit an already-revealed key.
func safeRelease(ctx context.Context, tx pgx.Tx, org string, out Sitting) (bool, error) {
	if _, err := tx.Exec(ctx, "SELECT id FROM person WHERE id=$1 FOR UPDATE", out.LearnerID); err != nil {
		return false, err
	}
	ids := []string{}
	for _, sec := range out.Paper.Spec.Sections {
		for _, q := range sec.Questions {
			ids = append(ids, q.ID)
		}
	}
	var conflict bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM assignment a JOIN paper p ON p.id=a.paper_id AND p.institution_id=a.institution_id CROSS JOIN LATERAL jsonb_array_elements(p.spec->'sections') sec CROSS JOIN LATERAL jsonb_array_elements(sec->'questions') item JOIN question_revision v ON v.id=item->>'id' AND v.institution_id=p.institution_id JOIN question_revision selected ON selected.record_id=v.record_id AND selected.institution_id=v.institution_id WHERE a.institution_id=$1 AND a.learner_id=$2 AND a.closes_at>now() AND p.id<>$3 AND p.spec->>'purpose'='formal' AND selected.id=ANY($4::text[]))`, org, out.LearnerID, out.Paper.ID, ids).Scan(&conflict)
	return !conflict, err
}
