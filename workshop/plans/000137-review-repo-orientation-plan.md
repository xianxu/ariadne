# Boundary-Review Repo Orientation Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every boundary-review prompt (`sdlc close` / `milestone-close`) orient the fresh reviewer to the *actual* repo under review — derived from the live git context — instead of a hardcoded `ariadne#N` reference, so a review running in `pair` is told `pair#72` (not `ariadne#72`) with concrete repo/issue anchors.

**Architecture:** The repo derivation is IO (git root), so it lives in `cmd/sdlc` (a new `boundaryOrientation` helper over `gitx.RepoTopLevel` + the existing `issueFilePath`), and the result is passed as plain strings through `judge.PromptInput` into the prompt — `internal/judge` stays pure (ARCH-PURE), exactly as it does today for Base/Head/IssueRef. The hardcoded `"ariadne#N"` at the two call sites is replaced by the derived `<repo>#N` ref computed once in the shared `boundaryReviewDispatchOptions` (ARCH-DRY — one derivation, both close + milestone-close). `code-review.md`'s orientation header gains placeholders the existing `CodeReviewBody` substitutes — same templating seam, more anchors.

**Tech Stack:** Go; pure prompt-render unit tests in `internal/judge`; cmd-level orientation tests with a temp git repo fixture; the embedded `code-review.md` template; atlas docs.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `PromptInput` (orientation fields) | `cmd/sdlc/internal/judge/prompts.go:109` | modified |
| `CodeReviewBody` | `cmd/sdlc/internal/judge/review.go:29` | modified |
| `code-review.md` (orientation header) | `cmd/sdlc/internal/judge/code-review.md` | modified |
| `reviewOrientation` | `cmd/sdlc/orientation.go` | new |
| `renderReviewEntry` (H1) | `cmd/sdlc/reviewsidecar.go:71` | modified |

- **PromptInput** *(modified)* — gains `Repo, RepoRoot, IssueFile, Boundary, RepoNote string`. Callers populate them for the MilestoneReview category; other categories ignore them (the existing pattern). `IssueRef` stays but is now built with the repo prefix.
  - **DRY rationale:** one struct already carries all per-prompt data; orientation joins it rather than a parallel channel.
- **CodeReviewBody** *(modified)* — signature changes from `(issueRef, base, head string)` to `(in PromptInput)`; it substitutes the new `{{REPO}}`, `{{REPO_ROOT}}`, `{{ISSUE_FILE}}`, `{{BOUNDARY}}`, `{{REPO_NOTE}}` placeholders alongside the existing `{{ISSUE_REF}}/{{BASE}}/{{HEAD}}/{{ARCH_STAR}}`. Pure string templating.
  - **Relationships:** sole caller is `BuildPrompt` (prompts.go:354) + one test (judge_test.go:152).
- **code-review.md** *(modified)* — header (lines 3-6) replaced with an explicit "Repository under review" orientation block naming the repo, root, issue ref + file, boundary, and window, plus a base-vs-downstream note. The reviewer reads this first.
- **reviewOrientation** *(new)* — `boundaryOrientation(issuesDir string, issueNum int, milestone string) reviewOrientation` returns `{Repo, RepoRoot, IssueRef, IssueFile, Boundary, RepoNote}`, derived from `gitx.RepoTopLevel`, `issueFilePath`, and a `construct/base.manifest` existence check (base repo vs downstream). The single source of orientation truth.
  - **Relationships:** 1 per boundary review; consumed by `boundaryReviewDispatchOptions`.
  - **DRY rationale:** consolidates the repo-name derivation that `repoIdentity` (#136) also does — both route through a new `repoNameAndRoot()` (addresses the #136 review's triplication note).
  - **Future extensions:** could add remote URL / default-branch if a reviewer ever needs them.
- **renderReviewEntry (H1)** *(modified, #137 shadow-sweep)* — the persisted boundary-review sidecar's H1 currently hardcodes `# Boundary Review — ariadne#%d` (`reviewsidecar.go:71`, introduced #136), so a `pair` review's sidecar title reads `ariadne#72` while its `| repo |` cell reads `pair` — the *exact* misorientation #137 kills, persisted into the durable artifact. Fix: render the H1 from `m.Repo` (already populated by `repoIdentity()`): `# Boundary Review — <repo>#<num> (<boundary>)`. Empty-repo guard → `<unknown-repo>`.

> **Plan-quality gate revision (see `## Revisions`):** the sidecar H1 (above) and the full test-update surface (Task 3) were added after the change-code plan-quality judge flagged them.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `boundaryOrientation` | `cmd/sdlc/orientation.go` (new) | new | git root + fs |
| `repoNameAndRoot` | `cmd/sdlc/reviewsidecar.go` | new | `gitx.RepoTopLevel` |
| `boundaryReviewDispatchOptions` | `cmd/sdlc/milestoneclose.go:512` | modified | PromptInput build |
| close/milestone call sites | `close.go` / `milestoneclose.go` | modified | drop hardcoded ref |

- **boundaryOrientation** — the IO seam: resolves the repo + issue anchors. Injected once into `boundaryReviewDispatchOptions`, which sets the orientation fields + `IssueRef` on `PromptInput`. Returns best-effort empties (`<unknown-repo>`) if the git root can't be resolved — never blocks the review.
- **repoNameAndRoot** — `() (name, root string)`: `root, _ = gitx.RepoTopLevel(); name = base(root)` with empty-string guards (avoids `filepath.Base("") == "."`). `repoIdentity` (#136) is refactored to call it.
- **boundaryReviewDispatchOptions** *(modified)* — computes `boundaryOrientation(p.IssuesDir, p.IssueNum, p.Milestone)` and populates `PromptInput.{IssueRef,Repo,RepoRoot,IssueFile,Boundary,RepoNote}` from it, replacing `IssueRef: p.IssueRef`.
- **call sites** *(modified)* — `close.go` and `milestoneclose.go` stop setting `IssueRef: "ariadne#..."` on `boundaryReviewParams` (it's now derived); the `IssueRef` field is removed from `boundaryReviewParams` (Label stays for stderr messages). `p.IssueNum`/`p.Milestone` (added #136) feed the derivation.

**Test surface.** `CodeReviewBody`/`BuildPrompt` rendered output asserted purely (ariadne ref + a downstream `pair#72` ref + every anchor present, no `{{` left). `boundaryOrientation` tested against a `t.TempDir()` git repo fixture (named `pair`, with + without `construct/base.manifest`) → asserts `pair#72` and the downstream note, and that it never emits `ariadne#` for a non-ariadne root. Existing verdict/trailer/gate tests must stay green.

---

## Design decisions

- **D1 — derive, never hardcode.** Repo name = `filepath.Base(gitx.RepoTopLevel())` (the git root), matching the issue's "from git root/remote/cwd context." `IssueRef` becomes `<repo>#<num>[ <milestone>]`.
- **D2 — orientation computed in cmd/sdlc, rendered in judge (ARCH-PURE).** Git IO stays at the cmd boundary; `internal/judge` receives plain strings and stays unit-testable with no git.
- **D3 — single derivation point (ARCH-DRY).** Both close and milestone-close already funnel through `boundaryReviewDispatchOptions`; the orientation (incl. the ref) is computed there once, not at each call site. The hardcoded-`ariadne` strings are deleted, not duplicated-with-a-variable.
- **D4 — base-vs-downstream signal.** A repo is the base iff `construct/base.manifest` exists at its root (ariadne has it; downstream repos carry `construct/` but not the manifest). The note tells the reviewer to apply *this* repo's conventions — ariadne is named only when it IS the reviewed repo.
- **D5 — additive to the contract.** Verdict format, trailers, gates, the architecture/contract blocks, and the diff are unchanged; only the orientation header gains anchors. Existing tests that pass `IssueRef` still see it rendered.

---

## Chunk 1: Pure prompt orientation (judge)

### Task 1: PromptInput fields + CodeReviewBody + template

**Files:** Modify `internal/judge/prompts.go` (PromptInput, BuildPrompt:354), `internal/judge/review.go` (CodeReviewBody), `internal/judge/code-review.md`; Test `internal/judge/judge_test.go:152`

- [ ] **Step 1: Update the CodeReviewBody test (red).** Change the call to the new signature and assert the new anchors:

```go
body := CodeReviewBody(PromptInput{
    IssueRef: "pair#72 M1", Base: "BASE_SHA", Head: "HEAD_SHA",
    Repo: "pair", RepoRoot: "/w/pair", IssueFile: "workshop/issues/000072-x.md",
    Boundary: "milestone M1 close", RepoNote: "a downstream repo built on the ariadne base layer",
})
for _, want := range []string{"pair#72 M1", "Base: BASE_SHA", "Head: HEAD_SHA",
    "pair", "/w/pair", "workshop/issues/000072-x.md", "milestone M1 close",
    "downstream repo", "ARCH-DRY, ARCH-PURE, ARCH-PURPOSE", "Core concepts cross-check"} { ... }
// plus the existing no-"{{" check.
```

- [ ] **Step 2: Run → fails to compile** (CodeReviewBody signature).
- [ ] **Step 3: Implement.** Add the 5 fields to `PromptInput`. Change `CodeReviewBody(in PromptInput)` to build the replacer with the new placeholders (empty-string-safe). Edit `code-review.md` header (lines 3-6) to an orientation block:

```
You are conducting a fresh-context code review at a development boundary —
{{BOUNDARY}} — in the **{{REPO}}** repository.

- repository: {{REPO}}   (root: {{REPO_ROOT}})
- issue:      {{ISSUE_REF}}   (file: {{ISSUE_FILE}})
- window:     {{BASE}}..{{HEAD}}

Review the **{{REPO}}** repo and its tracker — {{REPO_NOTE}}. Do not assume any
other repository.
```

  Update `BuildPrompt`'s MilestoneReview case to call `CodeReviewBody(in)` (drop the 3-arg form). Keep `Base:`/`Head:` lines so the existing substring tests still pass.
- [ ] **Step 4: Run → PASS.** Then `go test ./cmd/sdlc/internal/judge/...`.
- [ ] **Step 5: Commit** — `#137: enrich boundary-review prompt with repo orientation (pure)`

## Chunk 2: cmd/sdlc orientation derivation + wiring

### Task 2: `repoNameAndRoot` + `boundaryOrientation`

**Files:** Modify `cmd/sdlc/reviewsidecar.go` (add `repoNameAndRoot`, route `repoIdentity` through it); Create `cmd/sdlc/orientation.go`; Test `cmd/sdlc/orientation_test.go`

- [ ] **Step 1: Write the failing test** — a `t.TempDir()` `git init` repo whose dir basename is `pair`, with an issue file `workshop/issues/000072-x.md`. Assert `boundaryOrientation("workshop/issues", 72, "M1")` → `Repo=="pair"`, `IssueRef=="pair#72 M1"`, `Boundary` mentions `milestone M1`, `IssueFile` ends `000072-x.md`, and the note says downstream. Add a sub-case with a `construct/base.manifest` file → note says base. Negative: `IssueRef` must NOT start with `ariadne#`.
- [ ] **Step 2: Run → fails** (undefined).
- [ ] **Step 3: Implement** `repoNameAndRoot()` + `boundaryOrientation(...)` (gitx root → name; ref `<repo>#<n>[ Mx]`; `issueFilePath` for the file; `os.Stat(root/construct/base.manifest)` for base-vs-downstream; `<unknown-repo>` fallback). Refactor `repoIdentity` to reuse `repoNameAndRoot`.
- [ ] **Step 4: Fix the sidecar H1 shadow-sweep (#136 bug).** In `renderReviewEntry` (`reviewsidecar.go:71`) change the H1 from the hardcoded `ariadne#%d` to `m.Repo`-derived: `# Boundary Review — <repo>#<num> (<boundary>)` (empty → `<unknown-repo>`). `TestRenderReviewEntry` (reviewsidecar_test.go:37) stays green — its `sidecarMeta` literal sets `Repo: "ariadne"`, so the H1 still renders `ariadne#136`.
- [ ] **Step 5: Run → PASS** (orientation + reviewsidecar tests).
- [ ] **Step 6: Commit** — `#137: boundaryOrientation + fix sidecar H1 repo hardcode`

### Task 3: Wire into the shared dispatch + drop the hardcoded ref

**Files:** Modify `cmd/sdlc/milestoneclose.go` (`boundaryReviewDispatchOptions`, `boundaryReviewParams` — drop `IssueRef`), `cmd/sdlc/close.go` + `cmd/sdlc/milestoneclose.go` call sites (drop `IssueRef: "ariadne#..."`)

- [ ] **Step 1:** In `boundaryReviewDispatchOptions`, compute `o := boundaryOrientation(p.IssuesDir, p.IssueNum, p.Milestone)` and build `judge.PromptInput{Diff: diff, Base: p.BaseLong, Head: "HEAD", IssueRef: o.IssueRef, Repo: o.Repo, RepoRoot: o.RepoRoot, IssueFile: o.IssueFile, Boundary: o.Boundary, RepoNote: o.RepoNote}`.
- [ ] **Step 2:** Remove the `IssueRef` field from `boundaryReviewParams` and its three production `"ariadne#%d"` assignments: `close.go:743` (dry-run), `close.go:755` (dispatch), `milestoneclose.go:156`. Keep `Label`. (`PromptInput.IssueRef` is untouched — it's still the field the prompt reads.)
- [ ] **Step 3: Update the test surface (the field removal + H1 change force these — they do NOT "stay green" unedited):**
  - **Drop `IssueRef:` from the two `boundaryReviewParams` literals** — `closereview_test.go:156`, `milestoneclose_test.go:85` (compile break otherwise). NOTE `milestoneclose_test.go:44` is a `judge.PromptInput{IssueRef:…}` literal — **keep** it (PromptInput still has the field).
  - **Relax the repo-coupled prompt/sidecar assertions to repo-agnostic** — `closereview_test.go:114` asserts `Contains(*lastPrompt, "ariadne#69")` and `:135` asserts the sidecar H1 `# Boundary Review — ariadne#69`. Both run under `closeRepo`, which `t.TempDir()`+chdir's to a **random-basename** repo, so post-change the derived ref is `<tmpbase>#69`, not `ariadne#69` — these would FAIL. Change them to assert `"#69"` + `"(whole-issue close)"` (repo-agnostic), or compute `repoIdentity()` and assert `<that>#69`. (Pre-change they passed only via the hardcode, not a real ariadne CWD.)
  - `reviewsidecar_test.go:37` — **stays green** (explicit `Repo: "ariadne"`), confirmed.
- [ ] **Step 4: Build + full suite** — `go build ./cmd/sdlc/ && go test ./cmd/sdlc/...` all green; the dispatch tests stub `judge.Run`, so only the assertions above change, not behavior.
- [ ] **Step 5: Commit** — `#137: dispatch derives repo-correct issue ref + orientation`

## Chunk 3: Docs

### Task 4: Atlas / prompt comment

**Files:** Modify `atlas/workflow/sdlc-binary.md` (boundary-review section — note the repo-orientation contract); the `code-review.md` header itself documents it inline.

- [ ] **Step 1:** Add a sentence to the boundary-review section: the prompt orients the reviewer to the live repo (`<repo>#N`, root, issue file, boundary) derived from the git context, distinguishing base (ariadne) from downstream repos (#137).
- [ ] **Step 2: Commit** — `#137: atlas — document boundary-review repo orientation`

---

## Done-when mapping

| Issue Done-when | Delivered by |
|---|---|
| prompts name the reviewed repo accurately (ariadne + downstream fixture) | Tasks 1–3 (D1, D3) |
| prompt includes repo root, issue file, ref, base/head, boundary, milestone | Tasks 1–3 |
| tests fail if it falls back to hardcoded `ariadne#N` for non-ariadne | Task 2 (negative assertion) |
| existing verdict/trailer/gate tests pass | Task 3 (full suite) |
| help/atlas/prompt-comment documents the contract | Task 4 + the template header |

## Non-goals

- Not changing the verdict format, severity buckets, architecture/contract blocks, or any gate.
- Not deriving from the git *remote* URL (root basename is the issue's stated primary signal; remote is a later extension).
- Not touching plan-quality/estimate-quality prompts (those already take an explicit `IssueRef` and aren't the misorientation surface; could follow later).

## Revisions

- **2026-06-29 — change-code plan-quality gate (FAILURE → addressed).** Two findings,
  both incorporated above:
  1. **ARCH-PURPOSE shadow-sweep:** `reviewsidecar.go:71`'s sidecar H1 hardcodes
     `ariadne#%d` (a #136 bug) — the same misorientation #137 targets, persisted into
     the durable artifact. Added as Task 2 Step 4 (render H1 from `m.Repo`) + entity table.
  2. **Test-surface inaccuracy:** removing `IssueRef` from `boundaryReviewParams` + the
     H1 change force edits the original plan called "stay green." Task 3 now enumerates the
     exact sites (two `boundaryReviewParams` literals to de-`IssueRef`; two repo-coupled
     assertions in `closereview_test.go` to make repo-agnostic; `reviewsidecar_test.go:37`
     confirmed green) and the three production assignments (`close.go:743/755`,
     `milestoneclose.go:156`).
