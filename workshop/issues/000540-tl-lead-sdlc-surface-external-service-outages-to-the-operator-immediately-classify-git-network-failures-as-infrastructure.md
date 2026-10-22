---
id: 000540
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '2067949d918a0295fede02eadf10a1ab60dd201d' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead + sdlc: surface external-service outages to the operator immediately; classify git network failures as infrastructure

## Problem

## Spec

Operator (10-21, after a ~45-minute github.com DNS failure on the host blocked every push, fetch, landing and filing):
"those external service outages, you should bring to my attention in the future, at least I'm aware."

Two parts:
1. **tl-lead rule:** an outage of an external service the fleet depends on (git host reachability or DNS, GitHub API,
   model APIs, the Cloudflare tunnel) goes to the operator **the moment the TL learns of it**: a top Needs-you item
   with what's affected and whether it's local or upstream, plus one `couch --notify`; and a second notice when it
   recovers. Slots that hit it report it to the hub as an incident, not a per-item failure.
2. **sdlc classification:** when a git network operation fails with a resolver or connect error (`Could not resolve
   host`, connection refused or timeout), sdlc names it as **infrastructure: network** in its error (the host, and
   whether `api.<host>` resolves), so slots and the hub don't have to diagnose it, and it's never logged as test
   evidence. A cheap probe verb (`sdlc net check`, light) reports the git host's DNS and HTTPS reachability through
   the system resolver and directly.

## Done when

- tl-lead help states the outage rule (immediate Needs-you + notify, recovery notice).
- A simulated DNS failure in a git fetch yields sdlc's "infrastructure: network" classification (test with a fake
  resolver or an unresolvable host).
- `sdlc net check` reports both lookups (test).

## Plan

- [ ]

## Log

### 2026-10-21
