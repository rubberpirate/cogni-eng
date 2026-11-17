package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/scholar"
)

func modelCheck(ctx context.Context) error {
	provider := scholar.ProviderFromEnv()
	in := scholar.Generation{Topic: scholar.Topic{ID: "synthetic-addition", Subject: "Mathematics", Name: "Whole number addition", Outcome: "Add two small positive whole numbers and explain the sum.", Language: "en"}, Count: 1, Difficulty: "introductory"}
	start := time.Now()
	generated, err := provider.Generate(ctx, in)
	if err != nil {
		return fmt.Errorf("live question-generation check: %w", err)
	}
	tutor, ok := provider.(interface {
		Tutor(context.Context, academic.Question, academic.Answer, []scholar.TutorMessage) (scholar.TutorOutput, error)
	})
	if !ok {
		return fmt.Errorf("configured provider has no real tutor adapter")
	}
	reply, err := tutor.Tutor(ctx, generated.Questions[0], academic.Answer{Selection: []string{"B"}}, []scholar.TutorMessage{{Role: "user", Text: "Explain the reasoning and give me a similar check question."}})
	if err != nil {
		return fmt.Errorf("generation succeeded; live tutor check: %w", err)
	}
	data, err := json.MarshalIndent(struct {
		CheckedAt  time.Time           `json:"checked_at"`
		DurationMS int64               `json:"duration_ms"`
		Generated  scholar.Generated   `json:"generated"`
		Tutor      scholar.TutorOutput `json:"tutor"`
	}{time.Now().UTC(), time.Since(start).Milliseconds(), generated, reply}, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(".local", 0700); err != nil {
		return err
	}
	if err = os.WriteFile(".local/model-check.json", append(data, '\n'), 0600); err != nil {
		return err
	}
	fmt.Printf("Live %s generation and tutoring succeeded; resolved model %s; generation tokens input=%v output=%v; elapsed=%s. Synthetic output saved privately in .local/model-check.json.\n", generated.Provider, generated.Model, tokenCount(generated.InputTokens), tokenCount(generated.OutputTokens), time.Since(start).Round(time.Millisecond))
	return nil
}
func tokenCount(n *int64) any {
	if n == nil {
		return "unknown"
	}
	return *n
}
