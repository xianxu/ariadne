package usage

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const MaxLineBytes = 4 << 20
const MaxResponses = 250000

// wire: Codex and Claude transcript records evolve with their producers.
type record struct {
	Type        string          `json:"type"`
	Timestamp   string          `json:"timestamp"`
	Payload     json.RawMessage `json:"payload"`
	RequestID   string          `json:"requestId"`
	Message     json.RawMessage `json:"message"`
	ModelUsage  json.RawMessage `json:"modelUsage"`
	TotalTokens json.RawMessage `json:"totalTokens"`
}

func recordTime(r record) (time.Time, error) { return time.Parse(time.RFC3339Nano, r.Timestamp) }

// Decode is the INTEGRATION streaming reader. Its pure record transitions are
// below the reader: it owns cancellation/line bounds but no filesystem or Git.
func Decode(ctx context.Context, agent, sourceID string, reader io.Reader) ([]Response, Coverage, error) {
	var coverage Coverage
	if err := ctx.Err(); err != nil {
		return nil, coverage, err
	}
	coverage.Add("sources", "")
	if agent == "qoder" {
		return decodeQoder(ctx, reader)
	}
	if agent != "codex" && agent != "claude" {
		coverage.Add("unsupported_formats", "unknown agent")
		return nil, coverage, nil
	}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), MaxLineBytes)
	partialTail := false
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		advance, token, err := bufio.ScanLines(data, atEOF)
		partialTail = atEOF && advance == len(data) && len(token) > 0 && !bytes.HasSuffix(data, []byte("\n"))
		return advance, token, err
	})
	state := codexState{source: sourceID, models: map[string]string{}}
	var out []Response
	for scanner.Scan() {
		if e := ctx.Err(); e != nil {
			return nil, coverage, e
		}
		coverage.Add("records", "")
		if len(scanner.Bytes()) == 0 {
			continue
		}
		var r record
		if e := json.Unmarshal(scanner.Bytes(), &r); e != nil {
			coverage.Add("malformed_records", "invalid JSON record")
			if agent == "codex" && !state.metadataSeen {
				// Discovery cannot recover identity across an unreadable prefix.
				// It may contain the actual first header; a later copied header
				// cannot safely replace that unknown executing identity.
				state.metadataSeen = true
				coverage.Add("unknown_metadata", "unreadable Codex metadata prefix")
			}
			if partialTail {
				coverage.Add("truncated_tails", "incomplete final JSONL record")
			}
			continue
		}
		if agent == "codex" {
			if e := state.step(r, &coverage); e != nil {
				coverage.Add("malformed_records", e.Error())
			}
		} else {
			if aggregate, err := claudeAggregateEvidence(r); err != nil {
				coverage.Add("malformed_records", err.Error())
			} else if aggregate {
				coverage.Add("aggregate_usage_unreconciled", "structured Claude aggregate usage lacks response-level reconciliation")
			}
			response, ok, e := decodeClaude(r, sourceID)
			if e != nil {
				coverage.Add("malformed_records", e.Error())
			} else if ok {
				out = append(out, response)
			}
		}
		if len(out)+len(state.modern)+len(state.legacy) > MaxResponses {
			coverage.Add("bound_exceeded", "response identity limit")
			return nil, coverage, fmt.Errorf("usage response limit exceeded")
		}
	}
	if e := ctx.Err(); e != nil {
		return nil, coverage, e
	}
	if e := scanner.Err(); e != nil {
		coverage.Add("unavailable_sources", "read failure or oversized JSONL line")
		return nil, coverage, e
	}
	if agent == "codex" {
		out = state.finish(&coverage)
	}
	return out, coverage, nil
}

func decodeCounters(raw json.RawMessage, claude bool) (Counters, error) {
	var values map[string]json.RawMessage
	if e := json.Unmarshal(raw, &values); e != nil || values == nil {
		return Counters{}, fmt.Errorf("missing usage dimensions")
	}
	get := func(k string) (*int64, error) {
		v, ok := values[k]
		if !ok || string(v) == "null" {
			return nil, nil
		}
		var n int64
		if e := json.Unmarshal(v, &n); e != nil || n < 0 {
			return nil, fmt.Errorf("invalid %s", k)
		}
		return &n, nil
	}
	in, e := get("input_tokens")
	if e != nil {
		return Counters{}, e
	}
	out, e := get("output_tokens")
	if e != nil {
		return Counters{}, e
	}
	if in == nil || out == nil {
		return Counters{}, fmt.Errorf("missing input/output usage")
	}
	c := Counters{Input: *in, Output: *out}
	read, write := "cached_input_tokens", "cache_write_input_tokens"
	if claude {
		read, write = "cache_read_input_tokens", "cache_creation_input_tokens"
	}
	if c.CacheRead, e = get(read); e != nil {
		return c, e
	}
	if c.CacheWrite, e = get(write); e != nil {
		return c, e
	}
	if c.Reasoning, e = get("reasoning_output_tokens"); e != nil {
		return c, e
	}
	if claude && c.Reasoning == nil {
		if raw, ok := values["output_tokens_details"]; ok && string(raw) != "null" {
			var details struct {
				Thinking *int64 `json:"thinking_tokens"`
			}
			if err := json.Unmarshal(raw, &details); err != nil {
				return c, fmt.Errorf("invalid output token details")
			}
			c.Reasoning = details.Thinking
		}
	}
	if claude {
		for _, v := range []*int64{c.CacheRead, c.CacheWrite} {
			if v != nil {
				if c.Input, e = checkedAdd(c.Input, *v); e != nil {
					return c, e
				}
			}
		}
	}
	if e = c.Validate(); e != nil {
		return c, e
	}
	if !claude {
		if total, e := get("total_tokens"); e != nil {
			return c, e
		} else if total != nil {
			want, e := c.Total()
			if e != nil || want != *total {
				return c, fmt.Errorf("inconsistent total usage")
			}
		}
	}
	return c, nil
}
