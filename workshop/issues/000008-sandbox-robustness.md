---
status: done
---

# 000008 — Sandbox Robustness & Troubleshooting

## Problem

`make sandbox` fails with cascading errors when prerequisites aren't met. The script doesn't pre-check requirements, so failures appear mid-flow with opaque messages.

## Spec

Document the sandbox lifecycle steps, identify failure modes, and add pre-flight checks to `sandbox.sh` so failures are caught early with actionable messages.

## Sandbox Lifecycle Steps

`make sandbox` → `make sandbox-build` → `.openshell/sandbox.sh build <name>`

### Step-by-step flow of `cmd_build`:

1. **Get sandbox phase** — `openshell sandbox list` to check if sandbox exists
2. **Clean bad state** — if sandbox exists but isn't Running/Ready, destroy and recreate
3. **Check base image updates** — fetch digest from GHCR, prompt user if outdated
4. **Create sandbox + bootstrap deps** (parallel):
   - `openshell sandbox create ...` — creates the container via OpenShell gateway
   - `bootstrap.sh` — downloads dependencies to `.openshell/.bootstrap/`
5. **Ensure SSH config** — writes Host block to `~/.ssh/config` using `openshell sandbox ssh-config`
6. **Ensure bootstrap sync** — mutagen one-way sync of `.bootstrap/` → sandbox `/tmp/bootstrap`
7. **Post-install** — SCP + SSH to run `post-install.sh` on sandbox (installs nvim, zellij, etc.)
8. **Apply config** — `setup.sh`, dotfiles, gh auth, git config
9. **Ensure file sync** — mutagen two-way sync for repo, worktree, .git, nvim-state, plenary

### Prerequisites (must be running/available):

| Prerequisite | How to check | Fix |
|---|---|---|
| OpenShell gateway | `openshell gateway status` | `openshell gateway destroy --name openshell && openshell gateway start` |
| Docker daemon | `docker info` | Start Docker Desktop |
| `mutagen` CLI | `which mutagen` | `brew install mutagen-io/mutagen/mutagen` |
| `openshell` CLI | `which openshell` | Install per OpenShell docs |
| `gh` CLI (optional) | `which gh && gh auth status` | `brew install gh && gh auth login` |
| Network to ghcr.io | `curl -sf https://ghcr.io/token` | Check VPN/proxy |

## Common Failure Modes

### 1. Gateway not running (the current failure)
- **Symptom**: `! Gateway 'openshell' is not reachable.` followed by `transport error / Connection refused (os error 61)`
- **Root cause**: OpenShell gateway process not running
- **Fix**: `openshell gateway destroy --name openshell && openshell gateway start`
- **Prevention**: Pre-flight check before any openshell commands

### 2. Docker not running
- **Symptom**: `openshell sandbox create` fails, docker socket not found
- **Root cause**: Docker Desktop not started
- **Fix**: Start Docker Desktop
- **Note**: `sandbox.sh` already has socket auto-detection (lines 7-14), but Docker must be running

### 3. Mutagen sync stale/stuck
- **Symptom**: Files not syncing, `mutagen sync list` shows errors or "Paused"
- **Root cause**: Mutagen daemon state corrupted, or sandbox was recreated without terminating old syncs
- **Fix**: `mutagen sync terminate <name>` then re-run, or `make sandbox-clean`

### 4. SSH config stale
- **Symptom**: `ssh: Could not resolve hostname openshell-*` or connection refused after sandbox recreate
- **Root cause**: SSH config block has old host/port from previous sandbox
- **Fix**: Re-run `make sandbox` (ensure_ssh_config overwrites the block), or manually edit `~/.ssh/config`

### 5. Bootstrap cache incomplete
- **Symptom**: Post-install fails, tools missing on sandbox
- **Root cause**: Previous bootstrap download was interrupted
- **Fix**: `make sandbox-nuke` to wipe cache and rebuild from scratch

### 6. Sandbox in bad state (not Running/Ready)
- **Symptom**: Script detects phase like "Creating", "Error", "Stopped"
- **Root cause**: Previous create failed, or sandbox was stopped externally
- **Fix**: Script auto-cleans this (lines 293-298), but if that fails: `make sandbox-stop && make sandbox`

## Plan

- [x] Document lifecycle steps and failure modes (this file)
- [x] Add pre-flight check function to `sandbox.sh` that validates gateway + docker before proceeding
- [x] Improve error messages with specific fix instructions

## Log

- 2026-04-21: User reported `make sandbox` failure due to gateway not running. All errors cascade from this single root cause. Created this issue to track robustness improvements.
- 2026-04-21: Added `preflight()` to `sandbox.sh` — auto-starts Docker Desktop and restarts OpenShell gateway. Added wait-for-Running loop before SSH/mutagen steps. Confirmed working.
