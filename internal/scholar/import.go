package scholar

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
)

type Import struct {
	ID        string               `json:"id"`
	TopicID   string               `json:"topic_id"`
	AssetID   string               `json:"asset_id"`
	State     string               `json:"state"`
	Rows      []academic.ImportRow `json:"rows"`
	RawText   string               `json:"raw_text"`
	Version   int                  `json:"version"`
	CreatedAt time.Time            `json:"created_at"`
}

func (s *Service) Import(ctx context.Context, tx pgx.Tx, actor, org, id string) (Import, error) {
	var out Import
	var data []byte
	err := tx.QueryRow(ctx, "SELECT id,topic_id,coalesce(asset_id,''),state,rows,raw_text,version,created_at FROM question_import WHERE institution_id=$1 AND id=$2", org, id).Scan(&out.ID, &out.TopicID, &out.AssetID, &out.State, &data, &out.RawText, &out.Version, &out.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, Problem(404, "import_not_found", "Import unavailable.")
	}
	if err != nil {
		return out, err
	}
	t, err := GetTopic(ctx, tx, org, out.TopicID)
	if err != nil {
		return out, err
	}
	if err = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err != nil {
		return out, err
	}
	err = json.Unmarshal(data, &out.Rows)
	return out, err
}
func (s *Service) Imports(ctx context.Context, tx pgx.Tx, actor, org, topic string) ([]Import, error) {
	t, err := GetTopic(ctx, tx, org, topic)
	if err != nil {
		return nil, err
	}
	if err = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err != nil {
		return nil, err
	}
	type item struct{ ID string }
	ids, err := rows[item](ctx, tx, "SELECT id FROM question_import WHERE institution_id=$1 AND topic_id=$2 ORDER BY created_at DESC LIMIT 30", org, topic)
	if err != nil {
		return nil, err
	}
	out := []Import{}
	for _, id := range ids {
		v, e := s.Import(ctx, tx, actor, org, id.ID)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, nil
}
func (s *Service) StageImport(ctx context.Context, tx pgx.Tx, actor, org, topic, filename string, data []byte, provenance, license string) (Import, error) {
	t, err := GetTopic(ctx, tx, org, topic)
	if err != nil {
		return Import{}, err
	}
	if err = Scope(ctx, tx, actor, org, t.CampusID, t.CourseID, t.Subject); err != nil {
		return Import{}, err
	}
	fp := access.Digest(string(data))
	var existing string
	err = tx.QueryRow(ctx, "SELECT id FROM question_import WHERE institution_id=$1 AND topic_id=$2 AND fingerprint=$3", org, topic, fp).Scan(&existing)
	if err == nil {
		return s.Import(ctx, tx, actor, org, existing)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Import{}, err
	}
	parsed, err := academic.ParseImport(filename, data, topic, provenance, license)
	if err != nil {
		return Import{}, Invalid(err)
	}
	for i := range parsed {
		var duplicate string
		e := tx.QueryRow(ctx, "SELECT latest_revision FROM question_record WHERE institution_id=$1 AND fingerprint=$2 AND topic_id=$3", org, fingerprint(parsed[i].Question), topic).Scan(&duplicate)
		if e == nil {
			parsed[i].ExistingID = duplicate
		} else if !errors.Is(e, pgx.ErrNoRows) {
			return Import{}, e
		}
	}
	id := access.Random()
	_, err = tx.Exec(ctx, "INSERT INTO question_import(id,institution_id,topic_id,fingerprint,state,rows,created_by) VALUES($1,$2,$3,$4,'preview',$5,$6)", id, org, topic, fp, Encode(parsed), actor)
	if err != nil {
		return Import{}, err
	}
	if err = Event(ctx, tx, org, actor, "questions.import_previewed", id, filename); err != nil {
		return Import{}, err
	}
	return s.Import(ctx, tx, actor, org, id)
}
func (s *Service) EditImport(ctx context.Context, tx pgx.Tx, actor, org, id string, index, version int, q academic.Question) (Import, error) {
	out, err := s.Import(ctx, tx, actor, org, id)
	if err != nil {
		return out, err
	}
	if out.Version != version || out.State != "preview" {
		return out, Problem(409, "import_changed", "Import changed or was committed. Refresh before editing.")
	}
	found := false
	for i, r := range out.Rows {
		if r.Index == index {
			if r.CommittedID != "" {
				return out, Problem(409, "row_committed", "Revise committed questions in the question bank; original import evidence is retained.")
			}
			found = true
			q.TopicID = out.TopicID
			out.Rows[i].Question = q
			out.Rows[i].Error = ""
			if e := academic.ValidateQuestion(q); e != nil {
				out.Rows[i].Error = e.Error()
			}
		}
	}
	if !found {
		return out, Problem(404, "row_not_found", "Import row unavailable.")
	}
	tag, err := tx.Exec(ctx, "UPDATE question_import SET rows=$3,version=version+1 WHERE institution_id=$1 AND id=$2 AND version=$4", org, id, Encode(out.Rows), version)
	if err != nil {
		return out, err
	}
	if tag.RowsAffected() == 0 {
		return out, Problem(409, "version_conflict", "Another reviewer edited the import.")
	}
	return s.Import(ctx, tx, actor, org, id)
}
func (s *Service) CommitImport(ctx context.Context, tx pgx.Tx, actor, org, id string, selected []int, version int) (Import, error) {
	if _, err := tx.Exec(ctx, "SELECT id FROM question_import WHERE institution_id=$1 AND id=$2 FOR UPDATE", org, id); err != nil {
		return Import{}, err
	}
	out, err := s.Import(ctx, tx, actor, org, id)
	if err != nil {
		return out, err
	}
	if out.State == "committed" {
		return out, nil
	}
	if out.State != "preview" || out.Version != version {
		return out, Problem(409, "import_changed", "Refresh the current import preview before commitment.")
	}
	indices := map[int]bool{}
	for _, i := range selected {
		indices[i] = true
	}
	if len(indices) == 0 {
		return out, Problem(400, "selection_required", "Select at least one clean question row.")
	}
	for i, row := range out.Rows {
		if !indices[row.Index] {
			continue
		}
		delete(indices, row.Index)
		if row.CommittedID != "" {
			continue
		}
		if row.Error != "" {
			return out, Problem(400, "invalid_import_row", "Correct selected row errors before committing.")
		}
		q, duplicate, e := s.CreateQuestion(ctx, tx, actor, org, row.Question, "imported")
		if e != nil {
			return out, e
		}
		if duplicate {
			out.Rows[i].ExistingID = q.ID
		}
		out.Rows[i].CommittedID = q.ID
	}
	if len(indices) > 0 {
		return out, Problem(400, "invalid_selection", "Selection contains an unknown row.")
	}
	state := "committed"
	for _, row := range out.Rows {
		if row.CommittedID == "" {
			state = "preview"
		}
	}
	if _, err = tx.Exec(ctx, "UPDATE question_import SET state=$4,rows=$3,version=version+1 WHERE institution_id=$1 AND id=$2", org, id, Encode(out.Rows), state); err != nil {
		return out, err
	}
	if err = Event(ctx, tx, org, actor, "questions.import_committed", id, "Drafts require academic approval"); err != nil {
		return out, err
	}
	return s.Import(ctx, tx, actor, org, id)
}
