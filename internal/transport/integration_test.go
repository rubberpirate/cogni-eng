//go:build integration

package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"
	"learning.local/workspace/internal/access"
)

type harness struct {
	t                                         *testing.T
	db                                        *pgxpool.Pool
	handler                                   http.Handler
	owner, org, campus, otherOrg, otherCampus string
	cookie                                    *http.Cookie
	csrf                                      string
}

func setup(t *testing.T) *harness {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL required: PostgreSQL integration not executed")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := "verify_" + access.Random()[:16]
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	c, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	c.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(ctx, c)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		_, err := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
		if err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	paths, err := filepath.Glob("../../migrations/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		sql, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		up := strings.Split(strings.TrimPrefix(string(sql), "-- +goose Up"), "-- +goose Down")[0]
		if _, err = db.Exec(ctx, up); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, db: db, owner: access.Random(), org: access.Random(), campus: access.Random(), otherOrg: access.Random(), otherCampus: access.Random()}
	h.handler = New(db, Config{Env: "test", Origin: "http://localhost:3000", MFAKey: make([]byte, 32)}).Handler()
	hash, err := access.Password("a unique test owner phrase 8!")
	if err != nil {
		t.Fatal(err)
	}
	h.exec("INSERT INTO person(id,email,name,password_hash,state,verified_at) VALUES($1,'owner@example.invalid','Test Owner',$2,'active',now())", h.owner, hash)
	for _, id := range []string{h.org, h.otherOrg} {
		h.exec("INSERT INTO institution(id,name,timezone,locale) VALUES($1,'Test Institute','Asia/Kolkata','en')", id)
	}
	h.exec("INSERT INTO campus(id,institution_id,name,code) VALUES($1,$2,'North','N'),($3,$4,'South','S')", h.campus, h.org, h.otherCampus, h.otherOrg)
	h.exec("INSERT INTO membership(id,institution_id,person_id,role) VALUES($1,$2,$3,'owner')", access.Random(), h.org, h.owner)
	h.login("owner@example.invalid", "a unique test owner phrase 8!", "")
	return h
}
func (h *harness) exec(q string, args ...any) {
	h.t.Helper()
	if _, err := h.db.Exec(context.Background(), q, args...); err != nil {
		h.t.Fatal(err)
	}
}
func (h *harness) request(method, path string, body any, status int) map[string]any {
	h.t.Helper()
	data, _ := json.Marshal(body)
	r := httptest.NewRequest(method, path, bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", "http://localhost:3000")
	r.Header.Set("X-CSRF-Token", h.csrf)
	if h.cookie != nil {
		r.AddCookie(h.cookie)
	}
	w := httptest.NewRecorder()
	h.handler.ServeHTTP(w, r)
	if w.Code != status {
		h.t.Fatalf("%s %s got %d want %d: %s", method, path, w.Code, status, w.Body.String())
	}
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "workspace_session" {
			h.cookie = cookie
		}
	}
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return out
}
func (h *harness) login(email, password, code string) {
	h.t.Helper()
	h.request("POST", "/api/v1/sign-in", map[string]any{"email": email, "password": password, "code": code}, 200)
	me := h.request("GET", "/api/v1/me", nil, 200)
	h.csrf = me["csrf"].(string)
}
func (h *harness) path(s string) string { return "/api/v1/institutions/" + h.org + s }
func TestInstitutionIsolationAndRevocation(t *testing.T) {
	h := setup(t)
	h.request("GET", "/api/v1/institutions/"+h.otherOrg+"/people", nil, 403)
	h.request("POST", h.path("/invitations"), map[string]any{"name": "Cross Tenant", "email": "bad@example.invalid", "role": "learner", "campus_id": h.otherCampus}, 400)
	var count int
	if err := h.db.QueryRow(context.Background(), "SELECT count(*) FROM person WHERE email='bad@example.invalid'").Scan(&count); err != nil || count != 0 {
		t.Fatal("cross tenant write persisted")
	}
	learner := access.Random()
	hash, _ := access.Password("a unique learner phrase 7!")
	h.exec("INSERT INTO person(id,email,name,password_hash,state) VALUES($1,'learner@example.invalid','Learner',$2,'active')", learner, hash)
	membership := access.Random()
	h.exec("INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,'learner')", membership, h.org, learner, h.campus)
	ownerCookie, ownerCSRF := h.cookie, h.csrf
	h.login("learner@example.invalid", "a unique learner phrase 7!", "")
	learnerCookie := h.cookie
	h.request("POST", h.path("/campuses"), map[string]any{"name": "Escalation", "code": "E", "address": ""}, 403)
	h.request("POST", h.path("/invitations"), map[string]any{"name": "Escalation", "email": "escalate@example.invalid", "role": "coordinator", "campus_id": h.campus}, 403)
	h.cookie, h.csrf = ownerCookie, ownerCSRF
	h.request("PATCH", h.path("/memberships/"+membership), map[string]any{"state": "suspended", "reason": "Test access suspension", "version": 1}, 200)
	h.cookie = learnerCookie
	h.request("GET", "/api/v1/me", nil, 401)
	h.cookie = ownerCookie
	h.request("PATCH", h.path("/memberships/"+membership), map[string]any{"state": "active", "reason": "Approved reactivation", "version": 2}, 200)
	h.cookie = learnerCookie
	h.request("GET", "/api/v1/me", nil, 401)
	h.cookie = ownerCookie
	h.csrf = "forged"
	h.request("POST", h.path("/campuses"), map[string]any{"name": "Forged", "code": "F", "address": ""}, 403)
}
func TestInvitationResetSingleUseAndMFA(t *testing.T) {
	h := setup(t)
	ctx := context.Background()
	h.request("POST", h.path("/invitations"), map[string]any{"name": "Invited Teacher", "email": "invite@example.invalid", "role": "instructor", "campus_id": h.campus}, 200)
	var payload []byte
	if err := h.db.QueryRow(ctx, "SELECT payload FROM delivery_job WHERE kind='invitation'").Scan(&payload); err != nil {
		t.Fatal(err)
	}
	var p map[string]string
	_ = json.Unmarshal(payload, &p)
	body := map[string]any{"token": p["token"], "password": "a unique invite phrase 4!", "accept_policy": true}
	h.request("POST", "/api/v1/accept-invitation", body, 200)
	h.request("POST", "/api/v1/accept-invitation", body, 400)
	h.login("invite@example.invalid", "a unique invite phrase 4!", "")
	oldCookie := h.cookie
	h.request("POST", "/api/v1/recovery", map[string]any{"email": "invite@example.invalid"}, 200)
	if err := h.db.QueryRow(ctx, "SELECT payload FROM delivery_job WHERE kind='reset'").Scan(&payload); err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal(payload, &p)
	h.request("POST", "/api/v1/reset", map[string]any{"token": p["token"], "password": "a fresh reset phrase 9!"}, 200)
	h.cookie = oldCookie
	h.request("GET", "/api/v1/me", nil, 401)
	h.login("invite@example.invalid", "a fresh reset phrase 9!", "")
	enrollment := h.request("POST", "/api/v1/mfa/enroll", nil, 200)
	secret := enrollment["secret"].(string)
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	result := h.request("POST", "/api/v1/mfa/confirm", map[string]any{"code": code}, 200)
	codes := result["recovery_codes"].([]any)
	h.request("POST", "/api/v1/sign-out", nil, 200)
	h.request("POST", "/api/v1/sign-in", map[string]any{"email": "invite@example.invalid", "password": "a fresh reset phrase 9!"}, 401)
	h.login("invite@example.invalid", "a fresh reset phrase 9!", codes[0].(string))
	h.request("POST", "/api/v1/sign-out", nil, 200)
	h.request("POST", "/api/v1/sign-in", map[string]any{"email": "invite@example.invalid", "password": "a fresh reset phrase 9!", "code": codes[0]}, 401)
}
func TestScopedCoordinatorAndAuditAtomicity(t *testing.T) {
	h := setup(t)
	id, otherCampus := access.Random(), access.Random()
	hash, _ := access.Password("a coordinator test phrase 8!")
	h.exec("INSERT INTO campus(id,institution_id,name,code) VALUES($1,$2,'East','E')", otherCampus, h.org)
	h.exec("INSERT INTO person(id,email,name,password_hash,state) VALUES($1,'coordinator@example.invalid','Coordinator',$2,'active')", id, hash)
	h.exec("INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,'coordinator')", access.Random(), h.org, id, h.campus)
	h.login("coordinator@example.invalid", "a coordinator test phrase 8!", "")
	h.request("GET", h.path("/people"), nil, 403)
	h.request("GET", h.path("/people?campus="+otherCampus), nil, 403)
	r := httptest.NewRequest("GET", h.path("/people?campus="+h.campus), nil)
	r.AddCookie(h.cookie)
	w := httptest.NewRecorder()
	h.handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	h.request("POST", h.path("/invitations"), map[string]any{"name": "Other Coordinator", "email": "no@example.invalid", "role": "coordinator", "campus_id": h.campus}, 403)
	h.request("POST", h.path("/courses"), map[string]any{"name": "Science", "campus_id": h.campus, "academic_year": "2026"}, 200)
	var n int
	err := h.db.QueryRow(context.Background(), "SELECT count(*) FROM audit_entry WHERE actor_id=$1 AND action='course.created'", id).Scan(&n)
	if err != nil || n != 1 {
		t.Fatal(fmt.Sprint("audit missing: ", err))
	}
}
