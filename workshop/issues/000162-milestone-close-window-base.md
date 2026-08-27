---
id: 000162
status: working
deps: []
github_issue:
created: 2026-07-02
updated: 2026-08-26
estimate_hours: 2.05
started: 2026-08-25T08:21:30-07:00
---

# sdlc milestone-close derives gate/review windows from a wrong base (far-back or HEAD)

## Problem

`sdlc milestone-close` computes the commit window it uses for (a) the auto
boundary-review and (b) the atlas-change gate from a **wrong base**. Observed two
distinct manifestations of the same root cause while shipping the pair `#99`
launcher port (a multi-milestone issue, `Mx` boundaries closed one at a time):

**Variant 1 — review window picks a far-back base → `argument list too long`.**
On the *first* milestone of a freshly-branched issue, the auto-computed
boundary-review window was `<far-back-unrelated-commit>^..HEAD` — a 566-file,
~6.8 MB diff. The review dispatch `fork/exec`s the `claude` CLI with the diff +
prompt inline, so the oversized argv trips **E2BIG (`fork/exec claude: argument
list too long`)**; the close aborts with verdict `not-run` and leaves the issue
`working`. Reproduced twice on pair `#99` M1. Not a PATH-size problem (a minimal
PATH still fails), not a code problem — purely the window base.

**Variant 2 — atlas-gate window picks `base = HEAD` → empty window.** On a *later*
milestone (pair `#99` M2), after committing the milestone code (whose atlas
updates landed in an *earlier* commit of the same milestone), `milestone-close`'s
atlas gate reported `no atlas/ changes in <lastCommit>..HEAD` and aborted — its
window base was the just-made HEAD commit, so the real, in-milestone atlas edits
one/two commits back were outside the window. The atlas requirement *was* met; the
gate's window was wrong.

Both windows should be anchored to the **milestone's own extent** — the milestone's
first commit, or (for multi-`Mx` issues) the previous `Mx` boundary — not a far-back
unrelated base and not `HEAD` itself.

## Spec

`milestone-close` (and the boundary-review + atlas/plan gates it drives) should
derive `BASE_SHA` deterministically as the milestone's lower bound:

- For a multi-`Mx` issue: the previous milestone's close boundary (the commit
  carrying the prior `Review-Verdict:` trailer / `closed M<k-1>` marker).
- For the first milestone (or un-tagged single-pass work): the branch point
  (`git merge-base <default-branch> HEAD`), not a far-back ancestor.
- The atlas/plan gates must use the **same** window as the review, so an atlas
  change anywhere in the milestone counts.

Guard the review dispatch against E2BIG regardless (pass the diff via a temp file
/ stdin, not an inline argv) so a large-but-legitimate window degrades gracefully.

## Revisions

### 2026-08-26 — replace the embedded boundary diff with a pinned review manifest

**Reason:** pair#149 reproduced the remaining transport failure with a correctly
bounded but large M5 window. Its unified diff was 930,583 bytes and the rendered
Codex prompt was 937,639 bytes. The npm Codex launcher on Node 26 deterministically
stack-overflowed when `child_process.spawn` forwarded a 930 KB argument, while the
native Codex binary accepted the same input. Passing the same payload over stdin
would avoid argv limits, but it would still duplicate data the reviewer already
loads from the repository: all ten inspected boundary-review sessions ran
`git diff` themselves (41 diff calls and 58 Git-inspection calls total).

**Decision:** a boundary review receives a compact, deterministic
`ReviewWindowManifest`, not unified diff bytes. The manifest contains the absolute
repository root, immutable full base/head commit SHAs, issue/boundary identity,
the issue path, prior structured findings, the current code-review path
exclusions, and exact read-only commands for `git diff --stat`,
`git diff --name-status`, and the full/targeted patch. Commands preserve the
existing exclusion of `workshop/issues/` and `workshop/history/`; the reviewer
reads the issue and plan explicitly through their named paths. Automatic close
reviews always use the already captured concrete head. Manual
`sdlc judge milestone-review` resolves supplied refs to immutable commits; when
the head is intentionally omitted, the manifest says the working tree is in
scope and renders the corresponding base-vs-working-tree command.

The reviewer must inspect the pinned range through repository tools and fail
closed if it cannot. The prompt keeps the compact `PriorFindings` handoff because
that is gate state, not repository patch data. `dry`, `pure`, `plan`, and `specs`
retain their existing inline-diff transport: they are separate judge contracts,
and widening all of them is not required to unblock boundary review. Reviewer
checkout isolation remains #204.

**Alternatives rejected:** sending the full patch on stdin fixes only transport
and still front-loads large generated mirrors into model context; writing a temp
patch file creates lifecycle/cleanup state while duplicating Git's object store.
Pinned Git commands reuse the repository as the single source of truth
(ARCH-DRY), keep command rendering pure while Git/ref resolution stays at the IO
boundary (ARCH-PURE), and preserve the exact reviewed window rather than solving
only the observed Codex wrapper crash (ARCH-PURPOSE). Git remains behind the
existing `gitx.RunGit` seam and repository fixtures exercise ref resolution and
working-tree behavior (ARCH-MOCK).

**Done-when delta:**

- An automatic boundary prompt remains bounded when the reviewed patch contains
  a multi-megabyte sentinel and contains no sentinel bytes.
- The manifest carries full immutable base/head SHAs and commands whose pathspecs
  match the former boundary `collectDiff` exclusions exactly.
- Manual milestone review pins explicit refs; omitted head is explicitly and
  correctly represented as a working-tree review.
- Existing verdict parsing, prior-finding convergence, fresh-process dispatch,
  and non-boundary judge prompts remain byte-compatible except for intentional
  documentation wording.

### 2026-08-26 — make the manifest and failure contract executable

**Reason:** fresh-context review found that the original checklist below still
prescribed stdin/temp transport and that the first manifest revision left the
working-tree variant, custom directories, and failure semantics implicit.

**Decision:** supersede the original `## Plan` transport item with the revised
plan below. The boundary-only prompt input retains the existing orientation and
gate fields (`Repo`, `RepoRoot`, `IssueRef`, `IssueFile`, `Boundary`, `RepoNote`,
and `PriorFindings`) and adds a typed review-window manifest with two explicit
variants:

- **Committed range:** `BaseSHA` and `HeadSHA` are full object IDs resolved as
  commits before dispatch; `WorkingTree` is false.
- **Working tree:** `BaseSHA` and `AmbientHeadSHA` are full commit object IDs,
  `HeadSHA` is empty, and `WorkingTree` is true. The range includes committed
  changes after the base plus staged and unstaged tracked changes; as with the
  current `git diff <base>` contract, untracked files are excluded.

Both variants carry the optional canonical durable-plan path when one exists.
The manifest renders argv as structured command plus arguments, and display
text shell-quotes each argument; it never constructs executable shell source.
Its stat, name-status, full-patch, and targeted-patch recipes reuse the effective
`--issues-dir` and `--history-dir` values, including environment overrides, for
the same exclusions as the former boundary `collectDiff` call.

Automatic dispatch refuses symbolic, missing, or non-commit base/head anchors
and an unavailable repository root before launching the reviewer. Manual
milestone review resolves explicit refs to commits first; omitted head selects
the working-tree variant and captures the ambient `HEAD`. The prompt instructs
the reviewer to return REWORK when repository inspection or a required command
is unavailable. This is the enforceable boundary: the dispatcher validates the
inputs and verdict protocol, while the reviewer remains responsible for using
the supplied read-only inspection recipes; `sdlc` does not claim to prove tool
use after launch.

Ref resolution is a PURE decision over an injected Git command seam. Unit tests
use a small stateful fake that records argv and maps refs to object IDs; a real
temporary Git repository provides conformance coverage for explicit refs and
working-tree semantics (ARCH-MOCK). Prompt rendering is pure and receives only
already-resolved data (ARCH-PURE).

**Revised implementation plan (supersedes the unchecked transport item below):**

- [x] Define and render the two `ReviewWindowManifest` variants with exact,
      safely quoted Git argv and effective issue/history exclusions.
- [x] Resolve and validate repository root, base/head commits, ambient `HEAD`,
      issue path, and optional canonical plan path at the IO boundary before
      automatic or manual boundary dispatch.
- [x] Replace only milestone-review's embedded `Diff` with the manifest; retain
      all orientation, prior-finding, verdict, and fresh-process contracts.
- [x] Add unit coverage with the stateful Git fake and live temporary-repository
      conformance tests for explicit ranges, omitted-head working-tree scope,
      invalid refs/repositories, custom directories, safe argument rendering,
      and a multi-megabyte sentinel absent from the resulting prompt.
- [x] Pin every non-boundary prompt golden byte-for-byte and update process/atlas
      documentation for the boundary-only transport change.

### 2026-08-26 — classify Git as integration and pin command sources

**Reason:** the second spec review caught an effects-based classification error
and ambiguity about whether this work should add a history-directory option to
automatic close.

**Decision:** Git execution and ref resolution are INTEGRATION even when invoked
through an injected runner. Only validation of already-returned object IDs and
manifest/command rendering are PURE (ARCH-PURE). The stateful fake models Git's
ref map and command history behind the same runner interface used by production;
temporary real repositories remain the live conformance layer (ARCH-MOCK).

Directory inputs preserve each caller's current contract rather than widening
flags in this issue:

- Manual `sdlc judge milestone-review` uses its effective `--issues-dir` and
  `--history-dir` values, including `WF_ISSUES_DIR` / `WF_HISTORY_DIR`.
- Automatic `close` / `milestone-close` uses its effective issues directory and
  the existing literal `workshop/history`. Adding an automatic history override
  is out of scope.

For repository root `R`, pinned commits `B` and optional `H`, effective
directories `I` and `A`, and a selected changed path `P`, the structured argv
recipes are exactly:

```text
committed stat:        git -C R diff --stat B H -- :!I/ :!A/
committed names:       git -C R diff --name-status B H -- :!I/ :!A/
committed full:        git -C R diff B H -- :!I/ :!A/
committed targeted:    git -C R diff B H -- P :!I/ :!A/
working-tree stat:     git -C R diff --stat B -- :!I/ :!A/
working-tree names:    git -C R diff --name-status B -- :!I/ :!A/
working-tree full:     git -C R diff B -- :!I/ :!A/
working-tree targeted: git -C R diff B -- P :!I/ :!A/
```

The renderer shell-quotes each argv element only for display; resolution invokes
the runner with an argv slice and no shell. `P` is a displayed substitution slot,
not interpolated executable text. Rooting every recipe with `-C R` makes the
manifest independent of the reviewer's ambient directory while preserving the
former exclusion pathspecs and their order.

### 2026-08-26 — preserve issue-less manual milestone review

**Reason:** `sdlc judge milestone-review` currently permits ad-hoc range review
without `--issue`; requiring an issue would be an unrelated API break.

**Decision:** the automatic close path always requires and names its issue file.
The manual path has two variants: with `--issue N`, it likewise requires the
issue file and names an optional canonical plan; without `--issue`, `IssueRef`
renders as `<unspecified>` and both `IssueFile` and `PlanFile` are explicitly
absent. The repository root, immutable range, exclusions, and inspection
commands remain mandatory in both variants. Manual `--plans-dir` (defaulting via
`WF_PLANS_DIR` / `workshop/plans`) selects plan discovery only when an issue is
present. This preserves ad-hoc review while making the absence of requirements
context visible to the reviewer rather than inventing a path.

## Done when

- `milestone-close` on the first milestone of a fresh branch reviews only the
  branch-point..HEAD diff (no far-back base, no E2BIG).
- The atlas gate on a later milestone sees atlas changes made anywhere in that
  milestone's window (no false "no atlas/ changes" when the edit is a commit back).
- A regression test pins the window base for both the first-milestone and
  Nth-milestone cases.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.30 impl=0.12
item: greenfield-go-module design=0.40 impl=0.32
item: cross-cutting-refactor design=0.20 impl=0.20
item: atlas-docs design=0.04 impl=0.08
item: milestone-review design=0.04 impl=0.20
design-buffer: 0.15
total: 2.05
```

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`
against `baseline-v3.1.md`. Method A only. The calibration source was marked
stale on 2026-08-26, so these v3.1 values are provisional.

## Plan

- [x] Locate the window/BASE_SHA computation in `sdlc milestone-close` + the
      judge dispatch (`sdlc judge milestone-review`) and the atlas/plan gates.
- [x] Anchor the window to the milestone's first commit / prior `Mx` boundary;
      share it across review + atlas + plan gates.
- [x] ~~Pass the review diff to `claude` via temp file / stdin (E2BIG-proof).~~
      Superseded by the compact pinned-manifest revision above.
- [x] Regression tests: first-milestone (branch point) + Nth-milestone (prior
      boundary) window bases.

## Log

### 2026-07-02
- Filed from downstream (pair `#99` launcher port). Current workaround used there:
  run the review manually with the real base
  (`sdlc judge milestone-review --base "$(git merge-base main HEAD)" --head HEAD --issue N`),
  address findings, then `sdlc milestone-close … --no-judge` (put the real verdict
  in the milestone commit's `Review-Verdict:` trailer); for the atlas variant, pass
  the precise `--no-atlas` with the atlas-carrying commit named in `--verified`.
  Recurs every milestone until fixed. See pair `workshop/lessons.md` (the
  "milestone-close's auto review-window" lesson, both manifestations).

### 2026-08-25

- Returned to `open` after a brief claim during pair#146 close recovery. Later
  work already pins prior-boundary/branch-point bases (#58/#77) and a concrete
  reviewed Head (#194); the immediate pair#146 failure was dominated by a raw
  Codex process transcript recursively entering the next diff. That artifact
  contract is now #201, while disposable reviewer isolation is #204. Keep #162
  as the remaining audit/consolidation point for review-window correctness and
  legitimately large prompt transport rather than mixing those fixes into the
  bounded side quest.

### 2026-08-26

- Plan-quality cleared after two rounds. PQ-1 restored the issue's full purpose:
  a first milestone now uses the feature branch point, not the parent of an
  issue commit that may predate unrelated main history (ARCH-PURPOSE). PQ-2
  compressed the executable test contract to named adversarial strategies.
- Added the pure `ReviewWindowManifest` / `ReviewCommand` renderer and a thin
  `boundaryGitRunner` resolution seam with a stateful fake plus temporary real
  Git conformance (ARCH-PURE, ARCH-MOCK). Automatic and manual boundary review
  now share that manifest (ARCH-DRY); only milestone-review changed transport.
- The large-range regression first proves its multi-megabyte sentinel exists in
  the pinned Git diff, then proves the generated prompt is under 100 KB and
  sentinel-free. Explicit-head, working-tree, custom tracker directory,
  issue-less manual review, fail-closed inspection wording, and non-boundary
  prompt compatibility are pinned separately.
- Verification: `go test ./cmd/sdlc/... -count=1` passed; `go test ./... -count=1`
  passed; helptext/process-manual focused suites passed; `git diff --check`
  passed. Rebuilt the live `bin/sdlc` via `make sdlc-build`; a manual
  `judge milestone-review --dry-run` named full immutable SHAs, the issue and
  canonical plan, and four repository-rooted Git recipes with no patch hunk.
- Boundary-review round 1 returned REWORK. BR-1 found that absolute custom
  tracker directories leaked into Git pathspecs; the resolver now canonicalizes
  in-repository paths to Git-relative form and the pure renderer rejects paths
  that bypass that invariant. BR-2 found the live test stopped at ref
  resolution; it now executes all four structured recipes against committed,
  staged, unstaged, untracked, and custom-directory fixture state (ARCH-MOCK,
  ARCH-PURPOSE).
- Boundary-review round 2 disposed BR-1 but proved BR-2's first repair was
  incomplete: `stat` and `targeted` were executed without positive assertions,
  so empty-range mutants passed. Every recipe now asserts its own positive and
  negative semantics. BR-3 added the missing README entry for manual
  `milestone-review` and `--plans-dir`.
- Boundary-review round 3 disposed BR-2 and required BR-3's documentation fix
  to be executable under the gate's claimed-fix rule. A scoped README contract
  now pins the committed-range form, omitted-head working-tree form,
  `--plans-dir`, and both defaults; deleting the section makes it fail.
