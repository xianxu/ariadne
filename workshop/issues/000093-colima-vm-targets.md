---
id: 000093
status: working
deps: []
github_issue:
created: 2026-06-12
updated: 2026-06-12
estimate_hours: 4
---

# make colima — Lima-VM testing targets mirroring make tart

## Problem

`.tart/` gives Apple-Silicon **macOS** VM testing (`make tart` / `tart-gui` /
`tart-stop` / `tart-clean`, parallel VMs via `TART_SUFFIX`). There is no
equivalent for **clean Linux** testing of the substrate (the Go `sdlc` binary,
agent CLIs, the install/bootstrap flow). Colima is already installed on the
operator's Mac; we want a sibling base-layer fragment that delivers a Linux
testbed with the same verbs and ergonomics as Tart.

## Spec

Mirror Tart's target set with Colima profiles as the per-repo unit (decided via
brainstorm 2026-06-12):

- **Isolation unit = a Colima *profile* (a Lima Linux VM) per repo**, named
  `$(REPO_NAME)-$(or $(COLIMA_SUFFIX),test)` — the structural twin of
  `TART_VM`. `COLIMA_SUFFIX=ying make colima` → a second profile alongside.
- **Workspace mount** mirrors Tart: the current repo + its transitive go.mod
  peers (reuse `construct/scripts/list-peers.sh`) made available inside the VM
  at a stable path. Unlike Tart's COW APFS clone, Colima/Lima mounts are live
  bind mounts — decide RO vs writable + host-safety in the plan.
- **Targets** (parallel to `.tart/Makefile`):
  - `make colima` — start profile (idempotent), mount workspace, `colima ssh`
    in, land the shell in the current repo.
  - `make colima-gui` — VNC path: provision a minimal X + VNC server in the
    guest, forward the port, point the operator at a VNC viewer / Screen
    Sharing. (Colima/Linux is headless; this is heavier than tart-gui, which
    inherits the macOS desktop for free.)
  - `make colima-stop` — stop the profile.
  - `make colima-clean` — delete the profile (Lima base image stays cached).
- **Help**: a `help-colima` fragment wired into the top-level `help` target
  alongside `help-tart` / `help-sandbox`.
- **Base-layer artifact**: lives in `.colima/`, included via `Makefile.workflow`
  parallel to `.tart`, and added to `construct/base.manifest` so downstream
  ariadne repos inherit it.

## Done when

- `make colima` boots a Linux Colima profile `<repo>-test`, mounts the
  workspace (current repo + go.mod peers), and drops into a shell in the repo.
- `COLIMA_SUFFIX=ying make colima` yields a second independent profile.
- `make colima-stop` / `make colima-clean` are idempotent and behave like their
  tart counterparts (stop vs delete).
- `make colima-gui` brings up a VNC-reachable Linux desktop and tells the
  operator how to connect.
- `.colima/` is in `base.manifest`; `make help` lists the colima verbs.
- Verification evidence: a real boot→ssh→stop→clean cycle on this Mac, captured
  in the Log.

## Plan

Durable design: `workshop/plans/000093-colima-vm-targets-plan.md`. Atomic
single-boundary feature → plain checkboxes, one `sdlc close` (no `Mx` split).

- [ ] `.colima/colima.sh` — orchestrator (`up|gui|stop|clean`), bash-3.2-safe
- [ ] `.colima/vnc-setup.sh` — guest TigerVNC+fluxbox provisioner
- [ ] `.colima/Makefile` — thin wrappers + `help-colima` + profile-name/mount derivation
- [ ] Wire into `Makefile.workflow` (`-include`), `Makefile` (`help-colima`), `construct/base.manifest`
- [ ] `.colima/test/colima.test.sh` — process-level fake test (command assembly + idempotency gates)
- [ ] Real boot→ssh→stop→clean cycle on this Mac; resolve R1–R4; log evidence
- [ ] `atlas/workflow/colima-vm.md` + link from `atlas/index.md`

## Log

### 2026-06-12

- Brainstorm settled three forks via AskUserQuestion: unit = **Colima profile
  (VM per repo)**; `colima-gui` = **VNC forwarding**; scope = **base-layer
  artifact** (parallel to `.tart`, into `base.manifest`).
- Durable plan written: `workshop/plans/000093-colima-vm-targets-plan.md`.
  Fresh-eyes plan review caught a critical bash-3.2 bug (empty-array under
  `set -u`) + two test/robustness gaps; all folded into the plan before code.

## Revisions

### 2026-06-12 — mount strategy: whole workspace, not the go.mod-peer subset

- **Delta:** The `## Spec` says "reuse `construct/scripts/list-peers.sh`" for
  the mount. The plan deliberately does **not** — it bind-mounts the whole
  workspace parent (`realpath $(CURDIR)/..`) writable instead.
- **Reason:** Tart restricts to the peer subset because `cp -cR` (clonefile) is
  linear in *file count*, so cloning a multi-GB workspace dominated boot
  (ariadne#32). Colima/Lima mounts are **live virtiofs shares** — nothing is
  copied — so that latency rationale does not transfer. Whole-workspace is
  simpler, always-complete (no pinned peer-set that goes stale on `go.mod`
  change), and keeps siblings adjacent so `replace ../peer` still resolves.
  **ARCH-DRY:** reuse machinery only where its justification holds. Per-peer
  mounting via `list-peers.sh` is recorded as a future extension.
