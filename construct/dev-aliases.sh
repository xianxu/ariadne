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
#   X() { ( cd OWNER && mkdir -p bin && d="$(mktemp -d bin/.X.XXXXXX)" &&
#     { go build -o "$d/X" ./cmd/X && mv -f "$d/X" bin/X; s=$?; rm -rf "${d:?}"; exit $s; } ) || return; OWNER/bin/X "$@"; }
#
# An owner whose build needs a generated input (a gitignored embed tree, say)
# declares a `dev-prepare` make rule (#299; what counts as declared is
# described once, at owner_prepares below). Its functions then run that step
# first, with make's stdout sent to stderr so the binary's own stdout stays
# clean; a failed step fails the call:
#
#   X() { ( cd OWNER && make -s dev-prepare >&2 && mkdir -p bin && ... ) || return; ...; }
#
# The step is the owner's to keep cheap when up to date, and so is parsing its
# Makefiles, which detection does once per owner whenever the aliases are
# emitted. An owner that declares nothing keeps the plain form above.
#
# The build goes to a private directory beside the target and is renamed over
# bin/X (#355): a rename on one filesystem is atomic, so the old binary stays until
# the new one replaces it, and a concurrent call never finds bin/X missing (a
# per-call mktemp directory, so concurrent builds never share a path; the last
# rename wins). The rename also gives the new binary a new inode, the macOS
# code-signing safety the owner Makefiles get from `rm -f` first. A build killed
# mid-way leaves its bin/.X.* directory behind (gitignored with bin/). The
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
		-h|--help) awk 'NR==1{next} /^#/{print; next} {exit}' "$0"; exit 0 ;;
		*) printf 'dev-aliases.sh: unknown arg: %s\n' "$1" >&2; exit 2 ;;
	esac
	shift
done

# NOUS_DEV gate: in prod (NOUS_DEV=0) emit nothing, so `source <(dev-aliases.sh)`
# is a no-op and callers fall back to the prebuilt binaries on PATH (e.g.
# ~/workspace/nous/bin/nous) instead of rebuilding on every call. Unset / 1 /
# anything-else => dev mode (emit) — the historical default, so this is
# backward-compatible. Diagnostic flags (--list / --strict) bypass the gate.
# NOTE: sourced via process substitution, so NOUS_DEV must be EXPORTED to be
# seen by this child process.
if [ "$list" -eq 0 ] && [ "$strict" -eq 0 ] && [ "${NOUS_DEV:-1}" = 0 ]; then
	exit 0
fi

warn() { printf 'dev-aliases: %s\n' "$*" >&2; }

# script_dir: the directory of the dev-aliases.sh PATH as invoked — NOT
# dereferencing the symlinked file (pwd -P resolves only path COMPONENTS). In a
# derivative this is that derivative's own construct/, so ariadne_root/workspace
# anchor the sibling scan at the CALLER's location exactly as before — the
# flat/peer semantics are unchanged.
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
ariadne_root="$(cd "$script_dir/.." && pwd -P)"
[ "$workspace_set" -eq 1 ] || workspace="$(dirname "$ariadne_root")"

# real_script_dir: the directory of the REAL dev-aliases.sh, following the
# symlink (this script is symlinked into every derivative's construct/, pointing
# back at the owning substrate's copy). Needed ONLY to locate the shared
# lib-deps.sh beside the real source — the substrate scan must reuse the owner's
# parser, not look for a (non-existent) scripts/lib-deps.sh in the derivative.
# readlink -f isn't on macOS bash 3.2; hand-roll a single-hop dereference
# (the link points straight at the source, so one hop suffices).
real_src="${BASH_SOURCE[0]}"
if [ -L "$real_src" ]; then
	link_target="$(readlink "$real_src")"
	case "$link_target" in
		/*) real_src="$link_target" ;;                       # absolute link
		*)  real_src="$(dirname "$real_src")/$link_target" ;; # relative to the link's dir
	esac
fi
real_script_dir="$(cd "$(dirname "$real_src")" && pwd -P)"
# A typo'd/empty --workspace must fail loudly, not silently emit nothing.
if [ -z "$workspace" ] || [ ! -d "$workspace" ]; then
	warn "workspace not found: ${workspace:-<empty>} (--workspace needs an existing dir)"
	exit 2
fi

# current_repo: the repo whose construct/deps substrate graph we ALSO follow for
# owners (directory-agnostic, #95). It's the repo the caller is in — find it by
# walking up from $PWD to the nearest dir holding construct/ (an ariadne-styled
# repo root). Empty when the caller isn't inside one (then only the sibling scan
# applies). This is what makes a NON-PEER layout work: the current repo's
# substrate ancestor (e.g. ariadne) need not be a workspace sibling.
find_current_repo() {
	local d="$PWD"
	while [ "$d" != "/" ] && [ -n "$d" ]; do
		[ -d "$d/construct" ] && { printf '%s\n' "$d"; return 0; }
		d="$(dirname "$d")"
	done
	return 0
}
current_repo="$(find_current_repo)"

# Transitive substrate repos of the current repo, via the shared construct/deps
# parser (lib-deps.sh — the bash parser, NOT weave: this runs pre-weave, so
# reusing weave here would be a chicken-egg). Walk the substrate graph BFS,
# deduping, so a diamond/chain resolves each substrate once. lib-deps.sh lives
# beside the REAL script ($real_script_dir/scripts/lib-deps.sh) — resolved
# through the dev-aliases.sh symlink above, so a derivative reaches the owner's
# copy (the derivative has no scripts/lib-deps.sh of its own to source).
substrate_repos() {
	[ -n "$current_repo" ] || return 0
	local lib="$real_script_dir/scripts/lib-deps.sh"
	[ -f "$lib" ] || { warn "lib-deps.sh not found at $lib (substrate scan skipped)"; return 0; }
	# shellcheck disable=SC1090
	. "$lib"
	local seen="" frontier="$current_repo" next repo tgt
	while [ -n "$frontier" ]; do
		next=""
		# Iterate the current frontier (newline-separated).
		while IFS= read -r repo; do
			[ -n "$repo" ] || continue
			while IFS= read -r tgt; do
				[ -n "$tgt" ] || continue
				# Dedup: skip if already seen (newline-bounded substring match).
				case "
$seen
" in *"
$tgt
"*) continue ;; esac
				seen="$seen
$tgt"
				next="$next
$tgt"
				printf '%s\n' "$tgt"
			done <<EOF
$(deps_substrate_targets "$repo")
EOF
		done <<EOF
$frontier
EOF
		frontier="$next"
	done
}

skip_repo() {
	local name="$1" g
	for g in $SKIP_GLOBS; do
		case "$name" in $g) return 0 ;; esac
	done
	return 1
}

# scan_repo emits "bin<TAB>repo" for every owned binary in ONE repo: cmd dirs
# that are real dirs (not re-export symlinks), buildable (>=1 .go), not .private.
scan_repo() {
	local repo="$1" cmddir
	[ -d "$repo/construct" ] || return 0
	skip_repo "$(basename "$repo")" && return 0
	[ -d "$repo/cmd" ] || return 0
	find "$repo/cmd" -mindepth 1 -maxdepth 1 | LC_ALL=C sort | while IFS= read -r cmddir; do
		[ -L "$cmddir" ] && continue          # re-export symlink → owner only
		[ -d "$cmddir" ] || continue
		ls "$cmddir"/*.go >/dev/null 2>&1 || continue   # buildable
		[ -e "$cmddir/.private" ] && continue # opt-out marker
		printf '%s\t%s\n' "$(basename "$cmddir")" "$repo"
	done
}

# Emit "bin<TAB>ownerRepo" for every owned binary, scanning the UNION of:
#   (1) workspace siblings under $workspace (flat/peer layout — unchanged), and
#   (2) the current repo's transitive substrate repos (non-peer layout — #95).
# Repos are deduped by absolute path (a substrate that's also a sibling scans
# once); deterministic last-win downstream handles any residual ordering.
collect() {
	{
		find "$workspace" -mindepth 1 -maxdepth 1 -type d
		substrate_repos
	} | LC_ALL=C sort -u | while IFS= read -r repo; do
		scan_repo "$repo"
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

# owner_prepares: the owner declares a dev-prepare step (#299) when make itself
# can build the target `dev-prepare` in the owner's tree. Make is the authority,
# so includes, multi-target rules and assignments mean what make says they do.
# A cheap text check runs first, over the files the base layer's Makefile reads
# (Makefile, Makefile.workflow, Makefile.local), so an owner that never mentions
# the name pays no make invocation; a rule kept in some other included file
# needs a mention of the name in one of those three to be found. `make -q` runs
# no recipe: exit 0/1 (up to date / would run) is declared; exit 2 saying there
# is no rule for dev-prepare itself is undeclared; any other exit 2 is a
# declared but broken step, emitted so the call fails loudly rather than
# silently skipping it.
owner_prepares() {
	local err rc
	grep -qs 'dev-prepare' "$1/Makefile" "$1/Makefile.workflow" "$1/Makefile.local" || return 1
	err="$(cd "$1" && env -u MAKEFLAGS -u MAKELEVEL make -q dev-prepare 2>&1 >/dev/null)"; rc=$?
	[ "$rc" -ne 2 ] && return 0
	case "$err" in *"No rule to make target "?"dev-prepare'"*) return 1 ;; esac
	return 0
}

prepared="|"
printf '%s\n' "$final" | while IFS="$(printf '\t')" read -r bin repo; do
	[ -n "$bin" ] || continue
	if [ "$list" -eq 1 ]; then
		printf '%s\t%s\n' "$bin" "$repo"
	else
		repo_q="$(printf '%q' "$repo")"
		# Build beside the owner's bin/X (the official path, gitignored) and
		# rename into place, so bin/X is never missing; run from the caller's cwd.
		prep=""
		# One detection per owner, not per binary: cache the answer by repo.
		case "$prepared" in
			*"|$repo=1|"*) prep="make -s dev-prepare >&2 && " ;;
			*"|$repo=0|"*) ;;
			*)
				if owner_prepares "$repo"; then
					prep="make -s dev-prepare >&2 && "; prepared="$prepared$repo=1|"
				else
					prepared="$prepared$repo=0|"
				fi ;;
		esac
		printf '%s() { ( cd %s && %smkdir -p bin && d="$(mktemp -d bin/.%s.XXXXXX)" && { go build -o "$d/%s" ./cmd/%s && mv -f "$d/%s" bin/%s; s=$?; rm -rf "${d:?}"; exit $s; } ) || return; %s/bin/%s "$@"; }\n' \
			"$bin" "$repo_q" "$prep" "$bin" "$bin" "$bin" "$bin" "$bin" "$repo_q" "$bin"
	fi
done

if [ "$strict" -eq 1 ] && [ "$had_dup" -eq 1 ]; then
	warn "strict: duplicate binary names found — exiting non-zero"
	exit 1
fi
exit 0
