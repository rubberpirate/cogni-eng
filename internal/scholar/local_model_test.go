package scholar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"learning.local/workspace/internal/academic"
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
	if schema, ok := input["format"].(map[string]any); !ok || schema["type"] != "object" {
		t.Fatal("document output schema missing")
	}
	schema := input["format"].(map[string]any)
	array := schema["properties"].(map[string]any)["rows"].(map[string]any)
	if array["minItems"] != float64(1) || array["maxItems"] != float64(1) {
		t.Fatal("source question count not constrained")
	}
	if rows[0].SourceNumber != 7 || rows[0].Question.TopicID != "topic" {
		t.Fatal("source identity lost")
	}
	m.Endpoint = "https://example.com"
	if _, err = m.Generate(context.Background(), Generation{}); err == nil {
		t.Fatal("remote endpoint accepted")
	}
}

func TestLocalModelTutor(t *testing.T) {
	promptEvaluated := int64(128)
	evaluated := int64(64)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Error("wrong path")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"done":              true,
			"done_reason":       "stop",
			"prompt_eval_count": promptEvaluated,
			"eval_count":        evaluated,
			"message": map[string]string{
				"content": "Here is an explanation of the problem...",
			},
		})
	}))
	defer server.Close()

	m := LocalModel{Endpoint: server.URL, Model: "qwen3.5:9b-q4_K_M", Client: server.Client()}
	q := questionExample("topic")
	ans := academic.Answer{Selection: []string{"A"}}
	out, err := m.Tutor(context.Background(), q, ans, []TutorMessage{{Role: "user", Text: "Can you help?"}})
	if err != nil {
		t.Fatalf("unexpected tutor error: %v", err)
	}
	if out.Provider != "local_ollama" || out.Model != "qwen3.5:9b-q4_K_M" {
		t.Fatalf("unexpected provider or model: %+v", out)
	}
	if out.Text != "Here is an explanation of the problem..." {
		t.Fatalf("unexpected tutor text: %q", out.Text)
	}
	if out.InputTokens == nil || *out.InputTokens != 128 || out.OutputTokens == nil || *out.OutputTokens != 64 {
		t.Fatalf("unexpected token counts: %+v", out)
	}
}

func TestLocalExtractionRejectsMissingSourceQuestion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, _ := json.Marshal(map[string]any{"rows": []academic.ImportRow{{SourceNumber: 1, Question: questionExample("topic")}}})
		_ = json.NewEncoder(w).Encode(map[string]any{"done": true, "message": map[string]string{"content": string(content)}})
	}))
	defer server.Close()
	m := LocalModel{Endpoint: server.URL, Model: "test-local", Client: server.Client()}
	if _, err := m.Extract(context.Background(), "Q1 First\nQ2 Numerical without options\n", "topic", "source", "license"); err == nil {
		t.Fatal("incomplete extraction accepted")
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

func TestDocumentChunksRecognizePrintedJEEQuestionLabels(t *testing.T) {
	text := "Paper instructions\n"
	for i := 1; i <= 19; i++ {
		label := fmt.Sprintf("Q.%d", i)
		if i%2 == 0 {
			label = fmt.Sprintf("Q%d", i)
		}
		text += label + " Complete scientific question?\n(A) One\n(B) Two\n"
	}
	chunks, err := documentChunks(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 3 || !strings.Contains(chunks[1], "Q.9 ") || !strings.Contains(chunks[2], "Q.17 ") {
		t.Fatalf("printed Q.1 labels not split into bounded groups: %d", len(chunks))
	}
}

func TestLocalGenerationKeepsInitialDraftAndUsesConsistencyReview(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		q := questionExample("topic")
		q.Stem = "Which option equals two plus three?"
		q.Options[0].Text = "Four"
		q.Options[1].Text = "Five"
		q.Solution = "Two plus three equals five, option B."
		q.Key = []string{"A"}
		if calls == 2 {
			q.Key = []string{"B"}
		}
		content, _ := json.Marshal(map[string]any{"questions": []academic.Question{q}})
		_ = json.NewEncoder(w).Encode(map[string]any{"done": true, "message": map[string]string{"content": string(content)}})
	}))
	defer server.Close()
	m := LocalModel{Endpoint: server.URL, Model: "test-local", Client: server.Client()}
	out, err := m.Generate(context.Background(), Generation{Topic: Topic{ID: "topic", Language: "en"}, Count: 1, Difficulty: "standard"})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(out.InitialQuestions) != 1 || out.InitialQuestions[0].Key[0] != "A" || out.Questions[0].Key[0] != "B" {
		t.Fatal("initial draft or review result lost")
	}
}
