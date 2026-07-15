---
id: 000028
status: done
deps: [000026, 000027]
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

## Insight

The existing `tart-gui` target already cold-boots when needed
(VNC vs headless is a cold-boot-only setting):

```make
if tart list ... | grep -qx running; then
    echo "==> Stopping VM to switch to VNC mode (cold-boot only setting)..."
    tart stop $(TART_VM)
fi
$(call _tart_wait_for_stop)
```

The plain `tart` target should follow the same pattern, with a
different reason: mount-cache invalidation rather than display
mode. We give up ~30s of latency per invocation; we get
correctness (operator can trust that `~/repo` on the VM
matches the host's current worktree, including just-committed
changes).

Per operator (2026-05-19): "it either works or not work. let's
make [cold-boot] the default. we have to tolerate some latency,
if correctness can't be guaranteed."

## Done when

- `make tart` always cold-boots: stops the VM first if running,
  then proceeds with the existing boot+ssh flow. Same pattern
  `tart-gui` already uses.
- Header comment block and `help-tart` mention the cold-boot
  default and the reason (VirtIO-FS cache invalidation).
- No new env-var escape hatch — operators who want to ssh into a
  running VM without restart can just `ssh admin@$(tart ip
  <name>)` directly. `make tart` is the "boot/refresh and
  connect" verb; if the operator doesn't want refresh, they
  don't run it.

## Spec

```make
tart:
    $(call _tart_check_others)
    $(call _tart_ensure_vm)
    @# Cold-boot for fresh VirtIO-FS mount. macOS's host-guest
    @# directory share caches dentries/inodes guest-side and
    @# diverges from the host over time — git's atomic-rename
    @# writes are especially affected. ~30s latency for
    @# correctness (see ariadne#28).
    @if tart list 2>/dev/null | awk -v vm=$(TART_VM) '$$2==vm {print $$NF}' | grep -qx running; then \
        echo "==> Cold-restarting VM to refresh host mount (VirtIO-FS cache invalidation)..."; \
        tart stop $(TART_VM); \
    fi
    $(call _tart_wait_for_stop)
    @if [ -n "$(RUN_FLAGS)" ]; then \
        echo "==> Mount: /Volumes/My Shared Files/$(REPO_NAME) (read-only; rsync source)"; \
        echo "    Mirror: ~/repo in the VM (writable, always fresh, wipes on each boot)"; \
    fi
    $(call _tart_boot_and_ssh)
```

(`tart-gui` already has the equivalent stop step.)

## Plan

- [x] M1: Added cold-boot stanza to `tart` recipe (same pattern
      `tart-gui` already uses for the VNC-mode case).
- [x] M2: Updated header comment block and `help-tart` to name
      cold-boot as the default behavior.
- [x] M3: `make refresh` in nous picked up the new Makefile.
      Other consumers (charon, pair, parley.nvim) refresh on
      their own next session — additive change, safe to lag.
- [ ] M4: Operator-side empirical verification — the originating
      bug (nous#26 M4 patch invisible on `nous-ying` despite
      multiple `make tart TART_SUFFIX=ying` invocations) was
      what motivated this change. Re-running with the cold-boot
      default should make the patch visible immediately. Will
      mark verified once the operator confirms the next cycle.

## Out of scope

- Detecting mount staleness without cold-boot (could compare
  host's HEAD against the mount's HEAD via git-in-the-mount).
  More precise but more code; cold-boot is the simple
  correctness-first answer.
- Filing the issue upstream against tart or Apple. The behavior
  is well-known in the macOS Virtualization framework community;
  no need to add to the pile. We work around at our layer.

## Log
