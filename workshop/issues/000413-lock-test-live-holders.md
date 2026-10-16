---
id: 000413
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
severity: medium
card_mirror: '3cb738432fcc5785cf69e3563cbec8f7de7f0b1e' # card fields mirrored from issue-cards; edit via sdlc
---

# Concurrent dead-holder lock test marks its live contenders dead and flakes with missing metadata

## Problem

During ariadne#411 recorded suite on 986653326c2242c56ef7a095a3ebab498f98918f, cmd/sdlc/internal/repolock failed TestConcurrentAcquireReclaimsDeadHolderOnce (0.03s): repolock_test.go:415 reported write sdlc repo lock metadata: open .../.git/sdlc.lock/meta.json: no such file or directory. The package is unchanged from current main. The full suite is still collecting other results; run ID 5f54a534ecf74f0c3b0c7af2873769a8.

The fixture supplies ProcessAlive: func(int) bool { return false } for the original dead PID 555 and both newly acquiring holders 111/222. This makes the live holders appear dead as well. That is a likely invalid liveness model; this observation alone does not establish a production lock bug.

## Spec

Model the original dead holder separately from the two live contenders, and assert safe ownership/release. Preserve production lock safeguards; change production only if a faithful liveness model demonstrates a production defect. Add deterministic ordering where needed rather than extending sleeps or budgets.

## Done when

- The concurrent-dead-holder test represents live contenders correctly and repeatedly passes while checking exclusive ownership and successful release.
- The observed ENOENT regression is addressed without weakening lock safety or merely increasing timeouts.

## Plan

- [ ]

## Log

- Found while verifying #411; file independently and leave unassigned.
- Scratch main control at 87e0feeb63e41946936ccb316583629ce8eb6b8a, go test -trimpath ./cmd/sdlc/internal/repolock -run '^TestConcurrentAcquireReclaimsDeadHolderOnce$' -count=100 -timeout 5m, passed in 4.158s. This did not reproduce the intermittent failure. Log /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne411-lock-main-gezq_l3f.log; scratch checkout removed.
