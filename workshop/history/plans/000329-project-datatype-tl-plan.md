# #329 — Project datatype as a TL-driven coordination plane: plan

Source: the issue's Spec, and the design note approved at the batch-2 recap
(`ops/workshop/pensive/2026-10-10-01-pensive-project-datatype-tl-centric.md`).
The hub (ops:0, 10-12) confirmed two points:
- The lifecycle is `defined → executing → closing → closed`, with `dropped` kept as the abandon exit.
- The ops conformance item is delivered as a patch the TL applies, because workers don't edit ops project files.

**Goal.** A project stops being a shipping plan (deadline, planned finish, commit baseline). It becomes a plane on
which a TL coordinates cross-issue work:
- a Charter;
- Trains with cross-train edges;
- append-only Decisions;
- a forecast derived from the trains and the calibration ledger, never typed.

**Architecture.**
- The vocabulary (`construct/vocabulary/project.cue`) carries the new statuses, the guards and a field ban.
- A pure train parser and a pure scheduler live in `cmd/sdlc/internal/project`.
- The IO seams in `cmd/sdlc` feed the parser and scheduler with issue cards and calibration ledger rows.
- The #182 calendar forecast (throughput share against a deadline) and the Phase-A fog and calendar ledgers go away.
  They existed only to serve the deadline baseline (ARCH-DRY: one forecast, not two).

## Decisions

- **D1. Statuses.** The categories are `forming: [defined]`, `executing: [executing, closing]` and `terminal: [closed,
  dropped]`. `ideation`, `committed` and `paused` are removed. `committed` and `paused` only described a baseline,
  and a project with no baseline has nothing for them to hold. Ideation lives in parley or pensive, as before.
  - Transitions:
    - `defined→executing` (event `execute`, guard `trains-present`);
    - `executing→closing` (event `wrap`, guard `mvp-landed`);
    - `closing→executing` (event `reopen`, used when new MVP work is filed);
    - `closing→closed` (event `close`, guard `retro-recorded`, run through `sdlc project close`);
    - `→dropped` from defined (no guard), and from executing or closing (guard `retro-recorded`).
  - `IsExecuting` covers both executing and closing, so close's retro nudge and `close --drop` keep working.
  - **Renames the terminal `done` to `closed`** (the Spec's word, which the hub confirmed). Archived records stay
    frozen at `done` (eight across the fleet), and none is migrated, because nothing validates or branches on an
    archived record's status:
    - `project validate` and the changed-instance gate read only `workshop/projects/` (`Discovery().Home`);
    - `find`/`resolve` read `workshop/history/projects/` but match by ref marker only, never by status or schema;
    - `dropTerminalLegacy`, the one status reader outside the active home, reads the legacy
      `brain/data/project/`, which is empty since the #171 migration (checked 10-12).
    
    So D2's `deadline` ban never sees the two archived records that carry deadline fields
    (`ariadne/.../project-management-primitive.md`, `metis/.../arena2-playground-s6e7.md`). A test pins this:
    `project validate` over a fixture with an archived `status: done` record plus `deadline:` passes. Hardcoded
    `"done"` sites in project code (`project.cue`, `projectsetstatus.go:140`, the comments in `pkg/vocab` and
    `discover.go`) change to `closed` or go through the vocabulary (`IsTerminal`, or the `close` edge's `To`).
- **D2. Deadline fields are banned.** In `#Project`, `deadline?: _|_` and `planned_finish?: _|_` make
  `sdlc project validate` flag stale records. The old schema was open, so without the ban validation could not flag
  a stale record (ARCH-DATA: the schema states the invariant). `Metadata.Deadline`/`PlannedFinish`, the `list`
  column and the `show` lines go. `list` shows `tl` instead.
- **D3. Sections.** The scaffold sections are, in order:
  - `## Charter`: problem, cross-issue requirements and scope; replaces `PRD`.
  - `## Trains`: the trains table and the edge list.
  - `## Breakdown`: the checkbox list that close ticks for non-TL projects. Detail blocks become optional pointers.
  - `## Log`.
  - `## Decisions`: append-only, at the end of the file.
  
  `## Estimate` is removed, since the forecast is derived.
- **D4. Trains grammar** (parsed from `## Trains`):
  - The table has the header `| Train | Cars, in order | Slot |`; columns are matched by header prefix
    (`Train`, `Cars`, `Slot`).
  - The cars cell is a sequence of refs. A bare `#N` inherits the repo of the previous car. Parenthesized text is
    stripped. Separators (`→`, `⇒`, `/`, `,`) all mean "next car in this train": one slot runs a train serially.
  - Edges are bullet lines containing `⇒`. Every ref before the first `⇒` is a prerequisite of every ref after it,
    and parenthesized text is stripped first. `~` lines are overlap notes, which the scheduler ignores.
  - A ref is normalized to `repo#N` (an int, so `#0329` equals `#329`). A ref with no repo, and no previous car to
    inherit from, takes the project's own repo.
- **D5. find and resolve.**
  - `ProjectMatch` gains `Train`, `Car` and `Cars` (1-based position and train length). They are filled by
    parsing the matched file's Trains.
  - Text output becomes `path  (train: <name>, car i/n)`.
  - The JSON `resolveFile` gains optional `train` and `car` fields. parley reads `.files[].path`, so extra fields
    are safe.
  - An issue in no train prints the path alone.
- **D6. Forecast.** `sdlc project forecast [--slug S | --path P] [--brain-dir]` is read-only.
  - **Calibration** uses the ledger rows with a build/converge split, `Estimate > 0` and `BuildHours > 0`, not
    excluded, newest per issue (the same eligibility as `ConvergeFactors`).
    - The point multiplier is `m = Σ(build+converge) / Σestimate`, which equals `1/r × (1+k)` with
      `r = Σest/Σbuild` and `k = Σconv/Σbuild`. That is exactly the Spec's formula, aggregated.
    - The range uses the 25th and 75th percentiles of the per-issue `m_i`.
    - With fewer than 3 rows, the range collapses to the point, and a note says so.
  - **Car hours.** A terminal issue counts 0. An open issue counts `estimate_hours × m` (low, point and high).
    An unestimated issue gets the class prior, the median `Estimate` of the eligible rows, and is named in a note.
    An unresolvable ref also gets the prior, with a warning.
  - **Schedule.** The cars form a DAG with two kinds of edge: the previous car in the same train, and the ⇒
    prerequisites that are in this project. `finish(c) = max(finish(preds)) + hours(c)`, computed separately for
    the low, point and high cases. A cycle is an error that names a car on it.
  - **Output**, one paragraph:
    - remaining focused hours in total;
    - the critical path `low–high h (point P)`;
    - the train that ends last;
    - the number of trains with work left;
    - calibration `n` and `k`;
    - notes.
    
    There is no date (a date is set against the forecast, never by it).
  - `project show` and `project status` print the same line on a best-effort basis (the existing `forecastLine`
    seam), quiet when there are no trains.
- **D7. Close and drop.**
  - `sdlc project close` runs `closing→closed` with `retro-recorded`. `--drop` works from executing or closing.
  - Gone: the fog-factor guard, the Phase-A rollup, the Calendar ledger, `phasea.go`, the `--no-ledger` flag and
    the brain fog ledger writer. That writer's file stays in brain as history, with no writer (ARCH-FUNERAL: the
    family ends here).
  - The retro line drops its deadline clause.
- **D8. set-status.**
  - The flags `--reality`, `--coverage` and `--planned-finish`, and the commit forecast, go.
  - `GuardCtx` gains an `IssueTerminal func(ref) (bool, error)`, built in `projectsetstatus.go` from the existing
    cross-repo card resolver `projectIssueLookupFn(ctx, ref, repoDir)` (the var in `projectstatus.go`, which reads
    tracker cards and details), with `vocab.Issue().IsTerminal`. Tests replace that same var with a map-backed fake
    (the `projectclose_test.go:56` pattern). The forecast's car hours use the same seam (ARCH-MOCK). With it,
    `mvp-landed` checks every `mvp_scope` ref; it refuses on an empty `mvp_scope`, naming the missing field.
  - `--to closed` is refused and points at `sdlc project close`.
- **D9.** `project throughput` and `estimate/throughput.go` (`SpanThroughput`, `ThroughputBaseline`,
  `ParseBaselineTSV`) are deleted. The blessed baseline's only reader was the #182 calendar forecast, which M1
  deletes. `brain/.../throughput-baseline.tsv` stays as history, with no writer and no reader (ARCH-FUNERAL). Kept as
  is: the close-time project tick and the TL skip (#326).
- **D10. ops conformance.** Run `sdlc project validate` (built from this branch) read-only against ops'
  `ariadne-robustness-1` and `couch-broadcast-web`. Write a patch file for the TL that renames PRD to Charter,
  promotes the trains table and edges to `## Trains`, adds `## Decisions`, and sets `status: executing`. Then run
  `sdlc project forecast` on the patched copy, held in scratch, to check that the trains parse. Fleet files that
  need migration are listed in the report:
  - `ideation`: pair `couch-slots-v2`, parley.nvim `parley-packaging` and `parley-v1-release`;
  - deadline fields: metis `arena3-rogii-wellbore` and `ml-research-workflow`.

**ARCH-FUNERAL.** The change creates nothing durable: the forecast is computed on read, and the new sections are
authored by the TL. It also ends two writers, the fog ledger and the calendar ledger.

## Plan

- [ ] M1 — Vocabulary and lifecycle:
  - `project.cue` (D1–D3), regenerate `project.json` with `make vocab-embed`, and update the cue testdata and
    `vet_test.sh` fixtures;
  - guards: `trains-present`, `mvp-landed` with `GuardCtx.IssueTerminal`, and removal of the old guards;
  - set-status (D8), close and drop (D7), and the deletion of `forecast.go`/`phasea.go` and the calendar forecast
    wiring;
  - `metadata`/`summary`/`list`/`show`, and the scaffold sections;
  - update every test listed in the survey;
  - `construct/datatype/project.md` rewrite (the prose drift test), `helptext/project.md`, README, and the
    `atlas/workflow/{sdlc-binary,vocabulary}.md` updates.

  Tests:
  - vocab predicates and the transition table;
  - each guard, including mvp-landed with a fake lookup;
  - set-status `executing→closing` refusing on an open MVP issue;
  - close `closing→closed`;
  - validate refusing `deadline:`.
**Non-goals:** the Spec's open questions (one file or several as a project grows, where cross-operator allocation
is negotiated, `mvp_scope` as a cut point on the trains) and any migration of peer files, which goes to the TL as a
list.

**Envelope (ARCH-CONSTRAINTS):** a forecast does one card lookup per car, the same per-row lookup cost `project
status` already pays for Breakdown rows: about 50 cars, local git reads, well under a second. `show` and `status`
print the forecast best-effort; a lookup error becomes a note and never fails the verb.

- [ ] M2 — Trains, find/resolve and forecast:
  - `internal/project/trains.go` (D4). Test strategy: table-driven over hand-authored markdown (the ops file's real
    rows among the cases), plus a fuzz seed of malformed rows. The invariant: it never panics, and a malformed row
    is skipped with a warning.
  - `trainforecast.go`, the pure scheduler. Test strategy: DAG invariants. The critical path is at least every
    train's own sum and at most the total; a ⇒ edge never shortens it; low ≤ point ≤ high; a cycle is always
    reported. Separate cases cover the prior and the collapse below 3 rows.
  - `ProjectMatch` train/car, plus find text and resolve JSON tests (D5);
  - `sdlc project forecast` on a fixture project with a fake lookup and a ledger fixture, printing a range;
  - show/status forecast line.
  
  Then the ops conformance patch (D10).

## Revisions

- 2026-10-12, plan-quality round 1. D1 now states the `done → closed` rename and why archived records need no
  migration; the first edit attempt failed to apply, and round 2 caught it. D9 now deletes `project throughput`, whose only reader goes away. D8 names the lookup seam. Added:
  non-goals, the envelope, and test strategies instead of a list of cases.
- 2026-10-12, during M1. The milestone boundary moved:
  - The `trains-present` guard needs the trains parser, so the parser moved into M1. The scheduler and
    `sdlc project forecast` were built in the same pass, because they replace the deleted calendar forecast that
    `show`/`status` printed.
  - M1 = vocabulary, lifecycle, the trains parser and the forecast. M2 = find/resolve train and car, plus the ops
    conformance patch.
  - D2's ban is a sentinel string, not `_|_`. cue's `_|_` diagnostic names neither the field nor a remedy (measured:
    `explicit error (_|_ literal) in source:`). The string constraint makes the diagnostic name the field and point
    at `sdlc project forecast`.
  - `ListActiveProjectFiles`/`ListActiveInRoots` (whose only reader was the contention forecast) and the
    `project close --no-ledger` gate row are deleted with what they served.
- 2026-10-12, M1 review round 1 (FIX-THEN-SHIP):
  - D6's `--path P` became a positional `[file]`, matching `sdlc project validate <file>`.
  - The cycle error now walks stalled predecessors, so it names a car on the cycle (BR-1).
  - D1's archive test landed as `TestProjectValidationScopeExcludesArchive` (BR-2). It pins that `validate --all`
    and the changed-instance gate never reach `workshop/history/projects/`.
- 2026-10-12, M2 (review round 3). Each M2 item matched to what was delivered, in both directions:
  - Promised and delivered:
    - D5 train and car in find/resolve;
    - D10 ops conformance as `workshop/plans/000329-ops-conformance.patch`. It was made from ops' current files and
      passes `git -C ops apply --check`; after it applies, both ops projects pass `sdlc project validate`. The TL
      applies it in ops and verifies Done-when 4.
  - Delivered beyond the plan, each because the plan's goal needed it:
    - **Train-only matching.** The ops trains table cites ariadne#329 with no `[ref]` row, so marker-only discovery
      could never name its train. `DiscoverInRoots` now matches a row marker or a train car, and is the single source
      of that rule.
    - **The close-tick skip for `RowRef=false`.** Without it, a train-only match in a project without `tl:` would
      demand a detail block at milestone-close.
    - **Section presence in `project validate`.** Frontmatter alone cannot tell a shipping-plan record from the new
      shape, and Done-when 4 needs a check. Only non-optional scaffold sections are required (a new `optional` field
      on `#ScaffoldSection`).
    - **The `claimant-ownership.md` migration.** ariadne's one live project would otherwise fail the new section
      check.
- 2026-10-12, after M2 (advisories in round 4):
  - D10's clauses differ per file:
    - `ariadne-robustness-1` moves to `status: executing`, with its trains table and edges moved into `## Trains`.
    - `couch-broadcast-web` stays `defined`. It runs in waves and has no trains, so it gets an empty trains table
      until the TL authors trains (`trains-present` guards the move to executing). Its Estimate line is kept as
      `**Sizing.**`, like the other file's.
  - Because the ops Log grows, a saved patch goes stale. The handoff is therefore the script
    `workshop/plans/000329-ops-conform.py`, re-run against the current files, with the patch beside it as a
    reviewable preview.
