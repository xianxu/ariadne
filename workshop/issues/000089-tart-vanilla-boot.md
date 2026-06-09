---
id: 000089
status: working
deps: []
github_issue:
created: 2026-06-09
updated: 2026-06-09
estimate_hours: 0.5
---

# VANILLA=1 make tart: boot pristine base image (skip mount/clone/setup, keep ssh)

## Problem

`make tart` always provisions the VM: builds a workspace APFS clone,
mounts it, and runs `tart-vm-setup.sh` (oh-my-zsh, extension rc,
`~/workspace`+`~/repo` wiring, per-repo vm-hooks). There's no way to boot
the *pristine* base image to reproduce a bug against an unmodified macOS,
or to test the setup script itself from a clean slate. `RUN_FLAGS=` drops
only the mount — setup still runs.

## Spec

Add `VANILLA=1 make tart` (and `tart-gui`) that boots `$(TART_VM)` with
**no customization except the passwordless-ssh pubkey install**:
- skip the workspace clone (`_tart_prepare_clone`),
- skip the mount (force `RUN_FLAGS` empty — a `--dir` at the un-built
  clone dir would otherwise fail the boot),
- skip the `tart-vm-setup.sh` push/run in `_tart_boot_and_ssh`,
- KEEP the pubkey install + the interactive ssh.

Design decision (operator-chosen): **reuse `$(TART_VM)`**, don't spin a
dedicated/fresh VM. So a VM previously customized by a plain `make tart`
keeps that on-disk state; `make tart-clean && VANILLA=1 make tart`
guarantees a from-scratch base (tart-clean deletes the disk → next boot
re-clones fresh `$(TART_BASE)` via `_tart_ensure_vm`). Presence-based flag
(any non-empty value = on), matching the make `V=1` idiom.

Base-layer file (`construct/base.manifest` symlink) → propagates
downstream.

## Done when

- `VANILLA=1 make -n tart` shows: no clone-prep, `tart run` without
  `--dir`, NO scp/`tart-vm-setup.sh`, but pubkey-install + interactive
  ssh still present.
- Plain `make -n tart` is unchanged (clone-prep + `--dir` + setup all
  present).
- Same for `tart-gui` (no `--dir`, no setup, VNC display flag kept).

## Plan

- [x] Add `VANILLA` var + `override RUN_FLAGS :=` when set; add a
      `_tart_clone_step` wrapper that no-ops the clone prep in vanilla.
- [x] Gate the `tart-vm-setup.sh` block in `_tart_boot_and_ssh` on
      `[ -z "$(VANILLA)" ]`; keep pubkey install.
- [x] Swap `$(call _tart_prepare_clone)` → `$(_tart_clone_step)` in
      `tart` + `tart-gui`; gate tart-gui's "APFS-cloned share" echoes.
- [x] Doc: header comment + `help-tart` line.
- [x] Verify via `make -n` diff (vanilla vs plain) for both targets.

## Log

### 2026-06-09
- Implemented in `.tart/Makefile`: `VANILLA` var (presence-based),
  `override RUN_FLAGS :=` + `_tart_clone_step` wrapper (parse-time gates),
  `[ -z "$(VANILLA)" ]` setup-skip (shell-time gate). Both gates kept
  presence-based so they never diverge (plan-quality judge's watch-item).
- Verified deterministically via `make -n` (recipe-level proof — the change
  is purely which commands the recipe emits; a live boot tests tart, not
  this change, and a prior live-boot attempt was denied):
  - PLAIN `make -n tart`: `tart run --no-graphics --dir=workspace:…clone`,
    "Preparing APFS clone" runs, setup gate `[ -z "" ]` (TRUE → setup
    runs), pubkey + `ssh -t … exec $SHELL -l` present.
  - VANILLA `make -n tart`: `tart run --no-graphics  ariadne-test` (NO
    `--dir`), clone-prep → one-line VANILLA notice, setup gate `[ -z "1" ]`
    (FALSE → setup skipped), pubkey + interactive ssh STILL present.
  - `tart-gui` PLAIN: `tart run --vnc --dir=…`, "Booting with VNC display
    + APFS-cloned share". VANILLA: `tart run --vnc  ariadne-test` (`--vnc`
    kept, no `--dir`), echo drops the share suffix, setup `[ -z "1" ]`.
  - `make -n help-tart` parses clean (no Makefile syntax errors).
- Atlas: `atlas/workflow/base-layer.md:41` documents the `make tart`
  family at a level that doesn't enumerate flags — added VANILLA there so
  the new no-customization boot is discoverable.
