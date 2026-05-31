#!/usr/bin/env bash
# bootstrap-peers.sh — recursively bootstrap substrate peer repos as
# siblings. Reads construct/go.mod for `replace <module> => ../<name>`
# patterns; for each missing peer, derives its URL from this repo's
# origin and clones; then runs `make bootstrap` in the peer to cascade.
#
# Called by Makefile.workflow's `bootstrap-peers` target. Designed for
# one-shot operator use after `git clone <derivative>`.
#
# Cycle detection: ARIADNE_BOOTSTRAP_VISITED env var carries the set of
# already-bootstrapped dirs across recursive `make bootstrap` calls.
# Depth limit: ARIADNE_BOOTSTRAP_DEPTH env var, hard-capped at 5.
#
# Peer URL convention: derived from current repo's `git remote get-url
# origin` by substituting this-repo's-name with peer's-name. Operators
# can override via PEER_URL_<name> in Makefile.local if convention
# doesn't fit (e.g., peer is hosted on a different org).
#
# Idempotent: if peer is already present at the resolved path, just
# recurse into it (idempotent make bootstrap re-runs are cheap).
set -euo pipefail

MAX_DEPTH=5
DEPTH="${ARIADNE_BOOTSTRAP_DEPTH:-0}"

if (( DEPTH > MAX_DEPTH )); then
    echo "Error: bootstrap recursion depth ($DEPTH) exceeded MAX_DEPTH ($MAX_DEPTH)" >&2
    echo "  Visited so far: ${ARIADNE_BOOTSTRAP_VISITED:-<none>}" >&2
    exit 1
fi

TARGET_DIR="$(pwd -P)"

# Cycle detection via visited-set in env var.
VISITED="${ARIADNE_BOOTSTRAP_VISITED:-}"
case ",${VISITED}," in
    *,"$TARGET_DIR",*)
        # Already visited in this cascade — silent no-op.
        exit 0
        ;;
esac
export ARIADNE_BOOTSTRAP_VISITED="${VISITED:+$VISITED,}$TARGET_DIR"
export ARIADNE_BOOTSTRAP_DEPTH=$((DEPTH + 1))

CONSTRUCT_GOMOD="$TARGET_DIR/construct/go.mod"
if [[ ! -f "$CONSTRUCT_GOMOD" ]]; then
    # No construct/go.mod = no substrate peers to bootstrap.
    exit 0
fi

# Derive URL convention for cloning missing peers.
THIS_REPO=$(basename "$TARGET_DIR")
ORIGIN_URL=""
if ORIGIN_URL=$(git -C "$TARGET_DIR" remote get-url origin 2>/dev/null); then
    : # got it
fi

# update_peer <dir> <name> — bring an already-present peer current without
# ever clobbering work in progress. Only fast-forwards a CLEAN tree on a branch
# that tracks an upstream; a dirty tree, detached HEAD, or no upstream is left
# untouched with a warning. Never creates a merge commit, and a diverged branch
# or offline remote is non-fatal — an already-cloned tree must still bootstrap
# without the network. Idempotent: a no-op when the peer is already current.
update_peer() {
    local dir="$1" name="$2" dirty upstream
    dirty="$(git -C "$dir" status --porcelain 2>/dev/null || true)"
    if [[ -n "$dirty" ]]; then
        printf "    skip pull %s — working tree not clean\n" "$name"
        return 0
    fi
    upstream="$(git -C "$dir" rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null || true)"
    if [[ -z "$upstream" ]]; then
        printf "    skip pull %s — no upstream tracking branch (detached or local-only)\n" "$name"
        return 0
    fi
    printf "==> updating peer %s (git pull --ff-only %s)\n" "$name" "$upstream"
    git -C "$dir" pull --ff-only --quiet \
        || printf "    warn: pull of %s failed (diverged or offline) — leaving as-is\n" "$name"
}

# Parse replace directives in construct/go.mod, matching the sibling
# pattern ../<name>. Each matched peer either exists (update + recurse) or
# needs cloning (clone then recurse).
while IFS= read -r line; do
    # Strip line comments.
    line="${line%%//*}"
    # Match: replace <module> [<version>] => ../<path>  (sibling-relative)
    if [[ "$line" =~ ^[[:space:]]*replace[[:space:]]+[^[:space:]]+([[:space:]]+[^[:space:]]+)?[[:space:]]+=\>[[:space:]]+(\.\.[^[:space:]]+) ]]; then
        rhs="${BASH_REMATCH[2]}"
        # rhs is relative to construct/go.mod's dir = $TARGET_DIR/construct
        peer_abs="$(cd "$TARGET_DIR/construct" 2>/dev/null && cd "$rhs" 2>/dev/null && pwd -P || true)"

        if [[ -z "$peer_abs" || ! -d "$peer_abs" ]]; then
            # Peer not present — derive expected path + clone.
            # Resolve path syntactically (without requiring it exists).
            peer_abs="$TARGET_DIR/construct/$rhs"
            peer_abs="$(cd "$(dirname "$peer_abs")" 2>/dev/null && pwd -P || true)/$(basename "$peer_abs")"
            peer_name="$(basename "$peer_abs")"

            if [[ -z "$ORIGIN_URL" ]]; then
                echo "Error: cannot clone peer '$peer_name'; current repo has no 'origin' remote" >&2
                echo "  Configure 'origin' or set PEER_URL_${peer_name} in Makefile.local" >&2
                exit 1
            fi

            # Convention: substitute this-repo-name with peer-name in origin URL.
            peer_url="${ORIGIN_URL//$THIS_REPO/$peer_name}"

            printf "==> cloning peer %s → %s\n" "$peer_name" "$peer_abs"
            printf "    from %s\n" "$peer_url"
            mkdir -p "$(dirname "$peer_abs")"
            if ! git clone "$peer_url" "$peer_abs"; then
                echo "Error: clone failed; check URL convention or override via PEER_URL_${peer_name}" >&2
                exit 1
            fi
        else
            # Peer already present — bring it current before recursing, so its
            # bootstrap runs against the latest substrate (ff-only, never WIP).
            update_peer "$peer_abs" "$(basename "$peer_abs")"
        fi

        # Recurse: run `make bootstrap` in the peer. The peer's bootstrap
        # cascades through its own peers (if any), refreshes its substrate,
        # builds its tools. ARIADNE_BOOTSTRAP_VISITED + _DEPTH are passed
        # via the exported env to prevent cycles + depth-exceed.
        printf "==> bootstrapping peer %s\n" "$(basename "$peer_abs")"
        ( cd "$peer_abs" && make bootstrap ) || {
            echo "Error: peer bootstrap failed at $peer_abs" >&2
            exit 1
        }
    fi
done < "$CONSTRUCT_GOMOD"
