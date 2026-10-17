---
id: 000427
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'f4d5f4aab0e3ab0b6567065766478cb66acbef79' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T08:56:11-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "3115a431", done: "832fc4bd"}
---

# weave's dependency step runs brew bundle install from slots (start-plan auto-refresh), changing the machine unattended

## Problem

A start-plan auto-refresh in pair:6 invoked weave dependency provisioning, which ran brew bundle install and auto-updated Homebrew metadata. Automated slot work must not change machine packages or metadata without the operator.

## Spec

Weave's shared dependency step checks packages instead of installing when its existing Git-proved setup policy identifies a numbered slot (including its private dependency checkouts), when COUCH_TREE identifies an active Couch context, or when an SDLC caller supplies the check-only subprocess context. Both compile/refresh and the dependencies subcommand derive this one policy. SDLC catch-up and propagate-base pass that context without mutating their parent environment.

Checks use brew bundle check --no-upgrade --verbose --file=Brewfile, streaming the missing-package diagnostics. A failed check stops before tools/composition and names the layer plus one shell-quoted absolute Brewfile install command for the operator. A failed probe is reported as failed, never as proven missing or success. Layers without Brewfiles still do nothing; dry runs print the selected operation without running it. Outside those automated contexts, hand runs keep the existing install --no-upgrade behavior. Every weave-owned brew invocation overrides HOMEBREW_NO_AUTO_UPDATE=1, including hand installs. No installer fallback, unsandboxed retry or new approval prompt.

Keep the fix in the existing startup/runner seam (ARCH-DRY/ARCH-PURE); use a stateful fake brew with an installed-package directory, call log and metadata-update sentinel (ARCH-MOCK). Existing discovery proves slot identity rather than adding a path heuristic. The policy creates no durable runtime records or workers (ARCH-FUNERAL); effects stay synchronous in existing bounded callers, with no new concurrency or workload (ARCH-CONSTRAINTS/ARCH-ORDER). Context only restricts effects, never grants installation (ARCH-SECURE); cover both SDLC subprocess paths, both weave entry points, and positive hand controls (ARCH-PURPOSE).

Alternatives: globally checking would remove existing operator provisioning; guarding only start-plan would leave slot compile/dependencies and propagate-base installing. Shared check selection plus restricted subprocess context covers the actual callers with no generic machine-permission framework.

2026-10-17: in pair:6, `sdlc start-plan`'s automatic refresh ran `weave compile`. Its dependency step ran `brew bundle install --no-upgrade --file=Brewfile` (see `cmd/weave/internal/startup/dependencies.go` and the README section on `weave dependencies`). Homebrew auto-updated its own and its taps' metadata before pair:6 interrupted the run, ahead of any package install.

Installing or updating machine-wide packages is a machine-level change, and the operator approves those. An agent's routine workflow step (start-plan, or any compile a slot triggers) must not make one unattended. `--no-upgrade` doesn't stop Homebrew's auto-update or new installs.

Wanted:
- In a Couch slot (`COUCH_TREE` set), or whenever weave runs non-interactively from an sdlc verb, the dependency step only checks: a dry-run comparison of the Brewfiles against installed packages. If anything is missing it fails with a clear message naming the packages and the command for the operator to run. It never installs.
- Any brew invocation weave does make sets `HOMEBREW_NO_AUTO_UPDATE=1`, so a check never updates metadata.
- An operator running `weave dependencies` or `weave compile` by hand in a terminal keeps today's behaviour.

## Done when

- Slot compile/refresh and dependencies only check Brewfiles, list missing packages and name the operator's quoted install command; package/metadata state stays unchanged and failed checks prevent later tools/composition.
- SDLC refresh and propagation enforce checks even outside Couch/numbered paths. Check-only context does not leak into the parent.
- Every dependency brew invocation has HOMEBREW_NO_AUTO_UPDATE=1; hand calls outside restricted contexts still install missing packages and remain repeatable.
- No-Brewfile and dry-run controls remain effect-free; focused tests and recorded full suite pass; help, README and atlas describe the policy.

- A test runs the dependency step with `COUCH_TREE` set and a fake brew: it never sees `bundle install`, only a check, and a missing package fails with the package name and the operator command.
- Brew invocations carry `HOMEBREW_NO_AUTO_UPDATE=1`; a test asserts the environment.
- A hand run outside a slot still installs, as today.

## Plan

- [ ] Reproduce the unattended install with stateful fake-brew regression tests; include numbered setup, Couch and SDLC contexts, missing/satisfied/failure, dry-run/no-bundle and hand controls.
- [ ] Select check-only mode once for both dependency callers; restrict both SDLC weave subprocess paths; force no-auto-update on every dependency brew call and print safe operator recovery.
- [ ] Update command help, README and relevant atlas provisioning passages; run focused weave/SDLC caller tests.
- [ ] Run the required recorded branch suite. Close the tested branch, then let merge integrate main under its turn and test integrated code as required; report landing to the hub.

- [ ]

## Log

### 2026-10-17

- Claimed and started from freshly fetched main 0408d048433e0073f5914787eb3f093480cfbf64. start-plan did not need auto-refresh, so no real brew was invoked. Issue initially had empty sections; the dispatch supplies the contract above.
- Root cause: startup.Dependencies unconditionally invokes install; compilePreparedSettings and buildDependencies share it. runInstalledWeaveArgs supplies no restricting environment; propagate-base has a second weave subprocess runner. prepareSetup already discovers the numbered environment and carries acquire.Policy, including private dependency and feature worktrees; reuse it.
- Homebrew's official bundle documentation confirms check --verbose reports unmet dependencies and --no-upgrade preserves installed versions: https://docs.brew.sh/Brew-Bundle-and-Brewfile . Tests will use the existing runner seam and fake binaries only; no real package install/update is authorized.

### 2026-10-17
