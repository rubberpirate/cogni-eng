package academic

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ImportRow struct {
	Index       int      `json:"index"`
	Question    Question `json:"question"`
	Error       string   `json:"error"`
	ExistingID  string   `json:"existing_id,omitempty"`
	CommittedID string   `json:"committed_id,omitempty"`
}

func ParseImport(filename string, data []byte, topic, provenance, license string) ([]ImportRow, error) {
	if len(data) > 8<<20 {
		return nil, errors.New("structured imports are limited to 8 MiB")
	}
	var questions []Question
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".json":
		d := json.NewDecoder(bytes.NewReader(data))
		d.DisallowUnknownFields()
		if err := d.Decode(&questions); err != nil {
			return nil, fmt.Errorf("invalid question JSON: %w", err)
		}
		if err := d.Decode(new(any)); err != io.EOF {
			return nil, errors.New("send one JSON array of questions")
		}
	case ".csv", ".xlsx":
		var records [][]string
		if strings.HasSuffix(strings.ToLower(filename), ".csv") {
			reader := csv.NewReader(bytes.NewReader(data))
			reader.FieldsPerRecord = -1
			var err error
			records, err = reader.ReadAll()
			if err != nil {
				return nil, fmt.Errorf("invalid CSV: %w", err)
			}
		} else {
			book, err := excelize.OpenReader(bytes.NewReader(data), excelize.Options{UnzipSizeLimit: 32 << 20, UnzipXMLSizeLimit: 8 << 20})
			if err != nil {
				return nil, fmt.Errorf("invalid XLSX: %w", err)
			}
			defer func() { _ = book.Close() }()
			sheets := book.GetSheetList()
			if len(sheets) == 0 {
				return nil, errors.New("the workbook has no worksheet")
			}
			records, err = book.GetRows(sheets[0], excelize.Options{RawCellValue: true})
			if err != nil {
				return nil, err
			}
		}
		if len(records) < 2 {
			return nil, errors.New("include a header and at least one question row")
		}
		header := map[string]int{}
		for i, h := range records[0] {
			h = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(h)), "\ufeff")
			h = strings.ReplaceAll(h, " ", "_")
			if _, exists := header[h]; exists {
				return nil, errors.New("duplicate import column")
			}
			header[h] = i
		}
		for _, record := range records[1:] {
			get := func(names ...string) string {
				for _, name := range names {
					if i, ok := header[name]; ok && i < len(record) {
						return strings.TrimSpace(record[i])
					}
				}
				return ""
			}
			if strings.TrimSpace(strings.Join(record, "")) == "" {
				continue
			}
			q := Question{Type: get("type", "question_type"), Stem: get("stem", "question", "question_text"), Marks: get("marks"), Penalty: get("penalty", "negative_marks"), Solution: get("solution", "explanation"), Difficulty: get("difficulty"), Language: get("language"), Provenance: get("provenance", "source"), License: get("license")}
			for _, label := range []string{"A", "B", "C", "D", "E", "F", "G", "H"} {
				v := get(strings.ToLower(label), "option_"+strings.ToLower(label))
				if v != "" {
					q.Options = append(q.Options, Option{ID: label, Text: v})
				}
			}
			if raw := get("options"); raw != "" {
				if err := json.Unmarshal([]byte(raw), &q.Options); err != nil {
					q.Options = nil
				}
			}
			key := get("answer", "key", "correct_answer")
			switch strings.ToLower(q.Type) {
			case "numeric", "numerical", "integer", "decimal":
				q.Numeric = &NumericRule{Values: strings.Split(key, "|"), Absolute: get("absolute_tolerance"), Relative: get("relative_tolerance"), Min: get("min"), Max: get("max"), Integer: strings.ToLower(q.Type) == "integer"}
				if key == "" {
					q.Numeric.Values = nil
				}
			default:
				for _, v := range regexp.MustCompile(`[,;|+]`).Split(key, -1) {
					v = strings.TrimSpace(strings.ToUpper(v))
					if len(v) == 1 && v >= "1" && v <= "8" {
						v = string(rune('A' + v[0] - '1'))
					}
					if v != "" {
						q.Key = append(q.Key, v)
					}
				}
			}
			if raw := get("rubric"); raw != "" {
				if err := json.Unmarshal([]byte(raw), &q.Rubric); err != nil {
					q.Rubric = nil
				}
			}
			questions = append(questions, q)
		}
	default:
		return nil, errors.New("use a JSON, CSV or XLSX question file")
	}
	if len(questions) == 0 || len(questions) > 500 {
		return nil, errors.New("import 1–500 questions per file")
	}
	out := make([]ImportRow, 0, len(questions))
	for i, q := range questions {
		switch strings.ToLower(strings.TrimSpace(q.Type)) {
		case "mcq", "single_choice", "single":
			q.Type = "single"
		case "multi", "multi_select", "multiple_choice", "multiple":
			q.Type = "multiple"
		case "numeric", "numerical", "integer", "decimal":
			q.Type = "numerical"
		case "short_answer", "short":
			q.Type = "short"
		case "essay", "extended":
			q.Type = "extended"
		}
		q.ID = ""
		q.Reference = ""
		q.State = ""
		q.Version = 0
		q.Origin = ""
		q.TopicID = topic
		if q.Marks == "" {
			q.Marks = "1"
		}
		if q.Penalty == "" {
			q.Penalty = "0"
		}
		if q.Difficulty == "" {
			q.Difficulty = "standard"
		}
		if q.Language == "" {
			q.Language = "en"
		}
		if q.Provenance == "" {
			q.Provenance = provenance
		}
		if q.License == "" {
			q.License = license
		}
		row := ImportRow{Index: i + 1, Question: q}
		if err := ValidateQuestion(q); err != nil {
			row.Error = err.Error()
		}
		out = append(out, row)
	}
	return out, nil
}
