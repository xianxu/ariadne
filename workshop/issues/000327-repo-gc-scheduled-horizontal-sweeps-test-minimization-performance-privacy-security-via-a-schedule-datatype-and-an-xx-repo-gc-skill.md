---
id: 000327
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: 'e50029de7042b29f06aee06386b1b3e0b6ded547' # card fields mirrored from issue-cards; edit via sdlc
---

# Repo GC: scheduled horizontal sweeps (test minimization, performance, privacy, security) via a schedule datatype and an xx-repo-gc skill

## Problem

Some work affects the whole codebase, belongs to no feature, and decays when nobody tends it: test minimization
(tests that catch no fault another test doesn't), test performance, privacy, security. Agents can now do a
cross-repo sweep in a day, so a periodic "repo GC" is affordable, even a stop-the-world one (e.g. the last day of
the month: agents make cross-repo updates, everyone rebases afterward). ariadne has no notion of a scheduled task.
Raised at ops' batch-2 recap of ariadne-robustness-1 (10-10).

## Spec

Keep scheduling deterministic and daemon-free; the sweeps themselves are ordinary issues.

- **A `schedule` datatype** (`workshop/schedules/<slug>.md`): cadence (e.g. monthly, or "last day of month"),
  scope (repos), the skill to run, and the issue template a run files. Revisions as for other plan artifacts.
- **`sdlc schedule due [--json]`**: read-only; lists schedules whose next run has come, computed from the cadence
  and the last run recorded in the ledger. No daemon in ariadne: whoever polls acts on it, whether that's a TL
  session start, a Couch tick or a launchd job. That keeps "when" (the datatype) apart from "who runs it".
- **`sdlc schedule run <slug>`**: files the run's issues (one per concern per repo, in a GC train) and records
  the run. A stop-the-world run also announces the freeze window to every slot.
- **The `xx-repo-gc` skill**: how each concern produces a report first, then a cleanup issue.
  - Test minimization: tests that catch no unique mutant or cover no unique line, merged or deleted, at equal coverage.
  - Test performance: time per package against a soft budget, as an alarm that triggers minimization, not a CI gate.
  - Flakes: quarantined by name with an owning issue.
  - Privacy and security: the respective checklists.

Open questions: the cadence grammar; whether the stop-the-world freeze is Couch's job; where the per-repo
reports live (ops?).

## Done when

- [ ] The `schedule` datatype and vocabulary exist; `sdlc schedule due` lists a due schedule and not a fresh one (test).
- [ ] `sdlc schedule run` files the run's issues and records the run, so `due` no longer lists it (test).
- [ ] `xx-repo-gc` exists, with the test-minimization report specified well enough to run on one repo.
- [ ] A first monthly run is scheduled for ariadne and pair.

## Plan

- [ ]

## Log

### 2026-10-10
