package processmanual

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
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

// scanTranscript extracts only anchored Bash(sdlc <verb>) calls, joined to their
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
	invs, _ := scanTranscript([]byte(strings.Join(lines, "\n")),
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

// ── M2 Task 6: firing-order (per-issue, iteration-aware) ─────────────────────

func foInv(verb, issue, transcript string, min int, output string) SdlcInvocation {
	return SdlcInvocation{Verb: verb, IssueID: issue, Transcript: transcript, Repo: "r",
		Time: time.Date(2026, 7, 14, 10, min, 0, 0, time.UTC), Output: output}
}

// reworkOutput carries the structured verdict block a REWORK boundary review
// streams back through the close/milestone-close output.
const reworkOutput = "review dispatched\n```verdict\nverdict: REWORK\nconfidence: high\n```\nrework required"

func TestDetectFiringOrderLadder(t *testing.T) {
	cases := []struct {
		name      string
		invs      []SdlcInvocation
		wantKinds []string
	}{
		{"change-code after a clean close flags (inverted order)", []SdlcInvocation{
			foInv("claim", "5", "t", 0, ""),
			foInv("close", "5", "t", 1, "closed."),
			foInv("change-code", "5", "t", 2, ""),
		}, []string{"change-code-after-close"}},
		{"milestone-close→change-code is the legal next-milestone loop", []SdlcInvocation{
			foInv("change-code", "5", "t", 0, ""),
			foInv("milestone-close", "5", "t", 1, "closed M1"),
			foInv("change-code", "5", "t", 2, ""),
			foInv("milestone-close", "5", "t", 3, "closed M2"),
			foInv("close", "5", "t", 4, "closed."),
		}, nil},
		{"start-plan re-runs are legal (AGENTS.md: re-run per design)", []SdlcInvocation{
			foInv("claim", "5", "t", 0, ""),
			foInv("start-plan", "5", "t", 1, ""),
			foInv("change-code", "5", "t", 2, ""),
			foInv("start-plan", "5", "t", 3, ""),
		}, nil},
		{"close→change-code after REWORK is the legal reopen loop", []SdlcInvocation{
			foInv("change-code", "5", "t", 0, ""),
			foInv("close", "5", "t", 1, reworkOutput),
			foInv("change-code", "5", "t", 2, ""),
			foInv("close", "5", "t", 3, "closed."),
		}, nil},
		{"cross-issue interleave keys per issue", []SdlcInvocation{
			foInv("claim", "5", "t", 0, ""),
			foInv("close", "5", "t", 1, "closed."),
			foInv("claim", "6", "t", 2, ""),
			foInv("start-plan", "6", "t", 3, ""),
		}, nil},
		{"a gate-REFUSED close does not raise the ladder (legal recovery)", []SdlcInvocation{
			foInv("change-code", "5", "t", 0, ""),
			// close refused by the actual-hours gate — boundary NOT crossed
			foInv("close", "5", "t", 1, "  Pass --no-actual (or --force) only when measurement is not applicable; close records actual_hours: N/A."),
			foInv("change-code", "5", "t", 2, ""),
			foInv("close", "5", "t", 3, "closed."),
		}, nil},
		{"--help invocations are not workflow steps", []SdlcInvocation{
			{Verb: "close", IssueID: "5", Transcript: "t", Repo: "r", IsHelp: true,
				Time: time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)},
			foInv("change-code", "5", "t", 1, ""),
		}, nil},
	}
	for _, c := range cases {
		res := detectFiringOrder(c.invs, nil)
		var kinds []string
		for _, a := range res.Anomalies {
			kinds = append(kinds, a.Kind)
		}
		if !reflect.DeepEqual(kinds, c.wantKinds) {
			t.Errorf("%s: anomalies %v, want %v (%+v)", c.name, kinds, c.wantKinds, res.Anomalies)
		}
	}
}

// merge/push carry no --issue (touched issues come from the git diff, invisible
// to the transcript) — attribute from segment context, or count unattributed and
// keep OUT of every per-issue ladder.
func TestDetectFiringOrderMergeAttribution(t *testing.T) {
	invs := []SdlcInvocation{
		foInv("change-code", "8", "t", 0, ""),
		foInv("close", "8", "t", 10, "closed."),
		foInv("merge", "", "t", 12, "merged."), // → attributed to 8 (nearest preceding --issue)
		foInv("change-code", "8", "t", 20, ""), // → after the attributed merge
	}
	res := detectFiringOrder(invs, nil)
	if len(res.Anomalies) != 1 || res.Anomalies[0].Kind != "change-code-after-close" ||
		!strings.Contains(res.Anomalies[0].Detail, "merge") {
		t.Fatalf("want one change-code-after-close with merge detail (attribution raised the ladder), got %+v", res.Anomalies)
	}
	if res.UnattributedPublish != 0 {
		t.Errorf("attributed merge counted as unattributed: %d", res.UnattributedPublish)
	}

	res2 := detectFiringOrder([]SdlcInvocation{foInv("merge", "", "t2", 0, "merged.")}, nil)
	if res2.UnattributedPublish != 1 || len(res2.Anomalies) != 0 {
		t.Errorf("context-free merge: want unattributed=1, no anomalies; got %d / %+v",
			res2.UnattributedPublish, res2.Anomalies)
	}
}

// skill-late: a plan/TDD Skill load AFTER a (non-doc) file edit in the same
// segment+issue — planning arriving once implementation already started.
func TestDetectFiringOrderSkillLate(t *testing.T) {
	mark := func(kind Kind, detail string, min int) ActivityMark {
		return ActivityMark{Kind: kind, Detail: detail, Transcript: "t", Repo: "r",
			Time: time.Date(2026, 7, 14, 10, min, 0, 0, time.UTC)}
	}
	invs := []SdlcInvocation{foInv("claim", "9", "t", 0, "")}

	res := detectFiringOrder(invs, []ActivityMark{
		mark(KindFileEdit, "cmd/sdlc/foo.go", 1),
		mark(KindSkill, "superpowers-writing-plans", 2),
	})
	if len(res.Anomalies) != 1 || res.Anomalies[0].Kind != "skill-late" || res.Anomalies[0].IssueID != "9" {
		t.Fatalf("code edit → plan skill: want one skill-late on issue 9, got %+v", res.Anomalies)
	}

	// doc/plan/issue (.md) edits are design work, not implementation — no flag
	res2 := detectFiringOrder(invs, []ActivityMark{
		mark(KindFileEdit, "workshop/plans/000009-x-plan.md", 1),
		mark(KindSkill, "superpowers-writing-plans", 2),
	})
	if len(res2.Anomalies) != 0 {
		t.Errorf(".md edit before plan skill must not flag, got %+v", res2.Anomalies)
	}

	// skill loaded BEFORE any edit is the correct order
	res3 := detectFiringOrder(invs, []ActivityMark{
		mark(KindSkill, "superpowers-writing-plans", 1),
		mark(KindFileEdit, "cmd/sdlc/foo.go", 2),
	})
	if len(res3.Anomalies) != 0 {
		t.Errorf("skill-then-edit must not flag, got %+v", res3.Anomalies)
	}
}

// Edit/Write/MultiEdit + Skill tool_use records surface as ActivityMarks
// alongside the anchored invocations (KindFileEdit — deferred from M1 Task 3).
func TestScanTranscriptMarks(t *testing.T) {
	mkTool := func(id, name string, input map[string]any) string {
		b, _ := json.Marshal(map[string]any{
			"type": "assistant", "timestamp": "2026-07-14T10:00:00Z",
			"message": map[string]any{"content": []any{
				map[string]any{"type": "tool_use", "id": id, "name": name, "input": input},
			}},
		})
		return string(b)
	}
	lines := []string{
		mkTool("e1", "Edit", map[string]any{"file_path": "/r/cmd/foo.go", "old_string": "a", "new_string": "b"}),
		mkTool("w1", "Write", map[string]any{"file_path": "/r/docs/x.md", "content": "c"}),
		mkTool("s1", "Skill", map[string]any{"skill": "superpowers-writing-plans"}),
		mkAssistantBash("tu1", "sdlc claim --issue 9"),
	}
	invs, marks := scanTranscript([]byte(strings.Join(lines, "\n")), map[string]bool{"claim": true})
	if len(invs) != 1 || invs[0].Verb != "claim" {
		t.Fatalf("want the one claim invocation, got %+v", invs)
	}
	want := []struct {
		kind   Kind
		detail string
	}{
		{KindFileEdit, "/r/cmd/foo.go"},
		{KindFileEdit, "/r/docs/x.md"},
		{KindSkill, "superpowers-writing-plans"},
	}
	if len(marks) != len(want) {
		t.Fatalf("want %d marks, got %+v", len(want), marks)
	}
	for i, w := range want {
		if marks[i].Kind != w.kind || marks[i].Detail != w.detail {
			t.Errorf("mark %d = {%s %q}, want {%s %q}", i, marks[i].Kind, marks[i].Detail, w.kind, w.detail)
		}
	}
}

// ── M2 Task 7: report fold + render/JSON shape (M1 review gap) ───────────────

func TestFrictionReportRenderAndJSON(t *testing.T) {
	esc := "\x1b"
	atlasRefusal := "  or pass --no-atlas (or --force) with the rationale in --verified"
	atlasAck := "  " + esc + "[1;33m[!]" + esc + "[0m --no-atlas (or --force): skipping atlas/ change check — rationale in --verified"
	invs := []SdlcInvocation{
		foInv("claim", "5", "t", 0, ""),
		foInv("close", "5", "t", 1, atlasRefusal), // refused …
		foInv("close", "5", "t", 2, atlasAck),     // … retried via bypass
		foInv("change-code", "5", "t", 3, ""),     // change-code after the clean close → anomaly
	}
	rep := buildFrictionReport(invs, nil, 1)

	md := renderFrictionReport(rep)
	for _, want := range []string{
		"| close | no-atlas | 1 | 1 |",  // per-gate row: 1 bypass, 1 refusal
		"## Refusal→retry",              // section present
		"| close | no-atlas | 1 | 1 | 1 | 1 |", // refusals/retried/resolved/via-bypass tallies
		"## Firing-order anomalies",
		"change-code-after-close",
		"go run",                        // stated-limitations footnote (dev invocations not anchored)
	} {
		if !strings.Contains(md, want) {
			t.Errorf("render missing %q in:\n%s", want, md)
		}
	}

	raw, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	rrs, ok := m["refusal_retries"].([]any)
	if !ok || len(rrs) != 1 {
		t.Fatalf("JSON refusal_retries: want 1 record, got %v", m["refusal_retries"])
	}
	fo, ok := m["firing_order"].(map[string]any)
	if !ok {
		t.Fatalf("JSON firing_order missing: %v", m)
	}
	if an, ok := fo["anomalies"].([]any); !ok || len(an) != 1 {
		t.Errorf("JSON firing_order.anomalies: want 1, got %v", fo["anomalies"])
	}
}

// toolResultText's legacy array-of-{text} form (older transcripts) — untested in
// M1 (review Minor).
func TestToolResultTextArrayForm(t *testing.T) {
	got := toolResultText(json.RawMessage(`[{"type":"text","text":"line one"},{"type":"text","text":"line two"}]`))
	if !strings.Contains(got, "line one") || !strings.Contains(got, "line two") {
		t.Fatalf("array-form content not joined: %q", got)
	}
}

// Zero enumerable transcripts must ERROR, not print an empty report — the #68
// lesson: a misinvocation must not look like a real empty answer (M1 review Minor).
func TestRunFrictionReportZeroTranscripts(t *testing.T) {
	if _, err := RunFrictionReport(filepath.Join(t.TempDir(), "no-such-dir"), false); err == nil {
		t.Fatal("want an error when zero transcripts enumerate, got nil")
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
