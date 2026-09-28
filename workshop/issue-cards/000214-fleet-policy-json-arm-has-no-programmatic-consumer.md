---
id: 000214
status: open
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
github_issue:
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
