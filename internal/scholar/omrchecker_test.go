package scholar

import (
	"context"
	"errors"
	"testing"
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
			if _, err := runOMRChecker(context.Background(), tc.image, nil, nil, []byte(tc.template), nil); err == nil {
				t.Fatal("invalid input reached runtime")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := runOMRChecker(ctx, nil, nil, nil, nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}
