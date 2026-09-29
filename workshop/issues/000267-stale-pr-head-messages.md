---
id: 000267
status: working
deps: []
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '930e109ebfee92436c8ca907989ae8d575532fdf' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-28T20:40:36-07:00
flow: {kind: quick, provenance: inferred, spec: "e980c95c", done: "754be3ce"}
---

# merge/pr: a PR behind the local branch reads as 'found 0 PRs'; pr re-run pushes then fails

## Problem

Seen landing #253 (PR #141). After a commit on the issue branch (a main
merge, a FIX-THEN-SHIP fix), the PR's head lags the local branch, and neither
verb says the fix is to push:

- `sdlc merge` reports `need one exact matching PR for <branch>; found 0
  (preserving local work)`. `selectLandingPR` (`cmd/sdlc/landing.go:186`)
  keeps only PRs whose head equals the local HEAD, so an open PR with a stale
  head counts as no PR at all. The later, accurate check at `landing.go:381`
  ("local, remote and PR heads must match; push selected branch before
  landing") is never reached.
- `sdlc pr` re-run pushes the new head (`landing.go:519`), then fails on
  `gh pr create`: "a pull request for branch … already exists". The push it
  needed did happen, but the verb reports an error.
- The branch was left with no upstream after `sdlc pr`, so a bare `git push`
  failed too (`fatal: no upstream configured`), despite `push -u`.

## Spec

- `selectLandingPR` matches by identity (repo, head ref, base, not closed)
  first; a single open PR whose head differs from local HEAD yields the
  head-mismatch refusal that names the fix: push the branch (or re-run
  `sdlc pr`), then retry.
- `sdlc pr` with an existing open PR for the branch pushes and reports
  "updated PR #N to <sha>" with exit 0, instead of failing on create.
- Find out why the durable `sdlc pr` path leaves no upstream configured, and
  fix it or say why not.

## Done when

- A test: open PR with a stale head → `sdlc merge` refuses with the push
  instruction, not "found 0".
- A test: `sdlc pr` with an existing PR pushes and exits 0, naming the PR.
- After `sdlc pr`, `git push` in the issue branch works without arguments (or
  the Log records why not).

## Plan

- [x] Split `selectLandingPR` into an identity pass (`liveLandingPRs`: repo,
  head ref, base, not closed) and a head filter; no head match plus exactly one
  OPEN PR at another head → "PR #N is at X, local is at Y; push with `sdlc pr`,
  then retry".
- [x] `runDurablePR` queries the branch's PRs (identity-checked) before pushing;
  one OPEN PR → push, report "updated PR #N to <sha>", exit 0, no create; >1 →
  refuse.
- [x] After the push, verify `branch.<b>.merge`; if git could not record it,
  warn with the cause instead of dropping git's stderr.
- [x] Tests for all three through the production entry points (`runMerge`,
  `runPR`) on the real-git landing fixture.

## Log

### 2026-09-28

- Filed from #253's landing, at the operator's request.
- Upstream root cause: the agent sandbox denies writes to the primary
  checkout's `.git/config` (`git config --local x y` → "could not lock config
  file … Operation not permitted"). `git push -u` then pushes, prints "error:
  unable to write upstream branch configuration", and still exits 0
  (reproduced with a held `config.lock`). `runGitCmd` drops stderr on
  success, so `sdlc pr` reported nothing. sdlc cannot write config the sandbox
  forbids; the fix is to verify and warn. The pushed path forward is re-running
  `sdlc pr`, which now updates the open PR.
- Taken over in ariadne:2 at 22:02. The previous session stopped at 20:46,
  with the implementation, three tests, helptext and atlas written but
  uncommitted. Kept as is apart from gofmt on `landing.go` (a stray blank
  line).
- Verified the tests are real regressions: with `HEAD`'s `landing.go`, all
  three fail (`TestLandingStalePRHeadNamesPush`,
  `TestLandingPRUpdatesOpenPR`, `TestLandingPRWarnsUnrecordedUpstream`); with
  the change, all pass.
- Scope note: this fixes the local `sdlc pr`/`merge` experience only. The
  GitHub CI failures on every PR are #266 (the ID lint in
  `internal/tracker/records.go`), a separate code path.
