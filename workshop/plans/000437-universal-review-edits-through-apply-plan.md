---
collaboration-mode: discussion
---

# Universal review one-edit-path implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every review of an existing tracked document uses atomic, reasoned `apply`; the hosted Pair workbench uses its matching pair-level verb. A document's `presentation:` alone chooses proposals.

**Architecture:** The skill's existing round and collaboration decisions remain pure. Its Couch adapter builds edit records against a fresh whole-file snapshot and lets Pair's engine own CAS, operator locking, proposal conversion, and round derivation. The Pair workbench follows the same record contract through `pair review apply`, then commits the returned body. The export manifest becomes v3 without the mode-level deletions field after parley#383 can consume both v2 and v3.

**Tech Stack:** Markdown skills, Go `pkg/reviewexport` parser/manifest and tests, Pair/Couch CLIs for live conformance.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| `ReviewMode` (`scope`, `frontier`, brief) | `construct/local/universal-review/modes/*.md`, `round.md`; `pkg/reviewexport/export.go` `Mode` | modified: remove `deletions` |
| `Presentation` (`direct`, `highlight`, `propose`) | reviewed document's frontmatter, explained in `round.md` | modified: one document-level proposal choice, missing means highlight |
| `EditRecord` (`old`, `occurrence`, `new`, `explain`) | `adapters/file-edit.md` and `construct/local/fix/SKILL.md`; implemented by Pair | modified contract consumer, not a new resolver |
| `RoundSpan` (`old`, `new`, `explain`, result position) | Pair's `round --json`; summarized by `round.md` reply rule | modified contract consumer |

`ReviewMode` still decides *where/how much* to edit. `Presentation` decides *how an applied edit is shown*. This removes two competing proposal rules (ARCH-DRY, ARCH-ORDER). `EditRecord` is the existing Pair engine's input, not a parser to replicate in ariadne. A blank/no-op answer may still use the current `write --all` path solely to fold an owed submit; it does not change document content. Pair's converter deliberately refuses frontmatter proposals: a necessary frontmatter-only change uses `write --frontmatter`, whose round is still derived by the same engine. Body edits use `apply`.

### Integration points

| Seam | Lives in | Status | Responsibility |
|---|---|---|---|
| Tracked-document adapter | `construct/local/universal-review/adapters/file-edit.md` | modified | status/diff/read snapshot; prepare records; Couch apply with fold IDs; refusal recovery; round read and reply |
| Side-doc fold and parent | `construct/local/universal-review/collaboration-modes.md` | modified | preserve write ordering; existing parents use their own apply/hash; new side docs are created through the existing new-file path |
| Hosted Pair workbench | `construct/local/fix/SKILL.md` | modified | Pair apply, exact returned Body to docflow, no handoff/landed/reconcile |
| Export | `pkg/reviewexport/export.go`, `construct/local/universal-review/export.json` | modified | v3 members omit deletions; no presentation copy in a mode or prompt flag |
| Consumer | parley#383 | external, lands first | accept v2/v3, reject v3 mode deletions, avoid missing-flag `propose-strike` default |

The adapters never implement matching, locking, hashing, marker conversion, or round positions themselves (ARCH-PURE, ARCH-DRY). They describe how to call Pair's one owner. No new durable file family is added here; the existing Pair round artifact owns retention (ARCH-FUNERAL). `export.json` keeps its existing lifecycle through regeneration.

## Actions and recovery

Actions: `read`, `rebuild records`, `apply`, `wait for next submit`, `write a side-doc proposal`, `fold in`, `commit returned body`.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Fresh snapshot and valid records; apply exit 0 | applied round | read `round --json` for the reply; in Pair workbench commit returned Body immediately | none |
| Apply exit 1 `edits-invalid`, including bad-record indices | refused, unchanged | fix the listed records and apply the whole batch again | no partial edits |
| Apply exit 1 `propose-too-large` or `propose-unencodable` | proposal needs side doc | write the rewrite in a side doc, then apply a `→ [proposal](…)` link; fold on later `🤖[fold in]` | if parent/link write refuses, the side doc remains unlinked and is relinked on retry |
| Apply exit 3 `operator-lock` | waiting for operator | wait for next submit, then read again | none; this turn is deferred |
| Apply exit 6 `stale` | needs rebase | read again and rebuild records from the new bytes | no partial edits; previous computation discarded |
| Successful apply but round read fails | document changed, reply incomplete | read round again; name uncertainty if still unavailable | no document loss, but explanation may be unavailable |
| Pair apply exit 0, docflow commit interrupted | applied, uncommitted | re-read Pair's recorded round/body and commit only if Git and document still match; otherwise stop for operator | uncommitted round history until recovery |

The skill never guesses success from an exit status it did not receive. For a lost response, read current file and round before retrying; a blind replay could double-apply. Pair's operator lock and hash are the single resources for edit ownership; no activation envelope or separate reconciliation state remains (ARCH-ORDER). The review's side-doc ordering and `parent:`/`marker:` relink logic remain as they are.

## Tasks

### Task 1: contract tests before prose

**Files:** `pkg/reviewexport/export_test.go`, `pkg/reviewexport/examples_test.go`; add a focused skill-contract test in `pkg/reviewexport` if the existing checker cannot express the cases.

- [x] Write red tests: v3 mode files without `deletions` parse and a mode still carrying that obsolete flag is rejected; v3 manifest mode records omit the field; `round.md` names document `presentation:` and no active `propose-strike` directive; adapter and workbench name apply, exit 1/3/6, and their distinct round/body outputs. Existing example checker should reject a case that still asserts the old proposal rule.
- [x] Run `go test ./pkg/reviewexport -count=1` and record the expected failures.

### Task 2: authored universal-review skill and examples

**Files:** `construct/local/universal-review/{round.md,collaboration-modes.md,adapters/file-edit.md,modes/*.md,examples/discussion-parent/{case.md,after/design.md},SKILL.md}` and any other example case whose expected output depended on `propose-strike`.

- [x] Remove the mode deletion flag and its directives. In `round.md`, read `presentation:` from each document's frontmatter every round; the mode still governs scope/frontier. Proposal markers come only from Pair's converter, never hand-written by the agent. Under `propose`, keep small changes as records and route `propose-too-large`/`propose-unencodable` to a linked side doc.
- [x] Replace the tracked adapter's body-file `write --all` step with EDITS.json records and `couch --session-sync apply EDITS.json FILE --if-hash H --fold ID[,ID] --json`. Every record has an operator-readable `explain`; `occurrence` is 1-based and `old` nonempty. On exit 1 invalid, fix the list and retry the entire batch; on exit 6 re-read/rebuild; on exit 3 defer. A successful round is read with `couch --session-sync round FILE --json`, and the reply lists its spans. Keep `write` only for generated content, no-op fold, and the converter's frontmatter-only exception.
- [x] Preserve side-doc ordering and move/fold semantics; rewrite examples and checker expectations to the new presentation rule, including a proposal-too-large side doc linked then folded. Run `go test ./pkg/reviewexport -count=1` until green. The live fixture must not require the agent to type marker syntax for a proposal.

### Task 3: hosted Pair workbench protocol

**Files:** `construct/local/fix/SKILL.md`, `atlas/workflow/universal-review.md`; add a focused phrase/contract test if none guards this section.

- [x] Replace handoff-writing with `pair review apply EDITS.json FILE --if-hash H --json` and its 0/1/3/6 table. Delete the activation-context envelope, landed artifact, applied-poke and reconciliation instructions. Commit `Body` from a successful reply verbatim via `docflow round --side agent -m SUMMARY --body BODY` immediately; human-round and readiness rules stay.
- [x] Remove the Edit/Generate instructions to hand-write proposal markers; preserve their distinct scope/voice behavior and direct mechanical proofreading. Update the skill's description/pointers so agents find the current path. Add an explicit lost-response read-before-retry rule.
- [x] Test the workbench contract text and its Pair CLI example; run the focused checker.

### Task 4: export, conformance, and landing

**Files:** `pkg/reviewexport/export.go`, `export_test.go`, `construct/local/universal-review/export.json`, `atlas/workflow/universal-review.md`, issue Log.

- [x] After parley#383's dual-version consumer lands, bump manifest `Version` from 2 to 3, omit `Deletions` from `Mode` and `Member`, regenerate `export.json` with `UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport`, and verify its exact member hashes. No `presentation` flag is copied into mode records or Parley's instruction context: the document frontmatter is the single source.
- [x] Build with `make -k build` (and `make -k sdlc-build` if needed), run focused Go tests and the examples checker. With Pair M2 available, run a live agent round on a real tracked scratch document through `couch --session-sync apply`, show its round's explained spans and an invalid/stale refusal with no partial edit. In a Pair workbench scratch round, use `pair review apply` and verify the returned Body is what docflow commits. Do not modify pair:2's checkout.
- [ ] Push the branch, run `sdlc local-check run --name suite` on that exact head, and collect the live conformance and compatibility evidence for close. If the Pair M2 engine or parley dual-version consumer has not landed, wait rather than publishing a skill whose verb/export is unavailable.

After the evidence is complete, close with the mandatory review, land via `sdlc merge --yes`, and report the PR/merge hash to the hub. The TL owns the project tick.

### Task 5: close-review convergence

- [x] Add red examples and contract checks for untracked `propose` refusal, marker-free proposed links and index cleanup after rejection, and lost-response recovery with an older current round; keep the Pair engine's marker-present fixtures intact.
- [x] Update both review adapters, round and collaboration guidance, and the README to the approved recovery and presentation rules. Prove the new cases and the full examples checker pass.
- [ ] Re-run the recorded suite if code changes, then close review and merge without waiving the fresh-context verdict.

## Revisions

- 2026-10-18, pending-link conformance: a live marker-free Pair apply wrote the proposed link as `🤖{ → \\[proposal\\](…)}` while its `discussions:` entry applied directly. The shared `pkg/sidedoc` parser previously missed that link, which would remove a still-pending index entry. Recognize only the engine's measured escaped marker shape, retain its index and side doc until the operator accepts or rejects, then drop the index after rejection. Add examples for both transitions. Parley's mover separately misses the escaped pending link; file parley#389 without holding #437.

- 2026-10-18, close review round 1 (BR-1 through BR-4): the document-level `presentation: propose` promise cannot silently become a direct write on an untracked file. Refuse that combination with the exact Couch-tracking or hosted Pair workbench next action; direct/highlight untracked edits retain the local file path and report their observed before/after diff rather than claiming a Couch round. A marker-free large rewrite or chat answer may create its side doc and propose its link: put the matching `discussions:` entry in the same apply, regard the proposal as pending, and remove the entry in the next round if the operator rejects the link. A lost Pair apply reply does not authorize committing the last current round: compare its `base_hash` with the attempted H, require `Current`, then read the file and check that the spans, explanations and result match the attempted edit; stop without commit or retry if any part is uncertain. Add a compact README pointer for the new presentation and apply surface. Red fixtures cover untracked propose refusal, marker-free proposal and rejection, and an old-current versus newly-landed lost response before skill edits.

- 2026-10-18, initial plan: incorporate Pair M2's exact review-workbench delta and the hub's parley-first export order. The `presentation:` source remains document frontmatter, not a copied round flag. A tracked Couch document uses `write --frontmatter` for a frontmatter-only change under `propose`; body edits use apply.
- 2026-10-18, proposal-link conformance: Pair M1's converter proposed even the link replacing an operator request, leaving no active side-doc link; a later large fold would hit the same size refusal. The hub assigned Pair #661 M2 a base-text-checked exception: under `presentation: propose`, a record that answers the operator's own ready request marker (including `🤖[fold in]` beside a side-doc link) applies directly. No agent-declared exception exists; unrelated edits still become proposals. Update the skill and fixture to that landed contract before close.
- 2026-10-18, atomic side-doc index: a new link also adds the matching `discussions:` frontmatter entry. Pair M1 rejects body plus frontmatter in a proposed write; the hub chose a Pair-level extension to the same operator-marker exception so the matching link and index entry ride one `apply` atomically. Keep the skill's one-write rule and verify it against Pair #661 M2 rather than adding a two-write recovery state.
- 2026-10-18, hosted frontmatter path: Pair's review file is not Couch-registered, so the earlier `write --frontmatter` instruction in xx-fix named no runnable command there. Pair #661 M2 is deciding its direct frontmatter rule; keep the workbench on `pair review apply` with frontmatter-only records, then align its exact refusal semantics with the landed engine. The tracked Couch adapter's dedicated `write --frontmatter` exception is separate.
- 2026-10-18, narrow structural exception: Pair's first M2 rule made *any* record containing a ready marker direct, which would turn an ordinary `🤖[tighten this]` edit into a silent prose change under `propose`. The hub restricted direct apply to a record that replaces the operator request with a side-doc link or answers `🤖[fold in]` at that link; its matching index entry may ride with it. Pair adds a red regression before narrowing the engine. Update the skill to this precise rule and keep ordinary prose proposals.
- 2026-10-18, hosted frontmatter contract: Pair #661 M2 confirms that frontmatter-only records go through `pair review apply` directly under `propose`, while a record crossing the closing delimiter refuses. `presentation:` and `collaboration-mode:` are operator-owned and refuse with exit-1 `operator-owned` on every write path; the skill must stop and ask the operator to set them.
- 2026-10-18, Pair conformance and recovery: Pair M2 now keeps the same link and appends `(folded <date>)` when a ready fold-in is applied directly; its conformance fixtures copy #437's large-proposal and fold documents byte for byte. Add its ordinary `🤖[tighten this]` fixture here to defend the narrow exception in both repos. Pair applies wholly-frontmatter index records directly but does not verify their path matches the link; the skill must match them and keep one atomic batch. For an uncertain Pair apply, `pair review round FILE --json` returns the stored `Body` and `Current`; commit only when Current and the review branch/file still match, re-read/rebuild when stale, and retry only when Round is null.
- 2026-10-18, close checklist: change the final task from a self-referential close-and-land checkbox to its pre-close evidence. The same close, merge, report and dependency order remain required immediately afterward.
