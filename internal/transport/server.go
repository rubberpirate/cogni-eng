package transport

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"learning.local/workspace/internal/access"
)

type Config struct {
	Env, Origin, Address string
	MFAKey               []byte
}

func LoadConfig() (Config, error) {
	if os.Getenv("DATABASE_URL") == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	c := Config{Env: os.Getenv("APP_ENV"), Origin: os.Getenv("PUBLIC_ORIGIN"), Address: os.Getenv("HTTP_ADDR")}
	if c.Address == "" {
		c.Address = "127.0.0.1:8080"
	}
	var err error
	c.MFAKey, err = hex.DecodeString(os.Getenv("MFA_KEY"))
	if err != nil || len(c.MFAKey) != 32 {
		return c, errors.New("MFA_KEY must contain 32 random bytes encoded as hex")
	}
	if c.Env != "development" && c.Env != "production" && c.Env != "test" {
		return c, errors.New("APP_ENV must be development, test or production")
	}
	if !strings.HasPrefix(c.Origin, "https://") && !(c.Env != "production" && strings.HasPrefix(c.Origin, "http://localhost:")) {
		return c, errors.New("PUBLIC_ORIGIN requires HTTPS in production and localhost HTTP in development")
	}
	return c, nil
}

type Server struct {
	DB     *pgxpool.Pool
	Config Config
	dummy  string
}

func New(db *pgxpool.Pool, c Config) *Server {
	dummy, _ := access.Password(access.Random())
	return &Server{db, c, dummy}
}

type problem struct {
	Status        int
	Code, Message string
}

func (e *problem) Error() string              { return e.Message }
func fail(status int, code, msg string) error { return &problem{status, code, msg} }

type endpoint func(http.ResponseWriter, *http.Request) error

func (s *Server) wrap(f endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := access.Random()[:16]
		r.Header.Set("X-Internal-Request-ID", id)
		w.Header().Set("X-Request-ID", id)
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		if r.Method != "GET" && r.Method != "HEAD" && r.Header.Get("Origin") != s.Config.Origin {
			s.writeError(w, id, fail(403, "origin_denied", "Open the application from its configured origin."))
			return
		}
		started := time.Now()
		if err := f(w, r); err != nil {
			s.writeError(w, id, err)
		}
		slog.Info("request", "id", id, "method", r.Method, "route", r.Pattern, "duration_ms", time.Since(started).Milliseconds())
	}
}
func (s *Server) writeError(w http.ResponseWriter, id string, err error) {
	p := &problem{500, "internal_error", "The operation could not finish. Retry with this request reference."}
	var known *problem
	if errors.As(err, &known) {
		p = known
	} else {
		slog.Error("request failed", "request_id", id, "error", err)
	}
	write(w, p.Status, map[string]string{"code": p.Code, "message": p.Message, "request_id": id})
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return fail(415, "content_type", "Use application/json.")
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 65536))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return fail(400, "invalid_body", "Check the fields and request size.")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fail(400, "invalid_body", "Send one JSON object.")
	}
	return nil
}

type Actor struct {
	ID, Name, Email, SessionID, CSRF string
	MFA                              bool
}

func (s *Server) actor(ctx context.Context, tx pgx.Tx, r *http.Request, mutation bool) (Actor, error) {
	var a Actor
	cookie, err := r.Cookie("workspace_session")
	if err != nil {
		return a, fail(401, "sign_in_required", "Sign in to continue.")
	}
	err = tx.QueryRow(ctx, `SELECT p.id,p.name,p.email,s.id,s.csrf,p.mfa_enabled FROM session s JOIN person p ON p.id=s.person_id WHERE s.token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at>now() AND s.last_seen>now()-interval '30 minutes' AND p.state='active' AND (NOT p.mfa_enabled OR s.mfa_verified) FOR UPDATE OF s`, access.Digest(cookie.Value)).Scan(&a.ID, &a.Name, &a.Email, &a.SessionID, &a.CSRF, &a.MFA)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, fail(401, "session_expired", "Your session ended. Sign in again.")
	}
	if err != nil {
		return a, err
	}
	if mutation && !access.Equal(r.Header.Get("X-CSRF-Token"), a.CSRF) {
		return a, fail(403, "csrf_denied", "Refresh the workspace and try again.")
	}
	_, err = tx.Exec(ctx, "UPDATE session SET last_seen=now() WHERE id=$1", a.SessionID)
	return a, err
}
func (s *Server) transaction(r *http.Request, fn func(context.Context, pgx.Tx) error) error {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func audit(ctx context.Context, tx pgx.Tx, r *http.Request, a Actor, org, action, resource, reason string) error {
	var institution any
	if org != "" {
		institution = org
	}
	var actor any
	if a.ID != "" {
		actor = a.ID
	}
	_, err := tx.Exec(ctx, `INSERT INTO audit_entry(institution_id,actor_id,action,resource_id,reason,request_id) VALUES($1,$2,$3,$4,$5,$6)`, institution, actor, action, resource, reason, r.Header.Get("X-Internal-Request-ID"))
	return err
}
func authorize(ctx context.Context, tx pgx.Tx, a Actor, org, campus, capability string) error {
	var state string
	var requireMFA bool
	err := tx.QueryRow(ctx, "SELECT state,require_staff_mfa FROM institution WHERE id=$1 FOR SHARE", org).Scan(&state, &requireMFA)
	if errors.Is(err, pgx.ErrNoRows) {
		return fail(403, "forbidden", "This workspace is unavailable.")
	}
	if err != nil {
		return err
	}
	if state != "active" {
		return fail(403, "workspace_inactive", "This institution is not active.")
	}
	rows, err := tx.Query(ctx, "SELECT role,coalesce(campus_id,'') FROM membership WHERE person_id=$1 AND institution_id=$2 AND state='active' FOR SHARE", a.ID, org)
	if err != nil {
		return err
	}
	defer rows.Close()
	allowed := false
	for rows.Next() {
		var role, scope string
		if err := rows.Scan(&role, &scope); err != nil {
			return err
		}
		if access.Grants(role, capability) && (scope == "" || campus != "" && scope == campus) {
			if requireMFA && !a.MFA {
				return fail(403, "mfa_required", "Your institution requires authenticator protection. Set it up in Access and security.")
			}
			allowed = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if !allowed {
		return fail(403, "forbidden", "Your responsibilities do not allow this operation.")
	}
	return nil
}
func requireText(s string, min, max int) error {
	if len(strings.TrimSpace(s)) < min || len(s) > max {
		return fail(400, "invalid_field", fmt.Sprintf("Use between %d and %d characters.", min, max))
	}
	return nil
}
func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"state": "alive"}) })
	m.HandleFunc("GET /health/ready", s.wrap(func(w http.ResponseWriter, r *http.Request) error {
		ctx, c := context.WithTimeout(r.Context(), time.Second)
		defer c()
		if err := s.DB.Ping(ctx); err != nil {
			return fail(503, "not_ready", "Database unavailable.")
		}
		write(w, 200, map[string]string{"state": "ready"})
		return nil
	}))
	for route, handler := range map[string]endpoint{
		"POST /api/v1/verification/resend":                              s.resendVerification,
		"POST /api/v1/register":                                         s.register,
		"POST /api/v1/verify-email":                                     s.verifyEmail,
		"GET /api/v1/institutions/{org}/admission-codes":                s.admissionCodes,
		"POST /api/v1/institutions/{org}/admission-codes":               s.createAdmissionCode,
		"DELETE /api/v1/institutions/{org}/admission-codes/{id}":        s.revokeAdmissionCode,
		"GET /api/v1/institutions/{org}/admission-queue":                s.admissionQueue,
		"POST /api/v1/institutions/{org}/admission-queue/{id}/decision": s.placeAdmission,

		"GET /api/v1/profile":                               s.profile,
		"PATCH /api/v1/profile":                             s.updateProfile,
		"POST /api/v1/password":                             s.changePassword,
		"POST /api/v1/email-change":                         s.requestEmailChange,
		"POST /api/v1/email-change/confirm":                 s.confirmEmailChange,
		"GET /api/v1/account-activity":                      s.accountActivity,
		"GET /api/v1/institutions/{org}/settings":           s.settings,
		"PATCH /api/v1/institutions/{org}/settings":         s.updateSettings,
		"PATCH /api/v1/institutions/{org}/campuses/{id}":    s.updateCampus,
		"POST /api/v1/institutions/{org}/owner-transfer":    s.transferOwner,
		"PATCH /api/v1/institutions/{org}/enrollments/{id}": s.changeEnrollment,

		"POST /api/v1/sign-in": s.signIn, "POST /api/v1/sign-out": s.signOut, "GET /api/v1/me": s.me,
		"GET /api/v1/sessions": s.sessions, "DELETE /api/v1/sessions/{id}": s.revokeSession,
		"POST /api/v1/recovery": s.recoverPassword, "POST /api/v1/reset": s.resetPassword, "POST /api/v1/accept-invitation": s.acceptInvitation,
		"POST /api/v1/mfa/enroll": s.enrollMFA, "POST /api/v1/mfa/confirm": s.confirmMFA,
		"GET /api/v1/workspaces": s.workspaces, "POST /api/v1/institutions": s.createInstitution,
		"GET /api/v1/institutions/{org}/campuses": s.campuses, "POST /api/v1/institutions/{org}/campuses": s.createCampus,
		"GET /api/v1/institutions/{org}/people": s.people, "POST /api/v1/institutions/{org}/invitations": s.invite,
		"PATCH /api/v1/institutions/{org}/memberships/{id}": s.changeMembership,
		"GET /api/v1/institutions/{org}/audit":              s.auditEntries,
		"GET /api/v1/institutions/{org}/courses":            s.courses, "POST /api/v1/institutions/{org}/courses": s.createCourse,
		"GET /api/v1/institutions/{org}/enrollments": s.enrollments, "POST /api/v1/institutions/{org}/enrollments": s.createEnrollment,
		"GET /api/v1/institutions/{org}/families": s.families, "POST /api/v1/institutions/{org}/families": s.createFamily,
		"DELETE /api/v1/institutions/{org}/families/{id}": s.removeFamily,
		"GET /api/v1/institutions/{org}/responsibilities": s.responsibilities, "POST /api/v1/institutions/{org}/responsibilities": s.createResponsibility,
		"GET /api/v1/institutions/{org}/deliveries": s.deliveries, "POST /api/v1/institutions/{org}/deliveries/{id}/retry": s.retryDelivery,
	} {
		m.HandleFunc(route, s.wrap(handler))
	}
	return m
}
