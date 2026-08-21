---
id: 000198
status: open
deps: [ariadne#194]
github_issue:
created: 2026-08-20
updated: 2026-08-20
estimate_hours:
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

## Spec

To be designed, but the shape is constrained by what already exists:

- The plan-unchecked gate should cover the durable plan when one exists, not only the
  issue's `## Plan`. Resolution is already solved — `sdlc resolve --kind plan` and the
  `#144` stem convention find it.
- The Core-concepts table is machine-checkable in a way prose is not: every row names a
  path and a symbol. A deterministic check ("does `path` exist, and does it contain
  `symbol`?") would have caught all five #194 rows without an LLM. That is a *form* gate
  in the AGENTS.md §5 sense, and it belongs beside the existing structural checks rather
  than in the boundary review's judgement.
- Line numbers in the table are drift-generating by construction. Consider dropping them
  from the convention, or accepting them as non-checked.
- Needs a `--no-<gate>` flag per AGENTS.md §5, and a `Grammar`/`AckPat`/`RefusalPat` row in
  `processmanual.GateCatalog` so its friction is attributable (#172).

## Done when

- [ ] Closing with an unticked durable plan refuses, or is explicitly waived.
- [ ] A Core-concepts row whose file or symbol does not exist refuses, naming the row.
- [ ] The check is deterministic — no LLM — and runs at the same boundary as the existing
      plan-unchecked gate.
- [ ] A test covers the #194 shape directly: boxes ticked in the issue, durable plan
      untouched, table rows pointing at files the symbols left.
- [ ] The gate has its own `--no-<gate>` flag and a GateCatalog row.

## Plan

- [ ] Design via `sdlc start-plan` before implementing.

## Log

### 2026-08-20

Filed from ariadne#194's M3 boundary review (BR-32), which declined to accept a row-by-row
correction and asked for the rule instead: *"Do NOT stop at correcting the rows… the
durable plan in workshop/plans/ has no gate, which is why it lags."*

Worth noting the diagnosis is structural rather than a lapse: the durable plan is the only
major SDLC artifact with no automated check on it. The issue file has the plan-unchecked
gate and instance conformance, `atlas/` has the atlas gate, the ledgers have their own
parsers, projects have the detail-block requirement. The plan has the boundary reviewer's
attention and nothing else.
