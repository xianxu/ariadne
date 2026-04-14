#!/bin/bash
# The Construct — Emergency Rollback Script
# Reverts all managed skills and constitution files to a previous version.
# This script has ZERO dependencies on AI or the Construct skill.
# It must work when everything else is broken.
#
# Usage (from the project repo, via symlink):
#   ./construct/rollback.sh <version>        # e.g., ./construct/rollback.sh 0002
#   ./construct/rollback.sh <version-slug>   # e.g., ./construct/rollback.sh 0002-broken
#   ./construct/rollback.sh --list           # list available versions
#
# IMPORTANT: Run this through the symlink in the project repo, not directly
# from the ariadne repo. The script determines REPO_ROOT from its own location.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONSTRUCT_DIR="$SCRIPT_DIR"
REPO_ROOT="$(cd "$CONSTRUCT_DIR/.." && pwd)"
REPO_SKILLS_DIR="$REPO_ROOT/.claude/skills"
PERSONAL_SKILLS_DIR="$HOME/.claude/skills"
VERSIONS_DIR="$CONSTRUCT_DIR/versions"
CURRENT_FILE="$CONSTRUCT_DIR/current"

# List available versions
list_versions() {
    if [ ! -d "$VERSIONS_DIR" ] || [ -z "$(ls -A "$VERSIONS_DIR" 2>/dev/null)" ]; then
        echo "No versions available."
        exit 1
    fi

    echo "Available versions:"
    echo ""
    for dir in "$VERSIONS_DIR"/*/; do
        [ -d "$dir" ] || continue
        version_name="$(basename "$dir")"
        if [ -f "$CURRENT_FILE" ] && [ "$(cat "$CURRENT_FILE")" = "$version_name" ]; then
            echo "  $version_name  (current)"
        else
            echo "  $version_name"
        fi
    done
}

# Find version directory by prefix match
find_version_dir() {
    local target="$1"

    # Exact match first
    if [ -d "$VERSIONS_DIR/$target" ]; then
        echo "$VERSIONS_DIR/$target"
        return 0
    fi

    # Prefix match on numeric part (e.g., "0002" matches "0002-broken")
    local matches=()
    for dir in "$VERSIONS_DIR"/"${target}"*/; do
        [ -d "$dir" ] || continue
        matches+=("$dir")
    done

    if [ ${#matches[@]} -eq 0 ]; then
        echo "Error: No version matching '$target' found." >&2
        return 1
    elif [ ${#matches[@]} -gt 1 ]; then
        echo "Error: Multiple versions match '$target':" >&2
        for m in "${matches[@]}"; do
            echo "  $(basename "$m")" >&2
        done
        return 1
    fi

    echo "${matches[0]}"
}

# Rollback to a specific version
rollback() {
    local target="$1"
    local version_dir
    version_dir="$(find_version_dir "$target")"
    local version_name
    version_name="$(basename "$version_dir")"

    echo "Rolling back to version: $version_name"
    echo ""

    # Restore repo-scoped skills
    if [ -d "$version_dir/skills/repo" ]; then
        for skill_dir in "$version_dir"/skills/repo/*/; do
            [ -d "$skill_dir" ] || continue
            skill_name="$(basename "$skill_dir")"
            if [ -d "$REPO_SKILLS_DIR/$skill_name" ]; then
                echo "  Restoring repo skill: $skill_name"
                rm -rf "$REPO_SKILLS_DIR/$skill_name"
            else
                echo "  Adding repo skill: $skill_name"
            fi
            cp -r "$skill_dir" "$REPO_SKILLS_DIR/$skill_name"
        done
    fi

    # Restore personal-scoped skills
    if [ -d "$version_dir/skills/personal" ]; then
        for skill_dir in "$version_dir"/skills/personal/*/; do
            [ -d "$skill_dir" ] || continue
            skill_name="$(basename "$skill_dir")"
            if [ -d "$PERSONAL_SKILLS_DIR/$skill_name" ]; then
                echo "  Restoring personal skill: $skill_name"
                rm -rf "$PERSONAL_SKILLS_DIR/$skill_name"
            else
                echo "  Adding personal skill: $skill_name"
            fi
            cp -r "$skill_dir" "$PERSONAL_SKILLS_DIR/$skill_name"
        done
    fi

    # Restore constitution files
    if [ -f "$version_dir/constitution/AGENTS.md" ]; then
        echo "  Restoring AGENTS.md"
        cp "$version_dir/constitution/AGENTS.md" "$REPO_ROOT/AGENTS.md"
    fi
    if [ -f "$version_dir/constitution/CLAUDE.md" ]; then
        echo "  Restoring CLAUDE.md"
        cp "$version_dir/constitution/CLAUDE.md" "$REPO_ROOT/CLAUDE.md"
    fi

    # Update current marker
    echo "$version_name" > "$CURRENT_FILE"

    echo ""
    echo "Rolled back to $version_name successfully."
}

# Main
if [ $# -eq 0 ]; then
    echo "Usage: ./rollback.sh <version>    # revert to a version"
    echo "       ./rollback.sh --list       # list available versions"
    exit 1
fi

case "$1" in
    --list|-l)
        list_versions
        ;;
    --help|-h)
        echo "The Construct — Emergency Rollback"
        echo ""
        echo "Usage:"
        echo "  ./rollback.sh <version>     Revert skills and constitution to a previous version"
        echo "  ./rollback.sh --list        List available versions"
        echo ""
        echo "Examples:"
        echo "  ./rollback.sh 0002          Revert to version 0002"
        echo "  ./rollback.sh 0002-broken   Revert to version 0002-broken"
        echo "  ./rollback.sh 0001          Revert to the very first version"
        ;;
    *)
        rollback "$1"
        ;;
esac
