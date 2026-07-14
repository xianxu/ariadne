package processmanual

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mkAssistantBash(id, cmd string) string {
	b, _ := json.Marshal(map[string]any{
		"type": "assistant", "timestamp": "2026-07-14T10:00:00Z",
		"message": map[string]any{"content": []any{
			map[string]any{"type": "tool_use", "id": id, "name": "Bash",
				"input": map[string]any{"command": cmd}},
		}},
	})
	return string(b)
}

func mkUserResult(id, stdout string) string {
	b, _ := json.Marshal(map[string]any{
		"type": "user",
		"message": map[string]any{"content": []any{
			map[string]any{"type": "tool_result", "tool_use_id": id, "content": stdout},
		}},
		"toolUseResult": map[string]any{"stdout": stdout, "stderr": ""},
	})
	return string(b)
}

// sdlcInvocations extracts only anchored Bash(sdlc <verb>) calls, joined to their
// tool_use_id-linked output, with issueID + isHelp parsed from the command.
func TestSdlcInvocations(t *testing.T) {
	ack := "  \x1b[1;33m[!]\x1b[0m --no-atlas (or --force): skipping atlas/ change check — rationale in --verified"
	lines := []string{
		mkAssistantBash("tu1", "sdlc close --issue 173 --no-atlas"),
		mkUserResult("tu1", "some preamble\n"+ack+"\ndone"),
		mkAssistantBash("tu2", "sdlc change-code --help --issue 172"),
		mkUserResult("tu2", "usage: sdlc change-code ..."),
		mkAssistantBash("tu3", "git status"), // non-sdlc Bash → excluded
		mkUserResult("tu3", "clean"),
	}
	invs := sdlcInvocations([]byte(strings.Join(lines, "\n")),
		map[string]bool{"close": true, "change-code": true})

	if len(invs) != 2 {
		t.Fatalf("got %d invocations, want 2 (close, change-code; git excluded): %+v", len(invs), invs)
	}
	if invs[0].Verb != "close" || invs[0].IssueID != "173" || invs[0].IsHelp {
		t.Errorf("inv0 = {verb:%q issue:%q help:%v}, want {close 173 false}", invs[0].Verb, invs[0].IssueID, invs[0].IsHelp)
	}
	if !strings.Contains(invs[0].Output, "--no-atlas (or --force): skipping") {
		t.Errorf("inv0 output missing the linked ACK: %q", invs[0].Output)
	}
	if invs[1].Verb != "change-code" || invs[1].IssueID != "172" || !invs[1].IsHelp {
		t.Errorf("inv1 = {verb:%q issue:%q help:%v}, want {change-code 172 true}", invs[1].Verb, invs[1].IssueID, invs[1].IsHelp)
	}
	// end-to-end: the linked ACK line classifies as a no-atlas bypass under `close`
	var got bool
	for _, ln := range strings.Split(invs[0].Output, "\n") {
		if ev, ok := classifyOutputLine(ln, invs[0].Verb); ok && ev.Kind == GateBypass && ev.Gate == "no-atlas" {
			got = true
		}
	}
	if !got {
		t.Errorf("expected a no-atlas bypass from inv0's output lines")
	}
}

// classifyOutputLine over REAL captured lines (copied verbatim from
// ~/.claude/projects) plus the documented-shape G2/refusal cases and — the whole
// point — the contamination cases the classifier MUST reject. Inventing fixtures is
// exactly the trap the #172 plan reviews caught; these are ground truth.
func TestClassifyOutputLine(t *testing.T) {
	esc := "\x1b"
	cases := []struct {
		name    string
		line    string
		verb    string
		wantOK  bool
		wantK   GateEventKind
		wantG   string
		wantObs Observability
	}{
		// ---- real runtime bypass ACKs (reset-bearing) ----
		{"g1 no-verdict ack", "  " + esc + "[1;33m[!]" + esc + "[0m --no-verdict (or --force): skipping Review-Verdict check for 1 milestone(s): M1",
			"close", true, GateBypass, "no-verdict", ObsFull},
		{"cinfo no-judge ack", esc + "[1;36m==>" + esc + "[0m skipping issue boundary review per --no-judge (or --force)",
			"close", true, GateBypass, "no-judge", ObsFull},
		{"g3 no-validate ack (⚠️ + double space)", "  " + esc + "[1;33m[!]" + esc + "[0m ⚠️  --no-validate: SKIPPING the instance-conformance gate (#124) — issue frontmatter NOT verified",
			"merge", true, GateBypass, "no-validate", ObsFull},
		// G2 change-code: force-only observability, ViaForce
		{"g2 plan-quality force-bypass", "  " + esc + "[1;33m[!]" + esc + "[0m plan-quality gate bypassed (--force: needed to iterate)",
			"change-code", true, GateBypass, "no-judge", ObsForceOnly},

		// ---- real runtime refusal (plain string, NO reset) ----
		{"no-actual refusal", "  Pass --no-actual (or --force) only when measurement is not applicable; close records actual_hours: N/A.",
			"close", true, GateRefusal, "no-actual", ObsFull},

		// ---- contamination: MUST classify as none ----
		{"warmup success line (not a refusal)", "             Pass --no-actual (or --force) only if there's genuinely nothing",
			"close", false, 0, "", 0},
		{"close.go source read (cat-n + append)", "944\t\ttail = append(tail, \"  Pass --no-actual (or --force) only when measurement is not applicable\")",
			"close", false, 0, "", 0},
		{"Sprintf source line", "        id := fmt.Sprintf(\"%06d\", issueNum)",
			"close", false, 0, "", 0},
		{"cat-n shell read", "226\t            --force) force=1; shift;;",
			"close", false, 0, "", 0},
		// verb anchoring: a close-gate ACK line seen under `merge` is not a merge gate
		{"wrong verb → none", "  " + esc + "[1;33m[!]" + esc + "[0m --no-verdict (or --force): skipping Review-Verdict check",
			"merge", false, 0, "", 0},
	}
	for _, c := range cases {
		ev, ok := classifyOutputLine(c.line, c.verb)
		if ok != c.wantOK {
			t.Errorf("%s: ok=%v want %v (ev=%+v)", c.name, ok, c.wantOK, ev)
			continue
		}
		if !ok {
			continue
		}
		if ev.Kind != c.wantK || ev.Gate != c.wantG || ev.Observability != c.wantObs {
			t.Errorf("%s: got {kind:%d gate:%q obs:%d} want {kind:%d gate:%q obs:%d}",
				c.name, ev.Kind, ev.Gate, ev.Observability, c.wantK, c.wantG, c.wantObs)
		}
	}
}

func TestRepoLabel(t *testing.T) {
	cases := []struct {
		slug, want string
		include    bool
	}{
		{"-Users-xianxu-workspace-ariadne", "ariadne", true},
		{"-Users-xianxu-workspace-parley.nvim", "parley.nvim", true},
		{"-Users-xianxu-workspace-worktree-ariadne-000095-weave", "ariadne", true},
		{"-private-tmp-claude-501", "", false},
		{"-private-var-folders-07-xyz-T", "", false},
	}
	for _, c := range cases {
		got, inc := repoLabel(c.slug)
		if inc != c.include || (inc && got != c.want) {
			t.Errorf("repoLabel(%q) = (%q,%v), want (%q,%v)", c.slug, got, inc, c.want, c.include)
		}
	}
}

// The load-bearing test: one real close bypass ACK buried among source-code +
// cat-n contamination lines must yield exactly ONE bypass, zero from the noise.
func TestAggregateAntiContamination(t *testing.T) {
	esc := "\x1b"
	ackLine := "  " + esc + "[1;33m[!]" + esc + "[0m --no-atlas (or --force): skipping atlas/ change check — rationale in --verified"
	mixed := strings.Join([]string{
		"reading close.go for context",
		`cmd/sdlc/close.go:437:  cwarn(stderr, "--no-atlas (or --force): skipping atlas/ change check")`, // source
		"944\t\ttail = append(tail, \"Pass --no-atlas (or --force) …\")",                                // cat-n
		ackLine, // the ONE real bypass
		"done.",
	}, "\n")
	invs := []SdlcInvocation{
		{Verb: "close", Output: mixed, Repo: "ariadne"},
		{Verb: "close", Output: "no gate events here", Repo: "ariadne"},
	}
	rep := aggregate(invs, 2)
	if len(rep.Gates) != 1 || rep.Gates[0].Flag != "no-atlas" || rep.Gates[0].Bypasses != 1 {
		t.Fatalf("want exactly one no-atlas bypass (the real ACK; source+cat-n rejected), got %+v", rep.Gates)
	}
	if rep.ByRepoBypass["ariadne"] != 1 {
		t.Errorf("want ariadne=1 bypass, got %v", rep.ByRepoBypass)
	}
}

// Observability must be keyed per (command, flag): no-judge is `full` for
// close/mclose but `force-only` for change-code — collapsing to the flag mislabels
// the honesty column (#172 M1 boundary review Important #1).
func TestObservabilityPerCommand(t *testing.T) {
	esc := "\x1b"
	closeAck := esc + "[1;36m==>" + esc + "[0m skipping issue boundary review per --no-judge (or --force)"
	ccAck := "  " + esc + "[1;33m[!]" + esc + "[0m plan-quality gate bypassed (--force: needed to iterate)"
	rep := aggregate([]SdlcInvocation{
		{Verb: "close", Output: closeAck, Repo: "r"},
		{Verb: "change-code", Output: ccAck, Repo: "r"},
	}, 2)
	got := map[string]string{}
	for _, g := range rep.Gates {
		if g.Flag == "no-judge" {
			got[g.Command] = g.Observability
		}
	}
	if got["close"] != "full" {
		t.Errorf("close no-judge observability = %q, want full", got["close"])
	}
	if got["change-code"] != "force-only" {
		t.Errorf("change-code no-judge observability = %q, want force-only (not last-write-wins)", got["change-code"])
	}
}

// ── M2 Task 5: per-invocation dedupe + refusal→retry pairing ─────────────────

// One no-validate refusal emits TWO matching lines — the validategate.go cwarn
// ("… nonconforming changed issue file(s) — fix and re-run, or --no-validate …")
// and the die-wrapped returned error ("… nonconforming issue file(s)") — which
// must collapse to ONE refusal per invocation, else refusal→retry resolution
// rates skew (#172 M1 review Minor).
func TestInvocationGateEventsDedupe(t *testing.T) {
	esc := "\x1b"
	out := strings.Join([]string{
		"  " + esc + "[1;33m[!]" + esc + "[0m instance-conformance gate: 2 nonconforming changed issue file(s) — fix and re-run, or --no-validate to bypass (loud):",
		"  - workshop/issues/000042-x.md (frontmatter):",
		"  " + esc + "[1;31m[✗]" + esc + "[0m instance-conformance gate: 2 nonconforming issue file(s)",
	}, "\n")
	evs := invocationGateEvents(SdlcInvocation{Verb: "merge", Output: out})
	if len(evs) != 1 || evs[0].Kind != GateRefusal || evs[0].Gate != "no-validate" {
		t.Fatalf("want exactly ONE deduped no-validate refusal, got %+v", evs)
	}
}

// detectRefusalRetries pairs a refusal with the next same-verb+same-issue
// invocation in the same transcript. Real captured line shapes throughout.
func TestDetectRefusalRetries(t *testing.T) {
	esc := "\x1b"
	atlasRefusal := "  or pass --no-atlas (or --force) with the rationale in --verified"
	atlasAck := "  " + esc + "[1;33m[!]" + esc + "[0m --no-atlas (or --force): skipping atlas/ change check — rationale in --verified"
	verdictRefusal := "  Or pass --no-verdict (or --force); record the reason in --verified."
	publishRefusal := "  " + esc + "[1;31m[✗]" + esc + "[0m publish gate: 2 commit(s) landed after `sdlc close` (anchor abc1234) — the boundary review no longer covers HEAD."
	warmup := "             Pass --no-actual (or --force) only if there's genuinely nothing"

	invs := []SdlcInvocation{
		// t1 / issue 5: no-atlas refusal → (unrelated issue-6 close between) → retried WITH the bypass flag
		{Verb: "close", IssueID: "5", Transcript: "t1", Repo: "ariadne", Output: atlasRefusal},
		{Verb: "close", IssueID: "6", Transcript: "t1", Repo: "ariadne", Output: "closed."}, // different issue — NOT the retry
		{Verb: "close", IssueID: "5", Transcript: "t1", Repo: "ariadne", Output: atlasAck},
		// t1 / issue 7: no-verdict refusal never retried; warmup line must add nothing
		{Verb: "close", IssueID: "7", Transcript: "t1", Repo: "ariadne", Output: verdictRefusal + "\n" + warmup},
		// t2: merge publish-gate refusal (flag never named) → paired by verb+context; retry clean
		{Verb: "merge", IssueID: "", Transcript: "t2", Repo: "brain", Output: publishRefusal},
		{Verb: "merge", IssueID: "", Transcript: "t2", Repo: "brain", Output: "merged 000042 into main."},
	}
	rrs := detectRefusalRetries(invs)
	if len(rrs) != 3 {
		t.Fatalf("want 3 refusal→retry records (warmup adds none), got %d: %+v", len(rrs), rrs)
	}
	byGate := map[string]RefusalRetry{}
	for _, rr := range rrs {
		byGate[rr.Gate] = rr
	}
	if rr := byGate["no-atlas"]; !rr.Retried || !rr.Resolved || !rr.ViaBypass || rr.IssueID != "5" {
		t.Errorf("no-atlas: want retried+resolved VIA BYPASS on issue 5, got %+v", rr)
	}
	if rr := byGate["no-verdict"]; rr.Retried || rr.Resolved {
		t.Errorf("no-verdict: want never-retried, got %+v", rr)
	}
	if rr := byGate["no-judge"]; !rr.Retried || !rr.Resolved || rr.ViaBypass || rr.Observability != "flag-omitted" {
		t.Errorf("merge no-judge: want retried+resolved (satisfied, not bypassed), flag-omitted attribution, got %+v", rr)
	}
}

// A refusal→bypass inside ONE compound invocation (`sdlc close … || sdlc close
// --no-atlas …` output carries both) resolves in place.
func TestDetectRefusalRetriesSameInvocation(t *testing.T) {
	esc := "\x1b"
	out := "  or pass --no-atlas (or --force) with the rationale in --verified\n" +
		"  " + esc + "[1;33m[!]" + esc + "[0m --no-atlas (or --force): skipping atlas/ change check — rationale in --verified"
	rrs := detectRefusalRetries([]SdlcInvocation{{Verb: "close", IssueID: "9", Transcript: "t", Output: out}})
	if len(rrs) != 1 || !rrs[0].Retried || !rrs[0].Resolved || !rrs[0].ViaBypass {
		t.Fatalf("want in-invocation refusal resolved via bypass, got %+v", rrs)
	}
}

func TestEnumerateClaudeTranscripts(t *testing.T) {
	root := t.TempDir()
	mk := func(slug string, n int) {
		d := filepath.Join(root, slug)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < n; i++ {
			os.WriteFile(filepath.Join(d, fmt.Sprintf("s%d.jsonl", i)), []byte("{}"), 0o644)
		}
	}
	mk("-Users-x-workspace-ariadne", 2)
	mk("-Users-x-workspace-worktree-ariadne-000-w", 1) // → labeled ariadne
	mk("-private-tmp-claude-501", 3)                   // excluded
	refs := enumerateClaudeTranscripts(root)
	byRepo := map[string]int{}
	for _, r := range refs {
		byRepo[r.Repo]++
	}
	if byRepo["ariadne"] != 3 || len(byRepo) != 1 {
		t.Errorf("want ariadne=3 (2 + 1 worktree), temp excluded; got %v", byRepo)
	}
}
