---
id: 000250
status: codecomplete
deps: []
github_issue:
created: 2026-09-24
updated: 2026-09-24
estimate_hours:
started: 2026-09-24T20:15:40-07:00
flow: {kind: quick, provenance: inferred, spec: "d0c76448", done: "99f79d25"}
actual_hours: 0.09
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

- [x] Test first: portable-ci fixture gains a tap-unpublished gateway (fake `brew tap` failure, pass-through `git` that stubs `clone`); published case now expects the `brew:tap` row
- [x] Workflow: consumer fallback branch in "Build candidate gateway"
- [x] Mutation-check; `bash scripts/test/portable-ci.test.sh`; note removal in #241

## Log

### 2026-09-24
- 2026-09-24: closed — scripts/test/portable-ci.test.sh executes the seeded workflow run blocks for published tap, unpublished tap (new: tap fails -> clone -> bundle -> candidate build -> compile, no brew install) and ariadne source; failed before the change; mutations (remove fallback branch; drop GITHUB_PATH export) both caught; go test ./cmd/weave/... and construct/scripts/test/bootstrap-transitive.test.sh green. No atlas: interim CI fallback owned by #241 for removal, no new architectural surface. Live proof pending: pair#323 re-seed + a green pair PR run.; review verdict: SHIP

- Filed from pair#323. Seed semantics (`construct/base.manifest`: seed tracks
  upstream, refreshed on drift) rule out a pair-local fix.
- Implemented. `portable-ci.test.sh` executes the real run blocks for published
  tap, unpublished tap and ariadne source; it failed before the workflow change.
  The fake `go` now logs to the leaf's events, because the fallback builds inside
  the source clone. Mutations caught: remove the fallback branch; drop its
  `$GITHUB_PATH` export. `go test ./cmd/weave/...` and
  `bootstrap-transitive.test.sh` green. Removal is noted in #241.

