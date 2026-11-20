package scholar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
)

type OMRIdentityRegion struct {
	Kind string            `json:"kind"`
	Rect academic.GridRect `json:"rect"`
}
type OMRIdentityReading struct {
	Kind       string  `json:"kind"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Uncertain  bool    `json:"uncertain"`
	Reason     string  `json:"reason"`
}

type OMRCheckerField struct {
	Label     string   `json:"label"`
	Value     string   `json:"value"`
	Selection []string `json:"selection"`
	Uncertain bool     `json:"uncertain"`
	Reason    string   `json:"reason"`
}
type OMRCheckerReading struct {
	Engine         string               `json:"engine"`
	Revision       string               `json:"revision"`
	Responses      map[string]string    `json:"responses"`
	Fields         []OMRCheckerField    `json:"fields"`
	MultipleMarked bool                 `json:"multiple_marked"`
	Overlay        string               `json:"overlay"`
	Notice         string               `json:"notice"`
	Identity       []OMRIdentityReading `json:"identity,omitempty"`
}

// OMRCheckerPreview reads a native template without assigning identity or saving marks.
func (s *Service) OMRCheckerPreview(ctx context.Context, tx pgx.Tx, actor, org, course string, data, marker, reference []byte, template, config json.RawMessage, identity []OMRIdentityRegion) (OMRCheckerReading, error) {
	var campus string
	if err := tx.QueryRow(ctx, "SELECT campus_id FROM course WHERE institution_id=$1 AND id=$2 AND state='active'", org, course).Scan(&campus); err != nil {
		return OMRCheckerReading{}, Problem(404, "course_unavailable", "Choose an available course.")
	}
	if err := Scope(ctx, tx, actor, org, campus, course, ""); err != nil {
		return OMRCheckerReading{}, err
	}
	return runOMRChecker(ctx, data, marker, reference, template, config, identity)
}

func runOMRChecker(ctx context.Context, data, marker, reference []byte, template, config json.RawMessage, identity []OMRIdentityRegion) (OMRCheckerReading, error) {
	var out OMRCheckerReading
	if err := ctx.Err(); err != nil {
		return out, err
	}
	if err := validateOMRIdentity(identity); err != nil {
		return out, err
	}
	if len(template) == 0 || len(template) > 256<<10 || !json.Valid(template) || len(config) > 64<<10 || len(config) > 0 && !json.Valid(config) {
		return out, Problem(400, "invalid_omr_template", "Provide a template JSON of at most 256 KiB and optional tuning JSON of at most 64 KiB.")
	}
	for _, source := range [][]byte{data, marker, reference} {
		if len(source) == 0 {
			continue
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(source))
		if err != nil || (format != "png" && format != "jpeg") || len(source) > 8<<20 || cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > 8_000_000 {
			return out, Problem(400, "invalid_omr_image", "Use PNG/JPEG sheet, marker and reference images of at most 8 MiB and 8 megapixels each.")
		}
	}
	if len(data) == 0 {
		return out, Problem(400, "file_required", "Choose a filled sheet image.")
	}
	python := os.Getenv("OMRCHECKER_PYTHON")
	if python == "" {
		python = "omr/.venv/bin/python"
	}
	python, err := filepath.Abs(python)
	if err != nil {
		return out, err
	}
	adapterPath := os.Getenv("OMRCHECKER_ADAPTER")
	if adapterPath == "" {
		adapterPath = "omr/bridge.py"
	}
	adapter, err := filepath.Abs(adapterPath)
	if err != nil {
		return out, err
	}
	root := os.Getenv("OMRCHECKER_ROOT")
	if root == "" {
		root = filepath.Join(filepath.Dir(adapter), "engine")
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return out, err
	}
	for _, path := range []string{filepath.Join(root, "src", "core.py"), python, adapter} {
		if _, err := os.Stat(path); err != nil {
			return out, Problem(503, "omrchecker_configuration_required", "OMRChecker is not configured. Restore the bundled omr/engine files and run bun run omr:setup.")
		}
	}
	directory, err := os.MkdirTemp("", "academic-omrchecker-")
	if err != nil {
		return out, err
	}
	defer func() { _ = os.RemoveAll(directory) }()
	for name, source := range map[string][]byte{"sheet.png": data, "marker.png": marker, "reference.png": reference} {
		if len(source) > 0 {
			if err := os.WriteFile(filepath.Join(directory, name), source, 0600); err != nil {
				return out, err
			}
		}
	}
	if len(config) == 0 {
		config = json.RawMessage(`{}`)
	}
	input := Encode(map[string]json.RawMessage{"template": template, "config": config, "identity_regions": Encode(identity)})
	runCtx, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()
	cmd := exec.CommandContext(runCtx, python, "-I", "-B", adapter, "--root", root, "--image", filepath.Join(directory, "sheet.png"))
	cmd.Dir = directory
	// Do not give the recognition process application credentials or provider keys.
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + directory, "TMPDIR=" + directory, "MPLBACKEND=Agg", "MPLCONFIGDIR=" + filepath.Join(directory, "matplotlib"), "OPENBLAS_NUM_THREADS=1", "OMP_NUM_THREADS=1", "NO_COLOR=1"}
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stderr = io.Discard
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return out, err
	}
	if err = cmd.Start(); err != nil {
		return out, Problem(503, "omrchecker_unavailable", "Could not start the OMRChecker runtime. Run bun run omr:setup.")
	}
	response, readErr := io.ReadAll(io.LimitReader(stdout, (6<<20)+1))
	if readErr != nil || len(response) > 6<<20 {
		cancel()
	}
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return out, ctx.Err()
	}
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return out, Problem(503, "omrchecker_timeout", "OMRChecker exceeded forty seconds. Retry with a smaller image or simpler template.")
	}
	if readErr != nil {
		return out, readErr
	}
	if len(response) > 6<<20 {
		return out, Problem(503, "omrchecker_output_limit", "OMRChecker response exceeded the preview limit.")
	}
	var failure struct {
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(response, &failure) == nil && failure.Error != nil {
		if failure.Error.Code == "invalid_omr_template" && len(failure.Error.Message) <= 500 {
			return out, Problem(400, failure.Error.Code, failure.Error.Message)
		}
		return out, Problem(503, "omrchecker_failed", "OMRChecker could not process this sheet. Check its template and uploaded marker/reference images.")
	}
	if waitErr != nil || json.Unmarshal(response, &out) != nil || out.Engine != "OMRChecker" || len(out.Revision) != 40 || len(out.Fields) < 1 || len(out.Fields) > 500 || len(out.Responses) > 1000 || !strings.HasPrefix(out.Overlay, "/9j/") {
		return out, Problem(503, "omrchecker_invalid_output", "OMRChecker did not return a complete reading. Check the installed revision and template.")
	}
	out.Notice = "OMRChecker template reading. Confirm uncertain fields and student identity before grading; this preview creates no marks."
	return out, nil
}
