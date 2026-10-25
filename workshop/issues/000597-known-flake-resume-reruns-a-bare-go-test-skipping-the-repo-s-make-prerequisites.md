---
id: 000597
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '1ccd5a1a9ee52ef3c9cd7e3f1c2dd16966ad642b' # card fields mirrored from issue-cards; edit via sdlc
---

# Known-flake resume reruns a bare go test, skipping the repo's make prerequisites (pair fails to build)

## Problem

## Spec

**`sdlc local-check resume` of a known-flake rerun runs a bare `go test -run` that skips the repo's make
prerequisites, so in pair it fails to build and the resume refuses (pair:6, pair#854, 10-24).** Resuming attempt
`da7653e4` on `8747012d` ran `go test -run TestFreshStartupRunSerializesLaunchThroughAdoption`; the build failed at
`runtimebundle/embed.go:8: pattern assets/runtime/files: no matching files found`, because the make prep that
generates those embedded assets never ran. The resume then refused with "proof unavailable, fix/advance the code and
run a suite on a new head", so a registered flake couldn't be cleared without a whole new head.

Fix: a known-flake rerun (and the merge suite's equivalent) must run the same preparation the recorded suite ran,
through the repo's declared entry point (its make target or test-recipe hook) rather than a bare `go test`, or the
repo declares a prep step that resume runs first.

## Done when

- A resume or flake rerun in a repo whose tests need generated assets (a fixture like pair's runtimebundle) builds
  and runs the selected test; a test pins that the prep runs before the rerun.
- A rerun whose prep itself fails reports that, distinctly from a test failure.

## Plan

- [ ]

## Log

### 2026-10-24
