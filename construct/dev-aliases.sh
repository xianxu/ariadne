#!/usr/bin/env bash
# dev-aliases.sh — emit dev shell functions for Go binaries owned by active
# ariadne-styled sibling repos (ariadne#57).
#
# Each owned cmd/X becomes a function that builds the binary to its OWNER's
# bin/ (the official path, gitignored — not a temp dir; safe for service
# binaries like nous) and runs it in the CALLER's cwd — fresh every call. Works
# for both repo-bound tools (sdlc, operates on whatever repo you're in) and
# run-anywhere tools (nous), which `go run`/`go tool` can't (their cwd is
# pinned to the module dir). Form:
#
#   X() { ( cd OWNER && mkdir -p bin && rm -f bin/X && go build -o bin/X ./cmd/X ) || return; OWNER/bin/X "$@"; }
#
# The `rm -f bin/X` mirrors the owner Makefiles' code-signing-inode safety. The
# function only builds + runs — it does NOT manage services (no `launchctl
# bootout`); use the owner's `make <name>-dev` target for stop-prod-then-serve.
#
# Ownership = location: a binary's source physically in a repo → that repo owns
# it. Re-export symlinks and non-buildable dirs are skipped, so a derivative
# that merely points at an ancestor's source never shadows the owner.
#
# Usage (in .zshrc):  source <(~/workspace/ariadne/construct/dev-aliases.sh)
#
# Flags:
#   --list           print "binary <TAB> owner-repo" instead of function defs
#   --strict         exit non-zero if any binary name is owned by >1 repo
#   --workspace DIR  scan siblings under DIR (default: parent of the ariadne root)
#
# stdout = function defs (or --list table); stderr = warnings. Bash 3.2-safe
# (no associative arrays / mapfile) — macOS system bash.
set -uo pipefail

# Dead-clone basename globs skipped by default (legacy snapshots, numbered
# copies). The symlink + buildable filters below also exclude clones that only
# carry re-export symlinks, so this is mainly noise control for full copies.
SKIP_GLOBS='*legacy* *.original *[0-9]'

list=0
strict=0
workspace=""
workspace_set=0
while [ $# -gt 0 ]; do
	case "$1" in
		--list) list=1 ;;
		--strict) strict=1 ;;
		--workspace) shift; workspace="${1-}"; workspace_set=1 ;;
		--workspace=*) workspace="${1#*=}"; workspace_set=1 ;;
		-h|--help) sed -n '2,30p' "$0"; exit 0 ;;
		*) printf 'dev-aliases.sh: unknown arg: %s\n' "$1" >&2; exit 2 ;;
	esac
	shift
done

warn() { printf 'dev-aliases: %s\n' "$*" >&2; }

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
ariadne_root="$(cd "$script_dir/.." && pwd -P)"
[ "$workspace_set" -eq 1 ] || workspace="$(dirname "$ariadne_root")"
# A typo'd/empty --workspace must fail loudly, not silently emit nothing.
if [ -z "$workspace" ] || [ ! -d "$workspace" ]; then
	warn "workspace not found: ${workspace:-<empty>} (--workspace needs an existing dir)"
	exit 2
fi

skip_repo() {
	local name="$1" g
	for g in $SKIP_GLOBS; do
		case "$name" in $g) return 0 ;; esac
	done
	return 1
}

# Emit "bin<TAB>ownerRepo" for every owned binary, in deterministic walk order:
# sorted ariadne-styled active repos, then sorted cmd dirs that are real dirs
# (not symlinks), buildable (>=1 .go), and not marked .private.
collect() {
	find "$workspace" -mindepth 1 -maxdepth 1 -type d | LC_ALL=C sort | while IFS= read -r repo; do
		[ -d "$repo/construct" ] || continue
		skip_repo "$(basename "$repo")" && continue
		[ -d "$repo/cmd" ] || continue
		find "$repo/cmd" -mindepth 1 -maxdepth 1 | LC_ALL=C sort | while IFS= read -r cmddir; do
			[ -L "$cmddir" ] && continue          # re-export symlink → owner only
			[ -d "$cmddir" ] || continue
			ls "$cmddir"/*.go >/dev/null 2>&1 || continue   # buildable
			[ -e "$cmddir/.private" ] && continue # opt-out marker
			printf '%s\t%s\n' "$(basename "$cmddir")" "$repo"
		done
	done
}

candidates="$(collect)"

# Warn on duplicate binary names (same name owned by >1 active repo).
had_dup=0
dups="$(printf '%s\n' "$candidates" | awk -F'\t' 'NF{print $1}' | LC_ALL=C sort | uniq -d)"
if [ -n "$dups" ]; then
	had_dup=1
	printf '%s\n' "$dups" | while IFS= read -r d; do
		[ -n "$d" ] || continue
		owners="$(printf '%s\n' "$candidates" | awk -F'\t' -v b="$d" '$1==b{printf "%s ", $2}')"
		warn "duplicate binary '$d' owned by: ${owners}(last in sorted order wins)"
	done
fi

# Final mapping: last-win per binary (deterministic since walk order is sorted),
# emitted sorted by binary name.
final="$(printf '%s\n' "$candidates" | awk -F'\t' 'NF{last[$1]=$2} END{for (k in last) printf "%s\t%s\n", k, last[k]}' | LC_ALL=C sort)"

printf '%s\n' "$final" | while IFS="$(printf '\t')" read -r bin repo; do
	[ -n "$bin" ] || continue
	if [ "$list" -eq 1 ]; then
		printf '%s\t%s\n' "$bin" "$repo"
	else
		repo_q="$(printf '%q' "$repo")"
		# Build to the owner's bin/ (official path, gitignored — not a temp
		# dir; safe for service binaries); rm -f first for code-signing-inode
		# safety; run from the caller's cwd.
		printf '%s() { ( cd %s && mkdir -p bin && rm -f bin/%s && go build -o bin/%s ./cmd/%s ) || return; %s/bin/%s "$@"; }\n' \
			"$bin" "$repo_q" "$bin" "$bin" "$bin" "$repo_q" "$bin"
	fi
done

if [ "$strict" -eq 1 ] && [ "$had_dup" -eq 1 ]; then
	warn "strict: duplicate binary names found — exiting non-zero"
	exit 1
fi
exit 0
