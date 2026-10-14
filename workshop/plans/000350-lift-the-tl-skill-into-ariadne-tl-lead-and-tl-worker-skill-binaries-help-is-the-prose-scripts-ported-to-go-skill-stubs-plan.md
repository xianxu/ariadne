# TL skill binaries (tl-lead, tl-worker) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Lift the ops TL skill draft into ariadne as two skill binaries, `tl-worker` and `tl-lead`. Each one's `--help` is its prose. `tl-lead` also gets Go subcommands that replace the ops Python scripts, and weave delivers a pointer stub for each skill.

**Architecture:** Each binary has the shape of `cmd/doc-review`: a cobra root whose `Long` is a `go:embed`ded markdown file. `tl-lead status …` is a thin shell over `couch --session-sync`, behind a `SessionSync` seam with a stateful fake. `tl-lead trains` renders from the project file, reusing sdlc's own trains parser and plan counter. Both move from `cmd/sdlc/internal` into importable `pkg/` packages, and the old paths re-export them, so there is no second parser (ARCH-DRY). Car status comes from `sdlc issue show --json` and off-train slots from `sdlc fleet inventory --json`, both behind an exec seam with fakes (ARCH-MOCK).

**Tech Stack:** Go 1.26, cobra, go:embed; `couch --session-sync` (pair), `sdlc issue show --json`, `sdlc fleet inventory --json`.

**Non-goals:** no Couch change (the §15 line is the load); `tl-status.py` is retired, not ported; no new field on `sdlc issue show --json`; the ops pointer and deleting `scripts/*.py` are the TL's ops change after M3.

**Source of the prose:** `~/workspace/ops/skills-draft/xx-tl/SKILL.md` and the TL's dispatch messages. The draft is still changing, so M2 reads the then-current file, not today's. Per the issue's Revisions, no stories or lessons are copied: the prose is current rules and procedure, and every reference has the full `repo#NNN` form.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `workerHelp` (embedded prose) | `cmd/tl-worker/help.md` | new |
| `leadHelp` (embedded prose) | `cmd/tl-lead/help.md` | new |
| `markdoc` (fence/section/plan scanners) | `pkg/markdoc/{fence,section,plan}.go` | new (moved from `cmd/sdlc/internal/issue`) |
| `trains` (`IssueRef`, `Train`, `Edge`, `TrainPlan`, `ParseTrains`) | `pkg/trains/trains.go` | new (moved from `cmd/sdlc/internal/project/trains.go`) |
| `PlanBar` | `cmd/tl-lead/bar.go` | new |
| `RenderTrains` / `RenderOffTrain` | `cmd/tl-lead/render.go` | new |
| `ThreadEntries` / `PrependThread` | `cmd/tl-lead/threads.go` | new |
| `SeenSet` (merge, prune, foldable) | `cmd/tl-lead/seen.go` | new |
| `MarkerLines` (operator 🤖 markers in a file) | `cmd/tl-lead/markers.go` | new |

- **markdoc**: `FenceSpans`, `ScanMarkdownLines`, `SectionBody`, `SectionLineBounds`, `PlanItemsBody`, `CountPlanItems`, `MilestonesInPlanOrder`, `TickedMilestones` and `MilestoneTagPattern`. All are pure and need only the standard library. `cmd/sdlc/internal/issue` keeps its names as forwarding aliases (`type UnterminatedPolicy = markdoc.UnterminatedPolicy`, `var SectionBody = markdoc.SectionBody`, …). Its 125 importers are unchanged, and the existing tests move with the code.
  - **DRY rationale:** the one plan/section parser that sdlc's gates use is also what renders the TL's bar, so a bar can't disagree with what `close` counts.
- **trains**: `trains.go` moves verbatim (it is stdlib-only). `cmd/sdlc/internal/project` aliases its types and keeps `(*Doc).Trains`.
- **PlanBar(planBody) string**: counts `Mx` rows when the plan has any (`MilestonesInPlanOrder`/`TickedMilestones`), and checkboxes otherwise (`CountPlanItems`). It returns `▉▉▕ M 2/3`, `▉▕▕▕ tasks 1/4` or `no plan yet`.
- **RenderTrains(plan TrainPlan, cars map[IssueRef]CarState, defaultRepo) string**: the fenced Trains block. Glyphs: ✓ done, ◑ codecomplete, ◐ in flight (claimed and not open), ○ waiting. `CarState{Status, Claimant, Bar, Err}`. A car whose read failed renders `?` with no bar, never as waiting (ARCH-SECURE: a failed read is not negative evidence).
- **RenderOffTrain(slots []SlotFact, listed set) string**: one line per slot whose branch is an issue branch no train lists, or `_None._` when there are none.
- **SeenSet**: the IDs the TL has looked at (`.<file>.ss-seen`, JSON — the same name and format as `ss.py`, so the cutover keeps the TL's seen IDs and leaves no residue). It holds no state between events beyond that file; see ARCH-ORDER below.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `SessionSync` | `cmd/tl-lead/sessionsync.go` | new | `couch --session-sync` |
| `fakeSessionSync` | `cmd/tl-lead/sessionsync_fake_test.go` | new | stateful model of a session-sync file |
| `SDLC` | `cmd/tl-lead/sdlc.go` | new | `sdlc issue show --json`, `sdlc fleet inventory --json` |
| `fakeSDLC` | `cmd/tl-lead/sdlc_fake_test.go` | new | fixture observations and inventory |
| skill stubs | `construct/local/tl-lead/SKILL.md`, `construct/local/tl-worker/SKILL.md` | new | weave skill delivery (`xx-tl-lead`, `xx-tl-worker`) |
| §15 pointer | `AGENTS.base.md` §15 | modified | the constitution every slot loads |

- **SessionSync**: the methods `ReadSection(file, heading) (body, hash)`, `ReadAll`, `WriteSection(file, heading, bodyPath, hash, fold []string)`, `Status(file) (owed []string, lock string)` and `Diff(file)`. The real implementation calls `exec.Command(couch, "--session-sync", args...)` with an argv array, never a shell. Exit codes become typed errors: `ErrLocked` (3), `ErrStale` (6) and `ErrDeliveryGap`.
  - **Fake state model:** the file's text, a per-section hash (sha of the body), a lock holder (`operator` or none), a list of owed IDs, and a write counter. Its write semantics match Couch's help: a stale hash refuses, a held lock refuses, a missing `--if-hash` refuses, and folding an ID that isn't owed refuses.
  - **Live conformance:** `TestSessionSyncConformance` runs when `couch` is on PATH and `TL_LEAD_LIVE=1`. Against a temporary session-sync file it checks read, write, stale (6) and an operator lock (3). It is opt-in and runs at every close of this issue and whenever pair changes session-sync (logged in `## Log`).
- **SDLC**: `IssueShow(repoPath, n)` and `FleetInventory()` decode into local structs that hold only the fields `tl-lead` reads. Decoding is lenient here: `tl-lead` is a reader, and the observation's own strictness binds sdlc. `schema_version != 1` is refused visibly. The fake serves canned JSON shaped like real `sdlc` output, captured once into `testdata/` from a scratch repository.

### ARCH notes

- **ARCH-PURPOSE:** the Done-when names four consumers: the worker's rules, the lead's prose, the scripts and the delivery path. Each milestone closes one of them end to end. The ops-side pointer and the TL switching to the binaries are the TL's ops change; M3 ends by handing that over, and nothing here defers it.
- **ARCH-CONSTRAINTS:** `trains` is an interactive TL pass, run about every few minutes. The Python baseline is about 10 s. It makes N ≤ 30 car reads, at most 8 concurrent, with a 60 s timeout each; a timed-out car renders `?`. `status` verbs are light, with one or two couch calls each. Load and memory are N/A: there is no hot loop and no large input.
- **ARCH-SECURE:** every input crosses a process boundary: project.md, issue files and couch/sdlc JSON. Each is parsed into a typed value, and malformed input degrades to a visible `?` or a refusal, never a fabricated count. Body files reach couch as an argv path. No credentials are touched.
- **ARCH-ORDER:** the state is Couch's file plus the seen-set. The events that matter:
  - **A submit lands mid-turn.** Its ID is not in the seen-set, so `--fold-owed` doesn't fold it until a later `submit` or `owed` lists it.
  - **Stale hash.** Re-read, recompute from the new body, and retry, at most 4 times, then refuse.
  - **Lock held.** Refuse with exit 3 and no retry. The update is queued by the TL.
  - **Process death between listing owed IDs and recording them as seen.** The IDs stay unfolded, which is the safe direction.
  - **Death between couch's write and the seen-set prune.** The prune is idempotent at the next call.

  Tests drive these through `fakeSessionSync`, which can inject a stale write or a lock at a chosen call number.
- **ARCH-FUNERAL:** the seen-set file is pruned on every write to the IDs still owed (a deliberate change from `ss.py`'s grow-only `mark_seen`), so it is bounded by the owed set. It dies with the status file and is gitignored alongside it. `<file>.clobbered` is deleted by `status submit` when it is identical, and by `status clobbered --done` otherwise. The binaries create nothing else that is durable.
- **ARCH-MOCK:** covered by the integration-point table. `git` is not called directly, because slot branches come from `sdlc fleet inventory`.

---

## M1 — tl-worker

**Files:** create `cmd/tl-worker/{main.go,help.md,main_test.go}` and `construct/local/tl-worker/SKILL.md`; modify `AGENTS.base.md` §15, `Makefile` (`tools`, `tl-worker-build`) and `atlas/` (a new `atlas/workflow/tl-skills.md`, linked from `atlas/index.md`).

- [x] Inventory every worker-facing rule in the ops draft ("Running batches", "Staging", "The project directory") and in recent TL dispatches (the Logs of the ops `teams/*/projects/*/project.md` files). Record the sorted table in `## Log`, with one row per rule and its destination:
  - **§15** if it applies to any Couch-fleet agent;
  - **couch skill** if it is messaging mechanics, already in `couch --skill`: slot labels, the bare call form, `--message-file`, and receipt rungs;
  - **tl-worker** if it applies only to a TL-run fleet.

  A rule already in §15 or `couch --skill` is cited, not restated. If a rule belongs in `couch --skill` but is missing there, file it as a pair issue and record it.
- [x] Write the failing test `TestWorkerHelpCarriesEveryRule`. It checks `buildRoot()`'s `--help` output for one distinctive phrase per tl-worker rule from the inventory, as a table. Each row asserts its own phrase, so no row can pass vacuously.
- [x] Write `TestWorkerHelpDoesNotRestateSection15`: the help names §15 and `couch --skill`, and contains none of §15's distinctive phrasing ("whole Bash call", "substitution").
- [x] Write `cmd/tl-worker/help.md` and `main.go` (in the shape of doc-review: `//go:embed help.md`, cobra `Long`, no subcommands). The help must cover these rules:
  - no permission is needed for targeted tests, a full suite before a close, or headless bounded browser runs; ask for a visible window, `:0`, machine changes or scope;
  - real hashes, from `git rev-parse`;
  - a cited known failure never ends the suite early (`make -k`; report every other failure);
  - retry an uncertain hub send per its receipt;
  - stage with `sdlc stage --build`, then clear it;
  - on SHIP, land yourself (`sdlc merge`), report with the PR, and stop;
  - never edit project files (the TL records);
  - file issues from the slot that should own them, or unclaim;
  - start fresh;
  - a feature driving the operator's real environment gets a live check.

  The final set is whatever the inventory step settled.
- [x] Write the stub `construct/local/tl-worker/SKILL.md`. Its frontmatter description triggers on work dispatched by a Couch hub. The body, a few lines in the shape of the couch stub, says: run `tl-worker --help` and follow it; if `tl-worker` is not installed, stop and tell the operator.
- [x] Add `TestSkillStubsPointAtBinaries` (in `cmd/tl-worker`): it reads `../../construct/local/tl-worker/SKILL.md` and asserts that it names `tl-worker --help` and carries no rule text (under 30 lines).
- [x] Add one bullet to `AGENTS.base.md` §15: "At the start of each dispatched item, run `tl-worker --help` and follow it; the hub runs `tl-lead --help` instead. If the binary is not installed, tell the hub and continue under these rules." This is the deterministic load: `couch --prepare` reloads the context, and the constitution is always loaded. No Couch change is needed; record that in `## Log`. Run `weave refresh` so the composed `AGENTS.md`/`CLAUDE.md` pick it up, and add a test asserting that §15 names both binaries.
- [x] Add `tl-worker` to the Makefile's `tools`. Confirm that `construct/dev-aliases.sh --list` lists it (ariadne owns `cmd/tl-worker`, so every environment that sources dev-aliases, downstream repos included, gets it), and that weave compiles `.claude/skills/xx-tl-worker`.
- [x] Atlas: `atlas/workflow/tl-skills.md` covers the two binaries, where their prose lives, the stubs and the §15 load. Link it from `atlas/index.md`.
- [ ] Run the targeted tests (`go test ./cmd/tl-worker/... ./cmd/weave/...`), then `sdlc milestone-close --milestone M1`.

## M2 — tl-lead prose

**Files:** create `cmd/tl-lead/{main.go,help.md,main_test.go}` and `construct/local/tl-lead/SKILL.md`; modify `Makefile` and `atlas/workflow/tl-skills.md`.

- [x] Re-read the then-current ops draft. List its sections and map each to the lead prose or to "dropped" (an incident narrative, history or a "kink N" reference), and record the map in `## Log`.
- [x] Write the failing `TestLeadHelpCoversPlaybook`, with one row per required topic:
  - the eight-step playbook;
  - talking to the operator (status file sections; `couch --notify` if pair#540 has landed, otherwise its sentence is omitted);
  - dispatch judgment;
  - product stewardship and the operator's two design rules;
  - the smoke loop via `sdlc stage`;
  - attention and load;
  - the project lifecycle;
  - grooming trains;
  - reading a slot's state;
  - fleet traps that are still true.
- [x] Write `TestLeadHelpHasNoNarratives`: there are no `kink \d+`, no dates such as `10-1\d` in parentheses, and no bare `#\d+` without a repo prefix (`(^|[^\w])#\d{3}` refuses).
- [x] Write `TestLeadHelpDoesNotRestateWorker`: the dispatch section cites `tl-worker --help` for the worker's rules and doesn't repeat them. It is checked against the same phrase table as M1, which is shared through a small exported `cmd/tl-worker` test-data file read by path.
- [x] Write `cmd/tl-lead/help.md` and `main.go`, rewriting the draft as current rules. A TL dispatch then carries only the issue and TL-only context, because the worker's half is in `tl-worker`.
- [x] Add the stub `construct/local/tl-lead/SKILL.md` (the description triggers on acting as the TL or hub), extend `TestSkillStubsPointAtBinaries` to it, and add it to the Makefile.
- [ ] Update the atlas, run the targeted tests, then `sdlc milestone-close --milestone M2`.

## M3 — Go subcommands; retire the Python

**Files:**
- Create `pkg/markdoc/*`, `pkg/trains/*` and `cmd/tl-lead/{status.go,trains.go,sessionsync.go,sdlc.go,bar.go,render.go,threads.go,seen.go,markers.go}` with their tests and `testdata/`.
- Modify `cmd/sdlc/internal/issue/{fence,section,plan}.go` (into forwarding aliases), `cmd/sdlc/internal/project/trains.go` (the same) and `cmd/tl-lead/help.md` (whose procedure names the subcommands).

- [x] **Extract markdoc.** Move the fence, section and plan functions and their tests to `pkg/markdoc`. Replace the originals with aliases. `go build ./... && go test ./cmd/sdlc/internal/issue/... ./pkg/markdoc/...` passes unchanged. Grep proves no logic remains behind in `internal/issue/{fence,section,plan}.go`.
- [x] **Extract trains** the same way, into `pkg/trains`. `go test ./cmd/sdlc/internal/project/... ./pkg/trains/...` passes.
- [x] Write the `PlanBar` tests first: Mx counted when present, tasks otherwise, `no plan yet`, and fenced checkboxes ignored. Then implement it.
- [x] Write the `RenderTrains`/`RenderOffTrain` golden tests against a crafted `testdata/project.md`, shaped like the ops project files but holding no operator content. They cover each glyph, a cross-repo car, a failed car read (`?`), and an off-train slot list plus `_None._`. Then implement them.
- [x] Write the `fakeSessionSync` model and its own test, which pins the stale, lock, missing-hash and fold-not-owed refusals from couch's help.
- [x] Write the `status` tests against the fake, then implement them. Strategy: table-driven over each verb (`section`, `thread`, `owed`, `seen`, `submit`, `clobbered`, `trains`); `fakeSessionSync` injects a stale or held-lock outcome at a chosen call number, and each row asserts the exit code, the fake's write count and the folded IDs. The rows that carry the risk: `TestOperatorLockRefuses` (exit 3, zero writes); stale-then-success and four-stales refusal; `--fold-owed` folds only seen ∩ owed; the seen-set prunes to owed.

  The legacy `holder: agent` lock deletion is ported only if parley.nvim#315 is still open when M3 starts; record the check in `## Log`.
- [x] Write the `SDLC` seam (exec and decode) with `fakeSDLC` fixtures, and `tl-lead trains PROJECT` (stdout only, no write) on top of it, using bounded concurrency of 8.
- [x] Write the live conformance test (opt-in), run it once here against the real `couch`, and log the result.
- [x] Update the procedure in `help.md`, which names `tl-lead status …` and `tl-lead trains` and no Python, and assert it in `TestLeadHelpNamesSubcommands`.
- [x] Measure: run `tl-lead trains` against the live ops project, compare its output with `render-trains.py`'s, and log the timing and any diff.
- [ ] Run the full `make test` in the background before the close, then `sdlc close`. Hand over to the TL: the ops draft becomes a pointer, `scripts/*.py` are deleted, and the TL switches its calls to `tl-lead`. This is the TL's ops change, named in the report.

## Revisions

- 2026-10-13 — how the build departed from the plan, recorded while ticking the rows:
  - **Rule-table loader.** It lives in `pkg/tlrules`, not inside a `cmd/` directory, because the settings test treats every `cmd/` directory as a binary that needs a sandbox exemption. `rules.tsv` gained a scope column (`own`/`cited`) so the lead may name the `sdlc stage` commands, which are the TL's half of the same step.
  - **SDLC seam.** It gained `RepoRoot` (`sdlc workspace <repo> --json`), so each car resolves its repository in 0.2 s while the 11.8 s `sdlc fleet inventory` runs alongside, needed only for off-train slots. This took a measured render from 21.8 s to 13.7 s.
  - **Render goldens.** They are inline in `trains_test.go` (`TestRenderTrainsGolden`, `TestRenderOffTrain`, `TestGatherTrains`) rather than in a `testdata/project.md`. The sdlc fixtures, `testdata/issue-show.json` and `testdata/fleet-inventory.json`, are captured from real output and scrubbed of machine and operator identifiers.
  - **Status tests.** These are one test per verb plus a fault-injection table for the write loop (`TestWriteSectionRetries`). The named risk rows exist under these names: `TestOperatorLockRefuses`, `TestFoldOwedFoldsOnlySeen`, `TestListingPrunesSeenSet` and `TestSeenSetMalformedIsAnError`.
  - **Legacy lock.** The `holder: agent` lock deletion is not ported, because parley.nvim#315 is done.
  - **Settings exemption.** The `tl-worker *` / `tl-lead *` sandbox exemptions are ariadne#354, a local-gates change built from main by another slot. M1's milestone-close waits for it to land.

- 2026-10-14 — the M1 boundary review (BR-1 to BR-5), answered in the same round:
  - **BR-1, the stale tool lists.** Every restatement of "the tools ariadne builds" now names `tl-worker` and `tl-lead`: `README.md`, `atlas/workflow/base-layer.md`, `atlas/workflow/setup-and-replication.md` and `atlas/workflow/sandbox.md` (its exempt-binaries list, now "one per `cmd/` directory").
  - **BR-2, "whole Bash call".** `cmd/tl-worker/help.md` no longer restates §15's phrasing ("the exact form a couch or sdlc call must take"), and `TestWorkerHelpDoesNotRestateSection15` now forbids "whole Bash call" as planned.
  - **BR-3, the live fold check.** `TestSessionSyncConformance` now folds an ID that isn't owed against the real couch and asserts the refusal changes nothing. The dummy owner has no owed events.
  - **BR-4, the delivery gap.** Couch's `delivery-gap` reply is the typed `ErrDeliveryGap` (`decodeReply`). A `--fold-owed` write refuses with the next action ("check recent messages, then write without --fold-owed"), and a plain write still lands (`TestDeliveryGapRefusesFoldOnly`). `--accept-gap` is not passed through: folding needs the owed set, which a gap leaves unknown. The fake for `SDLC` is `fakeSDLC`, in `cmd/tl-lead/trains_test.go`; there is no `sdlc_fake_test.go`. `cmd/tl-lead/fake_test.go` pins `fakeSessionSync`'s refusals.
  - **BR-5, a fleet failure.** A failed `sdlc fleet inventory` no longer aborts the render: the Trains block renders, and the off-train block reads `? off-train slots unread: <reason>` (`TestGatherTrainsFleetFailureDegrades`).
