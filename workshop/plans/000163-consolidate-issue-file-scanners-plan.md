# Shared Issue-File Scanner Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Consolidate the publish, warning, and archive issue-file scanners behind one parsed-record IO seam without changing behavior.

**Architecture:** A new `scanIssueFiles` integration seam owns window/directory enumeration and one-time parsing into `issueFileRef`. Pure named filters select codecomplete, not-done, or terminal records; existing callers retain mutation, logging, GitHub, and path-normalization side effects.

**Tech Stack:** Go, standard-library filesystem/path packages, existing `gitRunner`, `cmd/sdlc/internal/issue`, and `pkg/vocab`.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `issueFileRef` | `cmd/sdlc/issuefiles.go` | new |
| `issueFileScanError` | `cmd/sdlc/issuefiles.go` | new |
| `issueFilenamePattern` | `cmd/sdlc/issuefiles.go` | new |
| `issueFilename` | `cmd/sdlc/issuefiles.go` | modified |
| `codecompleteIssueFiles` | `cmd/sdlc/issuefiles.go` | new |
| `notDoneIssueFiles` | `cmd/sdlc/issuefiles.go` | new |
| `terminalIssueFiles` | `cmd/sdlc/issuefiles.go` | new |

- **`issueFileRef`** — one coherent snapshot of an issue file: path, parsed status,
  frontmatter, and body.
  - **Relationships:** one record owns one parsed snapshot; one scan returns zero or
    more records; publish/archive actions consume records without reparsing them.
  - **DRY rationale:** all four scanner families repeat read → parse → status
    extraction, and action callers otherwise need a second parse for mutation fields.
  - **Future extensions:** add another parsed field only when a new caller needs it;
    do not turn the record into a generic issue domain model.

- **`issueFileScanError`** — pure typed value carrying raw window-command output and
  the underlying cause; `Error` and `Unwrap` perform no IO.
  - **Relationships:** each failed window scan returns one error; the two callers map
    it back to their distinct established diagnostic formats.
  - **DRY rationale:** the scanner captures failure facts once without forcing callers
    to share presentation or error-wrapping policy.
  - **Future extensions:** none; add fields only if an existing diagnostic requires a
    fact unavailable from output/cause.

- **`issueFilenamePattern` / `issueFilename`** — the one six-digit issue-name grammar,
  shared by directory glob enumeration and existing issue/history path membership.
  - **Relationships:** one constant feeds both `filepath.Glob` and `filepath.Match`;
    `issueFilename` moves from `push.go` beside the new scanner without changing its
    callers.
  - **DRY rationale:** the refactor must not replace repeated scanners by introducing
    a repeated filename-pattern literal (ARCH-DRY).
  - **Future extensions:** grammar changes occur in the constant and are verified
    against both glob selection and predicate membership.

- **Named status filters** — select records for each existing caller policy while
  preserving input order.
  - **Relationships:** N:1 over `issueFileRef`; callers consume the filtered slice.
  - **DRY rationale:** `codecomplete`, non-terminal-except-codecomplete, and terminal
    membership become testable single sources instead of inline conditionals.
  - **Future extensions:** a fifth scanner reuses an existing filter or adds a focused
    predicate; avoid a callback framework until another policy demonstrates the need.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `scanIssueFiles` | `cmd/sdlc/issuefiles.go` | new | git diff, filepath glob, file reads, frontmatter parse |
| `mergedCodecompleteIssues` | `cmd/sdlc/publishgate.go` | modified | window scan through `gitx.RunGit` |
| `touchedIssuesNotDone` | `cmd/sdlc/push.go` | modified | window scan through caller `gitRunner` |
| `publishCodecompleteIssues` | `cmd/sdlc/publishgate.go` | modified | status/date file writes |
| `archiveDoneIssues` | `cmd/sdlc/push.go` | modified | GitHub close, rename, plan sweep |
| `archiveDoneIssuesInDir` | `cmd/sdlc/merge.go` | modified | main-worktree rename and relative staging paths |

- **`scanIssueFiles`** — with non-empty `baseRef`, asks the injected git function for
  `git diff --name-only baseRef..HEAD -- issuesDir/*.md`; with empty `baseRef`, globs
  only `NNNNNN-*.md`. It reads/parses each candidate once and silently skips unreadable
  or malformed records, matching current behavior.
  - **Injected into:** callers pass `gitx.RunGit` or `r.Git`; directory mode passes nil
    and does not invoke it. A typed scan error preserves raw output and unwraps the
    underlying error so caller-specific diagnostics remain unchanged. Pure filters
    receive only returned records.
  - **Future extensions:** an explicit scope type is the natural widening if a third
    enumeration mode appears; do not add it for the current two-mode contract.

- **Modified callers** — each consumes scanner results while retaining its concrete
  contract: publish gate wrapping, warning output, status/date writes, push-only GitHub
  close, rename/plan sweep, and merge-side absolute-to-relative path conversion.
  - **Injected into:** `scanIssueFiles` results feed existing action loops; no package-
    level caller seam changes.
  - **Future extensions:** side-effect consolidation is out of scope because these
    consequences intentionally differ.

## Chunk 1: Atomic scanner consolidation

### Task 1: Add the parsed scanner and pure filters with TDD

**Files:**
- Create: `cmd/sdlc/issuefiles.go`
- Create: `cmd/sdlc/issuefiles_test.go`

- [ ] **Step 1: Write failing pure-filter tests**

Add table-driven `TestIssueFileRefFilters` cases whose input order includes
`working`, `done`, `codecomplete`, missing status, `wontfix`, `open`, and `punt`.
Assert codecomplete-only, not-done (`working`, missing, `open`), and terminal
(`done`, `wontfix`, `punt`) results with order preserved.

- [ ] **Step 2: Run the pure tests and confirm RED**

Run: `go test ./cmd/sdlc -run 'TestIssueFileRefFilters' -count=1`

Expected: FAIL to compile because the record and filters do not exist.

- [ ] **Step 3: Implement the minimal record and pure filters**

```go
type issueFileRef struct {
	Path        string
	Status      string
	Frontmatter string
	Body        string
}

func codecompleteIssueFiles(refs []issueFileRef) []issueFileRef
func notDoneIssueFiles(refs []issueFileRef) []issueFileRef
func terminalIssueFiles(refs []issueFileRef) []issueFileRef
```

Use `vocab.Issue().IsTerminal` for category membership and keep `codecomplete` as the
value-specific carve-out. Return new slices in input order (ARCH-PURE, ARCH-DRY).

- [ ] **Step 4: Run the pure tests and confirm GREEN**

Run: `go test ./cmd/sdlc -run 'TestIssueFileRefFilters' -count=1`

Expected: PASS.

- [ ] **Step 5: Write failing integration tests for both scan modes**

Use a real temporary git repository plus `execGitRunner{}`. Pin:

- window mode includes changed `custom.md` and six-digit files in git order;
- directory mode includes only sorted six-digit `NNNNNN-*.md` files;
- `issueFilename` and directory-mode globbing accept/reject the same fixture names,
  and the six-digit glob literal appears only once in production source;
- deleted/unreadable/malformed candidates are skipped;
- missing `status` produces `Status == ""`;
- a failing window runner returns an error;
- the typed error retains raw command output and supports `errors.Is`/`errors.As` for
  the underlying failure;
- returned frontmatter/body support `SetField` + `Compose` without another read.

- [ ] **Step 6: Run the scanner tests and confirm RED**

Run: `go test ./cmd/sdlc -run 'TestScanIssueFiles' -count=1`

Expected: FAIL to compile because `scanIssueFiles` does not exist.

- [ ] **Step 7: Implement the minimal integration seam**

```go
func scanIssueFiles(baseRef, issuesDir string, runGit func(...string) ([]byte, error)) ([]issueFileRef, error)
```

Window mode uses `issuesDir+"/*.md"` and preserves git output order. Move the existing
`issueFilename` predicate from `push.go` into `issuefiles.go`, define one
`issueFilenamePattern`, and have directory mode join that constant into its glob while
the predicate passes it to `filepath.Match`. Sort directory matches. Read, parse, and extract status once per
path; silently skip read/parse failures. Return a failed window runner error. Perform
no writes or caller policy here. On git failure return an `issueFileScanError` with
`Output []byte`, `Err error`, `Error()`, and `Unwrap()`.

- [ ] **Step 8: Run focused tests and confirm GREEN**

Run: `go test ./cmd/sdlc -run 'Test(IssueFileRefFilters|ScanIssueFiles)' -count=1`

Expected: PASS.

- [ ] **Step 9: Commit the scanner core**

```bash
gofmt -w cmd/sdlc/issuefiles.go cmd/sdlc/issuefiles_test.go cmd/sdlc/push.go
git add cmd/sdlc/issuefiles.go cmd/sdlc/issuefiles_test.go cmd/sdlc/push.go
git commit -m "#163: add shared issue-file scanner" -m "Centralize issue enumeration and parsing while keeping status policy pure and caller effects outside the seam." -m "Co-Authored-By: OpenAI Codex <noreply@openai.com>"
```

### Task 2: Rewire window-scoped callers

**Files:**
- Modify: `cmd/sdlc/publishgate.go`
- Modify: `cmd/sdlc/push.go`
- Modify: `cmd/sdlc/publishgate_test.go`
- Modify: `cmd/sdlc/push_test.go`

- [ ] **Step 1: Strengthen caller tests before rewiring**

Pin that `mergedCodecompleteIssues` returns only codecomplete paths and preserves its
exact `git diff <base>..HEAD: <cause>` message plus `errors.Is` chain; that
`touchedIssuesNotDone` formats missing status as `unset`, preserves order, and excludes
terminal plus `codecomplete`, while its failure message retains combined runner output.

- [ ] **Step 2: Run the strengthened tests before refactor**

Run: `go test ./cmd/sdlc -run 'Test(MergedCodecompleteIssues|TouchedIssuesNotDone)' -count=1`

Expected: PASS, proving the assertions describe current behavior.

- [ ] **Step 3: Rewire `mergedCodecompleteIssues`**

Call `scanIssueFiles(baseRef, issuesDir, gitx.RunGit)`, filter with
`codecompleteIssueFiles`, and return record paths. Keep the function and
`runPublishGateFn` signatures unchanged. Convert `issueFileScanError` back to the
existing `%w` diagnostic.

- [ ] **Step 4: Rewire `touchedIssuesNotDone`**

Call `scanIssueFiles(baseRef, issuesDir, r.Git)`, filter with `notDoneIssueFiles`, and
format `path (status: valueOr(status, "unset"))`. Remove its read/parse/membership
boilerplate. Pass `r.Git` and preserve the current combined-output diagnostic.

- [ ] **Step 5: Run window caller regressions**

Run: `go test ./cmd/sdlc -run 'Test(MergedCodecompleteIssues|TouchedIssuesNotDone|RunPublishGate)' -count=1`

Expected: PASS.

- [ ] **Step 6: Commit the window rewiring**

```bash
gofmt -w cmd/sdlc/publishgate.go cmd/sdlc/publishgate_test.go cmd/sdlc/push.go cmd/sdlc/push_test.go
git add cmd/sdlc/publishgate.go cmd/sdlc/publishgate_test.go cmd/sdlc/push.go cmd/sdlc/push_test.go
git commit -m "#163: route window scans through shared helper" -m "Make publish and warning windows derive from one parsed source while preserving their distinct git diagnostics." -m "Co-Authored-By: OpenAI Codex <noreply@openai.com>"
```

### Task 3: Rewire directory-wide publish and archive callers

**Files:**
- Modify: `cmd/sdlc/publishgate.go`
- Modify: `cmd/sdlc/push.go`
- Modify: `cmd/sdlc/merge.go`
- Modify: `cmd/sdlc/publishgate_test.go`
- Modify: `cmd/sdlc/push_test.go`
- Modify: `cmd/sdlc/merge_test.go`
- Verify: `cmd/sdlc/archiveartifacts_test.go`
- Verify: `cmd/sdlc/merge_e2e_test.go`

- [ ] **Step 1: Strengthen directory characterization tests**

Before rewiring, pin the current externally visible contracts with exact named tests:

- `TestPublishCodecompleteIssues` asserts status and `updated` are rewritten while body
  bytes remain unchanged;
- `TestArchiveDoneIssues_MovesAndClosesGH` asserts only literal `done` closes GitHub and
  returned paths remain caller-relative;
- `TestArchiveDoneIssuesInDir_MovesTerminalAndRecordsRelativePaths` asserts terminal
  selection and mainPath-relative staging records;

- [ ] **Step 2: Run characterization tests before refactor**

Run: `go test ./cmd/sdlc -run 'Test(PublishCodecompleteIssues|ArchiveDoneIssues|ArchiveDoneIssuesInDir)' -count=1`

Expected: PASS, proving the assertions describe existing behavior. This refactor's RED
tests belong to the new scanner/filter entities; caller characterization is green
before and after.

- [ ] **Step 3: Rewire `publishCodecompleteIssues`**

Use `scanIssueFiles("", issuesDir, nil)` plus `codecompleteIssueFiles`. Update each
record's frontmatter/body, preserving updated-date behavior and order. The write loop
and its existing error return remain structurally unchanged.

- [ ] **Step 4: Rewire `archiveDoneIssues`**

Use directory scan plus `terminalIssueFiles`; read `github_issue` from the record.
Preserve push-only GitHub close, mkdir/rename, recorded paths, plan sweep, logging, and
the existing action-loop error returns.

- [ ] **Step 5: Rewire `archiveDoneIssuesInDir`**

Scan `filepath.Join(mainPath, issuesDir)`, filter terminals, preserve no-GitHub
behavior, and keep absolute scan paths separate from mainPath-relative staging paths.

- [ ] **Step 6: Run directory behavior tests**

Run: `go test ./cmd/sdlc -run 'Test(PublishCodecompleteIssues|ArchiveDoneIssues|ArchiveDoneIssuesInDir|PushPublishSequence|RunMerge_Codecomplete)' -count=1`

Expected: PASS, including real-repo plan/sidecar archive cases.

- [ ] **Step 7: Prove structural consolidation**

Run the Task 4 ARCH-DRY `rg` sweep before committing. Behavior-equivalent duplicated
code can keep characterization tests green, so the source sweep—not an artificial
mock seam—is the direct proof that all five caller functions derive from the helper.

- [ ] **Step 8: Format and commit directory caller rewiring**

```bash
gofmt -w cmd/sdlc/issuefiles.go cmd/sdlc/issuefiles_test.go cmd/sdlc/publishgate.go cmd/sdlc/publishgate_test.go cmd/sdlc/push.go cmd/sdlc/push_test.go cmd/sdlc/merge.go cmd/sdlc/merge_test.go
git add cmd/sdlc/issuefiles.go cmd/sdlc/issuefiles_test.go cmd/sdlc/publishgate.go cmd/sdlc/publishgate_test.go cmd/sdlc/push.go cmd/sdlc/push_test.go cmd/sdlc/merge.go cmd/sdlc/merge_test.go
git commit -m "#163: route directory scans through shared helper" -m "Remove parallel glob-and-parse loops while preserving publish mutations and the distinct push/merge archive consequences." -m "Co-Authored-By: OpenAI Codex <noreply@openai.com>"
```

### Task 4: Reconcile artifacts and verify the atomic change

**Files:**
- Modify: `workshop/issues/000163-consolidate-issue-file-scanners-into-a-shared-helper.md`
- Modify: `workshop/plans/000163-consolidate-issue-file-scanners-plan.md`
- Inspect: `atlas/`

- [ ] **Step 1: Format and run focused tests**

Run:

`gofmt -w cmd/sdlc/issuefiles.go cmd/sdlc/issuefiles_test.go cmd/sdlc/publishgate.go cmd/sdlc/publishgate_test.go cmd/sdlc/push.go cmd/sdlc/push_test.go cmd/sdlc/merge.go cmd/sdlc/merge_test.go`

Then:

`go test ./cmd/sdlc -run 'Test(IssueFileRefFilters|ScanIssueFiles|MergedCodecompleteIssues|TouchedIssuesNotDone|RunPublishGate|PublishCodecompleteIssues|ArchiveDoneIssues|ArchiveDoneIssuesInDir|PushPublishSequence|RunMerge_Codecomplete)' -count=1`

Expected: PASS.

- [ ] **Step 2: Run full verification**

Run `go test ./cmd/sdlc -count=1`, `go test ./... -count=1`,
`git diff --check "$(git merge-base main HEAD)"..HEAD`, and `git diff --check`.

Expected: all tests PASS and whitespace check prints nothing.

- [ ] **Step 3: Perform the ARCH-DRY shadow sweep**

Run:

```bash
rg -n 'issue.Parse|GetField\(fm, "status"\)|Glob\(filepath.Join\(.*\[0-9\]' cmd/sdlc/publishgate.go cmd/sdlc/push.go cmd/sdlc/merge.go
```

Expected: none of the four scanner families retains enumeration + parse + status-read
boilerplate. Explain any remaining parse as a behaviorally distinct job. Also run
`rg -n '\[0-9\]\[0-9\]\[0-9\]\[0-9\]\[0-9\]\[0-9\]-\*\.md' cmd/sdlc --glob '*.go'`
and confirm the production pattern has one definition (test fixtures may repeat it).

- [ ] **Step 4: Assess atlas impact**

Search `atlas/` for moved names and scanner descriptions. This is an internal refactor;
record “no atlas surface change” in the issue Log if no live map points at the old
implementation.

- [ ] **Step 5: Reconcile issue and plan state**

Check completed issue/plan boxes, append verification and ARCH outcomes to `## Log`,
and append a timestamped `## Revisions` entry if execution changed this plan.

- [ ] **Step 6: Commit completion records**

```bash
git add workshop/issues/000163-consolidate-issue-file-scanners-into-a-shared-helper.md workshop/plans/000163-consolidate-issue-file-scanners-plan.md
git commit -m "#163: record scanner consolidation verification" -m "Keep the durable execution record aligned with the verified implementation and close evidence." -m "Co-Authored-By: OpenAI Codex <noreply@openai.com>"
```

- [ ] **Step 7: Close through the single SDLC boundary**

Run `sdlc actual --issue 163`, inspect the measured window, then run:

```bash
sdlc close --issue 163 --no-atlas --verified '<focused + full Go tests; ARCH-DRY source sweep; branch + worktree diff checks; no command/workflow surface change>'
```

Do not run a separate boundary review: `sdlc close` owns the mandatory fresh-context
review and must report no unresolved Critical/Important findings before completion.

## Revisions

### 2026-07-13T00:27:00-07:00 — fresh-context plan review

- Replaced grouped concept-table rows with the five concrete modified caller symbols.
- Added a typed scan-error contract and exact caller diagnostic characterization so
  the shared IO seam does not erase distinct `gitx.RunGit`/`gitRunner.Git` behavior.
- Removed the artificial directory-caller RED/mutation test; new scanner entities use
  TDD, existing callers use green-before/green-after characterization, and the source
  sweep proves structural consolidation.
- Added per-commit formatting, why bodies, co-author trailers, exact verify-only test
  files, and branch-window plus working-tree whitespace checks.

### 2026-07-13T00:34:00-07:00 — plan review follow-up

- Added `issueFileScanError` to the load-bearing pure-entity inventory.
- Removed an optional partial-result test promise that had no deterministic named
  setup; action-loop error handling remains unchanged while scanner-specific failures
  have exact tests.
- Replaced the stale close-evidence “mutation check” label with the actual ARCH-DRY
  source sweep and both committed-window and working-tree diff checks.

### 2026-07-13T00:47:00-07:00 — change-code plan-quality refusal

- Added the existing `issueFilename` predicate and new shared pattern constant to the
  concept inventory. The implementation now relocates the predicate beside the
  scanner, derives both glob and match behavior from one grammar, tests their
  equivalence, and structurally sweeps for duplicate production literals.
