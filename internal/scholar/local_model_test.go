package scholar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLocalModelGenerationAndDocumentContract(t *testing.T) {
	var input map[string]any
	q := questionExample("topic")
	q.Stem = "What is two plus two?"
	q.Solution = "Two plus two is four."
	q.Options[0].Text = "Four"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Error("wrong path")
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		payload := map[string]any{"questions": []any{q}, "rows": []any{map[string]any{"index": 1, "source_number": 7, "question": q}}}
		b, _ := json.Marshal(payload)
		_ = json.NewEncoder(w).Encode(map[string]any{"done": true, "done_reason": "stop", "message": map[string]string{"content": string(b)}})
	}))
	defer server.Close()
	m := LocalModel{Endpoint: server.URL, Model: "test-local", Client: server.Client()}
	generated, err := m.Generate(context.Background(), Generation{Topic: Topic{ID: "topic", Subject: "Mathematics", Name: "Arithmetic", Language: "en"}, Count: 1, Difficulty: "standard", Exam: "NEET"})
	if err != nil {
		t.Fatal(err)
	}
	if len(generated.Questions) != 1 || generated.Provider != "local_ollama" || generated.Questions[0].State != "draft" {
		t.Fatal("invalid generation")
	}
	if input["stream"] != false || input["think"] != false || input["format"] != "json" {
		t.Fatal("structured bounded request not sent")
	}
	if !strings.Contains(string(Encode(input["messages"])), "NEET") {
		t.Fatal("target exam omitted from generation request")
	}
	rows, err := m.Extract(context.Background(), "7. What is two plus two? A. Four B. Five Answer: A", "topic", "Uploaded source", "Institution-owned")
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].SourceNumber != 7 || rows[0].Question.TopicID != "topic" {
		t.Fatal("source identity lost")
	}
	m.Endpoint = "https://example.com"
	if _, err = m.Generate(context.Background(), Generation{}); err == nil {
		t.Fatal("remote endpoint accepted")
	}
}
func TestLocalModelRejectsTruncationAndCountMismatch(t *testing.T) {
	for _, content := range []string{`{"done":true,"done_reason":"length","message":{"content":"{}"}}`, `{"done":true,"message":{"content":"{\"questions\":[]}"}}`, `{"done":false}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(content)) }))
		m := LocalModel{Endpoint: server.URL, Model: "test", Client: server.Client()}
		if _, err := m.Generate(context.Background(), Generation{Count: 1}); err == nil {
			t.Error("incomplete output accepted")
		}
		server.Close()
	}
}

func TestDocumentChunksKeepWholeNumberedQuestions(t *testing.T) {
	text := "Original paper\n"
	for i := 1; i <= 19; i++ {
		text += fmt.Sprintf("%d. Complete question %d?\nA. One\nB. Two\n", i, i)
	}
	chunks, err := documentChunks(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 3 || !strings.HasPrefix(chunks[1], "9.") || !strings.HasPrefix(chunks[2], "17.") {
		t.Fatal("question boundaries split incorrectly")
	}
	if _, err := documentChunks(strings.Repeat("unstructured ", 2000)); err == nil {
		t.Fatal("unbounded unstructured source accepted")
	}
}

func TestAssessmentProviderHasNoHostedFallback(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ACADEMIC_MODEL_PROVIDER", "gemini")
	if _, ok := ProviderFromEnv().(LocalModel); !ok {
		t.Fatal("hosted assessment provider selected")
	}
	t.Setenv("ACADEMIC_MODEL_PROVIDER", "development_fixture")
	if _, ok := ProviderFromEnv().(LocalModel); !ok {
		t.Fatal("fixture selected in production")
	}
	t.Setenv("APP_ENV", "development")
	if _, ok := ProviderFromEnv().(DevelopmentModel); !ok {
		t.Fatal("explicit development fixture unavailable")
	}
}
