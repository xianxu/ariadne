---
id: 000507
status: codecomplete
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-22
estimate_hours: 2.93
collaboration-mode: collaboration
card_mirror: '60a52bb63f7270e76c776b115b0e332f0da0f14b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T12:19:30-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 2.87
---

# workingdoc datatype and tl-lead doc settle/status: settle side docs into tracked copies, re-settle to update, drift visible

## Problem

Side docs (the operator–TL co-edited pages beside the session file) have become the main design medium, but they're
untracked, so settled decisions can be lost, and projects can't link them durably. Operator, 10-20: make a
`workingdoc` type; copy a settled side doc into a tracked place; projects link it as background. Concern raised:
once copied, the copy freezes while discussion continues in the side doc.

## Spec

- `construct/datatype/workingdoc.md`: a living, co-edited page on one topic, with owner, topic and a
  `settled-into: <path> @ <commit>` pointer.
- `tl-lead doc settle <side-doc> [--to PATH]`: copy with review markers stripped to `workshop/workingdoc/<date>-<slug>.md`
  (default) or the given path (e.g. a project's `design.md`), commit it, stamp the side doc's pointer. **Re-settling
  overwrites and commits the copy**, so git history is the copy's version history.
- `tl-lead doc status`: side docs edited since their last settle, so drift is visible.

## Done when

- The datatype, both verbs and their tests exist; `tl-lead --help` documents settling.

## Plan

- [x] Implement the workingdoc prototype, marker/refusal projection and pure settlement model, with regression tests.
- [x] Add guarded selected-copy publication and retry behavior through the existing product trunk publisher and ops main publication, with stateful fixtures.
- [ ] Wire doc settle/status, update help/maps, verify the suite, close, arrange the hub-owned live check, then merge.

Durable implementation plan: [workingdoc settlement](../plans/000507-workingdoc-settle-plan.md). Full flow; TL approved the plan and change-code passed. One implementation close boundary; no Mx tags.

## Log


- 2026-10-22: closed — Recorded sdlc-local/suite PASS on b061d0dd1f2aa6382b3b3c3d34f5fa7256439dba: 8 shards, 1960 SDLC tests and 75 other packages, zero failures; log sha256 prefix 982f0f76a78c. BR-7 shared literal owner and retired old scanners have four-context sidedoc regressions, pure/coordinator preservation and ordinary-link controls, mask-removal mutation failure, and restored package/race checks. Only Log bookkeeping changed since suite. --no-plan-check acknowledges remaining combined Plan row includes post-SHIP hub-owned staging/live check and merge, still pending.; review verdict: SHIP
### 2026-10-20

### 2026-10-21

- Claimed and started from main after #514's completed merge/cleanup. Existing reusable seams: SessionSync Peek/Apply and whole-file CAS, sidedoc link/index rebasing, and sdlc trunk publication. No Go marker-cleaning implementation exists; the source of marker semantics is construct/local/fix/review-convention.md. Bounded read-only exploration confirmed the file-aware fake gap and the need for a Git blob baseline rather than mtime.
- TL clarified publication and cleaning (receipt 5ea7ef7f-3e12-4473-8369-91aa89a8425c): product repos use trunk publish; ops operates directly on primary main, inferred from the repository. Unresolved commentary/proposals and malformed markers refuse with diagnostics; settling is not implicit acceptance. Literal code and anchors survive. These decisions are recorded in the durable plan, alongside explicit partial-publication outcomes and retries.
- Plan proposes a pinned per-path baseline and JSON receipt extension to the existing trunk publisher because its current merge-base-only comparison refuses repeat settlement from a stale hub checkout. This design delta remains for TL review; no implementation started. ARCH-DRY: reuse one publisher and one projection; ARCH-STATE: commit/push and Couch stamping have explicit separate confirmation boundaries.
- TL approved the plan and publisher baseline/receipt extension (7f2905de-b94c-4cc8-bb53-582cc8aeb5b7). Pinned full flow explicitly because the pre-code diff alone inferred quick flow. Plan-quality round 1 found PQ-1: mutable destination reread could publish unvalidated bytes, and PQ-2: testing needs named functions/strategies. TL approved the immutable payload guard/private ops Git index refinement (18ce25ca-d968-4481-8f80-879f39ab36ac). Added a plan Revisions section recording exact payload authority and current function test contracts; default trunk behavior must stay unchanged. No implementation started before gate clearance.
- Full-flow change-code passed with clean plan-quality and an informational estimate-quality verdict. The estimate review correctly noted optimistic aggregation of the ops publisher/pure core and implicit verification allocation; retain the 2.93h derived point as provisional calibration, and report actuals without back-fitting it. Task decomposition delegates only isolated publisher/marker/ops files; the main worker owns the shared model and coordinator.
- Pure slice verification: metadata/pointer/transition/classification tests began red for absent APIs and pass after implementation. Datatype discovery began red with workingdoc missing, then passes with the prototype and existing collaboration schema checks. Marker scanner behavioral red was observed against its stub; grammar/literal/location/linear-count fixtures pass, and live fuzz passed 45,182 executions. Combined focused pure/marker tests PASS (0.087s); datatype checks PASS (0.132s). No command/coordinator or live-document effect is implemented yet.
- Shared /Volumes/AgentScratch/go-build filled, causing setup failures before Go compilation. Focused checks use owned caches under TestScratch/tmp (parent: ariadne-slot4/ariadne-507-cache.DnAOfs); no shared cache sweep, queue change or unrelated process cancellation. A selected-index synchronization clarification is pending with TL: private-index commit alone would leave a staged inverse entry in the real index; preserve unrelated entries/live edits while reconciling only the selected path, or report explicit reconciliation if the index changed.


- Implemented both commands, bounded repository/Couch seams, source-only CAS stamping, batched status baselines, and publication confirmation/retry behavior. Product publication adds opt-in per-path baseline, immutable SHA256 payload guard and JSON receipt to the existing trunk publisher; default behavior retains regression coverage. Ops publication uses a private index and preserves concurrent target edits. TL approved selected-path real-index reconciliation (cf9b4eed-e7c1-4f55-9c83-4cdfd56f807d): compare the captured index under index.lock, preserve unrelated entries, and return an exact reconciliation instruction on independent changes. The durable plan records this revision.
- Regression discoveries: unsupported parent forms now refuse through the existing side-doc validator before projection; reserve pointer space before publication; reject annotated-tag baselines and recheck ancestry during each publisher CAS attempt. Git timeout tests exposed surviving hook descendants; bounded runners now own and retire their process groups, with readiness-based cancellation regressions.
- Focused verification PASS: complete cmd/tl-lead package 36.649s, cmd/datatype 0.072s, pkg/golint 0.515s; existing/new trunk publication tests 30.401s; ops publication tests 17.024s; repository adapter tests 6.089s. Pointer fuzz passed 63,618 executions; marker fuzz passed 45,182. Stateful real bare-remote fixtures cover immutable payloads, uncertain push/retry, source CAS failures, selected-index races, baseline drift and cancellation. No live hub documents were changed.
- TL explicitly requested a shared-cache hold (52d78dc9-dfa4-4e39-93c1-ff639b35bcc8). Focused checks are finished; no recorded full suite has been queued, and no GOCACHE override will be used for it. Full-suite verification, close, hub-owned live check and landing remain pending until the operator restores the shared cache and TL releases the hold.

- Shared-cache hold released by TL (3cbc4b1e-503b-4612-8b03-2cf84b06058d). Verified go env GOCACHE resolves to /Users/xianxu/Library/Caches/go-build. Earlier ENOSPC setup failures were infrastructure failures before compilation, not test evidence; no recorded #507 suite had run. Starting the recorded suite on the pushed implementation with no per-run cache override.

- Recorded suite PASS on c780bb6e83ae1f48b39dd26f382967daf222b8de: all 8 shards, 1,959 cmd/sdlc tests and 75 other packages, zero failures; wall 857.7s. Recorded as sdlc-local/suite; log SHA256 prefix 20c7a871af69. The retained run survived the Couch restart. Proceeding to close review; the last Plan item deliberately remains unchecked because it includes the hub-owned post-SHIP live check and merge.

- Close round 1 returned REWORK: BR-1 pointer YAML roundtrip, BR-2 first-settlement remote ownership, BR-3 interruption coverage and BR-4 host binary dependency. Verified each and added regressions. BR-1 behavioral RED covered 16 adversarial paths; generation now reparses with the production reader before effects. Path fuzz PASS 30,289 executions. BR-2 reproduced an actual overwrite of an unrelated committed target; TL approved --expect-absent (d89aae91-3650-4e01-aed3-92152cef7c2c), supplied only for first settlement and checked inside every CAS callback. Default trunk behavior remains unchanged. BR-3 added lost-ack/pre-write failure/late-lock/unavailable-readback/source-edit retry coverage and corrected the fake's retained commit identity. BR-4 reproduced failure without host sdlc and now fixtures the authoritative agent-id command; restricted-PATH ops tests PASS 41.161s.
- Post-fix affected checks PASS: existing/new trunk publication tests 65.113s; full tl-lead 84.140s, datatype 0.442s, golint 1.041s. Added adversarial CAS retry-base case PASS in first-publication family 7.530s and full-coordinator unsafe-pointer no-effects test PASS 0.164s. Code changed after the earlier recorded suite, so a new recorded suite is required before re-close/staging.

- Post-fix recorded suite PASS on 14b431532272b21654a59108cd6823ffda03e9d8: all 8 shards, 1,960 SDLC tests and 75 other packages, zero failures, wall 694.4s, log SHA256 prefix 6fe7463dffc6. Retained lease-2822675348 survived the restart and completed after its original FIFO wait. No flake resume was needed. Re-running full close review against the four fixes; staging and live conformance remain post-SHIP.

- Close round 2 disposed BR-1 through BR-4 and returned REWORK for BR-5 (cross-block backticks hiding unresolved commentary) and BR-6 (model observed effects without authorizing them). TL approved bounded corrections (af7edb73-8f1f-4e91-ac8e-fbce7dc9c910). Actual Goldmark code nodes now determine literal masks; scanner/coordinator regressions cover paragraph, heading, list and blockquote boundaries plus real code controls. Effect permissions returned by the pure model now gate publication and stamping before calls; production consumes failure and unconfirmed outcomes. Removing the effect guard failed both forbidden-sequence regressions (mutation exit 1).
- TL accepted Goldmark's bounded-input complexity limit (33dac06b-99a1-4982-90b1-425fec0e8fee): counted linear work covers our input/AST/masks/scanner, not dependency parser internals. On a temporary 1 MiB adversarial same-paragraph fixture with unmatched backtick widths 1..1,446 and an active question, scanning took 383.972ms and retained the question; own work count 2,098,616. No timing threshold remains in tests. Focused marker race PASS 1.528s, fuzz PASS 15,130 executions, final marker/control tests PASS 0.198s. A third-round mechanism-level finding must go to TL before further fixes.

- Combined post-BR-5/6 checks PASS: full cmd/tl-lead 59.286s, cmd/datatype 0.184s, pkg/golint 1.140s; diff check clean. Committing this coherent correction before queuing a fresh recorded suite. No staging or live hub effect has occurred.

- Post-BR-5/6 recorded suite PASS on 4886a014ea52f498d3475603ebc1e59fff7ca2a2: 8 shards, 1,960 SDLC tests and 75 other packages, zero failures, wall 908.7s; recorded sdlc-local/suite, log SHA256 prefix 4fb0c18091a0. Retained lease-2210632118 completed after its FIFO wait. Re-closing now per TL hourly sweep; any new mechanism-level review gap will be reported before further implementation.

- Close round 3 disposed BR-5/6 and returned REWORK with new Critical BR-7 (markdown-literal-preservation): the shared sidedoc.RebaseLinks still uses older line-based literal detection, so multiline spans, indented code and list/blockquote fences can have example links rewritten before successful publication/stamping. Reviewer reproduced four pure/coordinator probes. Per TL instruction, stopped before implementing another mechanism-level correction and reported the concrete shared-boundary proposal to hub. No stage or live effect occurred; awaiting TL decision.

- TL approved BR-7 as completing the existing shared mechanism (ec07e85b-8459-4f68-9338-9b232420fd29). Promoted Goldmark masking to sidedoc.CodeLiteralMask, used by marker scanning, RebaseLinks and SideLinks; deleted old codeSpans and forUnfenced implementations. Link/index public APIs and approved cost scope stay unchanged. Index rebasing is structured frontmatter-only and does not rewrite body examples. Shared tests first failed all four reviewer contexts; the nested-list case also showed an ordinary live link incorrectly skipped by the old fence scanner.
- Pure projection plus full settlement/retry tests cover the four cases and ordinary-fence/inline controls, exact published bytes, source preservation and live-link rebasing. Removing the shared mask from rebasing failed every pure/coordinator preservation case (mutation exit 1); restored focused controls PASS 0.421s. Shared sidedoc tests PASS 0.155s, focused race PASS 1.356s, reviewexport consumer PASS 0.325s. Any round-four finding outside literal detection will be reported to TL before implementation.

- Combined restored BR-7 checks PASS: cmd/tl-lead 76.228s, cmd/datatype 0.180s, pkg/sidedoc 0.095s, pkg/golint 1.258s. Diff check clean. Checkpointing before a new recorded suite, since shared production code changed after the previous pass.

- 2026-10-22 sweep: retained BR-7 suite lease-2547092102 completed PASS on b061d0dd1f2aa6382b3b3c3d34f5fa7256439dba: all 8 shards, 1,960 SDLC tests and 75 other packages, zero failures; wall 737.1s, log SHA256 prefix 982f0f76a78c, recorded sdlc-local/suite. Re-closing now; no duplicate suite or cache override. Round-four gaps outside literal detection remain a stop-and-report boundary.

- Round-four close returned SHIP: all seven findings disposed, no new findings; measured actual adopted 2.87h. Final close head 25c656f4eec004ea6ecd1bca71536c10a9115946 staged on ariadne:0 through sdlc stage --build. All seven built binaries, including tl-lead and sdlc, verified vcs.revision=25c656f4eec004ea6ecd1bca71536c10a9115946 and vcs.modified=false. Stage report sent to hub; holding for TL/operator disposable live check and ordered fixture-replay evidence before clearing/merge. No worker writes to live hub documents.

## Revisions

### 2026-10-21 — publication ownership and unresolved review items

The initial Spec's shorthand "review markers stripped" does not authorize deleting an unanswered operator item or accepting a pending proposal. Only already-accepted clean text can settle; remaining commentary, proposals and malformed markers refuse. Product copies publish through the trunk path; ops copies commit/push directly on primary main under its operating charter. The original working page remains live and receives its confirmed path/commit pointer through Couch Apply. Re-settling updates the same tracked path and Git retains its history. Done-when includes fixtures for these decisions, interrupted publication/stamping and content-based drift; a hub-owned live check precedes landing.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Calibration is tagged stale by estimate-source, so these numbers are provisional.

Primitive decomposition: issue/spec shaping (design 1.0 x0.2, implementation 0.2 x0.4); one typed prototype (1.0 x0.2, 0.35 x0.4); one greenfield pure document model/marker concern (1.0 x0.2, 0.6 x0.4); existing publisher extension and status surface as two smaller modules (each 0.2 x0.2, 0.4 x0.4); one stateful publication/Couch integration (2.0 x0.2, 1.25 x0.4); atlas/help (0.1 x0.2, 0.15 x0.4); one boundary review (0, 0.4 x0.4); external conformance discovery (0, 0.4 x0.4). The thorough approved plan supplies the x0.2 design discount; familiar Go/Cobra/Git/Couch seams use familiarity1.0. Existing yaml.v3, sidedoc and trunk transaction helpers cover the library-availability check; the custom marker scanner has no existing Go implementation, and is not credited with a nonexistent parser shortcut. No new stack or vendor propagation. Design subtotal1.10, plus15% buffer1.265; implementation subtotal1.66; total2.925 rounded2.93 hours.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.20 impl=0.08
item: typed-data-prototype design=0.20 impl=0.14
item: greenfield-go-module design=0.20 impl=0.24
item: smaller-go-module design=0.04 impl=0.16
item: api-integration design=0.40 impl=0.50
item: smaller-go-module design=0.04 impl=0.16
item: atlas-docs design=0.02 impl=0.06
item: milestone-review design=0.00 impl=0.16
item: real-api-discovery design=0.00 impl=0.16
design-buffer: 0.15
total: 2.93
```
