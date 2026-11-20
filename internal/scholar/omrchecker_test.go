package scholar

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"

	"learning.local/workspace/internal/academic"
)

func TestBundledOMRCheckerReadsSampleWithoutExternalCheckoutOrGit(t *testing.T) {
	if os.Getenv("RUN_OMRCHECKER_TESTS") != "1" {
		t.Skip("set RUN_OMRCHECKER_TESTS=1 after installing the frozen OMR runtime")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	t.Setenv("OMRCHECKER_ROOT", "")
	t.Setenv("OMRCHECKER_PYTHON", "")
	t.Setenv("OMRCHECKER_ADAPTER", "")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	sample := filepath.Join("apps", "web", "public", "samples", "omrchecker")
	image, err := os.ReadFile(filepath.Join(sample, "filled-sheet.png"))
	if err != nil {
		t.Fatal(err)
	}
	template, err := os.ReadFile(filepath.Join(sample, "template.json"))
	if err != nil {
		t.Fatal(err)
	}
	reading, err := runOMRChecker(context.Background(), image, nil, nil, template, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if reading.Revision != "5cf44a5a9c7a49e2e541c6e19aefd09a41a4d494" || len(reading.Fields) != 26 {
		t.Fatalf("unexpected bundled engine: revision=%s fields=%d", reading.Revision, len(reading.Fields))
	}
	if reading.Responses["Roll"] != "123456" || reading.Responses["q1"] != "A" || reading.Responses["q12"] != "AC" {
		t.Fatalf("unexpected sample recognition: %+v", reading.Responses)
	}
}

func TestOMRCheckerRejectsInvalidInputsBeforeStartingRuntime(t *testing.T) {
	for _, tc := range []struct {
		name, template string
		image          []byte
	}{
		{name: "missing template", template: ""},
		{name: "broken JSON", template: `{"fieldBlocks":`},
		{name: "invalid image", template: `{}`, image: []byte("not an image")},
		{name: "missing image", template: `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := runOMRChecker(context.Background(), tc.image, nil, nil, []byte(tc.template), nil, nil); err == nil {
				t.Fatal("invalid input reached runtime")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := runOMRChecker(ctx, nil, nil, nil, nil, nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}

func TestOMRIdentityRegionsRejectInvalidOrDuplicateBoxes(t *testing.T) {
	valid := OMRIdentityRegion{Kind: "student_name", Rect: academic.GridRect{X: .1, Y: .1, Width: .2, Height: .1}}
	for _, tc := range []struct {
		name    string
		regions []OMRIdentityRegion
	}{
		{name: "duplicate", regions: []OMRIdentityRegion{valid, valid}},
		{name: "unknown", regions: []OMRIdentityRegion{{Kind: "private_file", Rect: valid.Rect}}},
		{name: "not finite", regions: []OMRIdentityRegion{{Kind: "student_name", Rect: academic.GridRect{X: math.NaN(), Y: .1, Width: .2, Height: .1}}}},
		{name: "outside", regions: []OMRIdentityRegion{{Kind: "roll_number", Rect: academic.GridRect{X: .9, Y: .1, Width: .2, Height: .1}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runOMRChecker(context.Background(), nil, nil, nil, []byte(`{}`), nil, tc.regions)
			var problem *Error
			if !errors.As(err, &problem) || problem.Code != "invalid_identity_regions" {
				t.Fatalf("wrong boundary rejection: %v", err)
			}
		})
	}
}
