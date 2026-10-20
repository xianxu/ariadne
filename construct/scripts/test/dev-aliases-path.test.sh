#!/usr/bin/env bash
# #465: pure path splitting conforms to platform tools without per-owner execs.
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
generator="$root/construct/dev-aliases.sh"
# Load only the production pure helper and owner-location rule, not the CLI.
eval "$(sed -n '/^path_parts() {/,/^}/p; /^primary_owner() {/,/^}/p' "$generator")"
if ! declare -F path_parts >/dev/null; then
    echo 'FAIL: missing pure path splitter' >&2
    exit 1
fi
for input in '' / // /// a a/ a// /a //a ///a //a//b/// a//b ./ ../ 'one two/three four/' '-leading' $'one\ntwo/three\n'; do
    expected_dir="$(dirname -- "$input")"
    expected_base="$(basename -- "$input")"
    path_parts "$input"
    if [ "$path_dir" != "$expected_dir" ] || [ "$path_base" != "$expected_base" ]; then
        printf 'FAIL path <%s>: dir <%s> vs <%s>; base <%s> vs <%s>\n' "$input" "$path_dir" "$expected_dir" "$path_base" "$expected_base" >&2
        exit 1
    fi
done
fixture="$(mktemp -d "${TMPDIR:-/tmp}/alias-path-count.XXXXXX")" || exit 1
trap 'rm -rf "${fixture:?}"' EXIT
# Count actual calls, including calls in command substitutions. Identity/layout
# behavior is covered by the real Git fixtures; count plain and slot owners.
dirname() { printf 'dirname\n' >> "$fixture/calls"; command dirname "$@"; }
basename() { printf 'basename\n' >> "$fixture/calls"; command basename "$@"; }
mkdir -p "$fixture/ariadne/construct" "$fixture/worktree/peer-slot1/ariadne/construct"
repo_identity() { printf 'origin:github.com/example/ariadne\n'; }
export fixture
export -f dirname basename
mkdir -p "$fixture/caller/construct" "$fixture/owners"
for count in 2 4; do
    for ((i=0; i<count; i++)); do
        mkdir -p "$fixture/owners/repo-$i-x/construct" "$fixture/owners/repo-$i-x/cmd/tool_$i"
        printf 'package main\n' > "$fixture/owners/repo-$i-x/cmd/tool_$i/main.go"
    done
    : > "$fixture/calls"
    for ((i=0; i<count; i++)); do
        test "$(primary_owner "$fixture/plain")" = "$fixture/plain"
        test "$(primary_owner "$fixture/worktree/peer-slot1/ariadne")" = "$fixture/ariadne"
    done
    (cd "$fixture/caller" && bash "$generator" --workspace "$fixture/owners" --list >/dev/null)
    if [ -s "$fixture/calls" ]; then
        echo "FAIL: $count owners spawned path utilities" >&2
        exit 1
    fi
done
echo 'PASS path conformance; N/2N owner resolution and CLI spawn zero dirname/basename commands'
