// codex.go — codex rollout ingestion for the friction audit (#172 M3).
//
// ALL format knowledge here derives from the single spec shared with Python
// introspect: atlas/workflow/introspect.md → "Codex transcript format" (the DRY
// point is the spec, not shared code — keep them in lockstep). The essentials:
//
//   - one JSONL per session under ~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl;
//     every line is {timestamp, type, payload}.
//   - sdlc invocations ride response_item/function_call — `arguments` is a
//     JSON-ENCODED STRING carrying the command under `cmd` (a plain string, the
//     same shape agent_codex.py reads); output links by call_id via
//     response_item/function_call_output, whose `output` is a plain string. sdlc's
//     ANSI is unconditional, so the SAME classifyOutputLine reads codex output.
//   - ⚠️ fork-replay rollouts (`forked_from_id` on the FIRST session_meta) REPLAY
//     the parent's transcript and carry two session_meta — key off the FIRST and
//     SKIP the file, or every shared event double-counts (the spec's 66%-inflation
//     trap: skip 40, NOT 119). Sub-agent threads (parent_thread_id/agent_nickname
//     WITHOUT forked_from_id) have their own content and are processed.
//
// The first-session_meta loop mirrors transcripts/codex.go codexCWDFromBytes
// (cwd-only, unexported — the fork fields are extracted net-new here, per plan).
package processmanual

import (
	"bufio"
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"time"
)

type codexMetaKind int

const (
	codexRoot       codexMetaKind = iota
	codexForkReplay               // forked_from_id set — replays parent, SKIP
	codexSubAgent                 // parent_thread_id/agent_nickname, no fork — process
)

// codexMeta decodes the FIRST session_meta in a rollout (later ones belong to a
// replayed parent). ok=false when the file carries none (not a session rollout).
func codexMeta(data []byte) (kind codexMetaKind, cwd string, ok bool) {
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var r struct {
			Type    string `json:"type"`
			Payload struct {
				CWD            string `json:"cwd"`
				ForkedFromID   string `json:"forked_from_id"`
				ParentThreadID string `json:"parent_thread_id"`
				AgentNickname  string `json:"agent_nickname"`
			} `json:"payload"`
		}
		if json.Unmarshal(line, &r) != nil {
			continue
		}
		if r.Type != "session_meta" {
			continue
		}
		switch {
		case r.Payload.ForkedFromID != "":
			return codexForkReplay, r.Payload.CWD, true
		case r.Payload.ParentThreadID != "" || r.Payload.AgentNickname != "":
			return codexSubAgent, r.Payload.CWD, true
		default:
			return codexRoot, r.Payload.CWD, true
		}
	}
	return codexRoot, "", false
}

// codexExitRE reads the exec_command wrapper's exit line ("Process exited with
// code N"). A non-zero exit derives SdlcInvocation.Failed — did the command
// complete? — for the firing-order ladder. NOTE: this is deliberately NOT the
// atlas spec's taste-friction `is_error` gate (which additionally requires a
// FRICTION_HINT); that gate classifies friction *moments*, this field records
// command completion.
var codexExitRE = regexp.MustCompile(`Process exited with code (\d+)`)

func codexOutputFailed(output string) bool {
	m := codexExitRE.FindStringSubmatch(output)
	if m == nil {
		return false
	}
	n, err := strconv.Atoi(m[1])
	return err == nil && n != 0
}

// parseCodexInvocations extracts every `sdlc <verb>` exec from a codex rollout,
// joined to its call_id-linked output — the codex sibling of scanTranscript.
// Fork-replay rollouts return nil (skipped entirely). Codex has no Skill tool and
// its file edits (patch_apply_end) have no consumer here, so no ActivityMarks —
// the skill-late arm is Claude-only (stated in the report footer).
func parseCodexInvocations(data []byte, validVerbs map[string]bool) []SdlcInvocation {
	kind, _, ok := codexMeta(data)
	if !ok || kind == codexForkReplay {
		return nil
	}

	type pending struct {
		inv SdlcInvocation
		id  string
	}
	var pend []pending
	outByID := map[string]string{}

	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var r struct {
			Timestamp time.Time `json:"timestamp"`
			Type      string    `json:"type"`
			Payload   struct {
				Type      string `json:"type"`
				Arguments string `json:"arguments"`
				CallID    string `json:"call_id"`
				Output    string `json:"output"`
			} `json:"payload"`
		}
		if json.Unmarshal(line, &r) != nil || r.Type != "response_item" {
			continue
		}
		switch r.Payload.Type {
		case "function_call":
			var args struct {
				Cmd string `json:"cmd"`
			}
			if json.Unmarshal([]byte(r.Payload.Arguments), &args) != nil || args.Cmd == "" {
				continue
			}
			m := sdlcVerbRE.FindStringSubmatch(args.Cmd)
			if m == nil || !validVerbs[m[1]] {
				continue
			}
			pend = append(pend, pending{
				inv: SdlcInvocation{
					Verb: m[1], Command: args.Cmd, Time: r.Timestamp,
					IssueID: parseIssueID(args.Cmd),
					IsHelp:  bytes.Contains([]byte(args.Cmd), []byte("--help")),
					Agent:   "codex",
				},
				id: r.Payload.CallID,
			})
		case "function_call_output":
			if r.Payload.CallID != "" && r.Payload.Output != "" {
				outByID[r.Payload.CallID] = r.Payload.Output
			}
		}
	}

	out := make([]SdlcInvocation, 0, len(pend))
	for _, p := range pend {
		if o, ok := outByID[p.id]; ok {
			p.inv.Output = o
			p.inv.Failed = codexOutputFailed(o)
		}
		out = append(out, p.inv)
	}
	return out
}
