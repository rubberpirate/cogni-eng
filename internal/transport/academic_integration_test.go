//go:build integration

package transport

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
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

func TestAcademicAssetsRequireSubjectRelationship(t *testing.T) {
	h, svc, course, learner := academicFixture(t)
	_, _, q := academicPaper(t, h, svc, course, learner)
	if svc.Objects == nil {
		t.Skip("OBJECT_* required: private academic asset entitlement not executed")
	}
	ctx := context.Background()
	staff := []string{access.Random(), access.Random()}
	for i, subject := range []string{"Mathematics", "Physics"} {
		h.exec("INSERT INTO person(id,email,name,password_hash,state) VALUES($1,$2,'Synthetic Instructor','not-a-login','active')", staff[i], staff[i]+"@example.invalid")
		h.exec("INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,'instructor')", access.Random(), h.org, staff[i], h.campus)
		h.exec("INSERT INTO responsibility(id,institution_id,staff_id,course_id,subject,starts_on) VALUES($1,$2,$3,$4,$5,current_date)", access.Random(), h.org, staff[i], course, subject)
	}
	var asset scholar.Asset
	academicTx(t, h, func(tx pgx.Tx) error {
		var pixels bytes.Buffer
		if err := png.Encode(&pixels, image.NewRGBA(image.Rect(0, 0, 64, 64))); err != nil {
			return err
		}
		var err error
		asset, err = svc.StoreAsset(ctx, tx, h.owner, h.org, h.campus, course, "question_image", "original-diagram.png", pixels.Bytes())
		if err != nil {
			return err
		}
		q.AssetID = asset.ID
		q.Alt = "Original synthetic diagram"
		_, err = svc.ReviseQuestion(ctx, tx, h.owner, h.org, q.ID, q, 2)
		return err
	})
	academicTx(t, h, func(tx pgx.Tx) error {
		if _, err := svc.Asset(ctx, tx, staff[0], h.org, asset.ID); err != nil {
			return err
		}
		if _, err := svc.Asset(ctx, tx, staff[1], h.org, asset.ID); err == nil {
			t.Fatal("unrelated subject instructor read private asset")
		}
		if _, err := svc.Asset(ctx, tx, learner, h.org, asset.ID); err == nil {
			t.Fatal("new bank image leaked through an older published sitting")
		}
		return nil
	})
	h.exec("UPDATE responsibility SET starts_on=current_date-2,ends_on=current_date-1 WHERE institution_id=$1 AND staff_id=$2", h.org, staff[0])
	academicTx(t, h, func(tx pgx.Tx) error {
		if _, err := svc.Asset(ctx, tx, staff[0], h.org, asset.ID); err == nil {
			t.Fatal("ended teaching relationship retained asset access")
		}
		return nil
	})
}

func TestOriginalSampleKitFilesAndPermissions(t *testing.T) {
	h, svc, course, learner := academicFixture(t)
	ctx := context.Background()
	var kit scholar.SampleKit
	var files []byte
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		kit, err = svc.CreateSampleKit(ctx, tx, h.owner, h.org, course, learner)
		if err != nil {
			return err
		}
		files, err = svc.SampleFiles(ctx, tx, h.owner, h.org, kit.PaperID, kit.AssignmentID)
		return err
	})
	archive, err := zip.NewReader(bytes.NewReader(files), int64(len(files)))
	if err != nil {
		t.Fatal(err)
	}
	contents := map[string][]byte{}
	for _, file := range archive.File {
		r, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		contents[file.Name], err = io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		if err = r.Close(); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"START-HERE.txt", "question-paper.pdf", "marking-guide.pdf", "omr-blank.pdf", "omr-correct.png", "omr-mixed.png", "omr-needs-review.png", "duplicate-pages.pdf", "questions.json", "expected-results.json"} {
		if len(contents[name]) == 0 {
			t.Errorf("missing original sample %s", name)
		}
	}
	var p scholar.PaperRecord
	var layout scholar.SheetLayout
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		p, err = svc.Paper(ctx, tx, h.owner, h.org, kit.PaperID)
		if err != nil {
			return err
		}
		layouts, err := svc.Layouts(ctx, tx, h.owner, h.org, p.ID)
		if err != nil {
			return err
		}
		layout = layouts[0]
		return nil
	})
	for _, tc := range []struct{ Name, Score string }{{"correct", "12"}, {"mixed", "5"}} {
		img, err := png.Decode(bytes.NewReader(contents["omr-"+tc.Name+".png"]))
		if err != nil {
			t.Fatal(err)
		}
		read := academic.Recognize(img, layout.Spec, layout.Fingerprint)
		if read.Error != "" || read.AssignmentID != kit.AssignmentID {
			t.Fatalf("sample sheet identity %+v", read)
		}
		answers := map[string]academic.Answer{}
		for id, row := range read.Rows {
			if row.Uncertain {
				t.Fatal("clean sample needs unexpected review")
			}
			answers[id] = row.Answer
		}
		marked, err := academic.Mark(p.Spec, answers, nil, "paper")
		if err != nil || marked.Score != tc.Score {
			t.Fatalf("sample %s expected %s got %+v %v", tc.Name, tc.Score, marked, err)
		}
	}
	img, err := png.Decode(bytes.NewReader(contents["omr-needs-review.png"]))
	if err != nil {
		t.Fatal(err)
	}
	read := academic.Recognize(img, layout.Spec, layout.Fingerprint)
	if !read.Rows[p.Spec.Sections[0].Questions[0].ID].Uncertain {
		t.Fatal("ambiguous sample silently accepted")
	}
	h.login("academic-learner@example.invalid", "unique learner academic phrase 91!", "")
	h.request("GET", h.path("/education/sample-files/"+kit.PaperID+"?assignment="+kit.AssignmentID), nil, 403)
	h.request("POST", h.path("/education/sample-kits"), map[string]string{"course_id": course, "learner_id": learner}, 403)
	started := h.request("POST", h.path("/education/assignments/"+kit.AssignmentID+"/start"), map[string]bool{"takeover": false}, 200)
	sittingID := started["id"].(string)
	h.request("POST", h.path("/education/sittings/"+sittingID+"/submit"), map[string]any{}, 200)
	academicTx(t, h, func(tx pgx.Tx) error {
		evidence, err := svc.Evidence(ctx, tx, learner, h.org, sittingID)
		if err != nil {
			return err
		}
		if len(evidence) != 1 || evidence[0].RequiredBlank != 6 || evidence[0].Answered != 0 || evidence[0].Status != "insufficient_evidence" {
			t.Fatalf("blank sample fabricated response evidence: %+v", evidence)
		}
		return nil
	})

}

func TestSampleKitRetainsTeacherEditsAndRejectsChangedKeys(t *testing.T) {
	h, svc, course, learner := academicFixture(t)
	ctx := context.Background()
	academicTx(t, h, func(tx pgx.Tx) error {
		kit, err := svc.CreateSampleKit(ctx, tx, h.owner, h.org, course, learner)
		if err != nil {
			return err
		}
		original, err := svc.Paper(ctx, tx, h.owner, h.org, kit.PaperID)
		if err != nil {
			return err
		}
		q := original.Spec.Sections[0].Questions[3]
		q.Key = []string{"A"}
		revised, err := svc.ReviseQuestion(ctx, tx, h.owner, h.org, q.ID, q, 2)
		if err != nil {
			return err
		}
		if err = svc.ReviewQuestion(ctx, tx, h.owner, h.org, revised.ID, "approved", "Synthetic changed-key regression fixture", revised.Version); err != nil {
			return err
		}
		_, err = svc.CreateSampleKit(ctx, tx, h.owner, h.org, course, learner)
		known, ok := err.(*scholar.Error)
		if !ok || known.Code != "sample_changed" {
			t.Fatalf("modified key silently reused: %v", err)
		}
		bank, err := svc.Bank(ctx, tx, h.owner, h.org, q.TopicID)
		if err != nil {
			return err
		}
		found := false
		for _, v := range bank {
			if v.ID == revised.ID {
				found = true
				if len(v.Key) != 1 || v.Key[0] != "A" {
					t.Fatal("teacher revision was silently reset")
				}
			}
		}
		if !found {
			t.Fatal("teacher revision was removed")
		}
		pinned, err := svc.Paper(ctx, tx, h.owner, h.org, kit.PaperID)
		if err != nil {
			return err
		}
		if pinned.Spec.Sections[0].Questions[3].Key[0] != "B" {
			t.Fatal("published original sample was rewritten")
		}
		return nil
	})
}

func TestSeparateKeyRequiresFacultyApplicationAndRejectsStaleOrForeignImports(t *testing.T) {
	h, svc, course, learner := academicFixture(t)
	_, _, q := academicPaper(t, h, svc, course, learner)
	ctx := context.Background()
	var imp scholar.Import
	var job scholar.Job
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		imp, err = svc.StageImport(ctx, tx, h.owner, h.org, q.TopicID, "original.csv", []byte("type,question,A,B,answer\nsingle,Original question without a key,First,Second,\n"), "Original fixture", "CC0-1.0")
		if err != nil {
			return err
		}
		if imp.Rows[0].Error == "" {
			t.Fatal("missing key accepted")
		}
		job, err = svc.StageKey(ctx, tx, h.owner, h.org, imp.ID, "key.csv", []byte("question,answer\n1,B\n"), "tesseract")
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
	deadline := time.Now().Add(8 * time.Second)
	for {
		var state string
		err = h.db.QueryRow(ctx, "SELECT state FROM academic_job WHERE id=$1", job.ID).Scan(&state)
		if err != nil {
			t.Fatal(err)
		}
		if state == "completed" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("key worker state %s", state)
		}
		time.Sleep(50 * time.Millisecond)
	}
	academicTx(t, h, func(tx pgx.Tx) error {
		before, e := svc.Import(ctx, tx, h.owner, h.org, imp.ID)
		if e != nil {
			return e
		}
		if len(before.Rows[0].Question.Key) > 0 {
			t.Fatal("key was silently applied")
		}
		return nil
	})
	h.login("academic-learner@example.invalid", "unique learner academic phrase 91!", "")
	h.request("POST", h.path("/education/imports/"+imp.ID+"/apply-key"), map[string]any{"job_id": job.ID, "version": imp.Version}, 403)
	var applied scholar.Import
	academicTx(t, h, func(tx pgx.Tx) error {
		var e error
		applied, e = svc.ApplyKey(ctx, tx, h.owner, h.org, imp.ID, job.ID, imp.Version)
		return e
	})
	if len(applied.Rows[0].Question.Key) != 1 || applied.Rows[0].Question.Key[0] != "B" || applied.Rows[0].Error != "" {
		t.Fatalf("applied %+v", applied)
	}
	if err = scholar.Tx(ctx, h.db, func(tx pgx.Tx) error {
		_, e := svc.ApplyKey(ctx, tx, h.owner, h.org, imp.ID, job.ID, imp.Version)
		return e
	}); err == nil {
		t.Fatal("stale faculty version accepted")
	}
	var other scholar.Import
	academicTx(t, h, func(tx pgx.Tx) error {
		var e error
		other, e = svc.StageImport(ctx, tx, h.owner, h.org, q.TopicID, "other.csv", []byte("type,question,A,B,answer\nsingle,Another original question,One,Two,\n"), "Original fixture", "CC0-1.0")
		return e
	})
	if err = scholar.Tx(ctx, h.db, func(tx pgx.Tx) error {
		_, e := svc.ApplyKey(ctx, tx, h.owner, h.org, other.ID, job.ID, other.Version)
		return e
	}); err == nil {
		t.Fatal("key from another import accepted")
	}
	academicTx(t, h, func(tx pgx.Tx) error {
		committed, e := svc.CommitImport(ctx, tx, h.owner, h.org, imp.ID, []int{1}, applied.Version)
		if e == nil && committed.Rows[0].CommittedID == "" {
			t.Fatal("reviewed key did not persist to bank")
		}
		return e
	})
}

func TestCalibratedGridPreviewRequiresActiveCourseStaffAndCreatesNoMarks(t *testing.T) {
	h, svc, course, learner := academicFixture(t)
	ctx := context.Background()
	img := image.NewRGBA(image.Rect(0, 0, 400, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.White)
		}
	}
	var data bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	c := academic.GridCalibration{Options: 4, Radius: .008, Blocks: []academic.GridBlock{{First: 1, Count: 2, TopLeft: academic.Point{X: .2, Y: .2}, TopRight: academic.Point{X: .8, Y: .2}, BottomLeft: academic.Point{X: .2, Y: .8}, BottomRight: academic.Point{X: .8, Y: .8}}}}
	academicTx(t, h, func(tx pgx.Tx) error {
		out, e := svc.GridPreview(ctx, tx, h.owner, h.org, course, data.Bytes(), c)
		if e == nil && (len(out.Rows) != 2 || !out.Rows[0].Uncertain || len(out.Rows[0].Selection) != 0) {
			t.Fatal("blank preview fabricated a response")
		}
		return e
	})
	if err := scholar.Tx(ctx, h.db, func(tx pgx.Tx) error {
		_, e := svc.GridPreview(ctx, tx, learner, h.org, course, data.Bytes(), c)
		return e
	}); err == nil {
		t.Fatal("learner invoked staff diagnostic")
	}
	if err := scholar.Tx(ctx, h.db, func(tx pgx.Tx) error {
		_, e := svc.GridPreview(ctx, tx, h.owner, h.otherOrg, course, data.Bytes(), c)
		return e
	}); err == nil {
		t.Fatal("cross-institution course accepted")
	}
	var n int
	if err := h.db.QueryRow(ctx, "SELECT count(*) FROM result_revision").Scan(&n); err != nil || n != 0 {
		t.Fatalf("diagnostic fabricated marks: %d %v", n, err)
	}
}

func TestVisionModelJobsRetainDraftBoundariesAndFlagOMRDisagreement(t *testing.T) {
	h, svc, course, learner := academicFixture(t)
	_, _, q := academicPaper(t, h, svc, course, learner)
	ctx := context.Background()
	modelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer synthetic-vision-contract-key" {
			t.Error("missing private runtime authentication")
		}
		var input struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		text := "Q.1. Original model transcription fixture\n(a) First\n(b) Second\n(c) Third\n(d) Fourth\nAnswer: B"
		revision := "contract-revision"
		if input.Model == scholar.BubbleMarkModel {
			revision = scholar.BubbleMarkRevision
			text = `{"classes":[0,0,1,0,1,0,0,0],"margins":[5,5,5,5,5,5,5,5],"revision":"` + revision + `"}`
		}
		if input.Model == scholar.OMRVisionModel {
			text = `{"rows":[{"number":1,"selection":[3],"uncertain":false},{"number":2,"selection":[1],"uncertain":false}]}`
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"model": input.Model, "model_revision": revision, "choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]string{"content": text}}}}); err != nil {
			t.Error(err)
		}
	}))
	defer modelServer.Close()
	t.Setenv("VISION_BASE_URL", modelServer.URL+"/v1")
	t.Setenv("VISION_API_KEY", "synthetic-vision-contract-key")
	img := image.NewRGBA(image.Rect(0, 0, 400, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.White)
		}
	}
	for y := 115; y <= 125; y++ {
		for x := 235; x <= 245; x++ {
			img.Set(x, y, color.Black)
		}
	}
	var data bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	calibration := academic.GridCalibration{Options: 4, Radius: .008, Blocks: []academic.GridBlock{{First: 1, Count: 2, TopLeft: academic.Point{X: .2, Y: .2}, TopRight: academic.Point{X: .8, Y: .2}, BottomLeft: academic.Point{X: .2, Y: .8}, BottomRight: academic.Point{X: .8, Y: .8}}}}
	var imp scholar.Import
	var job scholar.Job
	academicTx(t, h, func(tx pgx.Tx) error {
		var err error
		imp, err = svc.StageDocument(ctx, tx, h.owner, h.org, q.TopicID, "original.png", data.Bytes(), "Original contract fixture", "CC0-1.0", "lightonocr")
		if err != nil {
			return err
		}
		job, err = svc.StageGridModel(ctx, tx, h.owner, h.org, course, "original-grid.png", data.Bytes(), calibration)
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
	deadline := time.Now().Add(10 * time.Second)
	for {
		var count int
		err = h.db.QueryRow(ctx, "SELECT count(*) FROM academic_job WHERE state='completed' AND (id=$1 OR resource_id=$2)", job.ID, imp.ID).Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count == 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("vision jobs did not finish")
		}
		time.Sleep(50 * time.Millisecond)
	}
	academicTx(t, h, func(tx pgx.Tx) error {
		preview, e := svc.Import(ctx, tx, h.owner, h.org, imp.ID)
		if e != nil {
			return e
		}
		if len(preview.Rows) != 1 || len(preview.Rows[0].Question.Key) != 0 || preview.Rows[0].Error == "" {
			t.Fatal("model OCR silently approved a solved-paper key")
		}
		j, e := svc.Job(ctx, tx, h.owner, h.org, job.ID)
		if e != nil {
			return e
		}
		var output scholar.GridModelOutput
		if e = json.Unmarshal(j.Output, &output); e != nil {
			return e
		}
		if len(output.Reading.Rows) != 2 || output.Reading.Rows[0].ModelState != "agrees" || output.Reading.Rows[1].ModelState != "disagrees" || !output.Reading.Rows[1].Uncertain || len(output.Reading.Rows[1].Selection) != 0 {
			t.Fatalf("model changed observed pixels: %+v", output.Reading.Rows)
		}
		return nil
	})
	var count int
	if err = h.db.QueryRow(ctx, "SELECT count(*) FROM result_revision").Scan(&count); err != nil || count != 0 {
		t.Fatalf("vision proposal created marks %d %v", count, err)
	}
}
