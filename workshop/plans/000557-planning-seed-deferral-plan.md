---
collaboration-mode: discussion
---

# Planning refresh seed deferral implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3. Execute using superpowers-executing-plans; bounded exploration is complete. Steps use checkboxes. One atomic issue-close review boundary, no Mx milestones.

**Goal:** A consumer start-plan can catch up its sources without regenerating seeded entrypoints on its resting branch, and pending regeneration can reach main through a separate operator-verified consumer change.

**Architecture:** Prevent seed publication inside the compiler, using its existing deferred-output machinery. Planning mode also keeps the existing protected-settings deferral. Ordinary compile/refresh stays explicit and produces the reviewable consumer change; no snapshot/restore or automatic commit.

**Tech Stack:** Go, Cobra, existing weave plan/ownership/setup components, native temporary Git repositories and filesystem fixtures.

## Decision requested from the hub

Recommend the seeded-output class: defer **all Seed and SeedOnce destinations**, tracked or untracked, before writing. This avoids a Git-tracking observation race and covers tracked workflow updates, mode/symlink changes and absent seeds without restoring caller bytes. Existing managed/ignored composition continues normally. Confirm this scope; a blanket guarantee against every tracked writer (including shipped generator outputs and .gitignore management) would require a broader publication policy, not a hidden reset added here. Any unrelated dirty-state refusal remains a refusal, never automatic cleanup.

Recommend full-flow review because this crosses start-plan's automatic refresh and compiler publication contracts. After approval use change-code's normal gates (no estimate before plan acceptance). No actual workflow regeneration is included in #557.

## Alternatives

1. **Prevent seeded publication in planning mode — recommended.** Reuse typed actions, OutputCurrent and DeferredOutput; failure or process death cannot publish a suppressed seed.
2. **Generic tracked-output deferral.** Wider semantics, including managed retirement, shipped outputs/sidecars and tracked .gitignore; substantially broader than the reported seeded-entrypoint failure. Do this only if the hub requires that stronger scope.
3. **SDLC snapshot/restore after refresh.** Reject: an interrupted refresh can leave writes behind, legitimate upstream fast-forwards change the restoration baseline, and generic restoration can erase caller edits. The live workaround remains manual recovery, not the implementation.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| compilePublicationPolicy (normal / settings-deferred / planning) | cmd/weave/deferred_seeds.go | new |
| seedDeferralPlan: retained actions plus deferred Seed/SeedOnce candidates | cmd/weave/deferred_seeds.go | new |
| plan.DeferredOutput action | cmd/weave/internal/plan | existing, reused |

The policy is invocation-local, not persisted state. The pure planDeferredSeeds classifier replaces every Seed/SeedOnce action with DeferredOutput at its destination, preserving order and retaining the original typed action for comparison. Nonseed actions remain identical. DeferredOutput retains prior ownership/recovery evidence and blocks overlapping publication; do not invent seed ownership. The settings-only mode remains compatible.

| Integration | Lives in | Status | Wraps |
|---|---|---|---|
| OutputCurrent | cmd/weave/internal/plan/current.go | existing, reused | native currentness observation |
| deferredSeedOutputs | cmd/weave/deferred_seeds.go | new | existing FS, OutputCurrent, diagnostic writer |
| compilePreparedWithPolicy (rename/extend compilePreparedSettings) | cmd/weave/main.go | modified | one existing compile/publication path under setup lease |
| refresh planning mode | cmd/weave/refresh.go | modified | refresh.RunWithOptions and compile policy |
| installed catch-up arguments | cmd/sdlc/catchup.go | modified | existing installed-weave process seam |
| explicit regeneration procedure | README.md, atlas/workflow/weave.md, command help | modified | existing weave compile and SDLC consumer PR flow |

No new external service or binary dependency. Reuse real private Git remotes, OSFS and the existing refresh/client seams. No live fleet checkout or :0 stage is touched by the isolated proof.

## State and recovery

Actions: start planning; regenerate on a dedicated consumer issue branch; repair the specifically reported input/build failure; retry.

| Observation / failure | Visible result | Recovery action | Possible loss |
|---|---|---|---|
| Planning seed already current | No seed write; planning proceeds | Start planning | None |
| Planning seed stale, absent, different mode or symlink | Name each destination as deferred and show regeneration route; planning proceeds | Regenerate on a dedicated consumer issue branch | None; destination/index bytes remain as observed |
| Dirty checkout before refresh | Existing preflight refuses before updates | Preserve/commit caller work, then retry | None; never auto-restore |
| Refresh advances sources then fails | Existing partial-progress refusal; seed destinations remain untouched | Repair named failure, retry | Completed fast-forwards remain |
| Compare/preflight/apply cancelled or fails | Existing error path; no deferred seed effects | Repair/retry | Existing nonseed partial-progress semantics remain; no seed rollback needed |
| Explicit ordinary compile on dedicated issue branch | Real workflow diff, branch/HEAD preserved | Review/commit it through its own SDLC PR; request operator verification | Only intentional generated changes |
| Older installed weave lacks new planning flag | CLI refuses before refresh effects | Rebuild the installed tool using existing source/build procedure | No seed mutation |

A source-unreadable seed retains existing OutputCurrent/apply no-op semantics; do not report invented desired bytes. Deferral is not a permanent decline and does not clear the pending update. No new automatic branch, issue, stash, patch file, approval label or remote publication.

## Ordering, authority and envelope

ARCH-ORDER: choose policy once; finish action construction and generation staging; apply seed/settings deferral **before CheckManaged and every ApplyManaged pass**; then use the existing publication order. Suppression rather than absence observation makes seed protection independent of Git index races. Existing setup lease is the single owner of cooperating compiler work; no new lock or lifetime. Arbitrary external writers remain outside that cooperative contract, but this mode executes no seed writes with which they can race.

ARCH-SECURE: manifest actions and current destination identity are inspected through the existing parser/FS seam. Never follow a destination symlink to overwrite another file and never restore HEAD over caller contents. No credentials or machine settings are created. Native proof uses portable temporary roots only.

ARCH-CONSTRAINTS: on-demand planning CLI, not a hot event path. One pass over A actions; one OutputCurrent observation per S seed; O(A + S + compared bytes), no per-seed Git subprocess. Count FS observations for N and 2N seeds to reject rescanning. Keep existing refresh limits/timeouts and sequential setup lifetime; no background work or new fan-out. No new writer-side storage budget is needed.

ARCH-DRY/PURE: classifier is pure, currentness/printing is the thin IO shell, and compilation still has one apply path. ARCH-MOCK: native owned resources replace production Git/FS state through existing seams. ARCH-PURPOSE: both seed action kinds and the explicit consumer regeneration route are delivered. ARCH-FUNERAL: no new durable artifact family; comparison state dies on return, generation stages retain existing cleanup, explicit updates use ordinary issue/PR/archive lifecycle. ARCH-EVOLVE: no persisted field, inventory format or seed semantics changes; old settings-only option and ordinary compile behavior stay valid.

## Function-level verification contracts

| Function / seam | Adversarial strategy and mechanical guard |
|---|---|
| planDeferredSeeds | Generate mixed action orders; assert every Seed/SeedOnce destination is represented only by DeferredOutput, all other actions/order unchanged, and repeated classification is stable. Pure tests, no IO mocks. |
| OutputCurrent | cmd/weave/internal/plan/current.go | existing, reused | native currentness observation |
| deferredSeedOutputs | Native current/stale/absent/edited/mode/symlink/missing-source inputs; independently compare destination bytes/mode/link and index before/after, and count constant-per-seed FS observations at N/2N. |
| compilePreparedWithPolicy | Run real managed publication with deferred seeds and unchanged ignored composition; force failure/cancellation around publication and require no deferred-path mutation or lost inventory evidence. Existing settings deferral remains covered. |
| refresh RunWithOptions callback | Bare-remote consumer/substrate capture reproduces the actual compiler-provenance pipeline diff, advances sources and continues; tracked consumer bytes/index remain unchanged and every pending path is named. |
| catchUpBeforeWork / startPlanBranch | Native end-to-end catch-up reaches issue branch creation without seeded dirt on the resting branch; existing obsolete-build/partial-failure refusals still apply. Installed argv test enforces planning mode. Use available Claude/Codex environment controls without launching real agents. |
| ordinary weave compile / refresh | On a separate temporary consumer issue branch, regenerate the captured pending update and assert the real workflow diff is reviewable, while HEAD/branch do not auto-commit or merge. |

## Implementation and proof — one atomic boundary

- [ ] Capture the verified 2dc46703 → 5785330c workflow change in a sanitized fixture. Add native planning refresh/start-plan regression first and observe the current dirty-tree refusal.
- [ ] Add typed publication policy and pure seed classification; add --planning to refresh while retaining --defer-settings. Wire installed SDLC catch-up to refresh --planning. Defer before publication, compare/report through existing helpers; red-to-green, then commit the coherent behavior.
- [ ] Exercise policy, ownership, partial failure/cancellation, counted observations and explicit regeneration contracts. Keep ordinary compile/refresh and settings protection controls; no live :0 touch.
- [ ] Document the dedicated consumer workflow-update issue/branch, weave compile, diff inspection and separate local-gates PR. Say the operator applies local-gates-verified; agents never do. Include README and help discovery pointers.
- [ ] Push the code head, retain the recorded suite, then close through the mandatory review and land through sdlc merge --yes. Any local-gates guard refusal is reported to the hub; never self-label or bypass it.

## Acceptance and review point

Hub review must confirm seeded-output scope versus the stronger all-tracked-writers alternative before change-code. Request isolated native production-path proof as the live check; no consumer workflow changes or live stage are needed to demonstrate prevention. Actual consumer regeneration stays a distinct operator-verified change. If preventing all tracked writers is required, revise this plan explicitly before implementation.

## Revisions

### 2026-10-22 — hub-approved scope

Authenticated hub approval b2a58869-feeb-45f0-b41c-973fb0167e90 confirms all Seed/SeedOnce suppression in planning mode, existing settings protection and explicit ordinary compile for the separate operator-verified consumer PR. The wider all-tracked-writers policy (Ship, managed retirement, .gitignore) is a known follow-up, not part of #557; file that follow-up if one of those families is actually observed dirtying a resting branch. No speculative follow-up issue is filed now. Proceed through change-code; no implementation preceded approval.
