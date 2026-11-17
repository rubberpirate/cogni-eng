package scholar

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/academic"
)

type TopicEvidence struct {
	TopicID       string `json:"topic_id"`
	Name          string `json:"name"`
	Counted       int    `json:"counted"`
	Answered      int    `json:"answered"`
	Correct       int    `json:"correct"`
	Wrong         int    `json:"wrong"`
	RequiredBlank int    `json:"required_blank"`
	Pending       int    `json:"pending"`
	Status        string `json:"status"`
	Explanation   string `json:"explanation"`
}

func (s *Service) Evidence(ctx context.Context, tx pgx.Tx, actor, org, sitting string) ([]TopicEvidence, error) {
	out, err := s.Sitting(ctx, tx, actor, org, sitting)
	if err != nil {
		return nil, err
	}
	if out.State != "released" || out.Result == nil {
		return nil, Problem(403, "evidence_unreleased", "Topic evidence is available after result release.")
	}
	byTopic := map[string]*TopicEvidence{}
	ordered := []string{}
	for _, item := range out.Result.Items {
		if !item.Counted {
			continue
		}
		if byTopic[item.TopicID] == nil {
			topic, err := GetTopic(ctx, tx, org, item.TopicID)
			if err != nil {
				return nil, err
			}
			ordered = append(ordered, item.TopicID)
			byTopic[item.TopicID] = &TopicEvidence{TopicID: item.TopicID, Name: topic.Name}
		}
		t := byTopic[item.TopicID]
		t.Counted++
		switch item.Status {
		case "correct":
			t.Correct++
			t.Answered++
		case "wrong", "invalid":
			t.Wrong++
			t.Answered++
		case "blank":
			t.RequiredBlank++
		case "reviewed":
			t.Answered++
		case "pending":
			t.Pending++
		}
	}
	result := []TopicEvidence{}
	for _, id := range ordered {
		t := byTopic[id]
		t.Status = "insufficient_evidence"
		t.Explanation = "This is observed evidence from one sitting, not an estimate of mastery. Optional uncounted items are excluded; a required blank records missing work, not a diagnosis of ignorance. Timing was not measured."
		if t.Answered >= 5 && t.Pending == 0 {
			t.Status = "observed_responses"
		}
		result = append(result, *t)
	}
	return result, nil
}

// Result history always reads the current revision; old corrections never leave stale derived topic totals.
type ResultHistory struct {
	Revision  int             `json:"revision"`
	Result    academic.Result `json:"result"`
	Reason    string          `json:"reason"`
	CreatedAt string          `json:"created_at"`
}

func (s *Service) History(ctx context.Context, tx pgx.Tx, actor, org, id string) ([]ResultHistory, error) {
	out, err := loadSitting(ctx, tx, org, id)
	if err != nil {
		return nil, err
	}
	if err = paperStaff(ctx, tx, actor, org, out.Paper); err != nil {
		return nil, err
	}
	type stored struct {
		Revision        int
		Result          []byte
		Reason, Created string
	}
	list, err := rows[stored](ctx, tx, "SELECT revision,result,reason,created_at::text FROM result_revision WHERE institution_id=$1 AND sitting_id=$2 ORDER BY revision DESC", org, id)
	if err != nil {
		return nil, err
	}
	result := []ResultHistory{}
	for _, v := range list {
		r := ResultHistory{Revision: v.Revision, Reason: v.Reason, CreatedAt: v.Created}
		if err = json.Unmarshal(v.Result, &r.Result); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
