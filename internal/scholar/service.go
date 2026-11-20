package scholar

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
	"learning.local/workspace/internal/blob"
)

type Error struct {
	Status        int
	Code, Message string
}

func (e *Error) Error() string                       { return e.Message }
func Problem(status int, code, message string) error { return &Error{status, code, message} }
func Invalid(err error) error                        { return Problem(400, "academic_validation", err.Error()) }

type Service struct {
	DB      *pgxpool.Pool
	Objects *blob.Store
	Queue   *river.Client[pgx.Tx]
	Models  ModelProvider
}

func New(db *pgxpool.Pool, objects *blob.Store) (*Service, error) {
	omr := os.Getenv("OMR_MODEL_PROVIDER")
	if omr != "" && omr != "bubble_classifier" && omr != "qwen_vision" {
		return nil, errors.New("OMR_MODEL_PROVIDER must be bubble_classifier or qwen_vision")
	}
	ocr := os.Getenv("OCR_PROVIDER")
	if ocr != "" && ocr != "tesseract" && ocr != "lightonocr" {
		return nil, errors.New("OCR_PROVIDER must be tesseract or lightonocr")
	}
	provider := os.Getenv("ACADEMIC_MODEL_PROVIDER")
	if provider != "" && provider != "local" && provider != "anthropic" && provider != "gemini" && provider != "development_fixture" {
		return nil, errors.New("ACADEMIC_MODEL_PROVIDER must be local, gemini, anthropic or development_fixture")
	}
	if provider == "development_fixture" && os.Getenv("APP_ENV") != "development" && os.Getenv("APP_ENV") != "test" {
		return nil, errors.New("development model fixtures are forbidden outside development/test")
	}
	q, err := river.NewClient(riverpgxv5.New(db), &river.Config{})
	if err != nil {
		return nil, err
	}
	return &Service{DB: db, Objects: objects, Queue: q, Models: ProviderFromEnv()}, nil
}
func Migrate(ctx context.Context, db *pgxpool.Pool) error {
	m, err := rivermigrate.New(riverpgxv5.New(db), &rivermigrate.Config{})
	if err != nil {
		return err
	}
	_, err = m.Migrate(ctx, rivermigrate.DirectionUp, nil)
	return err
}
func Tx(ctx context.Context, db *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if e := tx.Rollback(ctx); e != nil && !errors.Is(e, pgx.ErrTxClosed) {
			slog.Warn("academic transaction rollback failed", "error", e)
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func Event(ctx context.Context, tx pgx.Tx, org, actor, operation, resource, reason string) error {
	var person any
	if actor != "" {
		person = actor
	}
	_, err := tx.Exec(ctx, "INSERT INTO audit_entry(institution_id,actor_id,action,resource_id,reason,request_id) VALUES($1,$2,$3,$4,$5,$6)", org, person, operation, resource, reason, access.Random()[:16])
	return err
}
func Scope(ctx context.Context, tx pgx.Tx, actor, org, campus, course, subject string) error {
	var state string
	var enforced, enabled bool
	err := tx.QueryRow(ctx, "SELECT i.state,i.require_staff_mfa,p.mfa_enabled FROM institution i CROSS JOIN person p WHERE i.id=$1 AND p.id=$2 AND p.state='active'", org, actor).Scan(&state, &enforced, &enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return Problem(403, "academic_forbidden", "This academic workspace is unavailable.")
	}
	if err != nil {
		return err
	}
	if state != "active" {
		return Problem(403, "workspace_inactive", "Institution is not active.")
	}
	if enforced && !enabled {
		return Problem(403, "mfa_required", "Enable authenticator protection in Access and security.")
	}
	var allowed bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM membership m WHERE m.institution_id=$1 AND m.person_id=$2 AND m.state='active' AND (m.campus_id IS NULL OR m.campus_id=$3) AND (m.role IN ('owner','coordinator') OR (m.role='instructor' AND EXISTS(SELECT 1 FROM responsibility r JOIN course c ON c.id=r.course_id AND c.institution_id=r.institution_id WHERE r.institution_id=$1 AND r.staff_id=$2 AND ($4='' OR r.course_id=$4) AND ($5='' OR lower(r.subject)=lower($5)) AND c.campus_id=$3 AND r.starts_on<=current_date AND (r.ends_on IS NULL OR r.ends_on>=current_date)))))`, org, actor, campus, course, subject).Scan(&allowed)
	if err != nil {
		return err
	}
	if !allowed {
		return Problem(403, "academic_forbidden", "Your campus, course and subject responsibilities do not permit this operation.")
	}
	return nil
}
func LearnerScope(ctx context.Context, tx pgx.Tx, actor, org, campus, course string) error {
	var allowed bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM enrollment e JOIN membership m ON m.institution_id=e.institution_id AND m.person_id=e.person_id AND m.campus_id=e.campus_id JOIN institution i ON i.id=e.institution_id WHERE e.institution_id=$1 AND e.person_id=$2 AND e.campus_id=$3 AND e.course_id=$4 AND e.state='active' AND m.state='active' AND m.role='learner' AND i.state='active' AND e.start_date<=current_date AND (e.end_date IS NULL OR e.end_date>=current_date))`, org, actor, campus, course).Scan(&allowed)
	if err != nil {
		return err
	}
	if !allowed {
		return Problem(403, "learner_ineligible", "An active learner enrollment for this course is required.")
	}
	return nil
}
func rows[T any](ctx context.Context, tx pgx.Tx, query string, args ...any) ([]T, error) {
	r, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	out, err := pgx.CollectRows(r, pgx.RowToStructByPos[T])
	if out == nil {
		out = []T{}
	}
	return out, err
}
func Encode(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
func fingerprint(q academic.Question) string {
	options := []string{}
	for _, o := range q.Options {
		options = append(options, strings.Join(strings.Fields(strings.ToLower(o.Text)), " ")+o.AssetID)
	}
	return access.Digest(q.Type + strings.Join(strings.Fields(strings.ToLower(q.Stem)), " ") + strings.Join(options, "|") + q.AssetID)
}
func FindQuestion(p academic.Paper, id string) (academic.Question, bool) {
	for _, s := range p.Sections {
		for _, q := range s.Questions {
			if q.ID == id {
				return q, true
			}
		}
	}
	return academic.Question{}, false
}

type Topic struct {
	ID       string `json:"id"`
	CampusID string `json:"campus_id"`
	CourseID string `json:"course_id"`
	Subject  string `json:"subject"`
	Name     string `json:"name"`
	Outcome  string `json:"outcome"`
	Language string `json:"language"`
}

func GetTopic(ctx context.Context, tx pgx.Tx, org, id string) (Topic, error) {
	var t Topic
	err := tx.QueryRow(ctx, "SELECT id,campus_id,course_id,subject,name,outcome,language FROM topic WHERE institution_id=$1 AND id=$2", org, id).Scan(&t.ID, &t.CampusID, &t.CourseID, &t.Subject, &t.Name, &t.Outcome, &t.Language)
	if errors.Is(err, pgx.ErrNoRows) {
		return t, Problem(404, "topic_not_found", "Topic unavailable.")
	}
	return t, err
}
func (s *Service) Topics(ctx context.Context, tx pgx.Tx, actor, org, campus string) ([]Topic, error) {
	list, err := rows[Topic](ctx, tx, "SELECT id,campus_id,course_id,subject,name,outcome,language FROM topic WHERE institution_id=$1 AND ($2='' OR campus_id=$2) ORDER BY subject,name LIMIT 500", org, campus)
	if err != nil {
		return nil, err
	}
	out := []Topic{}
	for _, t := range list {
		if err = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err == nil {
			out = append(out, t)
		} else if _, ok := err.(*Error); !ok {
			return nil, err
		}
	}
	return out, nil
}
func (s *Service) CreateTopic(ctx context.Context, tx pgx.Tx, actor, org string, t Topic) (Topic, error) {
	t.Subject = strings.TrimSpace(t.Subject)
	t.Name = strings.TrimSpace(t.Name)
	if len(t.Subject) < 2 || len(t.Subject) > 100 || len(t.Name) < 2 || len(t.Name) > 150 || len(t.Outcome) > 4000 || len(t.Outcome) < 5 || t.Language != "en" && t.Language != "hi" {
		return t, Problem(400, "invalid_topic", "Provide subject, topic, learning outcome and language.")
	}
	if err := Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err != nil {
		return t, err
	}
	var valid bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM course WHERE institution_id=$1 AND id=$2 AND campus_id=$3 AND state='active')", org, t.CourseID, t.CampusID).Scan(&valid); err != nil {
		return t, err
	}
	if !valid {
		return t, Problem(400, "invalid_course", "Select an active course in the selected campus.")
	}
	t.ID = access.Random()
	_, err := tx.Exec(ctx, "INSERT INTO topic(id,institution_id,campus_id,course_id,subject,name,outcome,language) VALUES($1,$2,$3,$4,$5,$6,$7,$8)", t.ID, org, t.CampusID, t.CourseID, t.Subject, t.Name, t.Outcome, t.Language)
	if err != nil {
		return t, err
	}
	return t, Event(ctx, tx, org, actor, "topic.created", t.ID, "")
}
func (s *Service) Bank(ctx context.Context, tx pgx.Tx, actor, org, topicID string) ([]academic.Question, error) {
	t, err := GetTopic(ctx, tx, org, topicID)
	if err != nil {
		return nil, err
	}
	if err = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err != nil {
		return nil, err
	}
	type stored struct {
		ID, Reference, State, Origin string
		Version                      int
		Content                      []byte
	}
	list, err := rows[stored](ctx, tx, "SELECT v.id,q.reference,v.state,v.origin,q.version,v.content FROM question_record q JOIN question_revision v ON v.institution_id=q.institution_id AND v.id=q.latest_revision WHERE q.institution_id=$1 AND q.topic_id=$2 ORDER BY q.created_at DESC LIMIT 500", org, topicID)
	if err != nil {
		return nil, err
	}
	out := []academic.Question{}
	for _, v := range list {
		var q academic.Question
		if err = json.Unmarshal(v.Content, &q); err != nil {
			return nil, err
		}
		q.ID = v.ID
		q.Reference = v.Reference
		q.State = v.State
		q.Origin = v.Origin
		q.Version = v.Version
		out = append(out, q)
	}
	return out, nil
}
func (s *Service) CreateQuestion(ctx context.Context, tx pgx.Tx, actor, org string, q academic.Question, origin string) (academic.Question, bool, error) {
	t, err := GetTopic(ctx, tx, org, q.TopicID)
	if err != nil {
		return q, false, err
	}
	if err = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err != nil {
		return q, false, err
	}
	if err = academic.ValidateQuestion(q); err != nil {
		return q, false, Invalid(err)
	}
	if err = s.validateQuestionAssets(ctx, tx, org, t, q); err != nil {
		return q, false, err
	}
	fp := fingerprint(q)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", org+q.TopicID+fp); err != nil {
		return q, false, err
	}
	var existing string
	err = tx.QueryRow(ctx, "SELECT latest_revision FROM question_record WHERE institution_id=$1 AND fingerprint=$2 AND topic_id=$3", org, fp, q.TopicID).Scan(&existing)
	if err == nil {
		return academic.Question{ID: existing}, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return q, false, err
	}
	record, id := access.Random(), access.Random()
	var reference string
	err = tx.QueryRow(ctx, "INSERT INTO question_record(id,institution_id,topic_id,fingerprint,created_by) VALUES($1,$2,$3,$4,$5) RETURNING reference", record, org, q.TopicID, fp, actor).Scan(&reference)
	if err != nil {
		return q, false, err
	}
	q.ID = id
	q.Reference = reference
	q.State = "draft"
	q.Version = 1
	q.Origin = origin
	_, err = tx.Exec(ctx, "INSERT INTO question_revision(id,institution_id,record_id,content,state,origin,created_by) VALUES($1,$2,$3,$4,'draft',$5,$6)", id, org, record, Encode(q), origin, actor)
	if err != nil {
		return q, false, err
	}
	_, err = tx.Exec(ctx, "UPDATE question_record SET latest_revision=$2 WHERE id=$1", record, id)
	if err != nil {
		return q, false, err
	}
	return q, false, Event(ctx, tx, org, actor, "question.drafted", id, origin)
}
func (s *Service) validateQuestionAssets(ctx context.Context, tx pgx.Tx, org string, t Topic, q academic.Question) error {
	ids := []string{q.AssetID}
	for _, o := range q.Options {
		ids = append(ids, o.AssetID)
	}
	for _, id := range ids {
		if id == "" {
			continue
		}
		var valid bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM academic_asset WHERE institution_id=$1 AND id=$2 AND course_id=$3 AND purpose='question_image' AND state='ready')", org, id, t.CourseID).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return Problem(400, "invalid_asset", "An image must be a ready question asset in this course.")
		}
	}
	return nil
}
func (s *Service) ReviewQuestion(ctx context.Context, tx pgx.Tx, actor, org, id, state, note string, version int) error {
	if state != "approved" && state != "review" && state != "retired" {
		return Problem(400, "invalid_review", "Choose approved, review or retired.")
	}
	if len(strings.TrimSpace(note)) < 3 || len(note) > 2000 {
		return Problem(400, "review_note_required", "Provide a short review note.")
	}
	var content []byte
	var topicID string
	var actual int
	err := tx.QueryRow(ctx, "SELECT v.content,q.topic_id,q.version FROM question_revision v JOIN question_record q ON q.id=v.record_id AND q.institution_id=v.institution_id WHERE v.institution_id=$1 AND v.id=$2 AND q.latest_revision=v.id FOR UPDATE OF q,v", org, id).Scan(&content, &topicID, &actual)
	if errors.Is(err, pgx.ErrNoRows) {
		return Problem(409, "revision_changed", "Refresh the latest question revision.")
	}
	if err != nil {
		return err
	}
	if actual != version {
		return Problem(409, "version_conflict", "This question changed. Refresh first.")
	}
	t, err := GetTopic(ctx, tx, org, topicID)
	if err != nil {
		return err
	}
	if err = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err != nil {
		return err
	}
	var q academic.Question
	if err = json.Unmarshal(content, &q); err != nil {
		return err
	}
	if state == "approved" {
		if len(strings.TrimSpace(q.Solution)) < 3 {
			return Problem(400, "solution_required", "Provide a worked solution or marking guidance before approval.")
		}
		if err = academic.ValidateQuestion(q); err != nil {
			return Invalid(err)
		}
		if err = s.validateQuestionAssets(ctx, tx, org, t, q); err != nil {
			return err
		}
	}
	if _, err = tx.Exec(ctx, "UPDATE question_revision SET state=$3,reviewed_by=$4,review_note=$5 WHERE institution_id=$1 AND id=$2", org, id, state, actor, note); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "UPDATE question_record SET version=version+1 WHERE institution_id=$1 AND latest_revision=$2", org, id); err != nil {
		return err
	}
	return Event(ctx, tx, org, actor, "question."+state, id, note)
}
func (s *Service) ReviseQuestion(ctx context.Context, tx pgx.Tx, actor, org, id string, q academic.Question, version int) (academic.Question, error) {
	var record, topicID, reference, origin string
	var actual int
	err := tx.QueryRow(ctx, "SELECT q.id,q.topic_id,q.reference,q.version,v.origin FROM question_record q JOIN question_revision v ON v.id=q.latest_revision AND v.institution_id=q.institution_id WHERE q.institution_id=$1 AND q.latest_revision=$2 FOR UPDATE OF q", org, id).Scan(&record, &topicID, &reference, &actual, &origin)
	if errors.Is(err, pgx.ErrNoRows) {
		return q, Problem(409, "revision_changed", "Refresh the latest question before editing.")
	}
	if err != nil {
		return q, err
	}
	if actual != version {
		return q, Problem(409, "version_conflict", "Question changed. Refresh before saving.")
	}
	q.TopicID = topicID
	t, err := GetTopic(ctx, tx, org, topicID)
	if err != nil {
		return q, err
	}
	if err = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err != nil {
		return q, err
	}
	if err = academic.ValidateQuestion(q); err != nil {
		return q, Invalid(err)
	}
	if err = s.validateQuestionAssets(ctx, tx, org, t, q); err != nil {
		return q, err
	}
	q.ID = access.Random()
	q.Reference = reference
	q.State = "draft"
	q.Origin = origin
	q.Version = actual + 1
	if _, err = tx.Exec(ctx, "INSERT INTO question_revision(id,institution_id,record_id,content,state,origin,created_by) VALUES($1,$2,$3,$4,'draft',$5,$6)", q.ID, org, record, Encode(q), origin, actor); err != nil {
		return q, err
	}
	if _, err = tx.Exec(ctx, "UPDATE question_record SET latest_revision=$2,version=version+1,fingerprint=$3 WHERE id=$1", record, q.ID, fingerprint(q)); err != nil {
		return q, err
	}
	return q, Event(ctx, tx, org, actor, "question.revised", q.ID, "Previous revisions retained")
}

// Each application job contains only an opaque durable task reference.
type TaskArgs struct {
	ID string `json:"id"`
}

func (TaskArgs) Kind() string { return "academic_task_v1" }

type ExpiryArgs struct {
	InstitutionID string `json:"institution_id"`
	SittingID     string `json:"sitting_id"`
}

func (ExpiryArgs) Kind() string { return "sitting_expiry_v1" }
func (s *Service) enqueue(ctx context.Context, tx pgx.Tx, id string) error {
	job, err := s.Queue.InsertTx(ctx, tx, TaskArgs{ID: id}, &river.InsertOpts{MaxAttempts: 3})
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "UPDATE academic_job SET river_id=$2 WHERE id=$1", id, job.Job.ID)
	return err
}
func (s *Service) expireAt(ctx context.Context, tx pgx.Tx, org, id string, deadline time.Time) error {
	_, err := s.Queue.InsertTx(ctx, tx, ExpiryArgs{org, id}, &river.InsertOpts{ScheduledAt: deadline, MaxAttempts: 10})
	return err
}

type Course struct {
	ID           string `json:"id"`
	CampusID     string `json:"campus_id"`
	Name         string `json:"name"`
	AcademicYear string `json:"academic_year"`
}

func (s *Service) Courses(ctx context.Context, tx pgx.Tx, actor, org, campus string) ([]Course, error) {
	list, err := rows[Course](ctx, tx, "SELECT id,campus_id,name,academic_year FROM course WHERE institution_id=$1 AND ($2='' OR campus_id=$2) AND state='active' ORDER BY name LIMIT 200", org, campus)
	if err != nil {
		return nil, err
	}
	out := []Course{}
	for _, c := range list {
		if err = Scope(ctx, tx, actor, org, c.CampusID, c.ID, ""); err == nil {
			out = append(out, c)
		} else if _, known := err.(*Error); !known {
			return nil, err
		}
	}
	return out, nil
}
