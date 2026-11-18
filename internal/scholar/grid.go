package scholar

import (
	"bytes"
	"context"
	"image"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
)

// GridPreview is an authorized diagnostic. No learner identity, mark or sitting is inferred.
func (s *Service) GridPreview(ctx context.Context, tx pgx.Tx, actor, org, course string, data []byte, c academic.GridCalibration) (academic.GridReading, error) {
	var campus string
	if err := tx.QueryRow(ctx, "SELECT campus_id FROM course WHERE institution_id=$1 AND id=$2 AND state='active'", org, course).Scan(&campus); err != nil {
		return academic.GridReading{}, Problem(404, "course_unavailable", "Choose an available course.")
	}
	if err := Scope(ctx, tx, actor, org, campus, course, ""); err != nil {
		return academic.GridReading{}, err
	}
	if len(data) > 8<<20 {
		return academic.GridReading{}, Problem(400, "image_limit", "Use a PNG or JPEG of at most 8 MiB and 8 megapixels.")
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || int64(cfg.Width)*int64(cfg.Height) > 8_000_000 {
		return academic.GridReading{}, Problem(400, "image_limit", "Use a decodable PNG or JPEG of at most 8 megapixels.")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return academic.GridReading{}, Invalid(err)
	}
	out, err := academic.ReadGrid(img, c)
	if err != nil {
		return out, Invalid(err)
	}
	return out, nil
}
