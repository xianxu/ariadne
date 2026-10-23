#!/usr/bin/env bash
# #573: dev-aliases never defines a function that shadows a bash or zsh builtin,
# a reserved word, or a system command; ordinary names still get theirs.
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
generator="$root/construct/dev-aliases.sh"
fixture="$(mktemp -d "${TMPDIR:-/tmp}/dev-aliases-shadow.XXXXXX")" || exit 1
fixture="$(cd "${fixture:?}" && pwd -P)"
trap 'rm -rf "${fixture:?}"' EXIT
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
have_zsh=0
command -v zsh >/dev/null 2>&1 && have_zsh=1

# The pure rule, loaded alone from the production script.
eval "$(sed -n '/^shadow_reason() {/,/^}/p' "$generator")"
while IFS='|' read -r name bash_kind zsh_kind system want; do
	shadow_reason "$name" "$bash_kind" "$zsh_kind" "$system"
	[ "$reason" = "$want" ] || fail "shadow_reason $name/$bash_kind/$zsh_kind/$system: <$reason>, want <$want>"
done <<'EOF'
echo|builtin|builtin||the bash builtin echo
if|keyword|reserved||the bash keyword if
print|none|builtin||the zsh builtin print
foreach|none|reserved||the zsh reserved word foreach
ls|file|command|/usr/bin/ls|the system command /usr/bin/ls
sdlc|file|command||
oktool|none|none||
warn|function|none||
EOF

# A workspace whose one repository owns builtin, reserved, system and ordinary
# names. HOME and PATH are the fixture's: sys/ stands for a system directory.
ws="$fixture/ws"
home="$fixture/home"
mkdir -p "$ws/tools/construct" "$ws/tools/bin" "$home/bin" "$fixture/sys" "$fixture/caller" "$fixture/stubs"
command_dir() { mkdir -p "$ws/tools/cmd/$1"; printf 'package main\n' > "$ws/tools/cmd/$1/main.go"; }
refused="echo test ls sysdup lsalias"
allowed="oktool homedup linked"
[ "$have_zsh" -eq 0 ] || refused="$refused print repeat"  # zsh-only builtin and reserved word
for name in $refused $allowed; do command_dir "$name"; done
executable() { printf '#!/bin/sh\n' > "$1"; chmod +x "$1"; }
executable "$fixture/sys/sysdup"              # outside workspace and HOME
executable "$home/bin/homedup"                # the user's own install
executable "$ws/tools/bin/linked"
ln -s "$ws/tools/bin/linked" "$fixture/sys/linked"  # resolves into the workspace
ln -s /bin/ls "$fixture/sys/lsalias"                # resolves to a system command
ln -s /bin/echo "$fixture/sys/echo"  # refused by the shells: its link is never followed
# Counting stubs: zsh once per run, readlink once per symlink hop.
for tool in zsh readlink; do
	real="$(command -v "$tool" || true)"
	[ -n "$real" ] || continue
	printf '#!/bin/sh\necho %s >> "%s/calls"\nexec %s "$@"\n' "$tool" "$fixture" "$real" > "$fixture/stubs/$tool"
	chmod +x "$fixture/stubs/$tool"
done
test_path="$fixture/stubs:$home/bin:$fixture/sys:/usr/bin:/bin"
# Host-dependent expectations come from the generator's own resolution
# (follow_links, then cd -P), never a literal layout: merged-usr Linux finds
# /usr/bin/ls first and resolves /bin to /usr/bin, Debian's /usr/bin/print can
# be a link, busybox's /bin/ls is one. The hops file counts its readlinks.
eval "$(sed -n '/^path_parts() {/,/^}/p; /^follow_links() {/,/^}/p' "$generator")"
readlink() { echo hop >> "$fixture/hops"; command readlink "$@"; }
resolved() { follow_links "$1"; path_parts "$link_path"; (cd -P "$path_dir" && printf '%s/%s\n' "$PWD" "$path_base"); }
ls_want="$(resolved "$(PATH="$test_path" type -P ls)")"
lsalias_want="$(resolved "$fixture/sys/lsalias")"
# The readlinks a run spends: one per hop of each first PATH match outside the
# workspace and HOME, for the names no shell refuses.
want_hops() {
	local n p
	: > "$fixture/hops"
	for n in "$@"; do
		p="$(PATH="$test_path" type -P "$n" || true)"
		case "$p" in ''|"$ws"/*|"$home"/*) continue ;; esac
		follow_links "$p"
	done
	grep -c . "$fixture/hops" || true
}
run() {
	(cd "$fixture/caller" && HOME="$home" PATH="$test_path" \
		bash "$generator" --workspace "$ws" "$@")
}

defs="$(run 2>"$fixture/err")"
for name in $refused; do
	n="$(grep -c "^dev-aliases: not defining $name (cmd/$name in $ws/tools): it would shadow " "$fixture/err" || true)"
	[ "$n" -eq 1 ] || fail "$name: $n warnings, want 1: $(cat "$fixture/err")"
done
[ "$(grep -c . "$fixture/err")" -eq "$(printf '%s\n' $refused | grep -c .)" ] || fail "extra warnings: $(cat "$fixture/err")"
grep -q "not defining ls .*the system command $ls_want;" "$fixture/err" || fail "ls reason, want $ls_want: $(cat "$fixture/err")"
grep -q "not defining lsalias .*the system command $lsalias_want;" "$fixture/err" || fail "lsalias reason, want $lsalias_want"
grep -q "not defining sysdup .*the system command $fixture/sys/sysdup;" "$fixture/err" || fail "sysdup reason"
if [ "$have_zsh" -eq 1 ]; then
	grep -q "not defining print .*the zsh builtin print;" "$fixture/err" || fail "print reason"
	grep -q "not defining repeat .*the zsh reserved word repeat;" "$fixture/err" || fail "repeat reason"
fi
listed="$(run --list 2>/dev/null | cut -f1 | tr '\n' ' ')"
[ "$listed" = "homedup linked oktool " ] || fail "--list: <$listed>"

# Sourced in each shell, builtins and commands keep their meaning.
bash_kinds="$(bash -c 'eval "$1"; for n in echo test ls oktool homedup linked; do printf "%s=%s " "$n" "$(type -t "$n")"; done' _ "$defs")"
[ "$bash_kinds" = "echo=builtin test=builtin ls=file oktool=function homedup=function linked=function " ] || fail "bash: $bash_kinds"
if [ "$have_zsh" -eq 1 ]; then
	zsh_kinds="$(zsh -f -c 'eval "$1"; whence -w echo test print repeat ls oktool homedup linked' _ "$defs" | tr '\n' ' ')"
	want="echo: builtin test: builtin print: builtin repeat: reserved ls: command oktool: function homedup: function linked: function "
	[ "$zsh_kinds" = "$want" ] || fail "zsh: $zsh_kinds"
else
	echo 'SKIP zsh sourcing: zsh not installed'
fi

# N and 2N more ordinary names: one zsh call per run, and readlink only for the
# hops of outside PATH matches, never per name.
for count in 8 16; do
	for ((i=0; i<count; i++)); do command_dir "tool_$i"; done
	: > "$fixture/calls"
	run >/dev/null 2>&1
	zsh_calls="$(grep -c '^zsh$' "$fixture/calls" || true)"
	readlink_calls="$(grep -c '^readlink$' "$fixture/calls" || true)"
	[ "$zsh_calls" -eq "$have_zsh" ] || fail "$count names: $zsh_calls zsh calls, want $have_zsh"
	tools=""
	for ((i=0; i<count; i++)); do tools="$tools tool_$i"; done
	# shellcheck disable=SC2086
	want="$(want_hops ls sysdup lsalias linked oktool homedup $tools)"
	[ "$want" -ge 2 ] || fail "fixture links not counted: want_hops $want"
	[ "$readlink_calls" -eq "$want" ] || fail "$count names: $readlink_calls readlink calls, want $want"
done
echo 'PASS shadowing names refused with one warning each; bash and zsh keep builtins; zsh once, readlink per hop'
