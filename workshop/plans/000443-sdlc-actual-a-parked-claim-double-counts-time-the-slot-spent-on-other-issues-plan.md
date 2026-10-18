---
collaboration-mode: discussion
---

# Parked-claim actual exclusivity implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution. The SDLC close gate owns the fresh-eyes review.

**Goal:** Time already adopted for another issue from the same slot cannot be adopted again by an issue whose claim stayed open while that work happened.

**Architecture:** The other issue's numeric adopted actual and completion-pinned measurement inputs are authority. Replay those inputs through the existing active-time engine, gather runs credited to that issue by slot, and subtract their intervals from the candidate's runs before candidate attribution. Refuse a numeric result if a potentially overlapping adopted actual cannot be replayed or bound to a slot. Keep branch-scoped commits unchanged: unrelated main commits are deliberately invisible under #270 and cannot safely arbitrate this overlap.

**Tech stack:** Go, tracker cards and completion ledger, `activetime`, existing transcript replay.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| `CreditedInterval` (issue, slot, start, end) | `cmd/sdlc/internal/activetime/` | new |
| Activity run exclusion | `cmd/sdlc/internal/activetime/` | new |
| Segment source identity | `cmd/sdlc/internal/activetime/segment.go` | modified |

`CreditedInterval` is derived from a replayed run with positive allocation to an adopted issue. One adopted issue has many intervals; several intervals can overlap a candidate run. The subtraction is a pure union/intersection, performed after candidate ownership clipping and queue subtraction, before candidate attribution. Split allocations protect the whole run rather than guessing a fractional wall-time location; this can undercount, and the diagnostic says so. It cannot double-count. Segment source identity preserves the existing run's source through attribution so its pinned slot can be read without inferring it from a path. No new persisted entity is created (ARCH-FUNERAL).

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Adopted-overlap loader | `cmd/sdlc/actualinputs.go` or a focused adjacent file | new | tracker card scan, completion ledger and retained replay inputs |
| Actual measurement wiring | `cmd/sdlc/actualmeasurement.go` | modified | one exclusion policy for preview, close and correction |

The loader reads the already fetched tracker snapshot, selects cards with numeric adopted actuals and same-slot engagement overlapping the candidate window, and resolves each issue's immutable completion/correction pins. It must fail closed on unreadable cards, missing retained inputs, missing source-slot binding, or unavailable replay evidence when overlap remains possible. It need not read every issue's transcript: filter by slot and engagement first. Tests use the existing stateful temporary Git/tracker fixture and real small transcript fixtures, not call-count mocks (ARCH-MOCK). Replay uses the saved source byte prefixes so later transcript growth does not move historical intervals.

## Operating and recovery rules

- A claim alone does not reserve time; only positive, replayed allocation in an **adopted** numeric actual protects an interval. `N/A`, unclaimed, and other-slot issues do not protect it.
- For each candidate source slot, union all protected intervals from other adopted issues in that slot. Subtract once, even if protected intervals overlap. Name issue ID, slot, interval and excluded minutes in the report. A whole-run protection of a fractional allocation is explicitly conservative.
- If an adopted issue with potentially overlapping engagement lacks a trustworthy replay or slot identity, return unknown coverage and refuse numeric adoption with the issue ID and reason. Do not call zero a measurement. A historical preview of an already adopted issue replays its pinned generation; only a new close/correction candidate resolves today's other adopted issues, so a later adoption cannot rewrite past evidence (ARCH-STATE).
- No network write, extra artifact or long-lived cache is added. The input is a bounded issue-card scan followed by replay only for potentially overlapping adopted issues; tests count replays for N and 2N unrelated cards to defend filtering cost (ARCH-CONSTRAINTS). Existing completion snapshots remain the single durable source (ARCH-DRY).

## Tasks

### 1. Reproduce the defect and pin refusal behavior

**Files:** `cmd/sdlc/actualmeasurement_test.go`, `cmd/sdlc/actualinputs_test.go`, and a focused new `cmd/sdlc/internal/activetime/*_test.go`.

- [ ] Write a #636-shaped fixture with one slot claimed on A, then two completed/adopted issues B and C in that same slot, with transcript work during B/C and no A work. Assert A's present measurement double-counts B/C (red).
- [ ] Add countercases for an adopted issue in another slot, an `N/A` actual, absent legacy replay evidence, and a replayed prefix after transcript append. Confirm the intended pre-change assertions fail for the defect, not fixture setup.
- [ ] Add a pure interval test: overlap and touching boundaries, union of two adopted issues, and a fractional-credit run whose full interval is protected.

### 2. Build the exclusion core and loader

**Files:** `cmd/sdlc/internal/activetime/segment.go`, `compute.go` and a focused exclusion file; `cmd/sdlc/actualinputs.go` or a focused adjacent file; `cmd/sdlc/actualmeasurement.go`.

- [ ] Carry source on attributed segments, derive positive-credit intervals with pinned source slots, and implement pure same-slot subtraction and its excluded-minute report. Run targeted `go test ./cmd/sdlc/internal/activetime -run 'Test.*(Exclusion|Claim)' -count=1` to green.
- [ ] Reuse tracker `Records.All`, `EngagementHistory`, completion/correction binding and `resolveOwnedReplay` to enumerate overlapping adopted issues, replay their saved inputs, and supply protected intervals to `activetime.Compute`. Fail closed when relevant evidence is unknown. Keep the current issue out of its own exclusion set.
- [ ] Wire the same loader through `computeOwnedActualFor` for new preview/close/correction candidates while preserving pinned historical replay. Run `go test ./cmd/sdlc -run 'Test.*(Actual|Parked)' -count=1 -timeout 60m` to green.

### 3. Verify and land

**Files:** issue Log and any affected `atlas/` map.

- [ ] Update the issue Log with measured exclusions and the conservative-overlap tradeoff. Map the changed actual-time flow in the relevant atlas page.
- [ ] Run `make -k`, push the reviewed head, then `sdlc local-check run --name suite`; inspect its recorded result.
- [ ] Run `sdlc close --issue 443 --verified '<specific checks and output>'`, answer blocking findings, and merge on SHIP. Report the PR and exact merge commit to the hub.

## Revisions

- 2026-10-18 11:12 PDT — The integration fixture uses one adopted issue plus a later transcript append to test the pinned-replay boundary. The pure interval tests separately cover a union of two adopted issues and different slots. The same `computeActualForPurpose` path supplies preview, close and correction candidates, so the integration assertion is made at that shared seam.
