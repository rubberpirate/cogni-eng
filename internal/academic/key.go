package academic

import (
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type KeyEntry struct {
	Number    int      `json:"number"`
	Selection []string `json:"selection"`
}
type KeyPreview struct {
	Text    string     `json:"text"`
	Entries []KeyEntry `json:"entries"`
}

var keyLine = regexp.MustCompile(`(?i)^\s*(?:Q[.]?\s*)?([0-9]{1,3})[\s.,):=;-]+[([]?([A-H1-8](?:[|+][A-H1-8])*)[)\]]?\s*[.,]?\s*$`)

// ParseAnswerKey accepts explicit numbered entries, never guesses an answer from question text.
func ParseAnswerKey(text string) (KeyPreview, error) {
	out := KeyPreview{Text: text, Entries: []KeyEntry{}}
	if len(text) == 0 || len(text) > 65536 {
		return out, errors.New("answer keys must contain 1–65536 bytes")
	}
	if strings.HasPrefix(strings.TrimSpace(text), "{") {
		var values map[string]string
		if err := json.Unmarshal([]byte(text), &values); err != nil {
			return out, errors.New("use a JSON object mapping question numbers to answer labels")
		}
		text = ""
		for number, key := range values {
			text += number + " " + key + "\n"
		}
	}
	seen := map[int]bool{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.EqualFold(strings.ReplaceAll(line, " ", ""), "question,answer") || strings.EqualFold(line, "answer key") {
			continue
		}
		m := keyLine.FindStringSubmatch(line)
		if m == nil {
			return out, errors.New("unrecognized key entry; use one numbered entry per line, such as 1,B")
		}
		number, _ := strconv.Atoi(m[1])
		if number < 1 || seen[number] {
			return out, errors.New("answer key has a duplicate or invalid question number")
		}
		seen[number] = true
		entry := KeyEntry{Number: number, Selection: []string{}}
		labels := map[string]bool{}
		for _, key := range strings.FieldsFunc(strings.ToUpper(m[2]), func(r rune) bool { return r == '|' || r == '+' }) {
			if key >= "1" && key <= "8" {
				key = string(rune('A' + key[0] - '1'))
			}
			if labels[key] {
				return out, errors.New("answer key repeats an option")
			}
			labels[key] = true
			entry.Selection = append(entry.Selection, key)
		}
		out.Entries = append(out.Entries, entry)
	}
	if len(out.Entries) == 0 || len(out.Entries) > 500 {
		return out, errors.New("supply 1–500 numbered key entries")
	}
	sort.Slice(out.Entries, func(i, j int) bool { return out.Entries[i].Number < out.Entries[j].Number })
	return out, nil
}
