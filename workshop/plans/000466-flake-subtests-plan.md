# Flake Registration for Exact Subtests Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let the known-flake registry name an exact Go subtest. A failure is explained only by the registered subtest, never by its siblings. A card owner can narrow or remove a registered name through a verb.

**Architecture:**
- **Names:** a registry name may carry subtest segments: `<import path>.<TestName>/<sub>/…`.
- **No record change:** the suite log parser already sees subtests and currently drops them. It now also returns the failing subtests, as a log fact that is never persisted. The test-stats record format is unchanged, so no schema bump and no older-reader refusal.
- **Classification:** classification walks each failing top-level test down to its failing leaves. Every leaf must be covered by a registered name: the leaf itself, or an ancestor (the top-level test included).
- **Rerun:** the rerun is unchanged in shape. It reruns the top-level tests that own the known entries, and must show each known entry's exact `--- PASS:` line.
- **Card verbs:** `flake unregister` and `flake narrow` are compare-and-swap writes, like `flake register`.

**Tech Stack:** Go: `cmd/sdlc/internal/flakes` (pure), `cmd/sdlc` (log parse, verbs).

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `SplitTest` / `ValidTest` (names with subtest segments) | `cmd/sdlc/internal/flakes/registry.go` | modified |
| `Covers(registered, failing)`, `Leaves(top, failing)` | `cmd/sdlc/internal/flakes/registry.go` | new |
| `Result.Subtests`, `Classify` (leaf coverage) | `cmd/sdlc/internal/flakes/classify.go` | modified |
| `RerunTests` (the top-level tests a known set reruns), `RerunMissing` (exact path) | `cmd/sdlc/internal/flakes/rerun.go`, `classify.go` | new / modified |
| `removeFlakeTests`, `narrowFlakeTests` (card decisions) | `cmd/sdlc/flake_cmd.go` | new |
| `failingNames` (one scanner; top-level and subtests) | `cmd/sdlc/suitefailures.go` | modified (DRY) |

- **Names.**
  - **Split:** `SplitTest` finds the leftmost dot whose left side is a valid import path and whose right side starts with a valid `Test…` function, followed by `/` or the end. A subtest segment containing a dot (`dev-aliases.test.sh`) therefore cannot be mistaken for the package boundary.
  - **Segments:** each segment is 1–128 characters of `[A-Za-z0-9_.+=,:@#%~-]`, and never `.` or `..`. That is what `go test` prints after rewriting spaces. It is argv-safe, and `regexp.QuoteMeta` makes it regex-safe wherever it reaches `-run`.
  - **Return:** `SplitTest` returns the package, the top-level test, and the subtest path.
  - **ARCH-SECURE:** names are parsed and validated at the boundary, on the card read, register, narrow and the registry build.
- **Coverage.** `Covers(r, f)` holds when `f == r` or `f` starts with `r + "/"`. A registered subtest covers its own descendants, never its siblings, and never its parent's own failure.
- **Leaves.** For a failing top-level test `T`, `Leaves` is the set of failing subtests under `T` with no failing descendant. When none are listed, it is `{T}` itself.
- **Classify.** For each failing top-level test:
  - every leaf must be covered by some registered name, and the covering names become the known entries;
  - an uncovered leaf is unknown, and the reason names the exact subtest, for example "a sibling".
  - A parent that fails only because its subtests fail is explained by those subtests. A parent's own `t.Error` beside them can't be told apart in the log. That is accepted and documented: the parent's failure is re-proven by the rerun, which must pass the whole top-level test.
- **Rerun.** `RerunTests(known)` is the deduplicated top-level tests that own the known entries. The rerun command keeps today's one `go test -run '^(A|B)$'`. It reruns the whole top-level test, siblings included, so a sibling that fails in the rerun blocks (conservative). `RerunMissing` requires `--- PASS: <exact path> (` for every known entry and for its top-level test.
  - A rerun that failed only on covered leaves is `KnownFlakeTwice`.
  - A rerun failure on an uncovered leaf is `Unknown`, naming it.
- **Card decisions.** Both are pure, on the compare-and-swap card's current list:
  - `removeFlakeTests(have, name)`: drops the name; a name already absent is no write (idempotent).
  - `narrowFlakeTests(have, old, new)`: requires `new` to be strictly under `old` (`Covers(old, new) && new != old`), and replaces `old` with `new`. Already narrowed (`old` absent, `new` present) is no write. Neither present refuses ("nothing to narrow").
  - Both keep any names a concurrent registration added. Both refuse on a closed issue.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `parseSuiteLogFacts` (`Subtests`, `json:"-"`) | `cmd/sdlc/teststats_parse.go` | modified | the suite log |
| `knownFlakeStep` (passes subtests in; rerun of top-level tests) | `cmd/sdlc/localcheck_flakes.go` | modified | the rerun runner |
| `flake register` evidence for a subtest | `cmd/sdlc/flake_cmd.go` | modified | the saved failed log |
| `flake unregister`, `flake narrow` | `cmd/sdlc/flake_cmd.go` | new | `runCardUpdate` (the CAS setter) |
| recovery catalog entries | `cmd/sdlc/internal/recovery/catalog.go` | new | — |

- **Subtest evidence.** Registering a subtest needs the top-level test's same-head fail→pass (`SameHeadFailPass`, unchanged). The failing attempt's saved log (`failed/<id>.log`) must also show a failing leaf the name covers. A missing or unreadable log refuses: no evidence, no write.
- **Ownership** of unregister and narrow: the card's owner, or anyone while the issue is unclaimed. This is the same rule as `issue publish`'s details, via `issue.CardClaimant` and `MatchClaimant`. Register stays any-slot (#394). (Q1.)

## Persisted formats (ARCH-EVOLVE)

- **Test-stats attempt records:** unchanged. `failures` stays top-level, and subtests live only in the log and the in-memory facts (`json:"-"`). No version bump.
- **Card `flake_tests`:** same field. Names may now carry `/…` segments. An older binary's registry build refuses such a name as malformed, which makes the known-flake rule unavailable on that machine: every failure blocks, fail-closed. Rollout is therefore ":0 rebuilt before anyone registers a subtest". I'll state it to the hub.

## State design table (card verbs)

**Actions:** `flake register`, `flake unregister`, `flake narrow`, a repeat of any of them, a concurrent registration, the issue closing.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| The name is registered and the caller owns the card | `flake list` shows it | `unregister` / `narrow` | none |
| A concurrent register adds a name while narrow runs | both names after the CAS retry | none | none |
| A repeat of `unregister` or `narrow` (lost reply) | the card already shows the result | rerun the same command (no write) | none |
| `narrow` with `new` not under `old` | refusal names both | `unregister` old, then `register` new (with evidence) | none |
| A caller who is not the owner, on a claimed card | refusal names the owner | ask the owner (or the hub) | none |
| The issue is closed | refusal | nothing (closing already unregistered) | none |
| An older binary reads a subtest name | known-flake rule unavailable there; every failure blocks | rebuild :0 / `weave refresh` | landing time only |

## Operating envelope

N/A for latency. The rerun is the existing one, at most `MaxRerun` top-level tests. A subtest registration reruns its whole top-level test: `TestConstructScripts` is about 2–5 minutes, inside the 15-minute rerun bound (Q2).

## Plan

- [ ] Names: `SplitTest`/`ValidTest` with subtest segments; `Covers`, `Leaves`. Table tests, including the `dev-aliases.test.sh` dot case, a forged `pkg.TestA/sub.TestB`, and refusals of shell and regex text.
- [ ] Log: `failingNames` keeps subtests, `failingTests` filters to top-level (DRY), and `SuiteLogFacts.Subtests` stays unpersisted. Tests on the `go test -v` format and the sharded `--- FAIL pkg Test/sub` format.
- [ ] Classify: leaf coverage, known entries are the covering names, uncovered siblings are unknown. `RerunTests`, exact `RerunMissing`, rerun-failure attribution. Pure tests, including one where a registered subtest's failure reruns while its sibling's failure blocks.
- [ ] Gate: `knownFlakeStep` passes `Subtests` in and reruns `RerunTests`. Done-when test with a real nested Go fixture package (a test with two subtests, one failing). The registered subtest's failure takes the rerun and lands; the sibling's failure blocks.
- [ ] Verbs: register evidence for a subtest from the saved log; `flake unregister`; `flake narrow`. Owner rule, compare-and-swap with a concurrent addition kept, idempotent repeats, and real tracker cards, as in #394's `TestAddFlakeTestsMergesOntoTheCard`.
- [ ] Docs: `sdlc flake --help` and the register help, recovery catalog entries (writes, retry behaviour), the atlas, and the README.
- [ ] Targeted tests, then the recorded suite.

## Questions for the hub

- **Q1 Ownership:** unregister and narrow by the card owner, or anyone while unclaimed (the details-publish rule). Register stays any-slot. Or should removal be any-slot too, like register?
- **Q2 Rerun scope:** a subtest registration reruns its whole top-level test, not the subtest alone. That is simpler and conservative: a sibling failing in the rerun blocks, and one go invocation still covers every known entry. Mixing exact subtest patterns with whole tests would need one `go test` per subtest, because `-run` patterns apply per level. The cost is rerun time, for example `TestConstructScripts` at a few minutes.
- **Q3 Rollout:** subtest names in `flake_tests` make older binaries' registries fail closed, so :0 must be rebuilt before anyone registers one. #465's broad registration can then be narrowed with `flake narrow`.
