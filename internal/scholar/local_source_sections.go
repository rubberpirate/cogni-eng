package scholar

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"learning.local/workspace/internal/academic"
)

var printedSection = regexp.MustCompile(`(?im)^[ \t]*SECTION[ \t]+[0-9]+\b[^\r\n]*`)
var printedFullMarks = regexp.MustCompile(`(?i)Full\s+Marks\s*:\s*\+?([0-9]+(?:\.[0-9]+)?)`)
var printedNegativeMarks = regexp.MustCompile(`(?i)Negative\s+Marks\s*:\s*[-−]?([0-9]+(?:\.[0-9]+)?)`)

func checkSourceQuestionIdentity(text string, rows []academic.ImportRow) error {
	starts := printedQuestionBoundary.FindAllStringIndex(text, -1)
	if len(starts) == 0 {
		return nil
	}
	if len(starts) != len(rows) {
		return fmt.Errorf("source has %d numbered questions but extraction returned %d; retry or split the source", len(starts), len(rows))
	}
	for i, start := range starts {
		n, _ := strconv.Atoi(sourceNumber.FindString(text[start[0]:start[1]]))
		if n != rows[i].SourceNumber {
			return fmt.Errorf("source question order changed at position %d; retry or review the saved extraction", i+1)
		}
	}
	return nil
}

// Section instructions apply across chunk boundaries and subject-number resets.
func reconcileSourceSections(text string, rows []academic.ImportRow) error {
	sections := printedSection.FindAllStringIndex(text, -1)
	starts := printedQuestionBoundary.FindAllStringIndex(text, -1)
	if len(sections) == 0 {
		return nil
	}
	if err := checkSourceQuestionIdentity(text, rows); err != nil {
		return err
	}
	for i, start := range starts {
		header := -1
		for _, section := range sections {
			if section[0] > start[0] {
				break
			}
			header = section[0]
		}
		if header < 0 {
			continue
		}
		first := start[0]
		for _, question := range starts {
			if question[0] > header {
				first = question[0]
				break
			}
		}
		instructions := strings.Join(strings.Fields(text[header:first]), " ")
		upper := strings.ToUpper(instructions)
		q := &rows[i].Question
		switch {
		case strings.Contains(upper, "NUMERICAL VALUE") || strings.Contains(upper, "NON-NEGATIVE INTEGER"):
			q.Type, q.Options, q.Key, q.Numeric = "numerical", nil, nil, nil
		case strings.Contains(upper, "ONE OR MORE"):
			q.Type, q.Numeric = "multiple", nil
		case strings.Contains(upper, "ONLY ONE"):
			q.Type, q.Numeric = "single", nil
		}
		if marks := printedFullMarks.FindStringSubmatch(instructions); marks != nil {
			q.Marks = marks[1]
		}
		if penalty := printedNegativeMarks.FindStringSubmatch(instructions); penalty != nil {
			q.Penalty = penalty[1]
		} else if strings.Contains(upper, "ZERO MARKS") && strings.Contains(upper, "IN ALL OTHER CASES") {
			q.Penalty = "0"
		}
		rows[i].Error = ""
		if err := academic.ValidateQuestion(*q); err != nil {
			rows[i].Error = err.Error()
		}
		end := len(text)
		if i+1 < len(starts) {
			end = starts[i+1][0]
		}
		reconcileSourceKeys(text[start[0]:end], rows[i:i+1])
	}
	return nil
}
