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
	"learning.local/workspace/internal/scholar"
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
		up = strings.ReplaceAll(strings.ReplaceAll(up, "-- +goose StatementBegin", ""), "-- +goose StatementEnd", "")
		if _, err = db.Exec(ctx, up); err != nil {
			t.Fatal(err)
		}
	}
	if err = scholar.Migrate(ctx, db); err != nil {
		t.Fatal(err)
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

func TestAccountChangesPreserveVerificationAndRevokeOtherSessions(t *testing.T) {
	h := setup(t)
	originalCookie := h.cookie
	h.login("owner@example.invalid", "a unique test owner phrase 8!", "")
	h.request("PATCH", "/api/v1/profile", map[string]any{"name": "Renamed Owner", "phone": "+91 synthetic", "emergency_contact": "Synthetic family contact"}, 200)
	h.request("PATCH", "/api/v1/profile", map[string]any{"name": "Attacker", "email": "unverified@example.invalid", "phone": "", "emergency_contact": ""}, 400)
	h.request("POST", "/api/v1/password", map[string]any{"current_password": "a unique test owner phrase 8!", "password": "a newly chosen owner phrase 5!", "code": ""}, 200)
	currentCookie, currentCSRF := h.cookie, h.csrf
	h.cookie = originalCookie
	h.request("GET", "/api/v1/me", nil, 401)
	h.cookie, h.csrf = currentCookie, currentCSRF
	h.request("POST", "/api/v1/email-change", map[string]any{"email": "NEW-owner@example.invalid", "password": "a newly chosen owner phrase 5!", "code": ""}, 200)
	profile := h.request("GET", "/api/v1/profile", nil, 200)
	if profile["email"] != "owner@example.invalid" {
		t.Fatal("unverified email replaced identifier")
	}
	var payload []byte
	if err := h.db.QueryRow(context.Background(), "SELECT payload FROM delivery_job WHERE kind='email_change'").Scan(&payload); err != nil {
		t.Fatal(err)
	}
	var p map[string]string
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatal(err)
	}
	h.request("POST", "/api/v1/email-change/confirm", map[string]any{"token": p["token"]}, 200)
	h.request("POST", "/api/v1/email-change/confirm", map[string]any{"token": p["token"]}, 400)
	profile = h.request("GET", "/api/v1/profile", nil, 200)
	if profile["email"] != "new-owner@example.invalid" || profile["name"] != "Renamed Owner" {
		t.Fatal(profile)
	}
}
func TestPlacementReferenceHistoryAndStaleWrite(t *testing.T) {
	h := setup(t)
	ctx := context.Background()
	learner, course, target, targetCampus := access.Random(), access.Random(), access.Random(), access.Random()
	hash, _ := access.Password("synthetic learner transfer 7!")
	h.exec("INSERT INTO person(id,email,name,password_hash,state) VALUES($1,'transfer@example.invalid','Transfer Learner',$2,'active')", learner, hash)
	h.exec("INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,'learner')", access.Random(), h.org, learner, h.campus)
	h.exec("INSERT INTO campus(id,institution_id,name,code) VALUES($1,$2,'Target','T')", targetCampus, h.org)
	h.exec("INSERT INTO course(id,institution_id,campus_id,name,academic_year) VALUES($1,$2,$3,'Original','2026'),($4,$2,$5,'Target','2026')", course, h.org, h.campus, target, targetCampus)
	result := h.request("POST", h.path("/enrollments"), map[string]any{"person_id": learner, "campus_id": h.campus, "course_id": course, "start_date": "2026-01-01"}, 200)
	id, reference := result["id"].(string), result["reference"].(string)
	body := map[string]any{"campus_id": targetCampus, "course_id": target, "state": "active", "effective_on": "2026-11-17", "reason": "Approved campus transfer", "version": 1}
	h.request("PATCH", h.path("/enrollments/"+id), body, 200)
	h.request("PATCH", h.path("/enrollments/"+id), body, 409)
	var saved string
	var history int
	if err := h.db.QueryRow(ctx, "SELECT reference FROM enrollment WHERE id=$1", id).Scan(&saved); err != nil {
		t.Fatal(err)
	}
	if saved != reference {
		t.Fatal("reference changed")
	}
	if err := h.db.QueryRow(ctx, "SELECT count(*) FROM placement_event WHERE enrollment_id=$1", id).Scan(&history); err != nil || history != 1 {
		t.Fatal("missing or duplicate placement history")
	}
	var oldActive bool
	if err := h.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM membership WHERE person_id=$1 AND campus_id=$2 AND role='learner' AND state='active')", learner, h.campus).Scan(&oldActive); err != nil || oldActive {
		t.Fatal("old campus access remains")
	}
}
func TestOwnershipTransferNeverOrphansInstitution(t *testing.T) {
	h := setup(t)
	target := access.Random()
	hash, _ := access.Password("a successor owner phrase 8!")
	h.exec("INSERT INTO person(id,email,name,password_hash,state,verified_at) VALUES($1,'successor@example.invalid','Successor',$2,'active',now())", target, hash)
	h.exec("INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,'instructor')", access.Random(), h.org, target, h.campus)
	body := map[string]any{"person_id": target, "password": "incorrect", "code": "", "reason": "Approved ownership succession"}
	h.request("POST", h.path("/owner-transfer"), body, 401)
	body["password"] = "a unique test owner phrase 8!"
	h.request("POST", h.path("/owner-transfer"), body, 200)
	h.request("GET", h.path("/people"), nil, 403)
	h.login("successor@example.invalid", "a successor owner phrase 8!", "")
	h.request("GET", h.path("/people"), nil, 200)
	var owners int
	if err := h.db.QueryRow(context.Background(), "SELECT count(*) FROM membership WHERE institution_id=$1 AND role='owner' AND state='active'", h.org).Scan(&owners); err != nil || owners != 1 {
		t.Fatal("ownership integrity")
	}
}
func TestStaffMFAEnforcementAllowsEnrollmentRecoveryPath(t *testing.T) {
	h := setup(t)
	h.exec("UPDATE institution SET require_staff_mfa=true WHERE id=$1", h.org)
	h.request("GET", h.path("/people"), nil, 403)
	enrollment := h.request("POST", "/api/v1/mfa/enroll", nil, 200)
	code, err := totp.GenerateCode(enrollment["secret"].(string), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	h.request("POST", "/api/v1/mfa/confirm", map[string]any{"code": code}, 200)
	h.request("GET", h.path("/people"), nil, 200)
}

func TestLearnerRegistrationRequiresVerificationAndStaffPlacement(t *testing.T) {
	h := setup(t)
	ctx := context.Background()
	result := h.request("POST", h.path("/admission-codes"), map[string]any{"campus_id": h.campus, "label": "Synthetic intake", "expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339), "max_uses": 1}, 200)
	code := result["code"].(string)
	register := map[string]any{"name": "Registered Learner", "email": "register@example.invalid", "password": "a newly registered learner 9!", "enrollment_code": code, "accept_policy": true}
	forged := map[string]any{}
	for k, v := range register {
		forged[k] = v
	}
	forged["role"] = "owner"
	h.request("POST", "/api/v1/register", forged, 400)
	h.request("POST", "/api/v1/register", register, 200)
	h.request("POST", "/api/v1/sign-in", map[string]any{"email": "register@example.invalid", "password": "a newly registered learner 9!"}, 401)
	var admissionID string
	var version int
	var payload []byte
	if err := h.db.QueryRow(ctx, "SELECT id,version FROM admission_request").Scan(&admissionID, &version); err != nil {
		t.Fatal(err)
	}
	decision := map[string]any{"course_id": "", "start_date": "2026-11-17", "decision": "placed", "reason": "Approved initial placement", "version": version}
	h.request("POST", h.path("/admission-queue/"+admissionID+"/decision"), decision, 409)
	if err := h.db.QueryRow(ctx, "SELECT payload FROM delivery_job WHERE kind='verify'").Scan(&payload); err != nil {
		t.Fatal(err)
	}
	var p map[string]string
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatal(err)
	}
	h.request("POST", "/api/v1/verify-email", map[string]any{"token": p["token"]}, 200)
	h.request("POST", "/api/v1/verify-email", map[string]any{"token": p["token"]}, 400)
	register["email"] = "exhausted@example.invalid"
	h.request("POST", "/api/v1/register", register, 400)
	course := access.Random()
	h.exec("INSERT INTO course(id,institution_id,campus_id,name,academic_year) VALUES($1,$2,$3,'Registration course','2026')", course, h.org, h.campus)
	decision["course_id"] = course
	decision["version"] = version + 1
	h.request("POST", h.path("/admission-queue/"+admissionID+"/decision"), decision, 200)
	h.request("POST", h.path("/admission-queue/"+admissionID+"/decision"), decision, 409)
	h.login("register@example.invalid", "a newly registered learner 9!", "")
	h.request("POST", h.path("/admission-codes"), map[string]any{"campus_id": h.campus, "label": "Escalation", "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339), "max_uses": 10}, 403)
	var memberships, placements, acks int
	if err := h.db.QueryRow(ctx, "SELECT count(*) FROM membership m JOIN person p ON p.id=m.person_id WHERE p.email='register@example.invalid' AND m.role='learner'").Scan(&memberships); err != nil {
		t.Fatal(err)
	}
	if err := h.db.QueryRow(ctx, "SELECT count(*) FROM enrollment").Scan(&placements); err != nil {
		t.Fatal(err)
	}
	if err := h.db.QueryRow(ctx, "SELECT count(*) FROM policy_acknowledgment").Scan(&acks); err != nil {
		t.Fatal(err)
	}
	if memberships != 1 || placements != 1 || acks != 2 {
		t.Fatal("registration duplicated evidence or omitted acknowledgments", memberships, placements, acks)
	}
}
