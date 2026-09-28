---
id: '000077'
status: done
created: 2026-06-03
updated: 2026-06-03
estimate_hours: 1.0
actual_hours: 1.0
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
