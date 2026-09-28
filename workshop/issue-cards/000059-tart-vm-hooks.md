---
id: '000059'
status: done
created: 2026-06-01
updated: 2026-06-01
estimate_hours: 1
actual_hours: 0.5
---

# tart vm-hooks.d run-parts convention

## Problem

`make tart` boots a generic VM and runs `.tart/scripts/tart-vm-setup.sh`
(oh-my-zsh + workspace symlinks) — but there's no way for a consuming repo
to inject its own per-VM setup. Concretely, nous needs to make the headless
VM GPG-unattended for brain testing (nous#36), and a base-layer VM with no
extension point would force either a brain-specific hack in ariadne (wrong
layer) or a manual step every boot.

What's missing is a generic, opt-in extension point: a repo declares VM
setup steps; the base-layer setup runs them.
