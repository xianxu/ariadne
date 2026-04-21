#!/usr/bin/env bash
# Ariadne Base Layer Setup
# Bootstraps a target repo with ariadne's portable fragments.
#
# Usage:
#   From target repo:  ./ariadne-refresh.sh  (symlinked)
#   Or directly:       ../ariadne/construct/setup.sh
#
# Idempotent — safe to re-run for updates.
set -euo pipefail

# ── Resolve paths ───────────────────────────���───────────────────────────────��─
# Follow symlinks to find real ariadne location (construct/ dir)
SCRIPT_REAL="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || realpath "${BASH_SOURCE[0]}")")" && pwd)"
ARIADNE_DIR="$(dirname "$SCRIPT_REAL")"
TARGET_DIR="$(pwd)"
MANIFEST="$SCRIPT_REAL/base.manifest"

if [[ "$ARIADNE_DIR" == "$TARGET_DIR" ]]; then
    echo "Error: run this from the TARGET repo, not from ariadne itself."
    echo "Usage: cd /path/to/your-repo && ./ariadne-refresh.sh"
    exit 1
fi

if [[ ! -f "$MANIFEST" ]]; then
    echo "Error: base.manifest not found at $MANIFEST"
    exit 1
fi

# ── Colors ────────────────────────────────────────────────────────────────────
GREEN='\033[1;32m'
YELLOW='\033[1;33m'
CYAN='\033[1;36m'
RESET='\033[0m'

# ── Helpers ───────────────────────────────────────────────────────────────────
rel_path() {
    # Compute relative path from target to ariadne for symlinks
    python3 -c "import os.path; print(os.path.relpath('$1', '$2'))"
}

ensure_parent() {
    local path="$1"
    local parent
    parent=$(dirname "$path")
    [[ -d "$parent" ]] || mkdir -p "$parent"
}

create_symlink() {
    local src="$1"  # absolute path in ariadne
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
        printf "  ${YELLOW}skipped${RESET} %s (file exists, not a symlink)\n" "${dst#$TARGET_DIR/}"
        return 0
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
    # Add .gitkeep so empty dirs are tracked
    touch "$dir/.gitkeep"
    printf "  ${GREEN}created${RESET} %s/\n" "${dir#$TARGET_DIR/}"
}

merge_settings() {
    local base_file="$1"   # ariadne's settings.ariadne.json
    local target_file="$2" # target's settings.json (generated, gitignored)

    ensure_parent "$target_file"

    # Remove old symlink if present (from previous setup.sh versions)
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

# ── Process manifest ──────────────────────────────────────────────────────────
printf "${CYAN}Ariadne setup: %s → %s${RESET}\n" "$ARIADNE_DIR" "$TARGET_DIR"
printf "\n"

while IFS= read -r line; do
    # Skip comments and empty lines
    [[ "$line" =~ ^[[:space:]]*# ]] && continue
    [[ -z "${line// /}" ]] && continue

    # Parse: action source [target]
    read -r action source target <<< "$line"
    target="${target:-$source}"

    case "$action" in
        symlink)
            create_symlink "$ARIADNE_DIR/$source" "$TARGET_DIR/$target"
            ;;
        scaffold)
            create_scaffold "$TARGET_DIR/$target"
            ;;
        copy)
            ensure_parent "$TARGET_DIR/$target"
            if [[ ! -f "$TARGET_DIR/$target" ]]; then
                cp "$ARIADNE_DIR/$source" "$TARGET_DIR/$target"
                printf "  ${GREEN}copied${RESET}  %s\n" "$target"
            else
                printf "  ${YELLOW}skipped${RESET} %s (already exists)\n" "$target"
            fi
            ;;
        merge)
            merge_settings "$ARIADNE_DIR/$source" "$TARGET_DIR/$target"
            ;;
        *)
            printf "  ${YELLOW}unknown action: %s${RESET}\n" "$action"
            ;;
    esac
done < "$MANIFEST"

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

printf "\n${GREEN}Done.${RESET} Review changes, then commit.\n"
