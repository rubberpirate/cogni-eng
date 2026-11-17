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

type AdmissionCode struct {
	ID        string    `json:"id"`
	CampusID  string    `json:"campus_id"`
	Label     string    `json:"label"`
	ExpiresAt time.Time `json:"expires_at"`
	MaxUses   int       `json:"max_uses"`
	Used      int       `json:"used"`
	Revoked   bool      `json:"revoked"`
}

func (s *Server) admissionCodes(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org, campus := r.PathValue("org"), r.URL.Query().Get("campus")
		if err := authorize(ctx, tx, a, org, campus, "people.manage"); err != nil {
			return nil, err
		}
		return collect[AdmissionCode](ctx, tx, "SELECT id,campus_id,label,expires_at,max_uses,used,revoked_at IS NOT NULL FROM admission_code WHERE institution_id=$1 AND ($2='' OR campus_id=$2) ORDER BY created_at DESC LIMIT 100", org, campus)
	})
}
func (s *Server) createAdmissionCode(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		CampusID  string `json:"campus_id"`
		Label     string `json:"label"`
		ExpiresAt string `json:"expires_at"`
		MaxUses   int    `json:"max_uses"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	expiry, err := time.Parse(time.RFC3339, in.ExpiresAt)
	if err != nil || expiry.Before(time.Now()) || expiry.After(time.Now().Add(31*24*time.Hour)) {
		return fail(400, "invalid_expiry", "Choose an expiry within 31 days.")
	}
	if in.MaxUses < 1 || in.MaxUses > 1000 {
		return fail(400, "invalid_limit", "Choose 1–1000 uses.")
	}
	if err := requireText(in.Label, 2, 100); err != nil {
		return err
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, in.CampusID, "people.manage"); err != nil {
			return nil, err
		}
		id, code := access.Random(), access.Random()[:24]
		_, err := tx.Exec(ctx, "INSERT INTO admission_code(id,institution_id,campus_id,code_hash,label,expires_at,max_uses,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8)", id, org, in.CampusID, access.Digest(code), in.Label, expiry, in.MaxUses, a.ID)
		if err != nil {
			return nil, conflict(err)
		}
		err = audit(ctx, tx, r, a, org, "admission_code.created", id, in.Label)
		return map[string]string{"id": id, "code": code}, err
	})
}
func (s *Server) revokeAdmissionCode(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		var campus string
		err := tx.QueryRow(ctx, "SELECT campus_id FROM admission_code WHERE institution_id=$1 AND id=$2", org, r.PathValue("id")).Scan(&campus)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fail(404, "not_found", "Code unavailable.")
		}
		if err != nil {
			return nil, err
		}
		if err = authorize(ctx, tx, a, org, campus, "people.manage"); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, "UPDATE admission_code SET revoked_at=now() WHERE id=$1", r.PathValue("id")); err != nil {
			return nil, err
		}
		err = audit(ctx, tx, r, a, org, "admission_code.revoked", r.PathValue("id"), "")
		return map[string]string{"state": "revoked"}, err
	})
}
func (s *Server) register(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Name           string `json:"name"`
		Email          string `json:"email"`
		Password       string `json:"password"`
		EnrollmentCode string `json:"enrollment_code"`
		AcceptPolicy   bool   `json:"accept_policy"`
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
	if !in.AcceptPolicy {
		return fail(400, "policy_required", "Accept the current terms and privacy policy.")
	}
	if err = s.limit(r.Context(), "registration:"+email, 5); err != nil {
		return err
	}
	if err = s.limit(r.Context(), "admission-code:"+in.EnrollmentCode, 1000); err != nil {
		return err
	}
	hash, err := access.Password(in.Password)
	if err != nil {
		return fail(400, "weak_password", err.Error())
	}
	err = s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		var codeID, org, campus string
		err := tx.QueryRow(ctx, `SELECT c.id,c.institution_id,c.campus_id FROM admission_code c JOIN institution i ON i.id=c.institution_id JOIN campus x ON x.id=c.campus_id WHERE c.code_hash=$1 AND c.revoked_at IS NULL AND c.expires_at>now() AND c.used<c.max_uses AND i.state='active' AND x.state='active' FOR UPDATE OF c`, access.Digest(strings.TrimSpace(in.EnrollmentCode))).Scan(&codeID, &org, &campus)
		if errors.Is(err, pgx.ErrNoRows) {
			return fail(400, "invalid_enrollment_code", "This enrollment code is unavailable. Ask your institution for a current code.")
		}
		if err != nil {
			return err
		}
		var exists bool
		if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM person WHERE email=$1)", email).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return nil
		}
		id := access.Random()
		if _, err = tx.Exec(ctx, "INSERT INTO person(id,email,name,password_hash,state) VALUES($1,$2,$3,$4,'invited')", id, email, strings.TrimSpace(in.Name), hash); err != nil {
			return conflict(err)
		}
		if _, err = tx.Exec(ctx, "INSERT INTO admission_request(id,institution_id,campus_id,person_id,code_id) VALUES($1,$2,$3,$4,$5)", access.Random(), org, campus, id, codeID); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "UPDATE admission_code SET used=used+1 WHERE id=$1", codeID); err != nil {
			return err
		}
		for _, policy := range []string{"terms", "privacy"} {
			if _, err = tx.Exec(ctx, "INSERT INTO policy_acknowledgment(id,person_id,policy,version) VALUES($1,$2,$3,'v1')", access.Random(), id, policy); err != nil {
				return err
			}
		}
		if _, err = queueToken(ctx, tx, id, org, "verify", s.Config.Origin); err != nil {
			return err
		}
		return audit(ctx, tx, r, Actor{ID: id}, org, "admission.registered", id, "Awaiting verified email and staff placement")
	})
	if err != nil {
		return err
	}
	write(w, 200, map[string]string{"state": "requested", "message": "If eligible, verification instructions are queued. Existing users should sign in and contact the institution."})
	return nil
}
func (s *Server) verifyEmail(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Token string `json:"token"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if err := s.limit(r.Context(), "verify:"+in.Token, 8); err != nil {
		return err
	}
	err := s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		var person, id string
		err := tx.QueryRow(ctx, "SELECT person_id,id FROM security_token WHERE token_hash=$1 AND kind='verify' AND consumed_at IS NULL AND expires_at>now() FOR UPDATE", access.Digest(in.Token)).Scan(&person, &id)
		if errors.Is(err, pgx.ErrNoRows) {
			return fail(400, "invalid_token", "This link expired or has already been used.")
		}
		if err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, "UPDATE person SET state='active',verified_at=now() WHERE id=$1 AND state='invited'", person)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return fail(409, "account_unavailable", "Contact your institution about this account.")
		}
		if _, err = tx.Exec(ctx, "UPDATE security_token SET consumed_at=now() WHERE id=$1", id); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "UPDATE admission_request SET state='pending',version=version+1 WHERE person_id=$1 AND state='unverified'", person); err != nil {
			return err
		}
		return audit(ctx, tx, r, Actor{ID: person}, "", "email.verified", person, "")
	})
	if err != nil {
		return err
	}
	write(w, 200, map[string]string{"state": "verified"})
	return nil
}

type Admission struct {
	ID       string `json:"id"`
	PersonID string `json:"person_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	CampusID string `json:"campus_id"`
	State    string `json:"state"`
	Reason   string `json:"reason"`
	Version  int    `json:"version"`
}

func (s *Server) admissionQueue(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org, campus := r.PathValue("org"), r.URL.Query().Get("campus")
		if err := authorize(ctx, tx, a, org, campus, "people.manage"); err != nil {
			return nil, err
		}
		return collect[Admission](ctx, tx, "SELECT q.id,p.id,p.name,p.email,q.campus_id,q.state,q.reason,q.version FROM admission_request q JOIN person p ON p.id=q.person_id WHERE q.institution_id=$1 AND ($2='' OR q.campus_id=$2) ORDER BY q.created_at DESC LIMIT 200", org, campus)
	})
}
func (s *Server) placeAdmission(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		CourseID  string `json:"course_id"`
		StartDate string `json:"start_date"`
		Decision  string `json:"decision"`
		Reason    string `json:"reason"`
		Version   int    `json:"version"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if in.Decision != "placed" && in.Decision != "rejected" {
		return fail(400, "invalid_decision", "Choose place or reject.")
	}
	if err := requireText(in.Reason, 5, 500); err != nil {
		return err
	}
	if in.Decision == "placed" {
		if _, err := time.Parse("2006-01-02", in.StartDate); err != nil {
			return fail(400, "invalid_date", "Choose a valid start date.")
		}
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org, id := r.PathValue("org"), r.PathValue("id")
		var person, campus, state string
		var version int
		err := tx.QueryRow(ctx, "SELECT person_id,campus_id,state,version FROM admission_request WHERE institution_id=$1 AND id=$2 FOR UPDATE", org, id).Scan(&person, &campus, &state, &version)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fail(404, "not_found", "Admission unavailable.")
		}
		if err != nil {
			return nil, err
		}
		if err = authorize(ctx, tx, a, org, campus, "people.manage"); err != nil {
			return nil, err
		}
		if state != "pending" || version != in.Version {
			return nil, fail(409, "admission_changed", "Only a verified pending admission can be decided. Refresh first.")
		}
		if in.Decision == "placed" {
			var valid bool
			if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM course WHERE institution_id=$1 AND campus_id=$2 AND id=$3 AND state='active')", org, campus, in.CourseID).Scan(&valid); err != nil {
				return nil, err
			}
			if !valid {
				return nil, fail(400, "invalid_course", "Choose an active course at the admission campus.")
			}
			if _, err = tx.Exec(ctx, "INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,'learner')", access.Random(), org, person, campus); err != nil {
				return nil, conflict(err)
			}
			if _, err = tx.Exec(ctx, "INSERT INTO enrollment(id,institution_id,person_id,campus_id,course_id,state,start_date) VALUES($1,$2,$3,$4,$5,'active',$6)", access.Random(), org, person, campus, in.CourseID, in.StartDate); err != nil {
				return nil, conflict(err)
			}
		}
		if _, err = tx.Exec(ctx, "UPDATE admission_request SET state=$2,reason=$3,version=version+1 WHERE id=$1", id, in.Decision, in.Reason); err != nil {
			return nil, err
		}
		err = audit(ctx, tx, r, a, org, "admission."+in.Decision, id, in.Reason)
		return map[string]string{"state": in.Decision}, err
	})
}

func (s *Server) resendVerification(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Email string `json:"email"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return err
	}
	if err = s.limit(r.Context(), "resend-verify:"+email, 3); err != nil {
		return err
	}
	err = s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		var person, org string
		err := tx.QueryRow(ctx, `SELECT p.id,q.institution_id FROM person p JOIN admission_request q ON q.person_id=p.id WHERE p.email=$1 AND p.state='invited' AND q.state='unverified' FOR UPDATE OF p`, email).Scan(&person, &org)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "UPDATE security_token SET consumed_at=now() WHERE person_id=$1 AND kind='verify' AND consumed_at IS NULL", person); err != nil {
			return err
		}
		_, err = queueToken(ctx, tx, person, org, "verify", s.Config.Origin)
		return err
	})
	if err != nil {
		return err
	}
	write(w, 200, map[string]string{"state": "requested"})
	return nil
}
