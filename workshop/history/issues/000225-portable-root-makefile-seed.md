---
id: 000225
status: done
deps: []
github_issue:
created: 2026-09-13
updated: 2026-09-13
estimate_hours: 1.03
started: 2026-09-13T12:29:20-07:00
actual_hours: 0.10
---

# Publish a portable root Makefile as a safe seed

## Problem

parley.nvim#208 needs a standalone contributor checkout with a real root
Makefile. The inherited `symlink Makefile` declaration replaces a portable
root on the next weave. A leaf seed cannot counteract it safely: current
`applySeed` follows destination symlinks, potentially overwriting an ancestor.
Fresh clones also lack ignored Makefile.workflow, so bootstrap's handoff to
`make bootstrap` cannot work merely by making that include optional.

## Spec

- Keep the generic root Makefile upstream-owned, published via `seed Makefile`.
  Product-specific targets remain in each consumer's Makefile.local.
- Make workflow inclusion optional for a public checkout, with a sibling
  ariadne overlay fallback after bootstrap has cloned peers. Normal product
  targets and help work without maintainer tools. Existing maintainer targets
  remain available after bootstrap, even when all local helper links are absent.
- Materializing a seed unlinks a destination symlink before writing or chmod,
  including identical-content links; never change its old target's bytes/mode.
  Reuse the safe regular-file materialization pattern already in applyWriteFile.
- Preserve generic CI ownership too: its seeded merge-check workflow discovers
  the bootstrapped upstream runner when the local helper link is absent and
  invokes an optional executable repo-owned scripts/ci-setup.sh before checks.
  Consumer-specific tool provisioning lives in that hook; a repeated weave
  must not erase the effective CI setup. Parley needs Go/CUE/vocabulary for
  its generated-runtime-data drift check.
- No local restoration script or copy-after-weave workaround. Migration is
  idempotent for linked and already-seeded consumers.

## Done when

- Scratch leaf with real Makefile.local has working local targets and help
  without a sibling; bootstrap discovers its overlay after peers are present.
- Existing linked consumer becomes a real seeded Makefile; ancestor bytes and
  permissions remain unchanged. Repeated weave leaves content unchanged.
- Regression cases cover matching, differing, and dangling destination symlinks
  and unavailable/failed materialization paths.
- A parley-like scratch consumer survives maintainer refresh and still runs its
  product tests. Source/manifest/atlas documentation describe seed ownership.

## Plan

- [x] Design safe seed materialization and portable generic Makefile behavior.
- [x] Add regressions, implement the seed migration and bootstrap path.
- [x] Validate representative consumers; submit measured evidence to close review.

## Log

### 2026-09-13
- 2026-09-13: closed — Full weave/Make/CI/bootstrap verification passes. Narrow review policy repair c3c46d39 preserves behavior regressions and allows pinned prose evidence; rendered contract red-green, full judge suite, targeted sdlc Boundary/CloseReview/ReviewWindow tests pass. Owner binary vcs.revision=c3c46d39 vcs.modified=false. Full-window diff check clean.; review verdict: SHIP

Discovered during fresh-eyes review of parley.nvim#208's deployment plan.
This is a prerequisite for its maintainer-link cleanup, not a runtime package
dependency. No implementation or estimate yet; plan approval comes first.

### 2026-09-13 — implementation and verification

User approved the prerequisite as part of parley.nvim#208. Plan-quality accepted
round 2; estimate-quality informative pass, change-code created this branch.
Seed and Make/CI regressions were observed failing before implementation.
Safe materialization now shares the fail-closed destination-link guard.

`go test ./cmd/weave/... -count=1` passed. Portable Make fixture runs the real
Make recipes with persisted tool state and actual weave twice; portable CI
fixture executes actual workflow run blocks with the real generic check runner.
Both passed. ensure-go (3), bootstrap-transitive (12), peer-update (6) passed.
`git diff --check` clean. No live peer weave or propagation performed.

A separate actual bootstrap probe used an archived upstream plus current edits
in `/tmp/ariadne225-live.w94sow`: real `./bootstrap.sh` built tools, materialized
99 consumer actions, installed PATH only in isolated HOME, passed product target,
and repeated real weave with a regular root matching upstream. Initial probe
surfaced the implicit `.sh` bootstrap recipe; added failing regression and phony
classification, then reran successfully without the stray file. ARCH-ORDER and
ARCH-PURPOSE shaped the ordered bootstrap and missing-helper conformance.

## Revisions

### 2026-09-13 — CI ownership sweep

Reason: the generic merge-check workflow is also upstream-seeded. Delta: add
runner discovery and a repo-owned setup hook to the portable-maintainer boundary
so the consumer does not fork a workflow that weave would overwrite.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. Calibration flagged stale by estimate-source,
so provisional. Existing OSFS and Make infrastructure eliminate library discovery.
Thorough plan discounts design by 0.2; implementation uses v3.1's 0.4 scaling.
Smaller Go extension (0.2/0.5), two cross-cutting surfaces (each 0.3/0.5),
docs (0.1/0.1), and one review (0.1/0.4), before those multipliers.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.04 impl=0.20
item: cross-cutting-refactor design=0.06 impl=0.20
item: cross-cutting-refactor design=0.06 impl=0.20
item: atlas-docs design=0.02 impl=0.04
item: milestone-review design=0.02 impl=0.16
design-buffer: 0.15
total: 1.03
```

### 2026-09-13 — boundary round 1 repairs

Review returned REWORK: BR-1 incorrectly classified filesystem wildcard discovery
as pure; BR-2 lacked README guidance for the new consumer setup surface. Corrected
the classification across the plan and added README ownership/bootstrap/hook
instructions. No runtime defect identified; code evidence remains unchanged.

### 2026-09-13 — documentation-only disposition evidence

Boundary round 2 explicitly confirms both original defects are corrected and
finds no runtime correctness defect. Its remaining objection is that deleting
README or restoring the plan's old label would leave runtime tests green.
That is expected: these two changes document already-tested behavior and do not
change executable semantics. A string-presence assertion would test the wording
against itself, not bootstrap safety. Session developer instructions explicitly
forbid tests for reversible low-impact changes or tests mirroring implementation.
Accordingly no vacuous prose test is added. BR-1 is addressed by the corrected
integration classification and revision; BR-2 by README's public usage section.
The fresh reviewer independently verified both corrections at pinned 5cb5857c.
Existing meaningful filesystem, Make and CI tests remain the behavior evidence.
Removed the touched README line's inherited trailing whitespace; verified with
`git diff --check HEAD` and `git diff --check fa8746e` after this edit.

### 2026-09-13 — side-quest: repair the injected evidence contract

Round 3 again verified BR-1/BR-2 corrected and found no runtime defect, but the
unconditional injected regression rule prevented disposition. Parent authorized
only the narrow prompt repair. The rendered MilestoneReview contract now retains
fails-without-fix for executable changes, explicitly including prompts/config,
and requires pinned-diff/source inspection for prose-only corrections. The new
rendered-contract regression failed on all missing policy clauses before the
repair; the existing golden was intentionally edited only for this changed
contract. This fixes the real gate rather than bypassing judge/ledger or adding
vacuous documentation tests. Owner binary is rebuilt from this source for reclose.
