package transport

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"learning.local/workspace/internal/access"
)

func conflict(err error) error {
	var p *pgconn.PgError
	if errors.As(err, &p) {
		switch p.Code {
		case "23505":
			return fail(409, "duplicate", "This record already exists. Refresh and review it.")
		case "23503", "23514", "22007":
			return fail(400, "invalid_relationship", "Check the selected records, dates and institution.")
		}
	}
	return err
}

type Enrollment struct {
	ID        string `json:"id"`
	PersonID  string `json:"person_id"`
	Name      string `json:"name"`
	CampusID  string `json:"campus_id"`
	CourseID  string `json:"course_id"`
	Course    string `json:"course"`
	Reference string `json:"reference"`
	State     string `json:"state"`
	Version   int    `json:"version"`
}

func (s *Server) enrollments(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org, campus := r.PathValue("org"), r.URL.Query().Get("campus")
		if err := authorize(ctx, tx, a, org, campus, "people.read"); err != nil {
			return nil, err
		}
		return collect[Enrollment](ctx, tx, `SELECT e.id,e.person_id,p.name,e.campus_id,e.course_id,c.name,e.reference,e.state,e.version FROM enrollment e JOIN person p ON p.id=e.person_id JOIN course c ON c.id=e.course_id WHERE e.institution_id=$1 AND ($2='' OR e.campus_id=$2) ORDER BY e.reference LIMIT 200`, org, campus)
	})
}
func (s *Server) createEnrollment(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		PersonID  string `json:"person_id"`
		CampusID  string `json:"campus_id"`
		CourseID  string `json:"course_id"`
		StartDate string `json:"start_date"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if _, err := time.Parse("2006-01-02", in.StartDate); err != nil {
		return fail(400, "invalid_date", "Use a valid start date.")
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, in.CampusID, "people.manage"); err != nil {
			return nil, err
		}
		var eligible bool
		err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM membership m JOIN course c ON c.institution_id=m.institution_id AND c.campus_id=m.campus_id WHERE m.institution_id=$1 AND m.person_id=$2 AND m.campus_id=$3 AND m.role='learner' AND m.state='active' AND c.id=$4 AND c.state='active')`, org, in.PersonID, in.CampusID, in.CourseID).Scan(&eligible)
		if err != nil {
			return nil, err
		}
		if !eligible {
			return nil, fail(400, "ineligible", "Select a learner and course at the same authorized campus.")
		}
		id := access.Random()
		var reference string
		err = tx.QueryRow(ctx, "INSERT INTO enrollment(id,institution_id,person_id,campus_id,course_id,state,start_date) VALUES($1,$2,$3,$4,$5,'active',$6) RETURNING reference", id, org, in.PersonID, in.CampusID, in.CourseID, in.StartDate).Scan(&reference)
		if err != nil {
			return nil, conflict(err)
		}
		err = audit(ctx, tx, r, a, org, "learner.enrolled", id, reference)
		return map[string]string{"id": id, "reference": reference}, err
	})
}

type Family struct {
	ID           string `json:"id"`
	GuardianID   string `json:"guardian_id"`
	Guardian     string `json:"guardian"`
	LearnerID    string `json:"learner_id"`
	Learner      string `json:"learner"`
	Relationship string `json:"relationship"`
	State        string `json:"state"`
}

func (s *Server) families(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org, campus := r.PathValue("org"), r.URL.Query().Get("campus")
		if err := authorize(ctx, tx, a, org, campus, "people.read"); err != nil {
			return nil, err
		}
		return collect[Family](ctx, tx, `SELECT f.id,g.id,g.name,l.id,l.name,f.relationship,f.state FROM family_link f JOIN person g ON g.id=f.guardian_id JOIN person l ON l.id=f.learner_id WHERE f.institution_id=$1 AND ($2='' OR EXISTS(SELECT 1 FROM membership m WHERE m.institution_id=$1 AND m.person_id=l.id AND m.campus_id=$2 AND m.role='learner' AND m.state='active')) ORDER BY l.name LIMIT 200`, org, campus)
	})
}
func (s *Server) createFamily(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		GuardianID   string `json:"guardian_id"`
		LearnerID    string `json:"learner_id"`
		CampusID     string `json:"campus_id"`
		Relationship string `json:"relationship"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if err := requireText(in.Relationship, 2, 80); err != nil {
		return err
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, in.CampusID, "people.manage"); err != nil {
			return nil, err
		}
		var valid bool
		err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM membership g JOIN membership l ON l.institution_id=g.institution_id WHERE g.institution_id=$1 AND g.person_id=$2 AND g.role='guardian' AND g.state='active' AND l.person_id=$3 AND l.role='learner' AND l.campus_id=$4 AND l.state='active')`, org, in.GuardianID, in.LearnerID, in.CampusID).Scan(&valid)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, fail(400, "invalid_relationship", "Select a guardian and learner in this institution and campus.")
		}
		id := access.Random()
		err = tx.QueryRow(ctx, "INSERT INTO family_link(id,institution_id,guardian_id,learner_id,relationship,state,approved_by) VALUES($1,$2,$3,$4,$5,'approved',$6) ON CONFLICT(institution_id,guardian_id,learner_id) DO UPDATE SET state='approved',relationship=excluded.relationship,approved_by=excluded.approved_by WHERE family_link.state='removed' RETURNING id", id, org, in.GuardianID, in.LearnerID, in.Relationship, a.ID).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fail(409, "duplicate", "This family link is already approved.")
		}
		if err != nil {
			return nil, conflict(err)
		}
		err = audit(ctx, tx, r, a, org, "family.approved", id, in.Relationship)
		return map[string]string{"id": id}, err
	})
}
func (s *Server) removeFamily(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, "", "people.manage"); err != nil {
			return nil, err
		}
		tag, err := tx.Exec(ctx, "UPDATE family_link SET state='removed' WHERE id=$1 AND institution_id=$2 AND state='approved'", r.PathValue("id"), org)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 0 {
			return nil, fail(404, "not_found", "Link unavailable.")
		}
		err = audit(ctx, tx, r, a, org, "family.removed", r.PathValue("id"), "Institution request")
		return map[string]string{"state": "removed"}, err
	})
}

type Responsibility struct {
	ID       string `json:"id"`
	StaffID  string `json:"staff_id"`
	Staff    string `json:"staff"`
	CourseID string `json:"course_id"`
	Course   string `json:"course"`
	Subject  string `json:"subject"`
	StartsOn string `json:"starts_on"`
	EndsOn   string `json:"ends_on"`
}

func (s *Server) responsibilities(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org, campus := r.PathValue("org"), r.URL.Query().Get("campus")
		if err := authorize(ctx, tx, a, org, campus, "people.read"); err != nil {
			return nil, err
		}
		return collect[Responsibility](ctx, tx, `SELECT x.id,p.id,p.name,c.id,c.name,x.subject,x.starts_on::text,coalesce(x.ends_on::text,'') FROM responsibility x JOIN person p ON p.id=x.staff_id JOIN course c ON c.id=x.course_id WHERE x.institution_id=$1 AND ($2='' OR c.campus_id=$2) ORDER BY p.name LIMIT 200`, org, campus)
	})
}
func (s *Server) createResponsibility(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		StaffID  string `json:"staff_id"`
		CourseID string `json:"course_id"`
		CampusID string `json:"campus_id"`
		Subject  string `json:"subject"`
		StartsOn string `json:"starts_on"`
		EndsOn   string `json:"ends_on"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if err := requireText(in.Subject, 1, 100); err != nil {
		return err
	}
	if _, err := time.Parse("2006-01-02", in.StartsOn); err != nil {
		return fail(400, "invalid_date", "Choose a start date.")
	}
	var end any
	if in.EndsOn != "" {
		if _, err := time.Parse("2006-01-02", in.EndsOn); err != nil || in.EndsOn < in.StartsOn {
			return fail(400, "invalid_date", "End date must follow start date.")
		}
		end = in.EndsOn
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, in.CampusID, "people.manage"); err != nil {
			return nil, err
		}
		var valid bool
		err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM membership m JOIN course c ON c.institution_id=m.institution_id WHERE m.institution_id=$1 AND m.person_id=$2 AND m.role='instructor' AND m.state='active' AND (m.campus_id=c.campus_id OR m.campus_id IS NULL) AND c.id=$3 AND c.campus_id=$4)`, org, in.StaffID, in.CourseID, in.CampusID).Scan(&valid)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, fail(400, "invalid_assignment", "Choose an instructor and course at the authorized campus.")
		}
		id := access.Random()
		_, err = tx.Exec(ctx, "INSERT INTO responsibility(id,institution_id,staff_id,course_id,subject,starts_on,ends_on) VALUES($1,$2,$3,$4,$5,$6,$7)", id, org, in.StaffID, in.CourseID, in.Subject, in.StartsOn, end)
		if err != nil {
			return nil, conflict(err)
		}
		err = audit(ctx, tx, r, a, org, "responsibility.created", id, in.Subject)
		return map[string]string{"id": id}, err
	})
}
func (s *Server) deliveries(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, "", "access.manage"); err != nil {
			return nil, err
		}
		type Delivery struct {
			ID        string    `json:"id"`
			Recipient string    `json:"recipient"`
			Kind      string    `json:"kind"`
			State     string    `json:"state"`
			Attempts  int       `json:"attempts"`
			Error     string    `json:"error"`
			CreatedAt time.Time `json:"created_at"`
		}
		return collect[Delivery](ctx, tx, `SELECT j.id,p.email,j.kind,j.state,j.attempts,j.last_error,j.created_at FROM delivery_job j JOIN person p ON p.id=j.recipient_id WHERE j.institution_id=$1 ORDER BY j.created_at DESC LIMIT 100`, org)
	})
}
func (s *Server) retryDelivery(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, "", "access.manage"); err != nil {
			return nil, err
		}
		tag, err := tx.Exec(ctx, "UPDATE delivery_job SET state='queued',run_at=now(),attempts=0,last_error='' WHERE id=$1 AND institution_id=$2 AND state='failed'", r.PathValue("id"), org)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 0 {
			return nil, fail(409, "invalid_state", "Only failed delivery can be retried.")
		}
		err = audit(ctx, tx, r, a, org, "delivery.retried", r.PathValue("id"), "Staff recovery")
		return map[string]string{"state": "queued"}, err
	})
}
