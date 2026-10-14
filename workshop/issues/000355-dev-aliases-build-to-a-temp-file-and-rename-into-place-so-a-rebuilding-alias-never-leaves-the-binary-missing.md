---
id: 000355
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '7bf13e350d06aa6a77b6da89d034630747210fbb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T20:27:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "61090024", done: "1dc27e69"}
---

# dev-aliases: build to a temp file and rename into place, so a rebuilding alias never leaves the binary missing

## Problem

Each dev-alias function (`construct/dev-aliases.sh`) rebuilds its owner binary on every call as `rm -f bin/X && go build -o bin/X ./cmd/X`, then runs it. On 2026-10-13 the operator's .zshrc switched every interactive shell to these functions with a fixed workspace, so any `couch` call anywhere deletes `~/workspace/pair/bin/couch` for the few seconds of the build; other processes calling that path meanwhile (non-interactive Codex shells via PATH/links, the hub by full path) fail with 'no such file or directory'. The TL hit exactly that.

## Spec

Build to a temp file beside the target (`bin/.X.tmp.$$`) and `mv -f` it over `bin/X` (an atomic rename on the same filesystem), so the old binary stays until the new one replaces it. Concurrent alias calls must not corrupt each other (unique temp names; last rename wins). Keep the dev-prepare step (#299) and stdout-to-stderr behaviour; don't add a second staleness scheme (go's build cache already makes rebuilds fast).

## Done when

- The emitted functions build to a temp file and rename; a test runs two concurrent alias builds while a third process repeatedly execs the target path and never sees it missing; dev-aliases.test.sh passes under bash 3.2.

## Plan

- [x] Emitted functions build into a per-call `mktemp -d bin/.X.XXXXXX` directory beside
  the target and `mv -f` the binary over `bin/X` (atomic on one filesystem; unique per
  call, last rename wins), then remove the directory. A directory, not a temp file:
  `go build -o` refuses to overwrite an existing file that is not a Go build output. The
  rename also gives the new binary a new inode, the code-signing safety `rm -f` gave.
  dev-prepare and stdout-to-stderr unchanged; no second staleness scheme.
- [x] Shape assertions updated; concurrency test: two concurrent rebuilds through a slow
  stub go while a poller execs `bin/X` and requires a whole binary every time.

## Log

### 2026-10-13
- 2026-10-13: closed — Round 1 fixes: atlas base-layer.md now states the build-and-rename form (the only stale restatement; reviewer swept for rm -f bin/ and go build -o bin); failed-build test added (stub go writes a partial binary and exits 1: call returns non-zero, runs nothing, leaves no bin/.X.*, previous bin/X still runs; with origin/main emitter the keeps-previous-binary case fails); Spec Revisions entry for the per-call directory. dev-aliases.test.sh under /bin/bash 3.2.57: 48 passed, 0 failed; the concurrency case fails with 62 No such file or directory execs on origin/main emitter. Real go build through the emitted function in bash 3.2 and zsh. Full make test green (12 shards, 1239 cmd/sdlc tests, 49 packages, 0 failed). (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

### 2026-10-13 — ariadne:3

`dev-aliases.test.sh` under macOS /bin/bash 3.2.57: 44 passed, 0 failed. Teeth: with
origin/main's emitter swapped in, the concurrency case fails with 62 `No such file or
directory` execs (the TL's couch failure) and a final binary that is not whole, plus the
four old-shape assertions. The first stub wrote its output immediately, so the old
`rm -f` window would have been microseconds and the poller might never have seen it; the
stub now waits before writing, like a real link, and writes in two steps, and the poller
requires the binary's whole output. A build killed mid-way leaves its `bin/.X.*`
directory (gitignored with bin/); nothing sweeps it, since a sweep could remove a
concurrent build's directory. Nothing runs `construct/scripts/test/*.test.sh` on its
own; the only shared script-test runner is the local-gates `20-ci-workflow-tests` check.

Correction: the close passed `--no-atlas`, claiming no atlas surface, but
`atlas/workflow/base-layer.md` restated the old `rm -f` form. Updated it after the close
(doc-only).

Close review round 1 (FIX-THEN-SHIP; the run itself refused as stale because the atlas
fix above landed mid-review): two Important, one Minor.
- Important: the stale `base-layer.md` form, fixed in fac123dd (the same restatement
  the correction above names).
- Important: no test for a failed build. Added: a stub go that writes a partial binary
  and exits 1; the call returns non-zero, runs nothing, leaves no `bin/.X.*`, and the
  previous `bin/X` still runs (the old `rm -f` form destroyed it).
- Minor: the Spec's temp-file wording; see Revisions.

## Revisions

- 2026-10-13 — Spec said "build to a temp file beside the target (`bin/.X.tmp.$$`)".
  Built instead into a per-call `mktemp -d bin/.X.XXXXXX` directory, then `mv -f` over
  `bin/X`: `go build -o` refuses to overwrite an existing file that is not a Go build
  output, and `$$` is the parent shell's pid inside the function's subshell, so
  concurrent calls from one shell would share it. Same filesystem, so the rename stays
  atomic; the per-call directory gives the uniqueness the Spec asked for.
