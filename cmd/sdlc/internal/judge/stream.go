// stream.go — the reviewer's run as claude's stream-json event stream (#300).
// Plain `claude -p` prints only the final assistant message, so a reviewer
// that writes its verdict and then a short postscript (after a background
// job, say) loses the verdict. The stream carries every message; the review
// text is all of them in order, which lets the existing last-block-wins
// parsers take the latest verdict and findings. The terminal result event
// says, apart from the review's prose, whether the run itself failed.
package judge

import (
	"bytes"
	"encoding/json"
	"strings"
)

// AgentRun is one reviewer run read from its stdout.
type AgentRun struct {
	Stream   bool     // at least one line parsed as a stream event
	Messages []string // each assistant message's text, in order
	Tail     string   // lines that didn't parse (a truncated stream), or the whole output when not a stream
	Result   *Result  // the terminal result event, nil when absent
}

// Result is the stream's terminal result event.
type Result struct {
	Subtype string
	IsError bool
	Text    string
}

type streamEvent struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	IsError bool   `json:"is_error"`
	Result  string `json:"result"`
	Message struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"message"`
}

// ReadStream reads stdout as a stream-json event stream. Lines that don't
// parse are kept, in order, as the run's tail, so a truncated stream loses
// nothing it already wrote; output with no event at all is plain text.
func ReadStream(out []byte) AgentRun {
	var r AgentRun
	var tail []string
	for _, line := range bytes.Split(out, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e streamEvent
		if err := json.Unmarshal(line, &e); err != nil || e.Type == "" {
			tail = append(tail, string(line))
			continue
		}
		r.Stream = true
		switch e.Type {
		case "assistant":
			var parts []string
			for _, c := range e.Message.Content {
				if c.Type == "text" && strings.TrimSpace(c.Text) != "" {
					parts = append(parts, c.Text)
				}
			}
			if len(parts) > 0 {
				r.Messages = append(r.Messages, strings.Join(parts, "\n"))
			}
		case "result":
			r.Result = &Result{Subtype: e.Subtype, IsError: e.IsError, Text: e.Result}
		}
	}
	if !r.Stream {
		return AgentRun{Tail: string(out)}
	}
	r.Tail = strings.Join(tail, "\n")
	return r
}

// Text is the review text: every assistant message in order, then any
// unparsed tail.
func (r AgentRun) Text() string {
	parts := append([]string(nil), r.Messages...)
	if strings.TrimSpace(r.Tail) != "" {
		parts = append(parts, r.Tail)
	}
	return strings.Join(parts, "\n\n")
}

// HasVerdict reports whether text carries a verdict signal any parser
// recognises: a valid ```verdict block, a `VERDICT:` line (plan-quality's
// CLEAN, INFO and FAILURE included), a legacy sentinel Classify still reads,
// or a boundary verdict ParseVerdict finds without a VERDICT: line. It is the
// retry predicate, so one rule serves every recipe: a run is retried only when
// none of them would read anything.
func HasVerdict(text string) bool {
	if _, _, ok := ParseVerdictBlock(text); ok {
		return true
	}
	if _, ok := ParseVerdictToken(text); ok {
		return true
	}
	s := strings.TrimSpace(text)
	return infoRE.MatchString(s) || cleanRE.MatchString(s) || ParseVerdict(text) != VerdictUnknown
}
