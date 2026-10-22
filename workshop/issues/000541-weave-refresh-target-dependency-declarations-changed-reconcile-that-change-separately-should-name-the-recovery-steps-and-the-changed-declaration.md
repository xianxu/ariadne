---
id: 000541
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '0b22861c6ce7548c524e350dff06136f1d6627d5' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T00:31:21-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "1a665a97", done: "02bf4c2c"}
---

# weave refresh: 'target dependency declarations changed; reconcile that change separately' should name the recovery steps and the changed declaration

## Problem

## Spec

From ops-echo:0 (10-21 evening): after echo#1 landed (it changed echo's `construct/deps` ariadne source from SSH to
HTTPS, operator-approved), Couch's prepare/refresh on echo:1 failed with
`…/echo-slot1/echo: target dependency declarations changed; reconcile that change separately`
(`cmd/weave/internal/refresh/refresh.go:157`). Nobody knew what "reconcile separately" meant; echo:0 got through by
fast-forwarding first. The refusal is right (refresh won't apply a dependency change), but the error must name the
step.

Fix: the error names the exact recovery for the common case, a clean resting branch behind origin/main:
"dependency declarations change between HEAD and origin/main (construct/deps …): fast-forward this checkout
(`git merge --ff-only origin/main`) and run ordinary setup (`./bootstrap.sh`), then retry", listing which
declarations changed (names and fields, never credential-bearing URLs). Document it in the setup atlas.

## Done when

- A fixture where origin/main changes a deps source yields the new message, naming the changed declaration and the
  steps (test).
- The atlas setup page documents the recovery.
- Diagnostics identify changed substrate paths or data mounts and changed field names, including additions,
  removals and ordering changes; source values (including URLs with credentials or query secrets) never appear.
- Refresh still refuses before branch updates or compilation. Divergent branches receive no unconditional
  fast-forward instruction; existing dirty-checkout, malformed-declaration and substrate-only controls remain.

## Plan

- [x] Red first: extend the existing real bare-remote refresh fixture with an existing substrate whose
  origin/main declaration changes from SSH to HTTPS, matching echo's reported incident. Assert the changed
  declaration and `source` field, clean-checkout recovery commands, unchanged HEAD and zero compilation.
  Use synthetic source strings; record the incident sequence here rather than copying live URLs.
- [x] Add a pure declaration-difference formatter beside refresh's existing comparison. Reuse parsed
  `layergraph.Dependency` rows and the existing DeepEqual refusal; compare ordered rows, identifying old/new
  substrate paths or data mounts and changed Kind/Path/Source/Mount fields. Identify rows by declaration
  ordinal too, so additions, removals, reorders and duplicate declarations remain understandable. Never
  print Source values; quote path/mount labels and suppress URL/scp-like labels in favor of the ordinal.
  No additional Git calls, discovery, persistence or declaration parser.
- [x] In the existing refusal name `construct/deps`, HEAD and origin/main. For the already-observed
  fast-forward case say: in the named checkout, `git merge --ff-only origin/main`, then `./bootstrap.sh`,
  then retry the original refresh command. If ahead/divergent, say reconcile that branch first and use
  ordinary setup after adopting the declarations; preserve the existing divergence refusal.
- [x] Table-test the pure formatter for source/path/mount/kind changes, additions/removals, duplicates,
  reorder, unchanged rows, and synthetic credential/query-bearing URLs (including URL-like labels).
  Add real-Git divergent and substrate-only controls; retain existing malformed/dirty refusal tests.
  Document recovery and its clean/behind prerequisite in `atlas/workflow/setup-and-replication.md`.
- [x] Run targeted refresh/CLI tests, then the recorded pushed-head suite and `sdlc close` review after
  design approval and change-code. One quick-flow boundary, no milestones; target <100 production lines.

## Log

### 2026-10-21

### 2026-10-22 — claimed and planned

- TL assigned claim and plan, small quick flow. Claimed and started on a fresh issue branch from main.
- Root cause: `RunWithOptions` already fetches and parses both generations, but its DeepEqual mismatch
  discards the differing declaration details and returns only “reconcile that change separately”. The
  check is before all update effects. `TestDeclarationDrift` proves refusal but does not inspect guidance.
- Chosen approach: format the existing parsed rows, omitting source values. Generic fixed guidance alone
  misses the changed declaration; rendering raw diffs or redacting URLs after rendering unnecessarily
  exposes secrets to a new diagnostic path. No behavior change to refresh admission or mutation order.
- Architecture lenses: ARCH-DRY reuses parsed rows and the current real-Git fixture; ARCH-PURE puts formatting
  in a deterministic helper; ARCH-PURPOSE covers names, fields, recovery and secrecy together. ARCH-MOCK
  uses existing temporary bare Git repositories, not fleet checkouts. ARCH-CONSTRAINTS stays within the
  existing 1 MiB declaration/128-layer bounds and adds only linear in-memory formatting, no subprocesses.
  ARCH-SECURE treats target declarations as untrusted: omit all source values and suppress URL-like labels;
  keep current malformed-input refusal. ARCH-ORDER holds no new state between events: refusal remains in
  preflight and no automatic recovery is executed. ARCH-FUNERAL creates nothing durable at runtime because
  diagnostics are returned through the existing error. ARCH-EVOLVE changes no persisted format.
- Awaiting TL approval of this committed design; no implementation or test changes yet.

## Revisions

- 2026-10-22: Made the quick-flow plan executable from the current refresh seam; added explicit safety
  controls for divergent branches and URL-like declaration labels while preserving the original recovery goal.

### 2026-10-22 — red and implementation

- TL approved this plan (Couch receipt 611efdae-ebbd-48a0-90df-ab6a1261e5e9, submitted);
  `sdlc change-code --issue 541` admitted quick flow.
- Red evidence: `go test ./cmd/weave/internal/refresh -run '^TestDeclarationTransportRecovery$' -count=1`
  failed all three original subcases at the intended old refusal, missing construct/deps, named declaration,
  changed source field and recovery steps. HEAD/no-compilation controls held. An initial fixture attempt
  used Git insteadOf, which changes remote get-url's observed identity and stopped at origin validation;
  corrected the fixture's transport seam to keep the actual SSH origin and fetch only from its local bare
  remote. No real network endpoints or credentials are used.
- Green: that regression plus `TestDeclarationChanges` passed (6.4s). Formatter omits source values, quotes
  local path/mount labels, suppresses URL/scp/query-like labels and reports ordered row/field changes.
  Existing preflight ordering and DeepEqual admission remain unchanged. Added end-to-end synthetic
  credential-bearing target-source control before full targeted verification.

- Targeted verification passed: all `./cmd/weave/internal/refresh` tests (76.107s), including the four
  real-Git transport/recovery variants and 12 pure diff cases; `./cmd/weave -run TestRefresh` (17.082s).
  Existing dirty/malformed/admission and substrate-only controls passed. `git diff --check` clean.
  Production additions: 56-line formatter plus 5 added integration lines, inside the quick-flow shell.
  No live fleet mutation or :0 smoke is needed to prove a preflight diagnostic; the incident-shaped private
  real-Git fixtures and CLI controls exercise the changed path. Recorded suite and close review remain.

### 2026-10-22 — recorded suite passed; closing

- PR #387: https://github.com/xianxu/ariadne/pull/387. Current GitHub checks all green, including
  local-gates-guard; no protected gate configuration changed.
- Recorded `sdlc local-check run --name suite` on c04d7afed5768941f083205ddf95a17bd15ed7a7 passed:
  8/8 shards, 2103 cmd/sdlc tests and 76 other packages, zero failures; wall 706.5s (11m47s).
  Lease lease-2288301901 waited 877s for capacity before starting, then completed normally. Saved log
  `.git/sdlc/local-checks/c04d7afed5768941f083205ddf95a17bd15ed7a7/suite.log`, SHA256 prefix 13a615c98585.
- All Done when criteria have executable evidence above; invoking the binary-owned close review next.
  Plan's final checkbox records verification and entry into that boundary, not a claimed review verdict.
