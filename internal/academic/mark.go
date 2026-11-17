package academic

import (
	"errors"
	"math/big"
	"slices"
	"sort"
)

func numerical(q Question, value string) bool {
	n, err := Number(value)
	if err != nil || q.Numeric.Integer && !n.IsInt() {
		return false
	}
	rule := q.Numeric
	if rule.Min != "" {
		lo, _ := Number(rule.Min)
		hi, _ := Number(rule.Max)
		if n.Cmp(lo) >= 0 && n.Cmp(hi) <= 0 {
			return true
		}
	}
	absolute, relative := new(big.Rat), new(big.Rat)
	if rule.Absolute != "" {
		absolute, _ = Number(rule.Absolute)
	}
	if rule.Relative != "" {
		relative, _ = Number(rule.Relative)
	}
	for _, v := range rule.Values {
		expected, _ := Number(v)
		distance := new(big.Rat).Abs(new(big.Rat).Sub(n, expected))
		tolerance := new(big.Rat).Mul(new(big.Rat).Abs(expected), relative)
		if tolerance.Cmp(absolute) < 0 {
			tolerance.Set(absolute)
		}
		if distance.Cmp(tolerance) <= 0 {
			return true
		}
	}
	return false
}
func Mark(p Paper, answers map[string]Answer, manual map[string]ManualMark, channel string) (Result, error) {
	if err := ValidatePaper(p); err != nil {
		return Result{}, err
	}
	out := Result{Policy: "objective-v1", Items: []ItemMark{}}
	var score, maximum int64
	for _, section := range p.Sections {
		type candidate struct {
			index     int
			award     int64
			order     int
			attempted bool
		}
		candidates := []candidate{}
		items := []ItemMark{}
		var sectionMaximum int64
		for index, q := range section.Questions {
			a := answers[q.ID]
			m, _ := Units(q.Marks)
			penalty, _ := Units(q.Penalty)
			item := ItemMark{QuestionID: q.ID, TopicID: q.TopicID, SectionID: section.ID, Maximum: q.Marks, Status: "blank", Award: "0", Explanation: "No response recorded.", Answer: a}
			var award int64
			if !Blank(a) {
				item.Status = "wrong"
				award = -penalty
				item.Explanation = "The response does not match the pinned answer rule; the published penalty applies."
				if a.Invalid {
					item.Status = "invalid"
					item.Explanation = "Reviewer recorded an invalid response; the published penalty applies."
				} else if q.Type == "short" || q.Type == "extended" {
					mark, ok := manual[q.ID]
					if !ok || !mark.Approved {
						item.Status = "pending"
						award = 0
						item.Explanation = "Written work awaits approved rubric marking."
					} else {
						award = 0
						for _, c := range q.Rubric {
							v, e := Units(mark.Criteria[c.ID])
							max, _ := Units(c.Marks)
							if e != nil || v > max {
								return Result{}, errors.New("approved criterion marks exceed the pinned rubric")
							}
							award += v
						}
						for id := range mark.Criteria {
							if !slices.ContainsFunc(q.Rubric, func(c Criterion) bool { return c.ID == id }) {
								return Result{}, errors.New("unknown rubric criterion")
							}
						}
						item.Status = "reviewed"
						item.Explanation = mark.Feedback
					}
				}
				if q.Type == "single" || q.Type == "multiple" {
					actual := slices.Clone(a.Selection)
					key := slices.Clone(q.Key)
					slices.Sort(actual)
					slices.Sort(key)
					if !a.Invalid && slices.Equal(actual, key) {
						item.Status = "correct"
						award = m
						item.Explanation = "Exact selection set matches the pinned key."
					}
				}
				if q.Type == "numerical" && !a.Invalid && numerical(q, a.Number) {
					item.Status = "correct"
					award = m
					item.Explanation = "The exact decimal satisfies the pinned numerical acceptance rule."
				}
			}
			item.Award = Format(award)
			items = append(items, item)
			order := a.Order
			if channel == "paper" || order == 0 {
				order = index + 1
			}
			candidates = append(candidates, candidate{index, award, order, !Blank(a)})
		}
		selected := map[int]bool{}
		if section.Policy == "all" {
			for i := range items {
				selected[i] = true
			}
		} else {
			sort.SliceStable(candidates, func(i, j int) bool {
				a, b := candidates[i], candidates[j]
				if a.attempted != b.attempted {
					return a.attempted
				}
				if section.Policy == "best" && a.award != b.award {
					return a.award > b.award
				}
				if section.Policy == "first" && a.order != b.order {
					return a.order < b.order
				}
				return a.index < b.index
			})
			for _, c := range candidates[:section.Count] {
				selected[c.index] = true
			}
		}
		var sectionScore int64
		for i, item := range items {
			item.Counted = selected[i]
			item.Required = item.Counted
			if item.Counted {
				m, _ := Units(item.Maximum)
				sectionMaximum += m
				v, _ := new(big.Rat).SetString(item.Award)
				v.Mul(v, big.NewRat(1000, 1))
				sectionScore += v.Num().Int64()
				out.Required++
				if item.Status != "blank" {
					out.Answered++
				}
				switch item.Status {
				case "correct":
					out.Correct++
				case "wrong", "invalid":
					out.Wrong++
				case "pending":
					out.Pending++
				}
			} else {
				if item.Status == "blank" {
					item.Status = "not_required"
					item.Explanation = "Optional blank outside the required contribution count."
				} else {
					item.Status = "uncounted_optional"
					item.Explanation = "Response retained but excluded by the pinned optional-section policy."
				}
				item.Award = "0"
			}
			out.Items = append(out.Items, item)
		}
		if section.Cap != "" {
			cap, _ := Units(section.Cap)
			if sectionMaximum > cap {
				sectionMaximum = cap
			}
			if sectionScore > cap {
				sectionScore = cap
			}
		}
		score += sectionScore
		maximum += sectionMaximum
	}
	out.Score = Format(score)
	out.Maximum = Format(maximum)
	return out, nil
}
