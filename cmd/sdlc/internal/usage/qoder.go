package usage

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/qodertranscript"
)

// decodeQoder collects bounded metadata epochs first so a replayed old config
// cannot rewind the config applicable at a response's actual timestamp.
func decodeQoder(ctx context.Context, reader io.Reader) ([]Response, Coverage, error) {
	var coverage Coverage
	coverage.Add("sources", "")
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), MaxLineBytes)
	partialTail := false
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		n, b, e := bufio.ScanLines(data, atEOF)
		partialTail = atEOF && n == len(data) && len(b) > 0 && !bytes.HasSuffix(data, []byte("\n"))
		return n, b, e
	})
	var state qodertranscript.SessionState
	var candidates []qodertranscript.Record
	for scanner.Scan() {
		if e := ctx.Err(); e != nil {
			return nil, coverage, e
		}
		coverage.Add("records", "")
		if len(scanner.Bytes()) == 0 {
			continue
		}
		r, e := qodertranscript.ParseRecord(scanner.Bytes())
		if e != nil {
			coverage.Add("malformed_records", "invalid Qoder record")
			if partialTail {
				coverage.Add("truncated_tails", "incomplete final JSONL record")
			}
			continue
		}
		if e = state.Step(r); e != nil {
			// Parsed metadata can disagree about attribution without hiding charges.
			// Session identity conflicts are rejected by Identity below.
			coverage.Add("unknown_metadata", "ambiguous Qoder attribution metadata")
		}
		if r.Type == "assistant" {
			if r.Message == nil || r.Message.ID == "" {
				coverage.Add("malformed_records", "missing Qoder response identity")
				continue
			}
			if len(candidates) >= MaxResponses {
				coverage.Add("bound_exceeded", "Qoder response candidate limit")
				return nil, coverage, fmt.Errorf("usage response limit exceeded")
			}
			candidates = append(candidates, r)
		}
	}
	if e := ctx.Err(); e != nil {
		return nil, coverage, e
	}
	if e := scanner.Err(); e != nil {
		coverage.Add("unavailable_sources", "read failure or oversized JSONL line")
		return nil, coverage, e
	}
	source, e := state.Identity()
	if e != nil {
		coverage.Add("session_conflicts", "unknown or conflicting Qoder session identity")
		return nil, coverage, nil
	}
	if _, e = state.Metadata(); e != nil {
		coverage.Add("unknown_metadata", "incomplete Qoder workspace metadata")
	}
	finished := map[string]bool{}
	for _, r := range candidates {
		m := r.Message
		if m.StopReason == "end_turn" || m.StopReason == "tool_use" {
			finished[m.ID] = true
		}
	}
	var out []Response
	for _, r := range candidates {
		m := r.Message
		terminal := m.StopReason == "end_turn" || m.StopReason == "tool_use"
		if !terminal && finished[m.ID] {
			continue
		}
		model, _, configErr := state.ConfigAt(r.Timestamp)
		if configErr != nil || model == "" {
			model = ""
			coverage.Add("unknown_model", "no unambiguous applicable Qoder runtime configuration")
			coverage.Add("unknown_metadata", "incomplete Qoder response configuration")
		}
		response := Response{ID: m.ID, SourceID: source, Agent: "qoder", At: r.Timestamp, Model: model, Origin: "qoder-session:" + source, Provisional: !terminal}
		if !terminal {
			out = append(out, response)
			continue
		}
		amount, e := ParseCredit(m.Usage["credits"])
		if e != nil {
			coverage.Add("malformed_records", "missing or invalid Qoder credits")
			response.Provisional = true
			out = append(out, response)
			continue
		}
		response.Credits = &amount
		if original, ok := m.Usage["original_credits"]; ok {
			if _, e = ParseCredit(original); e != nil {
				coverage.Add("malformed_records", "invalid Qoder original credits")
			}
		}
		if !bytes.Equal(m.Usage["billable"], []byte("true")) {
			coverage.Add("unknown_charge", "Qoder charge semantics not confirmed billable")
		}
		// This format has token placeholders, not a supported token measurement.
		out = append(out, response)
	}
	return out, coverage, nil
}
