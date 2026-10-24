---
id: 000615
status: open
deps: [pair#862]
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '4033a67dfbb6017a4ebec077014fb3bc8359ff71' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead session-sync fake preserves unread ranges after apply

## Problem

ariadne#588 made fakeSessionSync.Apply in cmd/tl-lead/sessionsync_fake_test.go call seeAll, and TestFakeApplyAdvancesBaseline explicitly models Pair's old whole-file baseline consumption. pair#862 is correcting that production behavior. The fake must follow the landed contract so tests do not hide unread operator edits.

## Spec

After pair#862 lands, update the tl-lead session-sync fake to advance its baseline only for safely mapped actual apply-changed ranges. Preserve operator unread edits elsewhere, including inside the same touched section. Do not independently resolve occurrence numbers on the shadow: operator insertions/removals can change them. Pair's approved conservative outcome: any unmappable changed range or missing baseline keeps the entire old/empty shadow while the document apply succeeds with a visible baseline-retained warning and explicit recovery: inspect diff, review the remaining edits, then read --all to acknowledge them. Section reads do not advance the baseline. No new persistent state or retry.

Use the landed pair#862 plan, implementation and portable fixture cases as the oracle. The fake remains independent; no Pair runtime import or branch stacking. Converting instructions link to one apply is separate work, not required here.

## Done when

- FakeApply's obsolete seeAll behavior and TestFakeApplyAdvancesBaseline are replaced with range-scoped baseline assertions.
- Frontmatter-plus-answer-link apply preserves unread Needs you changes elsewhere and within a touched section.
- Changed duplicate counts, broad/no-op records and missing/unmappable baseline cases match landed Pair behavior.
- Relevant tl-lead tests pass; all full verification is through the SDLC queue under the overload restriction.

## Plan

- [ ] Wait for pair#862 to land; pin its actual contract and fixture evidence.
- [ ] Update fake and regressions, retaining portable stateful behavior and existing guards.
- [ ] Qualify and land independently through SDLC.

## Log

Filed unclaimed by pair:3 on authenticated ops:0 instruction 5ef6e854-e86a-48d7-b966-adec8c332ce2. The hub assigns this half after pair#862 lands; Pair's acceptance no longer waits for this dependent follow-up.

### 2026-10-24 — final recovery contract from Pair

Pair#862's executable recovery tests showed Read(section) does not save the
shadow. Authenticated ops:0 approval4f69397b-6649-40b4-9135-6541b1b756c4
corrects recovery to inspect diff, review remaining edits, then explicit
read --all, with no change to Read(section). Pair close BR-1 found this
receiving ticket's obsolete wording; corrected here before the handoff closes.

## Revisions

### 2026-10-24 — remove obsolete scoped-read recovery

The original Spec prescribed “scoped-read recovery”. That operation does not
acknowledge the baseline in current Pair. Per the approved Pair contract,
recovery now explicitly reviews the remaining diff before read --all;
section reads remain non-acknowledging. No fake implementation is included
in this handoff correction; the issue stays unclaimed and depends on pair#862.
