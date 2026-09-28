# Boundary Review — ariadne#252 (milestone M4)

| field | value |
|-------|-------|
| issue | 252 — Issue cards: card fields on a tracker ref, details on the branch |
| repo | ariadne |
| issue file | workshop/issues/000252-issue-cards-tracker-ref.md |
| boundary | milestone M4 |
| milestone | M4 |
| window | c0bfb39594c27f2949fb2ff1e21b17d0a98cf9c2..5df86677266267d4e761f9243b73bf1ff3a340ae |
| command | sdlc milestone-close --issue 252 --milestone M4 |
| reviewer | claude |
| timestamp | 2026-09-26T18:17:37-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M4 is mostly solid. The migration is split into a pure plan (`PlanTrackerMigration` plus the card derivation) and a thin IO shell, the cutover marker is parsed strictly, the guard is wired into the single snapshot read path, and the slot-cycle e2e covers every Task 8 scenario the plan names. One correctness bug blocks the boundary. When a legacy **codecomplete** issue is migrated, main's converted details point `card_mirror` at the *unbound* card. The tracker actually holds the *bound* card, whose bytes and blob ID differ. After the cutover that issue's baseline blob is missing, and merging its reconciled branch conflicts on the `card_mirror` line. That is exactly the path the plan says "can next merge to done/archive". The test for that case stops before the merge, so it passes.

1. **Strengths**
   - `internal/tracker/migration.go:95`: the plan never fails. Every problem becomes a refusal, so one dry run reports everything, and the digest covers cards, conversions, duplicates and refusals. The apply is tied to what the operator reviewed (`issuemigrate.go:113`).
   - `internal/tracker/cutover.go:84`: the guard sits in `Repository.read`, the one snapshot path. `PrepareCreate`, `PrepareUpdate`, `ChangeCard`, `UpdateCard` and `LocalSnapshot` were all rerouted through it (`candidates.go`, `repository.go`), so writes cannot get past it. `records.go:119` correctly refuses to answer a cutover mismatch with a stale read.
   - `issuemigrate.go:398-453`: a resumed bootstrap adopts an existing tracker only if its root has exactly the plan's blobs. `TestIssueMigrateResumesItsOwnTrackerOnly` proves a foreign tracker is refused.
   - Branch refusals are grouped by (path, reason) across stacked branches (`migration.go:186-212`). This fixed a real 53→10 noise problem found in a rehearsal.
   - `stale_guard_test.go` turns the stale-read labelling rule into a source-level check over every PreferFresh consumer, instead of fixing only the one instance the e2e found.

2. **Critical**
   - `cmd/sdlc/internal/tracker/migration.go:133-157`: for a codecomplete issue, `card` is replaced with the `bindLegacyClose` output, but `mirrored` still carries `mirrorLine(CardBlobOID(unbound card))` from `SplitCardWithFormat`.
     - That unbound blob is never written; the tracker root holds the bound card.
     - Main's converted details therefore name a baseline the tracker cannot supply, so any refresh or validation refuses on a missing baseline.
     - `--reconcile` on the branch writes the *bound* card's OID (`issuemigrate.go:529`). Main and the branch now add different `card_mirror` lines to the same hunk, and the PR merge conflicts.
     - This contradicts the atlas claim "a reconciled branch copy equals main's conversion byte for byte" and the Task 7 promise that the imported PR can next merge to done/archive.
     - **Fix:** attach the mirror only after the card's final bytes are known. For example, have `MigrateActiveDetails` return normalized details without a mirror, and have the plan add `mirrorLine(CardBlobOID(finalCard))` once, after binding.
     - **Tests:**
       - (a) A pure invariant over generated populations that includes a bound codecomplete case: every conversion's `MirrorBaselineOID` equals `CardBlobOID` of its manifest card.
       - (b) Extend `TestIssueMigrateImportsAProvableLegacyClose` so it merges the reconciled branch into migrated main without conflict, then runs the landing to done.

3. **Important:** none beyond the Critical.

4. **Minor**
   - `issuemigrate.go:466`: the moved-main error says "the new plan resumes from it". That only holds if main's movement touched no issue files; otherwise `sameTrackerFiles` refuses, and the only way forward is the hand abandonment in the atlas Recovery section. The message should say that.
   - `cutover.go:105`: every guarded read now spawns an extra `git rev-list --max-parents=0`. The measured envelope ("2 Git processes") was taken on an unguarded repository and should be re-measured with the guard.
   - `issuemigrate.go:407` re-implements `TrunkFile.HasRoot`'s root listing with `env.git` (ARCH-DRY).
   - The consumer inventory assigns "tracker ID validation" to `40-duplicate-issue-id.sh`. The Revision says `issuelintids.go` needed no change, and only an atlas note was added. A hand-made details file with no card can still land through the GitHub UI and later collide with a card-allocated ID published by `move-detail`. Either add a card check or record the accepted risk.
   - `Makefile.workflow:36` and `close-issue.py` detect a cut-over repository by the marker file alone, while Go's `CutOver` also treats a fetched tracker as cut over. A pre-cutover branch without `bin/sdlc` can still run the fallbacks. This is low impact.
   - `bindLegacyClose`: an issue whose branch already landed on main but was never marked done is refused with "code after its close". The next-action text ("land the issue…") is misleading because it has already landed.

5. **Test coverage**
   - The e2e is thorough: slot cycle, spin-off handoff, competing claim, fresh CI clone, disconnected read/mutation.
   - The gap is the codecomplete import path. It asserts the binding and `ownedCompletions`, but never checks the conversion's mirror baseline, never merges, and never lands. That gap is exactly what let the Critical ship.
   - `TestPlanTrackerMigrationOverGeneratedPopulations` checks only `HasMirror`; it should also check that each mirror's baseline equals its card's blob ID.

6. **Architecture**
   - **ARCH-DRY — flag (Minor):** duplicated root listing (above). The marker filename is repeated as a literal across Go, Make and Python; that is acceptable across languages, but it deserves a comment pointing at `CutoverMarkerPath`.
   - **ARCH-PURE — pass:** plan and derivation are pure; `issuemigrate.go` only gathers inventory and applies.
   - **ARCH-PURPOSE — flag:** the codecomplete migration path, one of Task 7's stated purposes, doesn't work end to end (the Critical).
   - **ARCH-MOCK — pass:** real Git fixtures with bare remotes, and the e2e runs the production seams.
   - **ARCH-CONSTRAINTS — flag (Minor):** the guard's extra process per read is outside the measured envelope.
   - **ARCH-SECURE — pass:**
     - The marker is parsed strictly (unknown fields refused, version and OID shape checked).
     - Archived frontmatter is treated as untrusted and every inference is reported.
     - Reconcile skips non-regular files.
     - Duplicate YAML keys are refused.
   - **ARCH-ORDER — pass:** the two apply phases are each observable and resumable. Tracker-without-marker, marker-without-tracker and foreign-root states are enumerated and refused. The remaining gap is the moved-main hint (Minor).
   - **ARCH-FUNERAL — pass:** the marker and tracker persist for the repository's lifetime by design, and the migration leaves no refs or sidecars behind.

7. **Plan revisions**
   - Add a Revision recording the mirror/binding ordering fix and the added end-to-end merge proof for an imported close.
   - Correct the atlas "Verification pointers" byte-for-byte claim, or make it true.
   - Record the disposition of the `40-duplicate-issue-id` "tracker ID validation" row.

```findings
findings:
  - id: new
    severity: Critical
    family: deferred-effect-input-drift
    title: |
      Codecomplete migration: main conversion mirrors the unbound card; tracker holds the bound card (missing baseline, merge conflict)
    detail: |
      This is the 4th finding in family deferred-effect-input-drift. migration.go:133-157 derives `mirrored` from the pre-binding card, then replaces `card` with SetCardCompletion's output, so main's card_mirror names a blob never written, while --reconcile on the branch mirrors the bound blob. Result: refresh refuses on the missing baseline, and the PR merge conflicts on the card_mirror line. Rule for the class: any artifact that pins a source identity (OID/digest) must be computed from the source's final bytes by one constructor called after the last mutation, and the pure plan must assert pin == identity(final). Fix by attaching the mirror after binding. Add a population invariant (each conversion's MirrorBaselineOID == CardBlobOID(card)) and extend TestIssueMigrateImportsAProvableLegacyClose through merge and landing.
  - id: new
    severity: Minor
    family: next-action-hint-correctness
    title: |
      Moved-main apply error promises resumption that sameTrackerFiles refuses when issue files changed
    detail: |
      issuemigrate.go:466 says the new plan resumes from the existing tracker; any change to active or archived details changes the plan's cards, so the bootstrap check refuses and only the atlas's manual abandonment works. The message should name that path.
  - id: new
    severity: Minor
    family: operating-envelope-enforcement
    title: |
      Cutover guard adds a rev-list process per guarded read, outside the measured envelope
    detail: |
      cutover.go:105 runs HasRoot on every read(), including each CAS retry; the plan's 2-process, 0.85 s measurement predates the guard.
  - id: new
    severity: Minor
    family: shared-helper-extraction
    title: |
      applyTrackerBootstrap re-implements TrunkFile.HasRoot's root listing
    detail: |
      issuemigrate.go:407 runs rev-list --max-parents=0 through env.git; a TrunkFile Roots helper shared with HasRoot would keep one implementation.
  - id: new
    severity: Minor
    family: core-concepts-inventory-drift
    title: |
      Consumer inventory's tracker ID validation for 40-duplicate-issue-id is neither delivered nor dispositioned
    detail: |
      This is the 5th finding in family core-concepts-inventory-drift. Only an atlas note was added; a cardless details file merged through the UI can later collide with a card-allocated ID. Rule: every inventory row needs either a delivering diff or an explicit Revision disposition. The close gate could check this by diffing inventory rows against the window's name-status.
```

---

## Re-review — 2026-09-26T18:51:40-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 252 — Issue cards: card fields on a tracker ref, details on the branch |
| repo | ariadne |
| issue file | workshop/issues/000252-issue-cards-tracker-ref.md |
| boundary | milestone M4 |
| milestone | M4 |
| window | c0bfb39594c27f2949fb2ff1e21b17d0a98cf9c2..4824965e6d6c6783a39da203fc70a531e624b626 |
| command | sdlc milestone-close --issue 252 --milestone M4 |
| reviewer | claude |
| timestamp | 2026-09-26T18:51:40-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: medium
```

Round 10 fixes BR-34 at the root, not just at the site it named. `issue.MirrorDetails` (`internal/issue/migrate.go`) is now the only function that builds a mirror. `PlanTrackerMigration` calls it after the close is bound (`internal/tracker/migration.go:157`), and `ReconcileLegacyDetails` uses the same function, so main and the reconciled branch now produce identical bytes. Two regression tests would fail without the fix. The generated-population test now includes bound codecomplete cards and checks every conversion's pin against its final card blob. It also checks that an unchanged branch copy reconciles to main's exact bytes. The end-to-end imported-close test now checks main's pin against the blob the tracker holds, then merges, lands and settles to done. The four Minors are also handled. `TrunkFile.Roots` is now the only root listing. The guard caches the (root, tip OID) pair it has already checked; the tip is an OID, so the cache is safe. `lint-ids` has a pure core, a real-Git exit-code test, and exits 2 when it cannot run. The fallbacks now use the same cut-over test as `tracker.CutOver`. The issue, tracker, gitx and targeted cmd/sdlc tests all pass at head. Nothing blocks shipping; there are two new Minors.

1. **Strengths**
   - `internal/issue/migrate.go`: `MigrateActiveDetails` no longer builds the mirror, and `MirrorDetails` is the only constructor. This removes the whole ordering hazard behind BR-34, and the new lesson states the general rule (build a pin from the final bytes).
   - `internal/tracker/migration_test.go:90-108`: the population test checks two things that don't depend on the implementation: the pin equals the final card's blob identity, and a branch copy reconciles to main's exact bytes.
   - `cmd/sdlc/issuemigrate_test.go:255-274`: the end-to-end test now runs the full path the bug broke (merge, land, reconcile, done).
   - `issuelintids.go`: pure `addedDetails`/`detailsWithoutCards` behind a thin IO shell. An unreadable tracker exits 2 instead of reporting clean.
   - `cutover.go:105`: the cache key is (marker root, tip OID). A re-created tracker has a new tip, so it cannot hit a stale pass.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `Makefile.workflow:40` and `scripts/close-issue.py:50` hardcode `issue-tracker`, while Go reads the name from `vocab.Issue().Discovery().Tracker`. The test also hardcodes it, so a vocab rename would pass the test while the fallbacks quietly stop refusing (ARCH-DRY).
   - `migration.go:245-262`: a branch whose close already landed is now treated as on main, and `CodeAfter` is ignored. Code the branch added after its close and never landed is dropped when the card settles to done (ARCH-PURPOSE).

5. **Test coverage**
   - BR-34: a regression test would fail without the fix. The population test's codecomplete cases with the pin assertion would go red under the old order.
   - BR-36's cache has no test that counts processes. The benchmark measures `HasRoot` itself, not whether the cache stops a second call. This is acceptable because the change only affects performance.
   - The landed-close binding has a pure-planner case. There is no end-to-end test where the anchor is an ancestor of main.

6. **Architecture**
   - **ARCH-DRY**: pass in Go (`Roots`, `MirrorDetails`). Flag (Minor) for the tracker branch name hardcoded in the Makefile and Python fallbacks.
   - **ARCH-PURE**: pass. The planner and the lint core are pure; IO sits in `cardlessAdditions`, `legacyCloseAnchors` and `applyTrackerBootstrap`.
   - **ARCH-PURPOSE**: pass for BR-34 and BR-38; the whole class is fixed and the inventory row is delivered. Flag (Minor) for landed-close binding.
   - **ARCH-MOCK**: pass. Tests run against real Git in temporary repositories.
   - **ARCH-CONSTRAINTS**: pass. The guard costs 54 ms and one process per tracker generation over a 10,000-write history, and the benchmark was added.
   - **ARCH-SECURE**: pass. The marker is parsed strictly, and the lint exits 2 when it cannot run.
   - **ARCH-ORDER**: pass. The one-constructor-after-last-mutation rule is now structural. The cache holds no state between events other than one checked pair.
   - **ARCH-FUNERAL**: pass. This round creates nothing durable (the cache lives only in memory for one Repository).

7. **Plan revisions:** none needed. The round-10 Revisions entry matches the code.

```findings
dispose:
  - id: BR-34
    disposition: addressed
    note: |
      MirrorDetails pins after binding (migration.go:157); population test asserts pin==CardBlobOID(final) incl. bound codecomplete; e2e merges, lands, settles.
  - id: BR-35
    disposition: addressed
    note: |
      issuemigrate.go:471-474 now names both resume and abandon-unwritten-tracker outcomes.
  - id: BR-36
    disposition: addressed
    note: |
      Guard caches verified (root, tip OID) in Repository.verified; benchmark measures 1 process per generation.
  - id: BR-37
    disposition: addressed
    note: |
      TrunkFile.Roots is shared by HasRoot and applyTrackerBootstrap.
  - id: BR-38
    disposition: addressed
    note: |
      lint-ids refuses cardless added details in tracker repos; pure unit test plus real-Git exit-code test.
findings:
  - id: new
    severity: Minor
    family: shared-helper-extraction
    title: |
      Makefile and close-issue.py hardcode the issue-tracker ref name that Go derives from vocab
    detail: |
      This is the 4th finding in shared-helper-extraction. The shell fallbacks cannot call the binary, so the rule is that every non-Go copy of a vocab constant is pinned by a test that reads vocab. trackedlegacy_test also hardcodes the literal, so a rename would not turn it red.
  - id: new
    severity: Minor
    family: landing-completion-proof
    title: |
      A landed branch close ignores CodeAfter, so unlanded post-close branch code settles the card to done
    detail: |
      This is the 3rd finding in landing-completion-proof. Rule: done requires proof that every ref carrying the close has nothing beyond main after it. Apply CodeAfter to a landed anchor's branch tip relative to main, not only to unlanded anchors.
```
