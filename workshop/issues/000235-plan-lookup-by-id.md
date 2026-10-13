---
id: 000235
status: codecomplete
deps: []
github_issue:
created: 2026-09-18
updated: 2026-10-12
estimate_hours: 1.4
card_mirror: '0d7079ac40a01dfe8d66d4cf19ad0bc859fe414e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T17:21:51-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 0.19
---

# Durable-plan lookup is exact-name: a plan under another slug is invisible

## Problem

sdlc finds an issue's durable plan by ONE exact name: the issue file's stem plus
`-plan.md` (`readOptionalPlanFile`, `cmd/sdlc/changecode.go`). A plan saved under
any other slug for the same issue id is invisible to it, silently.

Found in pair#283 (Log, 2026-09-18). The plan was saved as
`000283-default-cursor-style-plan.md`, while the issue file is
`000283-default-cursor-style-overrides-terminal.md`. Two consequences:

- **#231's flow inference said quick** although a durable plan existed. Renaming
  the file fixed it.
- **Plan-quality never saw the plan.** This predates #231: a misnamed plan is
  skipped by the plan gate too, and so the plan the agent wrote goes unreviewed.

The close-time review manifest resolves the plan the same way
(`reviewPlanPaths` / `optionalReviewPlanPath`, `cmd/sdlc/reviewwindow.go`), so
the boundary reviewer misses a misnamed plan as well.

## Spec

Find the plan by issue ID, in one place that every reader uses:

- A plan is `<plansDir>/NNNNNN-*-plan.md` for the issue's zero-padded id. The
  exact `-plan.md` suffix excludes the sidecars (`-plan-gate.md`,
  `-close-gate.md`, `-*-review.md`) by construction.
- Exactly one match is the plan, whatever its slug. A match that isn't at the
  canonical name also prints a one-line note naming the canonical name.
- Two or more matches are ambiguous: refuse, naming them, rather than pick one.
- No match means no plan, as today.

Route `readOptionalPlanFile` (change-code: flow inference and plan-quality) and
the review manifest's plan resolution through it. `familyFiles`
(`cmd/sdlc/resolve.go`) already globs an issue's family by id, so reuse it rather
than write a third lookup (ARCH-DRY).

## Done when

- A plan named `NNNNNN-<any-slug>-plan.md` is found by change-code (the flow
  inference and plan-quality) and by the close review manifest. The canonical
  name still works.
- Two candidate plans for one issue id refuse, naming both.
- A sidecar (`-plan-gate.md`, `-close-gate.md`, `-*-review.md`) is never taken
  for the plan.
- A test reproduces pair#283's names: an off-slug plan infers full and reaches
  plan-quality.

## Plan

Design: `planArtifactPath(plansDir, name)` (`cmd/sdlc/changecode.go`) is already
the one path every reader goes through (#189). It becomes the id lookup and
returns `(path, error)`. Its IO is a glob of `plansDir` with the family pattern
`familyFiles` uses, now one shared `familyPattern(id)` (ARCH-DRY). The plan
classification is `classifyFamily`'s `-plan.md` rule, so sidecars are excluded
by the same rule `sdlc resolve` uses. The decision is a pure
`pickPlanArtifact(canonical, plans)` (ARCH-PURE): none gives the canonical path,
which readers treat as absent; one gives that plan; two or more refuse, naming
them. Readers to route (the full enumeration, ARCH-PURPOSE):
change-code's plan read and its design checkpoint, `recordChangeCodeFlow`,
close's quick-flow measure (`readOptionalPlanFile`), `judge plan-quality
--dry-run`, and the review manifest plus close review snapshot
(`reviewPlanPaths`, which globs under the repository root). The off-slug note is
printed where the plan is first read: change-code step 2 and `judge plan-quality`.

The id comes from the stem's `NNNNNN-` prefix (`issueIDFromPath`). A stem
without one keeps today's exact canonical name. The lookup returns an error, and
every reader propagates it: `readOptionalPlanFile` becomes `(string, error)`, and
`recordChangeCodeFlow` and close's measure die on it. So an ambiguous plan can
never quietly read as "no plan".

Tests: the unit-tested core is `pickPlanArtifact`, fed through
`planArtifactPath` over sets of `NNNNNN-*` names: sidecars (`-plan-gate.md`,
`-close-gate.md`, `-m1-review.md`, `-close-review.md`), a longer id sharing the
prefix, and two plans. The guard is that the result is never a non-`-plan.md`
file and never a silent pick between two plans. The integration tests are
pair#283's names through `runChangeCode --dry-run`, which shows full inferred
and the plan text in the plan-quality prompt, and the review manifest finding an
off-slug plan.

Non-goals: no lookup in the archive or history, and no automatic rename of an
off-slug plan; the note names the canonical name instead. Docs:
`helptext/change-code.md` stops saying the plan is found only at
`<issue>-plan.md`.

ARCH-MOCK/CONSTRAINTS/SECURE/ORDER: N/A. This is a local directory glob with no
external dependency, a few files per issue, and no state held between events.
ARCH-FUNERAL: N/A, since it creates nothing durable.

- [x] One id-based plan lookup; route change-code's plan read and the review
      manifest through it
- [x] Tests: the canonical name, pair#283's off-slug name, two candidates
      refused, sidecars ignored

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module   design=0.2 impl=0.2
item: issue-spec          design=0.5 impl=0.04
item: milestone-review    design=0.0 impl=0.14
item: atlas-docs          design=0.05 impl=0.04
design-buffer: 0.30
total: 1.395
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.* It extends one lookup and reroutes six call
sites, which matches the well-specced extend row (impl 0.5h × 0.4). The design
buffer stays at 30% because there is no separate plan doc.

## Log

### 2026-09-18

Filed from pair#283 at the operator's request. It is meant as an example of the
small bug the quick flow (#231) should accelerate: a few lines in one or two
files, with a crisp Done-when and no durable plan.

One conflict to resolve before it can be one: in ariadne, #231 declared sdlc's
whole build closure (`cmd/sdlc/`, `pkg/`, go.mod/sum) a shared surface. sdlc
runs from the branch under review, so a quick branch could otherwise change the
gate it closes under. As a result every sdlc fix in ariadne, this one included,
is upgraded to full at close. Whether that declaration should be narrower is
the operator's call, and #231's trial is the place to decide it.

Decided the same day: the operator keeps the full-closure declaration ("not worth
additional rule/risk"). This issue takes the full flow in ariadne. It remains a
fair example of the class the quick flow targets, just not a candidate for it here.

### 2026-10-12
- 2026-10-12: closed — make test green (12 shards, 1124 cmd/sdlc tests, 0 failed); planlookup_test.go covers canonical, pair#283 off-slug, sidecars, two-plan refusal, review manifest, and change-code end to end (full inferred, plan in plan-quality prompt and judge dry-run, committed by checkpoint); checkpoint assertion mutation-checked; review verdict: SHIP

- Dispatched by the TL (ops:0). Claimed and started on ariadne:2. Pinned
  `--flow full` per the operator's 2026-09-18 decision above.
- plan-quality round 1: INFO with 4 Minors (id derivation, the refusal path
  through readers that can't refuse, test strategy, non-goals and help). All 4
  are folded into the Plan above.
- Implemented: `planArtifactPath` → `(path, error)` over `planArtifactPathIn`,
  `pickPlanArtifact` and `planNameNote` (`cmd/sdlc/changecode.go`), plus
  `familyPattern` shared with `familyFiles`. All six readers are rerouted. Gotcha:
  `issueIDFromPath` parses a *filename*, so the stem needs `.md` appended first.
- Tests (`cmd/sdlc/planlookup_test.go`): the id table (canonical, pair#283's
  off-slug, none, sidecars only, off-slug among sidecars, two refuse, a stem
  without an id); `readOptionalPlanFile` refusing ambiguity; the review manifest
  finding an off-slug plan and refusing two; end to end, a 600-line off-slug
  plan infers full, reaches the plan-quality prompt and `judge plan-quality
  --dry-run`, and is committed by the design checkpoint. A mutation check
  confirmed the checkpoint assertion catches an exact-name revert.
- Close round 1: SHIP with 3 Minors, all fixed in the same round. The canonical
  name is now one `canonicalPlanPath`. The lookup lists the directory with
  `ReadDir` and matches names against `familyPattern`, so metacharacters in the
  repository path are no longer wildcards (tested with `repo[1]*?`). And
  `TestReviewPlanPathsRefusesTwoPlans` covers the seam close's review snapshot
  reads through.

