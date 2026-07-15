# post-FIX-THEN-SHIP protocol — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** After a FIX-THEN-SHIP verdict the agent-facing next action is stated by `sdlc close` itself (candidate A); post-close doc-only bookkeeping no longer trips the merge/push publish gate (candidate C); the reclose refusal names the new-issue path for post-publish work. Candidate B (`sdlc reverify`) is explicitly skipped — operator decision 2026-07-14.

**Architecture:** All three legs are deterministic message/predicate changes on existing seams — no new verbs, no new state. (A) `reviewThenFinalize`'s `closeFinalize` arm gains a verdict-conditional protocol block (`review.Verdict == judge.VerdictFixThenShip`), rendered by a pure formatter. (C) `runPublishGate`'s refusal path first checks whether the `anchor..HEAD` delta is docs-only via the existing `hasCodePath` classifier (#177 — the same "no code surface" definition the atlas gate auto-satisfies with, ARCH-DRY); docs-only passes with an info line, code deltas keep the pinned refusal. The reclose refusal appends the sanctioned recovery. Every new/changed line is checked against the gatesig catalog (`assertNoGatesigCollision` / preserved `RefusalPat` spans, #172).

**Tech Stack:** Go (cmd/sdlc); existing seams: `stubJudge`/`closeRepo`/`rewriteIssuePlan` (closereview_test.go), `publishRepo`/`writeIssueStatus`/`commitCode` (publishgate_test.go).

**Why this dissolves the measured friction (#172):**
- Shape 2 (6/6 merge/push publish refusals = post-close bookkeeping commits): those deltas are `workshop/`, `atlas/`, `*.md` — exactly `!hasCodePath` — so leg C turns all six into deterministic passes. The reviewed-HEAD-unchanged invariant is preserved *for code*: the boundary review's correctness claims are about behavior, and #177 already established "docs are not reviewable code surface" as gate policy.
- Shape 1 (3/3 reclose bypasses + the "review ALREADY RAN" no-judge rationale): the loop starts because close's FIX-THEN-SHIP output states no protocol. Leg A states it at the moment of ambiguity: fix now, bundle into ONE commit, don't re-close. The reclose guard (fires only at `done`, i.e. post-publish) gets the missing rule: follow-up work is a new issue.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `formatFixThenShipProtocol` | `cmd/sdlc/close.go` | new |
| `formatPublishGateDocsOnly` | `cmd/sdlc/publishgate.go` | new |
| reclose refusal string | `cmd/sdlc/close.go:421` | modified |
| publish-gate refusal string | `cmd/sdlc/publishgate.go:115-118` | modified |

- **`formatFixThenShipProtocol()`** — the post-FIX-THEN-SHIP next-action block. Contract (pinned by test): says fix the findings **now / before committing**; says bundle fixes + the issue-file close mutations into **ONE commit** (or amend); says do **NOT re-run `sdlc close`**; explains why (the publish gate anchors on the close commit); names the escape hatch (fixes landing *after* the close commit → re-run close to re-review + advance the anchor). Pure, no args (the protocol is verdict-generic across issue/milestone closes).
  - **DRY rationale:** first occurrence; colocated with the other close formatters.
  - **Future extensions:** if verdict-specific protocols grow (e.g. a REWORK checklist), this becomes a `formatVerdictProtocol(v)` dispatch.
- **`formatPublishGateDocsOnly(n int, anchor string)`** — the docs-only pass info line: phrased to avoid the G3 RefusalPat — e.g. `publish gate: N doc-only commit(s) after close (anchor X) — no code surface, reviewed-HEAD-unchanged holds for code (#174)`; never echo the refusal vocabulary "landed after". Must NOT match any gatesig `AckPat`/`RefusalPat` (`assertNoGatesigCollision`).
- **reclose refusal** — keep the gatesig-pinned span `is already status: done — pass --no-reclose-guard (or --force) to re-close` verbatim (`gatesig.go:84`); append: post-publish follow-up work is a new issue (`side-quest:` commit or `sdlc issue new`), not a re-close.
- **publish-gate refusal** — keep the pinned span `publish gate: N commit(s) landed after` (`gatesig.go:120,126`); the trailing guidance gains the protocol line ("bundle post-close bookkeeping into the close commit — doc-only deltas pass automatically (#174); code deltas require re-running close").

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `closeFinalize` arm | `cmd/sdlc/close.go` (`reviewThenFinalize`) | modified | stderr |
| `runPublishGate` drift branch | `cmd/sdlc/publishgate.go:114-119` | modified | `git diff --name-only` + stderr |

- **`closeFinalize` arm** — after `applyClose`/trailer emission, when `review.Verdict == judge.VerdictFixThenShip`, `cwarn(stderr, formatFixThenShipProtocol())`. Fires for both issue and milestone closes (the bundle-into-the-boundary-commit advice is identical; the publish gate is issue-level but the fix-before-crossing rule is not).
- **`runPublishGate` drift branch** — when `minAhead > 0`: `gitx.DiffNames(newestAnchor, "HEAD")` (the same helper the atlas window uses at close.go:439 — ARCH-DRY); on git error keep fail-closed (refuse, mirroring the existing `revCount` posture); if `!hasCodePath(paths)` → `cinfo` the docs-only line and return nil; else the (augmented) refusal.

**Test surface:** both formatters get contract tests (+ `assertNoGatesigCollision` for the info line). Wiring pinned per the un-wired-helper lesson: a FIX-THEN-SHIP finalize integration test on the `closeRepo`/`stubJudge` seams (and a SHIP negative — no protocol block), plus `TestRunPublishGate` subtests on the `publishRepo` harness (docs-only drift passes; code drift still refuses; mixed drift refuses; git-error fail-closed unchanged). `go test ./cmd/sdlc/internal/processmanual/` guards the preserved pins.

---

## Chunk 1: all three legs

### Task 1 (leg A): FIX-THEN-SHIP protocol block

**Files:**
- Modify: `cmd/sdlc/close.go` (`reviewThenFinalize` closeFinalize arm + new formatter near the other `format*` helpers)
- Test: `cmd/sdlc/close_test.go` (formatter contract), `cmd/sdlc/closereview_test.go` (wiring)

- [ ] **Step 1: Failing contract + wiring tests**

```go
// close_test.go — contract
func TestFormatFixThenShipProtocol_ContractElements(t *testing.T) {
	msg := formatFixThenShipProtocol()
	for _, w := range []string{
		"FIX-THEN-SHIP",
		"before committing",      // fix NOW, pre-commit
		"ONE commit",             // bundle fixes + close mutations
		"do NOT re-run",          // the anti-loop instruction
		"publish gate",           // the why (anchor semantics)
		"re-run `sdlc close`",    // the post-commit escape hatch
	} {
		if !strings.Contains(msg, w) { t.Errorf(...) }
	}
	assertNoGatesigCollision(t, msg)
}

// closereview_test.go — wiring (mirrors TestClose_TrailingUnclosedMilestones_AcceptedByCloseReview shape)
func TestClose_FixThenShip_EmitsProtocol(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	stubJudge(t, "VERDICT: FIX-THEN-SHIP (confidence: high)\n\nFinding: nit.\n")
	// flags as in TestRunCloseWithReview_IssueClose_Dispatches
	// run runCloseWithReview with captured stderr; assert err == nil (finalizes),
	// stderr contains "ONE commit" + "do NOT re-run",
	// readIssue contains "status: codecomplete" + "review verdict: FIX-THEN-SHIP".
}

func TestClose_Ship_NoProtocolBlock(t *testing.T) {
	// SHIP close (reuse the dispatch-test posture): stderr must NOT contain
	// "FIX-THEN-SHIP" / "ONE commit" — the protocol is verdict-conditional.
}
```

- [ ] **Step 2: Run — red** (undefined formatter; missing stderr content).
- [ ] **Step 3: Implement** the formatter + the one-line conditional in the closeFinalize arm (after `annotateLogLineWithVerdict`, before the lessons reminder, so the reminder's "write lessons" lands inside the same pre-commit window the protocol describes).
- [ ] **Step 4: Run — green**; `go test ./cmd/sdlc/internal/processmanual/` green.
- [ ] **Step 5: Commit** — `#174: close states the post-FIX-THEN-SHIP protocol (leg A)`

### Task 2 (leg C): doc-tolerant publish gate

**Files:**
- Modify: `cmd/sdlc/publishgate.go` (`runPublishGate` drift branch + new formatter)
- Test: `cmd/sdlc/publishgate_test.go` (new `TestRunPublishGate` subtests + formatter contract)

- [ ] **Step 1: Failing tests** on the `publishRepo` harness:

```go
t.Run("docs-only drift after close passes (#174)", func(t *testing.T) {
	git, base := publishRepo(t)
	writeIssueStatus(t, git, 69, "codecomplete", "#69 close")
	// lessons/bookkeeping commit — the measured 6/6 shape
	os.WriteFile("lessons.md", ...); git("add", "lessons.md"); git("commit", ...)
	// note: root-level *.md IS docs per hasCodePath
	if err := runPublishGate(base, "workshop/issues", io.Discard); err != nil {
		t.Errorf("docs-only delta should pass: %v", err)
	}
})
t.Run("code drift still refuses", ...)          // commitCode → err contains "landed after"
t.Run("multi-issue: two codecomplete anchors + trailing docs commit passes", ...) // the anchor-selection interaction
t.Run("mixed docs+code drift refuses", ...)     // both commits → refuse
// + formatter contract test with assertNoGatesigCollision
```

  (`assertNoGatesigCollision` lives in close_atlasskip_test.go — same package, directly callable.)

- [ ] **Step 2: Run — red.**
- [ ] **Step 3: Implement**: in the `minAhead > 0` branch — `git diff --name-only` over `newestAnchor..HEAD`; git error → fail-closed refusal (same posture as `revCount`); `!hasCodePath(paths)` → `cinfo(stderr, formatPublishGateDocsOnly(minAhead, shortSHA(newestAnchor)))`, return nil; else the augmented refusal (pinned span preserved).
- [ ] **Step 4: Run — green**; processmanual green (refusal pin `publish gate: \d+ commit\(s\) landed after` intact).
- [ ] **Step 5: Mutation-check** (lessons #63): invert the `hasCodePath` call → docs-only subtest must refuse (red) AND code subtest must pass (red); restore.
- [ ] **Step 6: Commit** — `#174: publish gate tolerates doc-only post-close deltas (leg C)`

### Task 3: reclose refusal names the new-issue path

**Files:**
- Modify: `cmd/sdlc/close.go:421`
- Test: extend whatever pins the reclose message (check `gates_test.go` / close_test.go for an existing reclose test; else assert via the die message in a small computeClose-level test)

- [ ] **Step 1:** Failing assertion: refusal contains `new issue` and `side-quest` (and still the pinned span).
- [ ] **Step 2:** Append to the die message (after the pinned span, never rewording it — the old text is also frozen in processmanual codex classifier fixtures): `Post-publish follow-up work is a new issue (side-quest: commit or sdlc issue new), not a re-close.`
- [ ] **Step 3:** Green + processmanual green.
- [ ] **Step 4: Commit** — `#174: reclose refusal names the new-issue recovery`

### Task 4: docs + bookkeeping

**Files:**
- Modify: `cmd/sdlc/helptext/close.md` (FIX-THEN-SHIP protocol under WHAT IT DOES / the review paragraph), `cmd/sdlc/helptext/merge.md` + `push.md` (publish-gate doc-tolerance) + `milestone-close.md` (protocol fires there too), `atlas/workflow/sdlc-binary.md` (publish-gate + close-review paragraphs)
- Modify: issue file (tick Plan, Log with ARCH citations)

- [ ] **Step 1:** Shadow-doc sweep: grep `reviewed-HEAD-unchanged` + `re-close` across helptext/, atlas/, AGENTS.base.md; update live explanatory copies.
- [ ] **Step 2:** Verification: `go build -o /dev/null ./cmd/sdlc/ && go vet ./cmd/sdlc/... && go test ./cmd/sdlc/...` all green; `git diff --check`.
- [ ] **Step 3:** Commit docs; then `sdlc close --issue 174 --verified '<evidence>'` (single-pass, plain checkboxes — one boundary). Done-when's re-measure line is a lagging metric; the new tests are the leading proof.

### Notes for the implementer

- gatesig pins to preserve verbatim: `close.go` reclose spans (`gatesig.go:83-84`), publish-gate refusal head (`gatesig.go:120,126`). Any NEW info/warn line goes through `assertNoGatesigCollision`.
- `hasCodePath` treats root-level `*.md` (lessons.md) as docs and `Makefile`/extensionless as code — the conservative direction for the publish gate too.
- The protocol block prints via `cwarn` (it's an action demand, not a status) — but verify `cwarn` vs `cinfo` against how the REWORK arm styles its instructions, and keep the two consistent.
