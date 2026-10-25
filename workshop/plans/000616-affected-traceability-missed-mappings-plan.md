---
collaboration-mode: discussion
---

# Affected planning: traceability input and missed-mapping ledger Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Two more inputs to the declared-ownership mapping that ariadne#598 M1 built:
- parley-style traceability files, so that "change the toolbar, run these" selects those tests;
- a ledger of missed mappings: failures the merge's full suite caught that the branch's affected selection would not have run, listed for curation.

**Architecture:**
- Traceability entries compile into the same `ownerRule` list as `owners` and sidecars. `Plan`, `claimPath` and `Owned` are unchanged (ARCH-DRY).
- Missed mappings are a pure function over a selection and a failure list, plus a pure ledger value, behind one thin IO shell.
- The shell is fed by two hooks: the merge's suite diagnosis and the train's split attribution.

**Tech Stack:** Go (`cmd/sdlc/internal/affected`, `cmd/sdlc`). Traceability files are parsed through `yaml.Node` with `go.yaml.in/yaml/v3`, which is already a dependency. Files are read through `gitx.TreeFilesSelected`, and locking uses `internal/flock`.

This plan carries out the approved design in `workshop/history/plans/000598-affected-declared-input-ownership-plan.md` (M2, M3, Part C, Ordering, Lifecycle). The decisions below that refine it are marked **(refines #598)**.

---

## Context (facts this plan rests on)

- **The #598 M1 pieces on main:**
  - `affected.Compile(cfg, sidecars)` (`internal/affected/mapping.go`);
  - `loadAffectedMapping(ctx, dir, rev, cfg)` (`localcheck_affected.go`), which reads the sidecar prefixes with `gitx.TreeFilesSelected` at main's pinned rev;
  - `claimPath`, which follows `as` targets one level, and names `r.source` when a target claims nothing or matches `full`.
- **Parley's `atlas/traceability.yaml`** (parley.nvim ee938b5a):
  - 1,581 lines, 59 KB, `version: 1`, `atlas: {key: {code: [...], tests: [...]}}`;
  - 65 entries with `code`, 64 with non-empty `tests`, and one `tests: []`;
  - no globs, anchors or other keys;
  - its tests are Lua specs (`tests/unit/*_spec.lua`), so in a repo that declares this file they select through a group whose `paths` match.
- **`diagnoseSuiteFailure(ctx, gateErr, main, deps)`** (`suitebaseline.go`) runs at a failed merge-gate suite (`landing.go:628`). It reruns main unrecorded and classifies `introduced` and `preexisting` from the two logs' `parseSuiteLogFacts(...).Failures`.
- **`conductor.assessSplit`** (`landingtrain_attribution.go`) feeds `splitImpact` the exact failing packages of a failed train prefix and its passing baseline. `splitImpact` (`landingattribution_impact.go`) plans each endpoint from its own config, mapping and graph:
  - `ImpactUntouched` means that neither endpoint's selection (non-full, no groups) reached the package.
  - `ClassifySplit` reports such a failure as `unattributed: untouched package …`, unless it is a registered flake.
- **Persisted-record codec:** `landingline/codec.go` has the strict, preserving codec (`decodeStrict`, `encodeWithExtra`). It is unexported and used only by landingline.
- **Packages:** `flakes.ParseName` gives a failure's package. `internal/affected` and `internal/flakes` don't import each other.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Config` (`traceability` member) | `cmd/sdlc/internal/affected/config.go` | modified |
| `Sources` | `cmd/sdlc/internal/affected/mapping.go` | new |
| `Compile` | `cmd/sdlc/internal/affected/mapping.go` | modified |
| `ParseTraceability` | `cmd/sdlc/internal/affected/traceability.go` | new |
| `TraceabilityFile` / `TraceEntry` | `cmd/sdlc/internal/affected/traceability.go` | new |
| `TreeFilesAt` | `cmd/sdlc/internal/gitx/snapshot.go` | new |
| `Missed` / `Miss` | `cmd/sdlc/internal/affected/missed.go` | new |
| `MissedLedger` | `cmd/sdlc/internal/affected/missed.go` | new |
| `DecodeObject` / `EncodeWithExtra` | `cmd/sdlc/internal/jsonstrict/object.go` | new (moved from landingline) |

- **Config.Traceability**: `"traceability": ["atlas/traceability.yaml"]`.
  - Each entry is an exact, clean, repository-relative path: no globs and no duplicates.
  - A non-array value or a bad path refuses, under the same strict parse as today.
  - An older binary refuses the member, so it uses the suite instead, which is safe. The atlas states this, as it already does for `owners`.
- **Sources** `{Sidecars []SidecarFile; Traceability []TraceabilityFile}`. `Compile(cfg, Sources)` replaces the two-argument form (the M1 review asked for it), so a future input format adds a field, not a parameter.
  - **Relationships:** 1 per (config, rev). Its only producer is `loadAffectedMapping`; its consumers are the planner tests and `Compile`.
- **ParseTraceability(path, data) ([]TraceEntry, error)**: a strict parse over a `yaml.Node` tree, not a struct decode.
  - `yaml.v3` coerces a scalar `1` into a string field and accepts aliases, so the walk checks every node. It requires:
    - one document;
    - a mapping root with exactly `version` (int 1) and `atlas`;
    - `atlas` a mapping of string keys;
    - each value a mapping with only `code` and `tests`;
    - each a sequence of `!!str` scalars that are clean relative paths. `code` may also be a valid glob.
  - These refuse: alias and anchor nodes, merge keys (`<<`), duplicate keys, a second document, and an empty string.
  - **The walk owns every refusal (PQ-1).** Decoding into a `yaml.Node` skips all of yaml.v3's decode-time checks, including `uniqueKeys` (decode.go:492 returns early for a Node target). So one helper, `mappingPairs(node, where)`, refuses a duplicate key while it yields a mapping's pairs. The root, `atlas` and each entry all use it, so no level relies on the library.
  - An entry missing `code`, or with empty `tests`, yields nothing; parley's tooling owns its meaning (#598).
  - **DRY rationale:** it reuses `validGlob` and `cleanRelPath` from config.go.
  - **Future extensions:** a `version: 2` is a new branch here; v1 stays strict.
- **Compile**: for each `TraceabilityFile`, each entry becomes `ownerRule{source: "<file> <key>", paths: code, as: tests}`.
  - `claimPath` does the rest:
    - a Go test target claims its package through the graph's owners;
    - a Lua spec target claims the group whose `paths` match;
    - a target that claims nothing goes full, naming `<file> <key>`;
    - a target in `full` goes full.
  - So the planner needs no new code (ARCH-PURPOSE: every consumer of `Mapping` — Plan, `Owned`, `splitImpact` — sees traceability for free).
- **Missed(sel, failures, known) []Miss**: when `sel.Full` is false and `sel.Groups` is empty, it returns each failure whose `flakes.ParseName` package is not in `sel.GoPkgs` and that `known` (the flake registry's cover) does not explain. A `Miss` is `{Test, Pkg}`.
  - Nothing is returned when:
    - the selection is full;
    - the selection has a group. **(refines #598)**: a group's command may run any test, so its reach is unknown. `splitImpact` already treats groups the same way (precision over recall).
    - the failure has a bare or unparseable name.
  - **DRY rationale:** one definition of a miss for both hooks; the package grammar is flakes'.
- **MissedLedger**: `{Version: 1; Entries []MissedEntry; Extra}`. A `MissedEntry` is `{ID, Package, Test, Paths (≤20), PathsTotal, Head, Main, Source ("merge" | "train"), FirstSeen, LastSeen, Hits, Extra}`.
  - `ID` is the first 12 hex digits of sha256(package + "\n" + the full sorted path set). Paths are capped at 20 for storage, but the dedupe key uses the full set.
  - Pure operations:
    - `Record(misses, paths, meta, now)` dedupes by ID: hits+1, newest evidence, earliest FirstSeen.
    - `Dismiss(id)` refuses an unknown ID.
    - `Prune(covered func(MissedEntry) bool)`.
    - The cap is 200: Record drops the oldest `LastSeen` and reports how many it dropped.
  - Codec: `DecodeMissedLedger`, plus `MarshalJSON` through `jsonstrict.DecodeObject` and `EncodeWithExtra`.
    - Unknown members survive at the top level and on every entry (ARCH-EVOLVE).
    - Known members stay strict: types are exact and non-null, duplicate keys refuse.
    - A version other than 1 refuses, because a v1 writer can't know what a v2 entry's members mean for dedupe or pruning. The IO shell then skips recording with a note.
- **DecodeObject / EncodeWithExtra**: landingline's codec moves to `internal/jsonstrict` (it wraps `cmd/internal/jsonstrict`; the codec goes beside the wrappers, not into the shared package). landingline calls it in place of its private copy (ARCH-DRY: one strict, preserving codec, now with two users).
  - Its existing tests move with it, and landingline's suite keeps passing unchanged.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `loadAffectedMapping` | `cmd/sdlc/localcheck_affected.go` | modified | `gitx.TreeFilesSelected` at main's rev |
| `affectedSelectionAt` | `cmd/sdlc/landingattribution_impact.go` | new (extracted) | detached worktree, `graph_prepare`, `LoadGoGraph`, config and mapping reads |
| `missedLedgerStore` | `cmd/sdlc/affected_missed.go` | new | `<git-common-dir>/sdlc/affected-missed.json` under flock |
| `diagnoseSuiteFailure` hook | `cmd/sdlc/suitebaseline.go` | modified | the merge gate's baseline comparison |
| `assessSplit` hook | `cmd/sdlc/landingtrain_attribution.go` | modified | train split attribution |
| `sdlc affected missed list\|dismiss` | `cmd/sdlc/affected_missed_cmd.go` | new | CLI |

- **loadAffectedMapping** also reads each `cfg.Traceability` path at the same pinned rev, with one `TreeFilesSelected(dir(path), exact-path select)` per file.
  - A declared file that is absent on main is an error, so the selection is full: "declared ownership unavailable: traceability <path> missing". A declaration that points at nothing is stale, like a stale root.
  - Reads stay within the existing 1 MiB per-blob limit; parley's file is 59 KB.
- **affectedSelectionAt(ctx, root, git, cleanupGit, declRev, graphRev, paths)** returns `(Selection, Graph, Config, Mapping)`. It is the body of `splitImpact`'s per-endpoint loop, extracted.
  - `splitImpact` calls it with declRev = graphRev = the endpoint.
  - The merge hook calls it with declRev = main and graphRev = the merged head. That is the selection an affected run of that delta would have made (main's config, the head's graph).
  - **DRY rationale:** one place builds a selection from commits.
  - **Tests:** splitImpact's existing tests cover it unchanged; the merge hook test drives it on a real temp repo. Real git in temp repos is the git fake already used here (ARCH-MOCK).
- **missedLedgerStore**: `record(meta, misses, paths)`, `list`, `dismiss`, `prune`. Each is one read-modify-write:
  1. Open `affected-missed.lock` beside the ledger.
  2. Take `flock.Exclusive`, blocking, under the caller's ctx deadline.
  3. Read, then decode (a missing file is empty).
  4. Apply the pure op.
  5. Write a temp file in the same dir and rename it.
  - The path is from `git rev-parse --path-format=absolute --git-common-dir`, so every worktree of a clone shares one ledger.
  - The lock file stays: it is one empty file per clone and is reused.
  - **Injected into:** the hooks and the CLI take the store as a value built from the repo root. Tests point it at a temp dir.
- **diagnoseSuiteFailure hook**, after `introduced` is classified and only if `baseline.passed()` and `len(introduced) > 0`, calls `deps.RecordMissed(ctx, main, head, introduced)`.
  - Production wiring:
    1. Changed paths come from `ChangedPathArgs(main, head)`.
    2. `affectedSelectionAt(main, head)` gives the selection.
    3. The flake registry gives `known`, as `assessSplit` reads it.
    4. `Missed` finds the misses, and `store.record` stores them.
  - It runs under its own 2-minute budget (`missedRecordTimeout`, from ctx). Any error becomes a label: `missed mappings: not recorded (<why>)`.
  - Success adds `missed mappings: recorded N (sdlc affected missed list)`.
  - The gate's error is returned as before, so recording never changes the outcome.
  - `localRunDeps.RecordMissed` is nil in the existing tests; nil means skip.
- **assessSplit hook**: `splitImpact` also returns the changed paths it used, in a small result struct instead of a second bare value.
  - After classification, every failure with `Impact == ImpactUntouched` and no `RegisteredIssue` is a miss. The paths are hi's delta from lo; the Source is "train".
  - This is the existing classification — Untouched only arises for non-full, group-free selections — so the hook doesn't call `Missed` again. It builds the same `Miss` values and records them with the same store.
  - A record error is written to the conductor's diagnostic stream and never changes `done.Verdict`.
  - The ctx is the split's `splitAttributionTimeout` ctx.
- **`sdlc affected missed list [--json]`** prints the open entries: id, package, test, hits, last seen, paths, head and main.
  - Before printing, it prunes entries that main now covers. An entry is covered when `affectedSelectionAt(main, main, entry.Paths)` is non-full and selects `entry.Package`. That is one graph load per `list`, with selections computed per entry.
  - If the graph or mapping is unavailable, it lists without pruning and says why.
- **`sdlc affected missed dismiss --id ID --reason TEXT`** removes the entry and prints the reason. An unknown ID refuses.
  - The help text states the RECOVERY contract: dismiss is idempotent by ID, and a repeat refuses as unknown.

## Operating envelope (ARCH-CONSTRAINTS)

- **Traceability load:** batch, before planning. It adds one `ls-tree` and one `cat-file` per declared file, and is bounded by the snapshot limits; over a limit, the selection is full.
  - Compile: about 1,000 exact paths go into the `exact` index, so the planner cost doesn't change.
  - The M1 benchmark gains a traceability case of 65 entries × 20 paths. The budget is unchanged: 2,000 changed paths stay under 100 ms/op.
- **Merge hook:** runs only on a failed merge suite whose main baseline passed. It adds one detached worktree, one `graph_prepare` and one `LoadGoGraph`: tens of seconds, against a suite of about 20 minutes. Capped at 2 minutes; past that, nothing is recorded and a label says why.
- **Train hook:** adds no IO beyond one ledger write; the selection is already computed.
- **list:** interactive, run by hand: one worktree plus a graph load (seconds), then one `Plan` per entry (≤ 200).
- **Ledger:** at most 200 entries, about 2 KB each, so about 400 KB. Each write is one lock, one read and one rename.
- **Concurrency:** none added.

## Trust and formats (ARCH-SECURE, ARCH-EVOLVE)

- **Traceability bytes** are read from main's rev only, like every declaration (#598): a branch can't add its own traceability to narrow its selection.
  - They are untrusted text, parsed into typed entries at the boundary.
  - Aliases are refused, so there is no alias expansion and no billion-laughs case.
  - Paths must be clean, so there is no traversal.
- **The ledger** is a file in the common git dir, written by any sdlc of any version on this clone, and possibly hand-edited.
  - The decode is strict on known members. A refused file makes `record` skip with a note and makes `list` refuse with the path, so a curator repairs or deletes it. It is never overwritten.
  - Unknown members round-trip. A test has an older-shaped writer rewrite a newer record and keep its top-level and per-entry unknown members.
  - The ledger holds no credentials: test names, paths and SHAs only.
- **affected.json** gains one optional member (additive).

## Ordering (ARCH-ORDER)

- **The planner, `Missed` and the ledger ops are pure:** they hold no state between events, because each is a function of its inputs per call.
- **The ledger's events:** `record` from concurrent merges or trains in other worktrees, `list` (which prunes) and `dismiss`.
  - Every one is a read-modify-write under one exclusive flock, with a rename, so no interleaving loses another writer's entry.
  - A crash mid-write leaves the previous file, and the orphaned temp file is named `.affected-missed-*`. **(refines #598)**: `record` removes such temp files older than an hour while holding the lock.
- **The event most likely to be mishandled:** two merges recording the same ID at once. The lock serializes them, so the result is one entry with Hits = 2. A test runs two goroutine writers against one temp-dir store. A second test interleaves `dismiss` with a `record` of a different ID and checks that both effects survive.
- **Repeats:** a retried merge diagnosis records the same ID again, which adds to Hits. That is intended evidence (the miss recurred), not a duplicate entry.

## Lifecycle (ARCH-FUNERAL)

- **Ledger entries:**
  - created by the two hooks;
  - last needed by the curator;
  - removed by `dismiss`, by `list`'s prune once main covers them, or by the 200 cap, oldest first, with a "dropped N" line on that record's label.
  - Bound: about 400 KB.
- **The temp files and the lock file:** temp files are swept by `record` after an hour. The lock file is one per clone and permanent, at zero bytes.
- **Mapping and Sources:** in memory per run; they create nothing durable.

## Non-goals

- No misses for non-Go group failures; suite logs name Go tests.
- `list` never rewrites the traceability file or affected.json. Curation is the repo's edit.
- No change to which check pre-close or a car requires, and no change to `ClassifySplit`.

## Plan

### M1: traceability input

- [x] Move landingline's `decodeStrict` and `encodeWithExtra` to `internal/jsonstrict` (`object.go`) as `DecodeObject` and `EncodeWithExtra`, with tests. landingline calls them; `go test ./cmd/sdlc/internal/landingline/ ./cmd/sdlc/internal/jsonstrict/` passes. (This lands in M1 because it is a no-behavior refactor; M2 is its new user.)
- [x] `config.go`: the `traceability` member. Strategy: the existing strict-parse fuzz target gains the member in its seeds, with the guard that a parsed config validates again.
- [x] `traceability.go` `ParseTraceability`. Strategy:
  - a golden test on `testdata/parley-traceability.yaml` (parley.nvim ee938b5a, copied in);
  - a fuzz target seeded with hand-edit forms (duplicate keys at each level, aliases, merges, non-string scalars, extra documents, traversals), with the guard that it either refuses or yields only clean paths, and never panics.
- [x] `mapping.go`: `Sources` and `Compile(cfg, Sources)`, plus the traceability rules. Update the callers: `loadAffectedMapping` and the M1 tests.
- [x] Traceability through `Plan`. Strategy: the existing `Owned`-includes-graph-owners property extends to traceability rules, plus one `TestPlanTraceability` table over the claim contract (Go target, group target, unclaimed and `full` targets go full, naming `<file> <key>`). The benchmark gains the traceability case.
- [x] `loadAffectedMapping` reads traceability. Strategy: extend the real-git declared fixture so the file is read from main, not the head, and a missing or malformed file on main makes the selection full, naming it, in both modes.
- [x] atlas `ci-merge-check.md` "The affected suite": the traceability input, and the fact that parley's `make test-spec` reads the same file unchanged. README member list.
- [ ] `sdlc milestone-close --issue 616 --milestone M1`

### M2: missed-mapping ledger

- [ ] `missed.go` `Missed`. Strategy: one table over the "nothing recorded" contract (full, grouped, bare or unparseable name, known flake), plus the property that a returned miss's package is never in `sel.GoPkgs`.
- [ ] `missed.go` `MissedLedger` codec and ops. Strategy:
  - a round-trip property: an older-shaped writer rewrites a newer record and keeps its unknown members at both levels;
  - a fuzz target over the decoder, guarding that it refuses or re-encodes losslessly;
  - a sequence test of Record, Dismiss and Prune checking the invariants (unique IDs, ≤ 200 entries, Hits equal to the record count per ID).
- [ ] `affected_missed.go` store. Strategy: a concurrent-writers test on a temp common dir (same ID gives Hits = 2; an interleaved dismiss and a record of another ID both survive), and a refused file is never overwritten.
- [ ] Extract `affectedSelectionAt` from `splitImpact`; `landingattribution_impact_test.go` passes unchanged.
- [ ] `suitebaseline.go` hook with `localRunDeps.RecordMissed`. Strategy: drive `diagnoseSuiteFailure` with fake runs over the record/skip contract (records only on introduced + green main), with the guard that the returned gate error wraps the original in every case. Plus one real-git test of the production recorder: a branch that changes a non-Go input a Go test reads, with no mapping, records the test's package.
- [ ] `assessSplit` hook: `splitImpact` returns its paths. Strategy: extend `landingtrain_attribution_test.go` so that only untouched, unregistered failures record, with the guard that the verdict is identical whether or not the store fails.
- [ ] `sdlc affected missed list|dismiss` (`affected_missed_cmd.go`), registered as a new `affected` command group in `main.go`'s workflow list (none exists today; `missed` is its only child), with its Long help from the embedded helptext, help text and a RECOVERY block. Strategy: a command test over a temp store for list, covered-prune, prune-unavailable, dismiss and an unknown ID.
- [ ] atlas: the curation loop (miss → declare in affected.json, traceability or a sidecar's `reads` → `list` prunes it), in `ci-merge-check.md`, with the `atlas/index.md` link if a new file is added. A `workshop/lessons.md` entry, if review finds a class.
- [ ] `sdlc close --issue 616`

## Revisions

### 2026-10-24: M1 as built

- Traceability files are read with a new `gitx.TreeFilesAt(ctx, dir, rev, paths)`, not `TreeFilesSelected`: a root-level file would otherwise need a full-tree listing, since a "." prefix is refused. Both share one `treeFiles` body, so the listing, batch and limits are one implementation.
- `ParseTraceability(data)` takes the bytes alone; `Compile` adds the file's path to its error.
- The `Owned` property test (`TestOwnedProperty`) compiles every source, traceability included, and checks the traceability paths too (M1 review).
- A bare trailing `---` (an empty document) is tolerated; any second document with content still refuses (M1 review).
- The drop-cost guard (`scripts/test-landing-drop-cost.py`) builds with `-coverpkg` over landingline and jsonstrict, and counts `jsonstrict/object.go` as the line's own, because the codec it requires moved there.

### 2026-10-24: plan-quality round 1

- PQ-1: decoding into a `yaml.Node` skips yaml.v3's duplicate-key check. The class is any decode-time check the walk assumed from the library. The walk now owns every refusal through one `mappingPairs` helper, and duplicate-key seeds are in the fuzz corpus. No other library behavior is assumed: aliases, merges, tags and documents were already checked by the walk.
- Minor: the Plan's test rows are compressed to one strategy line per risky function.
