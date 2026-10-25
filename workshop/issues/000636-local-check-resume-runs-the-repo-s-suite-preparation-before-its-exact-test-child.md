---
id: 000636
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '3df79f7a33078134292f0a084143a4b8007f2676' # card fields mirrored from issue-cards; edit via sdlc
---

# local-check resume runs the repo's suite preparation before its exact-test child (pair's generated assets)

## Problem

## Spec

**A local-check resume builds its exact-test child without the repository's suite preparation (pair:4 on pair#887,
10-25).** The resume of failed attempt 4e5a9fb7 (a registered flake, pair#897) waited 42 minutes in the queue, then its
child failed to build before any test ran: `cmd/internal/runtimebundle/embed.go:8:41: pattern assets/runtime/files: no
matching files found`. pair's suite generates its embedded runtime assets first (the build target the full suite runs);
the resume child ran `go test -run …` directly in a checkout that never generated them. sdlc then refused: "resume proof
… is unavailable: child identity, completion or log digest is invalid; fix/advance the code and run a suite on a new
head", which forces a whole new suite for a one-test rerun.

The resume child runs the same preparation step the repository's suite runs (the declared suite setup, e.g. pair's
asset generation) before its exact-test command, through the same suite-command declaration, not a pair special case.
A build failure in the child is reported as "resume preparation failed" (an infrastructure outcome, retryable), not as
an invalid proof.

## Done when

- A resume in a repository whose suite generates files first succeeds on a clean checkout; tests with a fake repo
  whose build needs a generated file.
- A child that fails to build reports a preparation failure distinct from a proof failure.

## Plan

- [ ]

## Log

### 2026-10-25
