---
collaboration-mode: discussion
---

# Workingdoc settlement implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to choose bounded delegation or warm-context execution. Steps use checkbox tracking. Do not implement before TL approval and `sdlc change-code`.

**Goal:** Preserve an operator-approved side document as a tracked, versioned workingdoc, retain a pointer in the live document, and make subsequent drift visible.

**Architecture:** A pure projection and settlement classifier consume snapshots; a thin coordinator publishes one selected copy and stamps the source through Couch's guarded Apply. Git's committed blob is the baseline, so no settlement journal, duplicate content hash or new registry is needed. Product repositories use the existing trunk publication mechanism; the ops operating repository commits and pushes on its primary main checkout.

**Tech Stack:** Go/Cobra, YAML nodes plus byte-preserving scalar edits, existing sidedoc link helpers, SessionSync/Couch Apply, Git and SDLC trunk publication.

## Decisions and alternatives

- **2026-10-21, TL decision:** product repositories publish exactly the clean copy through `sdlc trunk publish`; extend its allowlist with `workshop/workingdoc`. The ops operating repository instead makes a plain selected-path commit and push on main from its primary `:0` checkout. Infer the mode from repository identity, never a caller flag. Do not change ops files during implementation.
- **2026-10-21, TL decision:** an unresolved human question/comment must never disappear through settling. List and refuse remaining active commentary, unaccepted edit proposals and malformed markers. Alt+a acceptance produces clean text before settlement; the verb does not treat its invocation as accepting proposals. Literal code and anchors remain intact. No automatic marker resolution or acceptance flag.
- **Recommended baseline:** compare normalized source bytes with the exact destination blob at `settled-into`'s commit. An mtime baseline misses edits and a separate digest would duplicate Git's fact. A live mirror would erase the operator's deliberate settlement boundary.
- **Recommended retry:** confirm any already-published identical copy, then retry only the Couch stamp. Do not roll back a confirmed publication or introduce a recovery journal. A differing copy or unreadable remote is a conflict/unknown, never evidence of successful settlement.
- **Publication seam gap for review:** current trunk publication compares against the branch merge base. Re-settling from a stale hub checkout needs a pinned per-path baseline instead. Add an optional `--expected-commit SHA` to trunk publish: compare the selected path against that commit's blob, require an ancestor of fetched main, preserve existing staged/branch-edit refusals, and still accept an already-identical remote copy. Omission keeps existing behavior. Add `--json` returning a versioned `{path, commit}` receipt after confirmed publication/readback; do not parse human output. This is a narrow extension of the existing CAS mechanism, not a second publisher.

## Command contract

`tl-lead doc settle SIDE-DOC [--to PATH]` is an explicit write/publication command. Paths are relative to the source repository root (absolute source paths are accepted after containment checks). Source must be a regular Markdown side document within that repository; support existing `type: side-doc` pages as well as new `type: workingdoc` pages. Validate known metadata and preserve unknown frontmatter bytes. Do not automatically convert existing live side docs or invent missing business owners/topics.

For the first settle, the default is `workshop/workingdoc/YYYY-MM-DD-<source-slug>.md`, using the invocation's clock and removing the dated side-doc prefix when present. Once stamped, omission of `--to` uses the recorded target regardless of today's date. An explicit different target after settlement refuses; relocation is a deliberate future operation. An occupied first destination refuses unless it is the identical clean copy whose publication can be confirmed. Re-settle overwrites the previous copy only when its published and local bytes are the last recorded baseline or the desired bytes; independently edited content refuses with the path to reconcile.

The copy excludes `settled-into` to avoid a self-referential commit, preserves the rest of the document, and rebases relative Markdown links with `sidedoc.RebaseLinks` and index paths with `sidedoc.RebaseIndex`. The original source remains the live editing surface. Report links into ignored/untracked side material and dangling links, so the operator sees references that are not durable; never silently rewrite their content or copy unrelated side trees. Preserve fenced examples and inline code.

The source pointer is one plain scalar, `settled-into: <repo-relative-path> @ <full-commit>`. Validate its destination and full Git object ID; read its blob before treating it as authority. Source writes use Couch Apply with the exact whole-file hash of the pre-publication snapshot. Unknown keys, comments, order and body bytes survive. Duplicate or non-scalar known fields refuse; for a managed pointer whose YAML spelling cannot be updated without loss, name it and refuse rather than reserialize the whole frontmatter.

`tl-lead doc status` is read-only. Resolve root `session.md` through the existing hub path logic and walk its `session.side/` Markdown tree, skipping hidden/tool directories and refusing symlink entries rather than following them. Include legacy side-doc and workingdoc pages, sort by relative source path, and show `unsettled`, `settled`, `drift`, `conflict` or `unverifiable`, with target/commit when available. Drift means source projection differs from the recorded blob; target edits/missing targets are reported separately as conflict. Metadata-only pointer stamping does not create drift. An absent side tree is an empty result; a read or Git error is visible and exits nonzero. Status makes no Couch acknowledgments or writes and does not fetch or publish.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| `WorkingdocMetadata` and `SettlementPointer` | `cmd/tl-lead/doc_model.go` | new |
| `WorkingdocProjection` | `cmd/tl-lead/doc_project.go` | new |
| `SettlementState`, `SettlementEvent`, transition/classification | `cmd/tl-lead/doc_model.go` | new |
| marker scanner diagnostics | `cmd/tl-lead/doc_markers.go` | new |
| pinned-path publication decision | `cmd/sdlc/trunkpublish.go` | modified |

- One source has zero or one settlement pointer, naming one destination at one commit. Metadata parsing validates the known fields once; projection, status and writes share it (ARCH-DRY). A workingdoc is a living one-topic page, distinct from a pensive's dated thought or a project's execution record. The prototype carries owner/topic, collaboration mode and optional settlement pointer; its body stays freeform with optional decisions/open questions, without empty required sections.
- Projection is deterministic in source bytes, source/target coordinates and a supplied link-existence inventory. It removes settlement metadata, validates active marker absence and rebases links. It performs no filesystem or Git operations (ARCH-PURE).
- State is invocation-local; the pointer plus Git blobs reconstruct subsequent invocations. Only the transition model grants the next effect. Expected edit and failure sequences below are the test surface (ARCH-STATE).
- The marker scanner follows the authored grammar in `construct/local/fix/review-convention.md`; there is no Go cleaner to reuse. Scan once, skip literal fenced/inline code, list line/column and marker kind, recognize legacy chains, and refuse malformed openings. Preserve standalone anchor text and never discard an open item. Table-driven and fuzz tests cover each grammar shape.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| `DocRepository` | `cmd/tl-lead/doc_repository.go` | new | typed Git snapshots/blob lookup, publication mode, SDLC/operating Git publication |
| `SessionSync` | `cmd/tl-lead/sessionsync.go` | reused | Couch Peek, Status, RegistrationSnapshot, Apply |
| file-aware stateful SessionSync fake | `cmd/tl-lead/doc_fake_test.go` | new | per-file bytes, hashes, registration and lock state |
| stateful document repository fake | `cmd/tl-lead/doc_fake_test.go` | new | commits, remote head, selected blobs, local changes and injected unknown outcomes |
| Cobra command assembly | `cmd/tl-lead/doc.go`, `main.go` | modified/new | user arguments and output |
| trunk-publish CAS and JSON receipt | `cmd/sdlc/trunkpublish.go` | modified | existing TrunkFile transaction and verified Git object |

Production and tests call the same interfaces. Reuse existing Git argv helpers for reads, sidedoc path/link utilities, SessionSync.Apply and editRecord verification; avoid the migration wrapper that discards lock/stale exit identities. Product publication uses the existing portable Git test repositories and trunk tests; ops-mode tests use an isolated bare remote and primary checkout, never ops:0. Real Couch conformance is a hub-arranged check after SHIP, backed by recorded fixture replay.

## Publication, ownership and concurrency

1. Validate everything available before effects: source/destination containment, regular paths, frontmatter/markers, pointer baseline, target state, repository mode and source/destination registration/locks. Require the copy to be outside ignored/side trees and eligible for the selected publisher; operating destinations are non-issue workshop prose or Markdown below `teams/<team>/projects/`. An unregistered source refuses with the existing Couch register instruction. If the destination is registered, prepare its bytes through Couch Apply and its own whole-file CAS too; never bypass an operator lock on either file. Unregistered destinations use an atomic selected-path write after checking the observed old bytes. No per-side-doc subprocess lookup loop: one registration snapshot per phase where registration information is needed.
2. Hold an OS-released exclusive settlement lock in the Git common directory for the complete invocation; linked checkouts share it. This serializes this verb's copies, publication and stamps. Existing Git/SDLC transaction locks still own their resources; do not acquire their lock recursively or bypass their API.
3. A quiet editing window is required for the selected source/target. Couch has no editor lease spanning the Git publication; a late source edit/lock can therefore leave a published snapshot awaiting a stamp. Preserve it and report the exact outcome. Couch's whole-file CAS prevents claiming that a newer edit was settled.
4. Product mode uses selected non-issue workshop Markdown only: `workshop/workingdoc` plus the existing trunk allowlist. The optional expected commit is the previous settlement baseline. On initial creation use the existing merge-base predicate and absent/identical destination checks. The publisher preserves index, unrelated work, branch and its output copy, as its existing contract requires.
5. Operating mode is limited to the Git-origin identity of the ops repository, verified primary checkout/slot0 and branch main; derive identity from Git remote/workspace metadata, not a directory spelling, prose grep or a user flag. Refuse a brain repo and unknown mode. Do not invent a generic operating-repo registry. The implementation must document this explicit current scope for future extension.
6. Ops mode refreshes main before the selected write; refuses divergence/conflicts and unrelated staged or dirty tracked work. Commit only the selected Markdown path with the agent trailers, then push normally. No force push, automatic conflict resolution, stash, unrelated commit or branch switch. A failed/uncertain push retains the local commit and source pointer; retry confirms the remote selected blob before stamping, without duplicating the commit. Refuse if the new local commit contains other paths.
7. Immediately before the source stamp, verify the published blob equals the planned copy; then Apply only the pointer against the original whole-file hash. On uncertain Apply reread the pointer and source. A confirmed matching pointer is completion; otherwise return an explicit retry/reconcile result. No background operation survives return.

## States and interrupted outcomes

Actions: resolve markers, settle/retry the same command, reconcile source/target, release operator lock.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| No pointer | unsettled | settle | none |
| Pointer and current source/copy match baseline | settled | edit, then settle when ready | none; history retained |
| Source projection differs; target still baseline | drift | settle | prior copy remains in Git |
| Human marker, unaccepted proposal, malformed metadata/marker or invalid path | refused, with file and items | resolve markers or reconcile source/target | none; no effects |
| Source locked before effects | locked | release operator lock, retry | none |
| Target has independent edits, missing baseline, or conflicting remote path | conflict | reconcile source/target | none; never overwrite independent edits |
| Copy prepared but commit/publication fails | not settled; exact local copy/commit named | retry same command | no source loss; local prepared copy remains |
| Push result unknown | publication unconfirmed | retry same command and confirm remote | may already be published; never roll back or blindly duplicate |
| Publication confirmed, stamp locked/stale/fails | copy published; source not stamped | release lock/reconcile source, retry | published snapshot retained; newer source edits remain unsettled |
| Stamp response lost | stamp unconfirmed | retry and read exact pointer/source | may already be settled; retry must not create a new commit |
| Source changes after publication | drift or source not stamped | reconcile source, then settle | none; published older snapshot remains |
| Git/Couch unavailable, invalid reply, missing object | unverifiable | retry same command after service repair | none; unknown is not absent |

Transitions: snapshot -> planned -> copy prepared -> publication confirmed -> stamp confirmed -> done. Any failure stops effects. A retry reconstructs from source, local/remote blob and pointer; it can skip only a confirmed identical publication. If source changed after an unstamped publication, surface conflict rather than silently overwrite or label that newer source settled. This intentionally requires a quiet old snapshot during repair; no durable recovery state is introduced.

## Operating envelope and lifecycle

- Manual batch commands, off the editor hot path. Reuse the trunk publisher's 1 MiB selected-document limit for source and destination, bounded reads even for growing files. An oversize page refuses before writes. No truncation.
- Status is one directory traversal plus one read/projection per eligible Markdown file, O(N + total bytes), with one batched Git baseline read rather than a process per doc. Snapshot registration once if needed; do not enumerate the entire repository or open every registered document. Count directory/file reads and external calls at N=8 and 2N=16 to reject quadratic growth.
- Couch calls reuse a cancellation-aware bounded adapter with the existing five-second migration call budget, named as the chosen interactive limit; Git remote publication reuses trunk's two-minute remote bound. One sequential settlement at a time per repository. No concurrent fan-out or independent asynchronous writes. Status is cancellable during its traversal; failed/cancelled effects cannot become success.
- Durable copies are authored knowledge: one path per first settlement; re-settle updates the same path, with Git as history. The operator archives/removes a copy through normal repository editing; status names a missing target instead of recreating or deleting it automatically. Source pages and their lifecycle remain with their parent session/side tree. Temporary payloads/lock handles die with the command; no per-run receipt directory, cache or append-only registry is created.
- Trust boundaries are command paths, hand-edited YAML/pointers, Git responses and Couch replies. Parse typed inputs at the seam, use argv arrays/literal pathspecs, refuse traversal/symlink and ambiguous YAML, preserve unknown fields, and verify exact published bytes. No credentials are read or logged. No vocabulary lifecycle enum changes are needed: status here is a derived observation, not a new persisted noun status set.

## Implementation tasks (one close boundary)

### Task 1 — prototype and pure model

Files: create `construct/datatype/workingdoc.md`, `cmd/tl-lead/doc_model.go`, `doc_model_test.go`, `doc_project.go`, `doc_project_test.go`, `doc_markers.go`, `doc_markers_test.go`.

- [ ] Write failing table tests for pointer parsing/formatting, duplicate and unsupported YAML fields, unknown-byte preservation, first/re-settle target choice, and all state/refusal rows.
- [ ] Write marker tests from the authored table: commentary, proposals, bare deletion, legacy chains, malformed delimiters, literal robot glyph, fences, inline code and anchors; add a fuzz invariant that refusal never changes source bytes.
- [ ] Implement the pure parser/projection/classifier and byte-preserving pointer edit; reuse sidedoc rebasing. Add workingdoc prototype with type/name, discovery description, collaboration default, owner/topic, pointer contract, freeform body, creation/editing guidance and search recipes. Add registry discovery/schema verification using existing datatype test surfaces.
- [ ] Run `go test ./cmd/tl-lead ./cmd/datatype -run 'Workingdoc|Settlement|DocMarker|EveryType|Collaboration' -count=1`; confirm red then green. Commit the coherent pure slice.

### Task 2 — publication seam and guarded settlement

Files: modify `cmd/sdlc/trunkpublish.go`, `trunkpublish_test.go`, `helptext` if needed; create `cmd/tl-lead/doc_repository.go`, `doc_repository_test.go`, `doc_fake_test.go`, `doc_settle.go`, `doc_settle_test.go`; extend `deps` in `main.go` narrowly.

- [ ] Start with failing existing-publisher tests for workingdoc allowlisting, expected per-path baseline, stale main, identical retry, unknown push and JSON receipt. Retain existing unselected/staged/branch-committed and unsafe-path refusals. Implement the additive trunk contract.
- [ ] Build the stateful repository/SessionSync fakes, injecting effects at each transition boundary. Drive production transitions through first settle, no-op retry, re-settle, source lock/edit arriving between publish and Apply, commit-before-stamp interruption, uncertain push/Apply and destination conflicts. Assert independent invariants: stamp only names confirmed blob; source body/open items survive; only selected destination is published; no duplicate commit on retry.
- [ ] Implement typed repository mode detection, selected-path publication and post-publication verification, plus the shared settlement lock. Verify ops mode against temporary bare remotes: clean main, unrelated dirty/staged work, divergent main, local commit/push interruption and duplicate retry. Product mode uses the real trunk executable seam or existing temporary Git fixtures for conformance, not a stateless call-count stub.
- [ ] Add counted publication/read tests for batching and cancellation/deadline tests synchronized by readiness or test deadlines, never helper-start wall time.
- [ ] Run focused new tests and all existing trunk publication tests, then full `cmd/tl-lead` and `pkg/golint`; commit the coherent feature slice.

### Task 3 — command surface, status and discovery

Files: create `cmd/tl-lead/doc.go`, `doc_status.go`, `doc_status_test.go`; modify `main.go`, `help.md`, `README.md`, `atlas/workflow/tl-skills.md`; document trunk additions in `atlas/workflow/sdlc-binary.md`. Existing atlas index links already reach both pages.

- [ ] Add failing Cobra tests for both verbs, positional/flag validation, deterministic outputs and exit behavior. Add status fixtures for unsettled/settled/drift/conflict/unverifiable, nested side docs, ignored tool folders and unreadable paths, plus N/2N operation-count assertions.
- [ ] Wire commands and read-only discovery/classification using the shared pure projection. Status must not acknowledge Couch events, edit files or publish. Update help and maps with first settle, re-settle, marker refusals, publication modes and exact recovery instructions.
- [ ] Run focused commands/package tests and `datatype lint`; inspect source diff and emitted help. Commit and push the implementation before the recorded suite.
- [ ] Run `sdlc local-check run --name suite` on the pushed head. On green, run the binary-owned `sdlc close --issue 507 --verified '<observed behavior>'` (no ad-hoc duplicate boundary review).
- [ ] Arrange the live smoke with the hub after SHIP: stage final build and supply revision evidence; operator/TL create a disposable side document in ops, settle to a disposable tracked copy, edit/re-settle, verify status and marker/lock refusal. Replay its ordered source/lock/remote evidence through the stateful fixture; record outcome. Do not manipulate the hub's real files yourself. Product publication is verified against an isolated product fixture plus recorded command proof; the hub decides any additional live check.
- [ ] Clear stage, enter `sdlc merge --yes` promptly and report landing/PR to the hub. Keep #514's completed outcome separate from #507 evidence.

## Approval requested

The TL's publication-mode and unresolved-marker decisions above are binding. Review the pinned-path trunk extension, ops identity boundary, copy metadata/link policy and visible interruption outcomes before implementation. Any change after approval is appended under Revisions rather than silently replacing this design.

## Revisions

### 2026-10-21 — approved design and plan-quality refinement

TL approved the initial design and pinned-path/JSON publisher seam (receipt 7f2905de-b94c-4cc8-bb53-582cc8aeb5b7). Plan-quality round 1 then raised PQ-1 (publication payload authority) and PQ-2 (function test-strategy contract). TL approved the payload correction below (receipt 18ce25ca-d968-4481-8f80-879f39ab36ac). These refinements supersede the affected publication and testing instructions above; state outcomes and scope remain unchanged.

**PQ-1: bind effects to validated bytes, not a mutable path.** Add optional `sdlc trunk publish --expected-content-sha256 HASH`, validating its syntax and comparing it to the publisher's one bounded captured local read before any publication effect. A mismatch refuses with no target or remote write. The existing publisher already writes its captured buffer; retain that immutable buffer through every CAS retry. Omission preserves current behavior. tl-lead passes the hash of its pure projection and the expected prior path commit when applicable. A target edit before publisher capture is refused; an edit after capture remains in the live file and cannot change the selected published bytes. Do not restore or clean that newer target. JSON receipt still binds the confirmed commit/path, whose blob must match the captured projection.

Ops publication uses a private temporary Git index loaded from the captured HEAD and the validated in-memory payload (hash-object from stdin, update-index for the one selected path). Verify the prospective tree against that HEAD before the plain commit, commit from the private index, then verify the resulting commit's parent, tree and selected blob before push. If HEAD moved or the commit differs, stop before push and name the retained local outcome; no rollback or automatic duplicate commit. Do not copy the live file into this index or use a commit form that rereads its worktree path. The real index and any concurrently edited live target remain untouched. Push/readback confirmation and source CAS stamp follow the original sequence. Private index, temporary Git environment and payload files are scoped resources removed at command return, including cancellation.

**PQ-2: executable test strategies.** The earlier Task 1–3 examples are historical design context, not an inventory to transcribe into cases. Implement tests from these named function contracts; this table is the current testing instruction. The tasks' file ownership, red/green/commit sequence and one close boundary remain in force.

| Function | Adversarial strategy and independent guard |
|---|---|
| `parseWorkingdoc` / `parseSettlementPointer` | Fuzz hand-edited/truncated/duplicate YAML and pointer text; malformed known fields refuse and unknown bytes round-trip exactly. |
| `projectWorkingdoc` / `stampWorkingdoc` | Fuzz documents and path coordinates; projection never alters source input, excludes its own settlement pointer, and stamping changes only the one managed scalar. |
| `workingdocMarkers` | Fuzz the authored marker grammar with literal-code contexts; unresolved or malformed active markers are reported with location, never resolved or silently discarded. |
| `transitionSettlement` | Stateful generated event sequences with failure/cancellation at effect boundaries; independently assert that a stamp can reference only confirmed exact publication and retries cannot duplicate a confirmed commit. |
| `classifyWorkingdoc` | Vary source, recorded blob and current target independently; bytes, not mtime, decide drift and unverifiable evidence can never become settled. |
| `pinnedProsePublishDecision` / `runTrunkPublish` | Mutate target/remote before and after captured read with controlled hooks; only the expected immutable payload reaches the remote, unrelated/index/live edits survive, and no-option behavior matches existing publication tests. |
| `publishOperatingWorkingdoc` | Real temporary Git repository and bare remote with controlled HEAD/target mutation and push uncertainty; verify selected tree/parent before push and preserve live bytes/index on every stop. |
| `settleWorkingdoc` | Stateful Couch/repository fakes inject edits/locks and lost acknowledgments; the whole-file source hash prevents stamping a newer source and target hash binds publication. |
| `discoverWorkingdocs` / `readWorkingdocBaselines` | Count reads/traversals and external commands at N and 2N; one walk, bounded per-file read and batched Git access enforce linear growth. |

Existing behavior references verified during review: trunk capture/publication at `cmd/sdlc/trunkpublish.go:158` and `:223`, merge-base baseline at `:182`, link projection at `pkg/sidedoc/rebase.go:44`, index projection at `pkg/sidedoc/index_rebase.go:18`. No new marker acceptance mechanism, journal or publication backend is introduced.

### 2026-10-21 — selected-index completion policy

TL approved this clarification (receipt cf9b4eed-e7c1-4f55-9c83-4cdfd56f807d): the private payload index remains isolated, but ordinary success must reconcile the selected real-index entry to the confirmed commit so the checkout is not left with a staged inverse change. Capture the real index before the private-index commit. After publication, take Git's index.lock, compare the complete captured index, and update only the selected entry. Unrelated entries and all live target bytes remain untouched. If another actor changed the index, preserve it entirely and print the exact selected-path reconciliation command. A typed confirmed-publication warning may carry that instruction while the coordinator independently verifies the published blob and performs its source CAS; it must not turn an unknown publication into success. This supersedes earlier wording that the entire real index is always untouched after successful publication.

The pre-target refresh required above is an explicit `DocRepository.Prepare` effect after source validation: product preparation fetches the published baseline objects without editing the checkout; ops preparation refreshes primary main with a safe fast-forward before preparing the live target, refusing divergence/dirty conflicts without stash/reset/rebase. A selected-only retained commit is validated for retry rather than discarded. Status remains local/read-only and does not fetch. This fills the existing caller obligation of the ops backend and makes ordering testable.
