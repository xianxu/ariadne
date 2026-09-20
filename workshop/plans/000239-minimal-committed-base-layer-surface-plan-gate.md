---
gate: plan-quality
issue: 239
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-19T20:38:59-07:00"
      agent: claude
      blocked: true
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-09-19T20:46:16-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Critical
          title: M4's fleet sweep would git rm --cached 1160 repo-owned tracked files in kbench
          detail: |-
            commitConsumption (cmd/sdlc/propagatebase.go:248) untracks everything
            `git ls-files -i -c --exclude-standard` reports — the whole ignore config,
            including nested and repo-owned blanket patterns — not the set weave just
            produced. kbench returns 1160 tracked-but-ignored files under
            competition/arc-agi-3/runs/, matched by its own nested
            competition/arc-agi-3/.gitignore:24 (runs/20*/). One
            `sdlc propagate-base --repo kbench` rm-caches and commits all of them. The
            plan's "per-path derivation makes the untrack set structurally safe" is a
            property of the managed BLOCK, not of the sweep. Scope the untrack to the
            planned action set in Task 4.0, or refuse when the set contains a path the
            plan did not produce.
          family: untrack-scope-exceeds-weave-surface
          round: 2
        - id: PQ-2
          severity: Critical
          title: Untracking scripts/merge-checks.d/* silently voids CI checks in astro, parli, tools
          detail: |-
            CI never runs weave — merge-check.yml:38 uses BOOTSTRAP_CLONE_ONLY=1 and
            bootstrap.sh:34-36 skips the `make bootstrap` handoff — so merge-checks.d
            holds only committed files. The runner has an owner fallback
            (merge-check.yml:71); the checks DIR has none (run-merge-checks.sh:26-27
            reads $ROOT/scripts/merge-checks.d only). astro, parli and tools track
            base.manifest:138's symlink 40-duplicate-issue-id.sh as mode 120000; after
            the sweep astro and parli have no checks at all and run-merge-checks.sh:42
            prints "none defined — pass (no-op)" and exits 0. Done-when's "CI still
            passes on a derivative PR" cannot detect a vacuous pass, and the pilot
            (pair) tracks only .gitkeep there, so it is blind by construction. Write
            the pre-weave-consumer enumeration into the plan and resolve this member.
          family: pre-weave-consumer-fallback
          round: 2
        - id: PQ-3
          severity: Minor
          title: Done-when 7 contradicts the ownership rule the plan adopted
          detail: |-
            "Only bootstrap.sh, merge-check.yml and construct/deps remain committed
            from the weave surface" is false under the plan's own rule — Makefile
            (seed-once), workshop/lessons.md (touch) and the scaffold .gitkeeps also
            stay tracked, deliberately. Task 4.4 Step 2 walks the ten criteria
            literally, so restate criterion 7 in ownership terms via a ## Revisions
            entry.
          family: done-when-restates-superseded-framing
          round: 2
        - id: PQ-4
          severity: Minor
          title: IgnoreEntries panics on an unclassified Action though its caller returns error
          detail: |-
            planActions can return an error, so an unhandled Action type can surface as
            a diagnosable weave failure rather than a stack trace in every repo's
            `make weave`. The unit test asserting the guard works either way.
          family: pure-core-aborts-instead-of-erroring
          round: 2
        - id: PQ-5
          severity: Minor
          title: Task 4.2's "commit or discard" on pair/Makefile picks the wrong outcome half the time
          detail: |-
            pair/Makefile is 120000 in the index and a regular file on disk. Committing
            it makes applySeedOnce see presence and no-op forever, freezing a copy of
            ariadne's own root Makefile as pair's front door — the two-owners artifact
            this issue removes. Discarding restores the symlink so M1 materializes
            construct/Makefile.seed. Say discard.
          family: ambiguous-operator-instruction
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-09-19T20:51:51-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Task 4.0a filters by check-ignore -v provenance; re-measured kbench (1160, nested source) and parley.nvim (1, block source) — the filter separates them correctly.
          round: 3
        - id: PQ-2
          disposition: addressed
          note: Pre-weave-consumer class enumerated; Task 4.0b adds the owner fallback to run-merge-checks.sh:26-27, with a test keyed on the "none defined" vacuous-pass message.
          round: 3
        - id: PQ-3
          disposition: addressed
          note: Done-when 7 now states the re-derived/provisioned split instead of enumerating three paths; recorded in the issue's Revisions.
          round: 3
        - id: PQ-4
          disposition: addressed
          note: IgnoreEntries returns an error naming the type; only the risks-table row still says "panics" — stale prose, fix in passing.
          round: 3
        - id: PQ-5
          disposition: addressed
          note: Task 4.2 Step 1 says discard, with the freeze-ariadne's-Makefile reason stated.
          round: 3
      blocked: false
    - "n": 4
      timestamp: "2026-09-19T20:57:02-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Task 4.0a adds the check-ignore -v provenance filter; kbench's 1160 nested-gitignore files stay tracked.
          round: 4
        - id: PQ-2
          disposition: addressed
          note: Pre-weave consumers table enumerates the class; Task 4.0b adds the one missing owner fallback.
          round: 4
        - id: PQ-3
          disposition: addressed
          note: Issue Revisions restates Done-when 7 as the re-derived/provisioned split.
          round: 4
        - id: PQ-4
          disposition: addressed
          note: IgnoreEntries returns an error naming the type; planActions propagates it.
          round: 4
        - id: PQ-5
          disposition: addressed
          note: Task 4.2 Step 1 now says discard, with the two-owners reason stated.
          round: 4
      findings:
        - id: PQ-6
          severity: Critical
          title: propagate-base --dry-run returns before any untrack logic, so all three pre-sweep safety checks prove nothing
          detail: |-
            runPropagateBase prints the dependent list and returns at propagatebase.go:135 —
            it never weaves and never reaches commitConsumption, so it cannot name a path it
            would untrack. Task 4.0a Step 4 ("must report 0 paths to untrack, not 1160"),
            Task 4.0c Step 4, and Task 4.2 Step 1 all read that empty output as a pass and
            proceed into the irreversible sweep with the provenance filter never exercised
            end-to-end. The class is those three sites plus 4.0a Step 4's kbench git status,
            and the single fix is to make Task 4.0a extend --dry-run to run the classify pass
            and print, per repo, candidates split into would-untrack (inside the managed
            block's line range) versus left-tracked (repo-owned provenance), mutating nothing.
          family: verification-cannot-fail
          round: 4
        - id: PQ-7
          severity: Important
          title: Task 4.0b's checks-dir fallback collects the owner's whole merge-checks.d, importing ariadne-local checks into every derivative
          detail: |-
            base.manifest:133-137 states that scaffold scripts/merge-checks.d plus the single
            symlink row for 40-duplicate-issue-id.sh IS the propagation selection (#213).
            Resolving the fallback to ../ariadne/scripts/merge-checks.d discards it: ariadne's
            dir also holds 30-weave-drift.sh and, after Task 3.3 Step 3 which deliberately
            keeps it ariadne-local, 50-base-layer-tests.sh — so every derivative PR would run
            ariadne's portable-makefile and gitignore-surface conformance tests, which go build
            ./cmd/weave against ariadne's sources. State the rule in 4.0b: the fallback resolves
            only checks the leaf's manifest walk produces a Symlink for; the rest stay the owner's.
          family: owner-fallback-widens-propagation-set
          round: 4
      blocked: true
    - "n": 5
      timestamp: "2026-09-20T13:46:12-07:00"
      agent: codex
      dispose:
        - id: PQ-6
          disposition: withdrawn
          note: 'Overtaken by the restart: the active plan removes the dry-run-gated fleet sweep, requires scoped migration regressions and disposable pilots, and assigns actual rollout to #241.'
          round: 5
        - id: PQ-7
          disposition: withdrawn
          note: 'Overtaken by compile-before-consumption CI: the active plan does not introduce an owner-directory fallback, so manifest selection continues to govern inherited checks.'
          round: 5
      blocked: false
content_hash: e4a3ba5c0f520a409aca05d21dc6812c3a2605b5e27ad6eb5df221c2f5258224
---

# Gate ledger — ariadne#239 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-19T20:38:59-07:00 (claude) — BLOCKED

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-09-19T20:46:16-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Critical] `untrack-scope-exceeds-weave-surface` M4's fleet sweep would git rm --cached 1160 repo-owned tracked files in kbench
  commitConsumption (cmd/sdlc/propagatebase.go:248) untracks everything
  `git ls-files -i -c --exclude-standard` reports — the whole ignore config,
  including nested and repo-owned blanket patterns — not the set weave just
  produced. kbench returns 1160 tracked-but-ignored files under
  competition/arc-agi-3/runs/, matched by its own nested
  competition/arc-agi-3/.gitignore:24 (runs/20*/). One
  `sdlc propagate-base --repo kbench` rm-caches and commits all of them. The
  plan's "per-path derivation makes the untrack set structurally safe" is a
  property of the managed BLOCK, not of the sweep. Scope the untrack to the
  planned action set in Task 4.0, or refuse when the set contains a path the
  plan did not produce.
- **PQ-2** [Critical] `pre-weave-consumer-fallback` Untracking scripts/merge-checks.d/* silently voids CI checks in astro, parli, tools
  CI never runs weave — merge-check.yml:38 uses BOOTSTRAP_CLONE_ONLY=1 and
  bootstrap.sh:34-36 skips the `make bootstrap` handoff — so merge-checks.d
  holds only committed files. The runner has an owner fallback
  (merge-check.yml:71); the checks DIR has none (run-merge-checks.sh:26-27
  reads $ROOT/scripts/merge-checks.d only). astro, parli and tools track
  base.manifest:138's symlink 40-duplicate-issue-id.sh as mode 120000; after
  the sweep astro and parli have no checks at all and run-merge-checks.sh:42
  prints "none defined — pass (no-op)" and exits 0. Done-when's "CI still
  passes on a derivative PR" cannot detect a vacuous pass, and the pilot
  (pair) tracks only .gitkeep there, so it is blind by construction. Write
  the pre-weave-consumer enumeration into the plan and resolve this member.
- **PQ-3** [Minor] `done-when-restates-superseded-framing` Done-when 7 contradicts the ownership rule the plan adopted
  "Only bootstrap.sh, merge-check.yml and construct/deps remain committed
  from the weave surface" is false under the plan's own rule — Makefile
  (seed-once), workshop/lessons.md (touch) and the scaffold .gitkeeps also
  stay tracked, deliberately. Task 4.4 Step 2 walks the ten criteria
  literally, so restate criterion 7 in ownership terms via a ## Revisions
  entry.
- **PQ-4** [Minor] `pure-core-aborts-instead-of-erroring` IgnoreEntries panics on an unclassified Action though its caller returns error
  planActions can return an error, so an unhandled Action type can surface as
  a diagnosable weave failure rather than a stack trace in every repo's
  `make weave`. The unit test asserting the guard works either way.
- **PQ-5** [Minor] `ambiguous-operator-instruction` Task 4.2's "commit or discard" on pair/Makefile picks the wrong outcome half the time
  pair/Makefile is 120000 in the index and a regular file on disk. Committing
  it makes applySeedOnce see presence and no-op forever, freezing a copy of
  ariadne's own root Makefile as pair's front door — the two-owners artifact
  this issue removes. Discarding restores the symlink so M1 materializes
  construct/Makefile.seed. Say discard.

## Round 3 — 2026-09-19T20:51:51-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Task 4.0a filters by check-ignore -v provenance; re-measured kbench (1160, nested source) and parley.nvim (1, block source) — the filter separates them correctly.
- PQ-2 — addressed — Pre-weave-consumer class enumerated; Task 4.0b adds the owner fallback to run-merge-checks.sh:26-27, with a test keyed on the "none defined" vacuous-pass message.
- PQ-3 — addressed — Done-when 7 now states the re-derived/provisioned split instead of enumerating three paths; recorded in the issue's Revisions.
- PQ-4 — addressed — IgnoreEntries returns an error naming the type; only the risks-table row still says "panics" — stale prose, fix in passing.
- PQ-5 — addressed — Task 4.2 Step 1 says discard, with the freeze-ariadne's-Makefile reason stated.

## Round 4 — 2026-09-19T20:57:02-07:00 (claude) — BLOCKED

### Disposed

- PQ-1 — addressed — Task 4.0a adds the check-ignore -v provenance filter; kbench's 1160 nested-gitignore files stay tracked.
- PQ-2 — addressed — Pre-weave consumers table enumerates the class; Task 4.0b adds the one missing owner fallback.
- PQ-3 — addressed — Issue Revisions restates Done-when 7 as the re-derived/provisioned split.
- PQ-4 — addressed — IgnoreEntries returns an error naming the type; planActions propagates it.
- PQ-5 — addressed — Task 4.2 Step 1 now says discard, with the two-owners reason stated.

### Raised

- **PQ-6** [Critical] `verification-cannot-fail` propagate-base --dry-run returns before any untrack logic, so all three pre-sweep safety checks prove nothing
  runPropagateBase prints the dependent list and returns at propagatebase.go:135 —
  it never weaves and never reaches commitConsumption, so it cannot name a path it
  would untrack. Task 4.0a Step 4 ("must report 0 paths to untrack, not 1160"),
  Task 4.0c Step 4, and Task 4.2 Step 1 all read that empty output as a pass and
  proceed into the irreversible sweep with the provenance filter never exercised
  end-to-end. The class is those three sites plus 4.0a Step 4's kbench git status,
  and the single fix is to make Task 4.0a extend --dry-run to run the classify pass
  and print, per repo, candidates split into would-untrack (inside the managed
  block's line range) versus left-tracked (repo-owned provenance), mutating nothing.
- **PQ-7** [Important] `owner-fallback-widens-propagation-set` Task 4.0b's checks-dir fallback collects the owner's whole merge-checks.d, importing ariadne-local checks into every derivative
  base.manifest:133-137 states that scaffold scripts/merge-checks.d plus the single
  symlink row for 40-duplicate-issue-id.sh IS the propagation selection (#213).
  Resolving the fallback to ../ariadne/scripts/merge-checks.d discards it: ariadne's
  dir also holds 30-weave-drift.sh and, after Task 3.3 Step 3 which deliberately
  keeps it ariadne-local, 50-base-layer-tests.sh — so every derivative PR would run
  ariadne's portable-makefile and gitignore-surface conformance tests, which go build
  ./cmd/weave against ariadne's sources. State the rule in 4.0b: the fallback resolves
  only checks the leaf's manifest walk produces a Symlink for; the rest stay the owner's.

## Round 5 — 2026-09-20T13:46:12-07:00 (codex) — passed

### Disposed

- PQ-6 — withdrawn — Overtaken by the restart: the active plan removes the dry-run-gated fleet sweep, requires scoped migration regressions and disposable pilots, and assigns actual rollout to #241.
- PQ-7 — withdrawn — Overtaken by compile-before-consumption CI: the active plan does not introduce an owner-directory fallback, so manifest selection continues to govern inherited checks.

## Open findings

(none — every finding has been disposed)
