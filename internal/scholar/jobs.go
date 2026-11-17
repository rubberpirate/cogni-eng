package scholar

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"learning.local/workspace/internal/access"
)

type Job struct {
	ID        string          `json:"id"`
	Org       string          `json:"-"`
	Campus    string          `json:"campus_id"`
	Course    string          `json:"course_id"`
	Actor     string          `json:"actor_id"`
	Kind      string          `json:"kind"`
	Resource  string          `json:"resource_id"`
	Input     json.RawMessage `json:"input"`
	State     string          `json:"state"`
	Progress  int             `json:"progress"`
	Error     string          `json:"error"`
	Attempts  int             `json:"attempts"`
	Output    json.RawMessage `json:"output"`
	RiverID   int64           `json:"-"`
	Provider  string          `json:"provider"`
	Model     string          `json:"model"`
	CreatedAt time.Time       `json:"created_at"`
}

func loadJob(ctx context.Context, tx pgx.Tx, org, id string) (Job, error) {
	var j Job
	err := tx.QueryRow(ctx, "SELECT id,institution_id,campus_id,course_id,actor_id,kind,resource_id,input,state,progress,error,attempts,output,coalesce(river_id,0),provider,model,created_at FROM academic_job WHERE institution_id=$1 AND id=$2", org, id).Scan(&j.ID, &j.Org, &j.Campus, &j.Course, &j.Actor, &j.Kind, &j.Resource, &j.Input, &j.State, &j.Progress, &j.Error, &j.Attempts, &j.Output, &j.RiverID, &j.Provider, &j.Model, &j.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return j, Problem(404, "job_unavailable", "Academic job unavailable.")
	}
	return j, err
}
func (s *Service) Job(ctx context.Context, tx pgx.Tx, actor, org, id string) (Job, error) {
	j, err := loadJob(ctx, tx, org, id)
	if err != nil {
		return j, err
	}
	if j.Kind == "tutor" {
		var in map[string]string
		if err = json.Unmarshal(j.Input, &in); err != nil {
			return Job{}, err
		}
		_, _, err = s.tutorScope(ctx, tx, actor, org, in["sitting_id"], in["question_id"])
		return j, err
	}
	if err = Scope(ctx, tx, actor, org, j.Campus, j.Course, ""); err != nil {
		return Job{}, err
	}
	if j.Kind == "generate_questions" || j.Kind == "extract_document" {
		topicID := j.Resource
		if j.Kind == "extract_document" {
			var in map[string]string
			if err = json.Unmarshal(j.Input, &in); err != nil {
				return Job{}, err
			}
			topicID = in["topic_id"]
		}
		t, e := GetTopic(ctx, tx, org, topicID)
		if e != nil {
			return Job{}, e
		}
		if err = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err != nil {
			return Job{}, err
		}
	}
	return j, nil
}
func (s *Service) Jobs(ctx context.Context, tx pgx.Tx, actor, org, campus string) ([]Job, error) {
	type row struct{ ID string }
	ids, err := rows[row](ctx, tx, "SELECT id FROM academic_job WHERE institution_id=$1 AND ($2='' OR campus_id=$2) ORDER BY created_at DESC LIMIT 100", org, campus)
	if err != nil {
		return nil, err
	}
	out := []Job{}
	for _, v := range ids {
		j, e := s.Job(ctx, tx, actor, org, v.ID)
		if e == nil {
			out = append(out, j)
		} else if _, ok := e.(*Error); !ok {
			return nil, e
		}
	}
	return out, nil
}
func (s *Service) Generate(ctx context.Context, tx pgx.Tx, actor, org, topic string, count int, difficulty string) (Job, error) {
	t, err := GetTopic(ctx, tx, org, topic)
	if err != nil {
		return Job{}, err
	}
	if err = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err != nil {
		return Job{}, err
	}
	if count < 1 || count > 20 || (difficulty != "introductory" && difficulty != "standard" && difficulty != "challenging") {
		return Job{}, Problem(400, "invalid_generation", "Choose 1–20 questions and a supported difficulty.")
	}
	id := access.Random()
	_, err = tx.Exec(ctx, "INSERT INTO academic_job(id,institution_id,campus_id,course_id,actor_id,kind,resource_id,input) VALUES($1,$2,$3,$4,$5,'generate_questions',$6,$7)", id, org, t.CampusID, t.CourseID, actor, topic, Encode(Generation{t, count, difficulty}))
	if err != nil {
		return Job{}, err
	}
	if err = s.enqueue(ctx, tx, id); err != nil {
		return Job{}, err
	}
	if err = Event(ctx, tx, org, actor, "generation.requested", id, "Drafts require teacher review"); err != nil {
		return Job{}, err
	}
	return s.Job(ctx, tx, actor, org, id)
}
func (s *Service) ControlJob(ctx context.Context, tx pgx.Tx, actor, org, id, action string) error {
	if _, err := tx.Exec(ctx, "SELECT id FROM academic_job WHERE institution_id=$1 AND id=$2 FOR UPDATE", org, id); err != nil {
		return err
	}
	j, err := s.Job(ctx, tx, actor, org, id)
	if err != nil {
		return err
	}
	switch action {
	case "cancel":
		if j.State == "completed" {
			return Problem(409, "job_completed", "Completed output is retained. Retire its drafts instead.")
		}
		if j.RiverID != 0 {
			if _, err = s.Queue.JobCancelTx(ctx, tx, j.RiverID); err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, "UPDATE academic_job SET state='cancelled',error='Cancelled by staff',updated_at=now() WHERE id=$1", id)
	case "retry":
		if j.State != "failed" && j.State != "configuration_required" && j.State != "cancelled" {
			return Problem(409, "job_not_retryable", "Only failed, cancelled or configuration-blocked jobs can be retried.")
		}
		_, err = tx.Exec(ctx, "UPDATE academic_job SET state='queued',error='',updated_at=now() WHERE id=$1", id)
		if err == nil {
			if j.RiverID != 0 {
				_, err = s.Queue.JobRetryTx(ctx, tx, j.RiverID)
			} else {
				err = s.enqueue(ctx, tx, id)
			}
		}
	default:
		return Problem(400, "invalid_job_action", "Choose retry or cancel.")
	}
	if err != nil {
		return err
	}
	return Event(ctx, tx, org, actor, "academic_job."+action, id, "")
}
func (s *Service) jobAllowed(ctx context.Context, tx pgx.Tx, j Job) error {
	var state string
	if err := tx.QueryRow(ctx, "SELECT state FROM academic_job WHERE id=$1 FOR UPDATE", j.ID).Scan(&state); err != nil {
		return err
	}
	if state == "cancelled" {
		return river.JobCancel(errors.New("academic job cancelled"))
	}
	_, err := s.Job(ctx, tx, j.Actor, j.Org, j.ID)
	return err
}

type taskWorker struct {
	river.WorkerDefaults[TaskArgs]
	Service *Service
}

func (w *taskWorker) Timeout(*river.Job[TaskArgs]) time.Duration { return 5 * time.Minute }
func (w *taskWorker) Work(ctx context.Context, job *river.Job[TaskArgs]) error {
	var j Job
	s := w.Service
	err := Tx(ctx, s.DB, func(tx pgx.Tx) error {
		var org string
		if err := tx.QueryRow(ctx, "SELECT institution_id FROM academic_job WHERE id=$1", job.Args.ID).Scan(&org); err != nil {
			return err
		}
		var err error
		j, err = loadJob(ctx, tx, org, job.Args.ID)
		if err != nil {
			return err
		}
		if j.State == "completed" {
			return nil
		}
		if err = s.jobAllowed(ctx, tx, j); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, "UPDATE academic_job SET state='running',attempts=attempts+1,error='',updated_at=now() WHERE id=$1", j.ID)
		return err
	})
	if err != nil {
		if j.ID != "" {
			saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()
			_, saveErr := s.DB.Exec(saveCtx, "UPDATE academic_job SET state='failed',error=$2,updated_at=now() WHERE id=$1 AND state NOT IN ('completed','cancelled')", j.ID, err.Error())
			if saveErr != nil {
				return errors.Join(err, saveErr)
			}
		}
		return err
	}
	if j.State == "completed" {
		return nil
	}
	switch j.Kind {
	case "tutor":
		err = s.tutorJob(ctx, j)
	case "generate_questions":
		err = s.generateJob(ctx, j)
	case "extract_document":
		err = s.extract(ctx, j)
	case "recognize_scan":
		err = s.recognize(ctx, j)
	default:
		err = errors.New("unsupported academic job kind")
	}
	if err != nil {
		state := "failed"
		var known *Error
		if errors.As(err, &known) && known.Code == "provider_configuration_required" {
			state = "configuration_required"
		}
		saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		saveErr := Tx(saveCtx, s.DB, func(tx pgx.Tx) error {
			_, e := tx.Exec(saveCtx, "UPDATE academic_job SET state=$2,error=$3,updated_at=now() WHERE id=$1 AND state NOT IN ('completed','cancelled')", j.ID, state, err.Error())
			return e
		})
		if saveErr != nil {
			return errors.Join(err, saveErr)
		}
		if state == "configuration_required" {
			return river.JobCancel(err)
		}
		return err
	}
	return Tx(ctx, s.DB, func(tx pgx.Tx) error {
		if err := s.jobAllowed(ctx, tx, j); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, "UPDATE academic_job SET state='completed',progress=100,updated_at=now() WHERE id=$1", j.ID)
		return err
	})
}
func (s *Service) generateJob(ctx context.Context, j Job) error {
	var in Generation
	if err := json.Unmarshal(j.Input, &in); err != nil {
		return err
	}
	var generated Generated // Persist provider output before creating drafts; a retry after commit cannot call the provider again.
	if err := json.Unmarshal(j.Output, &generated); err != nil {
		return err
	}
	if len(generated.Questions) == 0 {
		var err error
		generated, err = s.Models.Generate(ctx, in)
		if err != nil {
			return err
		}
		if err = Tx(ctx, s.DB, func(tx pgx.Tx) error {
			if err := s.jobAllowed(ctx, tx, j); err != nil {
				return err
			}
			_, err := tx.Exec(ctx, "UPDATE academic_job SET output=$2,provider=$3,model=$4,input_tokens=$5,output_tokens=$6,progress=70 WHERE id=$1", j.ID, Encode(generated), generated.Provider, generated.Model, generated.InputTokens, generated.OutputTokens)
			return err
		}); err != nil {
			return err
		}
	}
	return Tx(ctx, s.DB, func(tx pgx.Tx) error {
		if err := s.jobAllowed(ctx, tx, j); err != nil {
			return err
		}
		origin := "generated"
		if generated.Provider == "development_fixture" {
			origin = "development_fixture"
		}
		for _, q := range generated.Questions {
			if _, _, err := s.CreateQuestion(ctx, tx, j.Actor, j.Org, q, origin); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx, "UPDATE academic_job SET state='completed',progress=100,updated_at=now() WHERE id=$1", j.ID)
		return err
	})
}

type expiryWorker struct {
	river.WorkerDefaults[ExpiryArgs]
	Service *Service
}

func (w *expiryWorker) Work(ctx context.Context, j *river.Job[ExpiryArgs]) error {
	return Tx(ctx, w.Service.DB, func(tx pgx.Tx) error {
		out, err := loadSitting(ctx, tx, j.Args.InstitutionID, j.Args.SittingID)
		if err != nil {
			return err
		}
		if out.State != "active" {
			return nil
		}
		if time.Now().Before(out.Deadline) {
			return river.JobSnooze(time.Until(out.Deadline))
		}
		return w.Service.finalize(ctx, tx, j.Args.InstitutionID, &out, "", "server_expiry")
	})
}
func (s *Service) WorkerClient() (*river.Client[pgx.Tx], error) {
	workers := river.NewWorkers()
	river.AddWorker(workers, &taskWorker{Service: s})
	river.AddWorker(workers, &expiryWorker{Service: s})
	return river.NewClient(riverpgxv5.New(s.DB), &river.Config{Queues: map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 4}}, Workers: workers})
}
