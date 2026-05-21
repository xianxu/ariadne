---
id: 000029
status: working
deps: [000028]
created: 2026-05-21
updated: 2026-05-21
estimate_hours: 3
---

# tart: replace rsync with APFS clonefile (O(1) init, write-isolated)

## Problem

`tart-vm-setup.sh` currently rsync-mirrors the host share into
`~/repo` inside the VM at every boot. For a ~100MB repo (nous
today) that's 5–10 seconds. The cost is linear in repo size —
as soon as the repo grows to several hundred MB (or anyone
checks in large test fixtures), the boot delay becomes painful.

The rsync exists for two reasons:

  1. Cache isolation: `~/repo` is VM-local, so reads bypass the
     VirtIO-FS cache layer that misbehaves on host edits. This
     is the dominant value — see ariadne#28's cold-boot
     workaround for the underlying staleness issue.
  2. Write isolation: VM builds (`make nous-build`, etc.) write
     to `~/repo/...`, not back to the host. Host stays clean.

APFS's `cp -cR` (`clonefile(2)` under the hood) gives us both
properties cheaper: a metadata-only copy that shares data blocks
COW-style with the source. Initial cost is O(1) regardless of
repo size; physical disk usage is bounded by the VM's write
divergence (which is just build outputs + edits during the
session).

Verified empirically (2026-05-21): `cp -cR` of nous's 117MB tree
takes 0.57s. `df` before+after shows zero physical disk delta —
APFS COW working as expected.

## Insight

The Mac host is always APFS (since Mojave 2018). VirtIO-FS will
continue to misbehave on host-side renames (Apple's bug, see
ariadne#28's discussion), so we can't drop cache-coherence
workarounds entirely — cold-boot remains the default. But we
CAN drop the rsync that's only there to populate a VM-local
filesystem, by exposing an APFS clone of the host repo as the
VM's writable workspace directly.

Structure shift:

  Today:
    host /Users/.../nous  ──[mount :ro]──→  /Volumes/Shared/nous (VM read)
                                            ──[rsync at boot]──→ ~/repo (VM writable)

  Proposed:
    host /Users/.../nous  ──[cp -cR]──→  /tmp/$(TART_VM)-clone
                                        ──[mount :rw]──→  /Volumes/Shared/nous (VM read+write)
                                                          ──[symlink]──→ ~/repo

## Done when

- `.tart/Makefile` does `cp -cR $(CURDIR) /tmp/$(TART_VM)-clone`
  before tart-run (skipping if clonefile fails for any reason —
  fall back to plain `cp -R` as a safety net).
- The mount target is the clone, not `$(CURDIR)`. Writable
  (`--dir=...:rw`), not read-only.
- Build artifacts excluded from the clone, matching the existing
  rsync excludes:
  - `bin/`, `cmd/*/bin/` (built binaries — they carry codesign +
    provenance xattrs that block VM-side writes, the original
    nous#11 / pre-rsync EPERM failure).
  - `.nous-mode`, `.nous-plugins` (ariadne local state).
- `.tart/scripts/tart-vm-setup.sh` no longer rsyncs. Instead,
  symlinks `~/repo` → `/Volumes/My Shared Files/$(REPO_NAME)`.
  VM reads and writes pass through to the host clone via the
  writable VirtIO-FS share.
- `tart-stop` / `tart-clean` remove `/tmp/$(TART_VM)-clone` for
  the targeted VM. Idempotent.
- A "GC orphan clones" step runs at tart-run: remove any
  `/tmp/*-clone` older than 7 days. Catches the force-kill case
  where the cleanup didn't fire.
- Help-tart text updates to reflect the new behavior + the
  trade-off note ("VM writes still don't reach host source; via
  APFS COW into the clone instead of via a separate
  filesystem").

## Open questions

- **clonefile + VirtIO-FS interaction.** I assume the cloned
  tree behaves identically to the source as far as VirtIO-FS is
  concerned (they're both ordinary APFS dir trees). Worth
  verifying empirically — boot tart against a clone, do a build
  inside the VM, confirm normal behavior + the VM cache
  staleness story is the same as today (cold-boot still needed).
- **Symlink semantics inside the clone.** nous has some
  in-repo symlinks (e.g., `bin/nous` → `cmd/nous/bin/nous`,
  construct/ vendor symlinks). `cp -cR` clones symlinks as
  symlinks (verified). VM-side resolution should work. Quick
  check inside the VM after boot.

## Out of scope

- VirtIO-FS cache coherence proper. Still requires cold-boot
  (ariadne#28). The clonefile substitution is orthogonal —
  same cache layer, same staleness, same workaround. We're
  reducing the rsync cost only.
- Disk-image-based VM with full block-device caching (the path
  B sketch from the operator discussion 2026-05-21). Bigger
  effort, deferred.

## Plan

- [x] M1: `_tart_prepare_clone` macro in `.tart/Makefile` —
      cp -cR + targeted rm of bin/, cmd/*/bin/, .nous-mode,
      .nous-plugins. Fallback to plain `cp -R` if clonefile
      fails (non-APFS volume edge case).
- [x] M2: `TART_CLONE_DIR` knob (default `/tmp/$(TART_VM)-clone`);
      `RUN_FLAGS` defaults to mount the clone writable.
- [x] M3: VM-side: `tart-vm-setup.sh` drops the rsync, symlinks
      `~/repo` → `/Volumes/My Shared Files/<repo>` (the writable
      clone-backed share). Old rsync-era `~/repo` directories
      get wiped on first boot post-rename.
- [x] M4: Cleanup wired into `tart-stop` (single-VM clone
      removal) and `tart-clean` (same). Added
      `_tart_gc_orphan_clones` step at every tart-run boot:
      removes any `/tmp/*-clone` older than 7 days as the
      safety net for the force-kill / unclean-exit case.
- [x] M5: Updated `help-tart`, the header comment block in
      `.tart/Makefile`, and the historical-evolution comment in
      `tart-vm-setup.sh` (layout #4 added).
- [ ] M6: Empirical verification — operator-side, needs running
      VM. Steps: `make tart`, observe ~150ms prepare-clone step
      vs the old ~5–10s rsync; `nous-build` inside VM, verify
      build outputs land in `/tmp/<vm>-clone/bin/`, host's
      `cmd/nous/bin/nous` mtime unchanged; `make tart-stop`,
      verify `/tmp/<vm>-clone` removed.

## Log
