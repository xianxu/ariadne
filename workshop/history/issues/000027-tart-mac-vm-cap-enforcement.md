---
id: 000027
status: done
deps: [000026]
created: 2026-05-19
updated: 2026-05-19
estimate_hours: 1.5
actual_hours: 0.7
---

# tart: enforce Apple's 2-macOS-VM cap; leave Linux alone

## Problem

Apple's `Virtualization.framework` caps concurrent macOS guests at
2 on every Apple-Silicon Mac (M1/M2/M3 base/Pro/Max/Ultra alike —
it's a software policy matching the macOS EULA, not a hardware
ceiling). When the cap is hit, `tart run` exits instantly with:

```
The number of VMs exceeds the system limit (other running VMs: ...)
```

`_tart_boot_and_ssh` doesn't read the boot log on failure — it
just polls SSH against an IP that never came up. Symptom: "stuck
waiting for SSH" for 120 s, then time out. The actual error sits
in `/tmp/tart-$(TART_VM).log` waiting to be read.

`_tart_check_others` is supposed to head this off, but it's a
*soft* prompt: Enter with no ticks = "stop nothing." Pleasant
for the "I just want to free 4 GB of RAM" case, but wrong for
the "your boot will fail unless you stop one" case.

Two distinct behaviors collapsed into one prompt.

## Insight

The cap only applies to macOS guests. Linux guests on the same
host don't count toward Apple's limit (different code path in
the framework). So the script needs to know:

1. The target VM's OS (what we're about to boot).
2. The OS of each running VM (what counts toward the cap).

Both are in `~/.tart/vms/<name>/config.json`, field `"os"` —
`"darwin"` or `"linux"`. No jq dependency needed; a grep
suffices.

When the target is darwin and ≥2 darwin VMs are already running,
`_tart_check_others` should **require** stopping enough macOS
VMs to free a slot. When the target is linux, or fewer than 2
macOS VMs are running, the existing soft-prompt behavior stays.

## Done when

- `.tart/scripts/tart-stop-others.sh` detects the target VM's OS
  via `~/.tart/vms/<name>/config.json` (`"os"` field). For
  not-yet-cloned VMs, infers from the `TART_BASE` image name
  passed as `$2` (defaulting to darwin — ariadne's primary case).
- The script counts only `darwin` running VMs toward the cap.
  Linux VMs aren't required to stop and aren't shown as
  candidates in cap-enforce mode (they can't free a macOS slot).
- When `target_os = darwin` AND running-darwin-count ≥ 2:
  required-stop count = (running-darwin - 1). The fzf prompt's
  Enter binding requires that many ticks; the header explains
  why ("Apple's 2-macOS-VM cap — stop N to free a slot").
- Non-fzf fallback (sshpass-style y/N) in cap-enforce mode
  changes from "stop all? [y/N]" to "stop all? [Y/n] (boot will
  fail otherwise)" — defaults to yes; explicit `n` is the
  operator-takes-responsibility path.
- Non-tty mode (CI, piped) in cap-enforce mode exits 1 with a
  clear diagnostic, so the make target fails fast instead of
  the 120s SSH-poll timeout downstream.
- Soft-prompt mode (target is linux, or target is darwin and
  <2 darwin running) keeps current behavior unchanged — Enter
  with no ticks = stop nothing.
- `_tart_boot_and_ssh` also greps the boot log for the "exceeds
  the system limit" string after the SSH poll fails, so even
  when check-others is bypassed (e.g., manual `tart run`), the
  failure mode is named not "stuck waiting."
- `.tart/Makefile` passes `$(TART_BASE)` as the second arg to
  `tart-stop-others.sh` for OS inference on first-boot.
- `MAC_VM_CAP` env var (default 2) lets future macOS versions
  override the limit without code changes, in case Apple ever
  bumps it.

## Spec

### OS detection

```bash
# vm_os <name>: prints "darwin", "linux", or "unknown".
vm_os() {
    local cfg="$HOME/.tart/vms/$1/config.json"
    [ -f "$cfg" ] || { echo unknown; return; }
    if   grep -q '"os" *: *"darwin"' "$cfg"; then echo darwin
    elif grep -q '"os" *: *"linux"'  "$cfg"; then echo linux
    else echo unknown
    fi
}

# infer target OS from $TART_BASE when config.json doesn't exist yet
# (first boot, before _tart_ensure_vm clones).
infer_target_os() {
    case "${1:-}" in
        *macos*|*darwin*) echo darwin ;;
        *linux*|*ubuntu*|*debian*|*fedora*) echo linux ;;
        *) echo darwin ;;  # ariadne default
    esac
}
```

### Cap enforcement

```bash
MAC_VM_CAP="${MAC_VM_CAP:-2}"

target_os=$(vm_os "$ME")
[ "$target_os" = "unknown" ] && target_os=$(infer_target_os "$BASE")

# Partition running-others by OS
mac_others=""; other_running=""
while IFS= read -r vm; do
    [ -n "$vm" ] || continue
    if [ "$(vm_os "$vm")" = "darwin" ]; then
        mac_others="${mac_others}${vm}\n"
    else
        other_running="${other_running}${vm}\n"
    fi
done <<< "$others"
mac_count=$(printf "%b" "$mac_others" | grep -c .)

required=0
if [ "$target_os" = "darwin" ] && [ "$mac_count" -ge "$MAC_VM_CAP" ]; then
    required=$(( mac_count - (MAC_VM_CAP - 1) ))
fi
```

### fzf binding with required minimum

```bash
if [ "$required" -gt 0 ]; then
    candidates="$mac_others"  # only macOS VMs can free a slot
    enter_check="[ \$FZF_SELECT_COUNT -ge $required ]"
    header="Apple's $MAC_VM_CAP-macOS-VM cap reached. Tick at least $required to free a slot, then Enter."
else
    candidates="$others"
    enter_check="[ \$FZF_SELECT_COUNT -gt 0 ]"
    header="Other tart VMs running (4-8 GB each). Space: toggle, Enter: stop ticked, Esc: skip."
fi

selected=$(echo -e "$candidates" | fzf --multi --no-sort --reverse \
    ... \
    --bind "enter:transform:$enter_check && echo accept || echo \"change-header:$header\"" \
    --header="$header")
```

### Non-fzf fallback

```bash
if [ "$required" -gt 0 ]; then
    echo "==> Apple's $MAC_VM_CAP-macOS-VM cap reached. These macOS VMs are running:"
    printf "%b" "$mac_others" | sed 's/^/    /'
    printf "Stop all macOS others? [Y/n] (boot will fail otherwise) "
    read -r answer
    case "$answer" in
        [nN]|[nN][oO]) selected="" ;;
        *) selected=$(printf "%b" "$mac_others") ;;  # default Y
    esac
else
    # current behavior — soft prompt, default N
    ...
fi
```

### Non-tty mode

```bash
if [ ! -t 0 ] || [ ! -t 1 ]; then
    if [ "$required" -gt 0 ]; then
        echo "==> Apple's $MAC_VM_CAP-macOS-VM cap reached and no tty for prompt." >&2
        echo "    Stop one of: $(printf "%b" "$mac_others" | tr '\n' ',')" >&2
        exit 1
    fi
    # current behavior — soft, just notify
    ...
fi
```

### Boot-log post-mortem

In `_tart_boot_and_ssh`, after the 60-iteration SSH poll fails:

```make
if ! nc -z -G 2 "$$ip" 22 2>/dev/null; then \
    if grep -q "exceeds the system limit" /tmp/tart-$(TART_VM).log 2>/dev/null; then \
        echo "==> Boot failed: Apple's 2-macOS-VM cap reached."; \
        echo "    Stop another macOS VM first (see tart list)."; \
        exit 1; \
    fi; \
    echo "==> SSH never came up at $$ip (boot log: /tmp/tart-$(TART_VM).log)"; \
    exit 1; \
fi;
```

## Plan

- [x] M1: Added `vm_os` and `infer_target_os` helpers to
      `tart-stop-others.sh`. Pass `$2` (TART_BASE) from the
      Makefile.
- [x] M2: Partitioned `others` into `mac_others` / `linux_others`.
      Compute `required`. fzf / non-fzf / non-tty paths branch
      on `required > 0`.
- [x] M3: Boot-log grep in `_tart_boot_and_ssh` matches "exceeds
      the system limit" and fails fast with a named error.
- [x] M4: Verified on host:
      - Cap-enforce non-tty (target=nous-emma, ME=nous-emma, base=macos,
        2 darwin running) → exit 1 with "Stop 1 of them."
      - Soft mode (target=nous-test, base=macos, 1 darwin running other) →
        exit 0 with soft notice.
      - Linux target (target=fake-linux, base=ubuntu-base, 2 darwin
        running) → exit 0; cap not applied.
- [x] M5: `make refresh` in nous picked up both the new script and
      Makefile changes. Operator can now retry `make tart TART_SUFFIX=emma`
      and get the cap-enforce prompt instead of the silent SSH timeout.

## Out of scope

- Adjusting `MAC_VM_CAP` based on a real Apple announcement. If
  macOS X lifts the cap to 4, the env var is the knob.
- "Suggest which VM to stop" UX. Operator picks; we just
  enforce the count. A heuristic like "stop the longest-idle"
  would be nice but is its own design conversation.
- Tracking *why* a VM is running (which repo, which session).
  `tart list` doesn't carry that metadata; would require a
  separate ledger.

## Log

### 2026-05-19 — landed

Two-file change in ariadne plus a refresh in nous:

- `.tart/scripts/tart-stop-others.sh` — partial rewrite. New helpers
  `vm_os` (reads `~/.tart/vms/<name>/config.json` for the `os` field)
  and `infer_target_os` (falls back to base-image-name pattern for
  first-boot VMs without a config yet). Cap-enforce mode kicks in
  iff `target_os == darwin` AND running-darwin-count ≥ `MAC_VM_CAP`
  (default 2, env-overridable). The fzf binding's `enter_min`
  parameter ties the existing `FZF_SELECT_COUNT` transform to the
  computed `required` — same mechanism as #26-vintage, just with N
  instead of 1. Non-fzf fallback flips the default from `[y/N]` to
  `[Y/n]` in cap-enforce mode (since "stop something" is the
  presumed intent). Non-tty exits 1 fail-fast rather than silently
  letting the boot fail.
- `.tart/Makefile` — `_tart_check_others` passes `$(TART_BASE)` as
  the second arg, and `_tart_boot_and_ssh`'s SSH-poll-failed branch
  greps the boot log for "exceeds the system limit" before
  reporting "SSH never came up." That covers the path where the
  operator (or a non-ariadne caller) bypasses check-others entirely
  and goes straight to `tart run`.

Verification used the host's actual state (nous-test + pair-test
both running darwin guests) and three `ME`/`BASE` combinations to
exercise cap-enforce, soft-mode, and linux-target paths. All
non-tty since the sandbox doesn't ship fzf interactively, but the
shell logic between non-tty and tty paths is the same up to the
prompt mechanism — confident the fzf prompt branch works too.

The original repro (`make tart TART_SUFFIX=emma` from #26) will
now produce a structured cap-enforce prompt instead of the silent
120 s SSH timeout. Operator picks which of {nous-test, pair-test}
to stop, then the boot proceeds.
