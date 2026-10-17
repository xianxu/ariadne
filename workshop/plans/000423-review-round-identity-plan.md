---
collaboration-mode: discussion
---

# Review round identity implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3. This is a single atomic correction in the current issue worktree; execute using the existing session context and the binary-owned review.

**Goal:** Report the completion-bound/latest verdict with an honest ledger round and chronological history despite timestamp drift.

**Architecture:** Keep the existing sidecar parser and completion selection. Replace timestamp matching with a pure reviewed-commit matcher over typed ledger facts. Older entries can use positional correspondence only when the entire boundary sequence is compatible with every known head and at least one known head anchors it; all other ambiguous round identities stay explicitly unknown.

**Tech Stack:** Go, existing observe/gatestate types, markdoc fence processing, existing Git fixture/collector seam.

## Core concepts

| Name | Kind | Lives in | Status |
|---|---|---|---|
| ReviewRound | PURE | cmd/sdlc/internal/observe/checkpoints.go | new |
| assignReviewRounds | PURE | cmd/sdlc/internal/observe/checkpoints.go | new |
| Review | PURE | cmd/sdlc/internal/observe/types.go | modified |
| collectEvidence | INTEGRATION | cmd/sdlc/observe.go | modified |

ReviewRound is a ledger number plus the reviewed commit, not a timestamp. One boundary owns an ordered list; Review owns transient round/error/history output. assignReviewRounds is the single matching rule (ARCH-DRY, ARCH-PURE). collectEvidence reuses observeGit and gatestate.ParseSidecar, derives heads from finalized Reviewed or retained ReviewerRuns, and passes facts to the pure assembler; no new external calls or backend. Existing controlled Git repositories provide stateful integration fixtures; the earlier native issue-show 416 check is the live collector conformance check. No new persisted state or artifact family; these values die with the observation (ARCH-FUNERAL).

## Matching contract

- Stable identity is the head after `..` in the sidecar window and a ledger round's Reviewed, or retained review-run intent ReviewedHEAD for nonfinal rounds.
- If both complete per-boundary sequences have equal length, at least one ledger head is known, and every known head matches the corresponding entry, use their ordered ledger numbers, including legacy missing-head rounds.
- Otherwise match only commit groups whose entry/ledger multiplicities agree; duplicate heads resolve in their recorded order only with that complete group evidence. Missing or extra duplicate artifacts must not select an arbitrary round.
- With no ledger or ambiguous/missing matches, keep the independently read verdict/window/time and expose round unknown with a reason. Never substitute an ordinal for a ledger number.
- Earlier/later labels use entry indices relative to the selected entry. Numeric gaps, unknown numbers or timestamp drift cannot reverse chronology.
- All logic reads one immutable committed snapshot and carries no state between events (ARCH-ORDER). Runtime/memory scale linearly with the existing entries/ledger; no extra Git/network work (ARCH-CONSTRAINTS). Inputs are existing parsed artifacts; absent identity is unknown, not evidence of absence; no secrets are read or produced (ARCH-SECURE). Existing collector/seams and real Git fixtures remain the integration boundary (ARCH-MOCK). The entire timestamp-map/lookup/relation family is replaced together (ARCH-PURPOSE).

## Steps

- [x] Reproduce the timestamp drift counterexample with nonconsecutive rounds 4 and 6 using the current implementation; expected round 6 and earlier history, observed round 2 and reversed label.
- [ ] Add pure coverage for stable-head matching with timestamp drift, equal-count aligned legacy sequences, incompatible legacy sequences, absent ledger, duplicate-head ambiguity, and close/milestone chronology.
- [ ] Replace ArtifactFacts timestamp maps with typed ordered ReviewRound facts. Derive the reviewed head from finalized or retained review evidence in collectEvidence; retain current open-finding counts.
- [ ] Implement assignReviewRounds under the matching contract above; expose round errors independently of verdict read quality. Label history by entry indices. Remove old timestamp lookups and ordinal fallbacks.
- [ ] Extend the real collector/lifecycle fixture with nonconsecutive boundary ledger rounds and intentionally different timestamps; assert before and after archive/branch deletion. Update golden output and atlas for explicit unknown identity.
- [ ] Run `go test -trimpath ./cmd/sdlc/internal/observe -count=1` and focused collector/writer tests, then commit/push and run `sdlc local-check run --name suite`.
- [ ] Re-close through the binary, resolve all blocking findings, integrate current main and land on SHIP using `sdlc merge --yes`. Record measured evidence and report the real merge hash/PR to the hub.
