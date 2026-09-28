---
gate: boundary-review
issue: 253
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-28T13:19:23-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: move-detail help text says "escape patch" (means escape hatch) and the sentence is ungrammatical
          detail: cmd/sdlc/issuemovedetail.go:38, side-quest commit f3e7c907.
          family: help-text-accuracy
          round: 1
        - id: BR-2
          severity: Minor
          title: scripts/test-shard.py committed 100644 though its usage line invokes it directly
          family: script-invocation-contract
          round: 1
        - id: BR-3
          severity: Minor
          title: test-shard TEST_NAME excludes Example functions, so future cmd/sdlc examples would silently not run under make test
          family: shard-test-selection-completeness
          round: 1
        - id: BR-4
          severity: Minor
          title: '"new git-driving package needs PreferRealGit TestMain" is enforced only by atlas prose, not a source guard'
          family: guidance-without-enforcement
          round: 1
      recipe: milestone-review
      blocked: false
    - "n": 2
      timestamp: "2026-09-28T13:21:25-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: cmd/sdlc/issuemovedetail.go:38 still reads "Consult operator before move, this is an escape patch" at HEAD 3f11b26f.
          round: 2
        - id: BR-2
          disposition: addressed
          note: git ls-tree 3f11b26f shows scripts/test-shard.py (and test-timing.py) as 100755.
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: TEST_NAME now admits Example, but no test pins it and no Example exists in cmd/sdlc, so no fixture reaches the change; scripts/test/ could host the check.
          round: 2
        - id: BR-4
          disposition: addressed
          note: cmd/sdlc/realgit_guard_test.go enforces it; per-dir grep shows fleet/tracker/gitx/activetime would fail without their TestMain call; passes at HEAD.
          round: 2
      recipe: milestone-review
      blocked: false
    - "n": 3
      timestamp: "2026-09-28T15:29:02-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: issuemovedetail.go:38 at dd78d856 still reads "this is an escape patch" in a comma-spliced sentence; unchanged since f3e7c907.
          round: 3
        - id: BR-3
          disposition: not-addressed
          note: Regex widened to Example (3f11b26f), but no test and no class check; add a runner coverage check that every -test.list name except Benchmark got a terminal event, else exit non-zero.
          round: 3
      findings:
        - id: BR-5
          severity: Minor
          title: test-shard.py write_atomic uses a fixed .tmp path in the shared git common dir, so concurrent make test runs from two slots can race
          detail: scripts/test-shard.py:151 always writes sdlc-test-timings.tmp. Use tempfile.NamedTemporaryFile(dir=path.parent, delete=False) and then os.replace.
          family: shared-state-write-race
          round: 3
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#253 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-28T13:19:23-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `help-text-accuracy` move-detail help text says "escape patch" (means escape hatch) and the sentence is ungrammatical
  cmd/sdlc/issuemovedetail.go:38, side-quest commit f3e7c907.
- **BR-2** [Minor] `script-invocation-contract` scripts/test-shard.py committed 100644 though its usage line invokes it directly
- **BR-3** [Minor] `shard-test-selection-completeness` test-shard TEST_NAME excludes Example functions, so future cmd/sdlc examples would silently not run under make test
- **BR-4** [Minor] `guidance-without-enforcement` "new git-driving package needs PreferRealGit TestMain" is enforced only by atlas prose, not a source guard

## Round 2 — 2026-09-28T13:21:25-07:00 (claude) — passed

### Disposed

- BR-1 — not-addressed — cmd/sdlc/issuemovedetail.go:38 still reads "Consult operator before move, this is an escape patch" at HEAD 3f11b26f.
- BR-2 — addressed — git ls-tree 3f11b26f shows scripts/test-shard.py (and test-timing.py) as 100755.
- BR-3 — not-addressed — TEST_NAME now admits Example, but no test pins it and no Example exists in cmd/sdlc, so no fixture reaches the change; scripts/test/ could host the check.
- BR-4 — addressed — cmd/sdlc/realgit_guard_test.go enforces it; per-dir grep shows fleet/tracker/gitx/activetime would fail without their TestMain call; passes at HEAD.

## Round 3 — 2026-09-28T15:29:02-07:00 (claude) — passed

### Disposed

- BR-1 — not-addressed — issuemovedetail.go:38 at dd78d856 still reads "this is an escape patch" in a comma-spliced sentence; unchanged since f3e7c907.
- BR-3 — not-addressed — Regex widened to Example (3f11b26f), but no test and no class check; add a runner coverage check that every -test.list name except Benchmark got a terminal event, else exit non-zero.

### Raised

- **BR-5** [Minor] `shared-state-write-race` test-shard.py write_atomic uses a fixed .tmp path in the shared git common dir, so concurrent make test runs from two slots can race
  scripts/test-shard.py:151 always writes sdlc-test-timings.tmp. Use tempfile.NamedTemporaryFile(dir=path.parent, delete=False) and then os.replace.

## Open findings

- **BR-1** [Minor] `help-text-accuracy` move-detail help text says "escape patch" (means escape hatch) and the sentence is ungrammatical
- **BR-3** [Minor] `shard-test-selection-completeness` test-shard TEST_NAME excludes Example functions, so future cmd/sdlc examples would silently not run under make test
- **BR-5** [Minor] `shared-state-write-race` test-shard.py write_atomic uses a fixed .tmp path in the shared git common dir, so concurrent make test runs from two slots can race
