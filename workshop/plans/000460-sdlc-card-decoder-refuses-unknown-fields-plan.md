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
| Unknown-preserving nested YAML merge | `cmd/sdlc/internal/issue/` | new |
| Suite attempt decoder | `cmd/sdlc/teststats_store.go` | modified |

The card document owns the original frontmatter. A nested merge receives the original node, the known-field replacement, and the known schema; it carries only unknown keys forward when a known record remains present. If a known record is deliberately removed, its unknown children leave with that record. The helper is pure and shared by claimant and transaction-envelope setters (ARCH-DRY, ARCH-PURE). The attempt decoder owns validation of the fields an older binary understands; extra metadata does not change the exact SHA, exit, timeout, or log hash (ARCH-EVOLVE).

| Integration point | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| Tracker card update | `cmd/sdlc/internal/tracker/` and `cmd/sdlc/internal/issue/` | existing | Git tracker CAS and YAML source bytes |
| Attempt record reader | `cmd/sdlc/teststats_store.go` | existing | Local 30-day JSON record files |

No new external service or durable file family is introduced. Tests use the existing temporary Git tracker and local file fixtures; the production CAS remains unchanged (ARCH-MOCK, ARCH-FUNERAL). The per-card parse and rewrite remain linear in card bytes; the attempt reader keeps its 1 MiB per-record and 64 MiB scan budget (ARCH-CONSTRAINTS).

## Contract and audit

Card decode accepts unknown keys at the top level, under `claimant`, and recursively under `tracker` records, while rejecting duplicate keys anywhere, YAML aliases, malformed known values, missing required known fields, and unsupported `tracker.version`. Unknown top-level frontmatter remains byte-identical through setters and is not copied from details into an authoritative tracker card. Unknown nested members survive each known-field rewrite; a deliberate deletion of a complete known record may delete its own children. A new field that changes the meaning of an old one requires a version boundary, not this compatibility rule (ARCH-EVOLVE, ARCH-SECURE).

The card rewrite table is: scalar/title setter → preserve untouched spans; claimant setter → replace known claimant children and retain unknown claimant children; tracker setter → retain unknown children at every retained mapping depth. Invalid card → refuse before Git write; CAS drift → existing retry path rereads the whole newer card and applies again; interrupted push → existing tracker receipt reconciles. The user-visible state and recovery remain the existing tracker states; this change adds no new state (ARCH-ORDER).

Landing events: `landingline.readLog` tolerates unknown JSON members and `sweep` retains original lines, so add an explicit retention regression but no production change. Test-stats v2: `decodeSuiteAttempt` currently refuses unknown keys; let unknown keys through the token-level duplicate check and JSON struct decoder while validating known key spelling, types, required fields, nested known facts and the exact-head safety fields. These are read-only records, so there is no attempt rewrite; `publishSuiteAttempt` creates a new attempt with a new ID. #449 readiness sidecars: its current branch uses tolerant JSON decode but writes a complete vouch on update. Hub directed its owner, ariadne:2, to preserve unknown members there before landing; #460 will verify that branch's result and log it, without making a competing edit. The audit result goes into the issue Log (ARCH-PURPOSE).

## Tasks

### 1. Red regressions for cards

**Files:** `cmd/sdlc/internal/issue/card_test.go`, `cardset_test.go`, `claimant_test.go`, `handoff_test.go`, and tracker update integration test.

- [ ] Build a card with an unknown top-level `flake_tests` surrogate, an unknown claimant key, and unknown nested tracker keys (including a nested mapping). Assert old-shape `ParseCard` accepts it. Feed it through `SetCardField`, `SetCardClaimant`, `SetCardHandoff`, `SetCardCompletion`, and tracker CAS; assert each unknown subtree is retained while the known change lands.
- [ ] Add negative rows for wrong known-field types, missing required values, duplicate unknown or known keys at every depth, and unsupported tracker version. Run `go test ./cmd/sdlc/internal/issue ./cmd/sdlc/internal/tracker -run 'Card|Tracker' -count=1` and confirm the new positive tests fail before implementation.

### 2. Implement lossless card compatibility

**Files:** `cmd/sdlc/internal/issue/card.go`, `claimant.go`, `handoff.go`, and a focused helper file in that package.

- [ ] Remove only the top-level unknown-field refusal in `ParseCard`; retain validation of known fields and tracker version. Teach claimant parsing to ignore unknown members after the global recursive duplicate-key check. Make claimant rendering preserve unknown children from the source node.
- [ ] Update the tracker-envelope rewrite so an old reader's new known record is combined with source unknown children at retained mapping depths. Keep known-key removal intentional and known-field validation intact. Test each setter family, then run the package tests.

### 3. Test-stats and audit regressions

**Files:** `cmd/sdlc/teststats_store.go`, `teststats_store_test.go`, `cmd/sdlc/internal/landingline/log_test.go`, issue Log.

- [ ] Add a newer attempt with an unknown top-level and nested metadata member and prove an older read-only decoder can still use the validated exit/SHA/log hash. Assert duplicates and wrong known field types still fail. Relax only unknown-key refusal; do not weaken versioned outcome semantics.
- [ ] Add a landing-event retention test with an extra member and assert the original JSON line survives a sweep. Check ariadne:2's #449 fix before it lands; record its test and code evidence in the audit.

### 4. Verify and land

- [ ] Run targeted tests, `make -k`, `git diff --check`, push, and `sdlc local-check run --name suite` on the exact head. Record the attempt and log hash.
- [ ] Complete `sdlc close` review and evidence, stage a short ariadne:0 live check only after hub grant, clear it back to main and verify binary revisions, then `sdlc pr` and `sdlc merge --yes` on SHIP. Report PR and real merge hash to the hub.

## Revisions

- 2026-10-19: Initial plan from the #460 incident and the current reader/writer audit. No implementation has begun.
- 2026-10-19: Hub approved and assigned readiness sidecar compatibility to #449's current owner before landing; #460 verifies that result instead of editing its branch.
