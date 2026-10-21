---
id: 000532
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '01f3abaa56a1d7b77880da6d8edfdce68cd0d95f' # card fields mirrored from issue-cards; edit via sdlc
---

# Archive preflight misses noncanonical side bundles

## Problem

pair#768 landed in PR521 (https://github.com/xianxu/pair/pull/521; reported merge 2026-10-21 14:37:41 PDT), but post-merge SDLC cleanup refuses with `active or conflicting archive artifact workshop/plans/000768-darwin-pty-final-output-plan.md.side/shared-owner-draft.md`. The card is done and the issue and parent plan are archived, while this tracked side document remains active. The documented retry `sdlc merge --branch 000768-darwin-pty-final-output --yes` repeats the refusal. Reported `sdlc issue recovery reconcile --issue 768` finds no unfinished operations and deleted the remote branch without archiving this leftover. Successful landing therefore leaves an incomplete artifact bundle and no successful documented cleanup retry.

## Spec

Treat a document and its owned side-document bundle as one archive transition, preserving exact bytes and provenance. Enumerate the complete family before publishing archive state, and safely reconcile the already-partially-archived state through a documented idempotent cleanup path. Preserve refusal for conflicting or changed generations; do not silently discard orphaned content or merely weaken archive proof.

The producer used a noncanonical `.md.side` directory; this report does not assert that canonical `.side` archival is broken. Read-only source inspection identifies a naming mismatch worth resolving explicitly: `pkg/sidedoc/tree.go:SideDir` derives `<stem>.side` by stripping `.md`, while the reported existing bundle is `<stem>.md.side`. `cmd/sdlc/landingarchive.go:planLandingSideMoves` collects only SideDir-derived members; plan discovery uses immediate children of the plans root. Later archive-proof family enumeration rejects the leftover at `confirmLandingArchiveCommit` (around line 801). Investigate recognition/validation of this noncanonical naming form before irreversible merge or any archive write: either safely include provably owned existing bundles or refuse before moving their parent, with an actionable recovery path. This is an archive-family completeness/recovery bug, distinct from active #488/#519 (duplicate-ID collectors counting .side directories).

ARCH-DRY: share artifact-family membership between planning, completeness proof and recovery. ARCH-STATE: explicitly cover retry after the parent archive has already committed, without treating absence of a tracker-operation receipt as proof of complete cleanup.

## Done when

- A related noncanonical `.md.side` bundle is detected before irreversible merge/archive effects and refused with an actionable correction, or the workflow provides deterministic safe recovery without data loss. Canonical `.side` archival behavior is preserved. Parent-only archival must not leave this case without a safe recovery path.
- A regression fixture includes the exact nested path `workshop/plans/000768-darwin-pty-final-output-plan.md.side/shared-owner-draft.md`, alongside the canonical `<stem>.side` form, and proves the selected compatibility/recovery behavior rather than normalizing the fixture away.
- A fixture starts with the issue and parent plan already archived, the card done, the side file still active, and no unfinished tracker operation; the documented cleanup retry reconciles it safely and a second retry is a no-op.
- Side contents remain byte-identical and archive provenance remains verifiable. Conflicting archived content or a changed active generation cannot be overwritten or deleted; interrupted/uncertain publication and retry preserve all content.
- Regression coverage exercises the production archive-planning/proof and cleanup/recovery seams; document the supported recovery command and naming policy.

## Plan

- [ ]

## Log

### 2026-10-21 — filed from pair#768 cleanup failure

Git verification in pair slot2: HEAD/reviewed source `8b7a72b86f618ab0b06f15e966570ad44a4e1d0d`; merge `68fac7f0bf4312610f1348601fe4f05765d865f0` has that source as its second parent (Git commit timestamp 14:37:40 -0700). Observed origin/main `1ed39dfdbbe0182b72c4b3a62b5504b2e725ce5d` archives the parent issue and plan. Both HEAD and origin/main contain the active exact side path above as regular mode 100644 blob `670ed082db7429077580eb805edcbcc3907a8ecc`; neither contains its history-side counterpart. Archived issue blob: `5451b5f38eeec15da8be9969cf4b2d4b2413f1d4`; archived parent-plan blob: `d321e0d969632c5818acb92916dab2c1d77efcdc`. No canonical plan.side counterpart was present. No tests, repair, checkout mutation, or cleanup replay performed during filing; retry/reconcile results above are supplied incident evidence.
