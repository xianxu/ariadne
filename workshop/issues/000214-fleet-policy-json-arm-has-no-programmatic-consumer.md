---
id: 000214
status: open
deps: [pair#170]
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
---

# fleet policy JSON arm has no programmatic consumer

## Problem

`sdlc fleet policy --path P --json` (`cmd/sdlc/internal/fleet/`) had exactly one
programmatic consumer: pair's couch, which shelled out to it per start
resolution and persisted the normalized result — repo identity, admission key,
capacity/action, provider version, declaration digest — on every occupied
incarnation.

**pair#170 M4 deleted that consumer.** couch was rescoped to "couch-lite": a
switcher over one operator's own sessions on one host. Fleet capacity and
incumbency defend the *multi-owner* case, which couch-lite does not have, so
admission and its provider dependency went together. couch no longer invokes
`sdlc` at all — `cmd/couch/main_test.go` now asserts the *absence* of any `sdlc`
invocation.

This is a note, not a request to delete anything. The CLI arm stands on its own
operator-facing merits, and pair has no standing to decide that. But an
integration surface whose last programmatic caller quietly disappeared is
exactly how a surface rots: the e2e tests keep passing, the helptext keeps
promising, and nobody notices the contract has no counterparty.

One consequence is already concrete: pair deleted its
`make test-couch-policy-live` conformance target, which is what checked
ariadne's real provider against pair's strict consumer, plus the weekly workflow
that ran it. **Whatever cross-repo conformance that target provided is gone.**

## Spec

Decide, and record the decision:

1. **Keep as operator CLI.** Fine — but the JSON arm's contract now has no
   automated consumer, so its stability guarantees are whatever ariadne says
   they are, not what an integration test enforces.
2. **Keep and find the next consumer.** If something else should consume it,
   name it; the shape was designed for a caller that no longer exists.
3. **Retire the `--json` arm**, keeping the human-readable one.

The one thing not to do is leave it undecided, which is the state this note
exists to prevent.

## Done when

- One of the three dispositions above is chosen and recorded in this issue.
- If kept: a sentence in ariadne's atlas stating the arm has no programmatic
  consumer, so the next reader does not assume pair still calls it.

## Plan

- [ ] Confirm the consumer census independently — `rg 'fleet policy' ~/workspace`
      across peer repos, since this note is written from pair's side and pair
      only knows about itself.
- [ ] Pick a disposition and record it.
- [ ] If kept, add the atlas sentence.

## Log

### 2026-09-02

Filed from pair#170 M4's close, which is the milestone that removed the
consumer. Raised by that milestone's boundary review: the plan (Task 15 Step 0)
required this note and the first close attempt skipped it — "leaving the
cross-repo consequence unstated is how a surface rots" was the plan's own
wording, and it was nearly proved right by omission.

Evidence at time of filing: `cmd/sdlc/internal/fleet/` present in ariadne;
`grep -rn "sdlc fleet policy" ~/workspace/pair/cmd` returns nothing.
