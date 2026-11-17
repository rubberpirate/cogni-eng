package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"learning.local/workspace/internal/access"
)

func normalizeEmail(input string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(input))
	a, err := mail.ParseAddress(s)
	if err != nil || a.Address != s || len(s) > 254 {
		return "", fail(400, "invalid_email", "Enter a valid email address.")
	}
	return s, nil
}
func (s *Server) limit(ctx context.Context, key string, max int) error {
	var attempts int
	err := s.DB.QueryRow(ctx, `INSERT INTO throttle(key,window_at,attempts) VALUES($1,now(),1) ON CONFLICT(key) DO UPDATE SET attempts=CASE WHEN throttle.window_at<now()-interval '15 minutes' THEN 1 ELSE throttle.attempts+1 END,window_at=CASE WHEN throttle.window_at<now()-interval '15 minutes' THEN now() ELSE throttle.window_at END RETURNING attempts`, access.Digest(key)).Scan(&attempts)
	if err != nil {
		return err
	}
	if attempts > max {
		return fail(429, "rate_limited", "Too many attempts. Wait 15 minutes and retry.")
	}
	return nil
}
func (s *Server) signIn(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Code     string `json:"code"`
		Remember bool   `json:"remember"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return err
	}
	if len(in.Password) > 256 {
		return fail(400, "invalid_password", "Check your password.")
	}
	if err = s.limit(r.Context(), "login:"+email, 12); err != nil {
		return err
	}
	var token string
	var expiry time.Time
	err = s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		var id, hash, state string
		var secret []byte
		var mfa bool
		var last int64
		err := tx.QueryRow(ctx, "SELECT id,password_hash,state,mfa_secret,mfa_enabled,last_totp_step FROM person WHERE email=$1 FOR UPDATE", email).Scan(&id, &hash, &state, &secret, &mfa, &last)
		if errors.Is(err, pgx.ErrNoRows) {
			access.CheckPassword(s.dummy, in.Password)
			return fail(401, "invalid_credentials", "Check your sign-in details.")
		}
		if err != nil {
			return err
		}
		if !access.CheckPassword(hash, in.Password) || state != "active" {
			return fail(401, "invalid_credentials", "Check your sign-in details.")
		}
		if mfa {
			if err := s.checkFactor(ctx, tx, id, secret, last, in.Code); err != nil {
				return err
			}
		}
		token = access.Random()
		expiry = time.Now().Add(12 * time.Hour)
		if in.Remember {
			expiry = time.Now().Add(30 * 24 * time.Hour)
		}
		_, err = tx.Exec(ctx, `INSERT INTO session(id,person_id,token_hash,csrf,expires_at,device,mfa_verified) VALUES($1,$2,$3,$4,$5,$6,$7)`, access.Random(), id, access.Digest(token), access.Random(), expiry, truncate(r.UserAgent(), 200), mfa)
		if err != nil {
			return err
		}
		return audit(ctx, tx, r, Actor{ID: id}, "", "session.started", id, "")
	})
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: "workspace_session", Value: token, Path: "/", HttpOnly: true, Secure: s.Config.Env == "production", SameSite: http.SameSiteStrictMode, Expires: expiry})
	write(w, 200, map[string]string{"state": "signed_in"})
	return nil
}
func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
func (s *Server) checkFactor(ctx context.Context, tx pgx.Tx, id string, encrypted []byte, last int64, code string) error {
	secret, err := access.Open(s.Config.MFAKey, encrypted)
	if err != nil {
		return err
	}
	now := time.Now()
	step := now.Unix() / 30
	valid, err := totp.ValidateCustom(code, secret, now, totp.ValidateOpts{Period: 30, Skew: 0, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		valid = false
	}
	if valid && step > last {
		_, err = tx.Exec(ctx, "UPDATE person SET last_totp_step=$2 WHERE id=$1", id, step)
		return err
	}
	tag, err := tx.Exec(ctx, "UPDATE recovery_code SET used_at=now() WHERE person_id=$1 AND code_hash=$2 AND used_at IS NULL", id, access.Digest(code))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	return fail(401, "factor_required", "Enter a current authenticator or unused recovery code.")
}
func (s *Server) authenticated(w http.ResponseWriter, r *http.Request, fn func(context.Context, pgx.Tx, Actor) (any, error)) error {
	var result any
	err := s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		a, err := s.actor(ctx, tx, r, r.Method != "GET")
		if err != nil {
			return err
		}
		result, err = fn(ctx, tx, a)
		return err
	})
	if err != nil {
		return err
	}
	write(w, 200, result)
	return nil
}
func (s *Server) me(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		return map[string]any{"id": a.ID, "name": a.Name, "email": a.Email, "csrf": a.CSRF, "mfa_enabled": a.MFA, "session_id": a.SessionID, "environment": s.Config.Env}, nil
	})
}
func (s *Server) signOut(w http.ResponseWriter, r *http.Request) error {
	err := s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		a, err := s.actor(ctx, tx, r, true)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, "UPDATE session SET revoked_at=now() WHERE id=$1", a.SessionID)
		if err != nil {
			return err
		}
		return audit(ctx, tx, r, a, "", "session.ended", a.SessionID, "")
	})
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: "workspace_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: s.Config.Env == "production", SameSite: http.SameSiteStrictMode})
	write(w, 200, map[string]string{"state": "signed_out"})
	return nil
}
func (s *Server) sessions(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		rows, err := tx.Query(ctx, `SELECT id,device,created_at,last_seen,expires_at FROM session WHERE person_id=$1 AND revoked_at IS NULL AND expires_at>now() AND last_seen>now()-interval '30 minutes' ORDER BY created_at DESC LIMIT 50`, a.ID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		out := []map[string]any{}
		for rows.Next() {
			var id, device string
			var created, last, expiry time.Time
			if err := rows.Scan(&id, &device, &created, &last, &expiry); err != nil {
				return nil, err
			}
			out = append(out, map[string]any{"id": id, "device": device, "created_at": created, "last_seen": last, "expires_at": expiry, "current": id == a.SessionID})
		}
		return out, rows.Err()
	})
}
func (s *Server) revokeSession(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		tag, err := tx.Exec(ctx, "UPDATE session SET revoked_at=now() WHERE person_id=$1 AND (id=$2 OR $2='all') AND revoked_at IS NULL", a.ID, r.PathValue("id"))
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 0 {
			return nil, fail(404, "not_found", "Session unavailable.")
		}
		err = audit(ctx, tx, r, a, "", "session.revoked", r.PathValue("id"), "User request")
		return map[string]string{"state": "revoked"}, err
	})
}
func (s *Server) enrollMFA(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		if a.MFA {
			return nil, fail(409, "mfa_enabled", "Authenticator protection is already enabled.")
		}
		key, err := totp.Generate(totp.GenerateOpts{Issuer: "Learning Workspace", AccountName: a.Email})
		if err != nil {
			return nil, err
		}
		encrypted, err := access.Seal(s.Config.MFAKey, key.Secret())
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, "UPDATE person SET mfa_secret=$2 WHERE id=$1 AND NOT mfa_enabled", a.ID, encrypted)
		return map[string]string{"secret": key.Secret(), "url": key.URL()}, err
	})
}
func (s *Server) confirmMFA(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Code string `json:"code"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		if err := s.limit(ctx, "mfa-confirm:"+a.ID, 10); err != nil {
			return nil, err
		}
		var encrypted []byte
		var enabled bool
		err := tx.QueryRow(ctx, "SELECT mfa_secret,mfa_enabled FROM person WHERE id=$1 FOR UPDATE", a.ID).Scan(&encrypted, &enabled)
		if err != nil {
			return nil, err
		}
		if enabled || len(encrypted) == 0 {
			return nil, fail(409, "invalid_state", "Begin authenticator enrollment first.")
		}
		secret, err := access.Open(s.Config.MFAKey, encrypted)
		if err != nil {
			return nil, err
		}
		if !totp.Validate(in.Code, secret) {
			return nil, fail(400, "invalid_factor", "Check the authenticator code.")
		}
		codes := []string{}
		for i := 0; i < 8; i++ {
			code := access.Random()[:20]
			codes = append(codes, code)
			if _, err = tx.Exec(ctx, "INSERT INTO recovery_code(person_id,code_hash) VALUES($1,$2)", a.ID, access.Digest(code)); err != nil {
				return nil, err
			}
		}
		if _, err = tx.Exec(ctx, "UPDATE person SET mfa_enabled=true,last_totp_step=$2 WHERE id=$1", a.ID, time.Now().Unix()/30); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, "UPDATE session SET revoked_at=now() WHERE person_id=$1 AND id<>$2", a.ID, a.SessionID); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, "UPDATE session SET mfa_verified=true WHERE id=$1", a.SessionID); err != nil {
			return nil, err
		}
		err = audit(ctx, tx, r, a, "", "mfa.enabled", a.ID, "")
		return map[string]any{"recovery_codes": codes}, err
	})
}
func queueToken(ctx context.Context, tx pgx.Tx, person, org, kind, origin string) (string, error) {
	token := access.Random()
	id := access.Random()
	ttl := time.Hour
	if kind == "invitation" {
		ttl = 72 * time.Hour
	}
	_, err := tx.Exec(ctx, "INSERT INTO security_token(id,person_id,token_hash,kind,expires_at) VALUES($1,$2,$3,$4,$5)", id, person, access.Digest(token), kind, time.Now().Add(ttl))
	if err != nil {
		return "", err
	}
	payload, _ := json.Marshal(map[string]string{"token": token, "origin": origin})
	var institution any
	if org != "" {
		institution = org
	}
	_, err = tx.Exec(ctx, "INSERT INTO delivery_job(id,institution_id,recipient_id,kind,payload) VALUES($1,$2,$3,$4,$5)", access.Random(), institution, person, kind, payload)
	return token, err
}
func (s *Server) recoverPassword(w http.ResponseWriter, r *http.Request) error {
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
	if err = s.limit(r.Context(), "recovery:"+email, 3); err != nil {
		return err
	}
	err = s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		var id string
		err := tx.QueryRow(ctx, "SELECT id FROM person WHERE email=$1 AND state='active' FOR UPDATE", email).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		_, err = queueToken(ctx, tx, id, "", "reset", s.Config.Origin)
		return err
	})
	if err != nil {
		return err
	}
	write(w, 200, map[string]string{"state": "requested", "message": "If this account is eligible, recovery instructions will be sent."})
	return nil
}
func (s *Server) resetPassword(w http.ResponseWriter, r *http.Request) error {
	return s.consumeToken(w, r, "reset")
}
func (s *Server) acceptInvitation(w http.ResponseWriter, r *http.Request) error {
	return s.consumeToken(w, r, "invitation")
}
func (s *Server) consumeToken(w http.ResponseWriter, r *http.Request, kind string) error {
	var in struct {
		Token        string `json:"token"`
		Password     string `json:"password"`
		Code         string `json:"code"`
		AcceptPolicy bool   `json:"accept_policy"`
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if kind == "invitation" && !in.AcceptPolicy {
		return fail(400, "policy_required", "Accept the current terms and privacy policy.")
	}
	hash, err := access.Password(in.Password)
	if err != nil {
		return fail(400, "weak_password", err.Error())
	}
	if err = s.limit(r.Context(), "token:"+in.Token, 8); err != nil {
		return err
	}
	err = s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		var id, tokenID, state string
		var mfa bool
		var secret []byte
		var last int64
		err := tx.QueryRow(ctx, `SELECT p.id,t.id,p.state,p.mfa_enabled,p.mfa_secret,p.last_totp_step FROM security_token t JOIN person p ON p.id=t.person_id WHERE t.token_hash=$1 AND t.kind=$2 AND t.consumed_at IS NULL AND t.expires_at>now() FOR UPDATE OF t,p`, access.Digest(in.Token), kind).Scan(&id, &tokenID, &state, &mfa, &secret, &last)
		if errors.Is(err, pgx.ErrNoRows) {
			return fail(400, "invalid_token", "This link expired or has already been used. Request a new one.")
		}
		if err != nil {
			return err
		}
		if state == "suspended" || state == "archived" {
			return fail(400, "invalid_token", "Contact your institution to restore access.")
		}
		if mfa {
			if err = s.checkFactor(ctx, tx, id, secret, last, in.Code); err != nil {
				return err
			}
		}
		if _, err = tx.Exec(ctx, "UPDATE person SET password_hash=$2,state='active',verified_at=coalesce(verified_at,now()) WHERE id=$1", id, hash); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "UPDATE security_token SET consumed_at=now() WHERE person_id=$1 AND consumed_at IS NULL", id); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "UPDATE session SET revoked_at=now() WHERE person_id=$1 AND revoked_at IS NULL", id); err != nil {
			return err
		}
		reason := ""
		if kind == "invitation" {
			reason = "Accepted terms-v1 and privacy-v1"
			for _, policy := range []string{"terms", "privacy"} {
				if _, err = tx.Exec(ctx, "INSERT INTO policy_acknowledgment(id,person_id,policy,version) VALUES($1,$2,$3,'v1') ON CONFLICT DO NOTHING", access.Random(), id, policy); err != nil {
					return err
				}
			}
		}
		return audit(ctx, tx, r, Actor{ID: id}, "", "credential."+kind, id, reason)
	})
	if err != nil {
		return err
	}
	write(w, 200, map[string]string{"state": "complete"})
	return nil
}
