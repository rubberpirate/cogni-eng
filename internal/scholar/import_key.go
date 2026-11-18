package scholar

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
)

// StageKey retains the source in a durable job; it does not replace any reviewed key.
func (s *Service) StageKey(ctx context.Context, tx pgx.Tx, actor, org, id, filename string, data []byte) (Job, error) {
	imp, err := s.Import(ctx, tx, actor, org, id)
	if err != nil {
		return Job{}, err
	}
	if imp.State != "preview" {
		return Job{}, Problem(409, "import_changed", "Choose an uncommitted import before attaching a key.")
	}
	t, err := GetTopic(ctx, tx, org, imp.TopicID)
	if err != nil {
		return Job{}, err
	}
	if len(data) == 0 || len(data) > 8<<20 {
		return Job{}, Problem(400, "invalid_key", "Use a key file of at most 8 MiB.")
	}
	input := map[string]string{"mode": "answer_key", "topic_id": t.ID, "filename": filename, "sha256": access.Digest(string(data))}
	mime := http.DetectContentType(data)
	if mime == "image/png" || mime == "image/jpeg" || mime == "application/pdf" {
		asset, e := s.StoreAsset(ctx, tx, actor, org, t.CampusID, t.CourseID, "document", filename, data)
		if e != nil {
			return Job{}, e
		}
		input["asset_id"] = asset.ID
		input["language"] = t.Language
	} else {
		if _, e := academic.ParseAnswerKey(string(data)); e != nil {
			return Job{}, Invalid(e)
		}
		input["text"] = string(data)
	}
	job := access.Random()
	_, err = tx.Exec(ctx, "INSERT INTO academic_job(id,institution_id,campus_id,course_id,actor_id,kind,resource_id,input) VALUES($1,$2,$3,$4,$5,'extract_document',$6,$7)", job, org, t.CampusID, t.CourseID, actor, id, Encode(input))
	if err != nil {
		return Job{}, err
	}
	if err = s.enqueue(ctx, tx, job); err != nil {
		return Job{}, err
	}
	if err = Event(ctx, tx, org, actor, "answer_key.uploaded", id, "Source retained in job "+job); err != nil {
		return Job{}, err
	}
	return s.Job(ctx, tx, actor, org, job)
}
func (s *Service) finishKey(ctx context.Context, j Job, text string) error {
	preview, err := academic.ParseAnswerKey(strings.TrimSpace(text))
	if err != nil {
		return Invalid(err)
	}
	return Tx(ctx, s.DB, func(tx pgx.Tx) error {
		if e := s.jobAllowed(ctx, tx, j); e != nil {
			return e
		}
		_, e := tx.Exec(ctx, "UPDATE academic_job SET output=$2 WHERE id=$1", j.ID, Encode(preview))
		return e
	})
}
func (s *Service) ApplyKey(ctx context.Context, tx pgx.Tx, actor, org, id, jobID string, version int) (Import, error) {
	imp, err := s.Import(ctx, tx, actor, org, id)
	if err != nil {
		return imp, err
	}
	if imp.State != "preview" || imp.Version != version {
		return imp, Problem(409, "import_changed", "Refresh the current import before applying the reviewed key.")
	}
	job, err := s.Job(ctx, tx, actor, org, jobID)
	if err != nil {
		return imp, err
	}
	var input map[string]string
	var key academic.KeyPreview
	if json.Unmarshal(job.Input, &input) != nil || job.Resource != id || input["mode"] != "answer_key" || job.State != "completed" || json.Unmarshal(job.Output, &key) != nil || len(key.Entries) == 0 {
		return imp, Problem(400, "key_not_ready", "Choose a completed answer-key extraction for this import.")
	}
	keys := map[int][]string{}
	for _, entry := range key.Entries {
		keys[entry.Number] = entry.Selection
	}
	for i, row := range imp.Rows {
		n := row.SourceNumber
		if n == 0 {
			n = row.Index
		}
		selection, ok := keys[n]
		if !ok {
			continue
		}
		if row.CommittedID != "" {
			return imp, Problem(409, "row_committed", "This key includes an already committed question. Revise that question in the bank.")
		}
		if row.Question.Type != "single" && row.Question.Type != "multiple" {
			return imp, Problem(400, "key_type", "Numbered selection keys only apply to choice questions.")
		}
		q := row.Question
		q.Key = selection
		// Validate the key independently; other extraction errors remain editable.
		allowed := map[string]bool{}
		for _, o := range q.Options {
			allowed[o.ID] = true
		}
		for _, label := range selection {
			if !allowed[label] {
				return imp, Problem(400, "key_option", "The key names an option absent from its question.")
			}
		}
		if q.Type == "single" && len(selection) != 1 {
			return imp, Problem(400, "key_option", "A single-choice question needs one answer.")
		}
		imp.Rows[i].Question = q
		imp.Rows[i].Error = ""
		if e := academic.ValidateQuestion(q); e != nil {
			imp.Rows[i].Error = e.Error()
		}
		delete(keys, n)
	}
	if len(keys) > 0 {
		return imp, Problem(400, "key_number", "The key contains question numbers not present in this import.")
	}
	tag, err := tx.Exec(ctx, "UPDATE question_import SET rows=$3,version=version+1 WHERE institution_id=$1 AND id=$2 AND version=$4", org, id, Encode(imp.Rows), version)
	if err != nil {
		return imp, err
	}
	if tag.RowsAffected() == 0 {
		return imp, Problem(409, "import_changed", "Another reviewer changed this import.")
	}
	if err = Event(ctx, tx, org, actor, "answer_key.applied", id, "Faculty confirmed key job "+jobID); err != nil {
		return imp, err
	}
	return s.Import(ctx, tx, actor, org, id)
}
