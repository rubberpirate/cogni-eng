package scholar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"learning.local/workspace/internal/academic"
)

func TestGeminiStructuredGenerationAndTutorContracts(t *testing.T) {
	input := Generation{Topic: Topic{ID: "topic", Subject: "Mathematics", Name: "Addition", Outcome: "Add whole numbers", Language: "en"}, Count: 1, Difficulty: "standard"}
	fixture, err := (DevelopmentModel{}).Generate(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("X-goog-api-key") != "private-test-key" || r.URL.RawQuery != "" || r.URL.Path != "/models/gemini-flash-latest:generateContent" {
			t.Error("Gemini authentication/route contract")
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		config := request["generationConfig"].(map[string]any)
		text := "Try adding equal-sized groups and check the total."
		if calls == 1 {
			if config["responseMimeType"] != "application/json" || config["responseJsonSchema"] == nil {
				t.Error("structured schema missing")
			}
			text = string(Encode(map[string]any{"questions": fixture.Questions}))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"candidates": []any{map[string]any{"content": map[string]any{"parts": []any{map[string]any{"text": "internal reasoning", "thought": true}, map[string]any{"text": text}}}, "finishReason": "STOP"}}, "modelVersion": "resolved-model-version", "usageMetadata": map[string]int{"promptTokenCount": 90, "candidatesTokenCount": 35}})
	}))
	defer server.Close()
	provider := Gemini{Key: "private-test-key", Model: "gemini-flash-latest", Endpoint: server.URL, Client: server.Client()}
	generated, err := provider.Generate(context.Background(), input)
	if err != nil || len(generated.Questions) != 1 || generated.Questions[0].State != "draft" || generated.Provider != "gemini" || generated.Model != "resolved-model-version" || generated.InputTokens == nil || *generated.InputTokens != 90 {
		t.Fatalf("generated %+v %v", generated, err)
	}
	reply, err := provider.Tutor(context.Background(), generated.Questions[0], academic.Answer{Selection: []string{"A"}}, []TutorMessage{{Role: "user", Text: "Explain my answer"}})
	if err != nil || strings.Contains(reply.Text, "internal reasoning") || reply.Provider != "gemini" || reply.OutputTokens == nil || *reply.OutputTokens != 35 {
		t.Fatalf("tutor %+v %v", reply, err)
	}
}
func TestGeminiFailureBoundaries(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 429, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"error":{"message":"private-test-key provider details"}}`))
			}))
			defer server.Close()
			provider := Gemini{Key: "private-test-key", Model: "configured", Endpoint: server.URL, Client: server.Client()}
			_, _, err := provider.call(context.Background(), "system", "prompt", nil, 20)
			if err == nil || strings.Contains(err.Error(), "private-test-key") {
				t.Fatalf("unbounded error %v", err)
			}
		})
	}
	for _, reason := range []string{"MAX_TOKENS", "SAFETY", ""} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"candidates": []any{map[string]any{"content": map[string]any{"parts": []any{map[string]string{"text": "partial"}}}, "finishReason": reason}}})
		}))
		provider := Gemini{Key: "test", Model: "test", Endpoint: server.URL, Client: server.Client()}
		if _, _, err := provider.call(context.Background(), "system", "prompt", nil, 20); err == nil {
			t.Errorf("incomplete candidate accepted: %s", reason)
		}
		server.Close()
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	provider := Gemini{Key: "test", Model: "test", Endpoint: "http://127.0.0.1:1", Client: &http.Client{Timeout: time.Second}}
	if _, _, err := provider.call(ctx, "system", "prompt", nil, 20); err == nil {
		t.Fatal("cancelled call accepted")
	}
	provider.Key = ""
	if _, _, err := provider.call(context.Background(), "system", "prompt", nil, 20); err == nil {
		t.Fatal("unconfigured provider accepted")
	}
}
