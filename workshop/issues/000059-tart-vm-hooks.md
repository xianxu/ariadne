---
id: 000059
status: working
deps: []
github_issue:
created: 2026-06-01
updated: 2026-06-01
estimate_hours: 1
---

# tart vm-hooks.d run-parts convention

## Problem

`make tart` boots a generic VM and runs `.tart/scripts/tart-vm-setup.sh`
(oh-my-zsh + workspace symlinks) — but there's no way for a consuming repo
to inject its own per-VM setup. Concretely, nous needs to make the headless
VM GPG-unattended for brain testing (nous#36), and a base-layer VM with no
extension point would force either a brain-specific hack in ariadne (wrong
layer) or a manual step every boot.

What's missing is a generic, opt-in extension point: a repo declares VM
setup steps; the base-layer setup runs them.

## Spec

A run-parts-style hook directory convention, owned by ariadne's
`.tart/scripts/tart-vm-setup.sh`:

- After the standard setup (oh-my-zsh, workspace + repo symlinks), look for
  the **booted repo's** hook dir at `~/workspace/$CURRENT_REPO/.tart/vm-hooks.d/`
  (the repo whose `make tart` was invoked — `$CURRENT_REPO` is already passed
  as `$1` to setup.sh).
- Run every `*.sh` in it, **lexical `LC_ALL=C` order** (so `00-`, `10-`
  prefixes sequence deterministically), each via `bash "$hook" "$CURRENT_REPO"`.
- **Every cold-boot** (no run-once marker) ⇒ hooks MUST be idempotent. This is
  part of the published contract.
- **Continue-on-error**: a failing hook prints a visible `[warn]` with its rc
  and the loop continues, so a broken hook never strands the operator out of
  the VM shell.
- Selection by `*.sh` glob + `bash`, not the executable bit — predictable and
  independent of mode bits surviving the APFS clone.

Design decisions:

- **Location `.tart/vm-hooks.d/`** — groups with the other tart machinery.
  A consumer's `.tart/` is a real dir holding symlinks to ariadne
  (`Makefile`, `scripts`); a real `vm-hooks.d/` subdir sits alongside them.
- **Additive + opt-in** — no dir → no-op. Zero behavior change for every
  existing consumer, so downstream impact (base.manifest propagation) is safe.
- **Booted-repo only** (not a walk over every peer in the clone) — running
  arbitrary peers' setup would be surprising; the booted repo is the one the
  operator chose.

Env available to hooks: `$1` / `$CURRENT_REPO` (repo name); `~/workspace`
(the mounted clone) and `~/repo` (symlink to the booted repo) already exist
by the time hooks run.

## Done when

- `tart-vm-setup.sh` runs `~/workspace/$CURRENT_REPO/.tart/vm-hooks.d/*.sh` in
  lexical order after standard setup, idempotently, continue-on-error.
- No-hook-dir consumers are byte-for-byte unaffected (no-op).
- The convention (location, lexical order, every-boot/idempotent, continue-on-
  error) is documented: a comment block in `tart-vm-setup.sh` + the `help-tart`
  / `.tart/Makefile` header, and a line in `atlas/`.
- Verified end-to-end: a throwaway `00-test.sh` hook in a booted repo runs on
  `make tart` and its output appears in the boot transcript.

## Plan

- [x] M1: add the run-parts loop to `.tart/scripts/tart-vm-setup.sh` after the
  workspace-symlink step; comment block documenting the convention.
- [x] M1: document in `.tart/Makefile` header + `help-tart`, and add an
  `atlas/` pointer (workflow/base-layer or a tart entry).
- [x] M1: verify with a throwaway hook on a real `make tart` (boot transcript
  shows `==> vm-hook: 00-…` + its output). Remove the throwaway after.

## Log

### 2026-06-01

Created as the base-layer dependency of nous#36 (headless brain testing). The
hook mechanism is generic; nous#36 supplies the first consumer
(`00-gpg-setup.sh` making the VM GPG-unattended). Single milestone — this is
atomic base-layer plumbing, not multi-stage work.

M1 implemented in 5d99f60 (loop + `.tart/Makefile`/`help-tart` docs + atlas
pointer). **E2e verified** (operator): throwaway `.tart/vm-hooks.d/00-test.sh`
echoing "hello world" fired in the boot transcript on a real `make tart`;
throwaway removed after.

**Post-milestone review (§3)** of 5d99f60 (fresh-eyes subagent) found 1
Important bug: the loop used `ls` inside `$()`, word-splitting any hook filename
with a space (spaced hook → two bogus rc=127 tokens, real hook silently never
runs). set -e / continue-on-error / no-op / ordering / quoting all verified
correct. Fixed in e0f7a47 (NUL-delimited glob via process substitution;
`nullglob`; LC_ALL=C pinned in-subshell). Re-verified under /bin/bash 3.2.57:
order 00→05→10→15, spaced hook runs as one unit, failing hook warns+continues,
empty dir is a clean no-op. No Critical.
