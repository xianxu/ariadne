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
floating ref (66 of the 70 archived sidecar files say so) — and its prose transcript is never
read back. Fix the anchor first (it is
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
| Durable review transcript | `cmd/sdlc/reviewsidecar.go`, #136 | **Built.** 70 sidecar files archived (86 window rows — a re-run appends `## Re-review`). Prose only; no reader. |
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

## Decisions this plan makes (do not re-litigate in code)

The plan-quality gate blocked round 1 on four unstated seam decisions. They are settled
here; each names the alternative it beat.

### D1 — One ledger per issue, but the round cap and open-findings scope per BOUNDARY

`gatestate.Decide` computes `CapReached: len(l.Rounds) > roundCap` with
`DefaultRoundCap = 3` (`decide.go:14,45`), and `OpenFindings(l)` spans the whole ledger.
So a naive issue-wide boundary ledger would arrive at the whole-issue close already past
the cap — silently demoting every Important finding on round 1 of the gate this issue
exists to strengthen — and would let an M1 finding left `not-addressed` block M3.

Resolution, which keeps both properties:

- **One ledger file per issue** (`NNNNNN-slug-close-gate.md`). Families must be visible
  across boundaries — the `tools#1` evidence spans M1 rounds *and* the close review, and
  a per-boundary file cannot see that recurrence, which is the whole point of M3.
- **Add `Boundary string` to `gatestate.Round`** (`ledger.go:57-75`) — `"M1"`, `"M2"`,
  `""` for the whole-issue close. It is a schema addition, listed in Core concepts.
- **Add pure `gatestate.FilterBoundary(l Ledger, boundary string) Ledger`**, returning a
  view with only that boundary's rounds. The boundary gate calls
  `Decide(FilterBoundary(l, b), cap)` and `OpenFindings(FilterBoundary(l, b))`.
  `FamilyCounts(l)` takes the **unfiltered** ledger.

`Decide` and `OpenFindings` keep their signatures, so plan-quality is untouched — the
scoping is a caller-side pure transform, not a widened API (ARCH-PURE).

*Rejected alternative:* per-boundary ledger files, mirroring the prose sidecar's
`-m2-review.md` shape. Simpler, cap works for free, but blind to cross-boundary families.

### D2 — ONE carry-forward channel, not two

`code-review.md:57-67` ("Plan-gate carry-forward", #187) already instructs the boundary
reviewer to read `workshop/plans/<stem>-plan-gate.md` off disk and re-raise its open
findings "at its original severity". Those carry `PQ-*` ids. Adding a `{{PRIOR_FINDINGS}}`
block rendered from the `BR-*` ledger would hand the reviewer two id namespaces and one
output fence, with no rule for a disposition naming an id the `BR` ledger has never seen.

Resolution:

- On the **first** boundary round for an issue, **seed** the `BR` ledger from the plan
  gate's still-open findings: each is re-issued as `BR-n` with `severity` and `family`
  preserved and a note recording its `PQ-n` origin.
- **Delete `code-review.md`'s "Plan-gate carry-forward" section** in the same commit as
  the seeding — never before, or the deferred findings vanish for one release. After
  this, the rendered `{{PRIOR_FINDINGS}}` block is the reviewer's only prior-findings
  input, in one namespace (ARCH-DRY: one mechanism, not two).
- A disposition naming an **unknown id**: warn to stderr, drop the disposition, and
  record it on the round as a protocol note. Never crash, and never invent the finding.
- Note the blast radius: `code-review.md` is embedded in the ad-hoc `sdlc judge` path too,
  so its prose changes there as well. That is correct — the file-reading instruction was
  always a stand-in for a rendered block.

### D3 — Family slugs are anchored three ways, and the test must prove it

`FamilyCounts` keys on exact string equality and `RenderPriorFindings`
(`gatestate/prompt.go:17-70`) renders `id / severity / title / detail` — **no family**. A
stateless reviewer writing `block-opener-rule` at round 2 and `block-opener` at round 3
leaves every count at 1, and the escalation instruction — the issue's stated purpose —
never fires. A fixture hand-built with consistent slugs passes by construction while the
live behavior fails (ARCH-PURPOSE).

Three mechanisms, because no one of them is sufficient:

1. **Render the in-play family vocabulary** into the prior-findings block, with an
   explicit instruction: *reuse an existing slug when the finding belongs to that family;
   coin a new one only when it genuinely does not.* This is what catches synonyms.
2. **Normalize on ingest** — `normalizeFamily`: lowercase, trim, non-alphanumeric runs to
   a single hyphen, strip leading/trailing hyphens. This catches casing and punctuation
   drift (`Block Opener Rule` → `block-opener-rule`), and **nothing else**.
3. **Test the gap explicitly** — a fixture whose rounds use *near-miss* slugs
   (`Block Opener Rule`, `block_opener_rule`) must still collapse to one family; and a
   fixture using a true synonym (`block-opener`) documents the residual risk in the test's
   own name rather than pretending it is solved.

Residual risk, accepted and recorded: a reviewer that coins a genuine synonym despite
seeing the vocabulary will under-count. Mechanism 1 makes that unlikely, not impossible.

### D5 — Seeded plan-gate findings are boundary-agnostic

D1 and D2 do not compose on their own: if the D2 seed round is stamped
`Boundary: "M1"`, then `OpenFindings(FilterBoundary(l, ""))` at the whole-issue close
never sees it — a **regression** against today, where `code-review.md:57` makes *every*
boundary reviewer read the plan-gate ledger.

Resolution: seed rounds carry the sentinel `gatestate.BoundaryAll = "*"`, and
`FilterBoundary(l, b)` retains a round when `r.Boundary == b || r.Boundary == BoundaryAll`.
A plan-gate finding was deferred to "the boundary review" generically, not to one
milestone, so it stays visible at every boundary until disposed — matching the behavior
being replaced.

*Rejected alternative:* stamping the seed round with an empty boundary. `""` is already
the whole-issue close's own value, so seeded findings would be invisible at M1–M3 —
the same bug pointing the other way.

Add a test that a seeded finding is visible from **both** a milestone boundary and the
whole-issue close.

### D4 — Verdict AND ledger must both clear; a boundary protocol miss warns

`closeVerdictOutcome` (`close.go:1064-1082`) derives finalize/rework/halt from
`vocab.Verdict()`. M2 adds a ledger-derived refusal beside it. Precedence:

- **Finalize iff the verdict is finalizing AND no blocking finding is undisposed** — an
  AND, not a fallback. SHIP with an undisposed Important means the reviewer contradicted
  itself; refuse and say which finding. REWORK with everything disposed still reworks.
- **A boundary protocol miss warns and persists — it does NOT halt.**

  > **Revised during M2 implementation.** This clause originally said *halt*, on the
  > reasoning that falling back would finalize a close carrying no ledger memory. That
  > reasoning was made without two facts that only surfaced in the code:
  >
  > 1. **The fallback's failure mode is the status quo, not a regression.** A round with
  >    no fence leaves the next round blind — which is exactly how every boundary review
  >    behaved before this milestone. Halting trades a known-tolerable state for a hard
  >    stop.
  > 2. **The reviewer is an LLM, and the only escapes are worse than the problem.** Fence
  >    compliance will not be 100%. The two ways past a halt are `--no-judge` and
  >    `--force`, both of which skip the review *entirely* — so a strict halt would
  >    convert an occasional formatting miss into a routine reason to run no review at
  >    all. That is the opposite of what this milestone is for.
  >
  > `finding.cue`'s own posture backs the softer read: *"the schema'd path is
  > authoritative; a fallback may exist transitionally."* So: mirror plan-quality — warn
  > loudly, persist the `ProtocolError` round (never drop it, or `len(Rounds)` stays 0
  > forever and the round cap can never fire), and let the verdict token decide. The
  > round is marked `Blocked` so the miss is visible in the ledger and in the close-time
  > metrics rather than being silently absorbed.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `reviewAnchorDelta` | `cmd/sdlc/reviewanchor.go` | new |
| `anchorOutcome` | `cmd/sdlc/reviewanchor.go` | new |
| `classifyReviewAnchor` | `cmd/sdlc/reviewanchor.go` | new |
| `formatAnchorDocsOnly` / `formatAnchorRefusal` | `cmd/sdlc/reviewanchor.go` | new |
| `#Finding.family` | `construct/vocabulary/finding.cue:91` | modified |
| `gatestate.Finding.Family` | `cmd/sdlc/internal/gatestate/ledger.go:43` | modified |
| `gatestate.Round.Boundary` | `cmd/sdlc/internal/gatestate/ledger.go:57` | modified (D1) |
| `gatestate.FilterBoundary` | `cmd/sdlc/internal/gatestate/ledger.go` | new (D1) |
| `gatestate.BoundaryAll` | `cmd/sdlc/internal/gatestate/ledger.go` | new (D5) |
| `gatestate.CountedRounds` | `cmd/sdlc/internal/gatestate/ledger.go` | new (M2 review) |
| `gatestate.Round.NoCap` | `cmd/sdlc/internal/gatestate/ledger.go` | modified (M2 review) |
| `gatestate.DecideScoped` | `cmd/sdlc/internal/gatestate/decide.go` | new (close review BR-37) |
| `gateLedgerKind` | `cmd/sdlc/planreview.go` | new (M2, shared by both gates) |
| `openScopeFor` | `cmd/sdlc/boundaryledger.go` | new (close review BR-37) |
| `gatestate.FamilyCounts` | `cmd/sdlc/internal/gatestate/family.go` | new |
| `gatestate.NormalizeFamily` | `cmd/sdlc/internal/gatestate/family.go` | new (D3; wraps `issue.Slugify`) |
| `gatestate.ConvergenceLine` | `cmd/sdlc/internal/gatestate/family.go` | new |
| `gatestate.RenderPriorFindingsScoped` | `cmd/sdlc/internal/gatestate/prompt.go` | new (M3 review BR-20) |

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

> **Corrected after M1's boundary review (I6).** `closeReviewSnapshot` and
> `resolveReviewWindow` were listed under Pure entities. Neither is —
> `validate()` does `os.ReadFile` plus git, and `resolveReviewWindow` shells
> `rev-parse` directly (M1 made it *more* impure). The code satisfies ARCH-PURE;
> the table was wrong, and was wrong before M1 touched it.

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `gatherReviewAnchorDelta` | `cmd/sdlc/reviewanchor.go` | new | `git` via `gitx` |
| `closeReviewSnapshot` | `cmd/sdlc/close.go:1182` | modified | filesystem + `git` |
| `resolveReviewWindow` | `cmd/sdlc/milestoneclose.go:243` | modified | `git` |
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
`collectDiff`, `judge.PromptInput`, the trailer, and the sidecar unchanged — 66 of the 70
archived sidecar files record `<base>..HEAD` across 84 of 86 window rows (a re-run
appends a `## Re-review` section, so rows outnumber files). And `reviewThenFinalizeLocked` releases
the repo lock **before** `dispatchBoundaryReview`, so `boundaryReviewDispatchOptions`
re-resolves `"HEAD"` independently: the snapshot's `rev-parse` and the reviewed diff can
already name different commits.

- [x] **Step 1: TDD the resolution.** Test that `resolveReviewWindow` returns a concrete
      SHA equal to `rev-parse HEAD`, not the literal `"HEAD"`. Red first.

- [x] **Step 2: Resolve the head** in `resolveReviewWindow`, keeping the documented
      `("?", "", "HEAD")` no-anchor return when `rev-parse` fails.

- [x] **Step 3: Spend the pinned head downstream.** `boundaryReviewDispatchOptions` passes
      literal `"HEAD"` twice — to `collectDiff` and to `judge.PromptInput`. Both become
      `p.Head`. **This is the actual defect fix**: those calls run after the repo lock is
      released, so today they can resolve a different commit than the snapshot recorded.

- [x] **Step 4: Take the SHA as a parameter.** `captureCloseReviewSnapshot(r, reviewedSHA,
      milestone)` — both call sites already compute `head` one line above, inside the lock.
      The `milestone` argument is for Task 1.2's refusal, which names the re-run verb via
      the existing `closeVerb`.

- [x] **Step 5: Follow the compile errors for the trailer/sidecar surface.** `Review-Window`
      becomes `<base>..<shortHead>`. Confirm it has no production parser before changing it
      — `grep -rn "Review-Window" --include='*.go' cmd/ pkg/ | grep -v _test` should show
      only the writer. Then update whatever assertions and help text the build and test run
      surface, plus `atlas/workflow/ledger-landscape.md`.

      Two traps worth knowing rather than a file list (a line-numbered inventory goes stale
      before it is read — and one entry in this plan's first draft already was):
      **(a)** fixtures that *construct* commit messages containing
      `Review-Window: abc1234..HEAD` feed the `Review-Verdict:` grep and must NOT change;
      **(b)** table-driven renderer tests that pass `Head: "HEAD"` as fixture *input* also
      need no change. Only assertions on *produced* output do.

- [x] **Step 6: Run the package** — `go test ./cmd/sdlc/ 2>&1 | tail -20`. Expected: PASS.
- [x] **Step 7: Commit** — `#194 M1: pin the boundary review to a concrete reviewed SHA`

### Task 1.2: Classify the mid-review delta

**Files:** Create `cmd/sdlc/reviewanchor.go` + `reviewanchor_test.go`; modify
`cmd/sdlc/close.go:1128-1152,1182-1227`; test `cmd/sdlc/close_finalize_test.go`.

- [x] **Step 1: Write the failing unit tests** — table over `classifyReviewAnchor` covering
      all four outcomes, plus `cmd/sdlc/helptext/close.md` classified as **code** (it is
      `//go:embed`ed — `publishGateHasCodeSurface` tightens `hasCodePath` for exactly this),
      plus two format tests:
  - the pass line shares **no** vocabulary with the refusal (`NOT finalized`,
    `unreviewed`, `re-run`, `stale`) — `gatesig` classifies transcripts by substring, so a
    pass line echoing refusal words corrupts friction attribution (#172, and
    `formatPublishGateDocsOnly`'s comment says so);
  - the refusal names **every** commit (short SHA + subject) and does not contain
    `"HEAD changed from"`.
- [x] **Step 2: Run, confirm they fail** (`undefined: reviewAnchorDelta`).
- [x] **Step 3: Write `reviewanchor.go`** — `reviewAnchorDelta`, `anchorOutcome`,
      `classifyReviewAnchor` (pure), `gatherReviewAnchorDelta` (thin IO shell),
      `formatAnchorDocsOnly`, `formatAnchorRefusal`. `classifyReviewAnchor` delegates the
      code-surface question to `publishGateHasCodeSurface`.
- [x] **Step 4: Run, confirm they pass.**
- [x] **Step 5: Write the failing integration test** — model on
      `TestCloseCommand_HEADChangedDuringBoundaryReview_DoesNotFinalize`
      (`close_finalize_test.go:139`), which already blocks the fake reviewer on a channel
      and commits concurrently. Add a doc-only twin that commits `workshop/lessons.md`
      mid-review and asserts the close **finalizes** (`status: codecomplete` present) with
      the pass line on stderr. Extend the existing code-delta test to assert the refusal
      names `"concurrent #69 side change"` and lacks `"HEAD changed from"`.
- [x] **Step 6: Rewrite `validate()`** to `func() (note string, err error)`: on
      `anchorDocsOnly` return the pass line; on `anchorCodeDelta`/`anchorDiverged` return
      `formatAnchorRefusal(..., closeVerb(s.milestone))`; fail closed on a git error. The
      issue-file and project-file checks stay **strict and unchanged** — the review read
      that prose, so a mid-review edit to it is a genuine invalidation.
- [x] **Step 7: Update `finalizeBoundaryReview`** to surface `note` via `cinfo`. Drop the
      now-redundant second `cwarn` (`"re-run … so the review covers the current repo
      state"`) — `formatAnchorRefusal` carries a precise instruction. Confirm nothing
      asserts on it: `grep -rn "so the review covers the current repo state" cmd/sdlc/`.
- [x] **Step 8: `go test ./cmd/sdlc/`** → PASS. **Commit.**

**Why a code delta must still refuse.** `runPublishGate` anchors on
`codecompleteAnchorCommit` — the *close commit*. Finalizing above an unreviewed code delta
would put the close commit on top of it; merge would compute `closeCommit..HEAD` = 0,
print `reviewed-HEAD-unchanged ✓`, and ship code no reviewer read. Making the other branch
safe means re-anchoring the publish gate — out of scope (see the issue's Spec F).

- [x] **Task 1.3: `sdlc milestone-close --issue 194 --milestone M1`**

---

## M2 — Give the boundary review the ledger

Spec B. The pair `planreview.go` says is coming.

### Task 2.1: Declare the boundary gate's pair

**Files:** Create `cmd/sdlc/boundaryledger.go` + test; modify
`construct/vocabulary/finding.cue:64-69` (discovery glob).

- [x] **Step 1:** Widen `finding.cue`'s `discovery.glob` (currently `"*-plan-gate.md"`)
      to cover the boundary ledger too. **This is a one-line documentation edit** —
      `FindingModel` (`pkg/vocab/finding.go:22-28`) has no `Disc` field, unlike
      `IssueModel`/`ProjectModel`, so `discovery` has no Go consumer to update. Do not
      widen the Go model "for symmetry"; add the field when a consumer needs it.
- [x] **Step 2:** Write `boundaryledger.go` mirroring `planreview.go`:
      `boundaryGateSuffix = "close-gate"`, `boundaryGateGate = "boundary-review"`,
      `boundaryGateIDPrefix = "BR"`. Reuse `sidecarPathFor` (already shared, `#144`) and
      `atomicWriteFile`.
      - **Do not** name it `*-review.md`: `construct/vocabulary/verdict.cue` declares
        `discovery.glob: "*-review.md"` to assert "this document carries a boundary
        verdict". A gate ledger carries findings and no verdict. `planreview.go:9-11`
        explains this trap — the boundary gate is *more* likely to fall into it, since its
        prose sidecar legitimately IS `*-review.md`.
      - Ledger shape and round-cap scoping are settled in **D1** — one file per issue,
        `Round.Boundary` added, `FilterBoundary` applied at the call site. Implement D1;
        do not re-derive it here.

- [x] **Step 3:** Copy `readPlanGateLedger`'s **parse-failure-is-an-error** behavior
      verbatim, with a test that a corrupt ledger errors rather than silently emptying.
- [x] **Step 4:** If read/write differ from the plan-gate pair only by the
      `Gate`/`IDPrefix`/suffix triple, extract the shared body and have both call it
      (ARCH-DRY). Commit.

### Task 2.2: Wire `PriorFindings` into the boundary review

**Files:** `cmd/sdlc/internal/judge/prompts.go:174-177`,
`cmd/sdlc/internal/judge/prompts/milestone-review.md`, `cmd/sdlc/milestoneclose.go:529-600`.

- [x] **Step 1:** Write a failing test: a second boundary review on an issue with an
      existing ledger renders the prior findings into its prompt. Use
      `judge.BuildPrompt(judge.MilestoneReview, in)` directly — no dispatch needed.
- [x] **Step 2:** Add `{{PRIOR_FINDINGS}}` and `{{FINDINGS_BLOCK}}` to
      `milestone-review.md`, mirroring `plan-quality.md:8-18,85`. Update the
      `PromptInput.PriorFindings` doc comment — it currently reads *"Empty for every
      category but plan-quality"* and would become false. **`golden_test.go` pins
      `BuildPrompt` output byte-for-byte**; regenerate its fixtures deliberately and read
      the diff.
- [x] **Step 3:** In `dispatchBoundaryReview`: read the ledger before building the prompt,
      **seed it from the plan gate's open findings on the first boundary round (D2)**,
      parse the `​```findings` fence from the output, `AssignIDs`, append the round stamped
      with its `Boundary` (D1), write the ledger. The prose sidecar keeps being written
      unchanged — two artifacts, two consumers.
- [x] **Step 4:** Delete `code-review.md`'s "Plan-gate carry-forward" section **in this
      same commit** as the seeding (D2) — earlier and the deferred findings vanish for a
      release; later and two channels coexist. Handle an unknown-id disposition by warning
      and dropping, never crashing.
- [x] **Step 5:** Make an undisposed blocking finding refuse the boundary via
      `Decide(FilterBoundary(l, boundary), cap)` (D1) — reuse `decide.go`, do not write a
      second rule. Wire the verdict/ledger precedence and the protocol-miss halt per **D4**.
- [x] **Step 6:** `go test ./...` → PASS. Commit.
- [x] **Task 2.3: `sdlc milestone-close --issue 194 --milestone M2`**

---

## M3 — Families, escalation, convergence

Spec C + D. This is the milestone that carries the actual insight; M1 and M2 are its
prerequisites.

- [x] **Task 3.1:** Add `family` to `#Finding` in `construct/vocabulary/finding.cue`
      (**closed schema** — an unmodeled key fails instance validation), to
      `gatestate.Finding`, to the parser, and to `RenderBlockInstruction`
      (`pkg/vocab/finding.go:97-100`) so the judge is *told* to emit it. Optional on
      emission, so an older transcript still parses. Test the round-trip.
- [x] **Task 3.2:** `gatestate.FamilyCounts(l Ledger) map[string]int` (unfiltered ledger,
      per D1) and `normalizeFamily` — both pure, unit-tested on in-memory ledgers. Per
      **D3**, `FamilyCounts` normalizes before counting.
- [x] **Task 3.3:** Render **the in-play family vocabulary plus a reuse instruction** into
      `RenderPriorFindings` (D3 mechanism 1 — `RenderPriorFindings` renders no family
      today, which is why escalation would otherwise never fire), and the escalation
      instruction when a family already has ≥1 prior finding, in the issue's words: *"This is the Nth finding
      in family `X`. Rounds … fixed instances. Do not fix this instance — state the rule
      that covers all of them, and fix that. If the rule cannot be stated, say why, and
      record the family in `Limits` with its measured prevalence."*
- [x] **Task 3.4:** `gatestate.ConvergenceLine` (pure, reusing `DispositionCounts`), and
      emit it with the verdict.
- [x] **Task 3.5: Verify against real history.** `tools#1`'s
      `000001-define-m1-review.md` has four rounds and two clear families. Build it as a
      fixture and assert a correct implementation flags `block-opener-rule` at **round 2,
      not round 3**. This is the issue's own acceptance test and the only one that
      measures whether the feature would have worked.
      - Reading a peer repo: per AGENTS.md, read `tools`' `AGENTS.local.md` + `MEMORY.md`,
        not its `AGENTS.md`. Copy the fixture into `cmd/sdlc/internal/gatestate/testdata/`
        — do not make the test depend on a sibling checkout existing.
      - **Also required by D3:** a near-miss-slug fixture (`Block Opener Rule`,
        `block_opener_rule`) that must collapse to one family, and a true-synonym fixture
        (`block-opener`) whose test name records the accepted residual risk. Without
        these, Task 3.5's consistent-slug fixture passes by construction.
- [x] **Task 3.6: Pin the window.** Regression test: a whole-issue close still resolves
      its review window to `merge-base(main, HEAD)`. M4 was rejected; this test is what
      keeps it rejected.
- [x] **Task 3.7: `sdlc milestone-close --issue 194 --milestone M3`**

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

- [x] `go build ./... && go test ./... 2>&1 | tail -20` — all packages pass.
- [x] `sdlc process-manual` renders without error (prompt templates changed).
- [x] Vocabulary conformance: the `finding.cue` edits pass whatever `make` target vets CUE
      instances — find it (`grep -rn "cue vet" Makefile scripts/`) and run it.
- [x] Self-hosting check: this issue's own M2/M3 boundary reviews should produce a
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

---

## Revisions

Appended per AGENTS.md §1 — mid-stream changes are recorded here, not folded silently
into the prose above.

### 2026-08-20 — M4 rejected (operator)

Round-scoping a re-review to `lastReviewedSHA..HEAD` was dropped: the reviewer keeps
reading the whole branch. See the `## M4 — REJECTED` section for the reasoning and the
regression test that keeps it rejected. The wall-clock win now rests entirely on M2+M3
reducing the NUMBER of rounds.

### 2026-08-20 — Core concepts table corrected (M1 boundary review, I6)

`closeReviewSnapshot` and `resolveReviewWindow` were listed under **Pure entities**.
Neither is: `validate()` does `os.ReadFile` plus git, and `resolveReviewWindow` shells
`rev-parse` directly. Moved to Integration points. The code satisfies ARCH-PURE — the
decision logic really is pure and really does unit-test with zero IO; the table was
wrong, and was wrong before M1 touched it.

### 2026-08-20 — D4's protocol-miss clause reversed (M2 implementation)

Originally *halt*; now *warn and persist*. Two facts from the code overturned it: the
fallback's failure mode (next round blind) is the pre-#194 status quo rather than a
regression, and the only escapes from a halt — `--no-judge`, `--force` — skip the review
entirely, so strictness would convert an occasional formatting miss into a routine reason
to run no review at all. Full argument in D4 itself.

The heading contradicted the body for one commit before M2's boundary review caught it
(I6) — a D-heading that says the opposite of its own text defeats the reason the headings
exist.

### 2026-08-20 — `--no-ledger` added (M2 boundary review, I4)

The ledger's open-findings refusal shipped with no per-gate bypass, leaving an operator
who hits "verdict SHIP, but the gate ledger still has open blocking finding(s)" with only
`--no-judge` (skip the review entirely) or `--force` (waive everything). AGENTS.md §5
makes a per-gate `--no-<gate>` flag a property of these commands, so the gate got one.

### 2026-08-20 — M3 review: the scoped/full split, and two rules (BR-20, BR-31, BR-32)

`RenderPriorFindings` was given the boundary-FILTERED ledger, so `FamilyCounts` could
never see a family recur across milestones — voiding the sole justification for one ledger
per issue, and rendering an empty vocabulary at the whole-issue close. Split into
`RenderPriorFindingsScoped(scoped, full)`: scoped supplies what must be disposed, full
supplies the families. New exported API, now listed in Core concepts.

Task 3.4 did **not** reuse `DispositionCounts` as the plan sketched — `ConvergenceLine`
needs per-round counts (new / repeat / disposed *this round*), while `DispositionCounts`
tallies the whole ledger by final state. Different questions; recorded rather than left as
an unexplained divergence from the plan.

Two rules adopted from this round, both written to `workshop/lessons.md` rather than
applied as one-off fixes:

1. **A fix is complete only when a test fails without it, verified by reverting.** Four
   fixes across M2 and M3 shipped with tests that passed either way.
2. **Grep `atlas/` for the mechanism you REMOVED.** Three instances across M1–M3.

The durable plan itself has no gate — `close`'s plan-unchecked check reads only the
ISSUE's `## Plan` — which is why this table drifted in the same commit that ticked its
boxes. Filed as its own issue rather than fixed here.

### 2026-08-20 — close review: two shared-gatestate behaviours, recorded (BR-18, BR-19, BR-37)

Three changes landed in `gatestate` that both gates now share, and none was in the
original design:

- **`ApplyChecked` rejects per-disposition, not per-round.** One typo'd id used to
  nullify a round's valid disposals — at the gate whose purpose is disposal.
- **`CountedRounds` / `Round.NoCap`.** The cap counts reviewer invocations, not persisted
  rounds. An interrupted review is deliberately NOT excluded; see `Round.NoCap`.
- **`DecideScoped` / `FilterBoundary` carrying BoundaryAll disposals.** The cap wants
  boundary scope; every other read wants the full issue. A seeded finding's *disposal*
  crosses boundaries with the finding, or it re-opens at every later boundary forever.
