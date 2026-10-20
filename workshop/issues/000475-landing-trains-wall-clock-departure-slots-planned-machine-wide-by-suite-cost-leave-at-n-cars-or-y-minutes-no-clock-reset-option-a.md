---
id: 000475
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '31bef3815adf52740bbcab93c11847dffa12c9cc' # card fields mirrored from issue-cards; edit via sdlc
---

# landing trains: wall-clock departure slots planned machine-wide by suite cost; leave at N cars or Y minutes; no clock reset (option A)

## Problem

Landing trains (#449) depart per repo on independent clocks: a train leaves `train_interval_minutes` after the previous one,
takes up to `train_max_cars`. Nothing coordinates repos on one machine, so heavy suites collide at random in the 2-slot suite
budget (ariadne's suite takes both slots; measured 7-day medians: ariadne 9.5 min, pair 12.5, parley 2.7; slot waits 11–19 min).
A full train waits for the clock even when N cars are ready, and a lone or granted landing resets the train clock
(`landingtrain.go` ~l.661, `NextDue = now + interval`), pushing the next train a whole interval back. Operator discussion and
decision (option A, 10-19): ops `teams/devinfra/tl-status-xian-xu.side/2026-10-19-11-landing-schedule.md`.

## Spec

- **Wall-clock departure slots, planned machine-wide.** Each train-enabled repo departs at `hh:00 + offset (mod interval)`.
  Offsets come from one machine-wide plan file next to the suite lock (the steward's to change), computed from measured suite
  cost (`sdlc test-stats` p50/p90 and suite weight) so heavy suites don't overlap: e.g. ariadne :00 (weight 2), pair and parley
  :15 together, the rest free for branch/affected suites. A repo without a plan entry keeps today's clock (back-compatible,
  ARCH-EVOLVE: readers tolerate unknown fields). `sdlc landing plan` shows and recomputes it.
- **N-or-Y departure:** a train also leaves early when `train_max_cars` cars are ready; and a car's wait is bounded: a train
  leaves when the oldest ready car has waited `train_max_wait_minutes`, whichever comes first (the slot then shifts to the next
  free offset).
- **Fix the clock reset:** lone and granted landings don't move the next train's departure.
- Measured outcome: slot-wait p50/p90 and ready → landed before/after (test-stats).

## Done when

- Fixtures (fake clock): two repos with offsets never start train suites within each other's expected window; a full train
  leaves before its slot; the oldest-car bound fires; a granted landing leaves the next departure unchanged.
- `sdlc landing plan` prints the machine plan with each repo's offset and its cost basis.

## Plan

- [ ]

## Log

### 2026-10-19
