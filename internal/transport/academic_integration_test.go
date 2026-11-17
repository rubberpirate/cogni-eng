//go:build integration

package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
	"learning.local/workspace/internal/blob"
	"learning.local/workspace/internal/scholar"
)

func academicFixture(t *testing.T) (*harness, *scholar.Service, string, string) {
	t.Helper()
	h := setup(t)
	ctx := context.Background()
	objects, err := blob.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	svc, err := scholar.New(h.db, objects)
	if err != nil {
		t.Fatal(err)
	}
	server := New(h.db, Config{Env: "test", Origin: "http://localhost:3000", MFAKey: make([]byte, 32)})
	server.Objects = objects
	h.handler = server.Handler()
	course, learner := access.Random(), access.Random()
	h.exec("INSERT INTO course(id,institution_id,campus_id,name,academic_year) VALUES($1,$2,$3,'Foundations','2026')", course, h.org, h.campus)
	hash, err := access.Password("unique learner academic phrase 91!")
	if err != nil {
		t.Fatal(err)
	}
	h.exec("INSERT INTO person(id,email,name,password_hash,state,verified_at) VALUES($1,'academic-learner@example.invalid','Synthetic Learner',$2,'active',now())", learner, hash)
	h.exec("INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,'learner')", access.Random(), h.org, learner, h.campus)
	h.exec("INSERT INTO enrollment(id,institution_id,person_id,campus_id,course_id,state,start_date) VALUES($1,$2,$3,$4,$5,'active',current_date)", access.Random(), h.org, learner, h.campus, course)
	if err = h.db.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	return h, svc, course, learner
}
func academicTx(t *testing.T, h *harness, fn func(pgx.Tx) error) {
	t.Helper()
	if err := scholar.Tx(context.Background(), h.db, fn); err != nil {
		t.Fatal(err)
	}
}
func academicPaper(t *testing.T, h *harness, svc *scholar.Service, course, learner string) (scholar.PaperRecord, string, academic.Question) {
	t.Helper()
	ctx := context.Background()
	var paper scholar.PaperRecord
	var assignment string
	var q academic.Question
	academicTx(t, h, func(tx pgx.Tx) error {
		topic, err := svc.CreateTopic(ctx, tx, h.owner, h.org, scholar.Topic{CampusID: h.campus, CourseID: course, Subject: "Mathematics", Name: "Whole number addition", Outcome: "Add two positive whole numbers accurately.", Language: "en"})
		if err != nil {
			return err
		}
		q = academic.Question{TopicID: topic.ID, Type: "single", Stem: "A tray has 7 seeds. Add 5 seeds. How many seeds are there?", Options: []academic.Option{{ID: "A", Text: "12"}, {ID: "B", Text: "13"}, {ID: "C", Text: "11"}, {ID: "D", Text: "2"}}, Key: []string{"A"}, Solution: "7 + 5 = 12 seeds.", Marks: "2", Penalty: "0.5", Difficulty: "standard", Language: "en", Provenance: "Original synthetic fixture", License: "CC0-1.0"}
		q, _, err = svc.CreateQuestion(ctx, tx, h.owner, h.org, q, "authored")
		if err != nil {
			return err
		}
		if err = svc.ReviewQuestion(ctx, tx, h.owner, h.org, q.ID, "approved", "Verified the answer and explanation", 1); err != nil {
			return err
		}
		paper, err = svc.Compose(ctx, tx, h.owner, h.org, scholar.Compose{CampusID: h.campus, CourseID: course, Title: "Seed arithmetic", Purpose: "practice", Instructions: "Select one answer", DurationMinutes: 30, ReleasePolicy: "immediate", Sections: []scholar.ComposeSection{{ID: "one", Title: "Addition", Policy: "all", Count: 1, QuestionIDs: []string{q.ID}}}})
		if err != nil {
			return err
		}
		if err = svc.Publish(ctx, tx, h.owner, h.org, paper.ID, 1); err != nil {
			return err
		}
		ids, err := svc.Assign(ctx, tx, h.owner, h.org, paper.ID, []string{learner}, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), 0)
		if err == nil {
			assignment = ids[0]
		}
		return err
	})
	return paper, assignment, q
}
func TestAcademicOnlineIsolationReplayImmutableAndExpiry(t *testing.T) {
	h, svc, course, learner := academicFixture(t)
	p, assignment, q := academicPaper(t, h, svc, course, learner)
	ownerCookie, ownerCSRF := h.cookie, h.csrf
	// Published paper snapshots survive creation and approval of a replacement key.
	academicTx(t, h, func(tx pgx.Tx) error {
		changed := q
		changed.Key = []string{"B"}
		_, err := svc.ReviseQuestion(context.Background(), tx, h.owner, h.org, q.ID, changed, 2)
		return err
	})
	h.login("academic-learner@example.invalid", "unique learner academic phrase 91!", "")
	h.request("GET", h.path("/education/papers/"+p.ID), nil, 403)
	h.request("GET", h.path("/education/questions?topic="+q.TopicID), nil, 403)
	out := h.request("POST", h.path("/education/assignments/"+assignment+"/start"), map[string]bool{"takeover": false}, 200)
	id := out["id"].(string)
	encoded, _ := json.Marshal(out)
	if bytes.Contains(encoded, []byte("7 + 5")) || bytes.Contains(encoded, []byte(`"key":["A"]`)) {
		t.Fatal("unreleased answer leaked")
	}
	save := map[string]any{"operation_id": "academic-save-0001", "version": 0, "answers": map[string]academic.Answer{q.ID: {Selection: []string{"A"}}}}
	out = h.request("POST", h.path("/education/sittings/"+id+"/save"), save, 200)
	if out["answer_version"] != float64(1) {
		t.Fatal("save version")
	}
	h.request("POST", h.path("/education/sittings/"+id+"/save"), save, 200)
	h.request("POST", h.path("/education/sittings/"+id+"/save"), map[string]any{"operation_id": "new-save-operation", "version": 0, "answers": map[string]academic.Answer{}}, 409)
	out = h.request("POST", h.path("/education/sittings/"+id+"/submit"), map[string]any{}, 200)
	result := out["result"].(map[string]any)
	if result["score"] != "2" || out["state"] != "released" {
		t.Fatalf("published snapshot key not respected: %+v", out)
	}
	h.request("POST", h.path("/education/sittings/"+id+"/submit"), map[string]any{}, 200)
	var count int
	if err := h.db.QueryRow(context.Background(), "SELECT count(*) FROM result_revision WHERE sitting_id=$1", id).Scan(&count); err != nil || count != 1 {
		t.Fatalf("duplicate result revisions %d %v", count, err)
	}
	h.cookie, h.csrf = ownerCookie, ownerCSRF
	h.request("GET", "/api/v1/institutions/"+h.otherOrg+"/education/sittings/"+id, nil, 404)
	// A worker expires a separate sitting while no browser is open.
	other := access.Random()
	h.exec("INSERT INTO paper(id,institution_id,campus_id,course_id,spec,state,created_by,published_by,published_at) SELECT $1,institution_id,campus_id,course_id,spec,'published',created_by,published_by,published_at FROM paper WHERE id=$2", other, p.ID)
	aid := access.Random()
	h.exec("INSERT INTO assignment(id,institution_id,paper_id,learner_id,opens_at,closes_at,assigned_by) VALUES($1,$2,$3,$4,now()-interval '1 hour',now()+interval '1 hour',$5)", aid, h.org, other, learner, h.owner)
	h.login("academic-learner@example.invalid", "unique learner academic phrase 91!", "")
	out = h.request("POST", h.path("/education/assignments/"+aid+"/start"), map[string]bool{"takeover": false}, 200)
	expired := out["id"].(string)
	h.exec("UPDATE sitting SET deadline=now()-interval '1 second' WHERE id=$1", expired)
	h.exec("UPDATE river_job SET scheduled_at=now() WHERE args->>'sitting_id'=$1", expired)
	worker, err := svc.WorkerClient()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err = worker.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := worker.Stop(context.Background()); err != nil {
			t.Error(err)
		}
	}()
	deadline := time.Now().Add(8 * time.Second)
	for {
		var state string
		if err = h.db.QueryRow(ctx, "SELECT state FROM sitting WHERE id=$1", expired).Scan(&state); err != nil {
			t.Fatal(err)
		}
		if state == "released" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("worker failed to expire browser-closed sitting: %s", state)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
func TestAcademicImportPreviewDedupAndScanPixelsToMarks(t *testing.T) {
	h, svc, course, learner := academicFixture(t)
	p, assignment, q := academicPaper(t, h, svc, course, learner)
	ctx := context.Background()
	var preview scholar.Import
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		data := []byte("type,question,A,B,answer,marks,penalty\nsingle,Original extra item,First,Second,A,1,0\nsingle,Broken answer key,First,Second,C,1,0\n")
		preview, err = svc.StageImport(ctx, tx, h.owner, h.org, q.TopicID, "original.csv", data, "Original synthetic material", "CC0-1.0")
		if err != nil {
			return err
		}
		if len(preview.Rows) != 2 || preview.Rows[1].Error == "" {
			t.Fatal("row validation missing")
		}
		repeat, e := svc.StageImport(ctx, tx, h.owner, h.org, q.TopicID, "original.csv", data, "Original synthetic material", "CC0-1.0")
		if e != nil {
			return e
		}
		if repeat.ID != preview.ID {
			t.Fatal("duplicate upload")
		}
		preview, err = svc.CommitImport(ctx, tx, h.owner, h.org, preview.ID, []int{1}, preview.Version)
		return err
	})
	if svc.Objects == nil {
		t.Skip("OBJECT_* required: private scan storage not executed")
	}
	var layout scholar.SheetLayout
	var batch scholar.Batch
	var job scholar.Job
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		layout, err = svc.CreateLayout(ctx, tx, h.owner, h.org, p.ID, 1)
		if err != nil {
			return err
		}
		batch, err = svc.CreateBatch(ctx, tx, h.owner, h.org, p.ID, layout.ID, time.Now())
		if err != nil {
			return err
		}
		data, err := academic.SheetImage(layout.Spec, assignment, layout.Fingerprint, "Seed arithmetic", "Synthetic learner", map[string]academic.Answer{q.ID: {Selection: []string{"A"}}})
		if err != nil {
			return err
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return err
		}
		var pngData bytes.Buffer
		if err = png.Encode(&pngData, img); err != nil {
			return err
		}
		job, err = svc.UploadScan(ctx, tx, h.owner, h.org, batch.ID, "filled-sheet.png", pngData.Bytes())
		if err != nil {
			return err
		}
		duplicate, err := svc.UploadScan(ctx, tx, h.owner, h.org, batch.ID, "duplicate.png", pngData.Bytes())
		if err == nil && job.ID != duplicate.ID {
			t.Fatal("source duplicated")
		}
		return err
	})
	worker, err := svc.WorkerClient()
	if err != nil {
		t.Fatal(err)
	}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err = worker.Start(workerCtx); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := worker.Stop(context.Background()); err != nil {
			t.Error(err)
		}
	}()
	var sheets []scholar.ScanSheet
	until := time.Now().Add(12 * time.Second)
	for {
		var state, problem string
		err = h.db.QueryRow(ctx, "SELECT state,error FROM academic_job WHERE id=$1", job.ID).Scan(&state, &problem)
		if err != nil {
			t.Fatal(err)
		}
		if state == "completed" {
			break
		}
		if state == "failed" || time.Now().After(until) {
			t.Fatalf("scan job %s: %s", state, problem)
		}
		time.Sleep(80 * time.Millisecond)
	}
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		sheets, err = svc.ScanSheets(ctx, tx, h.owner, h.org, batch.ID)
		if err != nil {
			return err
		}
		if len(sheets) != 1 || sheets[0].AssignmentID != assignment || sheets[0].Machine.Rows[q.ID].Answer.Selection[0] != "A" {
			t.Fatalf("scan recognition %+v", sheets)
		}
		return nil
	})
	var sitting string
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		sitting, err = svc.ReviewScan(ctx, tx, h.owner, h.org, sheets[0].ID, assignment, "Checked original sheet and identity", true, 1, map[string]academic.Answer{}, true)
		return err
	})
	academicTx(t, h, func(tx pgx.Tx) error {
		repeat, err := svc.ReviewScan(ctx, tx, h.owner, h.org, sheets[0].ID, assignment, "Replayed commitment", true, 1, nil, true)
		if err == nil && repeat != sitting {
			t.Fatal("duplicate academic evidence")
		}
		out, e := svc.Sitting(ctx, tx, h.owner, h.org, sitting)
		if e != nil {
			return e
		}
		if out.Result.Score != "2" {
			t.Fatalf("scan grade %+v", out.Result)
		}
		return err
	})
	// Private files do not become public guessed object URLs and are tenant scoped.
	h.request("GET", "/api/v1/institutions/"+h.otherOrg+"/education/assets/"+sheets[0].AssetID, nil, 404)
	request := httptest.NewRequest(http.MethodGet, h.path("/education/assets/"+sheets[0].AssetID), nil)
	response := httptest.NewRecorder()
	h.handler.ServeHTTP(response, request)
	if response.Code != 401 || strings.Contains(response.Body.String(), "academic/") {
		t.Fatal("unauthenticated private scan leak")
	}
}

func TestAcademicGenerationRequiresConfigurationAndReviewBeforeScoring(t *testing.T) {
	h, svc, course, learner := academicFixture(t)
	_, _, q := academicPaper(t, h, svc, course, learner)
	ctx := context.Background()
	svc.Models = scholar.Anthropic{}
	var job scholar.Job
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		job, err = svc.Generate(ctx, tx, h.owner, h.org, q.TopicID, 2, "standard")
		return err
	})
	start := func() func() {
		worker, err := svc.WorkerClient()
		if err != nil {
			t.Fatal(err)
		}
		if err = worker.Start(ctx); err != nil {
			t.Fatal(err)
		}
		return func() {
			stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := worker.Stop(stopCtx); err != nil {
				t.Fatal(err)
			}
		}
	}
	wait := func(want string) {
		deadline := time.Now().Add(8 * time.Second)
		for {
			var state, problem string
			if err := h.db.QueryRow(ctx, "SELECT state,error FROM academic_job WHERE id=$1", job.ID).Scan(&state, &problem); err != nil {
				t.Fatal(err)
			}
			if state == want {
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("job expected %s got %s: %s", want, state, problem)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
	stop := start()
	wait("configuration_required")
	stop()
	svc.Models = scholar.DevelopmentModel{}
	academicTx(t, h, func(tx pgx.Tx) error { return svc.ControlJob(ctx, tx, h.owner, h.org, job.ID, "retry") })
	stop = start()
	defer stop()
	wait("completed")
	academicTx(t, h, func(tx pgx.Tx) error {
		bank, err := svc.Bank(ctx, tx, h.owner, h.org, q.TopicID)
		if err != nil {
			return err
		}
		generated := []academic.Question{}
		for _, v := range bank {
			if v.Origin == "development_fixture" {
				generated = append(generated, v)
				if v.State != "draft" {
					t.Fatal("generated question bypassed approval")
				}
			}
		}
		if len(generated) != 2 {
			t.Fatalf("generated draft count %d", len(generated))
		}
		_, err = svc.Compose(ctx, tx, h.owner, h.org, scholar.Compose{CampusID: h.campus, CourseID: course, Title: "Reject unreviewed drafts", Purpose: "practice", DurationMinutes: 30, ReleasePolicy: "immediate", Sections: []scholar.ComposeSection{{ID: "one", Title: "Core", Policy: "all", Count: 1, QuestionIDs: []string{generated[0].ID}}}})
		if err == nil {
			t.Fatal("unapproved generated question entered scored paper")
		}
		return nil
	})
	if _, err := h.db.Exec(ctx, "UPDATE paper SET spec=jsonb_set(spec,'{title}','\"Illicit rewrite\"') WHERE institution_id=$1", h.org); err == nil {
		t.Fatal("published content changed")
	}
}
