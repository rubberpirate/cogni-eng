package academic

import (
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var questionStart = regexp.MustCompile(`(?i)^(?:Q(?:uestion)?[.,]?\s*)?([0-9]{1,3})[.)、:]\s*(.+)$`)
var optionInline = regexp.MustCompile(`(?:^|\s)[({]?([A-Ha-h])[)}.:]\s*`)
var keyStart = regexp.MustCompile(`(?i)^(?:answer|key|correct answer)\s*[:=]\s*(.+)$`)
var printedKey = regexp.MustCompile(`(?i)^["'’\s]*Ans[.,:]\s*`)

// ParseText retains printed numbering. Solved-paper Ans. lines are not trusted keys.
func ParseText(text, topic, provenance, license string) ([]ImportRow, error) {
	questions := []Question{}
	numbers := []int{}
	current := Question{Type: "single"}
	number := 0
	flush := func() {
		if number > 0 && strings.TrimSpace(current.Stem) != "" {
			sort.SliceStable(current.Options, func(i, j int) bool { return current.Options[i].ID < current.Options[j].ID })
			questions = append(questions, current)
			numbers = append(numbers, number)
		}
		current = Question{Type: "single"}
	}
	for _, line := range strings.Split(text, "\n") {
		line = strings.ReplaceAll(strings.TrimSpace(line), "{¢)", "(c)")
		if line == "" {
			continue
		}
		if match := questionStart.FindStringSubmatch(line); match != nil {
			flush()
			number, _ = strconv.Atoi(match[1])
			current.Stem = match[2]
			continue
		}
		if number == 0 {
			continue
		}
		if printedKey.MatchString(line) {
			continue
		}
		if match := keyStart.FindStringSubmatch(line); match != nil {
			current.Key = strings.FieldsFunc(strings.ToUpper(match[1]), func(r rune) bool { return r == ',' || r == '|' || r == ';' || r == ' ' })
			if len(current.Key) > 1 {
				current.Type = "multiple"
			}
			continue
		}
		if strings.HasPrefix(strings.ToLower(line), "solution:") {
			current.Solution = strings.TrimSpace(line[len("solution:"):])
			continue
		}
		matches := optionInline.FindAllStringSubmatchIndex(line, -1)
		if len(matches) > 0 && strings.TrimSpace(line[:matches[0][0]]) == "" {
			for i, m := range matches {
				end := len(line)
				if i+1 < len(matches) {
					end = matches[i+1][0]
				}
				current.Options = append(current.Options, Option{ID: strings.ToUpper(line[m[2]:m[3]]), Text: strings.TrimSpace(line[m[1]:end])})
			}
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
	out, err := ParseImport("transcription.json", data, topic, provenance, license)
	for i := range out {
		out[i].SourceNumber = numbers[i]
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].SourceNumber < out[j].SourceNumber })
	for i := range out {
		out[i].Index = i + 1
	}
	return out, err
}

type ocrWord struct {
	x, y, w, h int
	text       string
	line       string
}

var ocrQuestion = regexp.MustCompile(`(?i)^Q[.,]?(?:[0-9]|$)`)

// OrderedOCRText uses word geometry, so OCR block order cannot move options to another question.
func OrderedOCRText(tsv string, width int) string {
	words := []ocrWord{}
	starts := []int{}
	firstY := int(^uint(0) >> 1)
	for _, line := range strings.Split(tsv, "\n") {
		f := strings.SplitN(line, "\t", 12)
		if len(f) != 12 || f[0] != "5" || strings.TrimSpace(f[11]) == "" {
			continue
		}
		x, e1 := strconv.Atoi(f[6])
		y, e2 := strconv.Atoi(f[7])
		w, e3 := strconv.Atoi(f[8])
		h, e4 := strconv.Atoi(f[9])
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
			continue
		}
		word := ocrWord{x: x, y: y, w: w, h: h, text: f[11], line: strings.Join(f[1:5], "/")}
		words = append(words, word)
		if ocrQuestion.MatchString(word.text) {
			starts = append(starts, x)
			if y < firstY {
				firstY = y
			}
		}
	}
	sort.Ints(starts)
	boundaries := []int{0}
	for i := 1; i < len(starts); i++ {
		if starts[i]-starts[i-1] > width/4 {
			boundaries = append(boundaries, starts[i]-width/25)
		}
	}
	boundaries = append(boundaries, width+1)
	var out strings.Builder
	for col := 0; col < len(boundaries)-1; col++ {
		group := []ocrWord{}
		for _, w := range words {
			if w.x >= boundaries[col] && w.x < boundaries[col+1] && (firstY == int(^uint(0)>>1) || w.y+w.h >= firstY) {
				group = append(group, w)
			}
		}
		lines := map[string][]ocrWord{}
		for _, w := range group {
			lines[w.line] = append(lines[w.line], w)
		}
		ordered := [][]ocrWord{}
		for _, line := range lines {
			sort.Slice(line, func(i, j int) bool { return line[i].x < line[j].x })
			ordered = append(ordered, line)
		}
		sort.SliceStable(ordered, func(i, j int) bool {
			yi, yj := ordered[i][0].y, ordered[j][0].y
			if yi == yj {
				return ordered[i][0].x < ordered[j][0].x
			}
			return yi < yj
		})
		for _, line := range ordered {
			for i, w := range line {
				if i > 0 {
					out.WriteByte(' ')
				}
				out.WriteString(w.text)
			}
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// RawOCRText retains all native word lines, including headers and instruction text.
func RawOCRText(tsv string) string {
	var out strings.Builder
	previous := ""
	for _, line := range strings.Split(tsv, "\n") {
		f := strings.SplitN(line, "\t", 12)
		if len(f) != 12 || f[0] != "5" || strings.TrimSpace(f[11]) == "" {
			continue
		}
		id := strings.Join(f[1:5], "/")
		if previous != "" {
			if previous == id {
				out.WriteByte(' ')
			} else {
				out.WriteByte('\n')
			}
		}
		out.WriteString(f[11])
		previous = id
	}
	return out.String()
}
