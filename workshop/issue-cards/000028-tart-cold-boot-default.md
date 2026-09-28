---
id: '000028'
status: done
created: 2026-05-19
updated: 2026-05-19
estimate_hours: 0.5
actual_hours: 0.2
---

# tart: cold-boot on every `make tart` (correctness > latency)

## Problem

macOS Virtualization framework's directory-sharing layer (VirtIO-FS
between host and Apple-Silicon guest) caches dentries and inodes
guest-side. The cache state diverges from the host over time,
especially around git's atomic-rename writes: when commit/checkout
writes a file via `rename(2)`, the inode changes, but the guest's
cached dentry still points to the old (now-deleted) inode. Guest
reads return stale content or even empty strings.

Symptom that bit nous#26 today (2026-05-19):

  - Host commit landed M4-fix (cmd/nous/brain_join.go gains a
    `runBrainJoinRepublish` function).
  - Operator ran `make tart TART_SUFFIX=ying` to re-mirror to VM.
  - Setup script ran rsync from `/Volumes/My Shared Files/nous/`.
  - rsync copied stale guest-cached content — the patch wasn't
    in `~/repo` on the VM despite being in the host's HEAD.
  - `nous brain join xianxu/brain-family` from the rebuilt-but-
    actually-not-patched binary failed with cobra's "unknown
    command" — exact same error as the pre-patch code path.
  - `tart stop nous-ying && make tart TART_SUFFIX=ying` (cold
    boot) immediately fixed it. Fresh mount = fresh cache = the
    patch was visible to the rsync.

The current `make tart` recipe only boots the VM if it's not
already running:

```make
if [ "$state" != "running" ]; then
    nohup tart run ... &
fi
```

For long-running VMs (hours+), the boot-skip path means we ride
the same stale mount across many `make tart` invocations.
