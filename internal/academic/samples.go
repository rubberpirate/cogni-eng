package academic

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"strings"
)

//go:embed fixtures/questions.json
var sampleQuestions []byte

//go:embed fixtures/questions.csv
var SampleCSV []byte

//go:embed fixtures/question-paper.pdf
var SamplePaperPDF []byte

//go:embed fixtures/marking-guide.pdf
var SampleKeyPDF []byte

func Samples(topic string) ([]Question, error) {
	var out []Question
	if err := json.Unmarshal(sampleQuestions, &out); err != nil {
		return nil, err
	}
	for i := range out {
		out[i].TopicID = topic
		if err := ValidateQuestion(out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// ImagesPDF embeds generated sheet pixels without changing the registration geometry.
func ImagesPDF(pages ...image.Image) ([]byte, error) {
	if len(pages) < 1 || len(pages) > 4 {
		return nil, errors.New("sample PDFs require one to four pages")
	}
	objects := [][]byte{[]byte("<< /Type /Catalog /Pages 2 0 R >>"), nil}
	kids := ""
	for _, img := range pages {
		var encoded bytes.Buffer
		if err := jpeg.Encode(&encoded, img, &jpeg.Options{Quality: 98}); err != nil {
			return nil, err
		}
		n := len(objects) + 1
		kids += fmt.Sprintf("%d 0 R ", n)
		draw := []byte("q 595 0 0 833 0 4 cm /Sheet Do Q")
		objects = append(objects, []byte(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /XObject << /Sheet %d 0 R >> >> /Contents %d 0 R >>", n+2, n+1)), append([]byte(fmt.Sprintf("<< /Length %d >>\nstream\n", len(draw))), append(draw, []byte("\nendstream")...)...), append([]byte(fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length %d >>\nstream\n", img.Bounds().Dx(), img.Bounds().Dy(), encoded.Len())), append(encoded.Bytes(), []byte("\nendstream")...)...))
	}
	objects[1] = []byte(fmt.Sprintf("<< /Type /Pages /Count %d /Kids [%s] >>", len(pages), kids))
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n", i+1)
		out.Write(obj)
		out.WriteString("\nendobj\n")
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, v := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", v)
	}
	fmt.Fprintf(&out, "trailer << /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(offsets), xref)
	return out.Bytes(), nil
}

func IsPublicSample(q Question) bool {
	known, err := Samples("sample")
	if err != nil {
		return false
	}
	normalize := func(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }
	for _, original := range known {
		if q.Type != original.Type || normalize(q.Stem) != normalize(original.Stem) || len(q.Options) != len(original.Options) {
			continue
		}
		match := true
		for i, o := range q.Options {
			if o.ID != original.Options[i].ID || normalize(o.Text) != normalize(original.Options[i].Text) {
				match = false
			}
		}
		if match {
			return true
		}
	}
	return false
}
