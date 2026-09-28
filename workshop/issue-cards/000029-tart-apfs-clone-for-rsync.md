---
id: '000029'
status: done
created: 2026-05-21
updated: 2026-05-27
estimate_hours: 3
actual_hours: 3
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
