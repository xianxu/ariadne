# Colima-VM management — included from Makefile.workflow:
#   -include .colima/Makefile
#
# The Linux counterpart to .tart (macOS VMs). The per-repo unit is a
# Colima *profile* (a Lima Linux VM), named like TART_VM:
#   COLIMA_PROFILE = <repo>-<COLIMA_SUFFIX|test>
#
# Provides:
#   make colima        Start profile $(COLIMA_PROFILE) (idempotent), bind-mount
#                      the workspace ($(WORKSPACE_DIR)) writable, ssh in landing
#                      in the current repo. Mounts are LIVE virtiofs (no copy),
#                      so host edits — incl. uncommitted — are reflected live.
#   make colima-gui    Same, plus provision a TigerVNC desktop in the guest and
#                      print how to connect (vnc://localhost:5901). Colima/Linux
#                      is headless, so this apt-installs X+VNC — heavier than
#                      tart-gui (which inherits the macOS desktop for free).
#   make colima-stop   Stop $(COLIMA_PROFILE).
#   make colima-clean  Stop + delete $(COLIMA_PROFILE) (Lima base image cached).
#
# Parallel profiles per repo: COLIMA_SUFFIX=ying make colima → <repo>-ying.
# Resource knobs (env): COLIMA_CPUS/MEMORY/DISK/ARCH/VMTYPE — see colima.sh.
# amd64 testing: COLIMA_ARCH=x86_64 make colima (adds --vz-rosetta under vz).
#
# All Colima interaction lives in .colima/colima.sh; these recipes are thin
# wrappers (mirrors .openshell/sandbox.sh). Requires `brew install colima`.

# Profile name: sanitize like .openshell/Makefile (lowercase, '.'→'-').
COLIMA_PROFILE = $(shell echo "$(REPO_NAME)-$(or $(COLIMA_SUFFIX),test)" | tr '.' '-' | tr '[:upper:]' '[:lower:]')

# Workspace parent dir, bind-mounted writable into the VM (siblings adjacent
# so go.mod `replace ../peer` resolves). recursive (=) so realpath only runs
# when a colima target actually expands it, not on every `make`.
WORKSPACE_DIR = $(realpath $(CURDIR)/..)

.PHONY: colima colima-gui colima-stop colima-clean help-colima

help-colima:
	@printf '%s\n' \
	"Colima (Linux VMs):" \
	"  make colima            Start Linux profile $(COLIMA_PROFILE) (idempotent)," \
	"                         bind-mount the workspace writable, ssh in (lands in repo)." \
	"  make colima-gui        Same + provision a TigerVNC desktop (vnc://localhost:5901)." \
	"  make colima-stop       Stop $(COLIMA_PROFILE)." \
	"  make colima-clean      Stop + delete $(COLIMA_PROFILE) (base image stays cached)." \
	"" \
	"  Parallel profiles: COLIMA_SUFFIX=ying make colima → <repo>-ying." \
	"  amd64 guest: COLIMA_ARCH=x86_64 make colima. Knobs: COLIMA_CPUS/MEMORY/DISK." \
	""

colima:
	@.colima/colima.sh up   $(COLIMA_PROFILE) $(CURDIR) $(WORKSPACE_DIR)

colima-gui:
	@.colima/colima.sh gui  $(COLIMA_PROFILE) $(CURDIR) $(WORKSPACE_DIR)

colima-stop:
	@.colima/colima.sh stop $(COLIMA_PROFILE)

colima-clean:
	@.colima/colima.sh clean $(COLIMA_PROFILE)
