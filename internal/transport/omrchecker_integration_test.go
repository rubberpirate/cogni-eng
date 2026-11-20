//go:build integration

package transport

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/scholar"
)

func TestOMRCheckerReadsNativeTemplateWithinStaffScopeWithoutMarks(t *testing.T) {
	if os.Getenv("RUN_OMRCHECKER_TESTS") != "1" {
		t.Skip("opt-in installed OMRChecker runtime")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("OMRCHECKER_PYTHON", filepath.Join(root, "omr", ".venv", "bin", "python"))
	t.Setenv("OMRCHECKER_ADAPTER", filepath.Join(root, "omr", "bridge.py"))
	h, svc, course, learner := academicFixture(t)
	sample := filepath.Join(root, "apps", "web", "public", "samples", "omrchecker")
	image, err := os.ReadFile(filepath.Join(sample, "filled-sheet.png"))
	if err != nil {
		t.Fatal(err)
	}
	template, err := os.ReadFile(filepath.Join(sample, "template.json"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	academicTx(t, h, func(tx pgx.Tx) error {
		out, e := svc.OMRCheckerPreview(ctx, tx, h.owner, h.org, course, image, nil, nil, json.RawMessage(template), nil)
		if e != nil {
			return e
		}
		if out.Engine != "OMRChecker" || len(out.Revision) != 40 || len(out.Fields) != 26 || out.Responses["Roll"] != "123456" || out.Responses["q1"] != "A" || out.Responses["q7"] != "" || out.Responses["q12"] != "AC" || !out.MultipleMarked {
			t.Fatalf("unexpected native reading: engine=%s fields=%d", out.Engine, len(out.Fields))
		}
		return nil
	})
	for _, tc := range []struct{ name, actor, org string }{{name: "learner", actor: learner, org: h.org}, {name: "other institution", actor: h.owner, org: h.otherOrg}} {
		t.Run(tc.name, func(t *testing.T) {
			if err := scholar.Tx(ctx, h.db, func(tx pgx.Tx) error {
				_, e := svc.OMRCheckerPreview(ctx, tx, tc.actor, tc.org, course, image, nil, nil, json.RawMessage(template), nil)
				return e
			}); err == nil {
				t.Fatal("unauthorized native preview accepted")
			}
		})
	}
	var count int
	if err := h.db.QueryRow(ctx, "SELECT count(*) FROM result_revision").Scan(&count); err != nil || count != 0 {
		t.Fatalf("preview changed marks: %d %v", count, err)
	}
}
