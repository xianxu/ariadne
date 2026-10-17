package activetime

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// QueueInterval is a recorded machine-suite wait. Source is assigned only to
// admitted activity from the attempt's pinned owner slot.
type QueueInterval struct {
	Start, End              time.Time
	Source, Slot, AttemptID string
}

// QueueAttribution describes minutes intersecting admitted activity runs.
type QueueAttribution struct {
	RawMinutes, ExcludedMinutes, RetainedWorkMinutes float64
}

type toolKind uint8

const (
	toolUnknown toolKind = iota
	toolWait
	toolWork
	toolUncertainWork
)

// ToolSpan is direct work evidence, or an incomplete work invocation that
// prevents an uncertain queue overlap from becoming a calibrated actual.
type ToolSpan struct {
	Start, End time.Time
	Source     string
	Kind       toolKind
	id         string
}

type pendingTool struct {
	start   time.Time
	kind    toolKind
	invalid bool
	id      string
}

var sessionIDPattern = regexp.MustCompile(`(?i)(?:session|cell) ID\s+([A-Za-z0-9-]+)`)
var nestedCommandPattern = regexp.MustCompile(`tools\.exec_command\(\s*\{\s*cmd\s*:\s*("(?:\\.|[^"\\])*")`)
var nativePollPattern = regexp.MustCompile(`(?s)^\s*(?:// @exec:[^\n]*\n)?\s*(?:text\(\s*|(?:const|let|var)\s+[A-Za-z_$][A-Za-z0-9_$]*\s*=\s*)await\s+tools\.write_stdin\(\s*\{\s*session_id\s*:\s*([0-9]+)`)

// workSpansFromTranscript recognizes only structurally paired native tool
// calls. Polls and yielded commands confer no positive work evidence; yielded
// or otherwise incomplete work is explicit uncertainty for queue attribution.
func workSpansFromTranscript(data []byte, source string) []ToolSpan {
	pending := map[string]*pendingTool{}
	seen := map[string]*pendingTool{}
	invalidIDs := map[string]bool{}
	var spans []ToolSpan
	call := func(id, name, input string, ts time.Time) {
		if id == "" {
			return
		}
		if prior := seen[id]; prior != nil {
			invalidIDs[id] = true
			prior.invalid = true
			return
		}
		p := &pendingTool{start: ts, kind: classifyTool(name, input), id: id}
		seen[id] = p
		pending[id] = p
	}
	result := func(id, output string, ts time.Time) {
		p := pending[id]
		if p == nil {
			return
		}
		delete(pending, id)
		if p.invalid || p.kind == toolUnknown {
			return
		}
		if !ts.After(p.start) || isRunningOutput(output) || isFailedToolOutput(output) {
			if p.kind == toolWork {
				spans = append(spans, ToolSpan{Start: p.start, Source: source, Kind: toolUncertainWork, id: p.id})
			}
			return // a poll cannot prove this command's completion
		}
		spans = append(spans, ToolSpan{Start: p.start, End: ts, Source: source, Kind: p.kind, id: p.id})
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		var d rawLine
		if json.Unmarshal([]byte(line), &d) != nil || d.Timestamp == "" {
			continue
		}
		ts, err := parseISO(d.Timestamp)
		if err != nil {
			continue
		}
		if d.Type == "response_item" {
			var p codexPayload
			if json.Unmarshal(d.Payload, &p) != nil {
				continue
			}
			switch p.Type {
			case "function_call", "custom_tool_call":
				input := p.Arguments
				if input == "" {
					if json.Unmarshal(p.Input, &input) != nil {
						input = string(p.Input)
					}
				}
				if isPoll(p.Name) || nativePollPattern.MatchString(input) && (p.Name == "exec" || p.Name == "functions.exec") {
					call(p.CallID, "poll", input, ts) // reserve ID; never pair poll output to a command
				} else {
					call(p.CallID, p.Name, input, ts)
				}
			case "function_call_output", "custom_tool_call_output":
				result(p.CallID, toolOutputText(p.Output), ts)
			}
			continue
		}
		blocks, ok := decodeBlocks(d.Message.Content)
		if !ok {
			continue
		}
		for _, b := range blocks {
			switch {
			case d.Type == "assistant" && b.Type == "tool_use":
				call(b.ID, b.Name, string(b.Input), ts)
			case d.Type == "user" && b.Type == "tool_result":
				result(b.ToolUseID, string(b.Content), ts)
			}
		}
	}
	var valid []ToolSpan
	for _, span := range spans {
		if !invalidIDs[span.id] {
			valid = append(valid, span)
		}
	}
	for id, p := range pending {
		if p.kind == toolWork && !invalidIDs[id] {
			valid = append(valid, ToolSpan{Start: p.start, Source: source, Kind: toolUncertainWork, id: id})
		}
	}
	for id := range invalidIDs {
		if p := seen[id]; p != nil && p.kind == toolWork {
			valid = append(valid, ToolSpan{Start: p.start, Source: source, Kind: toolUncertainWork, id: id})
		}
	}
	return valid
}

func toolOutputText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var blocks []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) == nil {
		var parts []string
		for _, b := range blocks {
			parts = append(parts, b.Text)
		}
		return strings.Join(parts, "\n")
	}
	return string(raw)
}

func isPoll(name string) bool {
	return name == "write_stdin" || name == "functions.wait" || name == "wait"
}

func isRunningOutput(output string) bool {
	low := strings.ToLower(output)
	return strings.Contains(low, "still running") || strings.Contains(low, "script running") || strings.Contains(low, "process running") || yieldedSessionID(output) != ""
}

func isFailedToolOutput(output string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(output))
	return trimmed == "" || strings.HasPrefix(trimmed, "transport-error") || strings.HasPrefix(trimmed, "unknown process")
}

func yieldedSessionID(output string) string {
	if match := sessionIDPattern.FindStringSubmatch(output); len(match) == 2 && strings.Contains(strings.ToLower(output), "running") {
		return match[1]
	}
	for line := range strings.SplitSeq(output, "\n") {
		if id := yieldedJSONSessionID(line); id != "" {
			return id
		}
	}
	return ""
}

func yieldedJSONSessionID(output string) string {
	var obj map[string]json.RawMessage
	if json.Unmarshal([]byte(output), &obj) != nil || obj["exit_code"] != nil {
		return ""
	}
	for _, field := range []string{"session_id", "cell_id"} {
		if raw := obj[field]; raw != nil && string(raw) != "null" {
			var id string
			if json.Unmarshal(raw, &id) == nil && id != "" {
				return id
			}
			var num json.Number
			if json.Unmarshal(raw, &num) == nil {
				return num.String()
			}
		}
	}
	return ""
}

func classifyTool(name, input string) toolKind {
	if isPoll(name) {
		return toolUnknown
	}
	if name == "apply_patch" || name == "Edit" || name == "Write" || name == "MultiEdit" {
		return toolWork
	}
	if name == "functions.exec" || name == "exec" {
		var args struct {
			Code string `json:"code"`
		}
		code := input
		if json.Unmarshal([]byte(input), &args) == nil && args.Code != "" {
			code = args.Code
		}
		kind := toolUnknown
		for _, match := range nestedCommandPattern.FindAllStringSubmatch(code, -1) {
			var command string
			if json.Unmarshal([]byte(match[1]), &command) != nil {
				continue
			}
			classified := classifyCommand(command)
			if classified == toolWait {
				return toolWait
			}
			if classified == toolWork {
				kind = toolWork
			}
		}
		if kind == toolUnknown && strings.Contains(code, "tools.apply_patch(") {
			return toolWork
		}
		return kind
	}
	if name != "exec_command" && name != "Bash" {
		return toolUnknown
	}
	var args struct {
		Cmd     string `json:"cmd"`
		Command string `json:"command"`
	}
	if json.Unmarshal([]byte(input), &args) != nil {
		return toolUnknown
	}
	command := strings.TrimSpace(args.Cmd)
	if command == "" {
		command = strings.TrimSpace(args.Command)
	}
	return classifyCommand(command)
}

func classifyCommand(command string) toolKind {
	// A cwd wrapper is common in native transcripts. Admit only the trailing
	// executable command, never quoted mentions of a command in echo/log text.
	if strings.HasPrefix(command, "cd ") {
		parts := strings.SplitN(command, "&&", 2)
		if len(parts) != 2 {
			return toolUnknown
		}
		command = strings.TrimSpace(parts[1])
	}
	low := strings.ToLower(command)
	if (strings.HasPrefix(low, "sdlc local-check run ") && strings.Contains(low, "--name suite")) || strings.HasPrefix(low, "sdlc merge ") || low == "sdlc merge" {
		return toolWait
	}
	if low == "sdlc stage --build" || strings.HasPrefix(low, "sdlc stage --build ") || low == "sdlc stage --clear" || strings.HasPrefix(low, "sdlc stage --clear ") {
		return toolWork
	}
	return toolUnknown
}

// bindQueueSources applies a recorded wait to admitted activity sources in its
// pinned slot. No transcript poll or enclosing command supplies queue time.
func bindQueueSources(wait QueueInterval, runs []ActivityRun, slots map[string]string) ([]QueueInterval, error) {
	if wait.AttemptID == "" || wait.Slot == "" || wait.Start.IsZero() || !wait.End.After(wait.Start) {
		return nil, fmt.Errorf("invalid queue interval or identity")
	}
	seen := map[string]bool{}
	var bound []QueueInterval
	for _, run := range runs {
		if !run.Start.Before(wait.End) || !run.End.After(wait.Start) || seen[run.Source] {
			continue
		}
		seen[run.Source] = true
		slot := slots[run.Source]
		if slot == "" {
			return nil, fmt.Errorf("queue attempt %s overlaps source %s without unique slot ownership", wait.AttemptID, run.Source)
		}
		if slot == wait.Slot {
			copy := wait
			copy.Source = run.Source
			bound = append(bound, copy)
		}
	}
	return bound, nil
}

// subtractQueueIdle removes the union of source-bound waits from activity runs,
// preserving only the independently paired work inside each wait. Other source
// runs are unaffected. All intervals are half-open.
func subtractQueueIdle(runs []ActivityRun, waits []QueueInterval, work []ToolSpan) ([]ActivityRun, QueueAttribution) {
	bySource := map[string][]interval{}
	for _, w := range waits {
		bySource[sourceKey(w.Source)] = append(bySource[sourceKey(w.Source)], interval{w.Start, w.End})
	}
	for source, ivs := range bySource {
		bySource[source] = mergeIntervals(ivs)
	}
	workBySource := map[string][]interval{}
	for _, w := range work {
		if w.Kind == toolWork && w.End.After(w.Start) {
			workBySource[sourceKey(w.Source)] = append(workBySource[sourceKey(w.Source)], interval{w.Start, w.End})
		}
	}
	for source, ivs := range workBySource {
		workBySource[source] = mergeIntervals(ivs)
	}
	var out []ActivityRun
	var stats QueueAttribution
	// Raw queue time belongs to attempts, not to the number of concurrent
	// transcript sources from the owning slot.
	raw := map[string]interval{}
	for _, w := range waits {
		raw[w.AttemptID] = interval{w.Start, w.End}
	}
	var rawIntervals []interval
	for _, iv := range raw {
		rawIntervals = append(rawIntervals, iv)
	}
	for _, iv := range mergeIntervals(rawIntervals) {
		stats.RawMinutes += iv.e.Sub(iv.s).Minutes()
	}
	for _, run := range runs {
		if !run.End.After(run.Start) {
			continue
		}
		pieces := []interval{{run.Start, run.End}}
		for _, wait := range bySource[sourceKey(run.Source)] {
			overlap := intersect(interval{run.Start, run.End}, wait)
			if !overlap.e.After(overlap.s) {
				continue
			}
			for _, active := range workBySource[sourceKey(run.Source)] {
				kept := intersect(overlap, active)
				if kept.e.After(kept.s) {
					stats.RetainedWorkMinutes += kept.e.Sub(kept.s).Minutes()
				}
			}
			idle := subtractIntervals([]interval{overlap}, workBySource[sourceKey(run.Source)])
			for _, gap := range idle {
				stats.ExcludedMinutes += gap.e.Sub(gap.s).Minutes()
				pieces = subtractIntervals(pieces, []interval{gap})
			}
		}
		for _, part := range pieces {
			copy := run
			copy.Start, copy.End, copy.Active = part.s, part.e, part.e.Sub(part.s).Minutes()
			out = append(out, copy)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Start.Equal(out[j].Start) {
			return out[i].Source < out[j].Source
		}
		return out[i].Start.Before(out[j].Start)
	})
	return out, stats
}

func intersect(a, b interval) interval {
	if a.s.Before(b.s) {
		a.s = b.s
	}
	if a.e.After(b.e) {
		a.e = b.e
	}
	return a
}

func mergeIntervals(ivals []interval) []interval {
	sort.Slice(ivals, func(i, j int) bool { return ivals[i].s.Before(ivals[j].s) })
	var out []interval
	for _, iv := range ivals {
		if !iv.e.After(iv.s) {
			continue
		}
		if len(out) == 0 || iv.s.After(out[len(out)-1].e) {
			out = append(out, iv)
		} else if iv.e.After(out[len(out)-1].e) {
			out[len(out)-1].e = iv.e
		}
	}
	return out
}

func subtractIntervals(parts, cuts []interval) []interval {
	for _, cut := range cuts {
		var next []interval
		for _, p := range parts {
			x := intersect(p, cut)
			if !x.e.After(x.s) {
				next = append(next, p)
				continue
			}
			if x.s.After(p.s) {
				next = append(next, interval{p.s, x.s})
			}
			if p.e.After(x.e) {
				next = append(next, interval{x.e, p.e})
			}
		}
		parts = next
	}
	return parts
}
