package transport

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/scholar"
)

func (s *Server) academicAPI(w http.ResponseWriter, r *http.Request) error {
	svc, err := scholar.New(s.DB, s.Objects)
	if err != nil {
		return err
	}
	org, id, kind, action := r.PathValue("org"), r.PathValue("id"), r.PathValue("kind"), r.PathValue("action")
	query := r.URL.Query()
	if r.Method == "GET" && (kind == "assets" || kind == "sheet-image" || kind == "print-paper" || kind == "sample-files") {
		var data []byte
		content := "image/png"
		filename := "answer-sheet.png"
		err = s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
			a, e := s.actor(ctx, tx, r, false)
			if e != nil {
				return e
			}
			switch kind {
			case "sample-files":
				content = "application/zip"
				filename = "practice-sample-kit.zip"
				data, e = svc.SampleFiles(ctx, tx, a.ID, org, id, query.Get("assignment"))
				return e
			case "print-paper":
				p, e := svc.Paper(ctx, tx, a.ID, org, id)
				if e != nil {
					return e
				}
				if p.State != "published" {
					return fail(409, "paper_unpublished", "Publish the reviewed paper before printing.")
				}
				content = "text/html; charset=utf-8"
				filename = "assessment.html"
				data, e = scholar.PrintPaper(p, org, query.Get("answers") == "1")
				return e
			case "assets":
				asset, e := svc.Asset(ctx, tx, a.ID, org, id)
				if e != nil {
					return e
				}
				if s.Objects == nil {
					return fail(503, "storage_configuration_required", "Private storage is not configured.")
				}
				data, e = s.Objects.GetAcademic(ctx, asset.Key)
				content = asset.Mime
				filename = asset.Filename
				return e
			default:
				data, e = svc.Sheet(ctx, tx, a.ID, org, id, query.Get("assignment"))
				return e
			}
		})
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", content)
		w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": filename}))
		_, err = w.Write(data)
		return err
	}
	// Binary uploads have a bounded body and never expose an object-store URL.
	if r.Method == "POST" && (kind == "imports" && (action == "upload" || action == "key-upload") || kind == "batches" && action == "upload" || kind == "assets" || kind == "grid-preview") {
		if err = s.transaction(r, func(ctx context.Context, tx pgx.Tx) error { _, err := s.actor(ctx, tx, r, true); return err }); err != nil {
			return err
		}
		r.Body = http.MaxBytesReader(w, r.Body, 51<<20)
		if err = r.ParseMultipartForm(1 << 20); err != nil {
			return fail(400, "invalid_upload", "Upload one file of at most 50 MiB.")
		}
		defer func() {
			if r.MultipartForm != nil {
				_ = r.MultipartForm.RemoveAll()
			}
		}()
		file, header, e := r.FormFile("file")
		if e != nil {
			return fail(400, "file_required", "Choose a file.")
		}
		defer func() { _ = file.Close() }()
		limit := int64(50 << 20)
		if kind == "imports" {
			limit = 50 << 20
		}
		data, e := io.ReadAll(io.LimitReader(file, limit+1))
		if e != nil {
			return e
		}
		if int64(len(data)) > limit {
			return fail(400, "file_too_large", "File exceeds the upload limit.")
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			switch kind {
			case "grid-preview":
				var c academic.GridCalibration
				if json.Unmarshal([]byte(r.FormValue("calibration")), &c) != nil {
					return nil, fail(400, "invalid_calibration", "Provide the bubble-center calibration.")
				}
				return svc.GridPreview(ctx, tx, a.ID, org, r.FormValue("course_id"), data, c)
			case "imports":
				if action == "key-upload" {
					return svc.StageKey(ctx, tx, a.ID, org, id, header.Filename, data)
				}
				mime := http.DetectContentType(data)
				if mime == "application/pdf" || mime == "image/png" || mime == "image/jpeg" {
					return svc.StageDocument(ctx, tx, a.ID, org, r.FormValue("topic_id"), header.Filename, data, r.FormValue("provenance"), r.FormValue("license"))
				}
				return svc.StageImport(ctx, tx, a.ID, org, r.FormValue("topic_id"), header.Filename, data, r.FormValue("provenance"), r.FormValue("license"))
			case "batches":
				return svc.UploadScan(ctx, tx, a.ID, org, id, header.Filename, data)
			default:
				return svc.StoreAsset(ctx, tx, a.ID, org, r.FormValue("campus_id"), r.FormValue("course_id"), r.FormValue("purpose"), header.Filename, data)
			}
		})
	}
	if r.Method == "GET" {
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			switch kind {
			case "evidence":
				return svc.Evidence(ctx, tx, a.ID, org, id)
			case "result-history":
				return svc.History(ctx, tx, a.ID, org, id)
			case "tutor":
				return svc.Thread(ctx, tx, a.ID, org, query.Get("sitting"), query.Get("question"))
			case "course-learners":
				return svc.CourseLearners(ctx, tx, a.ID, org, id)
			case "courses":
				return svc.Courses(ctx, tx, a.ID, org, query.Get("campus"))
			case "topics":
				return svc.Topics(ctx, tx, a.ID, org, query.Get("campus"))
			case "questions":
				return svc.Bank(ctx, tx, a.ID, org, query.Get("topic"))
			case "imports":
				if id != "" {
					return svc.Import(ctx, tx, a.ID, org, id)
				}
				return svc.Imports(ctx, tx, a.ID, org, query.Get("topic"))
			case "papers":
				if id != "" {
					return svc.Paper(ctx, tx, a.ID, org, id)
				}
				return svc.Papers(ctx, tx, a.ID, org, query.Get("campus"))
			case "recipients":
				return svc.Recipients(ctx, tx, a.ID, org, id)
			case "assignments":
				return svc.Assignments(ctx, tx, a.ID, org, query.Get("paper"))
			case "sittings":
				return svc.Sitting(ctx, tx, a.ID, org, id)
			case "layouts":
				return svc.Layouts(ctx, tx, a.ID, org, query.Get("paper"))
			case "batches":
				return svc.Batches(ctx, tx, a.ID, org, query.Get("paper"))
			case "scan-sheets":
				return svc.ScanSheets(ctx, tx, a.ID, org, query.Get("batch"))
			case "jobs":
				if id != "" {
					return svc.Job(ctx, tx, a.ID, org, id)
				}
				return svc.Jobs(ctx, tx, a.ID, org, query.Get("campus"))
			default:
				return nil, fail(404, "academic_route_unavailable", "Academic operation unavailable.")
			}
		})
	}
	if r.Method != "POST" {
		return fail(405, "method_not_allowed", "Use GET or POST.")
	}
	// Each operation uses a narrow request type and strict JSON decoding.
	switch kind {
	case "sample-kits":
		var in struct {
			Course  string `json:"course_id"`
			Learner string `json:"learner_id"`
		}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			return svc.CreateSampleKit(ctx, tx, a.ID, org, in.Course, in.Learner)
		})
	case "tutor":
		var in struct {
			Sitting  string `json:"sitting_id"`
			Question string `json:"question_id"`
			Text     string `json:"text"`
		}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			return svc.AskTutor(ctx, tx, a.ID, org, in.Sitting, in.Question, in.Text)
		})
	case "topics":
		var in scholar.Topic
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			return svc.CreateTopic(ctx, tx, a.ID, org, in)
		})
	case "questions":
		if action == "review" {
			var in struct {
				State   string `json:"state"`
				Note    string `json:"note"`
				Version int    `json:"version"`
			}
			if err = decode(w, r, &in); err != nil {
				return err
			}
			return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
				return map[string]bool{"saved": true}, svc.ReviewQuestion(ctx, tx, a.ID, org, id, in.State, in.Note, in.Version)
			})
		}
		var in struct {
			Question academic.Question `json:"question"`
			Version  int               `json:"version"`
		}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			if id != "" {
				return svc.ReviseQuestion(ctx, tx, a.ID, org, id, in.Question, in.Version)
			}
			q, duplicate, e := svc.CreateQuestion(ctx, tx, a.ID, org, in.Question, "authored")
			return map[string]any{"question": q, "duplicate": duplicate}, e
		})
	case "imports":
		if action == "apply-key" {
			var in struct {
				JobID   string `json:"job_id"`
				Version int    `json:"version"`
			}
			if err = decode(w, r, &in); err != nil {
				return err
			}
			return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
				return svc.ApplyKey(ctx, tx, a.ID, org, id, in.JobID, in.Version)
			})
		}
		if action == "parse" {
			var in struct {
				Text       string `json:"text"`
				Provenance string `json:"provenance"`
				License    string `json:"license"`
				Version    int    `json:"version"`
			}
			if err = decode(w, r, &in); err != nil {
				return err
			}
			return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
				return svc.ParseDocument(ctx, tx, a.ID, org, id, in.Text, in.Provenance, in.License, in.Version)
			})
		}
		var in struct {
			Index    int               `json:"index"`
			Question academic.Question `json:"question"`
			Version  int               `json:"version"`
			Selected []int             `json:"selected"`
		}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			if action == "commit" {
				return svc.CommitImport(ctx, tx, a.ID, org, id, in.Selected, in.Version)
			}
			if action == "edit" {
				return svc.EditImport(ctx, tx, a.ID, org, id, in.Index, in.Version, in.Question)
			}
			return nil, fail(404, "academic_route_unavailable", "Import operation unavailable.")
		})
	case "papers":
		if action == "publish" {
			var in struct {
				Version int `json:"version"`
			}
			if err = decode(w, r, &in); err != nil {
				return err
			}
			return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
				return map[string]bool{"published": true}, svc.Publish(ctx, tx, a.ID, org, id, in.Version)
			})
		}
		var in scholar.Compose
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) { return svc.Compose(ctx, tx, a.ID, org, in) })
	case "blueprint":
		var in struct {
			Topic      string `json:"topic_id"`
			Count      int    `json:"count"`
			Difficulty string `json:"difficulty"`
		}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			return svc.Blueprint(ctx, tx, a.ID, org, in.Topic, in.Count, in.Difficulty)
		})
	case "assignments":
		if action == "start" {
			var in struct {
				Takeover bool `json:"takeover"`
			}
			if err = decode(w, r, &in); err != nil {
				return err
			}
			return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
				return svc.Start(ctx, tx, a.ID, a.SessionID, org, id, in.Takeover)
			})
		}
		var in struct {
			Paper    string    `json:"paper_id"`
			Learners []string  `json:"learners"`
			Opens    time.Time `json:"opens_at"`
			Closes   time.Time `json:"closes_at"`
			Extra    int       `json:"extra_minutes"`
		}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			return svc.Assign(ctx, tx, a.ID, org, in.Paper, in.Learners, in.Opens, in.Closes, in.Extra)
		})
	case "sittings":
		switch action {
		case "save":
			var in struct {
				Operation string                     `json:"operation_id"`
				Version   int                        `json:"version"`
				Answers   map[string]academic.Answer `json:"answers"`
			}
			if err = decode(w, r, &in); err != nil {
				return err
			}
			return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
				return svc.Save(ctx, tx, a.ID, a.SessionID, org, id, in.Operation, in.Version, in.Answers)
			})
		case "submit", "release":
			var in struct{}
			if err = decode(w, r, &in); err != nil {
				return err
			}
			return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
				if action == "submit" {
					return svc.Submit(ctx, tx, a.ID, a.SessionID, org, id)
				}
				return map[string]bool{"released": true}, svc.Release(ctx, tx, a.ID, org, id)
			})
		case "review":
			var in struct {
				Reason   string                         `json:"reason"`
				Revision int                            `json:"revision"`
				Manual   map[string]academic.ManualMark `json:"manual_marks"`
			}
			if err = decode(w, r, &in); err != nil {
				return err
			}
			return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
				return svc.ReviewMarks(ctx, tx, a.ID, org, id, in.Reason, in.Revision, in.Manual)
			})
		}
	case "layouts":
		var in struct {
			Paper   string `json:"paper_id"`
			Columns int    `json:"columns"`
		}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			return svc.CreateLayout(ctx, tx, a.ID, org, in.Paper, in.Columns)
		})
	case "batches":
		var in struct {
			Paper  string    `json:"paper_id"`
			Layout string    `json:"layout_id"`
			At     time.Time `json:"sitting_at"`
		}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			return svc.CreateBatch(ctx, tx, a.ID, org, in.Paper, in.Layout, in.At)
		})
	case "scan-sheets":
		var in struct {
			Assignment  string                     `json:"assignment_id"`
			Reason      string                     `json:"reason"`
			Confirmed   bool                       `json:"identity_confirmed"`
			Version     int                        `json:"version"`
			Corrections map[string]academic.Answer `json:"corrections"`
			Commit      bool                       `json:"commit"`
		}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			sitting, e := svc.ReviewScan(ctx, tx, a.ID, org, id, in.Assignment, in.Reason, in.Confirmed, in.Version, in.Corrections, in.Commit)
			return map[string]string{"sitting_id": sitting}, e
		})
	case "generate":
		var in struct {
			Topic      string `json:"topic_id"`
			Count      int    `json:"count"`
			Difficulty string `json:"difficulty"`
		}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			return svc.Generate(ctx, tx, a.ID, org, in.Topic, in.Count, in.Difficulty)
		})
	case "jobs":
		var in struct{}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
			return map[string]bool{"saved": true}, svc.ControlJob(ctx, tx, a.ID, org, id, action)
		})
	}
	return fail(404, "academic_route_unavailable", "Academic operation unavailable.")
}
