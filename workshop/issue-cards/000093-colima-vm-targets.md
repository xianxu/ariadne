---
id: '000093'
status: done
created: 2026-06-12
updated: 2026-06-12
estimate_hours: 4
actual_hours: 0.48
---

# make colima — Lima-VM testing targets mirroring make tart

## Problem

`.tart/` gives Apple-Silicon **macOS** VM testing (`make tart` / `tart-gui` /
`tart-stop` / `tart-clean`, parallel VMs via `TART_SUFFIX`). There is no
equivalent for **clean Linux** testing of the substrate (the Go `sdlc` binary,
agent CLIs, the install/bootstrap flow). Colima is already installed on the
operator's Mac; we want a sibling base-layer fragment that delivers a Linux
testbed with the same verbs and ergonomics as Tart.
