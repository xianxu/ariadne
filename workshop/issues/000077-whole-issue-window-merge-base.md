---
id: 000077
status: open
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-03
estimate_hours:
---

# whole-issue close window: base on merge-base(main) not first #N commit

## Problem

`boundaryWindowBase` (added in #58) bases a **whole-issue** close window on
`firstCommitReferencing("#N")^` — the parent of the *first* commit whose subject
contains `#N`. For an issue filed early and implemented much later, that first
`#N` commit is the **"#N: file issue"** commit, not the first *implementation*
commit. So the end-of-issue boundary review window spans everything since the
issue was filed — including unrelated, already-reviewed-and-merged work.

Surfaced concretely in #58's own boundary review: the whole-issue window
resolved to `6be88ca^..HEAD` = **147 commits across ~12 unrelated issues**, when
the actual #58 deliverable was a single commit. The review noted it reviewed the
right commit by inspection, but the *window* was diluted. This is "over-cover,
not under-cover" (the safe direction, same as #58's missing-trailer fallback), so
it's not a correctness bug — but it makes end-of-issue reviews noisy and #58
entrenched the first-`#N`-commit base as the single window source.

## Spec

Base the **whole-issue** close window (milestone == "") on a tighter anchor:
- `merge-base(main, HEAD)` — the branch point — so the window is exactly the
  work this branch added; OR
- the most-recent prior `Review-Verdict:` boundary on the branch (the same
  `previousReviewBoundary` the milestone path uses), whichever is tighter.

Milestone windows already base on the prior boundary (#58) and are unaffected.
The change is localized to `boundaryWindowBase`'s `milestone == ""` branch.

Open question to resolve in design: working directly on `main` (no feature
branch) makes `merge-base(main, HEAD)` == HEAD, which would empty the window —
need a fallback (first-`#N`-commit, or prior boundary) for the on-main case.
This is why #58 used first-`#N`-commit rather than merge-base in the first place.

## Done when

- A whole-issue close on a branch filed-early/implemented-late windows only the
  branch's own commits, not unrelated merged history.
- The on-`main` (no branch) case still resolves a sane non-empty window.
- Milestone windows unchanged; `go test ./cmd/sdlc/...` green.

## Plan

- [ ]

## Log

### 2026-06-03
Filed from #58's boundary review (Important, non-blocking) — the whole-issue
window over-captures when issue-file and implementation are far apart in history.
Pre-existing behavior #58 entrenched as the single window source. See [[000058]]
(`boundaryWindowBase`, `milestone == ""` branch).
