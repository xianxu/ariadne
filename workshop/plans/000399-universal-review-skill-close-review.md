# Boundary Review — ariadne#399 (whole-issue close)

| field | value |
|-------|-------|
| issue | 399 — Universal review skill and canonical round export |
| repo | ariadne |
| issue file | workshop/issues/000399-universal-review-skill.md |
| boundary | whole-issue close |
| milestone | — |
| window | f3bc999f925ab5137d1de82c6b12c3511d746f83..087c8af70c4b394008aadbdbeaf7aba5e79b1169 |
| command | sdlc close --issue 399 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T17:03:29-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The work does what #399's Spec and Plan promise.

- **Skill layout:** `construct/local/universal-review/` matches the Plan.
- **Modes:** the six modes are byte-identical to parley's `lua/parley/skills/review/modes/`. I checked this with `cmp` against the local parley.nvim checkout.
- **Grammar:** the grammar is a symlink to the one authored `fix/review-convention.md`, not a second copy.
- **Manifest:** `export.json` is built by a pure function over an `fs.FS`, and `TestExportCurrent` fails when it is stale.
- **Tests:** `pkg/reviewexport`, `cmd/tl-lead` and `cmd/tl-worker` all pass. `TestUniversalReviewSkillDiscoveredInDerivedRepo` fails under this sandbox's default TMPDIR (`.sdlc-review/tmp`) because `git rev-parse` can't find a repo there. Weave tests that were already in the repo fail the same way. With `TMPDIR=/tmp` it passes.

The one Done-when item not delivered is the live operator Alt+Return check. A TL decision, recorded in the Log (delivery f73c1f3a), moved it to parley.nvim#332's acceptance. That decision stands and doesn't block. Nothing blocks SHIP; the findings below are all Minor.

1. **Strengths**
   - `pkg/reviewexport/export.go:670-733`: `Build` is pure over `fs.FS` and sorts members in a total order (role → order with unordered modes last → name → path). It reads through the grammar symlink, so sizes and hashes cover the resolved contents. `TestBuildOrder` checks ties and unordered modes on a fixture.
   - `adapters/file-edit.md:270-283`: the tracked/untracked decision fails closed. A `type: session-sync` file is checked first, and an unknown lookup result writes nothing. Unknown ownership never counts as permission to edit directly.
   - `adapters/file-edit.md:287-309`: the order status → diff → `read --all` → `write --if-hash --fold` is spelled out, and it explains why the diff must come before the read. Deliveries that arrive late, and ones this write didn't answer, stay owed. The steps after exit 3 (lock) and exit 6 (stale) are both stated.
   - `round.md` names no Couch verb, so parley's standalone composition (round + grammar + modes + its own adapter) works without changes.
   - The discovery test builds a real derived repo, runs weave on it, and checks that the grammar bytes equal the canonical file in both `.claude/skills` and `.agents/skills`.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `round.md:32-33` hard-codes the list of known modes. That repeats `modes/` and no test guards it (ARCH-DRY / ARCH-PURPOSE). Either tell the agent to list `modes/`, or have `TestExportCurrent` assert that round.md's list equals the mode members.
   - `adapters/file-edit.md:302-303`: each tracked round creates a body file with `mktemp` and nothing removes it (ARCH-FUNERAL). Add "remove BODYFILE after the write".
   - `ParseMode` (`export.go:619-663`) doesn't handle CRLF line endings or quoted values. An empty frontmatter block (`---\n---`) is reported as "unterminated". This is fine for the authored modes, but the parity claim with parley's parser is only tested on happy-path inputs.

5. **Test coverage**
   - Manifest staleness, determinism, refusing missing members, and the six stock modes in order are all pinned.
   - The TL help phrase guards are pinned in both `main_test.go` and `rules.tsv`.
   - The adapter's Couch protocol is prose, and only the dry run plus the deferred live check exercise it. No fake `couch` exists (ARCH-MOCK). Since the "code" here is agent instructions, that's acceptable for now, but parley#332's acceptance is the only conformance evidence.

6. **Architecture**
   - **ARCH-DRY:** flag (Minor, the round.md mode list). Otherwise pass: the grammar is linked, not copied.
   - **ARCH-PURE:** pass. `Build` is pure; IO lives only in `TestExportCurrent`.
   - **ARCH-PURPOSE:** pass. The skill, the manifest and the TL pointers all derive from the one source. The live check was deferred by the TL, not by the implementer.
   - **ARCH-MOCK:** note. No `couch` fake for the adapter's protocol; conformance rests on the deferred live check.
   - **ARCH-CONSTRAINTS:** pass. It's a small manifest; the adapter bounds lock waits with a 30-minute ping and no busy-wait.
   - **ARCH-SECURE:** pass. The lookup fails closed and the CAS is never bypassed.
   - **ARCH-ORDER:** pass for prose. The step order and the late-delivery handling are explicit.
   - **ARCH-FUNERAL:** flag (Minor, the temp body file).
   - For upcoming work: when the section modes (round.md §4) arrive, extend the manifest schema with a `Version` bump rather than adding fields silently.

7. **Plan revisions**
   - Add a `## Revisions` entry saying the Done-when "live operator check" and its unchecked task moved to parley.nvim#332's combined acceptance (TL decision f73c1f3a). Then the Plan no longer claims this issue delivers it.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: hand-maintained-restatement-of-source
    title: |
      round.md hard-codes the known-mode list, which repeats modes/ with no test guarding it
    detail: |
      round.md:32-33 lists developmental, line-editing, copy-editing, proofreading, fact-check and free-form by hand (ARCH-DRY). Either tell the agent to list modes/, or assert in TestExportCurrent that the list equals the manifest's mode members.
  - id: new
    severity: Minor
    kind: defect
    family: artifact-without-removal-path
    title: |
      The file-edit adapter creates a mktemp body file on every tracked round and never removes it
    detail: |
      adapters/file-edit.md step 4 creates review.XXXXXX under TMPDIR for each round, and nothing removes it (ARCH-FUNERAL). Add an instruction to delete BODYFILE after the write, whether it succeeds or is refused.
  - id: new
    severity: Minor
    kind: defect
    family: parser-parity-untested-edges
    title: |
      ParseMode's parity with parley's parser is only tested on happy-path frontmatter
    detail: |
      pkg/reviewexport/export.go ParseMode doesn't handle CRLF line endings or quoted values, and it reports an empty frontmatter block as unterminated. Pin these against parley's behavior, or document them as out of scope.
  - id: new
    severity: Minor
    kind: defect
    family: plan-claims-undelivered-item
    title: |
      The Done-when live operator check is still claimed, though the TL moved it to parley.nvim#332
    detail: |
      Add a Revisions entry recording TL decision f73c1f3a, so the Plan and Done-when no longer claim this issue delivers the live Alt+Return check.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"4e42f121-e49f-47ce-bd4f-3f8e940ddb1a","repository":"github.com/xianxu/ariadne","issue_id":"000399","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/81476eaa8f017e33","reviewed_head":"087c8af70c4b394008aadbdbeaf7aba5e79b1169","slot":"ariadne:2","requested_at":"2026-10-15T23:57:15.441874Z"},"phase":"finished","pid":70098,"started_at":"2026-10-15T23:57:15.62149Z","finished_at":"2026-10-16T00:03:08.491485Z","session_id":"313ec06c-a55e-4137-9337-cbed3fffea00","outcome":"returned"}] -->

---

## Re-review — 2026-10-15T17:06:17-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 399 — Universal review skill and canonical round export |
| repo | ariadne |
| issue file | workshop/issues/000399-universal-review-skill.md |
| boundary | whole-issue close |
| milestone | — |
| window | 08a09e4755aed877713fd0f4aa17c8d473f726fc..e40d16b5fc230bf36f13c69415f21905f96d0698 |
| command | sdlc close --issue 399 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T17:06:17-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

All four prior Minors are fixed in this window, and `go test ./pkg/reviewexport/` passes. The problem is the rest of the window. It also delivers a new requirement that no prior finding asked for: operator requirement 5ef50ef6, "changes since your last read exclude the agent's own edits". That requirement adds a second `## Revisions` entry, rewrites the `## Untracked` section of `adapters/file-edit.md`, and changes the contract in `round.md` §1. It is new behavior added after close, so this narrow fix review doesn't cover it. Under the fix-review contract that is REWORK: run a full `sdlc close`. The new behavior also brings back a temp file that is never deleted, which repeats the family from BR-2.

1. **Strengths**
   - `pkg/reviewexport/export.go:86`: the `"\n"+rest` search is a minimal fix that makes an empty frontmatter block parse. The test `empty frontmatter` (`export_test.go:32`) expects a `basename` error; before this fix the code returned "unterminated", so the test would fail without the fix.
   - The CRLF and quoted-value edges are now documented in the ParseMode doc comment and pinned by tests, so their scope is explicit.
   - `export.json` hashes were regenerated for both edited members, and TestExportCurrent passes.

2. **Critical:** none.

3. **Important**
   - **The scope changed after close.** The requirement 5ef50ef6 (Revisions entry 2) came in after close. It changed `construct/local/universal-review/adapters/file-edit.md:64-76` (a saved base copy and `diff -u COPY FILE`) and the "changes since your last read" contract in `round.md:18-21`. None of the four prior findings asked for this. It needs a full `sdlc close` review.

4. **Minor**
   - **The new `review-base.XXXXXX` copy is never deleted** (`file-edit.md:70-72`). Each copy is deleted only when the next one replaces it, so the last copy of every conversation stays in TMPDIR. **This is the 2nd finding in family `artifact-without-removal-path`.** The rule that covers both: every temp file an adapter creates must state when it is deleted, including the end of the conversation or session. Apply that rule to the whole adapter, not just this one site.

5. **Test coverage**
   - The ParseMode edges are pinned.
   - The CRLF and quoted-value tests already passed before this change. They pin behavior, which is right for a parity claim. Parity with parley itself can't be checked from this repo.
   - Nothing tests that `round.md` refers the agent to `modes/`. That's acceptable, because the list itself is gone and the manifest already hashes `modes/`.

6. **Architecture**
   - **ARCH-DRY: pass.** BR-1 removed the hard-coded mode list.
   - **ARCH-PURE: pass.** The ParseMode change stays pure.
   - **ARCH-PURPOSE: flag.** This is via the scope finding: the window adds behavior outside the fix set, and its temp-file handling repeats the BR-2 class instead of applying a rule to the whole adapter.

7. **Plan revisions**
   - Revisions entry 2 should stay, but it has to go through a full close review rather than this narrow one.
   - Entry 1 (TL decision f73c1f3a) correctly records that the live check has moved.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      round.md:33 now says "list the known modes (the files in `modes/`, by `order`)"; the hand-kept list is removed.
  - id: BR-2
    disposition: addressed
    note: |
      file-edit.md step 4 now says "Then delete BODYFILE, whether the write succeeded or was refused."
  - id: BR-3
    disposition: addressed
    note: |
      export.go:86 handles the empty block. Tests crlf, quoted-value-literal and empty-frontmatter pin the edges, and the doc comment records the CRLF and literal-value behavior. The empty-frontmatter test fails without the fix.
  - id: BR-4
    disposition: addressed
    note: |
      A Revisions entry records TL decision f73c1f3a, moving the live Alt+Return check to parley.nvim#332's combined acceptance.
findings:
  - id: new
    severity: Important
    kind: defect
    family: scope-change-after-close
    title: |
      Scope changed after close: run a full `sdlc close`
    detail: |
      The interdiff adds operator requirement 5ef50ef6, which no prior finding asked for. It adds a second Revisions entry, the base-copy diff workflow in adapters/file-edit.md (the Untracked section), and a new "never your own edits" contract in round.md section 1. This is new behavior after close, outside the narrow fix review.
  - id: new
    severity: Minor
    kind: defect
    family: artifact-without-removal-path
    title: |
      The untracked adapter's last review-base copy is never deleted
    detail: |
      file-edit.md:70-72 deletes each review-base.XXXXXX copy only when the next one replaces it, so the last copy of every conversation stays in TMPDIR. This is the 2nd finding in this family. Rule: every temp file an adapter creates must state when it is deleted, including the end of the conversation or session. Apply that rule to the whole adapter, not just this site.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"438c7434-a49d-483b-b65d-c6bcc448e293","repository":"github.com/xianxu/ariadne","issue_id":"000399","recipe":"fix-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/da2ac20b4554708e","reviewed_head":"e40d16b5fc230bf36f13c69415f21905f96d0698","slot":"ariadne:2","requested_at":"2026-10-16T00:05:13.757063Z"},"phase":"finished","pid":36302,"started_at":"2026-10-16T00:05:14.072286Z","finished_at":"2026-10-16T00:06:07.030111Z","session_id":"c387247f-a9f9-448b-a918-d3bbea4ff4e7","outcome":"returned"}] -->
