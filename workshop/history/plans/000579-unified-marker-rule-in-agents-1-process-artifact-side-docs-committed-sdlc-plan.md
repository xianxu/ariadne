---
collaboration-mode: discussion
---

# Unified marker rule; process-artifact side docs committed; sdlc refuses uncommitted links — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** One marker rule for every artifact (a question opens a side doc, an instruction acts, "fold in" collapses), side docs of process artifacts committed with them, and an sdlc guard that keeps a committed artifact from linking to a file that will never reach the repository.

**Architecture:** The rule is prose in `AGENTS.base.md` §1 and the review convention, pointing at the model ariadne#575 already wrote (`construct/local/universal-review/collaboration-modes.md` §1). The guard is a pure link census over an issue's process documents (`sidedoc.LocalLinks`, extracted from `RebaseLinks` so both share one link grammar) judged against a git-backed status seam; `issue publish`, `change-code` and `close` call it before their first effect.

**Tech Stack:** Go (`pkg/sidedoc`, `cmd/sdlc`), git plumbing (`cat-file -e`, `check-ignore`), Markdown.

---

## Context the design rests on

- **What reaches main today** (surveyed 10-23): `issue publish` writes the details file and its `assets/<id>/…` bundle only (republish.go:223-273; `issueassets.ParseRefs` is a closed set). `change-code`'s `checkpointDesign` commits the issue, the plan and the plan-gate ledger only (changecode.go:284-318). Nothing commits a `.side/` tree before the archive (archiveside.go, push.go:725, landingarchive.go:553).
- **Consequence:** on an issue branch a side doc is committed by an ordinary `git commit`, so a refusal at change-code/close has a one-line fix. On a resting branch (shaping under a claim, before start-plan) nothing may be committed (#284), so a side doc can reach main only if **publish carries it**. A publish guard without that would leave "commit the side doc" impossible there.
- **Decision for the TL (D1):** publish carries the side docs the details link to, as it carries assets. Recommended; Task 4. The alternative is that publish refuses and the fix is "fold it in, or start-plan and commit it on the branch"; that is smaller but makes side docs unusable while shaping, which is where #575's questions happen most.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `LocalLink` / `LocalLinks` | `pkg/sidedoc/rebase.go` | new (extracted) |
| `RebaseLinks` | `pkg/sidedoc/rebase.go` | modified (uses `LocalLinks`) |
| `LinkState` | `cmd/sdlc/internal/linkguard/guard.go` | new |
| `Doc`, `Refusal`, `Census` | `cmd/sdlc/internal/linkguard/guard.go` | new |
| `FormatRefusals` | `cmd/sdlc/internal/linkguard/guard.go` | new |

- **LocalLinks(text, from) []LocalLink** — every relative link target in a document (inline links and images, reference definitions, frontmatter `parent:`), outside fenced code and code spans, resolved to a root-relative slash path; URLs, absolute paths, bare fragments and targets that leave the root are not local. `LocalLink{Target (as written), Path (root-relative, fragment dropped), Line}`.
  - **DRY rationale (ARCH-DRY):** `RebaseLinks` already owns this grammar privately; the guard is its second consumer, so the enumeration moves out and `RebaseLinks` rewrites over it. One grammar decides both "what moves with the archive" and "what must be committed".
  - **Future extensions:** a dangling-link check (a target that exists nowhere) is a third consumer.
- **LinkState** — `Committed | Uncommitted | Ignored | Absent`. A tagged enum, not booleans (ARCH-ORDER: one value per target, read once).
  - `Committed`: in the tree the verb publishes (HEAD for change-code/close; main-after-publish for publish). `Uncommitted`: on disk, not in that tree, not ignored. `Ignored`: on disk, not in the tree, matched by git's ignore rules. `Absent`: neither on disk nor in the tree — a dangling link, out of scope here, passes (the archive's `RebaseLinks` already reports dangling).
- **Census(docs []Doc, state func(path) (LinkState, error)) ([]Refusal, error)** — walks each document's `LocalLinks`, asks the state of each target once, refuses `Uncommitted` and `Ignored`, and follows links into Markdown targets that are themselves side docs (`sidedoc.InSideDir`) so a nested side doc is checked too; a visited set bounds the walk (ARCH-CONSTRAINTS: O(links) over a handful of documents, no fan-out beyond the issue's own tree). `Doc{Path, Text}`; `Refusal{Doc, Line, Target, State}`.
  - Pure: the state function is injected; tests pass a map.
- **FormatRefusals** — one line per refusal naming the document, line, link and the fix: for `Uncommitted` "commit it with the artifact (`git add <path>`), or fold it in and drop the link"; for `Ignored` "it is git-ignored (a private session side doc?): fold what it settled into the artifact and drop the link".

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `gitLinkState` | `cmd/sdlc/linkguard.go` | new | `git cat-file -e`, `git check-ignore -q`, `os.Lstat` |
| `issueProcessDocs` | `cmd/sdlc/linkguard.go` | new | issue + plans (`issueArchiveDocs`) + projects (`project.DiscoverInRoots`) |
| `guardIssueLinks` | `cmd/sdlc/linkguard.go` | new | the one call each verb makes |
| publish side carry | `cmd/sdlc/republish.go` | modified (D1) | the narrow main commit |

- **gitLinkState(r gitRunner, root, tree string)** — `cat-file -e <tree>:<path>` (committed; a directory target counts when the tree has it), else `Lstat` (absent), else `check-ignore -q` (ignored vs uncommitted). Runs through the verb's existing runner, so tests use the repos the verbs' fixtures already build (real git in a temp dir — the stateful fake the sdlc tests already use for git, ARCH-MOCK).
- **issueProcessDocs** — the issue file, its plan documents (the archive's own list, `issueArchiveDocs`, archiveside.go:34, so "what archives with the issue" and "what is guarded" are one list, ARCH-DRY), and the active project files that reference the issue (`project.DiscoverInRoots` with `projectWorkspaceRoots`, as close does at close.go:826). A project in another repository is judged against that repository's HEAD with its own runner root.
- **guardIssueLinks(...) error** — builds the docs, runs `Census`, returns a refusal error carrying `FormatRefusals`. Called:
  - `change-code`: a new `side-links` gate first in `changeCodeGates` (it is free like structural), bypass `--no-side-links` (or `--force`).
  - `close` / `milestone-close`: in `computeClose` beside the plan-check gate, bypass `--no-side-links` (or `--force`), with its row in `processmanual.GateCatalog` (the drift test `TestGateCatalogMatchesRegisteredFlags` requires it).
  - `issue publish`: in `runTrackedPublish`'s per-issue loop before any effect (issuepublish.go:93-117), so first-time creation is covered too. No bypass flag: the fix is always available (commit, carry or fold).

### Persisted formats (ARCH-EVOLVE)

None changed. The publish commit gains files under the details' `.side/` directory; no record or ledger shape changes.

### Lifecycle (ARCH-FUNERAL)

Creates nothing durable of its own: the guard reads. Side docs it makes publish carry are the artifact's own `.side/` tree, which the existing archive already moves to history with the issue (archiveside.go) — their end is the issue's.

### Trust (ARCH-SECURE)

Link targets are untrusted text from a hand-edited file. A target is only ever passed to git as one argv element after `path.Clean` and the root check `LocalLinks` already applies (no `..` escape, no absolute path, no scheme); a target beginning with `-` is refused as not local. Publish carries a side member only when it is a regular file inside the details' own `.side/` directory (no symlink, no other directory), so a crafted link cannot make publish write elsewhere on main.

### Ordering (ARCH-ORDER)

Holds no state between events: each verb computes the census once before its first effect, from the bytes it is about to publish. A side doc changed after the guard and before the commit is the same race as an issue edited mid-verb, which the verbs already handle by reading once.

## Tasks

### Task 1: The rule in prose

**Files:** `AGENTS.base.md:15`, `construct/local/fix/review-convention.md` (§6, Revisions), `construct/local/universal-review/collaboration-modes.md` §4 (one paragraph on process artifacts).

- [ ] Replace the §1 marker sentence with: markers as before; a ready `🤖[…]` in **any** artifact is the operator's ask — a question opens a side doc and the marker becomes `→ [discussion](…)`, an instruction is carried out and the marker dropped, and a `🤖[fold in]` beside a link folds the settled side doc back (model: collaboration-modes.md §1). Side docs of process artifacts (issues, plans, projects) live at `<artifact>.side/…`, are committed with the artifact and archive with it; the operator's private session side docs stay uncommitted, so a committed artifact never links to one (sdlc refuses the link).
- [ ] review-convention §6: the "answer it and resolve the marker that same turn" paragraph routes by the same rule (question → side doc, instruction → act); the clean-prose default applies to commands and fold-ins. Add a Revisions line (ariadne#579).
- [ ] collaboration-modes §4: process artifacts' side docs are committed with them; sdlc enforces it at publish, change-code and close.
- [ ] `weave refresh` regenerates the composed AGENTS.md and the `.agents/` copy; check both read right.
- [ ] Commit `#579: AGENTS §1 and the convention state the unified marker rule`.

### Task 2: `sidedoc.LocalLinks`, `RebaseLinks` over it

**Files:** `pkg/sidedoc/rebase.go`, `pkg/sidedoc/rebase_test.go`.

- [ ] Failing test `TestLocalLinks`: inline, image, `"title"`, reference definition, frontmatter `parent:`, `#fragment` dropped, trailing-slash directory; excluded: URL, `mailto:`, absolute, `#only`, `../` leaving the root, fenced code, code span, a `-`-leading target.
- [ ] Extract the enumeration from `RebaseLinks` into `LocalLinks`; `RebaseLinks` iterates it. Existing `rebase_test.go` passes unchanged (the refactor is behavior-preserving).
- [ ] Commit.

### Task 3: `linkguard` census and the three call sites

**Files:** create `cmd/sdlc/internal/linkguard/guard.go`, `guard_test.go`; create `cmd/sdlc/linkguard.go`, `linkguard_test.go`; modify `changecode.go` (gate + flag), `close.go` (flag, `skip`, check in `computeClose`), `internal/processmanual/gatesig.go` (row), `issuepublish.go`.

- [ ] Failing `TestCensus` (pure, map-backed state): committed passes; uncommitted and ignored refuse with document, line and target; absent passes; a nested side doc linking an uncommitted file refuses; a link cycle between two side docs terminates; each target is asked once.
- [ ] Implement `LinkState`, `Census`, `FormatRefusals`.
- [ ] Failing `TestGitLinkState` on a temp repo: committed file, committed directory, untracked file, ignored file, absent path.
- [ ] Implement `gitLinkState`, `issueProcessDocs`, `guardIssueLinks`.
- [ ] Failing verb tests (each refuses, then passes once the side doc is committed, and passes once the link is removed): `TestChangeCodeRefusesUncommittedSideLink`, `TestCloseRefusesIgnoredSideLink` (plus `--no-side-links` acknowledges), `TestCloseRefusesProjectLinkToUncommittedFile`, `TestIssuePublishRefusesUncommittedLink` (a link outside the details' `.side/`, which publish never carries).
- [ ] Wire the gate into change-code, close and publish; add the catalog row; `TestGateCatalogMatchesRegisteredFlags` passes.
- [ ] Commit.

### Task 4 (D1): publish carries the linked side docs

**Files:** `cmd/sdlc/republish.go`, `cmd/sdlc/republish_test.go`.

- [ ] Failing `TestIssuePublishCarriesLinkedSideDocs`: details on a resting branch link `→ [discussion](<issue>.side/2026-10/…md)`, which links a nested side doc; publish puts details and both side docs on main in its one commit; the guard passes; a side doc linked but outside the issue's `.side/` still refuses; a symlink in `.side/` refuses.
- [ ] In the per-issue read, collect the `Census`-reached targets inside `sidedoc.SideDir(details)`; their bytes join `w` in `prepare`. They count as `Committed` for the guard (the publish's tree is main-after-commit). Unlinked side files are not carried (the link set is the closed set, as with assets).
- [ ] Commit.

### Task 5: atlas, verification

- [ ] `atlas/workflow/sdlc-binary.md`: the side-link guard (verbs, states, bypass) and publish's side carry; `atlas/index.md` unchanged unless a file is added.
- [ ] Targeted runs: `go test ./pkg/sidedoc/ ./cmd/sdlc/internal/linkguard/` and `go test ./cmd/sdlc/ -run 'SideLink|LinkState|IssuePublish|GateCatalog|Rebase|Archive'`.
- [ ] Recorded suite before close: `sdlc local-check run --name suite`.

## Done when (from the issue)

- AGENTS §1 and the review convention state the unified rule and commit-with-artifact (doc check: read both after `weave refresh`).
- `sdlc issue publish`, `change-code` and `close` refuse a link to an uncommitted or ignored file and pass once it is committed or removed (Task 3 tests); with D1, publish passes once the side doc is carried (Task 4 test).

## Revisions

- **2026-10-23** — TL (ops:0) approved the plan as written, and D1 as recommended: publish carries the side docs the details link to (regular files inside the issue's own `.side/` only, a closed set like assets), because side docs must be usable while shaping. Task 4 is in scope.
- **2026-10-23** — as built:
  - change-code's check runs on every flow, before the gate loop, not as a `changeCodeGates` row: that list is skipped entirely on the quick flow (`activeChangeCodeGates`). Same `--no-side-links` flag, same `--force <reason>` bypass and catalog row (G2).
  - `issue publish` judges only what it publishes, the details (plans and projects reach main through the branch, where change-code and close judge them). A first publication (pre-#308 transfer) carries no side docs, so a link main lacks refuses there.
  - The issue's own asset bundle counts as committed at publish: publication carries and checks it.
  - No `-`-leading rule: targets reach git after `--` or as `tree:path`. Committed means `ls-tree` lists it (HEAD, or main for publish); ignored means `ls-files --others --ignored --exclude-standard` lists it.
  - A carried side doc is written only if main lacks it or holds the base's copy, re-checked inside the publish commit; the checkout finish restores (rest) or commits (issue branch) the carried paths with the details.
- **2026-10-23** — close review round 1 (FIX-THEN-SHIP):
  - BR-1: close skips the census entirely under `--no-side-links`/`--force` (the acknowledgement is "skipping the side-link check"), and a project that cannot be found or judged (discovery error, an unborn peer HEAD) is a warning, as close's own project update treats it; an error in the issue's own repository still refuses, with the bypass named.
  - BR-2: `TestChangeCodeRefusesUncommittedSideLink` drives `runChangeCode` (refusal and exit 1, `--force` acknowledgement, `--no-side-links`, pass once committed); `TestIssueSideLinkRefusalsJudgeProjects` covers the peer-repository project branch (unborn peer warns; committed peer refuses its uncommitted link, named `peerrepo/…`). The planned `TestCloseRefusesProjectLinkToUncommittedFile` is that test: the project branch is shared by change-code and close.
  - Minors: literal pathspecs (`:(literal)`); `gitLinkState` lost its unused tree parameter; publish's pre-effect check runs only when a first publication shares the run (republishOwned's own check is otherwise first), and no longer drops a `Require` error; change-code resolves the plans dir like close; a carried side doc is read only if it is the same regular file before and after the read.
