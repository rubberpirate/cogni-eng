package scholar

import (
	"context"
	"errors"
	"math"
	"testing"

	"learning.local/workspace/internal/academic"
)

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
