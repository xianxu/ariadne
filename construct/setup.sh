#!/usr/bin/env bash
# Ariadne / multi-layer Base Layer Setup
# Bootstraps a target repo by walking each transitive upstream's
# construct/base.manifest in topological order, then applies post-
# processing (creates Makefile, AGENTS.local.md, .gitignore entries,
# skill symlink sync).
#
# Symlink-only model (per ariadne#38): all substrate text is symlinked
# from upstream peers. Operator-divergent customization happens via
# per-operator branches in upstream source repos, not per-derivative
# copies. Six actions: symlink, tool, scaffold, touch, merge, seed.
# (`seed` is the lone copy-shaped action — write-once delivery of a
# real first-run entrypoint that can't be a symlink; see create_seed.)
#
# Upstream discovery
# ------------------
# Two paths:
#   1. Go-managed (target has go.mod) — `go list -m all` + replace
#      directive walk returns every transitive substrate peer.
#   2. Fallback (no go.mod, or no Go) — single ancestor = the script's
#      own resolved upstream. Preserves backward compat with today's
#      `../ariadne/construct/setup.sh` sibling invocation pattern.
#
# Usage:
#   cd /path/to/your-repo && ../ariadne/construct/setup.sh
#
# Idempotent — safe to re-run for updates.
set -euo pipefail

# ── Resolve paths ─────────────────────────────────────────────────────────────
# SCRIPT_REAL = where the script actually lives (followed through symlinks).
# When invoked via `../nous/construct/setup.sh` and that file is a symlink to
# ariadne's setup.sh, SCRIPT_REAL resolves to ariadne's path.
SCRIPT_REAL="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || realpath "${BASH_SOURCE[0]}")")" && pwd)"
# ARIADNE_DIR (legacy name) = the script's resolved upstream root. Used only
# as the fallback ancestor when go.mod-based discovery returns nothing —
# i.e., for first-time bootstrap and pre-Go consumers.
ARIADNE_DIR="$(dirname "$SCRIPT_REAL")"
# pwd -P canonicalizes to the physical path. Without this, on macOS where
# /tmp → /private/tmp, TARGET_DIR comes back as /tmp/... (logical) while
# SCRIPT_REAL resolves to /Users/... (physical). Python's relpath then
# computes a 3-level-up path that resolves wrong when the OS follows the
# symlink from its physical location (which is 4 levels deep). Forcing
# physical here makes both paths consistent so relpath math matches OS
# symlink resolution.
TARGET_DIR="$(pwd -P)"

# Shared construct/deps parser (#60). SCRIPT_REAL is this script's real dir
# (followed through the symlink), so the sibling scripts/lib-deps.sh is ariadne's.
# shellcheck source=/dev/null
. "$SCRIPT_REAL/scripts/lib-deps.sh"

# ── Colors ────────────────────────────────────────────────────────────────────
GREEN='\033[1;32m'
YELLOW='\033[1;33m'
CYAN='\033[1;36m'
BOLD_RED='\033[1;31m'
RESET='\033[0m'

# ── Helpers ───────────────────────────────────────────────────────────────────
rel_path() {
    python3 -c "import os.path; print(os.path.relpath('$1', '$2'))"
}

ensure_parent() {
    local path="$1"
    local parent
    parent=$(dirname "$path")
    [[ -d "$parent" ]] || mkdir -p "$parent"
}

create_symlink() {
    local src="$1"  # absolute path in upstream
    local dst="$2"  # absolute path in target

    ensure_parent "$dst"

    local rel
    rel=$(rel_path "$src" "$(dirname "$dst")")

    if [[ -L "$dst" ]]; then
        local existing
        existing=$(readlink "$dst")
        if [[ "$existing" == "$rel" ]]; then
            return 0  # already correct
        fi
        rm "$dst"
        printf "  ${YELLOW}updated${RESET} %s\n" "${dst#$TARGET_DIR/}"
    elif [[ -e "$dst" ]]; then
        rm -rf "$dst"
        printf "  ${YELLOW}relinked${RESET} %s (was a regular file/dir)\n" "${dst#$TARGET_DIR/}"
    else
        printf "  ${GREEN}linked${RESET}  %s\n" "${dst#$TARGET_DIR/}"
    fi

    ln -s "$rel" "$dst"
}

create_scaffold() {
    local dir="$1"
    if [[ -d "$dir" ]]; then
        return 0
    fi
    mkdir -p "$dir"
    touch "$dir/.gitkeep"
    printf "  ${GREEN}created${RESET} %s/\n" "${dir#$TARGET_DIR/}"
}

# create_seed — content-tracking real-file copy of an upstream file into the
# target. Unlike `symlink`, the result is a standalone file that survives a
# clone with NO upstream beside it — for first-run entrypoints (bootstrap.sh)
# that must run before any substrate is present, so they definitionally can't
# be symlinks. A seed is really a *flattened symlink*: its content is owned by
# upstream and carries no local edits, so it TRACKS upstream — refreshed on
# every run when it drifts (created on first run, updated when upstream changed,
# silent no-op when already identical). This converges a derivative whose seed
# predates an upstream change (e.g. nous's pre-#45 bootstrap.sh). For files the
# operator is meant to edit after delivery, use `scaffold`/`touch` (write-once),
# not `seed`. Mode is preserved via `cp -p` so an executable source lands
# executable.
create_seed() {
    local src="$1" dst="$2"
    if [[ ! -f "$src" ]]; then
        printf "  ${YELLOW}warn${RESET}    seed source missing: %s\n" "${src#$upstream/}"
        return 0
    fi
    if [[ -f "$dst" ]] && cmp -s "$src" "$dst"; then
        return 0   # already current — idempotent no-op, no churn on re-runs
    fi
    local verb=seeded
    [[ -f "$dst" ]] && verb=updated   # existed but drifted from upstream
    ensure_parent "$dst"
    cp -p "$src" "$dst"
    printf "  ${GREEN}%s${RESET}  %s\n" "$verb" "${dst#$TARGET_DIR/}"
}

merge_settings() {
    local base_file="$1"   # upstream's settings.<layer>.json
    local target_file="$2" # target's settings.json (generated, gitignored)

    ensure_parent "$target_file"

    [[ -L "$target_file" ]] && rm "$target_file"

    local target_dir
    target_dir=$(dirname "$target_file")
    local had_local=false
    [[ -f "$target_dir/settings.local.json" ]] && had_local=true

    "$SCRIPT_REAL/scripts/merge-settings.sh" "$base_file" "$target_dir"

    if "$had_local"; then
        printf "  ${YELLOW}merged${RESET}  %s (base + local)\n" "${target_file#$TARGET_DIR/}"
    else
        printf "  ${GREEN}created${RESET} %s (from base, no local overrides)\n" "${target_file#$TARGET_DIR/}"
    fi
}

# ── Ancestor discovery ────────────────────────────────────────────────────────
# Returns one ancestor path per line, in topological order (ancestors of
# transitive depth N appear before depth N-1, so manifests apply foundation-
# first). Empty output if no ancestors found.
#
# Three sources of ancestor candidates, accumulated then ordered:
#
#   1. Recursive `replace <module> => <local-path>` walk starting at
#      target's go.mod. When a replaced path itself has a go.mod with
#      further replace directives, recurse into it. This lets a baby
#      brain declare just `replace nous => ../nous` and have ariadne
#      get picked up transitively via nous's own go.mod. Discovery is
#      BFS; the resulting list is reversed at the end so deeper layers
#      (foundation) come first.
#
#   2. `go list -m -f '{{.Dir}}' all` for Go-imported transitive deps —
#      catches modules that are actually imported in Go code (which
#      require lines survive `go mod tidy`). Adds dirs that the replace
#      walk didn't already find.
#
#   3. Script's own resolved upstream — last-resort fallback for pre-Go
#      consumers or first-time-bootstrap cases where go.mod is absent.
#
# Candidates are filtered to dirs shipping construct/base.manifest; the
# target itself is never an ancestor of itself (its own manifest is
# walked separately AFTER ancestors).
discover_ancestors() {
    local ancestors=()
    local seen=()

    _seen_or_add() {
        # Adds to seen + ancestors. Returns 0 if added, 1 if already seen
        # or filtered out. Args: <abs-dir>
        local dir="$1"
        [[ -z "$dir" ]] && return 1
        [[ "$dir" == "$TARGET_DIR" ]] && return 1
        [[ ! -f "$dir/construct/base.manifest" ]] && return 1
        local s
        for s in "${seen[@]+"${seen[@]}"}"; do
            [[ "$s" == "$dir" ]] && return 1
        done
        seen+=("$dir")
        ancestors+=("$dir")
        return 0
    }

    _parse_replace_paths() {
        # Reads a go.mod from $1, prints each replace's local-path target.
        # Resolves relative paths against $1's dir (canonicalized to
        # physical via pwd -P so subsequent comparisons are consistent).
        local gomod_dir="$1"
        [[ -f "$gomod_dir/go.mod" ]] || return 0
        while IFS= read -r line; do
            line="${line%%//*}"
            if [[ "$line" =~ ^[[:space:]]*replace[[:space:]]+[^[:space:]]+([[:space:]]+[^[:space:]]+)?[[:space:]]+=\>[[:space:]]+([^[:space:]]+) ]]; then
                local rhs="${BASH_REMATCH[2]}"
                if [[ "$rhs" == /* || "$rhs" == ./* || "$rhs" == ../* ]]; then
                    local abs
                    if [[ "$rhs" == /* ]]; then
                        abs="$rhs"
                    else
                        abs="$(cd "$gomod_dir" && cd "$rhs" 2>/dev/null && pwd -P || true)"
                    fi
                    [[ -n "$abs" ]] && printf '%s\n' "$abs"
                fi
            fi
        done < "$gomod_dir/go.mod"
    }

    # Source 1: recursive replace walk (BFS). Each ancestor's own go.mod is
    # then probed for further replace directives, building the chain
    # without requiring the user to redeclare transitive replaces at the
    # leaf.
    #
    # Per node we walk BOTH go.mods: the repo-root go.mod (operator-owned app
    # deps — may carry self-declared sibling replaces) AND construct/go.mod
    # (substrate-tool deps — where the ariadne/upstream replace actually lives
    # in a derivative; see setup-and-replication.md). Walking only the root
    # misses the substrate ancestor for any depth-≥2 derivative whose upstream
    # parks its own ancestor in construct/go.mod (e.g. brain → nous → ariadne:
    # the nous→ariadne hop is in nous/construct/go.mod, so a root-only walk from
    # brain stops at nous and never applies ariadne's manifest). This matches
    # the both-go.mods convention already used by list-peers.sh and
    # bootstrap-peers.sh; discover_ancestors was the lone root-only walker (#50).
    if [[ -f "$TARGET_DIR/go.mod" || -f "$TARGET_DIR/construct/go.mod" || -f "$TARGET_DIR/construct/deps" ]]; then
        local queue=("$TARGET_DIR")
        while [[ ${#queue[@]} -gt 0 ]]; do
            local current="${queue[0]}"
            queue=("${queue[@]:1}")
            # Both go.mods (legacy) AND construct/deps (#60, dual-read). The
            # _seen_or_add filter (requires construct/base.manifest) drops any
            # absent/non-substrate target, so syntactic resolution is safe here.
            while IFS= read -r candidate; do
                if _seen_or_add "$candidate"; then
                    queue+=("$candidate")
                fi
            done < <(_parse_replace_paths "$current"; _parse_replace_paths "$current/construct"; deps_substrate_targets "$current")
        done

        # Source 2: go list -m all (for code-imported deps that aren't in
        # the replace chain). Order from go list is Go's MVS resolution,
        # which doesn't necessarily match topological-by-replace; we add
        # these after the replace walk so the latter wins for ordering.
        if command -v go >/dev/null 2>&1; then
            while IFS= read -r dir; do
                _seen_or_add "$dir" || true
            done < <(cd "$TARGET_DIR" && go list -m -f '{{.Dir}}' all 2>/dev/null)
        fi
    fi

    # Source 3: script's own upstream (last-resort fallback)
    if [[ ${#ancestors[@]} -eq 0 ]] && [[ "$ARIADNE_DIR" != "$TARGET_DIR" ]]; then
        _seen_or_add "$ARIADNE_DIR" || true
    fi

    # Topological ordering: BFS discovery visits depth-1 first, depth-2
    # second, etc. We want foundation-first (deepest first), so reverse.
    local i
    for ((i=${#ancestors[@]}-1; i>=0; i--)); do
        printf '%s\n' "${ancestors[$i]}"
    done
}

# ── Walk one upstream's manifest into the target ──────────────────────────────
walk_manifest() {
    local upstream="$1"
    local manifest="$upstream/construct/base.manifest"

    if [[ ! -f "$manifest" ]]; then
        printf "  ${YELLOW}skip${RESET}    no construct/base.manifest at %s\n" "$upstream"
        return 0
    fi

    printf "\n  ${CYAN}[%s]${RESET}\n" "$(basename "$upstream")"

    while IFS= read -r line; do
        [[ "$line" =~ ^[[:space:]]*# ]] && continue
        [[ -z "${line// /}" ]] && continue

        read -r action source target <<< "$line"
        target="${target:-$source}"

        # Self-reference filter: when walking target's own manifest (upstream
        # == target), entries whose source path equals target path would
        # destroy the canonical file by trying to symlink/copy it onto
        # itself. Skip them — the file is already where it belongs. This
        # lets a layer's manifest contain entries that ARE meaningful when
        # applied to that layer itself (e.g., `symlink construct/skills/X
        # .claude/skills/X` exposes a skill via the Claude-Code-expected
        # path) while protecting entries like `symlink Makefile.nous`
        # (declared for downstream consumers; tautological in nous itself).
        #
        # Exceptions: these actions are not file-shape operations and the
        # self-reference filter doesn't apply to them.
        #   merge — implicit source-rename (reads .X.<layer>.json, writes
        #           .X.json; different files). On self-walk, regenerates
        #           the layer's own settings.json from committed + local.
        #   tool  — modifies the target's go.mod via `go mod edit`. On
        #           self-walk, adds the tool directive to the upstream's
        #           own go.mod (so `go tool sdlc` works locally there too).
        if [[ "$action" != "merge" && "$action" != "tool" && "$upstream/$source" == "$TARGET_DIR/$target" ]]; then
            printf "  ${YELLOW}skipped${RESET} %s (self-reference at canonical location)\n" "$target"
            continue
        fi

        case "$action" in
            symlink)
                create_symlink "$upstream/$source" "$TARGET_DIR/$target"
                ;;
            scaffold)
                create_scaffold "$TARGET_DIR/$target"
                ;;
            seed)
                # Content-tracking real-file copy (NOT a symlink) for first-run
                # entrypoints that must work before substrate is present. Unlike
                # a symlink it materializes a standalone file; like a symlink it
                # tracks upstream — refreshed when it drifts. Self-walk in the
                # upstream is skipped by the self-reference filter above (source
                # path == target path), so ariadne's own copy is never touched.
                create_seed "$upstream/$source" "$TARGET_DIR/$target"
                ;;
            copy)
                # `copy` action was retired in ariadne#38 — substrate is
                # symlink-only. Stale manifest entries get a warning so
                # operators notice; the entry itself is a no-op.
                printf "  ${YELLOW}warn${RESET}    \`copy %s\` ignored — copy action retired in ariadne#38; switch manifest to \`symlink\`\n" "$source"
                ;;
            merge)
                merge_settings "$upstream/$source" "$TARGET_DIR/$target"
                ;;
            touch)
                ensure_parent "$TARGET_DIR/$source"
                if [[ ! -f "$TARGET_DIR/$source" ]]; then
                    touch "$TARGET_DIR/$source"
                    printf "  ${GREEN}created${RESET} %s\n" "$source"
                fi
                ;;
            tool)
                # Declare a Go tool dependency from this upstream in the
                # target's go.mod. Adds (idempotently) the require + replace
                # + tool directives so `go mod vendor` can populate the
                # source for `make sdlc-build` etc.
                #
                # The single-arg form (`tool cmd/sdlc`) names the path
                # within the upstream module. Self-walk for the upstream's
                # own go.mod skips require+replace (would be circular) but
                # still adds the tool directive (so `go tool sdlc` works
                # locally in the upstream too).
                ensure_go_tool_dependency "$upstream" "$source"
                ;;
            *)
                printf "  ${YELLOW}unknown action: %s${RESET}\n" "$action"
                ;;
        esac
    done < "$manifest"
}

# ── ensure_go_tool_dependency — wire an upstream's tool ownership into target
# Despite the legacy name, this is the `tool <path>` manifest action. It has two
# jobs split by whether the target IS the tool's owner:
#   • Cross-target (a derivative consuming the owner's tool): declare the
#     substrate dependency on the owner in construct/deps (#60). No Go needed —
#     the build role moved to build-in-owner (#60 M2) and the peer graph to
#     construct/deps (#60 M1). Pre-#60 this stubbed a `<name>-construct` go.mod
#     with require+replace+tool; that's gone. An existing construct/go.mod is
#     left UNTOUCHED — dual-read keeps it valid until #60 M4 deletes it.
#   • Self-walk (the owner, e.g. ariadne): add a `tool` directive to the owner's
#     own root go.mod so `go tool <name>` works locally. Still Go; still app-level.
ensure_go_tool_dependency() {
    local upstream="$1"      # absolute owner path
    local tool_path="$2"     # relative path within owner module (e.g. cmd/sdlc)

    # Cross-target: declare the owner as a substrate dep in construct/deps.
    if [[ "$upstream" != "$TARGET_DIR" ]]; then
        local construct_dir="$TARGET_DIR/construct"
        local deps_file="$construct_dir/deps"
        local rel_root                       # repo-root-relative (../ariadne)
        # Pass paths as argv (not interpolated into the literal) so a path with
        # a quote can't break the expression — matches clone-data-deps.sh.
        rel_root=$(python3 -c 'import os,sys; print(os.path.relpath(sys.argv[1], sys.argv[2]))' "$upstream" "$TARGET_DIR" 2>/dev/null || echo "$upstream")
        mkdir -p "$construct_dir"
        if awk -v t="$rel_root" '$1=="substrate" && $2==t {f=1} END{exit !f}' "$deps_file" 2>/dev/null; then
            printf "  ${CYAN}present${RESET} substrate %s in construct/deps\n" "$rel_root"
        else
            printf 'substrate %s\n' "$rel_root" >> "$deps_file"
            printf "  ${GREEN}declared${RESET} substrate %s in construct/deps\n" "$rel_root"
        fi
        return 0
    fi

    # Self-walk (owner adds its own tool directive to its root go.mod). Needs Go.
    if ! command -v go >/dev/null 2>&1; then
        printf "  ${YELLOW}skipped${RESET} tool %s (go toolchain not on PATH)\n" "$tool_path"
        return 0
    fi
    if [[ ! -f "$TARGET_DIR/go.mod" ]]; then
        printf "  ${YELLOW}skipped${RESET} tool %s (self-walk; no go.mod in target)\n" "$tool_path"
        return 0
    fi
    local upstream_module
    upstream_module=$(awk '/^module / {print $2; exit}' "$TARGET_DIR/go.mod")
    ensure_go_directive_24 "$TARGET_DIR/go.mod"
    ( cd "$TARGET_DIR" && go mod edit -tool "${upstream_module}/${tool_path}" ) \
        && printf "  ${GREEN}declared${RESET} tool %s/%s in go.mod (self; tool only)\n" "$upstream_module" "$tool_path"
}

# Bump the go directive in <gomod> to at least 1.24 (needed for the
# `tool` directive). No-op if already >= 1.24 or if directive absent.
ensure_go_directive_24() {
    local gomod="$1"
    local current_go
    current_go=$(awk '/^go / {print $2; exit}' "$gomod")
    [[ -z "$current_go" ]] && return 0

    local cur_major cur_minor
    cur_major="${current_go%%.*}"
    cur_minor="${current_go#*.}"; cur_minor="${cur_minor%%.*}"
    if (( cur_major < 1 )) || { (( cur_major == 1 )) && (( cur_minor < 24 )); }; then
        ( cd "$(dirname "$gomod")" && go mod edit -go=1.24 ) || true
        printf "  ${YELLOW}bumped${RESET}  go directive in %s to 1.24 (required for tool directive)\n" "$(basename "$(dirname "$gomod")")/go.mod"
    fi
}

# Lib-only seam (#60): a test can `SETUP_LIB_ONLY=1 source setup.sh` to get the
# functions (e.g. ensure_go_tool_dependency) without running discovery/apply.
# No-op in normal execution. Mirrors .openshell/sandbox.sh's SANDBOX_LIB_ONLY.
[[ -n "${SETUP_LIB_ONLY:-}" ]] && return 0 2>/dev/null

# ── Process manifest(s) ───────────────────────────────────────────────────────
ANCESTORS=()
while IFS= read -r dir; do
    [[ -n "$dir" ]] && ANCESTORS+=("$dir")
done < <(discover_ancestors)

# Test seam (#50): print the discovered ancestor list (foundation-first) and
# exit, without applying anything. Lets discover-ancestors.test.sh assert the
# both-go.mods walk hermetically. No-op in normal runs.
if [[ -n "${SETUP_DISCOVER_ONLY:-}" ]]; then
    for upstream in "${ANCESTORS[@]+"${ANCESTORS[@]}"}"; do
        printf '%s\n' "$upstream"
    done
    exit 0
fi

if [[ ${#ANCESTORS[@]} -eq 0 ]]; then
    # No upstreams — this is ariadne (the top of the chain). Skip the
    # ancestor walk, but still run self-walk + post-processing below
    # (settings merge, gitignore, skills sync).
    printf "${YELLOW}No upstream layers found${RESET} — running self-walk + post-processing only.\n"
elif [[ ${#ANCESTORS[@]} -eq 1 ]]; then
    printf "${CYAN}Setup:${RESET} %s → %s\n" "${ANCESTORS[0]}" "$TARGET_DIR"
else
    printf "${CYAN}Setup:${RESET} %d upstream layer(s) → %s\n" "${#ANCESTORS[@]}" "$TARGET_DIR"
    for upstream in "${ANCESTORS[@]}"; do
        printf "  • %s\n" "$upstream"
    done
fi

if [[ ${#ANCESTORS[@]} -gt 0 ]]; then
    for upstream in "${ANCESTORS[@]}"; do
        walk_manifest "$upstream"
    done
fi

# After ancestors: walk target's own manifest if it has one. This lets a
# layer's manifest contain entries that ARE meaningful when applied to that
# layer itself (e.g., `symlink construct/skills/X .claude/skills/X` exposes
# a skill at the path Claude Code expects). The walk_manifest function's
# self-reference filter protects entries that would be tautological.
if [[ -f "$TARGET_DIR/construct/base.manifest" ]]; then
    walk_manifest "$TARGET_DIR"
fi
printf "\n"

# ── Create Makefile if missing ────────────────────────────────────────────────
if [[ ! -f "$TARGET_DIR/Makefile" ]]; then
    cat > "$TARGET_DIR/Makefile" << 'MAKEFILE'
# Canonical repo name from git remote
REPO_NAME := $(shell git remote get-url origin 2>/dev/null | sed 's|.*/||; s|\.git$$||')

# Issue/history paths (override before include if non-standard)
WF_ISSUES_DIR = workshop/issues
WF_HISTORY_DIR = workshop/history

# Include ariadne workflow targets
include Makefile.workflow

# Include local targets (repo-specific)
-include Makefile.local

.PHONY: help

help: help-workflow
	@true
MAKEFILE
    printf "  ${GREEN}created${RESET} Makefile\n"
fi

# Create .parley marker (enables parley.nvim repo mode)
if [[ ! -f "$TARGET_DIR/.parley" ]]; then
    touch "$TARGET_DIR/.parley"
    printf "  ${GREEN}created${RESET} .parley\n"
fi

# Create Makefile.local if missing
if [[ ! -f "$TARGET_DIR/Makefile.local" ]]; then
    cat > "$TARGET_DIR/Makefile.local" << 'MAKEFILE'
# Repo-specific Makefile targets.
# This file is included by Makefile — add your own targets here.
MAKEFILE
    printf "  ${GREEN}created${RESET} Makefile.local\n"
fi

# ── Create AGENTS.local.md if missing ─────────────────────────────────────────
if [[ ! -f "$TARGET_DIR/AGENTS.local.md" ]]; then
    cat > "$TARGET_DIR/AGENTS.local.md" << 'EOF'
# Local Extensions

## Repo-specific rules

<!-- Add repo-specific workflow rules, conventions, or overrides here. -->
<!-- This file is referenced by AGENTS.md via @AGENTS.local.md -->
EOF
    printf "  ${GREEN}created${RESET} AGENTS.local.md\n"
fi

# ── Ensure .gitignore entries ─────────────────────────────────────────────────
APPLY_GITIGNORE="$TARGET_DIR/construct/scripts/apply-gitignore-entries.sh"
if [[ ! -f "$APPLY_GITIGNORE" ]]; then
    APPLY_GITIGNORE="$SCRIPT_REAL/scripts/apply-gitignore-entries.sh"
fi
if [[ -f "$APPLY_GITIGNORE" ]]; then
    bash "$APPLY_GITIGNORE" "$TARGET_DIR" || true
fi

# ── Sync skill symlinks ───────────────────────────────────────────────────────
SYNC_SCRIPT="$TARGET_DIR/construct/scripts/sync-local-skills.sh"
if [[ -f "$SYNC_SCRIPT" ]]; then
    printf "\n"
    bash "$SYNC_SCRIPT" 2>&1 | while read -r line; do printf "  %s\n" "$line"; done
fi

printf "\n${GREEN}Done.${RESET} Review changes, then commit.\n"
