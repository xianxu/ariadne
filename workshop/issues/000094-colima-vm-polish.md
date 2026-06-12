---
id: 000094
status: working
deps: [ariadne#93]
github_issue:
created: 2026-06-12
updated: 2026-06-12
estimate_hours: 4
---

# Colima VM post-login parity + unified colorized/dimmed VM logging

Two follow-up improvements to the `make colima` work (#93). Stacked on the
#93 branch; merges together.

## Problem

1. **No post-login customization in the Colima VM.** `make tart` runs
   `tart-vm-setup.sh` (oh-my-zsh, aliases, dev-aliases, workspace wiring,
   per-repo hooks); `make colima` drops you into a bare Ubuntu shell — no
   aliases, no nvim, no dev-aliases, no auto-cd.
2. **`make tart` logging is plainer than `make colima`.** Colima streams its
   boot log live under clear `==>` steps; tart redirects `tart run` to a
   logfile, so the operator sees only bare `==>` lines until SSH. Both want
   colorized step headers + dimmed underlying-process output.

## Spec

Decided via brainstorm 2026-06-12 (portability analysis + AskUserQuestion):
**lean parity**, **skip agent aliases**.

### Part A — Colima post-login setup (mirror the OpenShell *Linux* overlay, not Tart's macOS zsh rc)

- `.colima/vm-setup.sh` (guest bootstrap, run on boot via `colima ssh -- bash
  -s -- <repo-dir> <repo-name>`, idempotent): symlink `~/workspace` → host-abs
  mount + `~/repo` → current repo; write `~/.colima-current-repo` marker;
  install oh-my-bash (network, non-fatal); `apt install neovim`; append a
  source-line for `vm-rc.sh` to `~/.bashrc`; run `.colima/vm-hooks.d/*.sh`
  per-repo hooks (run-parts, lexical, continue-on-error).
- `.colima/vm-rc.sh` (bash rc, the Linux counterpart to `tart-vm-rc.zsh`): git
  aliases (`s ss a d p`), `v=nvim`, vi-mode + Ctrl+R/S, PATH (`~/repo/bin`,
  `~/.local/bin`), `EDITOR=nvim`, GPG_TTY guard, `dev-aliases.sh` Go
  build-on-demand wrappers, auto-cd into current repo on interactive login.
- **Deliberately NOT ported** (sandbox/macOS-specific): the egress-proxy config
  (`https_proxy`, `git http.sslVerify false`, pip/npm proxy), `/tmp/bootstrap`
  credentials, macOS DNS-flush + Homebrew PATH, the `script(1)`+DEBUG-trap
  output-capture machinery, AI-agent auto-approve aliases.

### Part B — Unified VM logging helper

- A shared `vm-log.sh` (one source of the ANSI codes — ARCH-DRY): `step`/`warn`
  print bold-cyan/`yellow` headers; `dim` filters stdin → grayed lines. tty +
  `NO_COLOR` guarded.
- `.colima/colima.sh`: replace plain `echo "==>"` with `vm-log.sh step`; pipe
  `colima start` + the vnc/setup sub-output through `vm-log.sh dim` (NOT the
  final interactive `colima ssh`).
- `.tart/Makefile`: colorize the `==>` recipe lines via `vm-log.sh step`; during
  the SSH-wait, stream the boot logfile dimmed (`tail -f … | vm-log.sh dim`)
  instead of leaving it hidden.

## Done when

- `make colima` lands you in a shell with: nvim installed, git aliases live,
  `~/workspace`/`~/repo` symlinks present, auto-cd into the repo, dev-aliases
  working (when ariadne is in the workspace).
- `.colima/vm-hooks.d/*.sh` run on boot (idempotent, continue-on-error).
- `make tart` and `make colima` both show bold-cyan `==>` steps + dimmed
  underlying-process output; the final interactive shell is NOT dimmed.
- `vm-log.sh` is one shared file used by both fragments; in `base.manifest`.
- Verification: real `make colima` boot showing the setup + new logging, and a
  `make tart`-path logging check, captured in the Log.

## Plan

Durable design: `workshop/plans/000094-colima-vm-polish-plan.md`. Atomic
single-boundary (verified in one boot) → one `sdlc close` (no `Mx`).

- [ ] `construct/scripts/vm-log.sh` + test — shared colorized-step / dimmed-passthrough logger
- [ ] `.colima/vm-rc.sh` — guest bash rc (aliases, vi-mode, dev-aliases, auto-cd)
- [ ] `.colima/vm-setup.sh` — guest bootstrap (symlinks, nvim, oh-my-bash, hooks)
- [ ] `colima.sh` — colorized steps, dimmed sub-output, run guest setup (gated on fresh, non-fatal)
- [ ] `.tart/Makefile` — colorized steps + dimmed boot-log tail during SSH wait
- [ ] `base.manifest` entries + colima test setup-call assertions
- [ ] Real colima boot verification (setup + logging); `make -n tart` check; log evidence

## Log

### 2026-06-12

- Brainstorm: portability analysis (Colima guest = Ubuntu/bash → mirror the
  OpenShell Linux overlay, not Tart's macOS zsh rc). AskUserQuestion settled
  **lean parity** + **skip agent aliases**. Non-portable bits explicitly
  excluded (proxy/creds/macOS-DNS/output-capture).
