---
id: 000162
status: working
deps: []
github_issue:
created: 2026-07-02
updated: 2026-08-26
estimate_hours:
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

## Done when

- `milestone-close` on the first milestone of a fresh branch reviews only the
  branch-point..HEAD diff (no far-back base, no E2BIG).
- The atlas gate on a later milestone sees atlas changes made anywhere in that
  milestone's window (no false "no atlas/ changes" when the edit is a commit back).
- A regression test pins the window base for both the first-milestone and
  Nth-milestone cases.

## Plan

- [ ] Locate the window/BASE_SHA computation in `sdlc milestone-close` + the
      judge dispatch (`sdlc judge milestone-review`) and the atlas/plan gates.
- [ ] Anchor the window to the milestone's first commit / prior `Mx` boundary;
      share it across review + atlas + plan gates.
- [ ] Pass the review diff to `claude` via temp file / stdin (E2BIG-proof).
- [ ] Regression tests: first-milestone (branch point) + Nth-milestone (prior
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
