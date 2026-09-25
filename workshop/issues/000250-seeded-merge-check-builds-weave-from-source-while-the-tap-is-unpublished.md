---
id: 000250
status: working
deps: []
github_issue:
created: 2026-09-24
updated: 2026-09-24
estimate_hours:
started: 2026-09-24T20:15:40-07:00
flow: {kind: quick, provenance: inferred, spec: "d0c76448", done: "99f79d25"}
---

# Seeded merge-check builds weave from source while the tap is unpublished

## Problem

The seeded `.github/workflows/merge-check.yml` (manifest `seed`) hands consumers
to `bootstrap.sh`, which runs `brew install xianxu/ariadne/weave`. The tap repo
`xianxu/homebrew-ariadne` is not published yet (#241), so every consumer that has
committed the post-#239 seed fails `merge-check` before any check runs:
`fatal: could not read Username for 'https://github.com'` (Homebrew cloning a
repo that does not exist). pair is the first: pair#323, every PR since
2026-09-24T02:43Z. The other consumers (parley.nvim, tools, …) still carry the
pre-#239 workflow and will break the same way on their next re-seed. ariadne's
own CI is unaffected because it builds a candidate `weave` from source.

## Spec

In the seeded workflow's "Build candidate gateway" step, add a consumer branch
after the ariadne-source branch:

- If `brew tap xianxu/ariadne` fails, emit
  `::warning::` naming #241, shallow-clone `https://github.com/xianxu/ariadne`
  into `$RUNNER_TEMP`, `brew bundle` its Brewfile, `go build ./cmd/weave` into
  the same `$RUNNER_TEMP/weave-candidate` dir, and append it to `$GITHUB_PATH`,
  the same candidate path the ariadne-source branch uses.
- `bootstrap.sh` is unchanged. It already prefers a `weave` on PATH that answers
  `weave dependencies --help`, so the published-formula path is skipped only
  while the fallback supplied weave.
- Once #241 publishes the tap, `brew tap` succeeds and the branch goes dormant;
  #241 owns deleting it (noted there).
- The build tracks ariadne `main`, as the published formula would.

## Done when

- `scripts/test/portable-ci.test.sh` executes the workflow's real run blocks for
  three gateways: tap published (tap → install → compile), tap unpublished
  (tap fails → clone → bundle → candidate build → compile, no `brew install`),
  and ariadne source (unchanged); plus the existing failure cases.
- The workflow text carries the `::warning::` naming #241.
- After re-seeding, a pair PR's `merge-check` passes (pair#323).

## Plan

- [ ] Test first: portable-ci fixture gains a tap-unpublished gateway (fake `brew tap` failure, pass-through `git` that stubs `clone`); published case now expects the `brew:tap` row
- [ ] Workflow: consumer fallback branch in "Build candidate gateway"
- [ ] Mutation-check; `bash scripts/test/portable-ci.test.sh`; note removal in #241

## Log

### 2026-09-24

- Filed from pair#323. Seed semantics (`construct/base.manifest`: seed tracks
  upstream, refreshed on drift) rule out a pair-local fix.
