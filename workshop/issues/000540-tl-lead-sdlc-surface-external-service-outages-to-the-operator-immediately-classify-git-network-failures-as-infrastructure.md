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
- Recognized transport errors retain original exit/context and publication uncertainty identities; local-check preparation network failures create no test verdict or status. Ordinary auth/ref/hook/local errors and cancellation remain unclassified.
- The explicit check reports both DNS and HTTPS paths within its total budget, validates certificates, omits credentials, and joins native probe groups on real CLI SIGINT/SIGTERM. Missing tools/partial observations stay visible.
- TL wording names the exact known operator remedy and treats a post-remedy recurrence as recurring, investigating its cause before repeating the remedy.
- The built binary passes the approved read-only github.com conformance, and the stage is cleared with :0 verified at origin/main.

## Plan

Durable plan: [network outage implementation](../plans/000540-network-outage-plan.md).

- [x] Add shared pure Git transport classification and production-runner regressions.
- [x] Add bounded system/direct DNS and HTTPS net check with fake-backed and native conformance controls.
- [ ] Add approved TL/worker incident rule, docs, live check, recorded suite, close review and landing.

## Revisions

### 2026-10-21 — keep Git failure paths free of probes

TL approval delivery `047b5c29-ca18-4ce0-9dba-841fbfc49d8a` revises the Spec sentence requesting API-host resolution inside each Git failure. Git errors classify immediately as `infrastructure: network`, name only a validated safe host, and print the exact `sdlc net check` command. All system/direct/API DNS and HTTPS probes belong exclusively to that bounded explicit command. This avoids extra resolver IO inside lock-held Git error paths. The incident rule also requires the exact operator command when the known remedy needs their privileges; otherwise distinguish awareness-only from an action request.

### 2026-10-21 — recurring incidents

TL wording approval `ae402fd5-8770-492c-9203-6881ccba84b3` adds: if the same outage recurs after the operator's remedy, report it as recurring and look for the cause behind it before repeating the remedy. This supplements the incident/recovery rule without an automatic remedy or retry mechanism. Added a red/green rendered-help control.

## Log

### 2026-10-21

- Claimed and entered planning from fresh main after #526 landed. Read current worker/SDLC contracts and lessons. Read-only delegated seam audit found the main Git runner, internal contextual/legacy runners and weekly transport runner; one shared pure classifier will serve them (ARCH-DRY/PURE). Preserve exit/error identity and uncertain publication outcomes; no automatic network retries.
- Compared automatic DNS-on-error against a pure classifier plus explicit bounded probe. TL approved the latter and reviewed incident wording. The durable plan records bounded native resolver/direct DNS/HTTPS seams, unknown/unavailable states, secret-safe output, and test/live-check scope. Implementation has not started; plan-quality gate follows this checkpoint.
- Plan-quality round 1 identified missing real CLI signal ownership and function-level test strategies. The revised plan names both SIGINT/SIGTERM admission and process-group conformance, and compresses inventories into named adversarial/property tests. Round 2 CLEAN disposed PQ-1/PQ-2 addressed, no new findings. Hub confirmed staged read-only github.com conformance; no forced outage/system change. Implementation admission is waiting only on the derived estimate.
- Full change-code admission subsequently passed. Estimate-quality INFO called the allocations optimistic; module rows include fixtures, fuzz, TLS and signal controls, and the review/discovery rows absorb targeted verification/staging/landing orchestration. Discovery distinguishes native resolver utility behavior from the pinned HTTPS/TLS lifecycle. Recorded suite execution may increase actual ship time; no tests or gates are waived to fit the estimate.
- Implemented one pure typed transport classifier across the four shared runners, preserving original exit/context and publication uncertainty identities. Fake production fetch failures reproduced ordinary exit-128 errors before the fix; isolated real local fetch success remains unchanged. gitx full package passed (99.044s), selected publication/lifetime controls passed, classifier and credential fuzz controls passed. Raw logs and summaries live on SSD under the primary Git common dir's `sdlc/test-stats/evidence/000540/`.
- Added bounded net-check with system/direct DNS and independent pinned-IP verified HTTPS, fixed safe reason classes, explicit partial/unavailable observations, and no mutation/retry. Native TLS is synchronously owned so cancellation cannot strand a transport dial. Private TLS/process controls, stateful probe sequencing and parser fuzz passed; late-cancellation and partial-A/failed-AAAA regressions observed red then green.
- Main CLI/remote/preparation/signal controls passed, then race controls passed (5.300s). The local-check preparation regression asserts no test-stat/local-check record or GH test status on failed fetch. A missing-signal-owner overlay fails both real CLI SIGINT and SIGTERM controls, while production passes; evidence `net-signal-mutation-red.log`. TL/worker packages passed (39.493s/0.271s), golint/shelllint passed (2.117s/0.691s), and vet passed for all touched Go packages. Approved incident wording and discovery docs are implemented. Next: commit and stage the approved live check, then recorded suite and close.
- Live conformance on built/staged source `8c3160b648b295106373d96ac4c1823d7b25f5c2` (all binaries vcs.modified=false) passed read-only against github.com: system/direct DNS both `172.182.252.133`; both independently resolved no-proxy HTTPS paths HTTP 200; api.github.com system/direct DNS both `140.82.116.6`. No forced outage, resolver/cache edit or system change. Hub approved the final wording with the recurring-incident sentence. Stage cleanup and :0 tip verification follow, then the recorded full suite.
- Stage cleared successfully and rebuilt main binaries. The restored main was two issue-only publications behind; using the documented host-only resting refresh procedure, verified clean main/upstream, fetched/pinned main, proved ancestry, fast-forwarded without autostash/overwrite, and verified :0 HEAD equals origin/main at `af0279c76507ac63de2c97bb45ae2dee549e3b47`. Hub received the live result and tip confirmation. PR #383 opened; recorded suite is next. Final recurring-outage help regression passed after observed red; the live-tested net-check code is unchanged by that prose addition.
- Close review round 1 REWORK found BR-1: native git:// uses resolver/connect diagnostics outside the HTTP/SSH recognizers, despite its host parser being supported. The recorded suite was still queued, so cancelled only our exact lease `lease-1862467006` via local-check cancel before changing code. No suite ran or passed on that source. Completing the transport family with a real local Git refusal and hermetic DNS/timeout controls; fresh corrected-head suite and close review follow. Command-tree help/recovery controls separately passed (11.021s).
- BR-1 now covers native Git resolver diagnostics and paired connect-header/errno entries, with strict matching host, contiguous address indexes, validated IPv4/IPv6 and supported refusal/timeout reasons. A real bounded production Git loopback refusal observed red then green; internal native DNS/timeout runner controls did likewise. Negative hook/local/auth/ref and malformed-pair mutations remain ordinary errors. Native-seeded classifier fuzz passed 53,028 executions; final main production controls passed (0.604s), network controls passed (1.162s). Post-fix main and gitx race controls passed (3.904s / 2.673s); vet passed. Raw BR1 logs and summary are persisted in the same SSD evidence directory. Probe code and approved wording/live conformance are unchanged. Fresh suite/review will run on the committed correction.
- Close review round 2 disposed BR-1 addressed, then REWORK BR-2: bounded main Git classified a diagnostic before adding context.Canceled, retaining both types. Cancelled only queued lease `lease-2592813877` before tests started. Sweep contextual ownership structurally: shared raw subprocess output until final context/cleanup checks, then classification, retaining existing cancelled-success controls and process/publication identities. The plan records this correction; production barrier regressions and fresh source suite/review follow. No live probe/wording change.

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
