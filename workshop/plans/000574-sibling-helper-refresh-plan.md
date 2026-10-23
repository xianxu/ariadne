---
collaboration-mode: discussion
---

# Slot-sibling helper refresh implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3. Use superpowers-executing-plans in this session; one atomic close boundary.

**Goal:** Before new suites queue, refresh a safe private sibling supplier; name the current supplier revision behind every legacy ticket.

**Architecture:** Preserve #563 checkout ownership. Add a narrow weave refresh mode under its existing setup lease; SDLC invokes it once at suite setup. Queue text observes legacy suppliers read-only. Scheduling and frozen JSON stay unchanged.

**Tech Stack:** Go, directory-explicit Git, existing workspace discovery, refresh Git seams, private native Git/Python fixtures.

## Evidence and alternatives

Authenticated TL dispatch and issue Spec report 21 stale siblings manually fast-forwarded this morning. No original incident TTY is available; preserve this as a labelled report, not a raw capture.

workflowFile/prepareSuiteHelper resolve the supplied Makefile.workflow. runLocalCheck calls preflight before logs/worktrees/tickets; saved resumes have the same seam. catchUpBeforeWork runs only for start-plan on a resting branch. Existing weave refresh --substrate also compiles and accepts fast-forwardable non-main branches, so calling it unchanged exceeds this issue.

Recommend suite setup: it handles long-lived work across rollouts. Claim/start-plan alone leaves long-lived issues stale. Replacing the supplier with the CLI source helper violates #563. Avoid a separate SDLC Git-update implementation by keeping mutation in weave.

Propose weave refresh --suite-helper, mutually exclusive with other refresh flags: refresh just the verified private sibling on main, without compile, build, generated outputs, missing-checkout acquisition or consumer updates. This extra mode is proposed for TL approval.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| Helper-refresh decision | cmd/weave/internal/refresh/helper.go | new |
| Legacy-ticket supplier diagnostic | cmd/sdlc/queue_suite_supplier.go | new |

Decision consumes typed Git observations and chooses skipped/current/fast-forward. One supplier belongs to one verified environment. Diagnostic relates each legacy ticket to a current checkout observation; it never invents a launch revision. These keep eligibility and wording single-sourced.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| Narrow helper refresh | cmd/weave/refresh.go, internal/refresh/helper.go | new | existing setup lease and refresh Git observation/revalidation |
| Suite setup refresh | cmd/sdlc/helper_refresh.go, localcheck_run.go, localcheck_resume.go | new | installed weave before helper preflight |
| Supplier observation | cmd/sdlc/queue_suite_supplier.go, queue_suite_status.go | new | workflowFile and bounded Git reads |

Reuse acquire.GitRunner and refresh's framed observations, ancestry, ignored-collision checks and stateful fake patterns. Integration fixtures use local bare origins and a stateful private executable that updates its supplier and records calls. Native conformance uses actual weave and Python against private roots. No production state is a fixture.

## Refresh boundary and safety

- Discover the numbered environment through workspace.DiscoverEnvironment; require resolved supplier to be its private ariadne sibling and distinct from consumer. Never update :0, fallback primary, ariadne issue checkout or another dependency. No authority from basename alone.
- Acquire existing .weave-setup.lock nonblocking before observations/fetch/update; retain/inherit its descriptor until Git children end. This owner-lifetime resource excludes cooperative setup without process scans. Busy prints one skip line.
- Require exact main, clean including untracked files, no active Git operation, matching identity, and ancestor relation to a captured origin/main OID. Fetch only main; use existing ff-only/no-autostash/no-submodule/ignored-collision protection. Revalidate identity/branch/HEAD/clean state immediately before effect and confirm resulting HEAD.
- External editors do not honor the lease. Preserve Git index locking, revalidation and safe merge behavior; never claim observations exclude arbitrary writers. No stash/reset/rebase, no rollback after uncertain effects.
- One optional call before each new managed suite/rerun setup, including merge/train/baseline and explicit resume launch; one shared seam. Proof-only resume repost and runner-resource checks do not refresh. Start-plan catch-up stays unchanged.
- Initial whole-call budget 10 seconds plus existing bounded child cleanup: conservative interactive setup budget, measure private conformance. Missing/old weave, busy, dirty/off-main, fetch failure or unavailable evidence prints one path/reason/repair line and continues existing preflight/admission. Caller cancellation still cancels the run. No cache, background updater or per-poll fetch.
- A completed ff may remain after timeout/error. Read back before reporting outcome; uncertain stays uncertain with inspect/retry guidance, never automatic rollback.
- Existing Python waiters keep their loaded policy. They drain normally; #570 grants remain gated. Source refresh does not upgrade a live ticket.

## Visible state table

Actions: continue command; reconcile named checkout then retry; inspect uncertain update then retry; let existing waiter drain.

| Internal state/failure | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Safe main behind, ff confirmed | refreshed path and old/new OIDs | continue command | none; history retained |
| Already current | current supplier | continue command | none |
| Dirty/off-main/detached/diverged/operation | one skipped path/reason line | reconcile then retry | no intentional local changes |
| Busy/unavailable/budget before mutation | one refresh-unavailable line | reconcile then retry | freshness unverified |
| Git effect may have happened | uncertain, observed state if readable | inspect then retry | completed ff may remain |
| Legacy ticket, supplier available | current path/revision, launch revision unknown | let waiter drain | no invented historical proof |
| Missing/unverified supplier or expired read budget | unavailable with named reason | reconcile then retry status | no fabricated revision |

## Status and cost

For text status, enrich each incompatible ticket using its existing worktree: strictly decode, verify fleet workspace identity, resolve workflowFile, observe current supplier HEAD/dirty state. Show full current revision per ticket and explicitly say launch revision unknown. A later checkout ff cannot prove the running helper changed. Missing worktree, invalid identity, Git error or deadline exhaustion yields unavailable. No ticket rewriting or inferred ordering marker.

Cache observations only within the request: at most one resolver operation per unique worktree and one revision/dirty observation per canonical supplier. O(T+U), T tickets and U suppliers, O(U) Git reads, zero network. Retain status deadline. Count 8/16 repeated-supplier tickets and 8/16 distinct suppliers; reject repeated probes or status/admission-poll fetch. New-suite setup adds one optional process and at most one fetch independent of queue duration. Grant/revoke gain no supplier probes. Frozen JSON and compatibility logic are unchanged.

## Tasks and function-level adversarial verification

- [x] Red pure eligibility/transition tests: safe main behind, current, dirty/untracked, off-main, detached, ahead/diverged; identity/branch/HEAD/dirty changes between observation and effect; cancellation before/after ff preserves truthful result.
- [x] Implement weave mode via existing lease/Git machinery. Private bare-origin tests assert consumer HEAD/index/generated bytes unchanged, no compile/build, ignored collisions protected, busy lease prevents writes, inherited child lease lifetime and incompatible flags rejected before effects.
- [x] Wire one SDLC setup seam before preflight/ticket. Stateful executable fixture upgrades a private old helper; real acquire then emits marked ticket, and no ticket precedes completed refresh. Exercise run/merge/train/baseline/resume paths; missing/old weave fallback, cancellation, runner exclusion and proof-only repost exclusion.
- [x] Implement status observation/rendering. Controls: multiple supplier revisions, deduplication, dirty supplier, ff after ticket launch, missing worktree, symlink/identity mismatch, malformed records, Git failure and deadline. Launch revision must stay unknown. Keep old-helper/frozen-JSON tests and grant/revoke probe counts.
- [x] Run red/green targeted tests and counted controls; race lease/update tests. Preserve labelled incident summary and #570 old/new helper reply fixtures. This is semantic lifecycle policy: behavioral/native controls enforce it; an arbitrary-Git lexical ban would not prove it.
- [x] Update atlas/workflow/sdlc-binary.md, weave/local-check help and lessons with boundary, skip/recovery and current-vs-launch provenance. No new map file unless necessary.
- [x] Private actual-executable proof with local origin/queue/grants; no staging or live grant. Seek TL approval for private conformance plus post-land :0 rebuild as rollout proof.
- [ ] Push PR; central recorded suite, mandatory close review, central merge; hub rebuilds primary weave/sdlc. Report verified Git merge and PR. Never apply local-gates label.

Targeted commands:
    go test ./cmd/weave/internal/refresh ./cmd/weave -run 'Test.*(HelperRefresh|SuiteHelper)' -count=1
    go test ./cmd/sdlc -run 'Test.*(HelperRefresh|SuiteQueue.*Supplier|SuiteQueue.*Policy|HelperCapability)' -count=1
Adjust names to actual tests, then existing catch-up/helper/queue controls and race new lease cases. Required full suite: sdlc local-check run --name suite on pushed final head. No timeout increases.

## Architecture and residue

ARCH-DRY reuses lease, Git observations and supplier resolver. ARCH-PURE keeps eligibility/rendering deterministic. ARCH-PURPOSE refreshes before tickets and identifies each legacy supplier. ARCH-MOCK uses stateful/private native fixtures. ARCH-CONSTRAINTS bounds optional setup and deduplicated status. ARCH-SECURE verifies environment identity, strict Git framing, never prints credential-bearing remote stderr. ARCH-ORDER enumerates observed -> captured target -> revalidated -> ff -> confirmed/uncertain, including interruption. ARCH-FUNERAL creates no cache/updater/new lease family: existing permanent setup inode keeps its lifecycle; temporary roots cleaned by owners. ARCH-EVOLVE leaves ticket storage/frozen JSON unchanged.

## Approval

Pending TL approval of suite boundary, narrow weave mode, optional fallback, current-vs-launch wording and private proof/post-land rebuild. No implementation started.

## Revisions

### 2026-10-23 09:08 PDT — TL approval

Authenticated ops:0 approved the plan, including the narrow weave mode, new-suite setup boundary across all launch paths, existing setup lease, clean-main-only captured fast-forward, skip-and-continue behavior, read-only per-ticket current-supplier diagnostics with unknown launch revision, and private conformance plus post-land primary rebuild. Proceed to change-code. The pending statement above records the original proposal state.
