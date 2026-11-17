package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/access"
)

type Profile struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	EmergencyContact string `json:"emergency_contact"`
}

func (s *Server) profile(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		var p Profile
		err := tx.QueryRow(ctx, "SELECT name,email,phone,emergency_contact FROM person WHERE id=$1", a.ID).Scan(&p.Name, &p.Email, &p.Phone, &p.EmergencyContact)
		return p, err
	})
}
func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Name             string `json:"name"`
		Phone            string `json:"phone"`
		EmergencyContact string `json:"emergency_contact"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if err := requireText(in.Name, 1, 120); err != nil {
		return err
	}
	if len(in.Phone) > 40 || len(in.EmergencyContact) > 300 {
		return fail(400, "invalid_contact", "Check contact lengths.")
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		_, err := tx.Exec(ctx, "UPDATE person SET name=$2,phone=$3,emergency_contact=$4 WHERE id=$1", a.ID, strings.TrimSpace(in.Name), strings.TrimSpace(in.Phone), strings.TrimSpace(in.EmergencyContact))
		if err != nil {
			return nil, err
		}
		err = audit(ctx, tx, r, a, "", "profile.updated", a.ID, "")
		return map[string]string{"state": "saved"}, err
	})
}
func (s *Server) verifyPassword(ctx context.Context, tx pgx.Tx, a Actor, password, code string) error {
	var hash string
	var secret []byte
	var enabled bool
	var last int64
	err := tx.QueryRow(ctx, "SELECT password_hash,mfa_secret,mfa_enabled,last_totp_step FROM person WHERE id=$1 FOR UPDATE", a.ID).Scan(&hash, &secret, &enabled, &last)
	if err != nil {
		return err
	}
	if !access.CheckPassword(hash, password) {
		return fail(401, "reauthenticate", "Check your current password.")
	}
	if enabled {
		return s.checkFactor(ctx, tx, a.ID, secret, last, code)
	}
	return nil
}
func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		CurrentPassword string `json:"current_password"`
		Password        string `json:"password"`
		Code            string `json:"code"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	hash, err := access.Password(in.Password)
	if err != nil {
		return fail(400, "weak_password", err.Error())
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		if err := s.limit(ctx, "reauth:"+a.ID, 10); err != nil {
			return nil, err
		}
		if err := s.verifyPassword(ctx, tx, a, in.CurrentPassword, in.Code); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, "UPDATE person SET password_hash=$2 WHERE id=$1", a.ID, hash); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, "UPDATE session SET revoked_at=now() WHERE person_id=$1 AND id<>$2 AND revoked_at IS NULL", a.ID, a.SessionID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, "UPDATE security_token SET consumed_at=now() WHERE person_id=$1 AND consumed_at IS NULL", a.ID); err != nil {
			return nil, err
		}
		err := audit(ctx, tx, r, a, "", "credential.changed", a.ID, "Other sessions and reset links revoked")
		return map[string]string{"state": "changed"}, err
	})
}
func (s *Server) requestEmailChange(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Code     string `json:"code"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return err
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		if err := s.limit(ctx, "reauth:"+a.ID, 10); err != nil {
			return nil, err
		}
		if err := s.verifyPassword(ctx, tx, a, in.Password, in.Code); err != nil {
			return nil, err
		}
		token, id := access.Random(), access.Random()
		if _, err := tx.Exec(ctx, "INSERT INTO identity_change(id,person_id,new_email,token_hash,expires_at) VALUES($1,$2,$3,$4,now()+interval '1 hour')", id, a.ID, email, access.Digest(token)); err != nil {
			return nil, err
		}
		payload, _ := json.Marshal(map[string]string{"token": token, "origin": s.Config.Origin, "to": email})
		if _, err := tx.Exec(ctx, "INSERT INTO delivery_job(id,recipient_id,kind,payload) VALUES($1,$2,'email_change',$3)", access.Random(), a.ID, payload); err != nil {
			return nil, err
		}
		err := audit(ctx, tx, r, a, "", "email.change_requested", id, "")
		return map[string]string{"state": "queued"}, err
	})
}
func (s *Server) confirmEmailChange(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Token string `json:"token"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		var id, email string
		err := tx.QueryRow(ctx, "SELECT id,new_email FROM identity_change WHERE person_id=$1 AND token_hash=$2 AND consumed_at IS NULL AND expires_at>now() FOR UPDATE", a.ID, access.Digest(in.Token)).Scan(&id, &email)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fail(400, "invalid_token", "This link expired or has been used.")
		}
		if err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, "UPDATE person SET email=$2,verified_at=now() WHERE id=$1", a.ID, email); err != nil {
			return nil, conflict(err)
		}
		if _, err = tx.Exec(ctx, "UPDATE identity_change SET consumed_at=now() WHERE person_id=$1 AND consumed_at IS NULL", a.ID); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, "UPDATE session SET revoked_at=now() WHERE person_id=$1 AND id<>$2", a.ID, a.SessionID); err != nil {
			return nil, err
		}
		err = audit(ctx, tx, r, a, "", "email.changed", id, "")
		return map[string]string{"state": "changed"}, err
	})
}
func (s *Server) accountActivity(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		type Event struct {
			ID     int64     `json:"id"`
			Action string    `json:"action"`
			Reason string    `json:"reason"`
			Time   time.Time `json:"time"`
		}
		return collect[Event](ctx, tx, "SELECT id,action,reason,occurred_at FROM audit_entry WHERE actor_id=$1 AND institution_id IS NULL ORDER BY id DESC LIMIT 100", a.ID)
	})
}
