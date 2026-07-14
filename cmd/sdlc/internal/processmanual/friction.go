package processmanual

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// SdlcInvocation is one observed `Bash(sdlc <verb> …)` tool call plus its linked
// result output — the atom the friction detectors read. Anchoring detection to these
// (vs any line mentioning a flag) is what drops the source/log-read contamination
// that saturates this repo's transcripts. Transcript/Agent/Repo are set by the
// corpus walk (M1 Task 4 / M3).
type SdlcInvocation struct {
	Verb    string
	Command string // full bash command (for --force / --issue parsing)
	IssueID string // from "--issue N" / "#N" in the command ("" if absent)
	Output  string // linked tool_result stdout
	Time    time.Time
	IsHelp  bool // `sdlc <verb> --help` — its output lists every flag; excluded
	// Failed: the command did not complete — Claude's tool_result is_error flag /
	// codex's non-zero "Process exited with code N". NOT the taste-friction
	// is_error gate (atlas spec); used so failed invocations don't raise the
	// firing-order ladder.
	Failed     bool
	Transcript string
	Agent      string
	Repo       string
}

// issueArgRE accepts ONLY the explicit `--issue N` form. A bare-`#N` fallback
// was tried and removed (M4): commit/stash messages inside compound commands
// carry `#N` constantly, and a `git stash -m "… #145" && sdlc merge` mis-keyed
// an unrelated merge onto #145's ladder (a live false anomaly). The spine verbs
// all take --issue; merge/push carry none and are attributed from segment
// context instead (precision over recall).
var issueArgRE = regexp.MustCompile(`--issue[ =]+0*(\d+)`)

func parseIssueID(command string) string {
	if m := issueArgRE.FindStringSubmatch(command); m != nil {
		return m[1]
	}
	return ""
}

// toolResultText extracts a tool_result block's content, which is polymorphic — a
// plain string, or an array of {type:"text", text:…} parts (older transcripts).
func toolResultText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var parts []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) == nil {
		var b strings.Builder
		for _, p := range parts {
			b.WriteString(p.Text)
			b.WriteByte('\n')
		}
		return b.String()
	}
	return ""
}

// ActivityMark is a non-sdlc transcript event the firing-order detector consumes:
// a file edit (Edit/Write/MultiEdit → KindFileEdit, deferred here from M1 Task 3)
// or a Skill load. Captured by scanTranscript alongside the anchored invocations;
// Transcript/Repo are stamped by the corpus walk.
type ActivityMark struct {
	Kind       Kind   // KindFileEdit | KindSkill
	Detail     string // file path / skill name
	Time       time.Time
	Transcript string
	Repo       string
}

// scanTranscript extracts every `Bash(sdlc <verb>)` call from a Claude transcript,
// joined to its `tool_use_id`-linked result output, plus the ActivityMarks (file
// edits + Skill loads) the firing-order detector needs. Pure over bytes; reuses the
// same scan/linkage shape as parseEvents but yields SdlcInvocations (verb + args +
// output) rather than the per-session FiredEvents (verb + verdict), because the
// friction audit needs the raw output parseEvents discards. `validVerbs` gates on
// the real verb set so a prose "sdlc foo" isn't counted.
func scanTranscript(data []byte, validVerbs map[string]bool) ([]SdlcInvocation, []ActivityMark) {
	type pending struct {
		inv SdlcInvocation
		id  string
	}
	var pend []pending
	var marks []ActivityMark
	outByID := map[string]string{}
	errByID := map[string]bool{}

	// Error deliberately discarded: a >16MB line truncates the rest of THAT
	// transcript only — for a whole-corpus aggregate, partial data from one
	// pathological file beats failing the walk (parseEvents, single-session,
	// does propagate it).
	_ = forEachRec(data, func(r rec) {
		switch r.Type {
		case "assistant":
			for _, c := range r.Message.Content {
				if c.Type != "tool_use" {
					continue
				}
				if c.Name != "Bash" {
					// Skill loads + file edits mark the activity stream for the
					// skill-late arm; classifyToolUse is the shared match table.
					if kind, detail, ok := classifyToolUse(c.Name, c.Input, nil); ok &&
						(kind == KindSkill || kind == KindFileEdit) {
						marks = append(marks, ActivityMark{Kind: kind, Detail: detail, Time: r.Timestamp})
					}
					continue
				}
				var in struct {
					Command string `json:"command"`
				}
				if json.Unmarshal(c.Input, &in) != nil {
					continue
				}
				m := sdlcVerbRE.FindStringSubmatch(in.Command)
				if m == nil || !validVerbs[m[1]] {
					continue
				}
				pend = append(pend, pending{
					inv: SdlcInvocation{
						Verb: m[1], Command: in.Command, Time: r.Timestamp,
						IssueID: parseIssueID(in.Command),
						IsHelp:  bytes.Contains([]byte(in.Command), []byte("--help")),
					},
					id: c.ID,
				})
			}
		case "user":
			for _, c := range r.Message.Content {
				if c.Type == "tool_result" && c.ToolUseID != "" {
					// The tool_result BLOCK content is the merged displayed output
					// (stdout+stderr) — more complete than toolUseResult.stdout, which
					// misses stderr-only ACKs like the cinfo no-judge skip (verified:
					// no-judge ACK in content-block 105× vs stdout 49×). #172.
					if txt := toolResultText(c.Result); txt != "" {
						outByID[c.ToolUseID] = txt
					}
					if c.IsError {
						errByID[c.ToolUseID] = true
					}
				}
			}
		}
	})

	out := make([]SdlcInvocation, 0, len(pend))
	for _, p := range pend {
		if sd, ok := outByID[p.id]; ok {
			p.inv.Output = sd
		}
		p.inv.Failed = errByID[p.id]
		out = append(out, p.inv)
	}
	return out, marks
}

// forEachRec is THE shared Claude-transcript scan core — parseEvents (the
// --session report) and scanTranscript (the friction audit) both iterate through
// it, so the two walkers can't drift on scanner buffer sizing or line tolerance
// (#172 M1 review watch item). Malformed lines are skipped, not fatal.
func forEachRec(data []byte, fn func(rec)) error {
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024) // transcript lines can be large
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var r rec
		if json.Unmarshal(line, &r) != nil {
			continue
		}
		fn(r)
	}
	return sc.Err()
}

// GateEvent is one classified bypass-ACK or gate-refusal observed in an sdlc
// invocation's output (#172 friction audit).
type GateEvent struct {
	Kind          GateEventKind
	Gate          string // flag name, e.g. "no-atlas"
	Command       string // the sdlc verb whose output this came from
	ViaForce      bool   // bypassed via --force (vs the specific --no-<gate>)
	Observability Observability
}

type GateEventKind int

const (
	GateBypass GateEventKind = iota
	GateRefusal
)

// Observability records how completely a gate's events can be measured.
type Observability int

const (
	ObsFull        Observability = iota // ACK/refusal is emitted and names the flag
	ObsForceOnly                        // change-code: bypass observable only via --force (silent alone)
	ObsFlagOmitted                      // merge/push: refusal never names the flag (best-effort attribution)
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// contamination markers: a line bearing any of these is source code, a cat-n file
// read, or a format-string template — NOT a real runtime ACK/refusal.
var (
	catnPrefixRE     = regexp.MustCompile(`^\s*\d+\t`) // "944\t…" cat-n read
	contamSubstrings = []string{
		"%s", "%d", "%q", "%v", // unexpanded format verbs
		"cwarn(", "cinfo(", "cok(", "Sprintf(", "Fprint", "append(", "stderr,",
	}
)

func isContamination(line string) bool {
	if catnPrefixRE.MatchString(line) {
		return true
	}
	for _, m := range contamSubstrings {
		if strings.Contains(line, m) {
			return true
		}
	}
	return false
}

// classifyOutputLine classifies ONE line of an sdlc invocation's output as a gate
// bypass, a gate refusal, or neither, given the invocation's verb. Anchoring to the
// verb (plus the runtime-reset requirement for ACKs and the grammar-anchored refusal
// patterns) is what separates a real event from the source/log-read contamination
// that saturates this repo's transcripts.
func classifyOutputLine(line, verb string) (GateEvent, bool) {
	if isContamination(line) {
		return GateEvent{}, false
	}
	hasReset := strings.Contains(line, "\x1b[0m ") // runtime cwarn/cinfo/cok marker
	stripped := ansiRE.ReplaceAllString(line, "")

	for i := range GateCatalog {
		g := &GateCatalog[i]
		if !contains(g.Commands, verb) {
			continue
		}
		// Bypass ACK — requires the runtime reset (source restatements lack it).
		if hasReset && g.ackRE != nil && g.ackRE.MatchString(stripped) {
			return GateEvent{
				Kind: GateBypass, Gate: g.Flag, Command: verb,
				ViaForce:      g.Grammar == grammarG2 || strings.Contains(stripped, "--force:"),
				Observability: gateObs(g),
			}, true
		}
		// Refusal — grammar+digit-anchored, NOT reset-gated (runtime refusals are
		// plain strings). The exact per-gate pattern rejects the warmup twin.
		if g.HasRefusal && g.refusalRE != nil && g.refusalRE.MatchString(stripped) {
			return GateEvent{
				Kind: GateRefusal, Gate: g.Flag, Command: verb,
				Observability: gateObs(g),
			}, true
		}
	}
	return GateEvent{}, false
}

// gateObs is a (command, flag)'s intrinsic measurement caveat — a property of the
// gate, not of which event type happened to be seen, so the caveat shows even when
// e.g. only refusals were observed for a change-code force-only gate. The two
// non-full caveats are mutually exclusive per gate: change-code gates are
// SilentAlone (bypass observable only via --force); merge/push no-judge refusals
// don't name the flag (best-effort attribution). Everything else is fully observable.
func gateObs(g *GateSig) Observability {
	if g.SilentAlone {
		return ObsForceOnly
	}
	if g.HasRefusal && !g.RefusalNamesFlag {
		return ObsFlagOmitted
	}
	return ObsFull
}

// invocationGateEvents classifies one invocation's output lines and collapses
// duplicate events per (kind, gate, command): a single refusal can emit two
// matching lines — no-validate prints the validategate cwarn AND the die-wrapped
// returned error — which would double-count refusals and skew M2's
// refusal→retry resolution rates (#172 M1 review Minor).
func invocationGateEvents(inv SdlcInvocation) []GateEvent {
	type evKey struct {
		kind      GateEventKind
		gate, cmd string
	}
	seen := map[evKey]bool{}
	var out []GateEvent
	for _, ln := range strings.Split(inv.Output, "\n") {
		ev, ok := classifyOutputLine(ln, inv.Verb)
		if !ok {
			continue
		}
		k := evKey{ev.Kind, ev.Gate, ev.Command}
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, ev)
	}
	return out
}

// ── Refusal→retry pairing (M2 Task 5) ────────────────────────────────────────

// RefusalRetry pairs one observed gate refusal with the agent's next attempt —
// the next invocation of the same verb + same issue in the same transcript.
// Resolved = the retry no longer refused this gate; ViaBypass distinguishes
// routing AROUND the gate (a bypass ACK on the retry) from satisfying it. For
// merge/push refusals that never name their flag, pairing is the same
// verb+context walk but Observability carries the flag-omitted caveat so the
// report can mark the attribution best-effort.
type RefusalRetry struct {
	Gate          string `json:"gate"`
	Command       string `json:"command"`
	IssueID       string `json:"issue_id,omitempty"`
	Repo          string `json:"repo,omitempty"`
	Agent         string `json:"agent,omitempty"` // claude | codex — M4's triage wants the split
	Retried       bool   `json:"retried"`
	Resolved      bool   `json:"resolved"`
	ViaBypass     bool   `json:"via_bypass"`
	Observability string `json:"observability"`
}

// allGateEvents classifies every invocation ONCE — the single events stream all
// three consumers (aggregate, detectRefusalRetries, detectFiringOrder) read, so
// there is one source of classification (#172 M2 review Minor).
func allGateEvents(invs []SdlcInvocation) [][]GateEvent {
	events := make([][]GateEvent, len(invs))
	for i := range invs {
		events[i] = invocationGateEvents(invs[i])
	}
	return events
}

// detectRefusalRetries is pure over the parsed invocations, which arrive in
// transcript order (the corpus walk appends per transcript chronologically).
// events is the precomputed allGateEvents stream, index-aligned with invs.
func detectRefusalRetries(invs []SdlcInvocation, events [][]GateEvent) []RefusalRetry {
	byTranscript := map[string][]int{}
	var order []string
	for i, inv := range invs {
		if _, ok := byTranscript[inv.Transcript]; !ok {
			order = append(order, inv.Transcript)
		}
		byTranscript[inv.Transcript] = append(byTranscript[inv.Transcript], i)
	}
	var out []RefusalRetry
	for _, t := range order {
		idxs := byTranscript[t]
		for pi, i := range idxs {
			for _, ev := range events[i] {
				if ev.Kind != GateRefusal {
					continue
				}
				rr := RefusalRetry{
					Gate: ev.Gate, Command: ev.Command,
					IssueID: invs[i].IssueID, Repo: invs[i].Repo, Agent: invs[i].Agent,
					Observability: obsString(ev.Observability),
				}
				if hasGateEvent(events[i], GateBypass, ev.Gate) {
					// A compound command refused then bypassed within ONE invocation
					// (`sdlc close … || sdlc close --no-atlas …`) — resolved in place.
					rr.Retried, rr.Resolved, rr.ViaBypass = true, true, true
				} else {
					for _, j := range idxs[pi+1:] {
						if invs[j].Verb != invs[i].Verb || invs[j].IssueID != invs[i].IssueID {
							continue
						}
						rr.Retried = true
						rr.Resolved = !hasGateEvent(events[j], GateRefusal, ev.Gate)
						rr.ViaBypass = hasGateEvent(events[j], GateBypass, ev.Gate)
						break
					}
				}
				out = append(out, rr)
			}
		}
	}
	return out
}

// ── Firing-order detector (M2 Task 6) ────────────────────────────────────────

// workflowStage orders the workflow verbs per AGENTS.md §2's flow:
// claim ≺ start-plan ≺ change-code ≺ milestone-close ≺ close ≺ merge
// (push is merge-without-PR — the same publish stage).
var workflowStage = map[string]int{
	"claim": 0, "start-plan": 1, "change-code": 2,
	"milestone-close": 3, "close": 4, "merge": 5, "push": 5,
}

// lateSkillRE matches the plan/TDD skills whose load AFTER implementation edits
// signals planning-arrived-late (matched on the suffix so adapted/prefixed skill
// names still hit).
var lateSkillRE = regexp.MustCompile(`(writing-plans|test-driven-development)$`)

// FiringOrderAnomaly is one detected workflow-order violation (#172 M2).
type FiringOrderAnomaly struct {
	Kind    string `json:"kind"` // change-code-after-close | skill-late
	Verb    string `json:"verb,omitempty"`
	IssueID string `json:"issue_id,omitempty"`
	Repo    string `json:"repo,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

// FiringOrderResult carries the anomalies plus the honesty counter: merge/push
// invocations with no attributable issue context are COUNTED, never bucketed
// under a global "" ladder that would cross-contaminate issues.
type FiringOrderResult struct {
	Anomalies           []FiringOrderAnomaly `json:"anomalies"`
	UnattributedPublish int                  `json:"unattributed_publish"`
}

// detectFiringOrder walks each (repo, issue)'s invocations (events precomputed,
// index-aligned) in time order against
// the workflowStage ladder, iteration-aware: the legal loops that must NOT flag
// are milestone-close→change-code (next milestone), start-plan re-runs
// (AGENTS.md: "re-run per design"), and close→change-code/start-plan after a
// REWORK verdict (codecomplete→working reopen, issue.cue). Only an observed
// ORDER INVERSION flags — change-code after a clean close/merge; the mere
// absence of earlier stages is treated as partial observation (sessions predate
// the corpus or ran on another agent), not an anomaly (precision over recall).
// merge/push carry no --issue → attributed from segment context (the nearest
// preceding --issue invocation in the same transcript within the gap boundary)
// or counted unattributed. The skill-late arm flags a plan/TDD Skill load after
// a non-doc file edit in the same segment+issue.
func detectFiringOrder(invs []SdlcInvocation, events [][]GateEvent, marks []ActivityMark) FiringOrderResult {
	var res FiringOrderResult

	// 1. Attribute merge/push from segment context. --help invocations are not
	//    workflow steps (their output lists every flag; running them moves nothing).
	effIssue := make([]string, len(invs))
	lastIssue := map[string]string{}
	lastIssueTime := map[string]time.Time{}
	for i, inv := range invs {
		if inv.IsHelp {
			continue
		}
		effIssue[i] = inv.IssueID
		if inv.IssueID != "" {
			lastIssue[inv.Transcript] = inv.IssueID
			lastIssueTime[inv.Transcript] = inv.Time
			continue
		}
		if inv.Verb != "merge" && inv.Verb != "push" {
			continue
		}
		if li := lastIssue[inv.Transcript]; li != "" && inv.Time.Sub(lastIssueTime[inv.Transcript]) <= gapBoundary {
			effIssue[i] = li
		} else {
			res.UnattributedPublish++
		}
	}

	// 2. Per-(repo, issue) ladder, merged across transcripts in time order (an
	//    issue's claim/change-code/close usually span several sessions).
	type issueKey struct{ repo, issue string }
	seqs := map[issueKey][]int{}
	var keys []issueKey
	for i, inv := range invs {
		if inv.IsHelp || effIssue[i] == "" {
			continue
		}
		if _, ok := workflowStage[inv.Verb]; !ok {
			continue
		}
		k := issueKey{inv.Repo, effIssue[i]}
		if _, ok := seqs[k]; !ok {
			keys = append(keys, k)
		}
		seqs[k] = append(seqs[k], i)
	}
	for _, k := range keys {
		idxs := seqs[k]
		sort.SliceStable(idxs, func(a, b int) bool { return invs[idxs[a]].Time.Before(invs[idxs[b]].Time) })
		maxStage := -1
		flagged := false
		for _, i := range idxs {
			inv := invs[i]
			if inv.Verb == "change-code" && maxStage >= workflowStage["close"] && !flagged {
				detail := "change-code after close (no REWORK observed)"
				if maxStage >= workflowStage["merge"] {
					detail = "change-code after merge/push"
				}
				res.Anomalies = append(res.Anomalies, FiringOrderAnomaly{
					Kind: "change-code-after-close", Verb: inv.Verb,
					IssueID: k.issue, Repo: k.repo, Detail: detail,
				})
				flagged = true // once per issue — repeated regressions are one finding
			}
			if (inv.Verb == "close" || inv.Verb == "milestone-close") && isReworkVerdict(inv.Output) {
				// REWORK reopens (codecomplete→working): roll the ladder back to the
				// implementing stage instead of raising it. Checked BEFORE the
				// failure skip — a REWORK close also reads as failed, but it must
				// still roll an already-raised ladder back.
				if maxStage > workflowStage["change-code"] {
					maxStage = workflowStage["change-code"]
				}
				continue
			}
			if inv.Failed || (anyGateEvent(events[i], GateRefusal) && !anyGateEvent(events[i], GateBypass)) {
				// A failed or gate-REFUSED invocation did not cross its boundary — a
				// refused close followed by change-code is legal recovery, not an
				// inversion (M2 boundary review, Important #1). Failed covers the
				// non-gate failures (dirty tree, no claim) via Claude's is_error /
				// codex's non-zero exit (M3 hardening); a refusal WITH a bypass ACK
				// is the compound retry (`close || close --no-X`) — that completed.
				continue
			}
			if s := workflowStage[inv.Verb]; s > maxStage {
				maxStage = s
			}
		}
	}

	// 3. skill-late — per transcript, over the merged invocation+mark stream.
	type tev struct {
		time  time.Time
		issue string
		mark  *ActivityMark
	}
	byT := map[string][]tev{}
	var tOrder []string
	add := func(t string, e tev) {
		if _, ok := byT[t]; !ok {
			tOrder = append(tOrder, t)
		}
		byT[t] = append(byT[t], e)
	}
	for i, inv := range invs {
		if inv.IsHelp {
			continue
		}
		add(inv.Transcript, tev{time: inv.Time, issue: effIssue[i]})
	}
	for i := range marks {
		add(marks[i].Transcript, tev{time: marks[i].Time, mark: &marks[i]})
	}
	for _, t := range tOrder {
		evs := byT[t]
		sort.SliceStable(evs, func(a, b int) bool { return evs[a].time.Before(evs[b].time) })
		var curIssue string
		var editSeen bool
		var last time.Time
		for _, e := range evs {
			if !last.IsZero() && e.time.Sub(last) > gapBoundary {
				editSeen = false // new segment
			}
			last = e.time
			switch {
			case e.mark == nil:
				if e.issue != "" && e.issue != curIssue {
					curIssue = e.issue
					editSeen = false // earlier edits belonged to the previous issue's work
				}
			case e.mark.Kind == KindFileEdit:
				if !strings.HasSuffix(e.mark.Detail, ".md") {
					// .md edits (plans/issues/docs) ARE design work; only
					// implementation edits make a later plan-skill load "late".
					editSeen = true
				}
			case e.mark.Kind == KindSkill && lateSkillRE.MatchString(e.mark.Detail):
				if editSeen {
					res.Anomalies = append(res.Anomalies, FiringOrderAnomaly{
						Kind: "skill-late", IssueID: curIssue, Repo: e.mark.Repo,
						Detail: e.mark.Detail,
					})
					editSeen = false // one finding per late load, not per subsequent load
				}
			}
		}
	}
	return res
}

// isReworkVerdict recovers a REWORK boundary-review verdict from a close/
// milestone-close invocation's output — judge.ParseVerdict is the exact parser
// `sdlc close` itself uses (ARCH-DRY), with the trailer fallback for re-closes.
func isReworkVerdict(output string) bool {
	v := judge.ParseVerdict(output)
	if v == judge.VerdictUnknown {
		v = judge.ParseVerdictTrailer(output)
	}
	return v == judge.VerdictRework
}

func hasGateEvent(evs []GateEvent, kind GateEventKind, gate string) bool {
	for _, e := range evs {
		if e.Kind == kind && e.Gate == gate {
			return true
		}
	}
	return false
}

func anyGateEvent(evs []GateEvent, kind GateEventKind) bool {
	for _, e := range evs {
		if e.Kind == kind {
			return true
		}
	}
	return false
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// ── Corpus walk + aggregation + report (M1 Task 4) ───────────────────────────

// SpineVerbs is the set of sdlc verbs that carry bypass gates — the only verbs the
// friction audit needs to recognize. Derived from the catalog so it can't drift.
func SpineVerbs() map[string]bool {
	out := map[string]bool{}
	for _, g := range GateCatalog {
		for _, c := range g.Commands {
			out[c] = true
		}
	}
	return out
}

type transcriptRef struct{ Path, Repo, Agent string }

// repoLabel maps a ~/.claude/projects slug to a repo label for per-repo grouping,
// normalizing ariadne worktrees to "ariadne" and excluding scratch/temp dirs
// (include=false). Slug→repo is inherently lossy (dash-joined); this is best-effort
// grouping for the peer-vs-ariadne concentration finding.
func repoLabel(slug string) (label string, include bool) {
	if strings.HasPrefix(slug, "-private-tmp-") || strings.HasPrefix(slug, "-private-var-folders-") {
		return "", false
	}
	if strings.Contains(slug, "-worktree-ariadne-") {
		return "ariadne", true
	}
	if i := strings.Index(slug, "-workspace-"); i >= 0 {
		return slug[i+len("-workspace-"):], true
	}
	return strings.TrimPrefix(slug, "-"), true
}

// repoLabelFromPath maps a codex session_meta.cwd (a real path, unlike the
// dash-joined Claude slug) to a repo label — the codex counterpart of repoLabel.
// Worktree checkouts normalize to their repo; scratch/temp dirs are excluded.
func repoLabelFromPath(cwd string) (label string, include bool) {
	if cwd == "" || strings.HasPrefix(cwd, "/tmp/") || strings.HasPrefix(cwd, "/private/tmp") ||
		strings.HasPrefix(cwd, "/private/var/folders") || strings.HasPrefix(cwd, "/var/folders") {
		return "", false
	}
	first := func(rest string) string {
		if j := strings.IndexByte(rest, '/'); j >= 0 {
			return rest[:j]
		}
		return rest
	}
	if i := strings.Index(cwd, "/workspace/worktree/"); i >= 0 {
		return first(cwd[i+len("/workspace/worktree/"):]), true
	}
	if i := strings.Index(cwd, "/workspace/"); i >= 0 {
		return first(cwd[i+len("/workspace/"):]), true
	}
	return filepath.Base(cwd), true
}

// enumerateCodexTranscripts walks ~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl
// under codexRoot. Repo labels need the file's session_meta.cwd, so this returns
// paths only; the walk labels (and fork-skips) per file.
func enumerateCodexTranscripts(codexRoot string) []string {
	var out []string
	_ = filepath.WalkDir(codexRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // tolerant: unreadable subtree is skipped
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), "rollout-") && strings.HasSuffix(d.Name(), ".jsonl") {
			out = append(out, path)
		}
		return nil
	})
	return out
}

// enumerateClaudeTranscripts walks every ~/.claude/projects/<slug>/*.jsonl under
// claudeRoot (injectable for tests). The IO seam; the pure detectors receive the
// parsed invocations.
func enumerateClaudeTranscripts(claudeRoot string) []transcriptRef {
	entries, err := os.ReadDir(claudeRoot)
	if err != nil {
		return nil
	}
	var out []transcriptRef
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		label, include := repoLabel(e.Name())
		if !include {
			continue
		}
		files, _ := filepath.Glob(filepath.Join(claudeRoot, e.Name(), "*.jsonl"))
		for _, f := range files {
			out = append(out, transcriptRef{Path: f, Repo: label, Agent: "claude"})
		}
	}
	return out
}

// GateStat is one (command, gate) pair's aggregated friction. Keyed per command —
// NOT per flag alone — because a gate like no-judge spans five commands with three
// different observabilities (close/mclose full, change-code force-only, merge/push
// flag-omitted); collapsing to the flag mislabels the honesty column (#172 M1 review).
type GateStat struct {
	Command       string `json:"command"`
	Flag          string `json:"flag"`
	Bypasses      int    `json:"bypasses"`
	Refusals      int    `json:"refusals"`
	Observability string `json:"observability"`
}

// FrictionReport is the whole-corpus aggregate (#172): per-gate bypass/refusal
// counts + per-repo bypass concentration, plus the M2 detectors — refusal→retry
// pairing and firing-order anomalies. Codex coverage is M3.
type FrictionReport struct {
	TranscriptsScanned int               `json:"transcripts_scanned"`
	InvocationsSeen    int               `json:"invocations_seen"`
	Gates              []GateStat        `json:"gates"`
	ByRepoBypass       map[string]int    `json:"by_repo_bypass"`
	ByAgentBypass      map[string]int    `json:"by_agent_bypass"`
	CodexForksSkipped  int               `json:"codex_forks_skipped"`
	RefusalRetries     []RefusalRetry    `json:"refusal_retries"`
	FiringOrder        FiringOrderResult `json:"firing_order"`
}

// buildFrictionReport composes the pure detectors over one parsed corpus — the
// single seam RunFrictionReport feeds (both agents' invocations merged). Events
// are classified once here and shared by all three consumers.
func buildFrictionReport(invs []SdlcInvocation, marks []ActivityMark, nTranscripts int) FrictionReport {
	events := allGateEvents(invs)
	rep := aggregate(invs, events, nTranscripts)
	rep.RefusalRetries = detectRefusalRetries(invs, events)
	rep.FiringOrder = detectFiringOrder(invs, events, marks)
	return rep
}

// WorkflowVerbs is SpineVerbs plus the gate-less workflow verbs the firing-order
// ladder anchors on (claim/start-plan — stages 0–1 carry no bypass gates, so the
// catalog can't supply them).
func WorkflowVerbs() map[string]bool {
	out := SpineVerbs()
	out["claim"] = true
	out["start-plan"] = true
	return out
}

func obsString(o Observability) string {
	switch o {
	case ObsForceOnly:
		return "force-only"
	case ObsFlagOmitted:
		return "flag-omitted"
	default:
		return "full"
	}
}

// aggregate is pure over the parsed invocations (events precomputed via
// allGateEvents, index-aligned) — tallies per-gate bypass/refusal + per-repo and
// per-agent bypass. Gates are sorted by bypass count descending.
func aggregate(invs []SdlcInvocation, events [][]GateEvent, nTranscripts int) FrictionReport {
	type key struct{ cmd, flag string }
	bypass := map[key]int{}
	refusal := map[key]int{}
	obs := map[key]Observability{}
	byRepo := map[string]int{}
	byAgent := map[string]int{}
	for i, inv := range invs {
		// NB: we do NOT skip IsHelp invocations — a compound Bash command
		// (`sdlc close --no-judge && sdlc x --help`) contains "--help" yet carries a
		// real close bypass, and help-text lines don't match the specific ACK/refusal
		// patterns anyway, so the classifier already rejects them (#172).
		// The events stream is deduped per (kind, gate, command) — one no-validate
		// refusal prints two matching lines and must count once (M2).
		for _, ev := range events[i] {
			k := key{ev.Command, ev.Gate}
			obs[k] = ev.Observability // uniform per (command, flag) — no last-write collapse
			if ev.Kind == GateBypass {
				bypass[k]++
				byRepo[inv.Repo]++
				byAgent[inv.Agent]++
			} else {
				refusal[k]++
			}
		}
	}
	seen := map[key]bool{}
	for k := range bypass {
		seen[k] = true
	}
	for k := range refusal {
		seen[k] = true
	}
	var gates []GateStat
	for k := range seen {
		gates = append(gates, GateStat{
			Command: k.cmd, Flag: k.flag, Bypasses: bypass[k], Refusals: refusal[k],
			Observability: obsString(obs[k]),
		})
	}
	// Deterministic order: bypasses desc, then refusals desc, then flag/command.
	sort.SliceStable(gates, func(i, j int) bool {
		a, b := gates[i], gates[j]
		if a.Bypasses != b.Bypasses {
			return a.Bypasses > b.Bypasses
		}
		if a.Refusals != b.Refusals {
			return a.Refusals > b.Refusals
		}
		if a.Flag != b.Flag {
			return a.Flag < b.Flag
		}
		return a.Command < b.Command
	})
	return FrictionReport{
		TranscriptsScanned: nTranscripts, InvocationsSeen: len(invs),
		Gates: gates, ByRepoBypass: byRepo, ByAgentBypass: byAgent,
	}
}

func renderFrictionReport(rep FrictionReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# sdlc friction report\n\n")
	fmt.Fprintf(&b, "%d transcripts scanned, %d workflow-verb invocations.\n\n", rep.TranscriptsScanned, rep.InvocationsSeen)
	fmt.Fprintf(&b, "## Per-gate bypasses (command-anchored, contamination-filtered)\n\n")
	fmt.Fprintf(&b, "| command | gate | bypasses | refusals | observability |\n|---|---|---|---|---|\n")
	for _, g := range rep.Gates {
		note := ""
		if g.Observability != "full" {
			note = " ⚠️"
		}
		fmt.Fprintf(&b, "| %s | %s | %d | %d | %s%s |\n", g.Command, g.Flag, g.Bypasses, g.Refusals, g.Observability, note)
	}
	fmt.Fprintf(&b, "\n## Bypass concentration by repo\n\n| repo | bypasses |\n|---|---|\n")
	type rc struct {
		repo string
		n    int
	}
	var repos []rc
	for r, n := range rep.ByRepoBypass {
		repos = append(repos, rc{r, n})
	}
	sort.SliceStable(repos, func(i, j int) bool { return repos[i].n > repos[j].n })
	for _, r := range repos {
		fmt.Fprintf(&b, "| %s | %d |\n", r.repo, r.n)
	}

	// ── Per-agent split (M3) ──
	if len(rep.ByAgentBypass) > 0 {
		fmt.Fprintf(&b, "\n## Bypass concentration by agent\n\n| agent | bypasses |\n|---|---|\n")
		var agents []string
		for a := range rep.ByAgentBypass {
			agents = append(agents, a)
		}
		sort.SliceStable(agents, func(i, j int) bool {
			if rep.ByAgentBypass[agents[i]] != rep.ByAgentBypass[agents[j]] {
				return rep.ByAgentBypass[agents[i]] > rep.ByAgentBypass[agents[j]]
			}
			return agents[i] < agents[j]
		})
		for _, a := range agents {
			name := a
			if name == "" {
				name = "(untagged)"
			}
			fmt.Fprintf(&b, "| %s | %d |\n", name, rep.ByAgentBypass[a])
		}
	}
	if rep.CodexForksSkipped > 0 {
		fmt.Fprintf(&b, "\n_%d codex fork-replay rollout(s) skipped (they replay their parent's transcript — counting them double-counts every shared event)._\n", rep.CodexForksSkipped)
	}

	// ── Refusal→retry (M2) ──
	fmt.Fprintf(&b, "\n## Refusal→retry (per gate)\n\n")
	if len(rep.RefusalRetries) == 0 {
		b.WriteString("_No gate refusals observed._\n")
	} else {
		type rrKey struct{ cmd, gate string }
		type rrTally struct {
			refusals, retried, resolved, viaBypass int
			obs                                    string
		}
		tally := map[rrKey]*rrTally{}
		var rrOrder []rrKey
		for _, rr := range rep.RefusalRetries {
			k := rrKey{rr.Command, rr.Gate}
			tl := tally[k]
			if tl == nil {
				tl = &rrTally{obs: rr.Observability}
				tally[k] = tl
				rrOrder = append(rrOrder, k)
			}
			tl.refusals++
			if rr.Retried {
				tl.retried++
			}
			if rr.Resolved {
				tl.resolved++
			}
			if rr.ViaBypass {
				tl.viaBypass++
			}
		}
		sort.SliceStable(rrOrder, func(i, j int) bool {
			a, bK := rrOrder[i], rrOrder[j]
			if tally[a].refusals != tally[bK].refusals {
				return tally[a].refusals > tally[bK].refusals
			}
			if a.gate != bK.gate {
				return a.gate < bK.gate
			}
			return a.cmd < bK.cmd
		})
		fmt.Fprintf(&b, "| command | gate | refusals | retried | resolved | via bypass | observability |\n|---|---|---|---|---|---|---|\n")
		for _, k := range rrOrder {
			tl := tally[k]
			note := ""
			if tl.obs != "full" {
				note = " ⚠️"
			}
			fmt.Fprintf(&b, "| %s | %s | %d | %d | %d | %d | %s%s |\n",
				k.cmd, k.gate, tl.refusals, tl.retried, tl.resolved, tl.viaBypass, tl.obs, note)
		}
		b.WriteString("\n_`resolved` = the retry no longer refused this gate; `via bypass` = it resolved by routing AROUND the gate (--no-<gate>/--force), not by satisfying it. `flag-omitted` rows pair refusal→retry by verb+context only (best-effort). Pairing is within-transcript, unbounded in time, and milestone-blind (an M1 refusal can pair with an M2 retry); refuse→refuse→satisfy chains count per-record — both mildly overstate `retried`/understate `resolved`._\n")
	}

	// ── Firing-order (M2) ──
	fmt.Fprintf(&b, "\n## Firing-order anomalies\n\n")
	if len(rep.FiringOrder.Anomalies) == 0 {
		b.WriteString("_None detected._\n")
	} else {
		byKind := map[string]int{}
		for _, a := range rep.FiringOrder.Anomalies {
			byKind[a.Kind]++
		}
		var kinds []string
		for k := range byKind {
			kinds = append(kinds, k)
		}
		sort.Strings(kinds)
		fmt.Fprintf(&b, "| kind | count |\n|---|---|\n")
		for _, k := range kinds {
			fmt.Fprintf(&b, "| %s | %d |\n", k, byKind[k])
		}
		const maxListed = 20
		b.WriteString("\n")
		for i, a := range rep.FiringOrder.Anomalies {
			if i == maxListed {
				fmt.Fprintf(&b, "- …and %d more (full list in --json)\n", len(rep.FiringOrder.Anomalies)-maxListed)
				break
			}
			loc := a.Repo
			if a.IssueID != "" {
				loc += " #" + a.IssueID
			} else {
				loc += " (no issue context)"
			}
			fmt.Fprintf(&b, "- **%s** · %s", a.Kind, loc)
			if a.Detail != "" {
				fmt.Fprintf(&b, " — %s", a.Detail)
			}
			b.WriteString("\n")
		}
	}
	if rep.FiringOrder.UnattributedPublish > 0 {
		fmt.Fprintf(&b, "\n_%d merge/push invocation(s) had no attributable --issue context (kept out of every per-issue ladder)._\n", rep.FiringOrder.UnattributedPublish)
	}

	b.WriteString("\n_Observability: `force-only` = change-code silent bypass (countable only via --force); `flag-omitted` = merge/push refusal doesn't name the flag._\n")
	b.WriteString("_Stated limits: dev-style invocations (`go run ./cmd/sdlc <verb>`, `bin/sdlc <verb>`) are not anchored — undercounts the repo that dogfoods sdlc; in a compound command only the FIRST `sdlc <verb>` anchors (a second verb's gate lines are conservatively dropped); the skill-late arm is Claude-only (codex has no Skill tool; its file edits have no plan-skill counterpart to pair with)._\n")
	return b.String()
}

// RunFrictionReport walks the whole corpus — every Claude transcript under
// claudeRoot AND every codex rollout under codexRoot (M3) — extracts sdlc
// invocations, and returns the friction report as markdown (or JSON). The IO
// entry; enumeration + file reads are the only side effects. A missing corpus on
// ONE side is fine (single-agent machines); zero transcripts on BOTH errors.
func RunFrictionReport(claudeRoot, codexRoot string, asJSON bool) (string, error) {
	refs := enumerateClaudeTranscripts(claudeRoot)
	codexPaths := enumerateCodexTranscripts(codexRoot)
	if len(refs)+len(codexPaths) == 0 {
		// #68 lesson: a misinvocation (wrong/missing corpus roots) must not look
		// like a real empty answer.
		return "", fmt.Errorf("no transcripts found under %s or %s — are these the Claude projects / codex sessions dirs?", claudeRoot, codexRoot)
	}
	verbs := WorkflowVerbs()
	var invs []SdlcInvocation
	var marks []ActivityMark
	for _, ref := range refs {
		data, err := os.ReadFile(ref.Path)
		if err != nil {
			continue
		}
		tInvs, tMarks := scanTranscript(data, verbs)
		for _, inv := range tInvs {
			inv.Transcript, inv.Repo, inv.Agent = ref.Path, ref.Repo, ref.Agent
			invs = append(invs, inv)
		}
		for _, mk := range tMarks {
			mk.Transcript, mk.Repo = ref.Path, ref.Repo
			marks = append(marks, mk)
		}
	}
	scanned := len(refs)
	forksSkipped := 0
	for _, path := range codexPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		kind, cwd, ok := codexMeta(data)
		if !ok {
			continue // not a session rollout
		}
		if kind == codexForkReplay {
			// Replays its parent's transcript — processing it double-counts every
			// shared event (the spec's 66%-inflation trap). Counted, not hidden.
			forksSkipped++
			continue
		}
		label, include := repoLabelFromPath(cwd)
		if !include {
			continue
		}
		scanned++
		for _, inv := range parseCodexInvocations(data, verbs) {
			inv.Transcript, inv.Repo = path, label
			invs = append(invs, inv)
		}
	}
	rep := buildFrictionReport(invs, marks, scanned)
	rep.CodexForksSkipped = forksSkipped
	if asJSON {
		out, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			return "", err
		}
		return string(out) + "\n", nil
	}
	return renderFrictionReport(rep), nil
}
