package usage

import (
	"encoding/json"
	"fmt"
)

const MaxAggregateModels = 128

// claudeAggregateEvidence is PURE and intentionally does not emit responses.
// Aggregate totals have neither an executing response ID nor per-response time;
// user message bodies and nested cost subdivisions cannot supply either proof.
func claudeAggregateEvidence(r record) (bool, error) {
	if r.Type == "user" {
		return false, nil
	}
	found := false
	if len(r.TotalTokens) > 0 && string(r.TotalTokens) != "null" {
		var total int64
		if err := json.Unmarshal(r.TotalTokens, &total); err != nil || total < 0 {
			return false, fmt.Errorf("invalid Claude aggregate totalTokens")
		}
		found = true
	}
	if len(r.ModelUsage) > 0 && string(r.ModelUsage) != "null" {
		var models map[string]struct {
			Input      *int64 `json:"inputTokens"`
			Output     *int64 `json:"outputTokens"`
			CacheRead  *int64 `json:"cacheReadInputTokens"`
			CacheWrite *int64 `json:"cacheCreationInputTokens"`
			Total      *int64 `json:"totalTokens"`
		}
		if err := json.Unmarshal(r.ModelUsage, &models); err != nil {
			return false, fmt.Errorf("invalid Claude aggregate modelUsage")
		}
		if len(models) > MaxAggregateModels {
			return false, fmt.Errorf("Claude aggregate model limit exceeded")
		}
		for _, model := range models {
			for _, value := range []*int64{model.Input, model.Output, model.CacheRead, model.CacheWrite, model.Total} {
				if value != nil && *value < 0 {
					return false, fmt.Errorf("negative Claude aggregate model usage")
				}
			}
		}
		found = found || len(models) > 0
	}
	return found, nil
}

// decodeClaude is PURE over one decoded record. It never adds nested cache TTL
// subdivisions or aggregate modelUsage/cost-state values to response spend.
func decodeClaude(r record, source string) (Response, bool, error) {
	if r.Type != "assistant" {
		return Response{}, false, nil
	}
	var q struct {
		ID    string          `json:"id"`
		Model string          `json:"model"`
		Stop  json.RawMessage `json:"stop_reason"`
		Usage json.RawMessage `json:"usage"`
	}
	if e := json.Unmarshal(r.Message, &q); e != nil {
		return Response{}, false, e
	}
	if len(q.Usage) == 0 || string(q.Usage) == "null" {
		return Response{}, false, nil
	}
	if q.ID == "" {
		return Response{}, false, fmt.Errorf("missing Claude message identity")
	}
	at, e := recordTime(r)
	if e != nil {
		return Response{}, false, fmt.Errorf("invalid response timestamp")
	}
	tokens, e := decodeCounters(q.Usage, true)
	if e != nil {
		return Response{}, false, e
	}
	return Response{ID: q.ID, SourceID: source, Agent: "claude", Model: q.Model, RequestID: r.RequestID, At: at, Tokens: &tokens, Provisional: len(q.Stop) == 0 || string(q.Stop) == "null", Origin: "message"}, true, nil
}
