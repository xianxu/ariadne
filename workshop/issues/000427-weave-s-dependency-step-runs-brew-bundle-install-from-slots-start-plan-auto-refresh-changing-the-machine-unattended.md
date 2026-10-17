---
id: 000427
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'caa058e3c0846aef1fa37992449cf74042b32f4b' # card fields mirrored from issue-cards; edit via sdlc
---

# weave's dependency step runs brew bundle install from slots (start-plan auto-refresh), changing the machine unattended

## Problem

## Spec

2026-10-17: in pair:6, `sdlc start-plan`'s automatic refresh ran `weave compile`. Its dependency step ran `brew bundle install --no-upgrade --file=Brewfile` (see `cmd/weave/internal/startup/dependencies.go` and the README section on `weave dependencies`). Homebrew auto-updated its own and its taps' metadata before pair:6 interrupted the run, ahead of any package install.

Installing or updating machine-wide packages is a machine-level change, and the operator approves those. An agent's routine workflow step (start-plan, or any compile a slot triggers) must not make one unattended. `--no-upgrade` doesn't stop Homebrew's auto-update or new installs.

Wanted:
- In a Couch slot (`COUCH_TREE` set), or whenever weave runs non-interactively from an sdlc verb, the dependency step only checks: a dry-run comparison of the Brewfiles against installed packages. If anything is missing it fails with a clear message naming the packages and the command for the operator to run. It never installs.
- Any brew invocation weave does make sets `HOMEBREW_NO_AUTO_UPDATE=1`, so a check never updates metadata.
- An operator running `weave dependencies` or `weave compile` by hand in a terminal keeps today's behaviour.

## Done when

- A test runs the dependency step with `COUCH_TREE` set and a fake brew: it never sees `bundle install`, only a check, and a missing package fails with the package name and the operator command.
- Brew invocations carry `HOMEBREW_NO_AUTO_UPDATE=1`; a test asserts the environment.
- A hand run outside a slot still installs, as today.

## Plan

- [ ]

## Log

### 2026-10-17
