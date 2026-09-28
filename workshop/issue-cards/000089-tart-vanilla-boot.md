---
id: '000089'
status: done
created: 2026-06-09
updated: 2026-06-09
estimate_hours: 0.5
actual_hours: 0.4
---

# VANILLA=1 make tart: boot pristine base image (skip mount/clone/setup, keep ssh)

## Problem

`make tart` always provisions the VM: builds a workspace APFS clone,
mounts it, and runs `tart-vm-setup.sh` (oh-my-zsh, extension rc,
`~/workspace`+`~/repo` wiring, per-repo vm-hooks). There's no way to boot
the *pristine* base image to reproduce a bug against an unmodified macOS,
or to test the setup script itself from a clean slate. `RUN_FLAGS=` drops
only the mount — setup still runs.
