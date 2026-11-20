package scholar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func pdfPageCount(ctx context.Context, input string) (int, error) {
	info, err := exec.CommandContext(ctx, "pdfinfo", input).Output()
	if err != nil {
		return 0, errors.New("PDF cannot be inspected; remove encryption or upload page images")
	}
	pages := 0
	for _, line := range bytes.Split(info, []byte{'\n'}) {
		var n int
		if _, e := fmt.Sscanf(string(line), "Pages: %d", &n); e == nil {
			pages = n
		}
	}
	if pages < 1 || pages > 50 {
		return 0, errors.New("PDF must contain 1–50 pages; split larger batches")
	}
	return pages, nil
}

// Prefer a usable PDF text layer to raster OCR: formulas and printed keys should
// not be discarded by OCR layout heuristics. Scans still use the selected OCR engine.
func embeddedPDFText(ctx context.Context, data []byte) (string, bool, error) {
	dir, err := os.MkdirTemp("", "academic-pdf-text-")
	if err != nil {
		return "", false, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	input, output := filepath.Join(dir, "source.pdf"), filepath.Join(dir, "text.txt")
	if err = os.WriteFile(input, data, 0600); err != nil {
		return "", false, err
	}
	if _, err = pdfPageCount(ctx, input); err != nil {
		return "", false, err
	}
	if err = exec.CommandContext(ctx, "pdftotext", "-layout", "-enc", "UTF-8", input, output).Run(); err != nil {
		if ctx.Err() != nil {
			return "", false, ctx.Err()
		}
		return "", false, nil
	}
	info, err := os.Stat(output)
	if err != nil {
		return "", false, err
	}
	if info.Size() > 1<<20 {
		return "", false, errors.New("extracted text exceeds 1 MiB; split the document")
	}
	b, err := os.ReadFile(output)
	if err != nil {
		return "", false, err
	}
	text := strings.TrimSpace(string(b))
	if len(text) < 100 || !utf8.ValidString(text) || strings.Count(text, "\uFFFD") > utf8.RuneCountInString(text)/100 || !printedQuestionBoundary.MatchString(text) {
		return "", false, nil
	}
	return text, true, nil
}
