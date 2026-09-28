---
id: 000258
status: open
deps: [000255]
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '1695812907049ba8df0754175488b14b0058c95f' # card fields mirrored from issue-cards; edit via sdlc
---

# Delete the legacy writers and legacy mode

## Problem

#252 kept a legacy mode so the fleet could cut over one repository at a time:
the binary runs the pre-#252 workflow wherever a repository has no issue
tracker. #255 cut every work repository over (2026-09-28); only the brains,
which the lifecycle refuses anyway, remain without a tracker. The legacy
writers and legacy mode are now dead weight, kept on purpose for a while as a
fallback in case the cutover missed something (operator, 2026-09-28).

## Spec

Remove, with their tests and docs:
- legacy mode: `cmd/sdlc/legacymode.go` and its dispatch in `issue new`,
  `claim`, `set-status`, `start-plan`, `change-code`, the card setters,
  `move-detail` and the transfer guard;
- `issue publish` and the `issue sync --push` path;
- the Makefile/Python shell fallbacks that allocate IDs or write status;
- the legacy-equivalence harness (`cmd/sdlc/testdata/legacy-equivalence.sh`);
- the "legacy repositories" text in help, README, atlas and `AGENTS.base.md`.

Before deleting, confirm nothing outside the brains still takes a legacy path
(e.g. a fleet dry-run sweep: every work repository reports already migrated).

## Done when

- The legacy writers and legacy mode are deleted, with their tests and docs.
- A repository without a tracker gets a clear refusal (migrate first), not a
  legacy fallback; the brains' charter refusal is unchanged.
- The full suite passes.

## Plan

- [ ] Fleet sweep: every work repository already migrated
- [ ] Delete legacy mode and its dispatch; refusal for tracker-less repositories
- [ ] Delete `issue publish`, `sync --push`, the shell fallbacks, the harness
- [ ] Docs: help, README, atlas, `AGENTS.base.md`

## Log

### 2026-09-28
- Spun off from #255 (its Phase D) on 2026-09-28 so #255 can close on the
  fleet cutover; deferred (`punt`) until the operator wants it.
