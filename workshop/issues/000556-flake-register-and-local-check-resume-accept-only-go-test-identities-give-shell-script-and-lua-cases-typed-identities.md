---
id: 000556
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '5db31aa363148bac5e6c2f46ee20f63ba30b534a' # card fields mirrored from issue-cards; edit via sdlc
---

# flake register and local-check resume accept only Go test identities; give shell-script and Lua cases typed identities

## Problem

## Spec

`sdlc flake register` accepts only Go `PKG.TestName` identities, so failures in non-Go suites can't be registered or
resumed:
- pair#773 (10-22): `tests/nvim-route-poke-admission-test.sh` timed out at 30s; the suite classified it
  "unattributable-failure" (no Go test identity), so `local-check resume` couldn't apply and the slot had to rerun the
  whole suite. The script is tracked as pair#685.
- parley.nvim#410 (10-22): a #408 native Lua case (o-retry, initial `modifiable=false` assertion) failed once and passed
  on an isolated rerun; it can't be registered at all.

Fix: a typed test identity for non-Go suites: shell scripts by path (`script:tests/x.sh`) and Lua or busted cases by
spec file plus case name (`lua:<spec>#<case>`), produced by the existing suite runners from their own output, so a
script timeout or a named Lua case failure gets an identity the flake registry and `local-check resume` can use.
Same evidence rules as Go flakes: a saved same-head fail-then-pass.

## Done when

- A script timeout in a recorded suite is reported with a `script:` identity (not unattributable), and can be
  registered and resumed (test).
- A failing busted case reports a `lua:` identity that registration accepts (test).

## Plan

- [ ]

## Log

### 2026-10-22
