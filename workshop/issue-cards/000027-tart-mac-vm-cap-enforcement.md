---
id: '000027'
status: done
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
