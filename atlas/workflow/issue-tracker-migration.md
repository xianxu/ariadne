# Issue tracker migration and cutover

How a legacy repository (full-frontmatter details on main) becomes an
[issue tracker](issue-tracker.md) repository, and the operator procedure for
doing it across the fleet (#252 M4). The command is `sdlc issue migrate`
(`cmd/sdlc/issuemigrate.go`); its decisions are the pure
`tracker.PlanTrackerMigration` (`internal/tracker/migration.go`) and the card
derivation in `internal/issue/migrate.go`.

## What the migration does

- **Cards.** One per ID, active and archived alike, so allocation (`max + 1`)
  never reuses an ID. An active file seeds its ID's card; otherwise the newest
  archived file does. Two active files sharing an ID refuse (renumber first);
  archived duplicates are reported and keep their own records.
- **Active details** gain `card_mirror` (and, for pre-template issues, a
  `## Problem` heading above their preamble, or a labelled placeholder when
  there was none). Nothing else in them changes.
- **Archived details are never rewritten.** Their frontmatter stays their
  terminal authority (`historyFileIsTerminal` reads unmirrored files as
  before); their cards are built tolerantly and every inferred value — N/A
  hours, inferred status, title from the slug, Problem from the preamble — is
  listed in the dry run.
- **Legacy codecomplete** imports bound (`tracker.completion`, token
  `migrate-<id>`) only to exactly one provable legacy close: the commit that
  recorded codecomplete on a branch (main's issue-sync copy of it does not
  count) with only docs after it. Then the branch lands like any tracked close.
  A close already in main's history (landed, never marked done) binds main's
  record and settles to done. The details' mirror pins the final, bound card.
- **Marker.** One main commit adds `workshop/issue-tracker.json`
  `{version, tracker_root}` beside the details conversion. The marker is what
  makes the era explicit to the binary.

## The cutover guard

Commands open trackers guarded (`Repository.GuardCutover`): every read and
card preparation proves the checkout's marker names the tracker's root commit.
A tracker without a marker (a checkout from before the cutover, or an
unfinished migration), a marker without a tracker, or a marker naming another
root refuses with the next action — never a stale or legacy read. Unmirrored
details in a cut-over repository refuse the legacy close/change-code paths
(`tracker.RefuseLegacyDetails`). Legacy-only entrypoints refuse on the marker:
`issue publish`, `issue sync --push` or on a resting branch, and the
`Makefile.workflow` / `scripts/close-issue.py` shell fallbacks.

## Operator procedure

1. **Build and propagate the owner binary.** Land #252 in ariadne; build
   `bin/sdlc` from it and propagate the base layer (`weave`) so each
   participating repository's composed `AGENTS.md`/`CLAUDE.md`,
   `Makefile.workflow` and scripts carry the tracker-era guidance.
2. **Inventory, read-only.** In each participating repository (never a brain),
   from a clone that has fetched every slot's branches, run
   `sdlc issue migrate`. It lists cards, inferences, duplicates and refusals and
   prints a digest; it changes nothing. Clones on other machines run it too:
   their local-only branches and dirty worktrees are invisible elsewhere.
3. **Freeze that repository's writers.** Stop agent sessions, aliases,
   scripts and cron entries that run `sdlc` or the workflow Makefile in it and
   its slots. The new binary still runs the legacy workflow wherever the marker
   is absent, so once every entrypoint runs it, repositories cut over one at a
   time — no fleet-wide freeze. Old binaries cannot honor the marker: the new
   build must be installed everywhere first.
4. **Resolve refusals under the old workflow:** publish unsynced card fields
   (`sdlc issue sync --push` on the old binary), land or drop branches that
   edit archived issues, renumber colliding active IDs, commit or discard dirty
   issue edits, and land or re-close unprovable codecomplete issues. Re-run the
   dry run until it reports none; review the inferences.
5. **Apply:** `sdlc issue migrate --apply --expect <digest>`. It bootstraps
   `issue-tracker`, then publishes the main conversion. If interrupted, rerun
   the same command: each phase is recognized. If main moved, re-run the dry
   run and apply the new digest (the tracker already in place is adopted only
   if its root holds exactly the new plan's cards).
6. **Bring checkouts across.** `git pull` every resting checkout (all slots).
   A branch from before the cutover refuses sdlc commands until brought
   across — now, or on its next use — with `sdlc issue migrate --reconcile`
   (or by merging origin/main). Reconcile proves every details file the branch
   changed kept its imported card's fields and has a card, then merges main's
   migration commit so later merges start from converted details; card-field
   edits are applied on the card from a caught-up checkout first.
7. **Verify, then unfreeze:** `sdlc issue list` shows card statuses,
   `sdlc issue recovery list` is empty, a claim on an open issue succeeds.

The step-by-step operator checklist is
`workshop/plans/000252-issue-cards-tracker-ref-cutover.md` (archived with #252).

**Recovery.** Before any tracker write after the cutover, the migration can be
abandoned by deleting the `issue-tracker` branch and reverting the main commit.
After tracker writes resume, never roll back to the import snapshot: repair
forward. Order: ariadne first (it owns the binary), then downstream repositories;
product-owned consumers that need changes get issues in their owning repositories.

## Verification pointers

- `internal/issue/migrate_test.go`: round trip over generated legacy files; a
  reconciled branch copy equals main's conversion byte for byte.
- `internal/tracker/migration_test.go`: generated populations (bound
  codecomplete cards included) parse as a valid snapshot with the right max
  ID; every conversion pins its final card's blob; duplicates, divergence and
  close provability.
- `internal/tracker/cutover_test.go`: marker strictness; the guard on reads,
  stale reads and card writes; a marker without a tracker never reads legacy.
- `cmd/sdlc/issuemigrate_test.go`: dry run, apply, resume, foreign tracker,
  reconcile then clean merge, an imported close merged, landed and settled.
- `cmd/sdlc/leftover_e2e_test.go`: a leftover branch is locked for every verb
  and the publish gate, catches up by reconcile or by merging main, then lands
  to done; an invisible branch's card edits recover via the card; an invisible
  branch-only issue cannot land; backing out a cutover (with the prune).
- `cmd/sdlc/trackedlegacy_test.go`: legacy entrypoints, Makefile fallbacks, and
  `lint-ids` refusing cardless details.
