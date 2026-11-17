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

type InstitutionSettings struct {
	Name            string `json:"name"`
	Timezone        string `json:"timezone"`
	Locale          string `json:"locale"`
	State           string `json:"state"`
	SupportEmail    string `json:"support_email"`
	AcademicYear    string `json:"academic_year"`
	RequireStaffMFA bool   `json:"require_staff_mfa"`
	Version         int    `json:"version"`
}

func (s *Server) settings(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, "", "organization.manage"); err != nil {
			return nil, err
		}
		var out InstitutionSettings
		err := tx.QueryRow(ctx, "SELECT name,timezone,locale,state,support_email,academic_year,require_staff_mfa,version FROM institution WHERE id=$1", org).Scan(&out.Name, &out.Timezone, &out.Locale, &out.State, &out.SupportEmail, &out.AcademicYear, &out.RequireStaffMFA, &out.Version)
		return out, err
	})
}
func (s *Server) updateSettings(w http.ResponseWriter, r *http.Request) error {
	var in InstitutionSettings
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
	if in.State != "active" {
		return fail(400, "invalid_state", "Institution suspension and closure require the dedicated lifecycle workflow.")
	}
	if len(in.AcademicYear) > 40 {
		return fail(400, "invalid_year", "Use at most 40 characters.")
	}
	if in.SupportEmail != "" {
		if _, err := normalizeEmail(in.SupportEmail); err != nil {
			return err
		}
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, "", "organization.manage"); err != nil {
			return nil, err
		}
		if in.RequireStaffMFA && !a.MFA {
			return nil, fail(409, "mfa_required", "Enable your authenticator before enforcing staff MFA.")
		}
		tag, err := tx.Exec(ctx, "UPDATE institution SET name=$2,timezone=$3,locale=$4,support_email=$5,academic_year=$6,require_staff_mfa=$7,version=version+1 WHERE id=$1 AND version=$8", org, strings.TrimSpace(in.Name), in.Timezone, in.Locale, in.SupportEmail, in.AcademicYear, in.RequireStaffMFA, in.Version)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 0 {
			return nil, fail(409, "version_conflict", "Settings changed. Refresh and review them.")
		}
		err = audit(ctx, tx, r, a, org, "institution.configured", org, "")
		return map[string]string{"state": "saved"}, err
	})
}
func (s *Server) updateCampus(w http.ResponseWriter, r *http.Request) error {
	var in Campus
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if err := requireText(in.Name, 2, 120); err != nil {
		return err
	}
	if err := requireText(in.Code, 1, 20); err != nil {
		return err
	}
	if len(in.Address) > 500 || in.State != "active" && in.State != "closed" {
		return fail(400, "invalid_campus", "Check campus state and address.")
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org, id := r.PathValue("org"), r.PathValue("id")
		if err := authorize(ctx, tx, a, org, id, "campus.manage"); err != nil {
			return nil, err
		}
		if in.State == "closed" {
			var occupied bool
			if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM enrollment WHERE institution_id=$1 AND campus_id=$2 AND state='active')", org, id).Scan(&occupied); err != nil {
				return nil, err
			}
			if occupied {
				return nil, fail(409, "active_enrollments", "Transfer or withdraw active enrollments before closing the campus.")
			}
		}
		tag, err := tx.Exec(ctx, "UPDATE campus SET name=$3,code=$4,address=$5,state=$6,version=version+1 WHERE institution_id=$1 AND id=$2 AND version=$7", org, id, strings.TrimSpace(in.Name), strings.ToUpper(strings.TrimSpace(in.Code)), in.Address, in.State, in.Version)
		if err != nil {
			return nil, conflict(err)
		}
		if tag.RowsAffected() == 0 {
			return nil, fail(409, "version_conflict", "The campus changed or is unavailable. Refresh first.")
		}
		err = audit(ctx, tx, r, a, org, "campus.updated", id, in.State)
		return map[string]string{"state": "saved"}, err
	})
}
func (s *Server) transferOwner(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		PersonID string `json:"person_id"`
		Password string `json:"password"`
		Code     string `json:"code"`
		Reason   string `json:"reason"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if err := requireText(in.Reason, 10, 500); err != nil {
		return err
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org := r.PathValue("org")
		if err := authorize(ctx, tx, a, org, "", "access.manage"); err != nil {
			return nil, err
		}
		if in.PersonID == a.ID {
			return nil, fail(400, "invalid_owner", "Choose a different verified staff member.")
		}
		if err := s.limit(ctx, "reauth:"+a.ID, 10); err != nil {
			return nil, err
		}
		if err := s.verifyPassword(ctx, tx, a, in.Password, in.Code); err != nil {
			return nil, err
		}
		// Serialize ownership changes and re-check both identities before relinquishing access.
		var required bool
		if err := tx.QueryRow(ctx, "SELECT require_staff_mfa FROM institution WHERE id=$1 FOR UPDATE", org).Scan(&required); err != nil {
			return nil, err
		}
		var eligible bool
		err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM person p JOIN membership m ON m.person_id=p.id WHERE p.id=$1 AND p.state='active' AND p.verified_at IS NOT NULL AND (NOT $3 OR p.mfa_enabled) AND m.institution_id=$2 AND m.state='active' AND m.role IN ('coordinator','instructor','owner'))`, in.PersonID, org, required).Scan(&eligible)
		if err != nil {
			return nil, err
		}
		if !eligible {
			return nil, fail(400, "ineligible_owner", "The successor must be active verified staff and meet the institution MFA policy.")
		}
		if _, err = tx.Exec(ctx, `INSERT INTO membership(id,institution_id,person_id,role) VALUES($1,$2,$3,'owner') ON CONFLICT(institution_id,person_id,campus_id,role) DO UPDATE SET state='active',version=membership.version+1`, access.Random(), org, in.PersonID); err != nil {
			return nil, err
		}
		tag, err := tx.Exec(ctx, "UPDATE membership SET state='archived',version=version+1 WHERE institution_id=$1 AND person_id=$2 AND role='owner' AND state='active'", org, a.ID)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 0 {
			return nil, fail(409, "ownership_changed", "Ownership changed. Refresh before trying again.")
		}
		err = audit(ctx, tx, r, a, org, "owner.transferred", in.PersonID, in.Reason)
		return map[string]string{"state": "transferred"}, err
	})
}
func (s *Server) changeEnrollment(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		CampusID    string `json:"campus_id"`
		CourseID    string `json:"course_id"`
		State       string `json:"state"`
		EffectiveOn string `json:"effective_on"`
		Reason      string `json:"reason"`
		Version     int    `json:"version"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if err := requireText(in.Reason, 5, 500); err != nil {
		return err
	}
	if in.State != "active" && in.State != "withdrawn" {
		return fail(400, "invalid_state", "Select active or withdrawn.")
	}
	if _, err := time.Parse("2006-01-02", in.EffectiveOn); err != nil {
		return fail(400, "invalid_date", "Choose an effective date.")
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		org, id := r.PathValue("org"), r.PathValue("id")
		var person, oldCampus, oldCourse, oldState string
		var version int
		err := tx.QueryRow(ctx, "SELECT person_id,campus_id,course_id,state,version FROM enrollment WHERE institution_id=$1 AND id=$2 FOR UPDATE", org, id).Scan(&person, &oldCampus, &oldCourse, &oldState, &version)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fail(404, "not_found", "Enrollment unavailable.")
		}
		if err != nil {
			return nil, err
		}
		if err = authorize(ctx, tx, a, org, oldCampus, "people.manage"); err != nil {
			return nil, err
		}
		if err = authorize(ctx, tx, a, org, in.CampusID, "people.manage"); err != nil {
			return nil, err
		}
		if version != in.Version {
			return nil, fail(409, "version_conflict", "Placement changed. Refresh before trying again.")
		}
		var eligible bool
		err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM course c JOIN campus x ON x.id=c.campus_id WHERE c.institution_id=$1 AND c.id=$2 AND c.campus_id=$3 AND c.state='active' AND x.state='active')", org, in.CourseID, in.CampusID).Scan(&eligible)
		if err != nil {
			return nil, err
		}
		if !eligible {
			return nil, fail(400, "invalid_course", "Choose an active course at the target campus.")
		}
		var today string
		if err = tx.QueryRow(ctx, "SELECT (now() AT TIME ZONE timezone)::date::text FROM institution WHERE id=$1", org).Scan(&today); err != nil {
			return nil, err
		}
		if in.EffectiveOn > today {
			return nil, fail(400, "future_transfer", "Scheduled transfers are not supported yet. Choose today or an earlier effective date.")
		}
		if _, err = tx.Exec(ctx, "UPDATE enrollment SET campus_id=$3,course_id=$4,state=$5,end_date=CASE WHEN $5='withdrawn' THEN $6::date ELSE NULL END,version=version+1 WHERE institution_id=$1 AND id=$2", org, id, in.CampusID, in.CourseID, in.State, in.EffectiveOn); err != nil {
			return nil, conflict(err)
		}
		if _, err = tx.Exec(ctx, "INSERT INTO placement_event(id,enrollment_id,institution_id,previous_campus_id,previous_course_id,previous_state,next_campus_id,next_course_id,next_state,effective_on,actor_id,reason) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)", access.Random(), id, org, oldCampus, oldCourse, oldState, in.CampusID, in.CourseID, in.State, in.EffectiveOn, a.ID, in.Reason); err != nil {
			return nil, err
		}
		if in.State == "active" {
			if _, err = tx.Exec(ctx, "INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,'learner') ON CONFLICT(institution_id,person_id,campus_id,role) DO UPDATE SET state='active',version=membership.version+1", access.Random(), org, person, in.CampusID); err != nil {
				return nil, err
			}
		}
		if oldCampus != in.CampusID || in.State == "withdrawn" {
			if _, err = tx.Exec(ctx, `UPDATE membership SET state='archived',version=version+1 WHERE institution_id=$1 AND person_id=$2 AND campus_id=$3 AND role='learner' AND NOT EXISTS(SELECT 1 FROM enrollment WHERE institution_id=$1 AND person_id=$2 AND campus_id=$3 AND state='active')`, org, person, oldCampus); err != nil {
				return nil, err
			}
		}
		if _, err = tx.Exec(ctx, "UPDATE session SET revoked_at=now() WHERE person_id=$1 AND revoked_at IS NULL", person); err != nil {
			return nil, err
		}
		err = audit(ctx, tx, r, a, org, "placement.changed", id, in.Reason)
		return map[string]string{"state": "saved"}, err
	})
}
