package academic

import (
	"encoding/json"
	"regexp"
	"strings"
)

var questionStart = regexp.MustCompile(`^(?:Q(?:uestion)?\s*)?([0-9]{1,3})[.)、:]\s*(.+)$`)
var optionStart = regexp.MustCompile(`^\(?([A-Ha-h])\s*[).:]\s*(.+)$`)
var keyStart = regexp.MustCompile(`(?i)^(?:answer|key|correct answer)\s*[:=]\s*(.+)$`)

// Text is a transcription proposal: missing/uncertain keys stay invalid until a teacher supplies them.
func ParseText(text, topic, provenance, license string) ([]ImportRow, error) {
	questions := []Question{}
	current := Question{}
	flush := func() {
		if strings.TrimSpace(current.Stem) != "" {
			questions = append(questions, current)
		}
		current = Question{Type: "single"}
	}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if match := questionStart.FindStringSubmatch(line); match != nil {
			flush()
			current.Stem = match[2]
			continue
		}
		if match := optionStart.FindStringSubmatch(line); match != nil {
			current.Options = append(current.Options, Option{ID: strings.ToUpper(match[1]), Text: match[2]})
			continue
		}
		if match := keyStart.FindStringSubmatch(line); match != nil {
			current.Key = append(current.Key, strings.FieldsFunc(strings.ToUpper(match[1]), func(r rune) bool { return r == ',' || r == '|' || r == ';' || r == ' ' })...)
			if len(current.Key) > 1 {
				current.Type = "multiple"
			}
			continue
		}
		if strings.HasPrefix(strings.ToLower(line), "solution:") {
			current.Solution = strings.TrimSpace(line[len("solution:"):])
			continue
		}
		if len(current.Options) > 0 {
			current.Options[len(current.Options)-1].Text += " " + line
		} else {
			current.Stem += " " + line
		}
	}
	flush()
	data, err := json.Marshal(questions)
	if err != nil {
		return nil, err
	}
	return ParseImport("transcription.json", data, topic, provenance, license)
}
