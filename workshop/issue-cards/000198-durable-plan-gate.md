---
id: 000198
status: open
created: 2026-08-20
updated: 2026-08-20
estimate_hours:
github_issue:
---

# the durable plan in workshop/plans/ has no close gate, so it silently drifts from the code it specifies

## Problem

`sdlc close` and `sdlc milestone-close` enforce a plan-unchecked gate — an issue cannot
close with unticked `## Plan` items (`cmd/sdlc/close.go`, `--no-plan-check` to waive).
That gate reads **only the issue file**. The durable plan at
`workshop/plans/NNNNNN-slug-plan.md` — the artifact AGENTS.md §1 calls the record of
truth for non-trivial work, and the one a milestone reviewer is told to cross-check
entities against — has no gate at all.

So it drifts, silently, and the drift is invisible at exactly the moment it matters: the
close gate reports the issue's Plan is fully ticked while the durable plan still says the
work is unstarted and its Core-concepts table names files that no longer exist.

### Observed on ariadne#194

Both instances were caught by boundary reviewers, not by any gate:

- **M3 review, BR-25:** Tasks 2.3 and 3.1–3.6 and every `## Verification` box still `- [ ]`
  after M2 and M3 had landed, while the issue's `## Plan` correctly ticked M3 as done.
- **M3 re-review, BR-32:** the Core-concepts table put `FamilyCounts`, `NormalizeFamily`
  and `ConvergenceLine` in `ledger.go`/`prompt.go` when all three are in `family.go`;
  named `normalizeFamily` when the symbol is exported; carried two stale line numbers; and
  omitted `RenderPriorFindingsScoped` entirely — a newly exported API that downstream
  gates will consume.

The second instance is the sharper one: **the table drifted in the same commit that ticked
the boxes.** Ticking is not verification, and nothing checked the rows against the code.

That table is not decoration. `code-review.md`'s "Core concepts cross-check" instructs the
boundary reviewer to grep each row against the diff and flag PURE entities whose tests need
mocks. A table with five wrong rows sends that reviewer looking in the wrong files, so the
artifact meant to *aid* review actively degrades it.
