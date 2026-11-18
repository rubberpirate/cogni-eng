package scholar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
)

type Asset struct {
	ID       string `json:"id"`
	CampusID string `json:"campus_id"`
	CourseID string `json:"course_id"`
	Purpose  string `json:"purpose"`
	Key      string `json:"-"`
	Filename string `json:"filename"`
	Mime     string `json:"media_type"`
	Bytes    int64  `json:"bytes"`
	SHA      string `json:"sha256"`
	State    string `json:"state"`
}

func (s *Service) StoreAsset(ctx context.Context, tx pgx.Tx, actor, org, campus, course, purpose, filename string, data []byte) (Asset, error) {
	out := Asset{}
	if s.Objects == nil {
		return out, Problem(503, "storage_configuration_required", "Configure private object storage before uploading.")
	}
	if err := Scope(ctx, tx, actor, org, campus, course, ""); err != nil {
		return out, err
	}
	if len(data) == 0 || len(data) > 50<<20 {
		return out, Problem(400, "invalid_file", "File exceeds 50 MiB.")
	}
	mime := http.DetectContentType(data)
	if mime != "application/pdf" && mime != "image/png" && mime != "image/jpeg" {
		return out, Problem(400, "invalid_file", "Use PDF, PNG or JPEG.")
	}
	if purpose != "document" && purpose != "question_image" && purpose != "scan" && purpose != "scan_page" {
		return out, Problem(400, "invalid_purpose", "Unsupported asset purpose.")
	}
	if purpose == "question_image" && mime == "application/pdf" {
		return out, Problem(400, "invalid_file", "Question images must be PNG or JPEG.")
	}
	if mime != "application/pdf" {
		cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || cfg.Width < 32 || cfg.Height < 32 || int64(cfg.Width)*int64(cfg.Height) > 32_000_000 {
			return out, Problem(400, "invalid_image", "Image must be decodable and at most 32 megapixels.")
		}
	}
	id := access.Random()
	key := "academic/" + org + "/" + id
	out = Asset{id, campus, course, purpose, key, filepath.Base(filename), mime, int64(len(data)), access.Digest(string(data)), "staged"}
	_, err := tx.Exec(ctx, "INSERT INTO academic_asset(id,institution_id,campus_id,course_id,owner_id,purpose,object_key,filename,media_type,bytes,sha256,state) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'staged')", id, org, campus, course, actor, purpose, key, out.Filename, mime, out.Bytes, out.SHA)
	if err != nil {
		return out, err
	}
	if err = s.Objects.PutAcademic(ctx, key, mime, data); err != nil {
		return out, Problem(503, "storage_failed", "The file could not be stored. Retry the upload; no academic evidence was committed.")
	}
	_, err = tx.Exec(ctx, "UPDATE academic_asset SET state='ready' WHERE id=$1", id)
	out.State = "ready"
	return out, err
}
func loadAsset(ctx context.Context, tx pgx.Tx, org, id string) (Asset, error) {
	var a Asset
	err := tx.QueryRow(ctx, "SELECT id,campus_id,course_id,purpose,object_key,filename,media_type,bytes,sha256,state FROM academic_asset WHERE institution_id=$1 AND id=$2 AND state='ready'", org, id).Scan(&a.ID, &a.CampusID, &a.CourseID, &a.Purpose, &a.Key, &a.Filename, &a.Mime, &a.Bytes, &a.SHA, &a.State)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, Problem(404, "asset_unavailable", "File unavailable.")
	}
	return a, err
}
func (s *Service) Asset(ctx context.Context, tx pgx.Tx, actor, org, id string) (Asset, error) {
	a, err := loadAsset(ctx, tx, org, id)
	if err != nil {
		return a, err
	}
	if err = Scope(ctx, tx, actor, org, a.CampusID, a.CourseID, ""); err == nil {
		var broad, owned bool
		if e := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM membership WHERE institution_id=$1 AND person_id=$2 AND state='active' AND role IN ('owner','coordinator') AND (campus_id IS NULL OR campus_id=$3)),EXISTS(SELECT 1 FROM academic_asset WHERE institution_id=$1 AND id=$4 AND owner_id=$2)", org, actor, a.CampusID, id).Scan(&broad, &owned); e != nil {
			return Asset{}, e
		}
		if broad || owned {
			return a, nil
		}
		type subject struct{ ID string }
		if a.Purpose == "question_image" || a.Purpose == "document" {
			query := "SELECT DISTINCT q.topic_id FROM question_record q JOIN question_revision v ON v.record_id=q.id AND v.institution_id=q.institution_id WHERE q.institution_id=$1 AND (v.content->>'asset_id'=$2 OR EXISTS(SELECT 1 FROM jsonb_array_elements(coalesce(v.content->'options','[]'::jsonb)) o WHERE o->>'asset_id'=$2))"
			if a.Purpose == "document" {
				query = "SELECT topic_id FROM question_import WHERE institution_id=$1 AND asset_id=$2 UNION SELECT input->>'topic_id' FROM academic_job WHERE institution_id=$1 AND input->>'asset_id'=$2 AND kind='extract_document'"
			}
			topics, e := rows[subject](ctx, tx, query, org, id)
			if e != nil {
				return Asset{}, e
			}
			for _, v := range topics {
				t, e := GetTopic(ctx, tx, org, v.ID)
				if e != nil {
					return Asset{}, e
				}
				if e = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); e == nil {
					return a, nil
				}
			}
		} else {
			papers, e := rows[subject](ctx, tx, `SELECT b.paper_id FROM scan_batch b JOIN scan_source src ON src.batch_id=b.id AND src.institution_id=b.institution_id WHERE b.institution_id=$1 AND src.asset_id=$2 UNION SELECT b.paper_id FROM scan_batch b JOIN scan_sheet sh ON sh.batch_id=b.id AND sh.institution_id=b.institution_id WHERE b.institution_id=$1 AND sh.asset_id=$2`, org, id)
			if e != nil {
				return Asset{}, e
			}
			for _, v := range papers {
				if _, e = s.Paper(ctx, tx, actor, org, v.ID); e == nil {
					return a, nil
				}
			}
		}
	} else {
		var known *Error
		if !errors.As(err, &known) {
			return Asset{}, err
		}
	}
	if a.Purpose != "question_image" {
		return Asset{}, Problem(403, "asset_forbidden", "This private file is outside your subject or paper responsibilities.")
	}

	type candidate struct{ ID string }
	list, e := rows[candidate](ctx, tx, "SELECT id FROM sitting WHERE institution_id=$1 AND learner_id=$2", org, actor)
	if e != nil {
		return a, e
	}
	for _, v := range list {
		out, e := s.Sitting(ctx, tx, actor, org, v.ID)
		if e != nil {
			continue
		}
		for _, sec := range out.Paper.Spec.Sections {
			for _, q := range sec.Questions {
				if q.AssetID == id {
					return a, nil
				}
				for _, o := range q.Options {
					if o.AssetID == id {
						return a, nil
					}
				}
			}
		}
	}
	return Asset{}, Problem(403, "asset_forbidden", "This file is outside your assigned work.")
}

type SheetLayout struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Spec        academic.Layout `json:"spec"`
	Fingerprint string          `json:"fingerprint"`
}

func (s *Service) CreateLayout(ctx context.Context, tx pgx.Tx, actor, org, paper string, columns int) (SheetLayout, error) {
	p, err := s.Paper(ctx, tx, actor, org, paper)
	if err != nil {
		return SheetLayout{}, err
	}
	if p.State != "published" {
		return SheetLayout{}, Problem(409, "paper_unpublished", "Publish the immutable paper before creating sheets.")
	}
	spec, err := academic.NewLayout(p.Spec, columns)
	if err != nil {
		return SheetLayout{}, Invalid(err)
	}
	out := SheetLayout{access.Random(), fmt.Sprintf("Registered circles / %d column(s)", columns), spec, access.Digest(string(Encode(spec)))}
	_, err = tx.Exec(ctx, "INSERT INTO sheet_layout(id,institution_id,paper_id,name,spec,fingerprint) VALUES($1,$2,$3,$4,$5,$6)", out.ID, org, paper, out.Name, Encode(spec), out.Fingerprint)
	return out, err
}
func (s *Service) Layouts(ctx context.Context, tx pgx.Tx, actor, org, paper string) ([]SheetLayout, error) {
	if _, err := s.Paper(ctx, tx, actor, org, paper); err != nil {
		return nil, err
	}
	type row struct {
		ID, Name    string
		Spec        []byte
		Fingerprint string
	}
	list, err := rows[row](ctx, tx, "SELECT id,name,spec,fingerprint FROM sheet_layout WHERE institution_id=$1 AND paper_id=$2 ORDER BY created_at", org, paper)
	if err != nil {
		return nil, err
	}
	out := []SheetLayout{}
	for _, r := range list {
		var spec academic.Layout
		if err = json.Unmarshal(r.Spec, &spec); err != nil {
			return nil, err
		}
		out = append(out, SheetLayout{r.ID, r.Name, spec, r.Fingerprint})
	}
	return out, nil
}
func loadLayout(ctx context.Context, tx pgx.Tx, org, id string) (SheetLayout, string, error) {
	var l SheetLayout
	var data []byte
	var paper string
	err := tx.QueryRow(ctx, "SELECT id,name,spec,fingerprint,paper_id FROM sheet_layout WHERE institution_id=$1 AND id=$2", org, id).Scan(&l.ID, &l.Name, &data, &l.Fingerprint, &paper)
	if errors.Is(err, pgx.ErrNoRows) {
		return l, paper, Problem(404, "layout_unavailable", "Sheet layout unavailable.")
	}
	if err != nil {
		return l, paper, err
	}
	err = json.Unmarshal(data, &l.Spec)
	return l, paper, err
}
func (s *Service) Sheet(ctx context.Context, tx pgx.Tx, actor, org, layout, assignment string) ([]byte, error) {
	l, paper, err := loadLayout(ctx, tx, org, layout)
	if err != nil {
		return nil, err
	}
	p, err := s.Paper(ctx, tx, actor, org, paper)
	if err != nil {
		return nil, err
	}
	var name string
	err = tx.QueryRow(ctx, "SELECT u.name FROM assignment a JOIN person u ON u.id=a.learner_id WHERE a.institution_id=$1 AND a.id=$2 AND a.paper_id=$3", org, assignment, paper).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, Problem(400, "assignment_unavailable", "Select an assignment for this paper.")
	}
	if err != nil {
		return nil, err
	}
	return academic.SheetImage(l.Spec, assignment, l.Fingerprint, p.Spec.Title, name, nil)
}

type Batch struct {
	ID        string    `json:"id"`
	PaperID   string    `json:"paper_id"`
	LayoutID  string    `json:"layout_id"`
	SittingAt time.Time `json:"sitting_at"`
}

func (s *Service) CreateBatch(ctx context.Context, tx pgx.Tx, actor, org, paper, layout string, at time.Time) (Batch, error) {
	_, err := s.Paper(ctx, tx, actor, org, paper)
	if err != nil {
		return Batch{}, err
	}
	_, layoutPaper, err := loadLayout(ctx, tx, org, layout)
	if err != nil {
		return Batch{}, err
	}
	if layoutPaper != paper || at.IsZero() || at.After(time.Now().Add(5*time.Minute)) {
		return Batch{}, Problem(400, "invalid_batch", "Choose the matching layout and the actual paper sitting time.")
	}
	out := Batch{access.Random(), paper, layout, at}
	_, err = tx.Exec(ctx, "INSERT INTO scan_batch(id,institution_id,paper_id,layout_id,sitting_at,created_by) VALUES($1,$2,$3,$4,$5,$6)", out.ID, org, paper, layout, at, actor)
	return out, err
}
func (s *Service) Batches(ctx context.Context, tx pgx.Tx, actor, org, paper string) ([]Batch, error) {
	if _, err := s.Paper(ctx, tx, actor, org, paper); err != nil {
		return nil, err
	}
	return rows[Batch](ctx, tx, "SELECT id,paper_id,layout_id,sitting_at FROM scan_batch WHERE institution_id=$1 AND paper_id=$2 ORDER BY created_at DESC LIMIT 100", org, paper)
}
func batch(ctx context.Context, tx pgx.Tx, org, id string) (Batch, error) {
	var b Batch
	err := tx.QueryRow(ctx, "SELECT id,paper_id,layout_id,sitting_at FROM scan_batch WHERE institution_id=$1 AND id=$2", org, id).Scan(&b.ID, &b.PaperID, &b.LayoutID, &b.SittingAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return b, Problem(404, "batch_unavailable", "Batch unavailable.")
	}
	return b, err
}
func (s *Service) UploadScan(ctx context.Context, tx pgx.Tx, actor, org, batchID, filename string, data []byte) (Job, error) {
	b, err := batch(ctx, tx, org, batchID)
	if err != nil {
		return Job{}, err
	}
	p, err := s.Paper(ctx, tx, actor, org, b.PaperID)
	if err != nil {
		return Job{}, err
	}
	digest := access.Digest(string(data))
	var existing string
	err = tx.QueryRow(ctx, "SELECT job_id FROM scan_source WHERE institution_id=$1 AND batch_id=$2 AND sha256=$3", org, batchID, digest).Scan(&existing)
	if err == nil {
		return s.Job(ctx, tx, actor, org, existing)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Job{}, err
	}
	asset, err := s.StoreAsset(ctx, tx, actor, org, p.CampusID, p.CourseID, "scan", filename, data)
	if err != nil {
		return Job{}, err
	}
	source, job := access.Random(), access.Random()
	_, err = tx.Exec(ctx, "INSERT INTO academic_job(id,institution_id,campus_id,course_id,actor_id,kind,resource_id,input) VALUES($1,$2,$3,$4,$5,'recognize_scan',$6,$7)", job, org, p.CampusID, p.CourseID, actor, source, Encode(map[string]string{"asset_id": asset.ID, "batch_id": batchID}))
	if err != nil {
		return Job{}, err
	}
	_, err = tx.Exec(ctx, "INSERT INTO scan_source(id,institution_id,batch_id,asset_id,job_id,sha256) VALUES($1,$2,$3,$4,$5,$6)", source, org, batchID, asset.ID, job, digest)
	if err != nil {
		return Job{}, err
	}
	if err = s.enqueue(ctx, tx, job); err != nil {
		return Job{}, err
	}
	return s.Job(ctx, tx, actor, org, job)
}

type ScanSheet struct {
	ID                string                     `json:"id"`
	AssetID           string                     `json:"asset_id"`
	AssignmentID      string                     `json:"assignment_id"`
	Page              int                        `json:"page"`
	Machine           academic.ScanRead          `json:"machine"`
	Corrections       map[string]academic.Answer `json:"corrections"`
	IdentityConfirmed bool                       `json:"identity_confirmed"`
	State             string                     `json:"state"`
	Version           int                        `json:"version"`
	SittingID         string                     `json:"sitting_id"`
}

func (s *Service) ScanSheets(ctx context.Context, tx pgx.Tx, actor, org, batchID string) ([]ScanSheet, error) {
	b, err := batch(ctx, tx, org, batchID)
	if err != nil {
		return nil, err
	}
	if _, err = s.Paper(ctx, tx, actor, org, b.PaperID); err != nil {
		return nil, err
	}
	type stored struct {
		ID, Asset, Assignment string
		Page                  int
		Machine, Corrections  []byte
		Confirmed             bool
		State                 string
		Version               int
		Sitting               string
	}
	list, err := rows[stored](ctx, tx, "SELECT id,asset_id,coalesce(assignment_id,''),page,machine,corrections,identity_confirmed,state,version,coalesce(committed_sitting,'') FROM scan_sheet WHERE institution_id=$1 AND batch_id=$2 ORDER BY source_id,page", org, batchID)
	if err != nil {
		return nil, err
	}
	out := []ScanSheet{}
	for _, v := range list {
		r := ScanSheet{ID: v.ID, AssetID: v.Asset, AssignmentID: v.Assignment, Page: v.Page, IdentityConfirmed: v.Confirmed, State: v.State, Version: v.Version, SittingID: v.Sitting}
		if err = json.Unmarshal(v.Machine, &r.Machine); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(v.Corrections, &r.Corrections); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}
func (s *Service) ReviewScan(ctx context.Context, tx pgx.Tx, actor, org, id, assignment, reason string, confirmed bool, version int, corrections map[string]academic.Answer, commit bool) (string, error) {
	var batchID, state string
	var actual int
	var machineData, oldData []byte
	var committed *string
	var previousAssignment *string
	var previousConfirmed bool
	err := tx.QueryRow(ctx, "SELECT batch_id,state,version,machine,corrections,committed_sitting,assignment_id,identity_confirmed FROM scan_sheet WHERE institution_id=$1 AND id=$2 FOR UPDATE", org, id).Scan(&batchID, &state, &actual, &machineData, &oldData, &committed, &previousAssignment, &previousConfirmed)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", Problem(404, "sheet_unavailable", "Scan sheet unavailable.")
	}
	if err != nil {
		return "", err
	}
	b, err := batch(ctx, tx, org, batchID)
	if err != nil {
		return "", err
	}
	p, err := s.Paper(ctx, tx, actor, org, b.PaperID)
	if err != nil {
		return "", err
	}
	if commit || confirmed {
		var allowed bool
		if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM membership WHERE institution_id=$1 AND person_id=$2 AND state='active' AND role IN ('owner','coordinator') AND (campus_id IS NULL OR campus_id=$3))", org, actor, p.CampusID).Scan(&allowed); err != nil {
			return "", err
		}
		if !allowed {
			return "", Problem(403, "coordinator_required", "A campus coordinator must confirm whole-sheet identity and commit paper evidence.")
		}
	}
	if state == "committed" && committed != nil {
		return *committed, nil
	}
	if actual != version {
		return "", Problem(409, "scan_conflict", "Another reviewer changed this sheet. Refresh first.")
	}
	if !validString(reason, 2000) {
		return "", Problem(400, "reason_required", "Record a review reason.")
	}
	var learner string
	var opens, closes time.Time
	err = tx.QueryRow(ctx, "SELECT learner_id,opens_at,closes_at FROM assignment WHERE institution_id=$1 AND id=$2 AND paper_id=$3 FOR UPDATE", org, assignment, b.PaperID).Scan(&learner, &opens, &closes)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", Problem(400, "invalid_identity", "Select an assigned learner for this paper.")
	}
	if err != nil {
		return "", err
	}
	if b.SittingAt.Before(opens) || b.SittingAt.After(closes) {
		return "", Problem(409, "paper_outside_window", "The recorded paper sitting is outside this learner's assigned window.")
	}
	if err = LearnerScope(ctx, tx, learner, org, p.CampusID, p.CourseID); err != nil {
		return "", err
	}
	var machine academic.ScanRead
	var old map[string]academic.Answer
	if err = json.Unmarshal(machineData, &machine); err != nil {
		return "", err
	}
	if err = json.Unmarshal(oldData, &old); err != nil {
		return "", err
	}
	if old == nil {
		old = map[string]academic.Answer{}
	}
	for qid, a := range corrections {
		q, ok := FindQuestion(p.Spec, qid)
		if !ok {
			return "", Problem(400, "unknown_question", "Correction references another paper.")
		}
		if !a.Invalid {
			if err = academic.ValidateAnswer(q, a); err != nil {
				return "", Invalid(err)
			}
		}
		previous := old[qid]
		if _, exists := old[qid]; !exists {
			previous = machine.Rows[qid].Answer
		}
		if !bytes.Equal(Encode(previous), Encode(a)) {
			if _, err = tx.Exec(ctx, "INSERT INTO scan_correction(id,institution_id,sheet_id,question_id,previous_value,new_value,reason,actor_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8)", access.Random(), org, id, qid, Encode(previous), Encode(a), reason, actor); err != nil {
				return "", err
			}
		}
		old[qid] = a
	}
	oldIdentity := ""
	if previousAssignment != nil {
		oldIdentity = *previousAssignment
	}
	if oldIdentity != assignment || previousConfirmed != confirmed {
		if _, err = tx.Exec(ctx, "INSERT INTO scan_correction(id,institution_id,sheet_id,question_id,previous_value,new_value,reason,actor_id) VALUES($1,$2,$3,'@identity',$4,$5,$6,$7)", access.Random(), org, id, Encode(map[string]any{"assignment_id": oldIdentity, "confirmed": previousConfirmed}), Encode(map[string]any{"assignment_id": assignment, "confirmed": confirmed}), reason, actor); err != nil {
			return "", err
		}
	}
	if _, err = tx.Exec(ctx, "UPDATE scan_sheet SET assignment_id=$2,corrections=$3,identity_confirmed=$4,version=version+1 WHERE id=$1", id, assignment, Encode(old), confirmed); err != nil {
		return "", err
	}
	if err = Event(ctx, tx, org, actor, "scan.reviewed", id, reason); err != nil {
		return "", err
	}
	if !commit {
		return "", nil
	}
	if !confirmed {
		return "", Problem(409, "identity_unconfirmed", "Confirm the learner identity before committing evidence.")
	}
	answers := map[string]academic.Answer{}
	for _, sec := range p.Spec.Sections {
		for _, q := range sec.Questions {
			a, corrected := old[q.ID]
			reading, present := machine.Rows[q.ID]
			if !corrected {
				if !present || reading.Uncertain || machine.Error != "" {
					return "", Problem(409, "uncertain_reading", "Explicitly review every uncertain or unreadable response before commit.")
				}
				a = reading.Answer
			}
			if !a.Invalid {
				if err = academic.ValidateAnswer(q, a); err != nil {
					return "", Invalid(err)
				}
			}
			answers[q.ID] = a
		}
	}
	var existing string
	err = tx.QueryRow(ctx, "SELECT id FROM sitting WHERE assignment_id=$1", assignment).Scan(&existing)
	if err == nil {
		return "", Problem(409, "evidence_exists", "This assignment already has a sitting. Review its evidence instead of creating a duplicate.")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	sitting := Sitting{ID: access.Random(), AssignmentID: assignment, LearnerID: learner, Channel: "paper", State: "active", Answers: answers, Manual: map[string]academic.ManualMark{}, Paper: p}
	_, err = tx.Exec(ctx, "INSERT INTO sitting(id,institution_id,assignment_id,learner_id,channel,state,started_at,deadline,answers,finish_reason) VALUES($1,$2,$3,$4,'paper','active',$5,$5,$6,'paper_commit')", sitting.ID, org, assignment, learner, b.SittingAt, Encode(answers))
	if err != nil {
		return "", err
	}
	if err = s.finalize(ctx, tx, org, &sitting, actor, "reviewed_paper_commit"); err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, "UPDATE scan_sheet SET state='committed',committed_sitting=$2 WHERE id=$1", id, sitting.ID)
	return sitting.ID, err
}
func rasterPages(ctx context.Context, mime string, data []byte) ([][]byte, error) {
	if mime != "application/pdf" {
		return [][]byte{data}, nil
	}
	dir, err := os.MkdirTemp("", "academic-pages-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	input := filepath.Join(dir, "source.pdf")
	if err = os.WriteFile(input, data, 0600); err != nil {
		return nil, err
	}
	info, err := exec.CommandContext(ctx, "pdfinfo", input).Output()
	if err != nil {
		return nil, errors.New("PDF cannot be inspected; remove encryption or upload page images")
	}
	pages := 0
	for _, line := range bytes.Split(info, []byte{'\n'}) {
		var n int
		if _, e := fmt.Sscanf(string(line), "Pages: %d", &n); e == nil {
			pages = n
		}
	}
	if pages < 1 || pages > 50 {
		return nil, errors.New("PDF must contain 1–50 pages; split larger batches")
	}
	out := [][]byte{}
	for page := 1; page <= pages; page++ {
		cmd := exec.CommandContext(ctx, "pdftoppm", "-f", fmt.Sprint(page), "-l", fmt.Sprint(page), "-scale-to", "2000", "-png", "-singlefile", input, filepath.Join(dir, "page"))
		if err = cmd.Run(); err != nil {
			return nil, errors.New("PDF rendering failed; retry or upload images")
		}
		data, err = os.ReadFile(filepath.Join(dir, "page.png"))
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}
	return out, nil
}
func (s *Service) recognize(ctx context.Context, j Job) error {
	var in map[string]string
	if err := json.Unmarshal(j.Input, &in); err != nil {
		return err
	}
	if in["mode"] == "grid_model" {
		return s.gridModel(ctx, j, in)
	}
	var a Asset
	var b Batch
	var l SheetLayout
	err := Tx(ctx, s.DB, func(tx pgx.Tx) error {
		var err error
		a, err = loadAsset(ctx, tx, j.Org, in["asset_id"])
		if err != nil {
			return err
		}
		b, err = batch(ctx, tx, j.Org, in["batch_id"])
		if err != nil {
			return err
		}
		l, _, err = loadLayout(ctx, tx, j.Org, b.LayoutID)
		return err
	})
	if err != nil {
		return err
	}
	data, err := s.Objects.GetAcademic(ctx, a.Key)
	if err != nil {
		return err
	}
	pages, err := rasterPages(ctx, a.Mime, data)
	if err != nil {
		return err
	}
	for index, data := range pages {
		if err = ctx.Err(); err != nil {
			return err
		}
		cfg, _, e := image.DecodeConfig(bytes.NewReader(data))
		if e != nil || int64(cfg.Width)*int64(cfg.Height) > 32_000_000 {
			return errors.New("scan page cannot be decoded within pixel limit")
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return err
		}
		reading := academic.Recognize(img, l.Spec, l.Fingerprint)
		var normalized bytes.Buffer
		if err = png.Encode(&normalized, img); err != nil {
			return err
		}
		err = Tx(ctx, s.DB, func(tx pgx.Tx) error {
			if err := s.jobAllowed(ctx, tx, j); err != nil {
				return err
			}
			var exists bool
			digest := access.Digest(normalized.String())
			if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM scan_sheet WHERE institution_id=$1 AND batch_id=$2 AND sha256=$3)", j.Org, b.ID, digest).Scan(&exists); err != nil {
				return err
			}
			if exists {
				return nil
			}
			asset, err := s.StoreAsset(ctx, tx, j.Actor, j.Org, a.CampusID, a.CourseID, "scan_page", fmt.Sprintf("page-%d.png", index+1), normalized.Bytes())
			if err != nil {
				return err
			}
			var assignment any
			if reading.AssignmentID != "" {
				var valid bool
				if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM assignment WHERE institution_id=$1 AND id=$2 AND paper_id=$3)", j.Org, reading.AssignmentID, b.PaperID).Scan(&valid); err != nil {
					return err
				}
				if valid {
					assignment = reading.AssignmentID
				} else {
					reading.Error = "Machine identity does not belong to this paper."
				}
			}
			_, err = tx.Exec(ctx, "INSERT INTO scan_sheet(id,institution_id,batch_id,source_id,asset_id,assignment_id,page,sha256,machine,state) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,'review') ON CONFLICT DO NOTHING", access.Random(), j.Org, b.ID, j.Resource, asset.ID, assignment, index+1, digest, Encode(reading))
			if err != nil {
				return err
			}
			_, err = tx.Exec(ctx, "UPDATE academic_job SET progress=$2,updated_at=now() WHERE id=$1", j.ID, (index+1)*95/len(pages))
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}
