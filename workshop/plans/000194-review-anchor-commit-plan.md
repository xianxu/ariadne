---
issue: 000194
created: 2026-08-20
revised: 2026-08-20
---

# Boundary Reviews: Anchor + Memory — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make a boundary review anchored (it records the commit it read) and stateful (it
reads what it said last round), so repeated rounds converge instead of re-deriving the
whole branch from scratch every time.

**Architecture:** Two gaps, one mechanism. A boundary review today records `..HEAD` — a
floating ref — and its prose transcript is never read back. Fix the anchor first (it is
the primitive everything else measures from), then extend `gatestate.Ledger` — which is
already gate-generic and whose own comments name the close boundary as an intended user —
to the boundary review. Families and convergence ride on the ledger; round-scoping rides
on the anchor plus the ledger.

**Tech Stack:** Go, `cmd/sdlc` (cobra), `cmd/sdlc/internal/{gitx,judge,gatestate}`,
`pkg/vocab`, CUE (`construct/vocabulary/finding.cue`), `go test` with the hermetic-repo
harness (`closeRepo`, `executeSDLCTestCommand`, `judge.Run` override).

**Milestone shape:** three review boundaries, genuinely closed separately (AGENTS.md §3) —
`M1` lands a standalone defect fix, `M2` a durable artifact, `M3` the prompt behavior.
Each gets its own `sdlc milestone-close`. (A fourth, round-scoped re-review, was
considered and rejected — see M4 below.)

---

## What is already built (do not rebuild)

Establishing this cut the scope roughly in half. Verify each before writing code:

| Thing | Where | State |
|---|---|---|
| Durable review transcript | `cmd/sdlc/reviewsidecar.go`, #136 | **Built.** 86 sidecars in `workshop/history/plans/`. Prose only; no reader. |
| Stable-id findings ledger | `cmd/sdlc/internal/gatestate/`, #187 | **Built and gate-generic** — `Ledger.Gate`, `Ledger.IDPrefix` are fields. |
| Prior-round prompt block | `gatestate.RenderPriorFindings` → `judge.PromptInput.PriorFindings` | **Built.** Wired to plan-quality **only**. |
| Severity / disposition taxonomy | `construct/vocabulary/finding.cue` | **Built and shared** — a drift test already pins that the boundary review and the plan gate use one taxonomy. |
| Doc-vs-code surface classifier | `publishGateHasCodeSurface`, `publishgate.go:175`, #174 | **Built.** Reuse; do not restate. |
| Ledger IO shell pattern | `cmd/sdlc/planreview.go` | **Built.** Its comment says verbatim: *"#183's close-boundary gate will declare its own pair."* |

The extension point is explicit in the source. `planreview.go:26-30`:

```go
// planGateGate / planGateIDPrefix identify the plan-quality gate within the gate-agnostic
// gatestate package. #183's close-boundary gate will declare its own pair.
```

**M2 is that pair.** Treat any temptation to write a second ledger as a plan defect.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `reviewAnchorDelta` | `cmd/sdlc/reviewanchor.go` | new |
| `anchorOutcome` | `cmd/sdlc/reviewanchor.go` | new |
| `classifyReviewAnchor` | `cmd/sdlc/reviewanchor.go` | new |
| `formatAnchorDocsOnly` / `formatAnchorRefusal` | `cmd/sdlc/reviewanchor.go` | new |
| `closeReviewSnapshot` | `cmd/sdlc/close.go:1182` | modified |
| `resolveReviewWindow` | `cmd/sdlc/milestoneclose.go:243` | modified |
| `#Finding.family` | `construct/vocabulary/finding.cue:78` | modified |
| `gatestate.Finding.Family` | `cmd/sdlc/internal/gatestate/ledger.go:34` | modified |
| `gatestate.FamilyCounts` | `cmd/sdlc/internal/gatestate/ledger.go` | new |
| `gatestate.ConvergenceLine` | `cmd/sdlc/internal/gatestate/prompt.go` | new |

- **`reviewAnchorDelta` / `classifyReviewAnchor`** — the git-free description of
  `reviewedSHA..HEAD` and the decision over it: `anchorUnchanged`, `anchorDocsOnly`,
  `anchorCodeDelta`, `anchorDiverged`.
  - **DRY rationale:** `classifyReviewAnchor` *calls* `publishGateHasCodeSurface`; it does
    not restate the docs-vs-code rule. This is the ARCH-DRY reuse the issue names.
  - **Why a fourth state:** `Reviewed` may not be an ancestor of `Current` at all (rebase,
    `reset --hard`, branch switch). `git diff A B` between unrelated commits happily
    returns paths, so without an ancestry check a rebase-away could masquerade as
    doc-only. The publish gate's `revCount` conflates diverged with zero-ahead; do not
    copy that.
  - **Future extensions:** add `Sidecar string` if a re-attachable review ever needs to
    point back at its transcript.

- **`#Finding.family`** — a slug naming the *underlying rule* a finding is an instance of.
  - **Why the model and not just the Go struct:** `#Finding` in `finding.cue` is
    **closed** — its comment says *"a finding is an atomic judgment, not an organically
    growing record"* — so an unmodeled `family` key fails instance validation. And
    `RenderBlockInstruction` (`pkg/vocab/finding.go:85`) builds the `​```findings` fence the
    judge must emit, pulling severities and dispositions from the model. Adding the field
    in Go alone would give a judge a key it was never told to emit and a schema that
    rejects it (ARCH-PURPOSE: every consumer derives from the source).
  - **Relationships:** N:1 — many findings to one family, across rounds. The family slug
    is reviewer-assigned free text, deliberately **not** an enum: the whole point is that
    a family is discovered, not pre-declared.

- **`gatestate.FamilyCounts(l Ledger) map[string]int`** — how many findings each family
  has across all prior rounds. Pure; the input to both the escalation instruction and the
  convergence line.

- **`gatestate.ConvergenceLine(l Ledger, round int) string`** — `"round 4 — 2 new
  findings, 0 repeat families, 6 disposed. Converging."` Pure; unit-tested on in-memory
  ledgers, no IO. Reuses the existing `DispositionCounts` (`ledger.go:202`).

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `gatherReviewAnchorDelta` | `cmd/sdlc/reviewanchor.go` | new | `git` via `gitx` |
| `readBoundaryGateLedger` / `writeBoundaryGateLedger` | `cmd/sdlc/boundaryledger.go` | new | filesystem + clock |

- **`gatherReviewAnchorDelta(reviewed string)`** — `rev-parse`, `merge-base --is-ancestor`,
  `log --format=%H %s`, `gitx.DiffNames`. Errors **only** on git failure so the caller
  keeps the publish gate's fail-closed posture.
  - **Injected into:** `closeReviewSnapshot.validate()`. All decision logic stays in
    `classifyReviewAnchor` (ARCH-PURE).

- **`readBoundaryGateLedger` / `writeBoundaryGateLedger`** — a near-exact sibling of
  `readPlanGateLedger` / `writePlanGateLedger` (`planreview.go:44,66`), differing only in
  the `Gate`/`IDPrefix`/suffix triple.
  - **Injected into:** `dispatchBoundaryReview` (reads before dispatch to build
    `PriorFindings`; writes after parsing the round's fence).
  - **Critical behavior to copy verbatim:** a ledger that *exists but does not parse* is
    an **error**, never a fresh empty ledger. `planreview.go:38-42` explains why —
    silently resetting erases every disposition and re-opens findings the operator already
    addressed, *"the exact forgetting this feature exists to prevent, and worse than the
    status quo because it would look like it worked."*
  - **If the two functions end up differing only by that triple, extract the shared body**
    into a parameterized helper and have both gates call it (ARCH-DRY). Decide this after
    M2's code exists, not before — but do not leave two copies.

- **ARCH-MOCK note.** The external dependencies here are `git` and the reviewer CLI.
  Both already have established seams in this repo: `git` via the hermetic real-git temp
  repo (`hermeticrepo_test.go`, `closeRepo`), and the reviewer via the `judge.Run`
  override that every close test uses to return a canned transcript. **Reuse both. Do not
  introduce a mocked git runner** — the existing tests drive the real binary, and a second
  double would fork the conformance story.

---

## M1 — Anchor the review to the commit it read

Spec A + F. Standalone value: fixes a live defect regardless of what M2–M4 do.

### Task 1.1: Resolve the reviewed SHA once, under the lock

**Files:**
- Modify: `cmd/sdlc/milestoneclose.go:243-256` (`resolveReviewWindow`), `:576-600`
  (`boundaryReviewDispatchOptions`), `:200-212` (locked call site)
- Modify: `cmd/sdlc/close.go:1189-1196` (`captureCloseReviewSnapshot`), `:1036-1043`
- Test: `cmd/sdlc/milestonewindow_test.go`

The defect: `resolveReviewWindow` returns `head = "HEAD"`, a literal string. It reaches
`collectDiff`, `judge.PromptInput`, the trailer, and the sidecar unchanged — every one of
the 86 archived sidecars records `<base>..HEAD`. And `reviewThenFinalizeLocked` releases
the repo lock **before** `dispatchBoundaryReview`, so `boundaryReviewDispatchOptions`
re-resolves `"HEAD"` independently: the snapshot's `rev-parse` and the reviewed diff can
already name different commits.

- [ ] **Step 1: Write the failing test**

```go
// cmd/sdlc/milestonewindow_test.go
func TestResolveReviewWindow_HeadIsConcreteSHA(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	_, _, head := resolveReviewWindow("69", "", filepath.Join(issuesDir, "000069-boundary-review.md"))
	if head == "HEAD" {
		t.Fatal(`head must be a resolved SHA, not the literal "HEAD"`)
	}
	if want := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD")); head != want {
		t.Fatalf("head = %q, want %q", head, want)
	}
}
```

- [ ] **Step 2: Run it, confirm it fails**

Run: `go test ./cmd/sdlc/ -run TestResolveReviewWindow_HeadIsConcreteSHA -v`
Expected: FAIL — `head must be a resolved SHA`.

- [ ] **Step 3: Resolve the head in `resolveReviewWindow`**

```go
// head is the CONCRETE SHA the review will read (#194) — the anchor the finalize
// check classifies against and the ledger measures the next round from, so it must
// be pinned under the caller's lock. Falls back to "HEAD" only when rev-parse fails,
// preserving the documented ("?", "", "HEAD") no-anchor return shape.
head = "HEAD"
if sha := strings.TrimSpace(gitx.Capture("rev-parse", "HEAD")); sha != "" {
	head = sha
}
```

- [ ] **Step 4: Use the pinned head everywhere downstream**

In `boundaryReviewDispatchOptions`, replace both literal `"HEAD"` uses with `p.Head`
(the `collectDiff` call and `judge.PromptInput{..., Head: ...}`). Update
`dispatchBoundaryReview`'s progress line to `shortSHA(p.Head)`.

- [ ] **Step 5: Pass the SHA into the snapshot instead of re-reading it**

```go
// captureCloseReviewSnapshot pins the state the boundary review is about to read.
// reviewedSHA is the window head the caller already resolved (#194) — passing it in
// rather than re-`rev-parse`ing guarantees the snapshot, the reviewed diff, and the
// finalize check all name the SAME commit.
func captureCloseReviewSnapshot(r closeResult, reviewedSHA, milestone string) closeReviewSnapshot
```

Both call sites already compute `head` one line above, inside the lock — pass it, plus
`f.Milestone` (M1 Task 1.2 needs it to name the right re-run verb via `closeVerb`).

- [ ] **Step 6: Update trailer/sidecar expectations**

`emitTrailerBlock` now renders `Review-Window: <base>..<shortHead>`. `Review-Window` has
**no production parser** — confirm with
`grep -rn "Review-Window" --include='*.go' cmd/ pkg/ | grep -v _test` (expect only the
writer at `milestoneclose.go:406` and help text). Update the assertions at
`closereview_test.go:226` and `milestoneclose_test.go:120,134`, the doc comment at
`milestoneclose.go:395`, `helptext/milestone-close.md:35`, and
`atlas/workflow/ledger-landscape.md:102`.

Fixtures that *construct* commit messages containing `Review-Window: abc1234..HEAD`
(`close_test.go:564,611,653`, `closereview_test.go:485`, `milestonewindow_test.go:85,164`)
feed the `Review-Verdict:` grep and need **no** change.

- [ ] **Step 7: Run the package** — `go test ./cmd/sdlc/ 2>&1 | tail -20`. Expected: PASS.
- [ ] **Step 8: Commit** — `#194 M1: pin the boundary review to a concrete reviewed SHA`

### Task 1.2: Classify the mid-review delta

**Files:** Create `cmd/sdlc/reviewanchor.go` + `reviewanchor_test.go`; modify
`cmd/sdlc/close.go:1128-1152,1182-1227`; test `cmd/sdlc/close_finalize_test.go`.

- [ ] **Step 1: Write the failing unit tests** — table over `classifyReviewAnchor` covering
      all four outcomes, plus `cmd/sdlc/helptext/close.md` classified as **code** (it is
      `//go:embed`ed — `publishGateHasCodeSurface` tightens `hasCodePath` for exactly this),
      plus two format tests:
  - the pass line shares **no** vocabulary with the refusal (`NOT finalized`,
    `unreviewed`, `re-run`, `stale`) — `gatesig` classifies transcripts by substring, so a
    pass line echoing refusal words corrupts friction attribution (#172, and
    `formatPublishGateDocsOnly`'s comment says so);
  - the refusal names **every** commit (short SHA + subject) and does not contain
    `"HEAD changed from"`.
- [ ] **Step 2: Run, confirm they fail** (`undefined: reviewAnchorDelta`).
- [ ] **Step 3: Write `reviewanchor.go`** — `reviewAnchorDelta`, `anchorOutcome`,
      `classifyReviewAnchor` (pure), `gatherReviewAnchorDelta` (thin IO shell),
      `formatAnchorDocsOnly`, `formatAnchorRefusal`. `classifyReviewAnchor` delegates the
      code-surface question to `publishGateHasCodeSurface`.
- [ ] **Step 4: Run, confirm they pass.**
- [ ] **Step 5: Write the failing integration test** — model on
      `TestCloseCommand_HEADChangedDuringBoundaryReview_DoesNotFinalize`
      (`close_finalize_test.go:139`), which already blocks the fake reviewer on a channel
      and commits concurrently. Add a doc-only twin that commits `workshop/lessons.md`
      mid-review and asserts the close **finalizes** (`status: codecomplete` present) with
      the pass line on stderr. Extend the existing code-delta test to assert the refusal
      names `"concurrent #69 side change"` and lacks `"HEAD changed from"`.
- [ ] **Step 6: Rewrite `validate()`** to `func() (note string, err error)`: on
      `anchorDocsOnly` return the pass line; on `anchorCodeDelta`/`anchorDiverged` return
      `formatAnchorRefusal(..., closeVerb(s.milestone))`; fail closed on a git error. The
      issue-file and project-file checks stay **strict and unchanged** — the review read
      that prose, so a mid-review edit to it is a genuine invalidation.
- [ ] **Step 7: Update `finalizeBoundaryReview`** to surface `note` via `cinfo`. Drop the
      now-redundant second `cwarn` (`"re-run … so the review covers the current repo
      state"`) — `formatAnchorRefusal` carries a precise instruction. Confirm nothing
      asserts on it: `grep -rn "so the review covers the current repo state" cmd/sdlc/`.
- [ ] **Step 8: `go test ./cmd/sdlc/`** → PASS. **Commit.**

**Why a code delta must still refuse.** `runPublishGate` anchors on
`codecompleteAnchorCommit` — the *close commit*. Finalizing above an unreviewed code delta
would put the close commit on top of it; merge would compute `closeCommit..HEAD` = 0,
print `reviewed-HEAD-unchanged ✓`, and ship code no reviewer read. Making the other branch
safe means re-anchoring the publish gate — out of scope (see the issue's Spec F).

- [ ] **Task 1.3: `sdlc milestone-close --issue 194 --milestone M1`**

---

## M2 — Give the boundary review the ledger

Spec B. The pair `planreview.go` says is coming.

### Task 2.1: Declare the boundary gate's pair

**Files:** Create `cmd/sdlc/boundaryledger.go` + test; modify
`construct/vocabulary/finding.cue:64-69` (discovery glob).

- [ ] **Step 1:** Widen `finding.cue`'s `discovery.glob` — it is currently
      `"*-plan-gate.md"` (singular). A boundary ledger is also a set of finding instances,
      so discovery must find both. Check whether the field is consumed as a single glob or
      a list (`grep -rn "Glob" pkg/vocab/finding.go`) and widen accordingly; if the shape
      must change from string to list, update every consumer — a hand-maintained
      restatement is a deferred consumer (ARCH-PURPOSE).
- [ ] **Step 2:** Write `boundaryledger.go` mirroring `planreview.go`:
      `boundaryGateSuffix = "close-gate"`, `boundaryGateGate = "boundary-review"`,
      `boundaryGateIDPrefix = "BR"`. Reuse `sidecarPathFor` (already shared, `#144`) and
      `atomicWriteFile`.
      - **Do not** name it `*-review.md`: `construct/vocabulary/verdict.cue` declares
        `discovery.glob: "*-review.md"` to assert "this document carries a boundary
        verdict". A gate ledger carries findings and no verdict. `planreview.go:9-11`
        explains this trap — the boundary gate is *more* likely to fall into it, since its
        prose sidecar legitimately IS `*-review.md`.
      - Milestone-aware naming: the plan gate has one ledger per issue; the boundary
        review has one per **boundary**. Decide `-m2-close-gate.md` vs one issue-wide
        ledger. **Recommendation: one ledger per issue, rounds tagged with their
        boundary** — families recur *across* milestones (the `tools#1` evidence spans M1
        rounds *and* the close review), and a per-boundary ledger cannot see that. This is
        a deliberate divergence from `sidecarPath`'s per-boundary shape; record it in
        `## Log`.
- [ ] **Step 3:** Copy `readPlanGateLedger`'s **parse-failure-is-an-error** behavior
      verbatim, with a test that a corrupt ledger errors rather than silently emptying.
- [ ] **Step 4:** If read/write differ from the plan-gate pair only by the
      `Gate`/`IDPrefix`/suffix triple, extract the shared body and have both call it
      (ARCH-DRY). Commit.

### Task 2.2: Wire `PriorFindings` into the boundary review

**Files:** `cmd/sdlc/internal/judge/prompts.go:174-177`,
`cmd/sdlc/internal/judge/prompts/milestone-review.md`, `cmd/sdlc/milestoneclose.go:529-600`.

- [ ] **Step 1:** Write a failing test: a second boundary review on an issue with an
      existing ledger renders the prior findings into its prompt. Use
      `judge.BuildPrompt(judge.MilestoneReview, in)` directly — no dispatch needed.
- [ ] **Step 2:** Add `{{PRIOR_FINDINGS}}` and `{{FINDINGS_BLOCK}}` to
      `milestone-review.md`, mirroring `plan-quality.md:8-18,85`. Update the
      `PromptInput.PriorFindings` doc comment — it currently reads *"Empty for every
      category but plan-quality"* and would become false. **`golden_test.go` pins
      `BuildPrompt` output byte-for-byte**; regenerate its fixtures deliberately and read
      the diff.
- [ ] **Step 3:** In `dispatchBoundaryReview`: read the ledger before building the prompt,
      parse the `​```findings` fence from the output, `AssignIDs`, append the round, write
      the ledger. The prose sidecar keeps being written unchanged — two artifacts, two
      consumers.
- [ ] **Step 4:** Make an undisposed blocking finding refuse the boundary, reusing
      `gatestate`'s existing decision code (`decide.go`) rather than a new rule. Honor
      `WF_PLAN_ROUND_CAP`'s analogue: past the cap only `hardBlocking` (Critical) blocks.
- [ ] **Step 5:** `go test ./...` → PASS. Commit.
- [ ] **Task 2.3: `sdlc milestone-close --issue 194 --milestone M2`**

---

## M3 — Families, escalation, convergence

Spec C + D. This is the milestone that carries the actual insight; M1 and M2 are its
prerequisites.

- [ ] **Task 3.1:** Add `family` to `#Finding` in `construct/vocabulary/finding.cue`
      (**closed schema** — an unmodeled key fails instance validation), to
      `gatestate.Finding`, to the parser, and to `RenderBlockInstruction`
      (`pkg/vocab/finding.go:97-100`) so the judge is *told* to emit it. Optional on
      emission, so an older transcript still parses. Test the round-trip.
- [ ] **Task 3.2:** `gatestate.FamilyCounts(l Ledger) map[string]int` — pure, unit-tested
      on in-memory ledgers.
- [ ] **Task 3.3:** Render the escalation instruction into `RenderPriorFindings` when a
      family already has ≥1 prior finding, in the issue's words: *"This is the Nth finding
      in family `X`. Rounds … fixed instances. Do not fix this instance — state the rule
      that covers all of them, and fix that. If the rule cannot be stated, say why, and
      record the family in `Limits` with its measured prevalence."*
- [ ] **Task 3.4:** `gatestate.ConvergenceLine` (pure, reusing `DispositionCounts`), and
      emit it with the verdict.
- [ ] **Task 3.5: Verify against real history.** `tools#1`'s
      `000001-define-m1-review.md` has four rounds and two clear families. Build it as a
      fixture and assert a correct implementation flags `block-opener-rule` at **round 2,
      not round 3**. This is the issue's own acceptance test and the only one that
      measures whether the feature would have worked.
      - Reading a peer repo: per AGENTS.md, read `tools`' `AGENTS.local.md` + `MEMORY.md`,
        not its `AGENTS.md`. Copy the fixture into `cmd/sdlc/internal/gatestate/testdata/`
        — do not make the test depend on a sibling checkout existing.
- [ ] **Task 3.6: Pin the window.** Regression test: a whole-issue close still resolves
      its review window to `merge-base(main, HEAD)`. M4 was rejected; this test is what
      keeps it rejected.
- [ ] **Task 3.7: `sdlc milestone-close --issue 194 --milestone M3`**

---

## M4 — REJECTED (recorded, not built)

Round-scoping a re-review to `lastReviewedSHA..HEAD` was considered and **rejected by
the operator on 2026-08-20**: the reviewer keeps reading the whole branch.

The reasoning, kept so it is not re-proposed: it was the only scope item that would
have shortened an individual round, but it means no single reviewer ever reads the
integrated branch, and the whole-issue window is `merge-base(main, HEAD)` *by design*
(#77) so the final review sees what ships. M2's recorded coverage certifies that
findings were handled — not that anyone read the result of handling them.

**Consequence for this plan:** the wall-clock win comes entirely from M2+M3 reducing
the NUMBER of rounds, not the size of each. That is also what the evidence supports —
family escalation would have collapsed at least two of `tools#1`'s four M1 rounds.

**Consequence for the code:** `boundaryWindowBase` is **unmodified**. Add a regression
test in M3 pinning that a whole-issue close still resolves its window to
`merge-base(main, HEAD)`, so a later change cannot quietly narrow it.

## Verification

- [ ] `go build ./... && go test ./... 2>&1 | tail -20` — all packages pass.
- [ ] `sdlc process-manual` renders without error (prompt templates changed).
- [ ] Vocabulary conformance: the `finding.cue` edits pass whatever `make` target vets CUE
      instances — find it (`grep -rn "cue vet" Makefile scripts/`) and run it.
- [ ] Self-hosting check: this issue's own M2/M3 boundary reviews should produce a
      `-close-gate.md` ledger with stable ids. If M3 lands and a repeat family shows up in
      this very issue's rounds, say so in `## Log` — that is the feature working.

## Out of scope

- Re-anchoring `runPublishGate` on a recorded reviewed-SHA (needed to finalize *through* a
  code delta; a change to the publish contract).
- Re-attachable / concurrent reviews — the issue lists these as a second-order effect this
  work merely unblocks.
- `revCount`'s diverged/zero-ahead conflation in the publish gate.
- Replacing the #136 prose sidecar. It has a different consumer (a human or a resuming
  agent) and stays.
