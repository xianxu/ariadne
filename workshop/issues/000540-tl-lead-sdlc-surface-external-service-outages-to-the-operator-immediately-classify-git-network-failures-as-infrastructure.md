---
id: 000540
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours: 2.7
collaboration-mode: collaboration
card_mirror: 'fadb435ae51f0dbbb581d8f78da5fe20d5ff8b56' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T20:12:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
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

Durable plan: [network outage implementation](../plans/000540-network-outage-plan.md).

- [ ] Add shared pure Git transport classification and production-runner regressions.
- [ ] Add bounded system/direct DNS and HTTPS net check with fake-backed and native conformance controls.
- [ ] Add approved TL/worker incident rule, docs, live check, recorded suite, close review and landing.

## Revisions

### 2026-10-21 — keep Git failure paths free of probes

TL approval delivery `047b5c29-ca18-4ce0-9dba-841fbfc49d8a` revises the Spec sentence requesting API-host resolution inside each Git failure. Git errors classify immediately as `infrastructure: network`, name only a validated safe host, and print the exact `sdlc net check` command. All system/direct/API DNS and HTTPS probes belong exclusively to that bounded explicit command. This avoids extra resolver IO inside lock-held Git error paths. The incident rule also requires the exact operator command when the known remedy needs their privileges; otherwise distinguish awareness-only from an action request.

## Log

### 2026-10-21

- Claimed and entered planning from fresh main after #526 landed. Read current worker/SDLC contracts and lessons. Read-only delegated seam audit found the main Git runner, internal contextual/legacy runners and weekly transport runner; one shared pure classifier will serve them (ARCH-DRY/PURE). Preserve exit/error identity and uncertain publication outcomes; no automatic network retries.
- Compared automatic DNS-on-error against a pure classifier plus explicit bounded probe. TL approved the latter and reviewed incident wording. The durable plan records bounded native resolver/direct DNS/HTTPS seams, unknown/unavailable states, secret-safe output, and test/live-check scope. Implementation has not started; plan-quality gate follows this checkpoint.
- Plan-quality round 1 identified missing real CLI signal ownership and function-level test strategies. The revised plan names both SIGINT/SIGTERM admission and process-group conformance, and compresses inventories into named adversarial/property tests. Round 2 CLEAN disposed PQ-1/PQ-2 addressed, no new findings. Hub confirmed staged read-only github.com conformance; no forced outage/system change. Implementation admission is waiting only on the derived estimate.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Calibration is flagged stale by estimate-source, so the hours are provisional. Issue shaping uses the low issue-spec design primitive without discount. Classifier, CLI/config wiring and HTTP/native adapter extensions use well-specced smaller-module primitives (raw design 0.2, times 0.2 spec discount; raw implementation 0.5, times 0.4). New orchestration uses greenfield design 1.0, halved because stdlib HTTP/context and existing processgroup supply the mechanisms, then times 0.2 for the approved detailed plan; implementation 0.8 times 0.4. The rule uses skill/dispatcher design 0.3 times 0.2 and implementation 0.4 times 0.4. Docs use 0.1 times 0.2 design and 0.2 times 0.4 implementation. One close review and two native resolver/HTTPS live observations account for review and discovery overhead. Familiarity 1.0: existing SDLC/Cobra/processgroup territory. Design 0.80 plus 15% thorough-plan buffer = 0.92; scaled implementation 1.78; total 2.70 focused ship hours. No cross-repo vendor coordination.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: issue-spec design=0.50 impl=0.08
item: smaller-go-module design=0.04 impl=0.20
item: greenfield-go-module design=0.10 impl=0.32
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: skill-or-dispatcher design=0.06 impl=0.16
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.00 impl=0.20
item: real-api-discovery design=0.00 impl=0.24
total: 2.70
```
