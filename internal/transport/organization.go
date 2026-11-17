package transport

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/access"
)

type Workspace struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
	Locale   string `json:"locale"`
	Role     string `json:"role"`
	CampusID string `json:"campus_id"`
}
type Campus struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Address string `json:"address"`
	State   string `json:"state"`
	Version int    `json:"version"`
}
type Person struct {
	MembershipID string `json:"membership_id"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	CampusID     string `json:"campus_id"`
	State        string `json:"state"`
	AccessState  string `json:"access_state"`
	Version      int    `json:"version"`
}
type Course struct {
	ID       string `json:"id"`
	CampusID string `json:"campus_id"`
	Name     string `json:"name"`
	Year     string `json:"academic_year"`
}

func collect[T any](ctx context.Context, tx pgx.Tx, sql string, args ...any) ([]T, error) {
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	out, err := pgx.CollectRows(rows, pgx.RowToStructByPos[T])
	if out == nil {
		out = []T{}
	}
	return out, err
}
func (s *Server) workspaces(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		return collect[Workspace](ctx, tx, `SELECT i.id,i.name,i.timezone,i.locale,m.role,coalesce(m.campus_id,'') FROM institution i JOIN membership m ON m.institution_id=i.id WHERE m.person_id=$1 AND m.state='active' AND i.state='active' ORDER BY i.name,m.role LIMIT 100`, a.ID)
	})
}
func (s *Server) createInstitution(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Name     string `json:"name"`
		Timezone string `json:"timezone"`
		Locale   string `json:"locale"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if err := requireText(in.Name, 2, 120); err != nil {
		return err
	}
	if _, err := time.LoadLocation(in.Timezone); err != nil {
		return fail(400, "invalid_timezone", "Choose an IANA timezone.")
	}
	if in.Locale != "en" && in.Locale != "hi" {
		return fail(400, "invalid_locale", "Choose English or Hindi.")
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		var allowed bool
		err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM membership WHERE person_id=$1 AND role='owner' AND state='active')", a.ID).Scan(&allowed)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, fail(403, "forbidden", "An institution owner must create additional institutions.")
		}
		id := access.Random()
		if _, err = tx.Exec(ctx, "INSERT INTO institution(id,name,timezone,locale) VALUES($1,$2,$3,$4)", id, strings.TrimSpace(in.Name), in.Timezone, in.Locale); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO membership(id,institution_id,person_id,role) VALUES($1,$2,$3,'owner')", access.Random(), id, a.ID); err != nil {
			return nil, err
		}
		err = audit(ctx, tx, r, a, id, "institution.created", id, "")
		return map[string]string{"id": id}, err
	})
}
func (s *Server) campuses(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		return collect[Campus](ctx, tx, `SELECT c.id,c.name,c.code,c.address,c.state,c.version FROM campus c JOIN institution i ON i.id=c.institution_id WHERE c.institution_id=$1 AND i.state='active' AND EXISTS(SELECT 1 FROM membership m WHERE m.institution_id=c.institution_id AND m.person_id=$2 AND m.state='active' AND (m.campus_id IS NULL OR m.campus_id=c.id)) ORDER BY c.name LIMIT 200`, r.PathValue("org"), a.ID)
	})
}
func (s *Server) createCampus(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Name    string `json:"name"`
		Code    string `json:"code"`
		Address string `json:"address"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if err := requireText(in.Name, 2, 120); err != nil {
		return err
	}
	if err := requireText(in.Code, 1, 20); err != nil {
		return err
	}
	if len(in.Address) > 500 {
		return fail(400, "invalid_address", "Use at most 500 characters.")
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, "", "campus.manage"); err != nil {
			return nil, err
		}
		id := access.Random()
		_, err := tx.Exec(ctx, "INSERT INTO campus(id,institution_id,name,code,address) VALUES($1,$2,$3,$4,$5)", id, org, strings.TrimSpace(in.Name), strings.ToUpper(strings.TrimSpace(in.Code)), in.Address)
		if err != nil {
			return nil, conflict(err)
		}
		err = audit(ctx, tx, r, a, org, "campus.created", id, "")
		return map[string]string{"id": id}, err
	})
}
func (s *Server) people(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org, campus := r.PathValue("org"), r.URL.Query().Get("campus")
		if err := authorize(ctx, tx, a, org, campus, "people.read"); err != nil {
			return nil, err
		}
		return collect[Person](ctx, tx, `SELECT m.id,p.id,p.name,p.email,m.role,coalesce(m.campus_id,''),m.state,p.state,m.version FROM membership m JOIN person p ON p.id=m.person_id WHERE m.institution_id=$1 AND ($2='' OR m.campus_id=$2) AND (p.name ILIKE '%'||$3||'%' OR p.email ILIKE '%'||$3||'%') ORDER BY p.name,m.id LIMIT 200`, org, campus, truncate(r.URL.Query().Get("q"), 100))
	})
}
func (s *Server) invite(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		CampusID string `json:"campus_id"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return err
	}
	if err = requireText(in.Name, 1, 120); err != nil {
		return err
	}
	if !strings.Contains("|coordinator|instructor|learner|guardian|reviewer|", "|"+in.Role+"|") || in.Role == "" {
		return fail(400, "invalid_role", "Select a supported role. Owner transfer requires its protected workflow.")
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, in.CampusID, "people.manage"); err != nil {
			return nil, err
		}
		if in.Role == "coordinator" {
			if err := authorize(ctx, tx, a, org, "", "access.manage"); err != nil {
				return nil, err
			}
		}
		if in.CampusID == "" {
			return nil, fail(400, "campus_required", "Assign a campus responsibility.")
		}
		var valid bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM campus WHERE id=$1 AND institution_id=$2 AND state='active')", in.CampusID, org).Scan(&valid); err != nil {
			return nil, err
		}
		if !valid {
			return nil, fail(400, "invalid_campus", "Select an active campus in this institution.")
		}
		id := access.Random()
		var state string
		err := tx.QueryRow(ctx, "SELECT id,state FROM person WHERE email=$1 FOR UPDATE", email).Scan(&id, &state)
		fresh := errors.Is(err, pgx.ErrNoRows)
		if err != nil && !fresh {
			return nil, err
		}
		if !fresh && (state == "suspended" || state == "archived") {
			return nil, fail(409, "account_unavailable", "This person requires an account recovery review.")
		}
		if fresh {
			hash, err := access.Password(access.Random())
			if err != nil {
				return nil, err
			}
			if _, err = tx.Exec(ctx, "INSERT INTO person(id,email,name,password_hash,state) VALUES($1,$2,$3,$4,'invited')", id, email, strings.TrimSpace(in.Name), hash); err != nil {
				return nil, conflict(err)
			}
		}
		membership := access.Random()
		_, err = tx.Exec(ctx, "INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,$5)", membership, org, id, in.CampusID, in.Role)
		if err != nil {
			return nil, conflict(err)
		}
		if fresh {
			if _, err = queueToken(ctx, tx, id, org, "invitation", s.Config.Origin); err != nil {
				return nil, err
			}
		}
		if err = audit(ctx, tx, r, a, org, "membership.invited", membership, in.Role); err != nil {
			return nil, err
		}
		return map[string]string{"id": membership, "state": "assigned", "delivery": map[bool]string{true: "queued", false: "existing_account"}[fresh]}, nil
	})
}
func (s *Server) changeMembership(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		State   string `json:"state"`
		Reason  string `json:"reason"`
		Version int    `json:"version"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if in.State != "active" && in.State != "suspended" && in.State != "archived" {
		return fail(400, "invalid_state", "Choose active, suspended or archived.")
	}
	if err := requireText(in.Reason, 5, 500); err != nil {
		return err
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, "", "access.manage"); err != nil {
			return nil, err
		}
		var person, role string
		var version int
		err := tx.QueryRow(ctx, "SELECT person_id,role,version FROM membership WHERE institution_id=$1 AND id=$2 FOR UPDATE", org, r.PathValue("id")).Scan(&person, &role, &version)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fail(404, "not_found", "Membership unavailable.")
		}
		if err != nil {
			return nil, err
		}
		if version != in.Version {
			return nil, fail(409, "version_conflict", "This membership changed. Refresh before trying again.")
		}
		if person == a.ID || role == "owner" {
			return nil, fail(403, "protected_owner", "Self changes and owner removal require the protected transfer workflow.")
		}
		if _, err = tx.Exec(ctx, "UPDATE membership SET state=$3,version=version+1 WHERE institution_id=$1 AND id=$2", org, r.PathValue("id"), in.State); err != nil {
			return nil, err
		}
		// Revoke globally so a suspended membership cannot retain an old authenticated context.
		if in.State != "active" {
			if _, err = tx.Exec(ctx, "UPDATE session SET revoked_at=now() WHERE person_id=$1 AND revoked_at IS NULL", person); err != nil {
				return nil, err
			}
		}
		err = audit(ctx, tx, r, a, org, "membership."+in.State, r.PathValue("id"), in.Reason)
		return map[string]string{"state": in.State}, err
	})
}
func (s *Server) auditEntries(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, "", "audit.read"); err != nil {
			return nil, err
		}
		type Entry struct {
			ID         int64     `json:"id"`
			Actor      string    `json:"actor"`
			Action     string    `json:"action"`
			ResourceID string    `json:"resource_id"`
			Reason     string    `json:"reason"`
			RequestID  string    `json:"request_id"`
			Time       time.Time `json:"time"`
		}
		return collect[Entry](ctx, tx, `SELECT e.id,coalesce(p.name,'Operator'),e.action,e.resource_id,e.reason,e.request_id,e.occurred_at FROM audit_entry e LEFT JOIN person p ON p.id=e.actor_id WHERE e.institution_id=$1 ORDER BY e.id DESC LIMIT 200`, org)
	})
}
func (s *Server) courses(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org, campus := r.PathValue("org"), r.URL.Query().Get("campus")
		if err := authorize(ctx, tx, a, org, campus, "people.read"); err != nil {
			return nil, err
		}
		return collect[Course](ctx, tx, "SELECT id,campus_id,name,academic_year FROM course WHERE institution_id=$1 AND ($2='' OR campus_id=$2) ORDER BY name LIMIT 200", org, campus)
	})
}
func (s *Server) createCourse(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Name     string `json:"name"`
		CampusID string `json:"campus_id"`
		Year     string `json:"academic_year"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if err := requireText(in.Name, 2, 120); err != nil {
		return err
	}
	if err := requireText(in.Year, 4, 20); err != nil {
		return err
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, in.CampusID, "people.manage"); err != nil {
			return nil, err
		}
		id := access.Random()
		_, err := tx.Exec(ctx, "INSERT INTO course(id,institution_id,campus_id,name,academic_year) VALUES($1,$2,$3,$4,$5)", id, org, in.CampusID, in.Name, in.Year)
		if err != nil {
			return nil, conflict(err)
		}
		err = audit(ctx, tx, r, a, org, "course.created", id, "")
		return map[string]string{"id": id}, err
	})
}
