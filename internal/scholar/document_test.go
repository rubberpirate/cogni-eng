package scholar

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// This newly authored PDF is deliberately simple: it qualifies the actual Poppler path, not handwriting fidelity.
func originalPDF() []byte {
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", "<< /Type /Pages /Kids [3 0 R 3 0 R] /Count 2 >>", "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 600 800] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>", "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"}
	content := "BT /F1 22 Tf 50 730 Td (1. Add seven and five.) Tj 0 -45 Td (A. Twelve) Tj 0 -45 Td (B. Thirteen) Tj 0 -45 Td (Answer: A) Tj 0 -45 Td (Solution: Seven plus five is twelve.) Tj ET"
	objects = append(objects, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content))
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, object := range objects {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	start := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(offsets), start)
	return out.Bytes()
}
func TestActualPDFRasterizationAndBounds(t *testing.T) {
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		t.Skip("Poppler unavailable; PDF execution not qualified")
	}
	pages, err := rasterPages(context.Background(), "application/pdf", originalPDF())
	if err != nil || len(pages) != 2 || !bytes.HasPrefix(pages[0], []byte("\x89PNG")) {
		t.Fatalf("actual multi-page rasterization %d %v", len(pages), err)
	}
	if _, err = rasterPages(context.Background(), "application/pdf", []byte("%PDF-1.4\nmalformed")); err == nil {
		t.Fatal("malformed PDF accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = rasterPages(ctx, "application/pdf", originalPDF()); err == nil || !strings.Contains(err.Error(), "PDF") {
		t.Fatalf("cancellation %v", err)
	}
}
