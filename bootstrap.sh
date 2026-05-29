#!/usr/bin/env bash
# bootstrap.sh — first-run entrypoint for a fresh, standalone clone of an
# ariadne-style repo whose upstream peer(s) aren't checked out yet.
#
# WHY THIS FILE IS NOT A SYMLINK
# ------------------------------
# Almost every workflow entrypoint in an ariadne derivative (Makefile,
# Makefile.workflow — which defines `bootstrap:` — construct/setup.sh,
# construct/scripts/bootstrap-peers.sh, AGENTS.md, .tart/, .claude/skills/, …)
# is a sibling-relative SYMLINK into ../<upstream>. On a bare `git clone` with
# no upstream beside it those all dangle, and `make` can't even read its own
# Makefile:
#
#     $ make bootstrap
#     make: Makefile: No such file or directory
#     make: *** No rule to make target `Makefile'.  Stop.
#
# This script is a real committed file precisely so it runs with ZERO substrate
# present. It reads the one substrate file that survives a peerless clone — the
# real go.mod / construct/go.mod — clones the upstream peers as siblings, then
# hands off to `make bootstrap` (whose symlinks now resolve) for the full
# cascade: bootstrap-peers (re-walk + refresh) → tools → sdlc-install.
#
# WHY THE CLONE WALK IS TRANSITIVE (ariadne#45)
# ---------------------------------------------
# Each derivative declares only its DIRECT upstream in construct/go.mod
# (nous → ariadne; pair → ariadne). For a 2-deep chain the direct peer IS the
# Makefile's symlink target, so cloning it is enough for `make` to start. But a
# 3-deep chain (foo → mid → ariadne) symlinks foo/Makefile → ../mid/Makefile →
# ../ariadne/Makefile: `make` cannot read its Makefile until the ENTIRE chain is
# present. Cloning only the direct peer (mid) leaves ariadne absent, the symlink
# chain dangles, and the very `make bootstrap` that would have cloned ariadne
# (via bootstrap-peers.sh) can never start. Chicken-and-egg.
#
# So this script walks the chain TRANSITIVELY (BFS): clone a peer, then read
# THAT peer's go.mod / construct/go.mod and continue, until the whole substrate
# tree is on disk. Only then does it hand off. We do an in-process BFS rather
# than `cd <peer> && ./bootstrap.sh` recursion for two reasons: (1) each
# bootstrap.sh ends in `exec make bootstrap`, so the deepest peer's exec would
# replace the process and orphan the top repo; (2) the BFS depends only on each
# peer's go.mod — guaranteed present on a bare clone — not on each peer shipping
# a committed bootstrap.sh (a recent `seed`; older peers may lack it).
#
# It is intentionally generic (no repo-specific knowledge) and idempotent:
# peers already present are skipped (just traversed); re-running re-hands-off to
# make. Delivered to derivatives via the manifest `seed` action (write-once).
#
# PARSER PARITY: the `replace … => <local-path>` matcher below is intentionally
# identical to construct/scripts/.../list-peers.sh and setup.sh's
# discover_ancestors. bootstrap.sh keeps its own inline copy (zero-substrate
# constraint above), locked to the canonical walker by the drift test in
# construct/scripts/test/. Keep the three in sync.
#
# Test / debug hooks (env vars):
#   BOOTSTRAP_DRY_RUN=1    walk + list the transitive peer set, clone nothing,
#                          no handoff. (drift test, "what would this clone?")
#   BOOTSTRAP_CLONE_ONLY=1 do the transitive clone, but skip `exec make
#                          bootstrap` (test the clone walk in isolation).
set -euo pipefail

repo_root="$(cd "$(dirname "$0")" && pwd -P)"
cd "$repo_root"

DRY_RUN="${BOOTSTRAP_DRY_RUN:-}"
CLONE_ONLY="${BOOTSTRAP_CLONE_ONLY:-}"
MAX_DEPTH="${BOOTSTRAP_MAX_DEPTH:-5}"

# Hand off to make. Honors the CLONE_ONLY/DRY_RUN test hooks.
handoff() {
    if [[ -n "$DRY_RUN" || -n "$CLONE_ONLY" ]]; then
        echo "bootstrap: (clone-only) skipping 'make bootstrap' handoff." >&2
        exit 0
    fi
    echo "bootstrap: peers ready — handing off to 'make bootstrap'"
    exec make bootstrap
}

# No substrate go.mod at all → nothing to clone; let make take over (or no-op
# under the test hooks).
if [[ ! -f construct/go.mod && ! -f go.mod ]]; then
    echo "bootstrap: no go.mod / construct/go.mod (no substrate peers) — handing off to make." >&2
    handoff
fi

# ── Peer URL resolution ───────────────────────────────────────────────────────
# Convention: substitute this-repo-name → peer-name in the current repo's
# `origin` URL. Operators can override per-peer via the PEER_URL_<name> env var
# (name sanitized: non-alphanumerics → '_', e.g. parley.nvim → PEER_URL_parley_nvim),
# mirroring the convention bootstrap-peers.sh documents.
peer_url() {
    local name="$1" origin="$2" this_name="$3" var val
    var="PEER_URL_$(printf '%s' "$name" | tr -c '[:alnum:]' '_')"
    val="${!var:-}"
    if [[ -n "$val" ]]; then printf '%s\n' "$val"; return 0; fi
    if [[ -z "$origin" ]]; then
        echo "bootstrap: peer '$name' missing and this repo has no 'origin' remote." >&2
        echo "  Clone it manually beside this repo, or set ${var}=<url>, then re-run." >&2
        return 1
    fi
    # Global substitution (every occurrence), matching bootstrap-peers.sh's
    # convention: an origin that embeds the repo name in the org too (e.g.
    # .../pair-org/pair.git) rewrites both — use PEER_URL_<name> to override.
    printf '%s\n' "${origin//$this_name/$name}"
}

# ── Transitive clone BFS ───────────────────────────────────────────────────────
# queue entries are "depth:abspath". seen + discovered track visited dirs and
# the (non-root) peer set respectively. Globals mutated by walk_gomod.
queue=("0:$repo_root")
seen=()
discovered=()

_is_seen() {
    local p="$1" s
    for s in "${seen[@]+"${seen[@]}"}"; do [[ "$s" == "$p" ]] && return 0; done
    return 1
}

# walk_gomod <gomod-file> <base-dir> <origin> <this-name> <depth>
# Parse one go.mod for local-path replace targets, resolve each (syntactically,
# so absent peers resolve too), clone if missing (unless DRY_RUN), and enqueue
# at depth+1. base-dir is the dir relative paths resolve against (the go.mod's
# own directory, per Go's replace semantics).
walk_gomod() {
    local gm="$1" base="$2" origin="$3" this_name="$4" depth="$5"
    local line rhs raw parent peer name url real
    [[ -f "$gm" ]] || return 0
    while IFS= read -r line; do
        line="${line%%//*}"
        [[ "$line" =~ ^[[:space:]]*replace[[:space:]]+[^[:space:]]+([[:space:]]+[^[:space:]]+)?[[:space:]]+=\>[[:space:]]+([^[:space:]]+) ]] || continue
        rhs="${BASH_REMATCH[2]}"
        # Local-path replace targets only (sibling ../, ./, or absolute).
        case "$rhs" in /*|./*|../*) ;; *) continue ;; esac
        if [[ "$rhs" == /* ]]; then raw="$rhs"; else raw="$base/$rhs"; fi
        # Resolve syntactically: dirname exists (at/above repo_root) even when
        # the peer itself doesn't yet.
        parent="$(cd "$(dirname "$raw")" 2>/dev/null && pwd -P || true)"
        [[ -n "$parent" ]] || { echo "bootstrap: cannot resolve peer path for '$rhs' (in $gm)" >&2; exit 1; }
        peer="$parent/$(basename "$raw")"
        name="$(basename "$peer")"

        if [[ ! -d "$peer" ]]; then
            if [[ -n "$DRY_RUN" ]]; then
                echo "bootstrap: (dry-run) would clone '$name' → $peer" >&2
            else
                url="$(peer_url "$name" "$origin" "$this_name")" || exit 1
                echo "bootstrap: cloning peer '$name'" >&2
                echo "    from $url" >&2
                echo "    into $peer" >&2
                mkdir -p "$(dirname "$peer")"
                git clone "$url" "$peer"
            fi
        else
            echo "bootstrap: peer '$name' already present ($peer)" >&2
        fi

        # Canonicalize if now present (post-clone); else keep syntactic path
        # (dry-run for an absent peer — still record + traverse what we can).
        if [[ -d "$peer" ]]; then real="$(cd "$peer" && pwd -P)"; else real="$peer"; fi
        discovered+=("$real")
        queue+=("$((depth + 1)):$real")
    done < "$gm"
}

while [[ ${#queue[@]} -gt 0 ]]; do
    entry="${queue[0]}"
    queue=("${queue[@]:1}")
    depth="${entry%%:*}"
    dir="${entry#*:}"

    _is_seen "$dir" && continue
    seen+=("$dir")

    if (( depth > MAX_DEPTH )); then
        echo "bootstrap: peer chain deeper than MAX_DEPTH ($MAX_DEPTH) at $dir" >&2
        echo "  Likely a replace cycle the visited-set didn't catch, or a genuinely" >&2
        echo "  deep chain — raise BOOTSTRAP_MAX_DEPTH if the latter." >&2
        exit 1
    fi

    # dir is on disk here (root, or cloned before being enqueued). Derive its
    # own origin + name so its peers clone from the right URL family: this_name
    # is the CURRENT node's basename, substituted out of its origin to form each
    # child's clone URL (a just-cloned mid's origin → base via mid→base subst).
    origin="$(git -C "$dir" remote get-url origin 2>/dev/null || true)"
    this_name="$(basename "$dir")"

    # Replace targets may live in the root go.mod and/or construct/go.mod (the
    # derivative convention — substrate replace lives in construct/). Walk both.
    walk_gomod "$dir/go.mod" "$dir" "$origin" "$this_name" "$depth"
    walk_gomod "$dir/construct/go.mod" "$dir/construct" "$origin" "$this_name" "$depth"
done

if [[ -n "$DRY_RUN" ]]; then
    # Emit the discovered peer set (deduped, sorted) for inspection / drift test.
    if [[ ${#discovered[@]} -gt 0 ]]; then
        printf '%s\n' "${discovered[@]}" | sort -u
    fi
    exit 0
fi

handoff
