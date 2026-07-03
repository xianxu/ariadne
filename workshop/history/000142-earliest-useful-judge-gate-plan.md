# Earliest-Useful Judge Gate — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move plan/specs acceptance-and-docs review to the earliest useful gate (`sdlc close`'s boundary review) and make `sdlc push`/`merge` re-judge only genuine post-close drift, so fixable gaps are caught before the close verdict is recorded — never as a redundant late pass after PR creation (pair#84).

**Architecture:** The `sdlc close` / `milestone-close` boundary review (`code-review.md`) already covers plan completeness (Requirements-traceability) and atlas sync (Atlas gate). We (1) strengthen that one review to also own **README** sync — the gap class pair#84 hit — so close is the earliest gate that fully owns plan+specs; and (2) replace push/merge's blanket `plan`+`specs`+`lessons` pass with a **post-close delta re-review**: re-judge only the window since the last `Review-Verdict:` commit, **skipping the LLM passes entirely when close already covered HEAD**. `lessons` (a no-LLM reminder ping) and the deterministic #124 conformance gate stay at merge. This satisfies both Done-when #3 (no redundant slow passes) and #4 (merge still catches drift after PR creation).

**Tech Stack:** Go (`cmd/sdlc`), embedded markdown prompts (`cmd/sdlc/internal/judge/prompts`, `code-review.md`), golden tests, `gitx` git wrappers.

> ⚠️ **STATUS: PENDING SEQUENCING (2026-07-02).** An operator reframing tied this
> issue to #160 (`codecomplete` status). Under the two-gate model, **merge should
> run no LLM judge** and #160's status invariant replaces the post-close *delta-
> review* below (Tasks 2, 3, 6) — making that machinery potentially throwaway. What
> survives in **every** scenario is **Task 4 (strengthen close's boundary review to
> own README docs sync)** — the no-regret pair#84 fix. Do NOT execute Tasks 2/3/6
> until the #142×#160 sequencing is decided (see issue Log 2026-07-02 + `## Revisions`).

**Architectural principles applied:**
- **ARCH-DRY** — the merge-time `plan`/`specs` LLM passes duplicate the close boundary review's coverage; we collapse to one review system + one shared `runPreMergeJudges` (merge and push had identical `preflightOptions` blocks). Reuse `previousReviewBoundary`'s git-grep by generalizing it to `latestVerdictCommit`.
- **ARCH-PURE** — the delta-window decision (the load-bearing branching) is a pure function (`preMergeWindowDecision`) unit-tested without git; the git gathering is a thin glue seam.
- **ARCH-PURPOSE** — the issue's purpose is "catch fixable acceptance/doc gaps before close records a final verdict." Strengthening the close review's README coverage is that purpose; merely reordering the old judges would leave the pair#84 root cause (thin README coverage) unfixed.

---

## Audit (Done-when #1 — categories × current gate)

| Judge | Runs today | LLM pass? | Checks | Overlaps close boundary review? |
|-------|-----------|-----------|--------|---------------------------------|
| `plan` | push + merge | yes | Issue Plan completeness: done-but-unchecked items, Log entries, `status:` | **Yes** — "Requirements traceability" section; frontmatter also covered by deterministic #124 gate |
| `specs` | push + merge | yes | `atlas/` + README.md sync vs diff | **Partly** — "Atlas update gate" covers atlas well but README thinly (pair#84 root cause) |
| `lessons` | push + merge | no (reminder ping) | Prints "review lessons.md" | No — session reflection, not acceptance criteria |

**Target gates after this change (Done-when #2):**

| Judge | New gate | Rationale |
|-------|----------|-----------|
| `plan` | close (boundary review, full window) + merge/push (post-close delta only) | Close owns acceptance; merge re-checks only drift added after close |
| `specs` | close (boundary review, full window, **now incl. README**) + merge/push (post-close delta only) | Same; README gap now caught at earliest gate |
| `lessons` | merge/push (unchanged) | Whole-session pre-ship reflection; cheap, no redundancy |

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `preMergeWindowDecision` | `cmd/sdlc/preflight.go` | new |

- **preMergeWindowDecision** — the pure decision at the heart of the pre-merge gate: given the last `Review-Verdict:` SHA, the fallback diff-base, and how many commits sit after the verdict, return `(base, skip)`. No git, no IO.
  - **Relationships:** 1:1 with `preMergeReviewWindow` (the glue that feeds it live git values).
  - **DRY rationale:** Isolates the three-way branching (no verdict → whole window; verdict==HEAD → skip; verdict<HEAD → delta) so it is unit-tested exhaustively without a temp repo. First occurrence of "verdict-relative window decision" as pure logic.
  - **Future extensions:** Could take a per-issue verdict map if merge ever wants per-issue delta windows; today the branch-level last verdict is enough.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `latestVerdictCommit` | `cmd/sdlc/milestoneclose.go` | modified | `git log --grep=Review-Verdict:` |
| `preMergeReviewWindow` | `cmd/sdlc/preflight.go` | new | git (`log`, `rev-list`, `DiffBase`) |
| `runPreMergeJudges` | `cmd/sdlc/preflight.go` | new | judge dispatch (via `runPreflightJudgesFn`) |
| `code-review.md` "Docs update gate" | `cmd/sdlc/internal/judge/code-review.md` | modified | boundary-review prompt text |

- **latestVerdictCommit** — generalizes the existing `previousReviewBoundary(issuePath)`: `latestVerdictCommit(paths ...string)` returns the most recent commit reachable from HEAD carrying a `Review-Verdict:` trailer; with no paths it is **branch-level** (the pre-merge base), with a path it keeps the milestone-scoped behavior. `previousReviewBoundary` becomes a one-line delegator (keeps its empty-path guard).
  - **Injected into:** `preMergeReviewWindow` (branch-level) and `boundaryWindowBase` (path-scoped, unchanged behavior).
  - **DRY rationale:** One git-grep for "last review boundary," two scopings. Avoids a second near-identical `git log --grep` call site.
- **preMergeReviewWindow** — thin glue: finds the last **finalizing** `Review-Verdict:` commit (`latestVerdictCommit()` + `judge.ParseVerdictTrailer` on its message, gated by `vocab.Verdict().IsFinalizing`), gathers `gitx.DiffBase()` and `git rev-list --count <verdict>..HEAD`, then returns `preMergeWindowDecision(...)`. A **non-finalizing** boundary (a `--no-judge`/`--force` close emits `Review-Verdict: not-run`, which `ParseVerdictTrailer` maps to `VerdictUnknown`) establishes NO coverage → treated as "no verdict" → whole-window fallback. This is the fix for the review's Issue #1: without it, a bypassed *close* review would silently disable the *merge* plan/specs pass (`not-run` at HEAD → skip). `close.go:847` already uses `vocab.Verdict().IsFinalizing` — mirror that exact pattern (vocab is already imported in package main).
  - **Injected into:** `runPreMergeJudges`.
  - **Edge case (push-on-main):** on `main`, `latestVerdictCommit()` may return a *prior merged issue's* close commit — correct: base..HEAD is then "everything since the last review boundary," which is exactly this issue's work. If the merge strategy squashed trailers away, `latestVerdictCommit()` returns "" → whole-window fallback (`DiffBase()` = `origin/main`). Both directions are safe (over-cover, never under-cover).
- **runPreMergeJudges** — shared pre-merge gate for `sdlc push` and `sdlc merge` (replaces the duplicated `preflightOptions` blocks in both). Resolves the delta window; runs `[plan, specs, lessons]` on the delta, or `[lessons]` only when the boundary review already covered HEAD. Calls the existing `runPreflightJudgesFn` seam so current e2e stubs still intercept.
  - **Injected into:** `merge.go` step 5, `push.go` step 4.
- **code-review.md "Docs update gate"** — the boundary-review prompt section renamed from "Atlas update gate" to "Docs update gate (atlas + README)" with an explicit README bullet. Golden-tested; the golden testdata must be regenerated.

**Milestone structure:** single atomic boundary — the close-review strengthening and the merge/push rewiring are one coupled change reviewed together. Plain checkboxes, no `Mx` tags (AGENTS.md §3: tag `Mx` only for ≥2 separately-closed boundaries). One `sdlc close` → one boundary review → one log line.

---

## Task 1: Generalize `latestVerdictCommit` (ARCH-DRY)

**Files:**
- Modify: `cmd/sdlc/milestoneclose.go:264-273` (`previousReviewBoundary`)
- Test: `cmd/sdlc/milestonewindow_test.go` (existing tests for `previousReviewBoundary` must still pass)

- [ ] **Step 1: Read the existing test coverage**

Run: `go test ./cmd/sdlc/ -run 'BoundaryWindowBase|MilestoneWindow' -v`
Expected: PASS (baseline — these must keep passing after the refactor). Note: there is no test literally named `*PreviousReviewBoundary*`; `previousReviewBoundary` is exercised *inside* the `TestBoundaryWindowBase_*` tests (`milestonewindow_test.go`).

- [ ] **Step 2: Add `latestVerdictCommit` and delegate `previousReviewBoundary`**

Replace the body of `previousReviewBoundary` and add the generalized helper above it in `milestoneclose.go`:

```go
// latestVerdictCommit returns the SHA of the most recent commit reachable from
// HEAD whose message carries a Review-Verdict: trailer — the last review boundary.
// With no paths it is branch-level (the last close/milestone-close anywhere on the
// branch; the pre-merge delta base, #142). With a path it scopes to commits
// touching that file (the milestone window base). "" when none / on git error.
func latestVerdictCommit(paths ...string) string {
	args := []string{"log", "--grep=Review-Verdict:", "--max-count=1", "--pretty=format:%H"}
	if len(paths) > 0 {
		args = append(args, "--")
		args = append(args, paths...)
	}
	out, err := gitx.RunGit(args...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
```

Rewrite `previousReviewBoundary` to delegate (keep its doc comment + empty guard):

```go
func previousReviewBoundary(issuePath string) string {
	if issuePath == "" {
		return ""
	}
	return latestVerdictCommit(issuePath)
}
```

- [ ] **Step 3: Verify the refactor is behavior-preserving**

Run: `go test ./cmd/sdlc/ -run 'PreviousReviewBoundary|BoundaryWindowBase|MilestoneWindow' -v`
Expected: PASS (no behavior change for the path-scoped caller).

- [ ] **Step 4: Commit**

```bash
git add cmd/sdlc/milestoneclose.go
git commit -m "#142: generalize previousReviewBoundary → latestVerdictCommit (ARCH-DRY)"
```

---

## Task 2: Pure delta-window decision (ARCH-PURE)

**Files:**
- Modify: `cmd/sdlc/preflight.go` (add `preMergeWindowDecision` + `preMergeReviewWindow`; add `strconv` import)
- Test: `cmd/sdlc/preflight_test.go`

- [ ] **Step 1: Write the failing pure-decision test**

Add to `preflight_test.go`:

```go
func TestPreMergeWindowDecision(t *testing.T) {
	cases := []struct {
		name          string
		verdictSHA    string
		diffBase      string
		commitsAfter  int
		wantBase      string
		wantSkip      bool
	}{
		{"no verdict on branch → whole window", "", "origin/main", 0, "origin/main", false},
		{"HEAD == verdict → skip", "abc123", "origin/main", 0, "abc123", true},
		{"commits after verdict → delta", "abc123", "origin/main", 2, "abc123", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			base, skip := preMergeWindowDecision(c.verdictSHA, c.diffBase, c.commitsAfter)
			if base != c.wantBase || skip != c.wantSkip {
				t.Fatalf("got (%q,%v), want (%q,%v)", base, skip, c.wantBase, c.wantSkip)
			}
		})
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./cmd/sdlc/ -run TestPreMergeWindowDecision -v`
Expected: FAIL — `undefined: preMergeWindowDecision`.

- [ ] **Step 3: Implement the pure decision + glue**

Add to `preflight.go` (import `strconv`):

```go
// preMergeWindowDecision is the PURE core of the pre-merge gate (#142): given the
// last FINALIZING Review-Verdict SHA (or "" when there is no reliable boundary), a
// fallback diff base, and the count of commits after the verdict, decide what to
// review and whether to skip. No git, no IO.
//
//   verdictSHA == ""      → review the whole window (diffBase); the safe
//                           over-cover for work with no finalizing boundary
//                           (a --no-judge/--force close's "not-run", or a bare push).
//   commitsAfter == 0     → skip: the close boundary review already covered HEAD.
//   commitsAfter  > 0     → review verdictSHA..HEAD (post-close drift only).
func preMergeWindowDecision(verdictSHA, diffBase string, commitsAfter int) (base string, skip bool) {
	if verdictSHA == "" {
		return diffBase, false
	}
	if commitsAfter == 0 {
		return verdictSHA, true
	}
	return verdictSHA, false
}

// preMergeReviewWindow is the thin git seam that feeds preMergeWindowDecision the
// live values: the branch-level last FINALIZING verdict commit, the diff base, and
// the count of commits after it. A non-finalizing boundary (e.g. a --no-judge
// close's "not-run") establishes no coverage, so it degrades to the whole-window
// fallback rather than silently disabling the plan/specs pass (#142, review Issue #1).
func preMergeReviewWindow() (base string, skip bool) {
	v := latestVerdictCommit()
	if v != "" {
		msg := gitx.Capture("log", "-1", "--pretty=format:%B", v)
		if !vocab.Verdict().IsFinalizing(string(judge.ParseVerdictTrailer(msg))) {
			v = "" // non-finalizing (not-run/unknown) → no coverage; over-cover the whole window
		}
	}
	if v == "" {
		return preMergeWindowDecision("", gitx.DiffBase(), 0)
	}
	n, _ := strconv.Atoi(strings.TrimSpace(gitx.Capture("rev-list", "--count", v+"..HEAD")))
	return preMergeWindowDecision(v, gitx.DiffBase(), n)
}
```

Imports needed in `preflight.go`: add `strconv`, `github.com/xianxu/ariadne/pkg/vocab` (`judge` and `gitx` already imported).

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./cmd/sdlc/ -run TestPreMergeWindowDecision -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/preflight.go cmd/sdlc/preflight_test.go
git commit -m "#142: pure preMergeWindowDecision + git-seam preMergeReviewWindow (ARCH-PURE)"
```

---

## Task 3: Shared `runPreMergeJudges` + wire merge/push (ARCH-DRY)

**Files:**
- Modify: `cmd/sdlc/preflight.go` (add `runPreMergeJudges`)
- Modify: `cmd/sdlc/merge.go:312-329` (step 5)
- Modify: `cmd/sdlc/push.go:132-147` (step 4)
- Test: `cmd/sdlc/preflight_test.go` (category-selection test), `cmd/sdlc/merge_e2e_test.go` (adjust expectations if needed)

- [ ] **Step 1: Write the failing category-selection test**

Add to `preflight_test.go` a test that stubs `runPreflightJudgesFn`, captures the categories it receives, and asserts the delta/skip behavior. Because `runPreMergeJudges` calls `preMergeReviewWindow` (real git), drive category selection through a small seam: have `runPreMergeJudges` take `(base string, skip bool)` from `preMergeReviewWindow` internally, but expose selection via the captured `preflightOptions`. Test both branches by stubbing at the `runPreflightJudgesFn` seam and controlling `skip` through a temp repo OR by testing the pure selection helper. Prefer a pure selection helper:

```go
func TestPreMergeCategories(t *testing.T) {
	if got := preMergeCategories(false); len(got) != 3 {
		t.Fatalf("delta window: want [plan specs lessons], got %v", got)
	}
	if got := preMergeCategories(true); len(got) != 1 || got[0] != judge.Lessons {
		t.Fatalf("skip: want [lessons], got %v", got)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./cmd/sdlc/ -run TestPreMergeCategories -v`
Expected: FAIL — `undefined: preMergeCategories`.

- [ ] **Step 3: Implement `preMergeCategories` + `runPreMergeJudges`**

Add to `preflight.go`:

```go
// preMergeCategories is the PURE category selection: on a genuine post-close delta
// run the full plan+specs+lessons set; when the boundary review already covered
// HEAD (skip), run only the no-LLM lessons reminder (#142 Done-when #3 — no
// redundant slow passes).
func preMergeCategories(skip bool) []judge.Category {
	if skip {
		return []judge.Category{judge.Lessons}
	}
	return []judge.Category{judge.Plan, judge.Specs, judge.Lessons}
}

// runPreMergeJudges is the shared pre-merge gate for `sdlc push` and `sdlc merge`
// (ARCH-DRY: both had an identical preflightOptions block). It re-judges only the
// post-close delta — the window since the last Review-Verdict commit — and skips
// the plan/specs LLM passes when the close boundary review already covered HEAD
// (#142). It routes through runPreflightJudgesFn so existing e2e stubs intercept.
func runPreMergeJudges(stdout, stderr io.Writer, issuesDir, historyDir string, dryRun bool) error {
	base, skip := preMergeReviewWindow()
	if skip {
		cinfo(stderr, "boundary review already covered HEAD (no commits since the last "+
			"Review-Verdict) — skipping plan/specs re-check; lessons reminder only")
	} else if base != "" {
		cinfo(stderr, fmt.Sprintf("pre-merge judges scoped to post-close delta %s..HEAD", shortSHA(base)))
	}
	return runPreflightJudgesFn(preflightOptions{
		Categories: preMergeCategories(skip),
		Base:       base,
		IssuesDir:  issuesDir,
		HistoryDir: historyDir,
		DryRun:     dryRun,
		Stdout:     stdout,
		Stderr:     stderr,
	})
}
```

Note: `runPreflightJudgesFn` currently lives in `merge.go:77` (`var runPreflightJudgesFn = runPreflightJudges`). Keep it there (or move to `preflight.go`); `runPreMergeJudges` references it.

- [ ] **Step 4: Rewire `merge.go` step 5**

Replace the `preOpts := preflightOptions{…}; runPreflightJudgesFn(preOpts)` block (merge.go:313-326) with:

```go
if err := runPreMergeJudges(stdout, stderr, f.IssuesDir, f.HistoryDir, f.DryRun); err != nil {
	die(stderr, fmt.Sprintf("pre-merge judges failed: %v\n"+
		"  → fix the finding, commit, `git push`, then re-run `sdlc merge` "+
		"(the fix must reach origin — merge is server-side).", err))
}
```

- [ ] **Step 5: Rewire `push.go` step 4**

Replace the `preOpts := preflightOptions{…}; runPreflightJudges(preOpts)` block (push.go:134-144) with:

```go
if err := runPreMergeJudges(stdout, stderr, f.IssuesDir, f.HistoryDir, f.DryRun); err != nil {
	die(stderr, fmt.Sprintf("pre-merge judges failed: %v", err))
}
```

- [ ] **Step 6: Run tests; adjust merge_e2e_test expectations if the fixture trips skip**

Run: `go test ./cmd/sdlc/ -run 'PreMerge|Merge|Push|Preflight' -v`
Expected: PASS. If `merge_e2e_test.go`'s fixture has a `Review-Verdict:` commit at HEAD, `runPreMergeJudges` now selects `[lessons]` only — update the test's asserted categories or add a post-verdict commit to the fixture so the delta path is exercised. Confirm the stub-swap on `runPreflightJudgesFn` still intercepts (it does — `runPreMergeJudges` calls it).

- [ ] **Step 7: Commit**

```bash
git add cmd/sdlc/preflight.go cmd/sdlc/merge.go cmd/sdlc/push.go cmd/sdlc/preflight_test.go cmd/sdlc/merge_e2e_test.go
git commit -m "#142: shared runPreMergeJudges re-judges only the post-close delta"
```

---

## Task 4: Strengthen the close boundary review — README docs gate (ARCH-PURPOSE)

**Files:**
- Modify: `cmd/sdlc/internal/judge/code-review.md` (line 55 cross-reference **and** the line 73 "Atlas update gate" section)
- Modify: `cmd/sdlc/internal/judge/judge_test.go:376` (`TestBuildPrompt_MilestoneReview_HasContract` asserts the literal `"Atlas update gate"`)
- Test: `cmd/sdlc/internal/judge/golden_test.go` + regenerate golden testdata

> **Review Issue #2:** the phrase "Atlas update gate" appears **twice** in `code-review.md` — the section header (line 73) *and* a cross-reference at line 55 (`"Docs / atlas updated for new surface (see the Atlas update gate)"`). And `judge_test.go:376` asserts that literal string is present in the rendered prompt. All three must move together, or the test passes only by the stale leftover phrase.

- [ ] **Step 1: Run the golden + contract tests to see the baseline pins**

Run: `go test ./cmd/sdlc/internal/judge/ -run 'TestBuildPrompt_Golden|HasContract' -v`
Expected: PASS (baseline).

- [ ] **Step 2a: Fix the cross-reference at `code-review.md:55`**

Change `- Docs / atlas updated for new surface (see the Atlas update gate).` → `- Docs / atlas updated for new surface (see the Docs update gate).`

- [ ] **Step 2b: Update the contract assertion `judge_test.go:376`**

Change the `"Atlas update gate",` entry in `TestBuildPrompt_MilestoneReview_HasContract` to `"Docs update gate",`.

- [ ] **Step 2: Edit the gate section in `code-review.md`**

Replace the `## Atlas update gate (per AGENTS.md §8)` section (line 73) with:

```markdown
## Docs update gate (atlas + README, per AGENTS.md §8)

The boundary should update user-facing docs for any new surface introduced:

  - **atlas/** — new architectural surface, flow, or terminology. Scan the diff
    for new entity types, subcommands, conventions, file-tree locations. Any
    present without corresponding atlas/ changes in the same range = Important
    finding ("atlas update appears missing for <surface>").
  - **README.md** — new user-facing surface a reader runs or types: subcommands,
    flags, keybindings, config keys, install/usage steps. If the diff adds or
    changes such surface and README.md is not updated in the same range =
    Important finding ("README update appears missing for <surface>"). This is the
    class of gap that used to surface only at the merge-time `specs` judge (#142);
    catch it here, at the earliest gate, before the close verdict is recorded.
```

- [ ] **Step 3: Regenerate the golden testdata**

Run: `go test ./cmd/sdlc/internal/judge -run TestBuildPrompt_Golden -update-golden`
(The flag is `-update-golden`, defined at `golden_test.go:11`; the test is `TestBuildPrompt_Golden`.)
Then inspect: `git diff cmd/sdlc/internal/judge/testdata` — confirm ONLY the docs-gate wording changed.

> Note: this regen is legitimate — it captures an **intentional** prompt-content change. It is NOT the `golden_test.go:37` ⛔ ("never re-run `-update-golden` to *fix* a failure"), which forbids papering over *accidental* refactor drift. Say so in the commit body so a reviewer doesn't cite the ⛔ against it.

- [ ] **Step 4: Run the golden + contract tests to verify they pass**

Run: `go test ./cmd/sdlc/internal/judge/ -run 'TestBuildPrompt_Golden|HasContract' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/internal/judge/code-review.md cmd/sdlc/internal/judge/judge_test.go cmd/sdlc/internal/judge/testdata
git commit -m "#142: close boundary review owns README docs sync, not just atlas (ARCH-PURPOSE)"
```

---

## Task 5: Help text + atlas (Done-when #5)

**Files:**
- Modify: `cmd/sdlc/helptext/push.md` (lines ~31-32, ~48)
- Modify: `cmd/sdlc/helptext/merge.md` (line ~24, ~73)
- Modify: `cmd/sdlc/helptext/close.md` (boundary-review section — note it now owns plan+specs incl. README)
- Modify: `atlas/workflow/pre-merge-checks.md` (gate map + delta-window behavior)
- Test: `cmd/sdlc/helptext_render_test.go` / `estimate_helptext_test.go` if they pin help content

- [ ] **Step 1: Update push.md / merge.md**

Replace "Runs pre-merge judges: `sdlc judge plan`, `specs`, `lessons`" with a description of the new behavior, e.g.:

```
  2. Runs the pre-merge judges on the POST-CLOSE DELTA only — the window since
     the last boundary review (`Review-Verdict:` commit). plan + specs re-check
     any commits added after `sdlc close`; when nothing landed since the close
     review they are SKIPPED (the close boundary review already covered HEAD).
     `lessons` (a reminder ping) always runs. Any Failure aborts. --no-judge
     skips (emergency only). Full acceptance/docs review lives at `sdlc close`.
```

- [ ] **Step 2: Update close.md**

Add a line to the boundary-review section noting it now owns plan-completeness AND docs sync (atlas + README) — the earliest gate for those, so merge only re-checks post-close drift.

- [ ] **Step 3: Update atlas/workflow/pre-merge-checks.md**

Update the "Checks" table + add a section documenting: which judge runs at which gate (close vs push/merge), the post-close-delta rule, and that `plan`/`specs` acceptance+docs coverage is owned by the close boundary review. Note the atlas doc's stale `make check`/scripts references should point at `sdlc` (fix opportunistically or note it).

- [ ] **Step 4: Verify help renders + any help pins pass**

Run: `go build ./cmd/sdlc && ./sdlc merge --help && ./sdlc push --help && ./sdlc close --help`
Run: `go test ./cmd/sdlc/ -run 'Help|Helptext' -v`
Expected: help text shows the new wording; tests PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/helptext/ atlas/workflow/pre-merge-checks.md
git commit -m "#142: document per-gate judge placement + post-close-delta rule"
```

---

## Task 6: Regression tests (Done-when #6)

**Files:**
- Test: `cmd/sdlc/preflight_test.go` (delta window integration) — and/or `cmd/sdlc/merge_test.go`

- [ ] **Step 1: Integration test for `preMergeReviewWindow` over a real temp repo**

Follow the existing temp-repo test pattern (see `milestonewindow_test.go` for how it seeds commits). Assert four scenarios:
  - a) no `Review-Verdict:` commit → `skip == false`, base == `gitx.DiffBase()`;
  - b) a commit with `Review-Verdict: SHIP` in its message at HEAD → `skip == true`, base == that SHA;
  - c) an extra commit after the finalizing-verdict commit → `skip == false`, base == the verdict SHA;
  - d) **(review Issue #1)** a commit with `Review-Verdict: not-run` at HEAD → `skip == false`, base == `gitx.DiffBase()` (non-finalizing boundary degrades to whole-window; the merge gate is NOT silently disabled).

- [ ] **Step 2: Test that a specs FAILURE on the delta blocks (the pair#84 shape)**

Stub `runPreflightJudgesFn` to return an error for the `specs` category; call `runPreMergeJudges` in a repo where a post-close delta exists (scenario c); assert it returns non-nil (merge/push would `die`). Then a scenario-b (skip) repo: assert `runPreflightJudgesFn` was invoked with `[lessons]` only and returns nil.

- [ ] **Step 3: Run the full sdlc test suite**

Run: `go test ./cmd/sdlc/... ./cmd/sdlc/internal/judge/...`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add cmd/sdlc/preflight_test.go cmd/sdlc/merge_test.go
git commit -m "#142: regression tests — post-close delta skip + drift caught before merge"
```

---

## Verification (whole-issue, before `sdlc close`)

- [ ] `go test ./cmd/sdlc/... ./cmd/sdlc/internal/judge/...` — all green.
- [ ] `go build ./cmd/sdlc` — builds.
- [ ] Manual behavior check vs the pair#84 story:
  - Close an issue with a README gap → the close boundary review's Docs gate now flags README (Important) BEFORE the verdict is recorded. (Dry-run: `./sdlc close --issue N --dry-run` and read the emitted boundary-review prompt to confirm the README gate text is present.)
  - After a clean close with a pasted `Review-Verdict: SHIP` trailer at HEAD → `./sdlc merge --dry-run` logs "boundary review already covered HEAD … skipping plan/specs" (skip path). **Caveat (review Issue #5):** `sdlc merge` `die`s at its step 3/4 (branch-not-on-origin / commits-not-pushed, `merge.go:285-298`) *before* the step-5 judges — so this observation requires the branch pushed and HEAD-synced to origin. On a local-only branch, exercise the path via `./sdlc push --dry-run` instead, or the Task 6 integration test.
  - Add a commit after close → the pre-merge gate logs "scoped to post-close delta <sha>..HEAD" and runs plan+specs on it (drift path).
- [ ] Confirm `--no-judge` still bypasses the whole pre-merge gate (emergency escape intact), and that a `--no-judge` *close* (`Review-Verdict: not-run`) followed by a plain `sdlc merge` still runs plan+specs on the whole window (Issue #1 — gate not disabled).

## Done-when cross-check

- [x] #1 categories × current gate documented (Audit table above; also in atlas).
- [x] #2 explicit target gate per category (Audit "New gate" table).
- [x] #3 no redundant slow passes — merge/push skip plan/specs entirely when close covered HEAD (`preMergeWindowDecision` skip branch).
- [x] #4 merge still protects post-PR drift — delta window re-judges commits added after close.
- [x] #5 help text explains which judge runs where (Task 5).
- [x] #6 tests: moved-judge gate + a specs failure caught before merge (Task 6).

---

## Revisions

### 2026-07-02 — reframed against #160 (codecomplete); delta-review may be superseded

**Reason:** Operator clarified the architecture: merge's LLM judges are **local/
client-side**; the only **server-side** merge gate is the deterministic CI merge-
check (`ci-merge-check.md`), not an LLM. With #160 adding a `codecomplete` status,
the intended model is: `sdlc close` (local) = LLM acceptance review → `codecomplete`;
`sdlc merge` (remote) = deterministic CI + mechanical `codecomplete → done` + merge
+ push, **with no LLM judge**.

**Delta:**
- The post-close **delta-review** (Tasks 2, 3, 6 — `preMergeWindowDecision`,
  `preMergeReviewWindow`, `runPreMergeJudges`) was designed to catch commits landing
  *after* close. #160's status invariant (*codecomplete ⟹ close reviewed HEAD*)
  handles that structurally, so this machinery is likely **superseded** — do not
  build it until the #142×#160 sequencing is chosen.
- **Task 4 (README docs gate in the close boundary review) stands unchanged** — it
  is the no-regret pair#84 fix, required under every sequencing option.
- If sequencing = "fold into #160," this plan is absorbed into #160's plan and the
  merge side becomes "delete `plan`+`specs` from preflight; merge runs deterministic
  CI + `codecomplete→done`," with no delta-review.

**Not yet applied to the task bodies** (pending the sequencing decision) — the Tasks
below still describe the delta-review approach; treat Task 4 as independently shippable.
