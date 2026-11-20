package scholar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
)

func (s *Service) StageDocument(ctx context.Context, tx pgx.Tx, actor, org, topic, filename string, data []byte, provenance, license, engine string) (Import, error) {
	t, err := GetTopic(ctx, tx, org, topic)
	if err != nil {
		return Import{}, err
	}
	if err = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err != nil {
		return Import{}, err
	}
	if engine == "" {
		engine = os.Getenv("OCR_PROVIDER")
	}
	if engine == "" {
		engine = "tesseract"
	}
	if engine != "tesseract" && engine != "lightonocr" {
		return Import{}, Problem(400, "invalid_ocr_engine", "Choose Tesseract or LightOnOCR.")
	}
	digest := access.Digest(string(data) + "\x00" + engine)
	var existing string
	err = tx.QueryRow(ctx, "SELECT id FROM question_import WHERE institution_id=$1 AND topic_id=$2 AND fingerprint=$3", org, topic, digest).Scan(&existing)
	if err == nil {
		return s.Import(ctx, tx, actor, org, existing)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Import{}, err
	}
	if !validString(provenance, 2000) || !validString(license, 500) {
		return Import{}, Problem(400, "source_required", "Provide the document source and usage rights.")
	}
	asset, err := s.StoreAsset(ctx, tx, actor, org, t.CampusID, t.CourseID, "document", filename, data)
	if err != nil {
		return Import{}, err
	}
	id, job := access.Random(), access.Random()
	_, err = tx.Exec(ctx, "INSERT INTO question_import(id,institution_id,topic_id,asset_id,fingerprint,state,created_by) VALUES($1,$2,$3,$4,$5,'extracting',$6)", id, org, topic, asset.ID, digest, actor)
	if err != nil {
		return Import{}, err
	}
	input := map[string]string{"asset_id": asset.ID, "topic_id": topic, "provenance": provenance, "license": license, "language": t.Language, "provider": engine}
	_, err = tx.Exec(ctx, "INSERT INTO academic_job(id,institution_id,campus_id,course_id,actor_id,kind,resource_id,input) VALUES($1,$2,$3,$4,$5,'extract_document',$6,$7)", job, org, t.CampusID, t.CourseID, actor, id, Encode(input))
	if err != nil {
		return Import{}, err
	}
	if err = s.enqueue(ctx, tx, job); err != nil {
		return Import{}, err
	}
	return s.Import(ctx, tx, actor, org, id)
}
func (s *Service) extract(ctx context.Context, j Job) error {
	var in map[string]string
	if err := json.Unmarshal(j.Input, &in); err != nil {
		return err
	}
	var cached struct {
		OCR    string `json:"ocr_text"`
		Raw    string `json:"raw_ocr_text"`
		Method string `json:"text_method"`
	}
	if in["mode"] != "answer_key" && json.Unmarshal(j.Output, &cached) == nil && cached.OCR != "" && cached.Method == "embedded_pdf" {
		in["text_method"] = cached.Method
		return s.finishDocument(ctx, j, in, cached.Raw, cached.OCR)
	}
	if in["mode"] == "answer_key" && in["text"] != "" {
		return s.finishKey(ctx, j, in["text"])
	}
	var asset Asset
	err := Tx(ctx, s.DB, func(tx pgx.Tx) error {
		var err error
		asset, err = loadAsset(ctx, tx, j.Org, in["asset_id"])
		return err
	})
	if err != nil {
		return err
	}
	data, err := s.Objects.GetAcademic(ctx, asset.Key)
	if err != nil {
		return err
	}
	if asset.Mime == "application/pdf" && in["mode"] != "answer_key" {
		text, usable, e := embeddedPDFText(ctx, data)
		if e != nil {
			return e
		}
		if usable {
			in["text_method"] = "embedded_pdf"
			return s.finishDocument(ctx, j, in, text, text)
		}
	}
	if in["mode"] != "answer_key" && cached.OCR != "" {
		in["text_method"] = cached.Method
		return s.finishDocument(ctx, j, in, cached.Raw, cached.OCR)
	}
	pages, err := rasterPages(ctx, asset.Mime, data)
	if err != nil {
		return err
	}
	if in["provider"] == "lightonocr" {
		return s.extractVision(ctx, j, in, pages)
	}
	dir, err := os.MkdirTemp("", "academic-ocr-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	var text, rawText strings.Builder
	language := "eng"
	if in["language"] == "hi" {
		language = "hin+eng"
	}
	for index, page := range pages {
		input := filepath.Join(dir, "page.png")
		if err = os.WriteFile(input, page, 0600); err != nil {
			return err
		}
		output, err := exec.CommandContext(ctx, "tesseract", input, "stdout", "-l", language, "--psm", "3", "tsv").Output()
		if err != nil {
			return errors.New("OCR processing failed. Confirm Tesseract and the chosen language are installed, or upload a reviewed CSV/JSON transcription")
		}
		if text.Len()+len(output) > 1<<20 {
			return errors.New("extracted text exceeds 1 MiB; split the document")
		}
		cfg, _, decodeErr := image.DecodeConfig(bytes.NewReader(page))
		if decodeErr != nil {
			return decodeErr
		}
		fmt.Fprintf(&rawText, "\n%s\n", academic.RawOCRText(string(output)))
		fmt.Fprintf(&text, "\n%s\n", academic.OrderedOCRText(string(output), cfg.Width))
		if err = Tx(ctx, s.DB, func(tx pgx.Tx) error {
			if err := s.jobAllowed(ctx, tx, j); err != nil {
				return err
			}
			_, err := tx.Exec(ctx, "UPDATE academic_job SET progress=$2,provider='tesseract',model='local-ocr',updated_at=now() WHERE id=$1", j.ID, (index+1)*90/len(pages))
			return err
		}); err != nil {
			return err
		}
	}
	if in["mode"] == "answer_key" {
		return s.finishKey(ctx, j, text.String())
	}
	return s.finishDocument(ctx, j, in, rawText.String(), text.String())
}
func (s *Service) finishDocument(ctx context.Context, j Job, in map[string]string, raw, text string) error {
	parsed, err := academic.ParseText(text, in["topic_id"], in["provenance"], in["license"])
	if err != nil {
		parsed = []academic.ImportRow{}
	}
	parsed = academic.RequireFacultyKeys(parsed)
	if os.Getenv("ACADEMIC_MODEL_PROVIDER") != "development_fixture" || (os.Getenv("APP_ENV") != "development" && os.Getenv("APP_ENV") != "test") {
		if err = Tx(ctx, s.DB, func(tx pgx.Tx) error {
			if e := s.jobAllowed(ctx, tx, j); e != nil {
				return e
			}
			if _, e := tx.Exec(ctx, "UPDATE question_import SET raw_text=$2 WHERE id=$1 AND state='extracting'", j.Resource, raw); e != nil {
				return e
			}
			_, e := tx.Exec(ctx, "UPDATE academic_job SET output=output || $2::jsonb WHERE id=$1", j.ID, Encode(map[string]string{"ocr_text": text, "raw_ocr_text": raw, "text_method": in["text_method"]}))
			return e
		}); err != nil {
			return err
		}
		parsed, err = s.normalizeLocalDocument(ctx, j, in, text)
		if err != nil {
			return err
		}
	}
	return Tx(ctx, s.DB, func(tx pgx.Tx) error {
		if e := s.jobAllowed(ctx, tx, j); e != nil {
			return e
		}
		_, e := tx.Exec(ctx, "UPDATE question_import SET raw_text=$2,rows=$3,state='preview',version=version+1 WHERE id=$1 AND state='extracting'", j.Resource, raw, Encode(parsed))
		return e
	})
}

func (s *Service) ParseDocument(ctx context.Context, tx pgx.Tx, actor, org, id, text, provenance, license string, version int) (Import, error) {
	out, err := s.Import(ctx, tx, actor, org, id)
	if err != nil {
		return out, err
	}
	if out.State != "preview" || out.Version != version || len(text) > 1<<20 {
		return out, Problem(409, "import_changed", "Refresh a completed extraction before mapping the reviewed text.")
	}
	for _, r := range out.Rows {
		if r.CommittedID != "" {
			return out, Problem(409, "import_partially_committed", "Correct the remaining rows individually; committed evidence is retained.")
		}
	}
	parsed, err := academic.ParseText(text, out.TopicID, provenance, license)
	if err != nil {
		return out, Invalid(err)
	}
	tag, err := tx.Exec(ctx, "UPDATE question_import SET rows=$3,version=version+1 WHERE institution_id=$1 AND id=$2 AND version=$4", org, id, Encode(parsed), version)
	if err != nil {
		return out, err
	}
	if tag.RowsAffected() == 0 {
		return out, Problem(409, "import_changed", "Another reviewer changed this import.")
	}
	if err = Event(ctx, tx, org, actor, "document.transcription_mapped", id, "Original extracted text and private source retained; keys still require approval"); err != nil {
		return out, err
	}
	return s.Import(ctx, tx, actor, org, id)
}
