package academic

import (
	"strings"
	"testing"
)

func TestPaperTranscriptionPreservesNumberingOptionsAndUntrustedPrintedKeys(t *testing.T) {
	text := "Exam instructions\nQ. 2. Choose a prime.\n(a) Four (b) Five\n{c) Six (d) Eight\nAns. (a) Four.\nQ.1.Add two and three.\n(a) Four\n(b) Five\n(c) Six\n(d) Seven\nAnswer: B\n"
	rows, err := ParseText(text, "topic", "Original fixture", "CC0-1.0")
	if err != nil || len(rows) != 2 {
		t.Fatalf("rows %v %v", rows, err)
	}
	if rows[0].SourceNumber != 1 || rows[1].SourceNumber != 2 || rows[0].Question.Key[0] != "B" || len(rows[1].Question.Key) != 0 || rows[1].Error == "" {
		t.Fatal("numbering or printed-key boundary")
	}
	for _, row := range rows {
		if len(row.Question.Options) != 4 {
			t.Fatal("inline option lost")
		}
		for i, o := range row.Question.Options {
			if o.ID != string(rune('A'+i)) || strings.Contains(o.Text, "Ans.") {
				t.Fatal("option ordering/source answer contamination")
			}
		}
	}
}
func TestOrderedOCRReassemblesColumnsAndKeepsNativeLineMembership(t *testing.T) {
	header := "level\tpage_num\tblock_num\tpar_num\tline_num\tword_num\tleft\ttop\twidth\theight\tconf\ttext\n"
	// Original word-line membership keeps a sloping line intact; block order is deliberately reversed.
	tsv := header + "5\t1\t3\t1\t1\t1\t650\t110\t30\t20\t95\tQ.2.Second\n" + "5\t1\t1\t1\t1\t1\t80\t100\t30\t20\t95\tQ.1.First\n" + "5\t1\t2\t1\t1\t1\t50\t130\t30\t15\t95\tSloped\n" + "5\t1\t2\t1\t1\t2\t300\t142\t30\t22\t95\tsentence.\n"
	if raw := RawOCRText(tsv); !strings.HasPrefix(raw, "Q.2.Second") || !strings.Contains(raw, "Sloped sentence.") {
		t.Fatalf("raw native text lost: %q", raw)
	}
	text := OrderedOCRText(tsv, 1200)
	if text != "Q.1.First\nSloped sentence.\nQ.2.Second\n" {
		t.Fatalf("column reading order: %q", text)
	}
}
