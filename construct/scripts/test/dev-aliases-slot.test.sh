#!/usr/bin/env bash
# #461: real Git layouts and executable sentinels, without builds or network.
set -euo pipefail
source_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
fixture="$(mktemp -d "${TMPDIR:-/tmp}/dev-aliases-slot.XXXXXX")" || exit 1
fixture="$(cd "${fixture:?}" && pwd -P)"
trap 'rm -rf "${fixture:?}"' EXIT
export COUCH_TREE=fixture
init_repo() {
  mkdir -p "$1/construct"
  git -C "$1" init -q -b main
  git -C "$1" -c user.name=T -c user.email=t@example.invalid commit -qm initial --allow-empty
}
owner() {
  init_repo "$1"
  mkdir -p "$1/cmd/sdlc" "$1/bin" "$1/construct/scripts"
  printf 'package main\nfunc main() {}\n' > "$1/cmd/sdlc/main.go"
  cp "$source_root/construct/dev-aliases.sh" "$1/construct/dev-aliases.sh"
  cp "$source_root/construct/scripts/lib-deps.sh" "$1/construct/scripts/lib-deps.sh"
  printf '#!/bin/sh\nprintf "%%s\\n" "%s"\n' "$2" > "$1/bin/sdlc"
  chmod +x "$1/bin/sdlc"
}
run_alias() {
  local dir="$1" script="$2" definitions
  definitions="$(cd "$dir" && bash "$script" --strict)"
  (cd "$dir" && eval "$definitions" && sdlc 'argument with spaces')
}
owner "$fixture/ariadne" central
git -C "$fixture/ariadne" remote add origin https://github.com/example/ariadne.git
init_repo "$fixture/pair"
git -C "$fixture/pair" worktree add -qb main-slot2 "$fixture/worktree/pair-slot2/pair"
owner "$fixture/worktree/pair-slot2/ariadne" stale-sibling
git -C "$fixture/worktree/pair-slot2/ariadne" remote add origin git@github.com:Example/Ariadne.git
host="$fixture/worktree/pair-slot2/pair"
mkdir -p "$host/construct"
printf 'substrate ../ariadne\n' > "$host/construct/deps"
ln -s ../../ariadne/construct/dev-aliases.sh "$host/construct/dev-aliases.sh"
mkdir -p "$fixture/stubs"
printf '#!/bin/sh\necho unexpected-build >> "%s"\nexit 99\n' "$fixture/builds" > "$fixture/stubs/go"
chmod +x "$fixture/stubs/go"
export PATH="$fixture/stubs:$PATH"
test "$(run_alias "$fixture/ariadne" construct/dev-aliases.sh)" = central
# Exercise both a slot-local script path and an operator-loaded central script.
test "$(run_alias "$host" construct/dev-aliases.sh)" = central
test "$(run_alias "$host" "$fixture/ariadne/construct/dev-aliases.sh")" = central
test "$(run_alias "$fixture/worktree/pair-slot2/ariadne" construct/dev-aliases.sh)" = central
test ! -e "$fixture/builds"
owner "$fixture/plain/ariadne" plain
test "$(run_alias "$fixture/plain/ariadne" construct/dev-aliases.sh)" = plain

# Canonicalizing N/2N private candidates must scan the selected cmd tree once.
real_find="$(command -v find)"
printf '#!/bin/sh\n[ "$1" != "%s/cmd" ] || echo scan >> "%s"\nexec "%s" "$@"\n' "$fixture/ariadne" "$fixture/scans" "$real_find" > "$fixture/stubs/find"
chmod +x "$fixture/stubs/find"
for count in 2 4; do
  for ((i=1; i<=count; i++)); do
    clone="$fixture/worktree/peer-slot$i/ariadne"
    if [ ! -d "$clone" ]; then
      owner "$clone" stale-sibling
      git -C "$clone" remote add origin https://github.com/example/ariadne.git
    fi
  done
  : > "$host/construct/deps"
  for ((i=1; i<=count; i++)); do printf 'substrate %s/worktree/peer-slot%s/ariadne\n' "$fixture" "$i" >> "$host/construct/deps"; done
  : > "$fixture/scans"
  (cd "$host" && bash construct/dev-aliases.sh --list >/dev/null)
  test "$(wc -l < "$fixture/scans" | tr -d ' ')" = 1
done
printf 'substrate ../ariadne\n' > "$host/construct/deps"
git -C "$fixture/ariadne" remote set-url origin https://github.com/example/different.git
if (cd "$host" && bash construct/dev-aliases.sh > "$fixture/defs" 2> "$fixture/error"); then
  echo 'FAIL: mismatched primary silently accepted' >&2; exit 1
fi
test ! -s "$fixture/defs"
grep -F 'primary owner' "$fixture/error"
mv "$fixture/ariadne" "$fixture/ariadne.original"
if (cd "$host" && bash construct/dev-aliases.sh > "$fixture/defs" 2> "$fixture/error"); then
  echo 'FAIL: absent primary silently accepted' >&2; exit 1
fi
test ! -s "$fixture/defs"
grep -F 'primary owner' "$fixture/error"
test ! -e "$fixture/builds"
echo 'PASS dev-aliases slot, primary, plain clone; no stale fallback or builds'
