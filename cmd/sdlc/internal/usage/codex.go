package usage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

type codexState struct {
	source, model, thread string
	metadataSeen          bool
	models                map[string]string
	modern                []Response
	legacy                []Response
	modernSeen            bool
	baseline              *Counters
	ordinal               int
	embedded              map[string]bool
	original              map[string]time.Time
	expected              *Counters
	legacyExpected        *Counters
	turnExpected          map[string]*Counters
}
type codexUsage struct {
	ResponseID  string          `json:"response_id"`
	ThreadID    string          `json:"thread_id"`
	TurnID      string          `json:"turn_id"`
	Usage       json.RawMessage `json:"usage"`
	ThreadUsage json.RawMessage `json:"thread_token_usage"`
	TurnUsage   json.RawMessage `json:"turn_token_usage"`
}

func (s *codexState) step(r record, c *Coverage) error {
	switch r.Type {
	case "session_meta":
		if s.metadataSeen {
			return nil
		}
		s.metadataSeen = true
		id, err := transcripts.CodexExecutingID(r.Payload)
		if err != nil {
			c.Add("unknown_metadata", "invalid first Codex executing identity")
			return err
		}
		s.thread = id
	case "turn_context":
		if s.thread == "" {
			// Model context from an unidentified prefix may belong to copied
			// history. It cannot establish either turn or legacy stream models.
			return nil
		}
		var q struct {
			TurnID string `json:"turn_id"`
			Model  string `json:"model"`
		}
		if e := json.Unmarshal(r.Payload, &q); e != nil {
			return e
		}
		s.model = q.Model
		if old, ok := s.models[q.TurnID]; ok && old != q.Model {
			s.models[q.TurnID] = ""
		} else {
			s.models[q.TurnID] = q.Model
		}
	case "token_usage_record":
		s.modernSeen = true
		return s.response(r, r.Payload, false, c)
	case "compacted":
		var q struct {
			Latest json.RawMessage `json:"latest_token_usage_record"`
		}
		if e := json.Unmarshal(r.Payload, &q); e != nil {
			return e
		}
		if len(q.Latest) > 0 && string(q.Latest) != "null" {
			s.modernSeen = true
			return s.response(r, q.Latest, true, c)
		}
	case "event_msg":
		var q struct {
			Type string `json:"type"`
			Info *struct {
				Total json.RawMessage `json:"total_token_usage"`
			} `json:"info"`
		}
		if e := json.Unmarshal(r.Payload, &q); e != nil {
			return e
		}
		if q.Type != "token_count" {
			return nil
		}
		if s.thread == "" {
			c.Add("uncovered_segments", "cumulative snapshot rejected without executing identity")
			return nil
		}
		if q.Info == nil {
			c.Add("uncovered_segments", "token snapshot has no usage")
			return nil
		}
		at, e := recordTime(r)
		if e != nil {
			return fmt.Errorf("invalid usage timestamp")
		}
		v, e := decodeCounters(q.Info.Total, false)
		if e != nil {
			s.baseline = nil
			return e
		}
		s.ordinal++
		rememberCumulativeMaximum(&s.legacyExpected, v)
		if s.baseline == nil {
			s.baseline = &v
			c.Add("baseline_gaps", "initial cumulative snapshot has no prior baseline")
			return nil
		}
		old := *s.baseline
		s.baseline = &v
		if old.equal(v) {
			c.Add("duplicate_snapshots", "")
			return nil
		}
		if !old.monotone(v) {
			c.Add("reset_gaps", "cumulative usage reset or dimension change")
			return nil
		}
		d := Counters{Input: v.Input - old.Input, Output: v.Output - old.Output}
		for _, pair := range []struct {
			old, next *int64
			dest      **int64
		}{{old.CacheRead, v.CacheRead, &d.CacheRead}, {old.CacheWrite, v.CacheWrite, &d.CacheWrite}, {old.Reasoning, v.Reasoning, &d.Reasoning}} {
			if pair.old != nil && pair.next != nil {
				x := *pair.next - *pair.old
				*pair.dest = &x
			}
		}
		if e = d.Validate(); e != nil {
			c.Add("reset_gaps", "inconsistent cumulative subsets")
			return nil
		}
		s.legacy = append(s.legacy, Response{ID: fmt.Sprintf("legacy:%s:%d", s.source, s.ordinal), SourceID: s.source, Agent: "codex", Model: s.model, At: at, Tokens: &d, Origin: "legacy-cumulative"})
	}
	return nil
}
func (s *codexState) response(r record, raw json.RawMessage, embedded bool, c *Coverage) error {
	if s.thread == "" {
		// Reject before any counters, original timestamps or compaction evidence
		// can taint the eventual executing thread. A late header cannot prove
		// which earlier records belong to this execution.
		c.Add("uncovered_segments", "response rejected without executing identity")
		return nil
	}
	var q codexUsage
	if e := json.Unmarshal(raw, &q); e != nil {
		return e
	}
	if q.ResponseID == "" || q.ThreadID == "" || q.TurnID == "" {
		return fmt.Errorf("missing Codex response identity")
	}
	if s.thread != "" && q.ThreadID != s.thread {
		c.Add("copied_history", "")
		return nil
	}
	at, e := recordTime(r)
	if e != nil {
		return fmt.Errorf("invalid response timestamp")
	}
	tokens, e := decodeCounters(q.Usage, false)
	if e != nil {
		return e
	}
	if len(q.ThreadUsage) > 0 && string(q.ThreadUsage) != "null" {
		if expected, err := decodeCounters(q.ThreadUsage, false); err != nil {
			c.Add("uncovered_segments", "invalid thread cumulative evidence")
		} else {
			rememberCumulativeMaximum(&s.expected, expected)
		}
	}
	if len(q.TurnUsage) > 0 && string(q.TurnUsage) != "null" {
		if expected, err := decodeCounters(q.TurnUsage, false); err != nil {
			c.Add("uncovered_segments", "invalid turn cumulative evidence")
		} else {
			if s.turnExpected == nil {
				s.turnExpected = map[string]*Counters{}
			}
			maximum := s.turnExpected[q.TurnID]
			rememberCumulativeMaximum(&maximum, expected)
			s.turnExpected[q.TurnID] = maximum
		}
	}
	if embedded {
		if s.embedded == nil {
			s.embedded = map[string]bool{}
		}
		s.embedded[q.ResponseID] = true
	} else {
		if s.original == nil {
			s.original = map[string]time.Time{}
		}
		if old, ok := s.original[q.ResponseID]; !ok || at.Before(old) {
			s.original[q.ResponseID] = at
		}
	}
	s.modern = append(s.modern, Response{ID: q.ResponseID, SourceID: s.source, Agent: "codex", RequestID: q.TurnID, At: at, Tokens: &tokens, Origin: q.ThreadID, Model: q.TurnID})
	return nil
}

// Cumulative evidence is a lower bound on response coverage, never additional
// spend. Preserve each dimension independently: resets or a later incomparable
// snapshot must not erase an earlier observation of missing responses.
func rememberCumulativeMaximum(maximum **Counters, next Counters) {
	if *maximum == nil {
		*maximum = &Counters{}
	}
	if next.Input > (*maximum).Input {
		(*maximum).Input = next.Input
	}
	if next.Output > (*maximum).Output {
		(*maximum).Output = next.Output
	}
}

func (s *codexState) finish(c *Coverage) []Response {
	if s.thread == "" {
		if c.Counts["unknown_metadata"] == 0 {
			c.Add("unknown_metadata", "Codex executing identity metadata absent")
		}
		c.Add("uncovered_segments", "Codex source has no valid executing identity")
	}
	if !s.modernSeen {
		return s.legacy
	}
	// Cumulative snapshots are diagnostic only in modern files. Their initial
	// baseline gap does not make authoritative response evidence incomplete.
	delete(c.Counts, "baseline_gaps")
	delete(c.Examples, "baseline_gaps")
	for id := range s.embedded {
		if _, ok := s.original[id]; !ok {
			c.Add("reconstructed_compaction", "response reconstructed from compaction")
		}
	}
	for i := range s.modern {
		s.modern[i].Model = s.models[s.modern[i].Model]
	}
	// Keep the original timestamp for embedded duplicates, avoiding a compaction
	// timestamp moving an old response across an ownership/window boundary.
	for i := range s.modern {
		if at, ok := s.original[s.modern[i].ID]; ok {
			s.modern[i].At = at
		}
	}
	// Modern files are never supplemented with snapshot totals. A larger
	// authoritative cumulative counter exposes missing response evidence instead.
	known := Counters{}
	knownTurns := map[string]Counters{}
	seen := map[string]bool{}
	for _, r := range s.modern {
		if seen[r.ID] {
			continue
		}
		seen[r.ID] = true
		var err error
		known.Input, err = checkedAdd(known.Input, r.Tokens.Input)
		if err != nil {
			c.Add("uncovered_segments", "response sum overflows")
			break
		}
		known.Output, err = checkedAdd(known.Output, r.Tokens.Output)
		if err != nil {
			c.Add("uncovered_segments", "response sum overflows")
			break
		}
		// Each turn subtotal is bounded by the checked nonnegative thread sum.
		turn := knownTurns[r.RequestID]
		turn.Input += r.Tokens.Input
		turn.Output += r.Tokens.Output
		knownTurns[r.RequestID] = turn
	}
	// Legacy snapshots have no thread identity. Copied parent history cannot
	// prove that an excess snapshot belongs to the parent rather than this
	// session, so uncertainty stays visible without charging either total.
	for _, expected := range []*Counters{s.expected, s.legacyExpected} {
		if expected != nil && (expected.Input > known.Input || expected.Output > known.Output) {
			c.Add("uncovered_segments", "cumulative counters exceed recorded responses")
			break
		}
	}
	// Another turn's responses cannot reconcile missing spend in this turn.
	// Duplicate/compaction response IDs were excluded from both sums above.
	for turn, expected := range s.turnExpected {
		known := knownTurns[turn]
		if expected.Input > known.Input || expected.Output > known.Output {
			c.Add("uncovered_segments", "turn cumulative counters exceed recorded turn responses")
			break
		}
	}
	return s.modern
}
