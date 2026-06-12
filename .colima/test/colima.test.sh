#!/usr/bin/env bash
# colima.test.sh — drives .colima/colima.sh against a FAKE `colima` binary
# (records argv; canned status/list via env), asserting the command lines
# each verb builds. No real VM is started.
set -euo pipefail
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
LOG="$WORK/colima.log"

# Fake colima: append argv to LOG; emulate status/list from env.
mkdir -p "$WORK/bin"
cat > "$WORK/bin/colima" <<EOF
#!/usr/bin/env bash
echo "\$@" >> "$LOG"
case "\$1" in
  status) [ "\${FAKE_RUNNING:-0}" = 1 ] && exit 0 || exit 1 ;;
  list)   [ "\${FAKE_EXISTS:-0}" = 1 ] && echo '{"name":"p-test","status":"Stopped"}'; exit 0 ;;
  ssh)    exit 0 ;;   # non-interactive in test
  *)      exit 0 ;;
esac
EOF
chmod +x "$WORK/bin/colima"
export PATH="$WORK/bin:$PATH"

run() { : > "$LOG"; "$ROOT/.colima/colima.sh" "$@"; cat "$LOG"; }
assert_has()  { grep -q -- "$1" "$LOG" || { echo "FAIL: expected '$1' in:"; cat "$LOG"; exit 1; }; }
assert_none() { grep -q -- "$1" "$LOG" && { echo "FAIL: did not expect '$1' in:"; cat "$LOG"; exit 1; } || true; }

# up, profile stopped → starts with mount + arch + ssh
FAKE_RUNNING=0 run up p-test /Users/me/ws/repo /Users/me/ws >/dev/null
assert_has "start -p p-test"
assert_has "--mount /Users/me/ws:w"
assert_has "ssh -p p-test"

# up, profile already running → no start, just ssh (idempotent)
FAKE_RUNNING=1 run up p-test /Users/me/ws/repo /Users/me/ws >/dev/null
assert_none "start -p p-test"
assert_has  "ssh -p p-test"

# up (fresh) also pushes vm-rc.sh + runs vm-setup.sh before the interactive ssh
FAKE_RUNNING=0 run up p-test /Users/me/ws/repo /Users/me/ws >/dev/null
assert_has "colima-vm-rc.sh"                              # vm-rc push
assert_has "bash -s -- /Users/me/ws/repo /Users/me/ws"   # vm-setup invocation

# up on an ALREADY-RUNNING profile skips the heavy vm-setup (still pushes rc)
FAKE_RUNNING=1 run up p-test /Users/me/ws/repo /Users/me/ws >/dev/null
assert_has  "colima-vm-rc.sh"
assert_none "bash -s -- /Users/me/ws/repo /Users/me/ws"

# stop, running → stops
FAKE_RUNNING=1 run stop p-test >/dev/null
assert_has "stop -p p-test"

# stop, not running → no stop (idempotent)
FAKE_RUNNING=0 run stop p-test >/dev/null
assert_none "stop -p p-test"

# clean, exists → delete -f
FAKE_RUNNING=0 FAKE_EXISTS=1 run clean p-test >/dev/null
assert_has "delete -p p-test -f"

# clean, ABSENT → no delete (proves the _exists gate isn't a tautology: an
# empty `list -j` must suppress the delete)
FAKE_RUNNING=0 FAKE_EXISTS=0 run clean p-test >/dev/null
assert_none "delete -p p-test -f"

# amd64 → rosetta flag
FAKE_RUNNING=0 COLIMA_ARCH=x86_64 run up p-test /Users/me/ws/repo /Users/me/ws >/dev/null
assert_has "--vz-rosetta"

# gui → streams vnc-setup.sh with geometry + password positional args
FAKE_RUNNING=0 run gui p-test /Users/me/ws/repo /Users/me/ws >/dev/null
assert_has "bash -s -- 1600x1000 colima"

# up/gui with missing args → usage guard (exit 2), not an unbound-variable crash
set +e; "$ROOT/.colima/colima.sh" up p-test >/dev/null 2>&1; rc=$?; set -e
[ "$rc" -eq 2 ] || { echo "FAIL: 'up' with missing args should exit 2 (got $rc)"; exit 1; }

echo "PASS: .colima/colima.sh"
