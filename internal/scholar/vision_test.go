package scholar

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"learning.local/workspace/internal/academic"
)

func visionImage(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.White)
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func TestVisionImageOnlyOCRAndFailureBoundaries(t *testing.T) {
	data := visionImage(t)
	for _, tc := range []struct {
		name          string
		status        int
		finish, model string
		want          bool
	}{{"complete", 200, "stop", LightOnModel, true}, {"truncated", 200, "length", LightOnModel, false}, {"wrong model", 200, "stop", OMRVisionModel, false}, {"busy", 503, "", "", false}, {"authentication", 401, "", "", false}} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer private-contract-key" || r.URL.Path != "/v1/chat/completions" || r.URL.RawQuery != "" {
					t.Error("inference authentication/path")
				}
				var input struct {
					Model    string `json:"model"`
					Messages []struct {
						Content []struct {
							Type  string `json:"type"`
							Image struct {
								URL string `json:"url"`
							} `json:"image_url"`
						} `json:"content"`
					} `json:"messages"`
				}
				if json.NewDecoder(r.Body).Decode(&input) != nil || input.Model != LightOnModel || len(input.Messages) != 1 || len(input.Messages[0].Content) != 1 {
					t.Fatal("OCR instruction/data boundary")
				}
				part := input.Messages[0].Content[0]
				decoded, e := base64.StdEncoding.DecodeString(strings.TrimPrefix(part.Image.URL, "data:image/png;base64,"))
				if e != nil || !bytes.Equal(decoded, data) {
					t.Error("image altered")
				}
				w.WriteHeader(tc.status)
				if err := json.NewEncoder(w).Encode(map[string]any{"model": tc.model, "model_revision": "actual-revision", "choices": []any{map[string]any{"finish_reason": tc.finish, "message": map[string]string{"content": "Q.1. Original fixture"}}}, "usage": map[string]int{"prompt_tokens": 100, "completion_tokens": 6}}); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			out, err := (Vision{Endpoint: server.URL + "/v1", Key: "private-contract-key", Client: server.Client()}).Read(context.Background(), data, LightOnModel, "")
			if (err == nil) != tc.want {
				t.Fatalf("%+v %v", out, err)
			}
			if err == nil && (out.Revision != "actual-revision" || out.InputTokens == nil || *out.InputTokens != 100) {
				t.Fatal("lost actual inference provenance")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (Vision{Endpoint: "http://127.0.0.1:1/v1", Key: "key", Client: &http.Client{Timeout: time.Second}}).Read(ctx, data, LightOnModel, "")
	if err == nil {
		t.Fatal("cancelled inference succeeded")
	}
}
func TestModelOMRRejectsInventedMissingAndInvalidRows(t *testing.T) {
	expected := []academic.GridRow{{Number: 1}, {Number: 2}}
	for _, tc := range []struct {
		name, text string
		want       bool
	}{{"valid", `{"rows":[{"number":2,"selection":[],"uncertain":true},{"number":1,"selection":[3],"uncertain":false}]}`, true}, {"missing", `{"rows":[{"number":1,"selection":[3]}]}`, false}, {"duplicate", `{"rows":[{"number":1,"selection":[3]},{"number":1,"selection":[3]}]}`, false}, {"unknown label", `{"rows":[{"number":1,"selection":[5]},{"number":2,"selection":[]}]}`, false}, {"trailing payload", `{"rows":[{"number":1,"selection":[3]},{"number":2,"selection":[]}]} {}`, false}} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseModelGrid(tc.text, expected, 4)
			if (err == nil) != tc.want {
				t.Fatalf("%v", err)
			}
		})
	}
}
