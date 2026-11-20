package scholar

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
)

var printedQuestionBoundary = regexp.MustCompile(`(?im)^[ \t]*(?:Q(?:uestion)?[ .:]*[0-9]{1,3}\b[.)、:]*[ \t]*|[0-9]{1,3}[.)、:][ \t]+)`)

func documentChunks(text string) ([]string, error) {
	starts := printedQuestionBoundary.FindAllStringIndex(text, -1)
	if len(starts) < 2 {
		if len(text) > 18000 {
			return nil, fmt.Errorf("question boundaries are unclear in this large document; split it or correct the saved OCR text")
		}
		return []string{text}, nil
	}
	chunks := []string{}
	start := 0
	for i := 8; i < len(starts); i += 8 {
		end := starts[i][0]
		chunks = append(chunks, text[start:end])
		start = end
	}
	chunks = append(chunks, text[start:])
	for _, chunk := range chunks {
		if len(chunk) > 24000 {
			return nil, fmt.Errorf("a question group exceeds the local context bound; split the source paper")
		}
	}
	return chunks, nil
}
func (s *Service) normalizeLocalDocument(ctx context.Context, j Job, in map[string]string, text string) ([]academic.ImportRow, error) {
	chunks, err := documentChunks(text)
	if err != nil {
		return nil, err
	}
	m := localModel()
	hash := access.Digest(text + "\x00" + m.Model)
	var previous struct {
		Hash   string                 `json:"normalization_hash"`
		Chunks [][]academic.ImportRow `json:"normalized_chunks"`
	}
	if err = Tx(ctx, s.DB, func(tx pgx.Tx) error {
		if e := s.jobAllowed(ctx, tx, j); e != nil {
			return e
		}
		var output []byte
		if e := tx.QueryRow(ctx, "SELECT output FROM academic_job WHERE id=$1", j.ID).Scan(&output); e != nil {
			return e
		}
		return json.Unmarshal(output, &previous)
	}); err != nil {
		return nil, err
	}
	if previous.Hash != hash {
		previous.Chunks = [][]academic.ImportRow{}
	}
	rows := []academic.ImportRow{}
	for i, chunk := range chunks {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		var parsed []academic.ImportRow
		if i < len(previous.Chunks) && checkSourceQuestionIdentity(chunk, previous.Chunks[i]) == nil {
			parsed = previous.Chunks[i]
		} else {
			if err = Tx(ctx, s.DB, func(tx pgx.Tx) error { return s.jobAllowed(ctx, tx, j) }); err != nil {
				return nil, err
			}
			parsed, err = m.Extract(ctx, chunk, in["topic_id"], in["provenance"], in["license"])
			if err != nil {
				return nil, err
			}
			if len(rows)+len(parsed) > 200 {
				return nil, fmt.Errorf("more than 200 questions detected; split this assessment")
			}
			if i < len(previous.Chunks) {
				previous.Chunks[i] = parsed
			} else {
				previous.Chunks = append(previous.Chunks, parsed)
			}
			if err = Tx(ctx, s.DB, func(tx pgx.Tx) error {
				if e := s.jobAllowed(ctx, tx, j); e != nil {
					return e
				}
				provider := "ocr+local_ollama"
				if in["text_method"] == "embedded_pdf" {
					provider = "pdf_text+local_ollama"
				}
				_, e := tx.Exec(ctx, "UPDATE academic_job SET output=output || $2::jsonb,provider=$5,model=$3,progress=$4,updated_at=now() WHERE id=$1", j.ID, Encode(map[string]any{"normalization_hash": hash, "normalized_chunks": previous.Chunks}), m.Model, 90+(i+1)*9/len(chunks), provider)
				return e
			}); err != nil {
				return nil, err
			}
		}
		reconcileSourceKeys(chunk, parsed)
		for _, row := range parsed {
			row.Index = len(rows) + 1
			rows = append(rows, row)
		}
	}
	if len(rows) > 200 {
		return nil, fmt.Errorf("more than 200 questions detected; split this assessment")
	}
	if err := reconcileSourceSections(text, rows); err != nil {
		return nil, err
	}
	return rows, nil
}
