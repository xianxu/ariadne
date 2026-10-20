---
collaboration-mode: discussion
---

# Root session file implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution strategy. Steps use checkbox syntax. The TL owns approval and the live migration; workers must not migrate the live hubs.

**Goal:** Make repo-root `session.md` and `session.side/` the hub interface, with a safe, dry-run-first migration from the old location.

**Architecture:** Keep explicit FILE arguments compatible and normalize omitted FILE arguments at the command boundary. A pure move planner rebases copied documents through existing sidedoc helpers. A thin executor copies exclusively, moves Couch registrations, verifies, and only then removes the old files. No recovery journal or new persistent protocol.

**Tech Stack:** Go, Cobra, filesystem fixtures, existing SessionSync fake and real Couch scratch conformance; existing Parley attention finder.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| statusArgs | `cmd/tl-lead/status_path.go` | new |
| sessionMovePlan | `cmd/tl-lead/status_move.go` | new |
| sessionMoveState / transitionMove | `cmd/tl-lead/status_move.go` | new |
| moveSessionFields / moveIgnores | `cmd/tl-lead/status_move.go` | new |
| RebaseIndex | `pkg/sidedoc/index_rebase.go` | new |
| syncRegistration | `cmd/tl-lead/status_move_registry.go` | new |

- `statusArgs` supplies an already-resolved default path when a command has exactly one fewer positional argument than its explicit-FILE form. It never guesses from a filename suffix. One parser serves all status commands; the existing argument validators run on the normalized arguments. Tests require no IO.
- `sessionMovePlan` owns one old/new session pair, zero or one side tree, copied document bytes, registration pairs, conflicts, and ignore additions. It maps the old session exactly and the entire old side-tree prefix to their destinations. Reuse `sidedoc.RebaseLinks` and `sidedoc.SideDir`; promote index relocation into `sidedoc.RebaseIndex(text, from, to, moved)` and retain the old `rebaseIndex` wrapper for stationary callers. The promoted helper parses only the actual discussions block, accepts separate old/new coordinates, and preserves all other frontmatter bytes; do not duplicate Markdown parsing. Preserve arbitrary unknown frontmatter and binary attachments.
- `sessionMoveState` grants one prepare/copy/switch/verify/remove effect at a time through `transitionMove`, using confirmed progress and success/failure/unconfirmed observations. Failed or uncertain observations revoke authority; retries reconstruct from actual state without a journal.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| hub session path resolution | `cmd/tl-lead/status_path.go` | new | Git top-level and legacy-file discovery |
| status move executor | `cmd/tl-lead/status_move_apply.go` | new | exclusive file creation, locks, ignore file, verification, cleanup |
| SessionSync registry operations | `cmd/tl-lead/sessionsync.go` | modified | existing Couch Rename / Registrations / Status / Apply seam |
| fakeSessionSync | `cmd/tl-lead/sessionsync_fake_test.go` | modified | file-backed multi-document registry, lock, baseline and owed state |

Resolve default session paths against Git's checkout top level, including from subdirectories. Omitted FILE requires a hub (`.hub` at root); an explicit FILE remains usable in existing scratch/non-hub callers. Reuse the existing Git runner, not another Git adapter. Do not create session.md as a side effect of a status read. The operator field remains optional tooling metadata, with its default from the existing git-user-name convention where a template/default is needed; migration preserves an existing operator value verbatim and does not invent an init verb.

Couch remains the registration owner. Use `Rename`, not a second registration that drops the old identity. Extend the reply model only as needed to verify the effective new binding; JSON readers tolerate unknown fields. The fake must store registrations per document and preserve identity, baseline and owed events through rename. Verify its model with real Couch on isolated scratch documents, never hub files.

## Command contract

`tl-lead status move [OLD] [--apply]` always targets `<hub-root>/session.md`. OLD is repo-relative or absolute inside the same checkout. With OLD omitted, discover exactly one `teams/*/tl-status-*.md`; list candidates and refuse ambiguity. If no legacy file remains and session.md exists, report already migrated after verifying the destination; a missing source and missing destination refuses. No cross-repo moves. An explicit OLD equal to session.md is an already-at-destination check, not a destructive move.

Every other status leaf accepts its existing argument list and its omitted-FILE form:

| Explicit form | Default form |
|---|---|
| `submit FILE`, `seen FILE`, `owed FILE`, `check FILE`, `clobbered FILE` | verb alone |
| `section FILE HEADING BODYFILE` | `section HEADING BODYFILE` |
| `thread FILE TEXTFILE` | `thread TEXTFILE` |
| `slots FILE` | `slots` |
| `instructions list FILE` | `instructions list` |
| `instructions link FILE POS TARGET` | `instructions link POS TARGET` |
| `index FILE` | `index` |
| `migrate-answers FILE` | `migrate-answers` |

Keep flags, stdin body behavior and stale/lock exit codes unchanged. CLI help uses `[FILE]`. One shared normalizer is attached explicitly to each status leaf; a table-driven test enumerates the leaves so a future leaf cannot silently miss the default.

Dry run reads and prints the old/new paths, all copy/link/index/registration/ignore effects, conflicts and exact next command. It writes no files, does not acknowledge operator edits, and never folds owed events. Applying also leaves unread operator edits unread; do not call MarkAllSeen as migration cleanup.

## Copy, switch, verify, cleanup

1. Snapshot the source session, full side tree (including nested side docs and assets), existing destinations, registrations and relevant locks. Resolve physical root once; refuse symlinks in source/destination paths or trees, non-regular entries, malformed session metadata, `.lock`, and unresolved `.clobbered` evidence. Refuse unrelated destination documents even when names happen to match; identical migration copies may be reused. Check source/destination registrations do not bind different document identities.
2. Plan all rebasing before writing. Rewrite relative session links, side-document `parent:`, Markdown/reference/image links, and `discussions:` entries from old to new locations. Relative links to stationary tracked artifacts must still resolve after moving to root. Preserve URLs, fragments, code examples and unknown metadata. Report dangling links and bare old-path mentions without guessing replacements. Scope rewrites to the copied tree and session; report external backlinks for manual attention rather than sweeping tracked project files.
3. Append missing anchored ignores to the hub's `.gitignore`: `/session.md`, `/session.side/`, and the known local status sidecars such as `/.session.md.ss-seen`. Preserve other bytes and old ignore rules. New copies must be ignored before publication; refuse tracked destinations. Dry run only prints the additions. Guard `.gitignore` against intervening edits with the existing unchanged-content writer.
4. Copy final bytes exclusively, preserving file modes and mtimes. Existing byte-identical copies are no-ops; differing copies refuse without overwriting either. Include the bounded seen-ID set at `seenPath(OLD)` → `seenPath(NEW)`; move no lock or clobber evidence and let Couch own its registry/shadow state. Keep all old files until every new file is verified.
5. Recheck locks and source hashes, then rename registered side docs and finally the session registration using SessionSync.Rename. Verify each effective binding and the final destination bytes. No new registration for unregistered ordinary side docs. The source session must already be registered: otherwise refuse before effects with `couch --session-sync register OLD --owner SLOT`, substituting the exact path and its frontmatter owner. Couch imports legacy state during that explicit registration; tl-lead never copies legacy shadow/state files. Missing or unreadable registry evidence is not absence. On an uncertain rename, stop and print the identical retry and inspection instructions; do not blindly roll back or forget either path.
6. Only after all registration moves and copy verification succeed, remove unchanged old files and empty old directories. Each removal checks the expected source bytes. Never recursively remove an uninspected directory. Preserve concurrent edits; refuse and report the exact remaining path. Retrying after partial cleanup verifies existing new files and finishes remaining old files. A new-only child after a cleanup interruption is preserved, never interpreted as a collision to delete. Success names session.md and instructs the operator to reopen the editor at the new path.

The old source stays intact until the switch and verification. After a confirmed switch, the new session is the working interface even if cleanup fails. Do not promise that a partially removed old tree is still usable. Preserve the original session source until the final cleanup step so explicit OLD can anchor retries; if it is already gone, only a verified completed new destination is accepted.

### Actions and visible state

Actions: run dry run; resolve a named conflict; wait for operator submit/close; rerun the same apply; reopen session.md; inspect registry after an uncertain response.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Lock, ambiguity, symlink, collision or stale source before effects | refused; old in use | resolve named conflict, rerun dry run | none |
| Failure or cancellation copying | old in use; partial identical copies | rerun same apply | none; both copies kept |
| Some child registrations moved; session not switched | old session in use; registry migration incomplete | rerun same apply | none; old and new bytes kept |
| Rename outcome unknown | switch unconfirmed | inspect registry, retry identical rename through apply | none removed; routing temporarily uncertain |
| Session registration and new bytes verified; cleanup interrupted | new in use; old cleanup incomplete | rerun same apply | none; obsolete copies remain |
| Source changes after snapshot | conflict; source preserved | close editor, inspect both copies, rerun dry run | no automatic merge; stale new copy may need explicit human removal |
| All cleanup complete | migrated | reopen session.md | old path removed intentionally |

No atomic editor lease exists in the current Couch seam. A process-lifetime mutex local to tl-lead would exclude another tl-lead, but would not exclude Parley's operator editor; it cannot prove the lock invariant. Require a quiet window with editors closed, check `.lock` and Couch holder before copy/switch/removal, and test lock arrival between steps. This guarantees refusal for observed locks, not exclusion against an editor opening after the last check. No work continues in a background task after command return.

## Operating envelope and principles

- **ARCH-DRY:** reuse SessionSync and sidedoc transformations. Defaults have one normalizer. The existing answer migration remains separate because moving qa answers changes their document type/marker; this move preserves document meaning.
- **ARCH-PURE:** snapshot → pure plan/transition → thin executor; pure tests need no process mocks.
- **ARCH-PURPOSE:** include defaults, help/rule, ignore file, nested links/registrations and attention check. The last live acceptance remains an explicit operator/TL step, not silently waived.
- **ARCH-MOCK:** extend the existing stateful fake, including a rename that publishes then returns an uncertain response. Real conformance uses temporary registered files and forgets them at cleanup.
- **ARCH-CONSTRAINTS:** rare interactive maintenance, twice for the current fleet; one bounded sequential operation per file and registration, O(total bytes + files + links). Scan only old/new trees, not the whole workspace. No hot event path, so N/2N hot-path tests are unnecessary. No network beyond local Couch. No CPU fan-out; any read/permission error stops with paths and preserves sources.
- **ARCH-SECURE:** paths, frontmatter, registry JSON and lock contents cross trust boundaries. Reject traversal, ambiguous roots, symlinks and malformed known fields. Unknown fields survive. No credentials created or logged. Tests isolate Git/Couch state from live hubs.
- **ARCH-ORDER:** production transition model authorizes effects; injected step hooks test interruption, lock arrival and registration uncertainty at every boundary. Copies are exclusive and retries reconcile actual state; no success from an unknown response.
- **ARCH-FUNERAL:** copied files replace their sources only after verification; leftover partial copies are named in output and collected on successful retry. No journal/cache family. Seen IDs retain their existing bounded lifecycle. New ignored session/side docs are operator-owned and explicitly removed when retiring the hub, after decisions enter tracked artifacts. Scratch conformance registrations are forgotten in cleanup.
- **ARCH-EVOLVE:** preserve unknown frontmatter, side assets and Couch record fields; new JSON response members are tolerated. No on-disk format change.

## Implementation tasks

### Task 1: Defaults at command boundaries

**Files:** create `cmd/tl-lead/status_path.go`, `status_path_test.go`; modify `status.go`, `instructions.go`, `index.go`, `migrate_apply.go`.

- [x] Test `statusArgs` with generated positional arities and a table of real leaves; assert exact argument preservation/insertion and reject every other count. Test `defaultStatusPath` against temporary Git hub checkouts from nested working directories; assert root selection and no file creation.
- [x] Run `go test ./cmd/tl-lead -run 'TestStatus(Default|Args|Path)' -count=1`; expect failures for missing default support.
- [x] Implement the shared normalization and Git-root resolution; update Usage/Args together. Keep file-reading errors observable.
- [x] Re-run focused tests and existing status/instructions/index/migrate tests. Commit with issue reference and Agent/Co-Authored-By trailers.

### Task 2: Pure move planning and link corpus

**Files:** create `cmd/tl-lead/status_move.go`, `status_move_test.go`, `testdata/status-move/before/` and `after/`; extract shared rebase/edit helpers from `migrate.go` only if needed.

- [x] Test `planSessionMove` by comparing a committed before/after relocation corpus byte-for-byte; assert mapped link targets remain identical entities and untouched metadata/assets remain byte-identical.
- [x] Test `sidedoc.RebaseIndex` on malformed and unknown YAML members using strict node validation and byte-preservation assertions; generate old/new path pairs to prove index coordinates. Test `transitionMove` with interruption/outcome event sequences and assert cleanup never precedes verified switching. Test `moveIgnores` by repeated application and exact unrelated-byte preservation.
- [x] Run `go test ./cmd/tl-lead -run 'TestSessionMove(Plan|Transition|Links)' -count=1`; expect undefined planner/model, then implement minimal planner and rerun to pass.
- [x] Commit pure implementation and fixture expectations.

### Task 3: Executor, Couch identity and retry behavior

**Files:** create `cmd/tl-lead/status_move_apply.go`, `status_move_apply_test.go`; modify `status.go`, `sessionsync.go`, `sessionsync_fake_test.go`, `sessionsync_registry_test.go`, `conformance_test.go` as needed.

- [x] Test `readSessionMove` across adversarial filesystem/registry snapshots; real temporary directories plus the stateful fake guard path containment, registration prerequisites and every refusal before writes. CLI tests compare all tree bytes and registry state before/after dry run and apply.
- [x] Test `runSessionMove` with a controlled step hook after every effect and a stateful registry fake that can publish then return uncertainty; rerun actual production planning/execution and assert source preservation until verification, no overwritten changes, stable identities/baselines/owed sets, and convergent cleanup. Interleave late edits/locks at the hook rather than using timing. Test `copyMoveFile` and `removeMoveSource` against changed bytes and occupied paths, asserting exclusive creation and refusal before destructive effects.
- [x] Implement thin snapshot/executor and wire `newStatusMoveCmd`; reuse exclusive create/unchanged-write helpers without silently copying their limitations into the contract. Do not reuse answer migration's MarkAllSeen behavior.
- [x] Run `go test ./cmd/tl-lead ./pkg/sidedoc -count=1`; expect pass. Run scratch-only `TL_LEAD_LIVE=1 go test ./cmd/tl-lead -run 'TestSessionSync.*Conformance' -count=1 -v` for the used rename/verification behavior. Verify cleanup forgot scratch registrations.
- [x] Commit; report any design delta to TL before implementing it.

### Task 4: Operator documentation and attention evidence

**Files:** modify `cmd/tl-lead/help.md`, `main_test.go`, `README.md`, `atlas/workflow/tl-skills.md`; retain existing atlas/index link. No live session edits.

- [x] Write failing help assertions for root session.md/session.side, optional FILE, operator metadata, move dry run/apply/quiet-window/retry commands, and the decisions rule. Replace active default examples while retaining old paths in migration examples and compatibility fixtures.
- [x] Document that session/side docs are ephemeral TL/operator interface; decisions are copied immediately into tracked issue/project Log/charter. Workers act from tracked issues plus explicit TL/operator instructions. Keep the rule in lead-owned prose, not duplicate worker-owned rules.
- [x] Record Parley evidence: `lua/parley/attention.lua` scans a registered parent's `.side/*.md` without a team-path assumption; `attention_model.lua` inherits the slot; `tests/integration/attention_finder_spec.lua` covers root discovery and nested documents on a later finder open. Preserve mtimes and registrations. Arbitrary subdirectories without side-document parents are not newly promised.
- [x] Run existing Parley attention tests with `make test-spec SPEC=ui/attention_finder` in its isolated test environment; no Parley code change is currently needed. Run `go test ./cmd/tl-lead -count=1`, inspect the default/path consumer sweep, and commit docs/evidence.

### Task 5: Review, recorded suite, staging and manual acceptance

- [ ] Push issue head with SDLC; run recorded full suite `sdlc local-check run --name suite` on the pushed head. Targeted checks precede it. Do not commit during source-checkout test runs.
- [ ] Offer `sdlc stage --build` to the TL for scratch migration smoke; staging is authorized only at the TL's requested time. Report returned binary revision and exact fixture commands. Never invoke move on either live hub yourself.
- [ ] TL and operator choose a quiet moment, close their editor locks, inspect dry runs, and apply separately in ops:0 and ops-echo:0. They verify parent/side links, registration routing, attention finder, ignored files, and reopening session.md. Capture their evidence in this issue. The manual last step remains required; do not close without it or an explicit scope revision.
- [ ] Run `sdlc close --issue 504 --verified '<returned test and manual evidence>'`; this owns the fresh-context boundary review. Resolve findings and rerun relevant checks/suite after code changes. Report ready state to TL; on SHIP use `sdlc merge --yes` and report returned PR/merge evidence. Clear a stage if it was not cleared by landing.

## Approval and revisions

2026-10-20: TL accepted command discovery, compatibility and refusal semantics, then directed copy-then-switch instead of a recovery journal. This plan adopts that simplification. Approval of the complete plan is still required before change-code. The two visible failure locations around the registration switch are called out explicitly rather than promising the old registration stays active after it has moved.


## Revisions

### 2026-10-20 — plan-quality findings PQ-1 through PQ-3

- **PQ-1, migration-state-continuity:** TL approved an explicit source registration prerequisite. `readSessionMove` calls read-only lookup, and refuses an absent registration before any copy with the exact register command and effective slot. Couch's existing register imports legacy shadow/state before its rename; see pair `cmd/internal/sessionsync/registry.go:331`. The session's previous implicit legacy-routing promise is superseded by this prerequisite. Ordinary unregistered side docs are still copied without registration; nested legacy session-sync docs likewise require registration rather than losing path-specific state.
- **PQ-2, relocation-transform-contract:** the existing helper at `cmd/tl-lead/migrate.go:357` assumes a stationary document and scans all `- doc:` lines. Supersede direct reuse with `pkg/sidedoc.RebaseIndex`: read only the validated discussions node/span, resolve targets from `from`, spell them relative to `to`, and preserve every other byte. Existing stationary migration calls the new helper through a wrapper. Malformed/duplicate known index keys refuse instead of being reserialized. No whole-frontmatter YAML marshal.
- **PQ-3, function-level-test-strategy:** replaced procedural case inventories with named function/adversarial-class strategies above. Additional named boundaries: `sessionMovePaths` resolves untrusted OLD candidates using real Git/path fixtures; `verifyMoveRegistration` compares typed identity/owner lookups without acknowledging content; `LookupRegistration` decodes unknown or malformed Couch replies with strict known-field checks and explicit null-only absence. Their tests use generated path inputs and stateful injected registry observations. `renderSessionMove` is checked through CLI golden output and exact exit status. No separate review boundary was added.
- TL approved the full plan before the first gate and approved the registration-prerequisite refinement after its findings. These refinements preserve the approved copy-then-switch/no-journal mechanism and manual live acceptance.


### 2026-10-20 — implementation checkpoint

Tasks 1–4 implemented and verified with package tests, real scratch Couch migration
conformance, and isolated Parley attention-finder tests. Snapshot IO is in
`cmd/tl-lead/status_move_snapshot.go`; execution is in status_move_apply.go.
Effect-boundary tests exposed late-source membership and destination changes
before cleanup; bounded rescans and destination rechecks now refuse those
observed changes. No new journal or editor lease was added. Source Couch metadata
companions remain owned by Couch in their old location; no legacy shadow/state
is copied by tl-lead. Final live acceptance/staging order is awaiting TL
clarification because the requested stage-after-SHIP sequence changes the
original pre-close manual contract.


### 2026-10-20 10:40 PDT — accepted close/stage ordering

TL approved superseding Task 5's earlier requirement to capture live migration
before close. The binding order is now green recorded suite and SHIP, stage the
final build, TL/operator migrate both hubs and return evidence, append evidence,
clear stage, then merge. Live migration is an explicit post-close/pre-merge hold;
closing alone never authorizes merge. The issue's current Done when and Plan
match this ordering. No implementation behavior changed.


### 2026-10-20 10:54 PDT — BR-1/BR-2 narrow corrections

TL approved the smaller fix: `moveMetadataFields` validates an exact plain
parent scalar before the existing link rebase. Other parent spellings refuse
before effects; no structural YAML rewrite is built. Routing-only owner and
operator validation applies to session-sync, preserving ordinary business
metadata. `verifyMoveIgnores` in status_move_ignore.go uses real Git checks
for every destination after missing ignore additions and before copies. It
does not manage an owned block or override negations. Integration regressions
assert named parent refusals and unchanged sources/copies for exposed paths.
The earlier parent-syntax and effective-ignore assumptions are superseded by
these explicit input and publication boundaries, with no journal or state change.


### 2026-10-20 11:10 PDT — BR-3/BR-4 publication and effect authority

Every published destination must remain ignored under the resulting Git policy.
Refuse any `.gitignore` in the source side tree or existing destination tree,
at every depth, before effects; these files can change the policy after a
pre-copy check. Freshly rescan destinations and run real Git checks before
switching and before each deletion, catching late policy files and exposed
paths. No prospective Git policy simulator or managed ignore block is added.

The earlier phase-only tracker is superseded by invocation-local
`sessionMoveState`: counts of confirmed copies and bindings, whole-move
verification, and cleanup progress. `transitionMove` consumes the observed
success, failure or unconfirmed result of its authorized effect and grants
exactly one next effect. Production dispatches only that grant. Failure or
uncertainty stops authority; retry reconstructs from actual disk/Couch state,
without a journal. Pure sequence tests independently assert all copies precede
switches, all binding checks precede verification, verification precedes every
deletion, and interruptions at every boundary grant no later effects. Existing
production interruption fixtures exercise the same dispatch loop.

The green-suite, SHIP, staging and operator/TL live-migration landing holds remain.
