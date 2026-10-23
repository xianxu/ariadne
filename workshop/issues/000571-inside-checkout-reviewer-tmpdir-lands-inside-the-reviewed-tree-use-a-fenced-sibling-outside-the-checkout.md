---
id: 000571
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '686df5216293458e36b057f97a182e7b35675a2b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T02:04:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "96f72894", done: "43ce81e8"}
---

# Inside-checkout reviewer TMPDIR lands inside the reviewed tree; use a fenced sibling outside the checkout

## Problem

## Spec

From parley.nvim#422 (root cause by parley.nvim:1, 10-23): an inside-checkout `sdlc` close review sets `TMPDIR` to
`<clone>/.sdlc-review/tmp` (reviewbox/confine.go, by design, to stay inside the write fence). Repos whose test
harness derives HOME/XDG/scratch from `TMPDIR` then put all test scratch **inside the reviewed tree**; parley's
zero-config chat walks up, finds the clone's tracked `.parley` marker, and writes into `<clone>/workshop/parley`,
breaking parley's placement invariant (about 10 scattered failures plus a stray untracked artifact in review clones).

Fix: give the reviewer a `TMPDIR` that is **inside the write fence but outside the reviewed checkout** (a sibling of the
review checkout under the review box), so no repo's scratch can land in the tree under review. Keep every other
confinement property.

## Done when

- In an inside-checkout review, `TMPDIR` resolves outside the reviewed checkout and inside the write fence (test).
- A parley test run inside a review clone leaves the reviewed tree clean (integration test or documented manual
  probe with parley.nvim#422's guard).

## Plan

- [x] Red first: replace the inside-checkout expectation with a canonical outside-checkout TMPDIR assertion; reproduce project-marker discovery from scratch through the real confinement/dispatch seam. Retain parley.nvim#422 raw failure bytes and cover both Claude and Codex environments (plus other adapters sharing this seam).
- [x] Move the inside confinement base to the deterministic sibling `<checkout>.sdlc-review`, derived by one pure helper. Keep the existing temp/bin/env layout, Git ceilings/shims, dependency environment and all protected paths. Remove the now-unneeded checkout ignore insertion. Outside-checkout dispatch remains unchanged.
- [x] Bind sibling lifetime to the existing `<checkout>.lock`: checkout reset, close and dead-owner reclaim remove both trees; Sweep recognizes the sibling and consults that same owner lock rather than treating it as unowned debris. A live concurrent sweep preserves it. Retry removes prior scratch and recreates shims/env at the same path. Refuse a sibling symlink or non-directory before writing through it.
- [x] Verify real filesystem ownership/reset/close/crash-sweep and confinement failures; run existing Git-shim, dependency-env and preparation/retry controls. Run a real fence probe proving sibling scratch writable while source/common Git remain protected. In a short disposable Parley clone, run both supported mappings with the actual derived TMPDIR: `make test-spec SPEC=infra/config` and `make test-spec SPEC=infra/test_harness`; require scratch outside the clone, no fallback, and clean Git status. Save raw command/output bound to source revisions.
- [x] Update atlas/help comments and prevention lesson, publish the pushed head, run its declared affected check (suite if required), close through the binary-owned review and merge on SHIP. Confirm the owner-run disposable probe satisfies live conformance with the hub; do not stage on :0 without that approval.

Design: reuse the existing checkout lock as the single lifetime owner (ARCH-DRY/ORDER/FUNERAL). A fresh random sibling with no shared lock would be swept while live; putting checkout and scratch in a new enclosing container would change every clone path. The deterministic companion is the narrower correction. No new journal, lock type, metadata format, environment flag or fence allowance is introduced. The current OS fence denies protected source/common-Git paths and already permits this sibling; it is not an allowlist limited to the checkout.

Actions: run review; retry the same review after a reported failure.

| Internal state / failure | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Owner lock held; checkout and sibling ready | review runs | run review | scratch is disposable |
| Sibling creation/type/write probe fails | review did not run; named path error | retry after repairing named path | no accepted verdict |
| Invalid response triggers reset | existing retry, fresh checkout and scratch | run review | prior attempt scratch discarded, existing captured output retained |
| Review completes/cancels | existing result/error; both disposable trees removed | retry if needed | scratch discarded |
| Owner dies or cleanup cannot remove trees | existing stale/debris cleanup on next review; errors remain visible | retry review invokes existing Sweep | scratch removed only after owner lock can be acquired |

Risky-function strategies:

| Function / seam | Adversarial strategy | Independent guard |
| --- | --- | --- |
| ConfineWithDependencies / Prepare / Env | marked checkout, inherited hostile TMPDIR, sibling symlink/file and unwritable path | canonical path containment and real scratch marker walk; no source mutation; existing shell/env controls |
| Checkout.ResetContext / Close | scratch contains stale/read-only content; retry after mutation | old sentinel absent, both trees removed/recreated, checkout pin unchanged |
| Sweep / reclaim | sweep during held lock, then released/dead owner, orphan companion and failed removal | actual flock protects live sibling; source and unrelated live checkout survive |
| judge.Dispatch / preparation | native shell processes for Claude/Codex, invalid-response retry and cancellation | recorded effective TMPDIR outside checkout on each attempt, fresh scratch, clean reviewed tree |

ARCH-PURE: only companion-path derivation is pure; existing filesystem/process seams own IO. ARCH-PURPOSE: cover scratch placement plus the real Parley discovery regression, not just a changed env string. ARCH-MOCK: reuse portable real-Git checkout fixtures and existing stateful script reviewer adapters; actual Parley/fence probe checks conformance. ARCH-CONSTRAINTS: at most two attempts per dispatch; one companion per checkout, fixed-count setup operations; existing Sweep stays one pass over its entries with constant extra owner checks per companion. Cleanup remains proportional to disposable scratch files; no per-byte/keystroke path or extra repository scan. ARCH-SECURE: canonical sibling must not alias source; protected paths and shell shims remain authoritative; filesystem failure aborts before reviewer launch. ARCH-EVOLVE: no persisted schema changes. Mechanical enforcement is the behavioral containment/marker-walk and lock-lifetime regression family; a lexical ban on temporary path joins would not establish this semantic ownership boundary.

## Log

### 2026-10-23
- 2026-10-23: closed — Red-first scratch ancestry/reset/close/live-sweep/collision regressions reproduced, then focused reviewbox/judge packages PASS8.866s/7.653s and race controls PASS3.715s/6.378s; lint PASS0.897s. Real production-dispatch Seatbelt Parley infra/config and infra/test_harness mappings PASS for Claude/Codex260.27s, effective outer/inner temp paths verified, source writes denied, clean clones/process census. Raw pinned host proof is workshop/plans/000571-native-conformance.md. Final recorded suite lease-970537564 on e44b9bbb0aa6caaa598168aedbd15fe4ba123916 remains queued; landing held for its PASS, review overlap authorized by tl-worker. Hub accepts documented old-sweeper rollout limit and owns shared-provider rebuild at landing. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

### 2026-10-23 — claim, rebuild and bounded plan

Authenticated dispatch `03afd926-9b55-4821-975d-c32cb4751446` resumes this slot for #571, superseding the previous shutdown idle. Verified Go1.27.3; refreshed clean main-slot4 by fast-forward, ran make build and explicitly rebuilt bin/sdlc because its existing sentinel skips that binary. Build metadata reports Go1.27.3, revision e9e960c70a0513ad3235edd6ba7e1f56de188545, modified=false. Claimed #571 and ran start-plan; design is on its own issue branch. No implementation edits or test runs yet.

Read parley.nvim#422 and its harness guards. Peer receipt `d278dc8b-abbb-402a-878a-b438be66906a` supplies raw capture `/Volumes/TestScratch/tmp/r422-review.log` and a short disposable clone `/Volumes/TestScratch/tmp/r422`, source HEAD091392db (peer-reported; independently verify full hash before using evidence). Bounded reproducer sets TMPDIR to CLONE/.sdlc-review/tmp and runs infra/config; expected red is a chat created in CLONE/workshop/parley. infra/test_harness adds placement guards. Long clone paths can silently make Neovim tempname fall back to /tmp (parley.nvim#423), so inspect effective paths and use a short clone. Once #422 lands, its up-front test-placement refusal replaces downstream failures. Peer checkout remains untouched.

Current implementation inspection: Confinement.Prepare deliberately creates the base inside root and ignores it; Checkout.ResetContext deletes it implicitly; Sweep removes any directory without its own lock. Therefore relocation must move cleanup ownership as well. The proposed sibling shares the existing checkout lock, and the existing deny-list fence needs no widening. Await hub approval of this bounded plan before change-code.

### 2026-10-23 — approved correction implemented; regression red/green

Hub approval `dba59cb7-50af-41d4-bd40-4aba9d7349c9` authorizes the companion/lock correction and disposable conformance instead of :0 staging, plus notifying parley.nvim:1 after landing. Change-code passed on quick flow. The first frozen targeted run failed1.851s with expected failures: in-checkout marker discovery, reset/close residue, live Sweep deleting companion, and accepted symlink/file collisions. Minimal production correction now shares one companion derivation/removal boundary across confinement, checkout reset/close and dead-owner reclaim; Sweep preserves companions under the existing owner lock and groups orphan cleanup once. Removed the obsolete checkout ignore writer.

Reviewbox/judge focused controls PASS2.728s/1.388s; whole reviewbox/judge packages PASS8.866s/7.653s. Final fixture-integrated controls PASS2.614s/1.367s. Native shell dispatch exercises Claude/Codex/Gemini temp env equality, reset freshness, final cleanup and clean Git status; a compile-only test alias typo was corrected before accepting that verification. Cmd/sdlc preparation/dependency/fence controls compile and PASS1.019s; the opt-in Parley probe was skipped in that ordinary run. No test failure was registered as a flake.

Raw Parley failure capture copied byte-for-byte to `cmd/sdlc/internal/reviewbox/testdata/parley422-review.log`, SHA256 `5aa1db953f80ce37eebf0e2482caae090f6069129d15c66053c0b27454796346`. The normal marker-walk regression consumes its original scratch/escaped-chat evidence. Read-only peer source/repro independently resolved to `091392db58ad91f7f937e2e977d496894beab962`; peer provided supported infra/config and infra/test_harness mappings. Added opt-in native conformance through production judge.Dispatch and the real platform fence, checking protected-source writes refused and Neovim tempname remains under effective TMPDIR. Next: run that probe from a short disposable root against the retained Parley repro, record raw results and run focused race/lint controls before publishing the final recorded suite. No :0 or peer checkout mutation.

### 2026-10-23 — native conformance PASS and accepted rollout limit

Final real Seatbelt/production-dispatch Parley probe PASS260.27s (Claude107.29s; Codex152.75s), command exit0. Both infra/config and infra/test_harness mappings pass in each agent case, reviewed Git trees stay clean, protected-source/common-Git writes are refused, and process censuses report no survivors. Effective Neovim tempname stays under both outer TMPDIR and the deeper make-harness p/tmp root; no long-path fallback supplied a false pass. The source pinned at probe start was Parley `92bf8cbca1be44942bbfc078011cee7b89084812`, now carrying #422's placement guard; it differs from the earlier peer-reported repro091392db. Captured output prints the exact pin for both cases. Raw output SHA256 `bb16d91b3f5376c9c23f7352bc47035f9aea04700441e324d0f17519dae4d0ff`; [Native conformance evidence](../plans/000571-native-conformance.md) includes exact command, source/blob metadata and unchanged raw output. First probe PASS193.71s preceded the extra inner-temp check and is superseded by this final probe.

Focused race controls PASS: reviewbox3.715s, judge6.378s. Go lint catalog PASS0.897s. No affected declaration exists in this repo, so the final required recorded check is suite. No background targeted/live runner remains; sessions30947,47406,26068 completed exit0. Next: commit/push evidence, open PR, retain one recorded suite on the final source and close on its passing record; then merge under normal gates. At landing notify hub for its immediate shared-provider rebuild and parley.nvim:1 to release #422. No :0 stage is authorized or needed.

### 2026-10-23 — text-only evidence publication

The asset attach verb refused this repository's text-only policy before copying anything. Preserved the host evidence as an ordinary linked Markdown document under workshop/plans instead; no asset-policy change or attachment bypass. The committed text carries metadata and the unchanged raw combined output for independent review.

### 2026-10-23 — final head published; recorded suite retained

PR409 is open on source `e44b9bbb0aa6caaa598168aedbd15fe4ba123916`. Exactly one recorded suite is retained: terminal68032, lease-970537564, attemptda762f03dc128f47005122d50a07b374; log `.git/sdlc/test-stats/inflight/da762f03dc128f47005122d50a07b374/run.log`. Initial FIFO position5, held by ariadne:5. Keep this runner; no duplicate or restart. All implementation/checklist preparation is complete; the binary-owned close and final suite/merge gates remain outstanding.

The worker contract permits boundary review to overlap this queued suite, provided a pass on the final source still precedes landing and any code correction gets fresh verification. Run close with the focused/race/native evidence above; do not claim a recorded PASS until this retained runner returns it. On SHIP plus green, merge promptly and notify hub for provider rollout, then Parley owner for #422. This Log-only checkpoint does not alter the suite source.

## Revisions

### 2026-10-23 — mixed-binary rollout limit accepted

Authenticated hub receipt `3e9778b4-595a-4fb4-939e-8e74d1636927` accepts the deployment limit without another mechanism. An old Sweep sharing the review root does not recognize the new companion and can delete a live new review's scratch; that review fails and reruns. At landing the hub immediately rebuilds ariadne:0 tools (the shared sdlc for every slot) and sends a fleet notice to private-build owners to rebuild. This is a rollout step, not a lock alias, versioned root, journal, fence widening or schema change. The atlas records the same limitation.
