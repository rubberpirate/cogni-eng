package scholar

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (s *Service) extractVision(ctx context.Context, j Job, in map[string]string, pages [][]byte) error {
	var raw strings.Builder
	var previous struct {
		Pages []VisionOutput `json:"pages"`
	}
	if err := json.Unmarshal(j.Output, &previous); err != nil {
		return err
	}
	observations := []VisionOutput{}
	for i, page := range pages {
		// Recheck revocation/cancellation before each expensive inference call.
		if err := Tx(ctx, s.DB, func(tx pgx.Tx) error { return s.jobAllowed(ctx, tx, j) }); err != nil {
			return err
		}
		var observation VisionOutput
		var err error
		if i < len(previous.Pages) {
			observation = previous.Pages[i]
		} else {
			observation, err = VisionFromEnv().Read(ctx, page, LightOnModel, "")
		}
		if err != nil {
			return err
		}
		if raw.Len()+len(observation.Text) > 1<<20 {
			return errors.New("model OCR output exceeds 1 MiB; split the document")
		}
		raw.WriteString(observation.Text)
		raw.WriteByte('\n')
		observations = append(observations, observation)
		if err = Tx(ctx, s.DB, func(tx pgx.Tx) error {
			if e := s.jobAllowed(ctx, tx, j); e != nil {
				return e
			}
			_, e := tx.Exec(ctx, "UPDATE academic_job SET progress=$2,provider='lightonocr',model=$3,output=$4,updated_at=now() WHERE id=$1", j.ID, (i+1)*90/len(pages), observation.Model, Encode(map[string]any{"pages": observations}))
			return e
		}); err != nil {
			return err
		}
	}
	if in["mode"] == "answer_key" {
		return s.finishKey(ctx, j, raw.String())
	}
	return s.finishDocument(ctx, j, in, raw.String(), raw.String())
}
