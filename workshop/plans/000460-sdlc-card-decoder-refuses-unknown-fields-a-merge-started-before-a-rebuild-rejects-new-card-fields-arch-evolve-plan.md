---
collaboration-mode: discussion
---

# #460 additive persisted-record compatibility plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution. This is one issue branch; the SDLC gates own the fresh-context reviews.

**Goal:** A long-running sdlc binary can read and safely rewrite a tracker card that contains fields introduced by a newer binary, without losing those fields.

**Architecture:** Keep card frontmatter as the authoritative YAML document and patch known fields within it. Preserve unknown members, including nested ones, on every card rewrite. Relax the read-only test-stats decoder only for additive unknown members; keep every known field and safety-critical identity strict. Audit the landing and #449 sidecar paths against the same rule.

**Tech stack:** Go, `go.yaml.in/yaml/v3`, `encoding/json`, existing issue and landingline packages.

## Core concepts

| Pure entity | Lives in | Status |
| --- | --- | --- |
| Card document and field spans | `cmd/sdlc/internal/issue/card.go` | modified |
| `mergeUnknownCardNodes` | `cmd/sdlc/internal/issue/cardunknown.go` | new |
| Suite attempt decoder | `cmd/sdlc/teststats_store.go` | modified |

The card document owns the original frontmatter. `mergeUnknownCardNodes` receives the original node, the known-field replacement, and the known schema; it carries only unknown keys forward when a known record remains present. If a known record is deliberately removed, its unknown children leave with that record. The helper is pure and shared by claimant and transaction-envelope setters (ARCH-DRY, ARCH-PURE). The attempt decoder owns validation of the fields an older binary understands; extra metadata does not change the exact SHA, exit, timeout, or log hash (ARCH-EVOLVE).

| Integration point | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| Tracker card update | `cmd/sdlc/internal/tracker/` and `cmd/sdlc/internal/issue/` | existing | Git tracker CAS and YAML source bytes |
| Attempt record reader | `cmd/sdlc/teststats_store.go` | existing | Local 30-day JSON record files |

No new external service or durable file family is introduced. Tests use the existing temporary Git tracker and local file fixtures; the production CAS remains unchanged (ARCH-MOCK, ARCH-FUNERAL). The per-card parse and rewrite remain linear in card bytes; the attempt reader keeps its 1 MiB per-record and 64 MiB scan budget (ARCH-CONSTRAINTS).

## Contract and audit

Card decode accepts unknown keys at the top level, under `claimant`, and recursively under `tracker` records, while rejecting duplicate keys anywhere, YAML aliases, malformed known values, missing required known fields, and unsupported `tracker.version`. `ParseCard` itself calls the nested `cardEnvelope` validation after the version check, before returning a usable card; no tracker writer can publish a shape that the read boundary did not validate. Existing `parseClaimant` validates the known claimant fields at that same boundary. Unknown top-level frontmatter remains byte-identical through setters and is not copied from details into an authoritative tracker card. Unknown nested members survive each known-field rewrite; a deliberate deletion of a complete known record may delete its own children. A new field that changes the meaning of an old one requires a version boundary, not this compatibility rule (ARCH-EVOLVE, ARCH-SECURE).

The card rewrite table is: scalar/title setter → preserve untouched spans; claimant setter → replace known claimant children and retain unknown claimant children; tracker setter → retain unknown children at every retained mapping depth. Invalid card → refuse before Git write; CAS drift → existing retry path rereads the whole newer card and applies again; interrupted push → existing tracker receipt reconciles. The user-visible state and recovery remain the existing tracker states; this change adds no new state (ARCH-ORDER).

Landing events: `landingline.readLog` tolerates unknown JSON members and `sweep` retains original lines, so add an explicit retention regression but no production change. Test-stats v2: `decodeSuiteAttempt` currently refuses unknown keys; let unknown keys through its token-level duplicate check and JSON struct decoder while validating known key spelling, types, required fields, nested known facts and the exact-head safety fields. Because Go JSON struct decoding is case-insensitive, a key whose case-folded name matches a known key but whose spelling differs must be rejected at each object level. The `units` and `known_flakes` object validators apply the same rule. These are read-only records, so there is no attempt rewrite; `publishSuiteAttempt` creates a new attempt with a new ID. #449 readiness sidecars: its current branch uses tolerant JSON decode but writes a complete vouch on update. Hub directed its owner, ariadne:2, to preserve unknown members there before landing; #460 will verify that branch's result and log it, without making a competing edit. The audit result goes into the issue Log (ARCH-PURPOSE).

Excluded: no semantic schema migration, tracker transaction redesign, or competing edit on #449's branch. Those do not follow from the additive-field incident; changing existing field meaning would require its own explicit version design.

## Adversarial function strategies

| Function | Input class and mechanical guard |
| --- | --- |
| `ParseCard` / `cardEnvelope` | Unknown members at every mapping depth remain readable; wrong type or missing required **known** nested member and duplicate key at any depth must fail before a card is returned. |
| `mergeUnknownCardNodes` | Generate nested unknown YAML trees with distinct scalar, sequence and mapping leaves, rewrite a separate known leaf, then compare unknown subtrees before/after; known keys removed or replaced by the setter cannot be resurrected. |
| `SetCardClaimant` / `updateEnvelope` | Read a newer card, rewrite a known member, and assert every unrelated unknown subtree survives; run the same chain through tracker CAS and assert the written blob retains it. |
| `decodeSuiteAttempt` / `validateSuiteUnitsJSON` / known-flake validator | Unknown additive keys pass at root and nested objects; exact known names retain their type and required-field checks, duplicate keys fail, and case-folded aliases of known names fail before struct decode. |
| `landingline.sweep` | A retained event with extra JSON fields is rewritten as the same original line, while expired lines are removed. |

## Tasks

### 1. Red regressions for cards

**Files:** `cmd/sdlc/internal/issue/card_test.go`, `cardset_test.go`, `claimant_test.go`, `handoff_test.go`, and tracker update integration test.

- [ ] Build a card with an unknown top-level `flake_tests` surrogate, an unknown claimant key, and unknown nested tracker keys (including a nested mapping). Assert old-shape `ParseCard` accepts it. Feed it through `SetCardField`, `SetCardClaimant`, `SetCardHandoff`, `SetCardCompletion`, and tracker CAS; assert each unknown subtree is retained while the known change lands.
- [ ] Add negative rows for wrong known-field types, missing required values, duplicate unknown or known keys at every depth, and unsupported tracker version. Test that `ParseCard`, not a later setter alone, rejects a malformed known `tracker.completion` or `tracker.release.by`. Run `go test ./cmd/sdlc/internal/issue ./cmd/sdlc/internal/tracker -run 'Card|Tracker' -count=1` and confirm the new positive tests fail before implementation.

### 2. Implement lossless card compatibility

**Files:** `cmd/sdlc/internal/issue/card.go`, `claimant.go`, `handoff.go`, and a focused helper file in that package.

- [ ] Remove only the top-level unknown-field refusal in `ParseCard`; invoke `cardEnvelope` there for nested known-field validation after the version check. Teach claimant parsing to ignore unknown members after the global recursive duplicate-key check. Make claimant rendering preserve unknown children from the source node.
- [ ] Implement `mergeUnknownCardNodes` and update the tracker-envelope rewrite so an old reader's new known record is combined with source unknown children at retained mapping depths. Keep known-key removal intentional and known-field validation intact. Exercise direct generated-tree tests as well as each setter family, then run the package tests.

### 3. Test-stats and audit regressions

**Files:** `cmd/sdlc/teststats_store.go`, `teststats_store_test.go`, `cmd/sdlc/internal/landingline/log_test.go`, issue Log.

- [ ] Add a newer attempt with unknown top-level and nested metadata members and prove an older read-only decoder can still use the validated exit/SHA/log hash. Assert duplicate keys, case-folded known aliases and wrong known-field types fail at root, duration units and known flakes. Relax only unknown-key refusal; do not weaken versioned outcome semantics.
- [ ] Add a landing-event retention test with an extra member and assert the original JSON line survives a sweep. Check ariadne:2's #449 fix before it lands; record its test and code evidence in the audit.

### 4. Verify and land

- [ ] Run targeted tests, `make -k`, `git diff --check`, push, and `sdlc local-check run --name suite` on the exact head. Record the attempt and log hash.
- [ ] Complete `sdlc close` review and evidence, stage a short ariadne:0 live check only after hub grant, clear it back to main and verify binary revisions, then `sdlc pr` and `sdlc merge --yes` on SHIP. Report PR and real merge hash to the hub.

## Revisions

- 2026-10-19: Initial plan from the #460 incident and the current reader/writer audit. No implementation has begun.
- 2026-10-19: Hub approved and assigned readiness sidecar compatibility to #449's current owner before landing; #460 verifies that result instead of editing its branch.
- 2026-10-19: PQ-1/PQ-2 exposed missing decode-boundary validation and unnamed adversarial function checks. Put `cardEnvelope` under `ParseCard`, name the pure merge helper and its generated-tree property, and guard JSON canonical known names at every object level. State non-goals (PQ-3).
