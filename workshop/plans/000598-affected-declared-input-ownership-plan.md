---
collaboration-mode: discussion
---

# Affected planning: declared path ownership Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a repository declare, in a curated low-tech mapping, which Go packages and groups a non-Go path's change selects, with automatic Go-import selection still doing the rest; and list the mappings the merge's full suite proves missing.

**Architecture:** Every declaration source (inline `owners` rules, parley-style traceability files, per-test sidecar files) compiles into one pure value, `affected.Mapping`, read from main (the trusted base). `affected.Plan` consults it inside the existing ownership step, so the union, dependency closure, group roots and `full` precedence stay the one planner's (ARCH-DRY). A separate part records "missed mappings": an introduced failure at merge whose package the branch's selection would not have run.

**Tech Stack:** Go (`cmd/sdlc/internal/affected`, `cmd/sdlc`), `go.yaml.in/yaml/v3` (already a dependency) for traceability files, `gitx.TrunkView.FilesSelected` for batched reads from the base tree.

---

## Context (facts this plan rests on)

- `affected.Plan` (`internal/affected/plan.go`): `full` globs first; then per path the union of `Graph.owners` (dir, testdata, embeds) and group `paths`; an unclaimed non-doc path is `full`. Changed packages expand to dependents; a group whose `go` root is selected joins. `Group.Go` never adds packages.
- `readAffectedConfig(git, rev)` (`cigate.go`) reads `.sdlc/affected.json` from main for every consumer: the affected run (`localcheck_cmd.go`), resume, pre-close choice, car stand-in and `splitImpact` (landing attribution), which calls `Graph.Owners` directly when a selection is uncertain.
- Pair (pair#853): 583 certificates `tests/affected-readers/<test path>.json`, each `{version, disposition, sha256}`; no declared reads yet. `full` holds `tests/**` and `cmd/couch/**`. The Done-when's certificate is `tests/affected-readers/cmd/internal/testcouch/agentfixture/request_lifetime_test.go.json` (the issue's `request_lifetime_test.go.json`; it is not under cmd/couch).
- ariadne declares no `.sdlc/affected.json`; pair is the consumer.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Mapping` | `cmd/sdlc/internal/affected/mapping.go` | new |
| `OwnerRule` | `cmd/sdlc/internal/affected/mapping.go` | new |
| `FullRule` | `cmd/sdlc/internal/affected/config.go` | new |
| `Config` | `cmd/sdlc/internal/affected/config.go` | modified |
| `Plan` | `cmd/sdlc/internal/affected/plan.go` | modified |
| `ParseTraceability` | `cmd/sdlc/internal/affected/traceability.go` | new |
| `ParseSidecar` | `cmd/sdlc/internal/affected/sidecar.go` | new |
| `Missed` | `cmd/sdlc/internal/affected/missed.go` | new |

- **Mapping**: the compiled declarations: `[]OwnerRule`. Built by `Compile(cfg, traceFiles, sidecarFiles)` from already-read bytes; no IO. Tests in `mapping_test.go`.
  - **Relationships:** 1 per (config, base rev); Plan takes it as an argument alongside Graph and Config.
  - **DRY rationale:** three declaration formats, one consumer. Plan and `splitImpact` ask one function, `Mapping.Targets(path)`, what a path is declared to touch.
  - **Future extensions:** a new input format is a new compiler into `OwnerRule`, never a planner change.
- **OwnerRule**: `{Source string; Paths []glob or exact; Go []root; As []path; Strip *{Prefix, Suffix}}`. Exactly one of `Go`, `As` or `Strip` is set.
  - `Go` names package roots directly.
  - `As` aliases the path to target paths, which are classified by the graph's owners and by group `paths`.
  - `Strip` derives a single alias by removing a prefix and a suffix. The derived path must be clean, relative, have no `..` and be non-empty, or the plan goes full.
  - An alias is followed one level, so there are no chains and no cycles.
  - An alias target that matches `full` makes the selection full. Precedence stays conservative.
  - `Source` names where the rule came from (`affected.json owners[3]`, `atlas/traceability.yaml chat/spell_typeahead`, or a sidecar's path), for refusals.
- **FullRule**: a `full` entry may be a glob string (today's form) or `{"glob": G, "except": [globs]}`. A path matching `glob` but also `except` falls through to ordinary classification. This is how pair keeps `cmd/couch/**` full while `cmd/couch/*_test.go` selects only its package.
- **Config**: gains `owners` (inline rules: `{"paths": [...], "go": [...]}`, `{"paths": [...], "as": [...]}`, or `{"paths": [...], "strip": {"prefix": P, "suffix": S}}`), `traceability` (repo paths of traceability files) and `sidecars` (`[{"prefix": P, "suffix": S}]`). Parsing stays strict and duplicate-key safe.
- **Plan**: per path, after the `full` check and before the unclaimed test, the path claims through `Mapping.Targets`:
  - `Go` roots add packages. A root that is not a package in the head graph makes the plan full, exactly as stale group roots do.
  - `As` and `Strip` targets add the targets' graph owners and group matches. A target that neither claims leaves the path unclaimed, so the plan goes full.
  - Owners then flow into the existing `Dependents` and group-root steps unchanged.
- **ParseTraceability**: the parley format `version: 1`, `atlas: {key: {code: [..], tests: [..]}}`. Each entry becomes one `OwnerRule{Paths: code, As: tests}`.
  - Strict: a version other than 1, unknown keys, or a non-string or empty path refuses.
  - Entries without `code`, or without `tests`, contribute nothing; parley's own tooling owns their meaning.
- **ParseSidecar**: a sidecar at `<prefix><target><suffix>` (pair: `tests/affected-readers/` + `X_test.go` + `.json`) yields:
  1. a `Strip` rule: the sidecar's own path aliases `X`, so a certificate edit selects the test's package. This rule is path-only, so added, deleted and renamed certificates all qualify, including a certificate whose base copy does not exist.
  2. from a base copy's optional `reads: [glob]` member, `OwnerRule{Paths: reads, As: [X]}`: a change to a declared input selects the test's package.
  - ariadne reads only `reads`; every other member (`version`, `disposition`, `sha256`, future fields) is opaque and ignored.
  - ariadne never rewrites sidecars, so preserving those members is pair's tooling's job (ARCH-EVOLVE note below).
  - A malformed `reads` (not an array of valid globs, or a traversal) makes the plan full, naming the sidecar.
- **Missed**: `Missed(sel Selection, failures []string, pkgOf func(test string) (string, bool)) []Miss`. For a non-full selection, it returns each introduced failure whose package is not in `sel.GoPkgs`.
  - A failure whose package can't be told is not a miss. Group failures are out of scope, because suite logs name Go tests.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `readAffectedConfig` | `cmd/sdlc/cigate.go` | modified | git (`ls-tree`, `show`, batch `cat-file` via `gitx.TrunkView`) |
| `missedLedger` | `cmd/sdlc/affected_missed.go` | new | a JSON file in the common git dir |
| `diagnoseSuiteFailure` hook | `cmd/sdlc/suitebaseline.go` | modified | the merge's existing baseline comparison |
| `sdlc affected missed` | `cmd/sdlc/affected_missed_cmd.go` | new | CLI |

- **readAffectedConfig**: still the only reader. It additionally fetches the declared traceability files and the sidecar subtrees from the same rev: one `ls-tree` and one `cat-file --batch` per prefix, through the existing `gitx.TrunkView.FilesSelected`. It then calls the pure `Compile`. It returns `(Config, Mapping, ok, err)`.
  - Every consumer already calls it, so all of them derive from one source (ARCH-PURPOSE shadow sweep): the affected run, resume, pre-close, the car stand-in and `splitImpact`.
  - An unreadable or invalid declaration makes the reader return an error, so the run uses the suite instead, as an invalid affected.json does today.
  - Tests use real git in temp repos, which is how the existing `readAffectedConfig` and gitx tests run (ARCH-MOCK: git's fake is the existing temp-repo harness).
- **splitImpact**: its uncertain-selection fallback computes owners with `Graph.Owners` alone. It switches to a shared `affected.Owned(path, g, cfg, m)`, which Plan also uses, so landing attribution sees declared ownership too.

### Part C: missed mappings (separable)

- **Hook**: after `diagnoseSuiteFailure` classifies `introduced`, and only when the main baseline passed:
  - It plans the branch's delta (merge-base to head) with main's config, mapping and the head graph. That is what an affected run would have selected.
  - It records `Missed` entries. A red or unparsed main baseline records nothing, so load storms like 10-24's, where main failed too, add no noise.
  - Best-effort: a planning error is reported on the existing comparison line and never changes the gate's outcome.
- **missedLedger**: a file in the common git dir, `sdlc/affected-missed.json`, shared by every worktree of the clone. It is versioned, and unknown members are preserved on rewrite (ARCH-EVOLVE).
  - Each entry holds: changed paths (capped at 20 plus a count), failing test, package, head, main and time.
  - Entries are keyed by (package, sorted changed-path set) and deduped, keeping the newest evidence and a hit count.
  - It is written under a lock (temp file plus rename).
- **sdlc affected missed list | dismiss --id N**:
  - `list` prints the open entries. It drops any entry that main's current mapping now covers, i.e. the entry's paths now select its package (ARCH-FUNERAL: curation ends an entry).
  - `dismiss` removes an entry with a reason. The ledger is capped at 200 entries, and the oldest is dropped at the cap with a line saying so.
- **Train path**: a train's full-suite failure is attributed to a car (`landingtrain_attribution.go`). If the car boarded on an affected record that did not select the failing package, that is the same miss. The plan wires that hook after the merge hook, sharing `Missed` and the ledger.

## Operating envelope (ARCH-CONSTRAINTS)

- **Workload:** a batch check, run before an affected run, a pre-close and a merge diagnosis. It is not interactive.
- **Inputs:** pair has 583 sidecars, about 60 KB in total; parley's traceability file is 1,572 lines with about 1,000 paths.
  - Reading them costs one `ls-tree` and one batch `cat-file` per prefix, bounded by TrunkView's existing per-blob and aggregate output limits.
  - Over the limit, the reader errors and the run uses the suite instead.
- **Planning:** cost is (changed paths) × (rules).
  - Exact (glob-free) rule paths are indexed in a map, so sidecar `Strip` and traceability `code` lookups are O(1).
  - Globbed rules are scanned, as group paths are today. At 2,000 changed paths × 600 glob rules that is about 1.2M `Match` calls, well under a second; it is measured in a benchmark test.
  - There are no per-file subprocesses and no group multiplication. Aliases are one level deep.
- **Missed-ledger IO:** at most one read and one write per failed merge.

## Trust and formats (ARCH-SECURE, ARCH-EVOLVE)

- **Trust:** every declaration is read from main's rev, never from the head. A branch can't widen or narrow its own selection.
  - A branch that adds a certificate gets the path-only `Strip` alias from main's config. That is safe because it only adds coverage.
  - The certificate's `reads` count once it lands.
- **affected.json changes are additive:** new optional members (`owners`, `traceability`, `sidecars`) and an object form of a `full` entry.
  - An older sdlc binary refuses the new members, because of its strict parse. That makes it use the suite instead, which is safe; this is noted in the atlas.
  - Within the new members, the parse stays strict: unknown keys, duplicate keys and wrong types refuse.
- **Sidecars:** ariadne reads one member and writes nothing, so other members' preservation is untouched by ariadne. Pair's `--record` tool owns that, and pair#853 adds a test that a rewrite keeps `reads` and unknown members.
- **Traceability files:** read-only; parley owns them.
- **missed ledger:** the version and unknown members round-trip. An old reader rewriting a newer record keeps unknown members; there is a test for it.

## Ordering (ARCH-ORDER)

- **The planner holds no state between events.** It is a pure function of (paths, graph, config, mapping) per call.
- **The ledger's events** are concurrent merges in other worktrees recording, and `list`/`dismiss`. All of them read, modify and write under one flock with an atomic rename, so the last writer can't drop another writer's entry.
- **A crash mid-write** leaves the old file, because of the rename.
- **The event most likely to be mishandled** is two merges recording the same key at once. The lock plus key dedupe makes that a single entry with hits=2. There is a test with two goroutine writers.

## Lifecycle (ARCH-FUNERAL)

- **Mapping:** in-memory per run; it creates nothing durable.
- **The missed ledger:**
  - Created by merge diagnosis.
  - Last needed by the curator.
  - Removed by `dismiss`, or automatically once main's mapping covers the entry. It is capped at 200 entries, dropping the oldest with a notice, so at most about 200 × 2 KB.

## Plan

### M1: owners, full exceptions and sidecars (what pair#853 needs)

- [ ] `config.go`: `FullRule` (string or `{glob, except}`), `owners`, `sidecars`, `traceability` parse; strict tests for each new member, including bad roots, empty `as`, strip traversal, and unknown or duplicate keys (`config_test.go`).
- [ ] `mapping.go`: `OwnerRule`, `Compile`, the exact-path index and `Targets`. Tests: one-level alias, a strip derivation, rejecting a traversal-derived target, sidecar `reads` with opaque members ignored, a malformed `reads` resolving to a full reason.
- [ ] `plan.go`: consult `Mapping` in the claim step through a shared `Owned`.
  - Tests (`plan_test.go`): union with graph owners, dependents of a declared root, group roots selected via a declared root, `full` precedence over a declared rule, `except` falling through, a stale declared root going full, an alias target that claims nothing going full, an unmapped path keeping today's behavior, and an added/edited/deleted/renamed sidecar path selecting the test's package.
- [ ] `cigate.go readAffectedConfig`: fetch the sidecars and traceability files from the rev with `TrunkView.FilesSelected` and compile. Every caller now passes the Mapping. Real-git tests show that the config and sidecars come from the base, not the head, and that an unreadable sidecar falls back to suite.
- [ ] `landingattribution_impact.go`: the fallback uses `Owned`; extend its test with a declared owner.
- [ ] Qualification test: a fixture modeled on pair's `affected.json` minus `tests/**`, plus the sidecars. `tests/affected-readers/cmd/internal/testcouch/agentfixture/request_lifetime_test.go.json` selects `…/cmd/internal/testcouch/agentfixture` and its dependents. `cmd/couch/x_test.go` with `except` selects `cmd/couch`. `cmd/couch/main.go` stays full. `scripts/test-affected.py` (the selector) stays full.
- [ ] Benchmark: 2,000 paths × 600 sidecars × 600 glob rules.
- [ ] atlas `ci-merge-check.md` "The affected suite": owners, `except` and sidecars, a rewrite of "What the graph can't see" (a declared owner is now the answer for a non-Go input a Go test reads), and the older-binary fallback.
- [ ] `sdlc milestone-close --issue 598 --milestone M1`

### M2: parley-style traceability input

- [ ] `traceability.go` `ParseTraceability` with strict tests: parley's real file as a golden (copied into testdata), a bad version, unknown keys, non-string paths.
- [ ] Plan tests: a code path selects the test files' package (Go) or group (non-Go `paths` match). A test path that claims nothing goes full, naming the traceability entry.
- [ ] atlas: the traceability input, and that parley's own `make test-spec` reads the same file unchanged.
- [ ] `sdlc milestone-close --issue 598 --milestone M2`

### M3: missed mappings

- [ ] `missed.go` `Missed` and a package-of-test parser for suite failure names (`<import path>.TestName[/sub]`); pure tests.
- [ ] `affected_missed.go` ledger: lock, dedupe, cap, round-trip of unknown members, two concurrent writers.
- [ ] `suitebaseline.go` hook: record only when the main baseline passed and the selection was not full. A test with a fake run pair shows a recorded miss, no record on a red main, and no record for a full selection.
- [ ] Train attribution hook: a culprit car whose affected record missed the failing package records the same entry. The test extends `landingtrain_attribution_test.go`.
- [ ] `sdlc affected missed list|dismiss` with help text; auto-drop of entries that main now covers.
- [ ] atlas: the curation loop.
- [ ] `sdlc close --issue 598`

## Coordination with pair#853

ariadne owns the generic capability: owners, `except`, the sidecar convention with a `reads` member, traceability input and missed mappings. Pair#853 owns:
- adding `reads` to its certificates (audited per reader), and to its `--record` tool, preserving unknown members;
- replacing `tests/**` with narrower `full` entries for its policy/selector/harness paths (`.sdlc/**`, `scripts/test-affected.py`, `tests/lib/**`, `tests/affected-config-test.py` …);
- `"except": ["cmd/couch/*_test.go"]` on `cmd/couch/**`;
- the sidecar declaration `{"prefix": "tests/affected-readers/", "suffix": ".json"}`.

What each disposition means for selection stays pair's (for example, a `non_go_full_fallback` reader's inputs stay in pair's `full`). ariadne treats the disposition as opaque.
