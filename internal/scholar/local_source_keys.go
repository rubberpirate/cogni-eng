package scholar

import (
	"regexp"
	"strconv"
	"strings"

	"learning.local/workspace/internal/academic"
)

var sourceNumber = regexp.MustCompile(`[0-9]+`)
var sourceAnswer = regexp.MustCompile(`(?im)^[ \t]*(?:answer|correct answer|key)[ \t]*[:=][ \t]*([^\r\n]+)`)
var sourceSelections = regexp.MustCompile(`(?i)^[A-H](?:[ \t]*[,|+][ \t]*[A-H])*$`)
var sourceRange = regexp.MustCompile(`^\[\s*([+-]?[0-9]+(?:\.[0-9]+)?)\s+to\s+([+-]?[0-9]+(?:\.[0-9]+)?)\s*\]$`)

// Reconcile literal printed keys rather than asking a model to infer them.
// Repeated numbering within one chunk is ambiguous and requires faculty review.
func reconcileSourceKeys(text string, rows []academic.ImportRow) {
	starts := printedQuestionBoundary.FindAllStringIndex(text, -1)
	answers := map[int][]string{}
	for i, start := range starts {
		n, _ := strconv.Atoi(sourceNumber.FindString(text[start[0]:start[1]]))
		end := len(text)
		if i+1 < len(starts) {
			end = starts[i+1][0]
		}
		matches := sourceAnswer.FindAllStringSubmatch(text[start[1]:end], -1)
		if _, exists := answers[n]; exists {
			answers[n] = nil
			continue
		}
		answers[n] = []string{}
		for _, match := range matches {
			answers[n] = append(answers[n], strings.TrimSpace(strings.ReplaceAll(match[1], "−", "-")))
		}
	}
	for i := range rows {
		r := &rows[i]
		printed, exists := answers[r.SourceNumber]
		if !exists || printed != nil && len(printed) == 0 {
			continue
		}
		q := &r.Question
		q.Key, q.Numeric = nil, nil
		if len(printed) != 1 {
			r.Error = "Printed answer identity is ambiguous; review the source and enter the key."
			continue
		}
		answer := printed[0]
		if strings.ReplaceAll(strings.ReplaceAll(q.Solution, "−", "-"), " ", "") == strings.ReplaceAll(answer, " ", "") {
			q.Solution = ""
		}
		switch {
		case sourceSelections.MatchString(answer) && (q.Type == "single" || q.Type == "multiple"):
			q.Key = strings.FieldsFunc(strings.ToUpper(answer), func(r rune) bool { return r == ',' || r == '|' || r == '+' || r == ' ' || r == '\t' })
			if len(q.Key) > 1 {
				q.Type = "multiple"
			}
		case q.Type == "numerical":
			if bounds := sourceRange.FindStringSubmatch(answer); bounds != nil {
				q.Numeric = &academic.NumericRule{Values: []string{}, Min: bounds[1], Max: bounds[2]}
			} else {
				values := strings.Split(answer, " OR ")
				rule := &academic.NumericRule{Values: values, Integer: true}
				valid := true
				for _, value := range values {
					number, err := academic.Number(value)
					if err != nil {
						valid = false
						break
					}
					rule.Integer = rule.Integer && number.IsInt()
				}
				if valid {
					q.Numeric = rule
				}
			}
		}
		r.Error = ""
		if err := academic.ValidateQuestion(*q); err != nil {
			r.Error = "Review printed answer " + answer + ": " + err.Error()
		}
	}
}
