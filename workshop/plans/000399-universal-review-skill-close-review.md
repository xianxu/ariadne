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
