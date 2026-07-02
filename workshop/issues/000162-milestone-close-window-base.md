---
id: 000162
status: open
deps: []
github_issue:
created: 2026-07-02
updated: 2026-07-02
estimate_hours:
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
