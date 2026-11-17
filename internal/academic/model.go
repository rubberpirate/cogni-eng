package academic

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"slices"
	"strings"
)

type Option struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	AssetID string `json:"asset_id,omitempty"`
	Alt     string `json:"alt,omitempty"`
}
type NumericRule struct {
	Values   []string `json:"values"`
	Min      string   `json:"min,omitempty"`
	Max      string   `json:"max,omitempty"`
	Absolute string   `json:"absolute,omitempty"`
	Relative string   `json:"relative,omitempty"`
	Integer  bool     `json:"integer"`
	Unit     string   `json:"unit,omitempty"`
}
type Criterion struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Marks       string `json:"marks"`
}
type Question struct {
	ID         string       `json:"id,omitempty"`
	Reference  string       `json:"reference,omitempty"`
	TopicID    string       `json:"topic_id"`
	Type       string       `json:"type"`
	Stem       string       `json:"stem"`
	Options    []Option     `json:"options"`
	Key        []string     `json:"key"`
	Numeric    *NumericRule `json:"numeric,omitempty"`
	Rubric     []Criterion  `json:"rubric"`
	Solution   string       `json:"solution"`
	Marks      string       `json:"marks"`
	Penalty    string       `json:"penalty"`
	Difficulty string       `json:"difficulty"`
	Language   string       `json:"language"`
	Provenance string       `json:"provenance"`
	License    string       `json:"license"`
	AssetID    string       `json:"asset_id,omitempty"`
	Alt        string       `json:"alt,omitempty"`
	State      string       `json:"state,omitempty"`
	Version    int          `json:"version,omitempty"`
	Origin     string       `json:"origin,omitempty"`
}
type Section struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Policy    string     `json:"policy"`
	Count     int        `json:"count"`
	Cap       string     `json:"cap,omitempty"`
	Questions []Question `json:"questions"`
}
type Paper struct {
	Title           string    `json:"title"`
	Purpose         string    `json:"purpose"`
	Instructions    string    `json:"instructions"`
	DurationMinutes int       `json:"duration_minutes"`
	ReleasePolicy   string    `json:"release_policy"`
	Sections        []Section `json:"sections"`
}
type Answer struct {
	Selection []string `json:"selection"`
	Number    string   `json:"number"`
	Text      string   `json:"text"`
	Visited   bool     `json:"visited"`
	Review    bool     `json:"review"`
	Order     int      `json:"order"`
	Invalid   bool     `json:"invalid"`
}
type ManualMark struct {
	Criteria map[string]string `json:"criteria"`
	Feedback string            `json:"feedback"`
	Approved bool              `json:"approved"`
}
type ItemMark struct {
	QuestionID  string `json:"question_id"`
	TopicID     string `json:"topic_id"`
	SectionID   string `json:"section_id"`
	Status      string `json:"status"`
	Award       string `json:"award"`
	Maximum     string `json:"maximum"`
	Counted     bool   `json:"counted"`
	Required    bool   `json:"required"`
	Explanation string `json:"explanation"`
	Answer      Answer `json:"answer"`
}
type Result struct {
	Score    string     `json:"score"`
	Maximum  string     `json:"maximum"`
	Correct  int        `json:"correct"`
	Wrong    int        `json:"wrong"`
	Answered int        `json:"answered"`
	Required int        `json:"required"`
	Pending  int        `json:"pending"`
	Items    []ItemMark `json:"items"`
	Policy   string     `json:"policy"`
}

var decimal = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)$`)

func Number(s string) (*big.Rat, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 || len(s) > 64 || !decimal.MatchString(s) {
		return nil, errors.New("enter a finite decimal using a dot, without commas or scientific notation")
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, errors.New("invalid decimal")
	}
	return r, nil
}
func Units(s string) (int64, error) {
	r, err := Number(s)
	if err != nil {
		return 0, err
	}
	r.Mul(r, big.NewRat(1000, 1))
	if !r.IsInt() || !r.Num().IsInt64() {
		return 0, errors.New("marks support at most three decimal places")
	}
	v := r.Num().Int64()
	if v < 0 || v > 1000000 {
		return 0, errors.New("marks must be between 0 and 1000")
	}
	return v, nil
}
func Format(v int64) string {
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	if v%1000 == 0 {
		return fmt.Sprintf("%s%d", sign, v/1000)
	}
	return strings.TrimRight(fmt.Sprintf("%s%d.%03d", sign, v/1000, v%1000), "0")
}
func ValidateQuestion(q Question) error {
	if len(strings.TrimSpace(q.Stem)) < 3 || len(q.Stem) > 16000 || q.TopicID == "" {
		return errors.New("a topic and a readable question stem are required")
	}
	if len(q.Solution) > 16000 || len(q.Provenance) > 2000 || len(q.License) > 500 || len(q.Alt) > 2000 {
		return errors.New("question guidance or provenance exceeds its input limit")
	}
	m, err := Units(q.Marks)
	if err != nil || m == 0 {
		return errors.New("positive marks are required, with at most three decimal places")
	}
	p, err := Units(q.Penalty)
	if err != nil || p > m {
		return errors.New("penalty must be between zero and the question marks")
	}
	if !slices.Contains([]string{"introductory", "standard", "challenging"}, q.Difficulty) {
		return errors.New("choose introductory, standard or challenging difficulty")
	}
	if q.Language != "en" && q.Language != "hi" {
		return errors.New("question language must be en or hi")
	}
	if strings.TrimSpace(q.Provenance) == "" || strings.TrimSpace(q.License) == "" {
		return errors.New("source/provenance and license are required")
	}
	if q.AssetID != "" && q.Alt == "" {
		return errors.New("question images require alternative text")
	}
	switch q.Type {
	case "single", "multiple":
		if len(q.Options) < 2 || len(q.Options) > 8 {
			return errors.New("selection questions require 2–8 options")
		}
		seen := map[string]bool{}
		for i, o := range q.Options {
			if o.ID != string(rune('A'+i)) {
				return errors.New("option labels must be consecutive A–H in printed order")
			}
			if len(o.Text) > 8000 || len(o.Alt) > 2000 {
				return errors.New("option content exceeds its input limit")
			}
			if len(o.ID) != 1 || o.ID < "A" || o.ID > "H" || seen[o.ID] || strings.TrimSpace(o.Text) == "" && o.AssetID == "" || o.AssetID != "" && o.Alt == "" {
				return errors.New("options need unique A–H labels, content and accessible images")
			}
			seen[o.ID] = true
		}
		if len(q.Key) < 1 || q.Type == "single" && len(q.Key) != 1 {
			return errors.New("supply the approved selection key")
		}
		keys := map[string]bool{}
		for _, k := range q.Key {
			if !seen[k] || keys[k] {
				return errors.New("answer key contains an unknown or duplicated option")
			}
			keys[k] = true
		}
	case "numerical":
		if q.Numeric == nil {
			return errors.New("supply a numerical acceptance rule")
		}
		r := q.Numeric
		if len(r.Values) > 20 || len(r.Unit) > 100 {
			return errors.New("numerical rule exceeds its input limit")
		}
		if len(r.Values) == 0 && (r.Min == "" || r.Max == "") {
			return errors.New("supply accepted numbers or both range bounds")
		}
		for _, v := range r.Values {
			n, e := Number(v)
			if e != nil || r.Integer && !n.IsInt() {
				return errors.New("invalid accepted numerical value")
			}
		}
		if r.Min != "" || r.Max != "" {
			lo, e := Number(r.Min)
			hi, e2 := Number(r.Max)
			if e != nil || e2 != nil || lo.Cmp(hi) > 0 {
				return errors.New("invalid numerical range")
			}
		}
		for _, v := range []string{r.Absolute, r.Relative} {
			if v != "" {
				n, e := Number(v)
				if e != nil || n.Sign() < 0 {
					return errors.New("tolerances must be nonnegative exact decimals")
				}
			}
		}
	case "short", "extended":
		if len(q.Rubric) == 0 || len(q.Rubric) > 20 {
			return errors.New("written responses require 1–20 rubric criteria")
		}
		var total int64
		ids := map[string]bool{}
		for _, c := range q.Rubric {
			v, e := Units(c.Marks)
			if e != nil || v == 0 || c.ID == "" || ids[c.ID] || len(c.Description) < 3 {
				return errors.New("invalid or duplicate rubric criterion")
			}
			ids[c.ID] = true
			total += v
		}
		if total != m {
			return errors.New("rubric criteria must total the question marks")
		}
	default:
		return errors.New("unsupported question type")
	}
	return nil
}
func ValidatePaper(p Paper) error {
	if len(strings.TrimSpace(p.Title)) < 3 || len(p.Title) > 180 || p.DurationMinutes < 1 || p.DurationMinutes > 480 {
		return errors.New("paper needs a title and a duration of 1–480 minutes")
	}
	if !slices.Contains([]string{"baseline", "quiz", "practice", "mock", "formal", "homework"}, p.Purpose) {
		return errors.New("unsupported paper purpose")
	}
	if p.ReleasePolicy != "immediate" && p.ReleasePolicy != "staff" {
		return errors.New("release policy must be immediate or staff")
	}
	if p.Purpose == "formal" && p.ReleasePolicy == "immediate" {
		return errors.New("formal papers require controlled staff release")
	}
	if len(p.Sections) == 0 || len(p.Sections) > 20 {
		return errors.New("a paper needs 1–20 sections")
	}
	seen, sectionIDs := map[string]bool{}, map[string]bool{}
	total := 0
	for _, s := range p.Sections {
		if s.ID == "" || sectionIDs[s.ID] || len(s.Questions) == 0 {
			return errors.New("sections need unique identifiers and questions")
		}
		sectionIDs[s.ID] = true
		if s.Policy != "all" && s.Policy != "first" && s.Policy != "best" {
			return errors.New("section policy must be all, first or best")
		}
		if s.Count < 1 || s.Count > len(s.Questions) || s.Policy == "all" && s.Count != len(s.Questions) {
			return errors.New("section contribution count is invalid")
		}
		for _, q := range s.Questions {
			if q.ID == "" || seen[q.ID] {
				return errors.New("a paper may not contain a question twice")
			}
			seen[q.ID] = true
			if len(strings.TrimSpace(q.Solution)) < 3 {
				return errors.New("published questions require a worked solution or marking guidance")
			}
			if err := ValidateQuestion(q); err != nil {
				return err
			}
			if s.Policy != "all" && (q.Marks != s.Questions[0].Marks || q.Penalty != s.Questions[0].Penalty) {
				return errors.New("optional groups require equal marks and penalties; split unequal scoring into sections")
			}
			total++
		}
		if s.Cap != "" {
			v, e := Units(s.Cap)
			if e != nil || v == 0 {
				return errors.New("section cap must be positive")
			}
		}
	}
	if total > 200 {
		return errors.New("a paper supports at most 200 questions")
	}
	return nil
}
func Blank(a Answer) bool {
	return !a.Invalid && len(a.Selection) == 0 && strings.TrimSpace(a.Number) == "" && strings.TrimSpace(a.Text) == ""
}
func ValidateAnswer(q Question, a Answer) error {
	if len(a.Text) > 16000 || len(a.Number) > 64 || len(a.Selection) > 8 {
		return errors.New("answer exceeds its input limit")
	}
	if a.Invalid {
		return errors.New("only a paper reviewer may record an invalid mark")
	}
	switch q.Type {
	case "single", "multiple":
		if a.Number != "" || a.Text != "" || q.Type == "single" && len(a.Selection) > 1 {
			return errors.New("choose a valid selection")
		}
		seen := map[string]bool{}
		for _, v := range a.Selection {
			if seen[v] || !slices.ContainsFunc(q.Options, func(o Option) bool { return o.ID == v }) {
				return errors.New("unknown or duplicate selection")
			}
			seen[v] = true
		}
	case "numerical":
		if len(a.Selection) > 0 || a.Text != "" {
			return errors.New("enter a numerical answer")
		}
		if strings.TrimSpace(a.Number) != "" {
			n, e := Number(a.Number)
			if e != nil {
				return e
			}
			if q.Numeric.Integer && !n.IsInt() {
				return errors.New("this question accepts an integer answer")
			}
		}
	default:
		if len(a.Selection) > 0 || a.Number != "" {
			return errors.New("enter a written response")
		}
	}
	return nil
}
func PublicPaper(p Paper) Paper {
	out := p
	out.Sections = make([]Section, len(p.Sections))
	for i, s := range p.Sections {
		out.Sections[i] = s
		out.Sections[i].Questions = make([]Question, len(s.Questions))
		for j, q := range s.Questions {
			q.Key = nil
			q.Numeric = nil
			q.Rubric = nil
			q.Solution = ""
			q.Provenance = ""
			q.License = ""
			q.Origin = ""
			q.State = ""
			q.Version = 0
			out.Sections[i].Questions[j] = q
		}
	}
	return out
}
