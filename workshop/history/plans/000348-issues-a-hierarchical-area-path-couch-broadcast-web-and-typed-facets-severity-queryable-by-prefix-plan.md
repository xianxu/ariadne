# Repo-Defined Enums: Issue Area and Severity Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Issues carry a hierarchical `area` (one or more `/`-paths whose allowed values each repo declares) and a shared `severity` enum, set through the vocabulary model and card setters, validated wherever they are written and in the validation gates, and queryable by `issue list --area <prefix> --severity <level>`.

**Architecture:** The shared model (`construct/vocabulary/issue.cue`) gains two concrete sections: `enums` (shared closed enums: `severity`) and `repoEnums` (a field's type and syntax plus the path of the repo's declaration: `area`). It also gains two card fields whose kinds name those enums. A pure package `cmd/sdlc/internal/repoenum` parses a repo's `.sdlc/vocabulary.json`, decides membership (the ancestor rule), and matches segment prefixes. Card parsing checks shape only. Writers (`issue new`, `set-area`, `set-severity`) and gates (`issue validate`, the #124 instance gate) check membership against main's declaration.

**Tech Stack:** CUE (`construct/vocabulary`, `vocabulary export` → `pkg/vocab/issue.json`, `make vocab-embed`), Go (`cmd/sdlc`, cobra, yaml.v3).

**Approved design:** hub (ops:0), 2026-10-13. The decisions:
- The declaration is `.sdlc/vocabulary.json`, under the local-gates guard: the operator governs each repo's small taxonomy.
- Severity is `critical | high | medium | low`.
- area and severity are card fields.
- ariadne ships no starter set. An **absent file is the empty declaration**, so this change adds no `.sdlc/` file and stays off guarded paths.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `enums` / `repoEnums` model sections; `#Severity`, `#AreaPath` | `construct/vocabulary/issue.cue` | new |
| `card.fields` rows `area` (kind `repo-enum`), `severity` (kind `enum`) | `construct/vocabulary/issue.cue` | modified |
| `vocab.CardField.Enum`, `IssueModel.Enums`, `IssueModel.RepoEnums`, `RepoEnumSpec` | `pkg/vocab/vocab.go` | modified/new |
| `repoenum.Declaration`, `Parse`, `ValidPath`, `Allowed`, `MatchPrefix`, `CheckValues` | `cmd/sdlc/internal/repoenum/repoenum.go` | new |
| `validateCardScalar` kinds `enum`, `repo-enum` (shape only) | `cmd/sdlc/internal/issue/card.go` | modified |
| `issue.ScaffoldSpec.Area/Severity` + `Render` | `cmd/sdlc/internal/issue/scaffold.go` | modified |
| `issueListFilter` (`matches(IssueState)`) | `cmd/sdlc/issuelist.go` | new |
| `IssueState.Area/Severity` | `cmd/sdlc/state.go` | modified |

- **Model sections.** In CUE, concrete data reaches the JSON export and `#` definitions do not, so the enums are concrete lists with definitions derived from them, the same way `categories`/`#Status` work:
  ```cue
  enums: severity: ["critical", "high", "medium", "low"]
  #Severity: or(enums.severity)
  repoEnums: {
      declarationPath: ".sdlc/vocabulary.json"
      fields: area: {syntax: "path", multi: true, maxDepth: 4}
  }
  #AreaPath: =~"^[a-z0-9][a-z0-9-]*(/[a-z0-9][a-z0-9-]*){0,3}$"
  ```
  Card rows: `{name: "area", kind: "repo-enum", enum: "area", setter: "sdlc issue set-area"}` and `{name: "severity", kind: "enum", enum: "severity", setter: "sdlc issue set-severity"}`. `#Issue` gains `severity?: #Severity | null` and `area?: [...#AreaPath] | null` so the cue instance check (the `vocabulary validate-instance` path) catches shape. A law asserts every card row of kind `enum` names a key of `enums`, and every `repo-enum` row names a key of `repoEnums.fields`.
  - **DRY rationale:** one source for the severity values, the path syntax and the declaration location. Go reads them from `vocab`, and no Go constant restates them. The path regex appears once in CUE; Go builds its check from `maxDepth` and the segment rule it exports (`segment: "[a-z0-9][a-z0-9-]*"` is added to the `area` spec, so the Go regex derives from the model).
  - **Future extensions:** a second repo enum (e.g. `component`) is one `repoEnums.fields` row, one card row (or a details field), and one key in the repo's declaration, with no new Go.
- **repoenum.Declaration** (`{Version int; Enums map[string][]string}`) with **Parse(data []byte, spec map[string]vocab.RepoEnumSpec)**. Parsing is strict:
  - unknown top-level fields, a version other than 1, a key that the model does not declare, and a value failing its field's syntax are all errors;
  - duplicates are refused;
  - empty input is the empty declaration.
  - **Allowed(field, value)** is true iff the value is declared or is a strict ancestor of a declared value (by `/` segments).
  - **MatchPrefix(prefix, value)** is true iff `value == prefix` or `strings.HasPrefix(value, prefix+"/")`.
  - **CheckValues(decl, field, values) []string** returns the undeclared ones.
  - Tests in `repoenum_test.go` run without IO.
  - **Relationships:** 1 declaration per repo, N enum keys, each with N values.
- **Card kinds.**
  - `enum`: a scalar string that is a member of `vocab.Issue().Enums[field.Enum]`, or null.
  - `repo-enum`: a YAML sequence (flow or block) of strings, each passing the model's path syntax, with no duplicates; or null; and at most 8 entries.
  - **Membership is never checked in `ParseCard`.** A value later removed from a declaration must not make cards unparseable, which would freeze the tracker. Membership is a write-time and gate-time check.
- **issueListFilter**: built from `--status`, `--area` (repeatable: an issue matches if any of its areas matches any prefix), and `--severity` (comma list). It is a pure predicate on `IssueState`.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `loadRepoEnums(git, rev)` | `cmd/sdlc/repoenums.go` | new | `git ls-tree` / `git show <rev>:.sdlc/vocabulary.json` |
| `newIssueSetAreaCmd`, `newIssueSetSeverityCmd` | `cmd/sdlc/cardsetters.go` | new | `runCardUpdate` (tracker CAS) |
| `issue new --area/--severity` | `cmd/sdlc/issue.go` | modified | scaffold → card split |
| `checkRepoEnumsFn` in the #124 gate | `cmd/sdlc/validategate.go` | modified | `loadRepoEnums` at the gate's base |
| `issue validate` card + file checks | `cmd/sdlc/issue.go` | modified | tracker snapshot + files |

- **loadRepoEnums(git, rev)**: reads the declaration from `rev`'s tree. Absent gives the empty declaration, a git error is returned, and a parse error is returned with the file named. **Which rev:**
  - writers in tracker repos use `origin/main` (the tracker env has fetched; the same trust rule as #345, so a branch cannot declare its own values);
  - the #124 gate uses its `base` (the target main the caller already pinned);
  - `issue new` in a legacy repo uses `origin/main` when it exists, else `HEAD`.
  - **Injected into:** setters and `issue new` via a package var seam (`readRepoEnumsFn`), and the gate via `checkRepoEnumsFn`, so tests use a fake reader.
- **Setters.**
  - `sdlc issue set-area --issue N <path>...` replaces the list; `--clear` sets null.
  - `sdlc issue set-severity --issue N <level>`; `--clear` sets null.
  - Both go through `newCardSetterCmd` → `runCardUpdate` → `issue.SetCardField` (shape) after a membership check against `loadRepoEnums(origin/main)`. A refusal names `.sdlc/vocabulary.json` and the local-gates label.
  - Legacy repos refuse as the other setters do (`errLegacyOnly`, pointing at the frontmatter).
- **`issue new --area a/b --area c --severity high`**: membership is checked before anything is reserved. `Render` emits `area: [a/b, c]` and `severity: high`, and `SplitCardWithFormat` copies them onto the card (they are card fields).
- **Gates.** `validateChangedInstances(base, head, …)` additionally loads the declaration at `base` once (lazily, only if an issue instance changed) and, for each changed issue file, refuses undeclared `area` values and a `severity` outside the model. In tracker repos, details carry the card's values through the mirror, so the check sees them. `issue validate` runs the same check per file and, in tracker repos, over every card in the snapshot.

### Operating envelope (ARCH-CONSTRAINTS)

Interactive CLI reads plus one declaration read per command. A declaration is a few dozen paths (an operator-curated "small set"). `issue list` filtering is O(issues × areas), with hundreds of issues and ≤8 areas each, so no index is needed. Card area lists are capped at 8 entries by the shape check. N/A beyond that: no concurrency, network or disk growth beyond existing tracker writes.

### ARCH-SECURE

- **Boundaries:**
  - The declaration is untrusted repo content. It is read from main's tree (never the branch under change), parsed strictly into a typed value, and an invalid file refuses writers and gates with the file named.
  - Card values arrive from other slots through the tracker. `ParseCard` shape-checks them into typed lists, and membership is re-checked at gates.
  - Hand-edited details frontmatter (legacy repos) is checked by the gates.
- **No credentials touched.**
- **Governance:** `.sdlc/` is under the local-gates guard, so adding an area needs the operator's label. That is the point, because an agent cannot drift the taxonomy unreviewed.

### ARCH-ORDER

Holds no state between events. Each command reads the declaration and the card once, and the card write is the existing tracker CAS (a concurrent writer makes the CAS fail, and the setter's existing retry or refusal applies). The race that matters is a declaration that changes between a write and a gate (a value removed from main after it was set). The gate then refuses that issue's next change until its area is updated, which is the intended governance: removal is a decision the operator makes with the label, and stale values surface at the next touch rather than silently. Old cards stay parseable.

### ARCH-FUNERAL

Creates nothing durable beyond two optional fields on existing cards and details, which live and die with the issue (archived with it). The declaration is one operator-owned file per repo, with no growth path other than the operator's edits.

### ARCH-PURPOSE

The issue's purpose is that the TL can record and query which product feature an issue belongs to and how severe it is, with each repo's areas governed. Delivered:
- the model contract;
- the declaration format and loader;
- writers;
- both gates;
- list filters;
- show and `state` JSON output;
- help;
- the atlas.

Not here, and intentionally so: each repo's actual area list. That is the operator's per-repo declaration under the label, since ariadne ships none per the hub's decision.

### ARCH-MOCK

Seams are `readRepoEnumsFn` (git show at a rev) and the existing tracker env, tested through `newTrackerRepo`'s real git with a bare origin, which is the established stateful fixture. No new external dependency.

---

## Tasks

Single pass (one `sdlc close`), plain checkboxes.

### Task 1: Model

**Files:** `construct/vocabulary/issue.cue`, `construct/vocabulary/testdata/` (new invalid fixtures), `construct/vocabulary/vet_test.sh`, `pkg/vocab/vocab.go`, `pkg/vocab/issue.json` (regenerated), `pkg/vocab/card_test.go`, `pkg/vocab/vocab_test.go`

- [ ] Write failing tests:
  - `pkg/vocab`:
    - `TestEnums`: `Issue().Enums["severity"]` is `[critical high medium low]`.
    - `TestRepoEnums`: `area` has syntax `path`, multi, maxDepth 4, a segment regex, and declarationPath `.sdlc/vocabulary.json`.
    - Update `TestCardOwnershipDiscovery` with the two rows and their `Enum`.
  - `vet_test.sh`:
    - a fixture whose card row names an undeclared enum fails `vocabulary vet`;
    - a fixture whose `#Issue` instance has `severity: urgent` fails `validate-instance`;
    - `area: [Couch]` fails.
- [ ] Edit `issue.cue` (sections + laws + `#Issue` fields + card rows), then run `vocabulary vet` and `make vocab-embed`. Add `Enum string` to `CardField`, plus `Enums map[string][]string` and `RepoEnums RepoEnumModel` to `IssueModel`.
- [ ] Run `go test ./pkg/vocab/... && sh construct/vocabulary/vet_test.sh` → PASS. Commit `#348: model: shared enums, repo enums, area and severity card fields`.

### Task 2: repoenum core

**Files:** create `cmd/sdlc/internal/repoenum/repoenum.go`, `repoenum_test.go`

- [ ] Table tests:
  - **Parse:**
    - valid;
    - empty → empty;
    - an unknown key `component` → error naming the key;
    - version 2;
    - `Couch/x` (uppercase) → syntax error;
    - depth 5;
    - a duplicate;
    - trailing data;
    - an unknown top-level field.
  - **Allowed:**
    - declared;
    - ancestor `couch` of `couch/broadcast/web` → true;
    - a sibling `couch/session` → false;
    - a non-segment prefix `couch/broad` → false;
    - a descendant `couch/broadcast/web/x` → false.
  - **MatchPrefix:** `couch/broadcast` vs `couch/broadcaster` → false.
  - **CheckValues:** returns the undeclared values in order.
- [ ] Implement, deriving the path regex from the `vocab` spec.
- [ ] Run → PASS. Commit `#348: repoenum: declaration, ancestor rule, segment prefix`.

### Task 3: Card kinds (shape only) — superseded in part by Revision R1 (sequence class)

**Files:** `cmd/sdlc/internal/issue/card.go`, `card_test.go`, `cardset_test.go`

- [ ] Tests:
  - `ParseCard` accepts `severity: high`, `severity:` (null), `area: [couch/broadcast/web, sdlc]`, a block sequence, and `area:` (null).
  - It refuses `severity: urgent`, `area: couch` (scalar), `area: [Couch]`, a duplicate entry, 9 entries, and `area: [a, [b]]`.
  - It **accepts** `area: [never/declared]` (shape only: membership is not ParseCard's job).
  - `SetCardField(raw, "area", "[a/b, c]")` round-trips, and `SetCardField(raw, "severity", "")` gives null.
- [ ] Implement: handle `repo-enum` before the scalar guard (a sequence node) and `enum` in the switch, both reading their sets and syntax from `vocab`.
- [ ] Commit `#348: card: enum and repo-enum kinds, shape-only`.

### Task 4: Loader + writers — see Revision R1 (publish check, declaration rev)

**Files:** create `cmd/sdlc/repoenums.go` (+ test); modify `cmd/sdlc/cardsetters.go`, `cmd/sdlc/issue.go` (new flags, register setters), `cmd/sdlc/internal/issue/scaffold.go` (+ `scaffold_test.go`)

- [ ] Tests:
  - `TestLoadRepoEnums`: absent → empty; valid; invalid → error naming `.sdlc/vocabulary.json`; read at the named rev, not the work tree.
  - `TestRender_AreaSeverity`: byte-stable output, with the fields omitted when unset (existing `TestRender_ByteStable` unchanged).
  - Through `newTrackerRepo` with `readRepoEnumsFn` faked:
    - `TestIssueSetArea_Declared` updates only `area`.
    - `TestIssueSetArea_Undeclared` refuses, names the file, and publishes nothing.
    - `TestIssueSetArea_Ancestor` passes.
    - `TestIssueSetArea_Clear` works.
    - `TestIssueSetSeverity` covers valid, invalid (listing the model's values) and clear.
    - `TestIssueNew_AreaSeverity`: values land on the card; an undeclared one refuses before any reservation.
    - Legacy setters refuse with the frontmatter pointer.
- [ ] Implement.
- [ ] Commit `#348: writers validate area and severity against main's declaration`.

### Task 5: Gates

**Files:** `cmd/sdlc/validategate.go` (+ `validategate_test.go`), `cmd/sdlc/issue.go` (`validateIssueFull` / `issue validate`) (+ test)

- [ ] Tests:
  - Instance gate with a fake `diffNameStatusFn` and `checkRepoEnumsFn`:
    - a changed issue file with an undeclared area → refusal naming the file and the declaration;
    - an invalid severity → refusal;
    - a declared area → pass;
    - no changed issue file → the declaration is never read;
    - an invalid declaration → fail closed.
  - `issue validate` in a tracker repo flags a card with an undeclared area (and a details file in a legacy repo).
- [ ] Implement.
- [ ] Commit `#348: gates check area and severity membership`.

### Task 6: Query

**Files:** create `cmd/sdlc/issuelist.go` (filter) + test; modify `cmd/sdlc/issue.go` (flags, output), `cmd/sdlc/state.go` (`IssueState` fields, populated from records), `cmd/sdlc/issue_test.go`

- [ ] Tests:
  - `TestIssueListFilter`:
    - status;
    - an area prefix (`couch/broadcast` matches `couch/broadcast/web`, not `couch/broadcaster`);
    - repeated areas (OR);
    - severity lists;
    - combined (AND across flags).
  - `--severity urgent` → refuses, listing the model values.
  - `--area Bad/Path` → syntax refusal.
  - Output gains `AREA` and `SEV` columns, shown only when any listed issue has them, so existing output stays byte-stable.
  - The `sdlc state --json` schema carries `area`/`severity` (omitempty).
- [ ] Implement.
- [ ] Commit `#348: issue list filters by area prefix and severity`.

### Task 7: Help, atlas, base layer

**Files:** `cmd/sdlc/helptext/issue.md` (fields + setters + list flags; the drift test pins it to the model), `atlas/workflow/vocabulary.md` (the repo-enum facility), `atlas/workflow/sdlc-binary.md` (setter rows), `atlas/index.md` if needed, `construct/generated` regeneration via `make weave` if the vocabulary freshness check requires it.

- [ ] Update the docs, run the help drift tests, and commit `#348: help and atlas for repo enums`.

### Task 8: Verify and close

- [ ] Run targeted suites, then, if load < 40 and the suite lock is free, the full `make test` (after merging origin/main).
- [ ] Manual check in a scratch tracker repo is covered by the `newTrackerRepo` e2e tests. Record that in `--verified`.
- [ ] `sdlc close --issue 348 --verified '…'` → `sdlc pr` → CI green → `sdlc merge --yes` → report to the hub.

## Revisions

### R1 — 2026-10-13, plan-quality round 1

**PQ-1 (Important), the class "card helpers that assume scalar nodes".** The enumeration below comes from every consumer of `vocab.Issue().CardFields()` plus `sameCardValue`. Each one is fixed in Task 3 and tested with a flow-sequence `area`:

| Site | Problem | Fix |
|---|---|---|
| `issue/card.go` `sameCardValue` (used by `mirror.go:79,99`, `migrate.go:68`) | compares `.Value`, which is `""` for any sequence, so all lists compare equal | decode-compare `SequenceNode` like `MappingNode` |
| `issue/frontmatter.go` `SetField`/`dropBlockContinuation` | does not drop column-one `- item` lines | treat `- ` lines (any indent) directly under the key as continuation |
| `tracker/records.go` `IssueRecord.Field` | returns the raw line text | add `IssueRecord.List(name) ([]string, bool)` that YAML-decodes the field; `IssueState.Area` uses it |
| `issue/migrate.go:160+` legacy→card migration kind switch | unknown kinds fall through | `enum`/`repo-enum`: carry a valid-shape value; else drop it with an `infer(...)` note |
| `issue/card.go` `ParseCard` | scalar-only guard | `repo-enum` handled before the scalar guard |

**Card shape is the flow form only.** A card `area` must be a single-line flow sequence (`area: [a/b, c]`) or null, which matches `SetCardField`'s one-line contract. Block form is refused on cards. Hand-edited legacy frontmatter (checked by the gates) accepts either form, because it is YAML-decoded.

Tests:
- `TestSameCardValueSequences`;
- `TestSetFieldDropsBlockSequence`;
- `TestRefreshMirrorAreaChange`: a card area change updates the details mirror, and an unchanged area is a no-op;
- `TestMigrateCarriesAreaSeverity`;
- `TestRecordList`.

**PQ-2 (Minor), test strategy instead of enumerated cases.** Tasks name test functions. Strategy per risky function:
- `repoenum.Parse`: table tests plus `FuzzParse`. Any input either errors or yields a declaration whose values all pass `ValidPath`, and a JSON round-trip of a parsed declaration re-parses equal.
- Card `repo-enum` shape: `FuzzRepoEnumCardShape` over arbitrary YAML values for `area:`. ParseCard never panics, and accepted values are flow sequences of valid paths.
- `Allowed`/`MatchPrefix`: property tests. Every declared value's ancestors are allowed, and `MatchPrefix(p, v)` implies `v` has `p`'s segments as a prefix.
- The rest are example tests named in each task.

**PQ-3 (Minor), coverage of write paths.**
- Tracker card values are written only by `issue new` and the two setters, all of which check membership. The #124 gate checks changed details files, whose mirror carries the card values.
- `issue publish` now also checks the published details' `area`/`severity` against main's declaration before publishing (the Revision's third write site), in `issuepublish.go`. Test: `TestIssuePublish_RefusesUndeclaredArea`.
- `issue validate` sweeps every card. That covers a declaration that shrank after a write: such values surface at the next validate or touch, by design (see ARCH-ORDER).

**PQ-4 (Minor), declaration rev in legacy repos.** The declaration is always read from main, never from HEAD or the work tree:
- `origin/main` when that ref exists;
- else the local `main` ref;
- else refuse ("no main to read .sdlc/vocabulary.json from").

Test: `TestLoadRepoEnums_NeverReadsHead` (a declaration only on the current branch is ignored).
