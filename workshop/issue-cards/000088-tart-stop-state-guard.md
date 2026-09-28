---
id: '000088'
status: done
created: 2026-06-08
updated: 2026-06-08
estimate_hours: 0.3
actual_hours: 0.08
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
