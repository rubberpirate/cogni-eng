//go:build localmodel

package scholar

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

type recordedTransport struct{ output string }

func (r recordedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(r.output, b, 0600); err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(b))
	return resp, nil
}

// Opt-in, owner-provided input and private raw output; never part of ordinary unit tests.
func TestLiveDocumentModelFirstChunk(t *testing.T) {
	source, output := os.Getenv("LOCAL_MODEL_LIVE_PAPER"), os.Getenv("LOCAL_MODEL_LIVE_OUTPUT")
	if source == "" || output == "" {
		t.Skip("set explicit source and private output paths")
	}
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	text, usable, err := embeddedPDFText(ctx, data)
	if err != nil || !usable {
		t.Fatalf("PDF text: %v %v", usable, err)
	}
	chunks, err := documentChunks(text)
	if err != nil {
		t.Fatal(err)
	}
	m := localModel()
	m.Client = &http.Client{Timeout: 10 * time.Minute, Transport: recordedTransport{output: output}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	rows, err := m.Extract(ctx, chunks[0], "topic", "Owner-supplied local evaluation", "Private test source")
	if err != nil {
		t.Fatal(err)
	}
	if expected := len(printedQuestionBoundary.FindAllStringIndex(chunks[0], -1)); expected > 0 && len(rows) != expected {
		t.Fatalf("returned %d of %d printed questions", len(rows), expected)
	}
	t.Logf("Live first chunk: %d rows; %d source chunks; inspect private output for scientific fidelity", len(rows), len(chunks))
}
