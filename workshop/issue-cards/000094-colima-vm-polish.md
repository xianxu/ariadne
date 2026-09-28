---
id: '000094'
status: done
created: 2026-06-12
updated: 2026-06-12
estimate_hours: 4
actual_hours: 0.73
---

# Colima VM post-login parity + unified colorized/dimmed VM logging

## Problem

1. **No post-login customization in the Colima VM.** `make tart` runs
   `tart-vm-setup.sh` (oh-my-zsh, aliases, dev-aliases, workspace wiring,
   per-repo hooks); `make colima` drops you into a bare Ubuntu shell — no
   aliases, no nvim, no dev-aliases, no auto-cd.
2. **`make tart` logging is plainer than `make colima`.** Colima streams its
   boot log live under clear `==>` steps; tart redirects `tart run` to a
   logfile, so the operator sees only bare `==>` lines until SSH. Both want
   colorized step headers + dimmed underlying-process output.
