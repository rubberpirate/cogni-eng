package scholar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestModelContractKeepsOutputDraftAndReportsProviderFailures(t *testing.T) {
	input := Generation{Topic: Topic{ID: "topic", Subject: "Mathematics", Name: "Addition", Outcome: "Add two whole numbers", Language: "en"}, Count: 1, Difficulty: "standard"}
	generated, err := (DevelopmentModel{}).Generate(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	q := generated.Questions[0]
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.Header.Get("anthropic-version") != "2026-10-21" || r.Header.Get("x-api-key") != "contract-test-key" {
			t.Error("provider request contract")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["model"] != "configured-model" || body["tool_choice"].(map[string]any)["name"] != "submit_questions" {
			t.Error("model/tool not pinned")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"content": []any{map[string]any{"type": "tool_use", "name": "submit_questions", "input": map[string]any{"questions": []any{q}}}}, "usage": map[string]int{"input_tokens": 123, "output_tokens": 456}})
	}))
	defer server.Close()
	provider := Anthropic{Key: "contract-test-key", Model: "configured-model", Endpoint: server.URL, Client: server.Client()}
	out, err := provider.Generate(context.Background(), input)
	if err != nil || len(out.Questions) != 1 || out.Questions[0].State != "draft" || out.Provider != "anthropic" || out.InputTokens == nil || *out.InputTokens != 123 || calls != 1 {
		t.Fatalf("provider output %+v %v", out, err)
	}
	for _, status := range []int{401, 403, 429, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			failed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"error":{"message":"provider-private-detail"}}`))
			}))
			defer failed.Close()
			p := provider
			p.Endpoint = failed.URL
			_, err := p.Generate(context.Background(), input)
			if err == nil || strings.Contains(err.Error(), "provider-private-detail") {
				t.Fatalf("failure not bounded: %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = provider.Generate(ctx, input); err == nil {
		t.Fatal("cancelled provider request accepted")
	}
	unconfigured := provider
	unconfigured.Key = ""
	if _, err = unconfigured.Generate(context.Background(), input); err == nil || calls != 1 {
		t.Fatal("unconfigured request sent")
	}
}
func TestProviderTimeoutAndMalformedOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(40 * time.Millisecond)
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"not structured questions"}]}`))
	}))
	defer server.Close()
	provider := Anthropic{Key: "test", Model: "test", Endpoint: server.URL, Client: &http.Client{Timeout: 5 * time.Millisecond}}
	if _, err := provider.Generate(context.Background(), Generation{}); err == nil {
		t.Fatal("timeout accepted")
	}
	provider.Client = server.Client()
	if _, err := provider.Generate(context.Background(), Generation{}); err == nil {
		t.Fatal("malformed output accepted")
	}
}
