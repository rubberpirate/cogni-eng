package scholar

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const LightOnModel = "lightonai/LightOnOCR-2-1B"
const OMRVisionModel = "Qwen/Qwen3.5-2B"
const BubbleMarkModel = "learning/bubble-marks-v1"
const BubbleMarkRevision = "ae8401755d721992a87790a63da31c115004de264901632864a7a50ba850d36a"

type VisionOutput struct {
	RuntimeVersion string `json:"runtime_version,omitempty"`
	Text           string `json:"text"`
	Model          string `json:"model"`
	Revision       string `json:"revision"`
	InputTokens    *int64 `json:"input_tokens,omitempty"`
	OutputTokens   *int64 `json:"output_tokens,omitempty"`
	DurationMS     int64  `json:"duration_ms"`
}

type Vision struct {
	Endpoint string
	Key      string
	Client   *http.Client
}

func VisionFromEnv() Vision {
	return Vision{Endpoint: os.Getenv("VISION_BASE_URL"), Key: os.Getenv("VISION_API_KEY"), Client: &http.Client{Timeout: 150 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("vision redirects are forbidden") }}}
}

// Read calls an explicitly configured private inference endpoint; document text is never an instruction channel.
func (v Vision) Read(ctx context.Context, page []byte, model, prompt string) (VisionOutput, error) {
	var out VisionOutput
	endpoint, err := url.Parse(v.Endpoint)
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" || v.Key == "" {
		return out, Problem(503, "provider_configuration_required", "Start the configured vision runtime and set VISION_BASE_URL/VISION_API_KEY on the worker. The source is retained.")
	}
	if model != LightOnModel && model != OMRVisionModel && model != BubbleMarkModel {
		return out, errors.New("unsupported vision model")
	}
	if len(page) > 8<<20 {
		return out, errors.New("vision image exceeds 8 MiB")
	}
	mime := http.DetectContentType(page)
	if mime != "image/png" && mime != "image/jpeg" {
		return out, errors.New("vision inference requires a rasterized PNG or JPEG")
	}
	content := []any{map[string]any{"type": "image_url", "image_url": map[string]string{"url": "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(page)}}}
	if prompt != "" {
		content = append(content, map[string]string{"type": "text", "text": prompt})
	}
	body := Encode(map[string]any{"model": model, "messages": []any{map[string]any{"role": "user", "content": content}}, "max_tokens": 4096, "temperature": 0})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(v.Endpoint, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return out, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+v.Key)
	start := time.Now()
	response, err := v.Client.Do(req)
	if err != nil {
		return out, errors.New("vision request interrupted; source and saved job are retained")
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(response.Body, (2<<20)+1))
	if err != nil {
		return out, err
	}
	if len(raw) > 2<<20 {
		return out, errors.New("vision output exceeds 2 MiB")
	}
	if response.StatusCode != http.StatusOK {
		if response.StatusCode == 401 || response.StatusCode == 403 || response.StatusCode == 404 {
			return out, Problem(503, "provider_configuration_required", "Vision authentication or model endpoint is unavailable. Correct worker/runtime configuration and retry.")
		}
		return out, fmt.Errorf("vision runtime returned HTTP %d; source retained for retry", response.StatusCode)
	}
	var envelope struct {
		RuntimeVersion string `json:"runtime_version"`
		Model          string `json:"model"`
		Revision       string `json:"model_revision"`
		Choices        []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Finish string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			Input  *int64 `json:"prompt_tokens"`
			Output *int64 `json:"completion_tokens"`
		} `json:"usage"`
	}
	if json.Unmarshal(raw, &envelope) != nil || len(envelope.Choices) != 1 || envelope.Choices[0].Finish != "stop" || strings.TrimSpace(envelope.Choices[0].Message.Content) == "" || envelope.Model != model {
		return out, errors.New("vision output was incomplete, malformed or from a different model; no drafts or marks were published")
	}
	if model == BubbleMarkModel && envelope.Revision != BubbleMarkRevision {
		return out, Problem(503, "provider_configuration_required", "The mark-classifier weights do not match the reviewed revision. Deploy matching Go/runtime artifacts and retry.")
	}
	out = VisionOutput{RuntimeVersion: envelope.RuntimeVersion, Text: envelope.Choices[0].Message.Content, Model: envelope.Model, Revision: envelope.Revision, InputTokens: envelope.Usage.Input, OutputTokens: envelope.Usage.Output, DurationMS: time.Since(start).Milliseconds()}
	return out, nil
}
