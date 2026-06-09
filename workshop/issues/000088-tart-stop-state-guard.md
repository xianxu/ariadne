---
id: 000088
status: working
deps: []
github_issue:
created: 2026-06-08
updated: 2026-06-08
estimate_hours: 0.3
---

# tart-stop guard uses cached tart ip; errors on GUI-shutdown VM

## Problem

`make tart-stop` / `make tart-clean` error out when the VM was shut down
from inside the guest GUI (macOS  → Shut Down) instead of via `tart stop`.

Repro observed: operator ran `make tart-gui`, then shut the VM down from
the GUI. `tart list` then shows the VM `stopped`, but:

```
$ tart stop brain-test
VM "brain-test" is not running        # exit 2
```

Root cause — `tart-stop`'s guard (`.tart/Makefile:368`):

```makefile
@if tart ip $(TART_VM) >/dev/null 2>&1; then tart stop $(TART_VM); fi
```

Tart **caches the last-known IP for a stopped VM**, so `tart ip` exits 0
even when the VM is stopped. The guard concludes "running" → calls
`tart stop` on an already-stopped VM → tart exits 2 → the recipe line
fails → `make tart-stop` aborts. `tart-clean` lists `tart-stop` as a
prerequisite, so it dies at the same prerequisite and never reaches its
own `tart delete` — hence *both* targets error.

The same file already documents this exact gotcha for the **boot**
decision (`.tart/Makefile:258-263`): *"Tart caches the last-known IP for
a stopped VM … Gate the boot decision on `tart list`'s State column, not
on `tart ip`."* The boot path was fixed; `tart-stop` was missed
(`ARCH-DRY` — the state-check idiom exists, this site didn't reuse it).

## Spec

Gate `tart-stop` on the VM's real run-state (the `tart list` State
column), matching the idiom used at lines 235, 262, 331, 352 — not on the
lie-prone `tart ip`. Result: `tart-stop`/`tart-clean` become idempotent
against an already-stopped VM (GUI shutdown, prior `tart stop`, crash).

Scope is exactly one line (368). Lines 269/273 also call `tart ip` but
legitimately — they fetch the *live* IP to SSH into, gated by an `nc`
reachability probe; not a state guard, leave them.

Not chosen: `tart stop … || true`. That masks genuine stop failures —
a lazy null-check, not a root-cause fix (AGENTS.md Core Design
Principles). The state guard is the correct fix.

## Done when

- `tart-stop` skips `tart stop` when the VM is not in `running` state,
  and still runs `_tart_cleanup_clone` unconditionally.
- `make tart-clean` succeeds end-to-end against a stopped/GUI-shutdown VM.
- Verified before/after at the guard level (the lie-prone `tart ip`
  guard vs the state guard) against a real stopped VM.

## Plan

- [ ] Replace the `tart ip` guard at `.tart/Makefile:368` with a
      `tart list` State-column check (reuse the file's existing idiom).
- [ ] Verify: demonstrate old guard returns "running" for a stopped VM
      (the bug) while the new guard correctly returns not-running.

## Log

### 2026-06-08
- Diagnosed live against `brain-test` (GUI-shutdown, `stopped`):
  `tart stop brain-test` → `VM "brain-test" is not running` (exit 2),
  confirming the guard's false-positive. `brain-test` since deleted +
  its clone removed.
- `.tart/Makefile` is base-layer (`construct/base.manifest:149`,
  `symlink`) → fix propagates to every downstream repo via the symlink.
- DRY follow-up (not this issue): the state-extraction
  `tart list | awk '$$2==vm{print $$NF}'` is inlined at 5 sites after
  this fix; could collapse into a `_tart_state` define. Deferred —
  matching the existing inline idiom keeps this fix one line.
