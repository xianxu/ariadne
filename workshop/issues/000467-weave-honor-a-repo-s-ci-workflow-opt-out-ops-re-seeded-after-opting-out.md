---
id: 000467
status: codecomplete
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '23e8b6c117837f44b5828f48f7204541fbae436d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T10:00:05-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.13
---

# weave: honor a repo's CI-workflow opt-out (ops re-seeded after opting out)

## Problem

ops' AGENTS.local.md says "ops opts out of ariadne's seeded CI workflows", yet a weave refresh in ops modified the tracked .github/workflows/merge-check.yml and added .github/workflows/local-gates-guard.yml. On 2026-10-19 the operator removed ops' CI workflow; the next weave refresh would seed it again.

## Spec

Give the opt-out a machine-readable form (a repo-level weave setting, e.g. in the repo's weave configuration: seeded CI workflows off) and make weave honor it: no seeding, modifying or re-adding of .github/workflows/* in an opted-out repo, while other seeded files (skills, settings) still refresh. weave reports "CI workflows: opted out" so the choice is visible. Migrate ops to the setting, and drop the prose-only claim or point it at the setting.

## Done when

- weave refresh in an opted-out fixture repo leaves .github/workflows untouched (and absent stays absent); a normal repo still gets them.
- ops carries the setting, and a refresh there leaves its workflows absent.

## Plan

Design: a leaf-only `decline <path>` row in the repo's own `construct/base.manifest` — the repo's weave configuration. ops writes `decline .github/workflows`.

Why a generic path verb rather than a `ci-workflows off` flag: the workflow paths are owned by ariadne's manifest. A named flag would make weave hardcode `.github/workflows` (ARCH-DRY), and it would need a new config file. A decline row is the same size, and the manifest stays the one source of what a repo receives.

Semantics:
- `decline P` is honored only in the leaf (the repo being woven). A decline row in an ancestor is ignored, so ariadne cannot decline on its derivatives' behalf; a visibility token on it is ignored.
- It removes every **inherited** (non-leaf) file-op intent (symlink, seed, seed-once, scaffold, touch, merge) whose target is P or under P. Prose, skill and ship intents have no single target and are unaffected. The leaf's own rows are never declined.
- It is applied once, as a pure function over the walked layers (`layer`/`walk`), so the plan, the completeness check and dry-run all see the same filtered set (ARCH-DRY, ARCH-PURE). The leaf Layer records what it declined, and that drives the report.
- Nothing already on disk is touched. Seeds are not in the managed inventory (`actionIdentities` covers only links and written files), so a declined seed is neither retired nor rewritten: absent stays absent, and a hand-kept workflow stays as it is. A declined *symlink* is in the inventory and would retire like any no-longer-produced output; that is the expected meaning of declining a link.
- Report: compile/refresh prints `weave: declined .github/workflows: opted out of 2 inherited row(s)` (dry-run too). A decline that matches no inherited row prints a warning, the same as an unmatched `ship` row, so a stale opt-out stays visible.

ARCH notes: EVOLVE: `decline` is a new verb, and an older weave skips it as an unknown verb. That older weave would re-seed, which is today's behavior. That's acceptable because the fleet runs :0's weave. SECURE/ORDER/FUNERAL/MOCK: N/A. It adds no state between events, nothing durable, and no external call: it filters the manifest rows that are read. CONSTRAINTS: N/A, one linear pass over the rows.

Steps:
- [x] `intent`: add a `Decline` kind and the `decline` verb (target = source).
- [x] Pure `ApplyDeclines(layers)`: filter inherited intents under the leaf's decline paths, recording `Declined{Path, Rows}` on the leaf. Call it in `walk.Load`. Unit tests: seed or symlink under a path is dropped, a sibling path is kept, the leaf's own row is kept, an ancestor's decline is ignored, a prefix like `.github/workflowsX` does not match, and an unmatched decline gives Rows 0.
- [x] compile prints the declined/unmatched lines (`runCompile`, dry-run as well).
- [x] Fixture test (cmd/weave): a two-layer fixture whose foundation seeds `.github/workflows/x.yml` plus another seed. Refreshing an opted-out leaf leaves `.github/workflows` absent while the other seed lands. An opted-out leaf holding its own edited workflow keeps it byte-for-byte. A normal leaf still gets the workflow.
- [x] Docs: verb comment in `construct/base.manifest` header and `construct/manifest.md`, atlas weave entry.
- Handed to the hub, outside this branch: ops adds `decline .github/workflows` to ops' `construct/base.manifest`; change AGENTS.local.md line to point at it; refresh shows the report line and no `.github/workflows`.

## Log

### 2026-10-19
- 2026-10-19: closed — go test ./cmd/weave/... ./pkg/layergraph/... ok; go vet clean; fixture tests: TestCompileDeclinedWorkflowsStayAbsent, TestCompileDeclineLeavesExistingWorkflowsUntouched, TestCompileSeedsWorkflowsWithoutDecline, TestCompileReportsStaleAndDependentDeclines; live: origin/main weave 931a4ad68 skips the decline row (exit 0, still seeds), branch weave drops it and reports; recorded suite queued on d1b016bab; review verdict: SHIP
- 2026-10-19: flow upgraded quick → full — 117 added lines in code files (limit 100)

- Hub approved the design and added two test asks: (1) prove an older weave skips the unknown verb; (2) a declined row that another row depends on (a merge into a declined path) must be reported, not silent.
- Built: `layer.Decline` (pure, called once in `walk.Load`), the `decline` verb, a `Kind.Verb()` read back from `kindByVerb`, and `printDeclined` in compile. The report lists every declined row by verb and target, which covers ask (2); `TestCompileReportsStaleAndDependentDeclines` tests a declined merge and a stale decline. ARCH-DRY: the completeness check's hand-kept `verbName` switch now derives from `Kind.Verb()`. Before, it was missing `seed-once`.
- Ask (1): the skip path in `intent.ParseManifest` (unknown verb, `continue`) has been unchanged since 52fe08c2d (#95 M2), and `TestParseManifestUnknownActionSkips` covers it. I also checked it live. The origin/main weave (931a4ad68) ran `compile --dry-run` on a two-layer fixture whose leaf has `decline .github/workflows`, exited 0, and still planned `seed .github/workflows/merge-check.yml`. That is today's behavior and not a refusal. The branch build on the same fixture drops that seed and prints `weave: declined .github/workflows: opted out of 1 inherited row(s): seed .github/workflows/merge-check.yml`.
- Naming: the plan's `ApplyDeclines` became `layer.Decline`.
- Tests: `go test ./cmd/weave/... ./pkg/layergraph/...` all ok, and `go vet` is clean.
