#!/usr/bin/env bash
# test-lock.test.sh — drives scripts/test-lock.py and its Makefile.workflow wiring (#319).
# Each case uses a private lock file, so it never touches the real machine-wide lock.
# Run:  bash scripts/test-lock.test.sh   (exit 0 = all pass)
set -u
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d "${TMPDIR:-/tmp}/test-lock.XXXXXX") || exit 1; trap 'kill $(jobs -p) 2>/dev/null; rm -rf "$tmp"' EXIT
export WF_TEST_LOCK_FILE="$tmp/lock"
unset WF_TEST_LOCK WF_TEST_LOCK_TIMEOUT WF_TEST_LOCK_HELD_BY
fails=0
check() { if eval "$2"; then echo "ok   $1"; else echo "FAIL $1"; fails=$((fails + 1)); fi; }
lock() { python3 scripts/test-lock.py acquire --watch "$1"; }

# Two runs serialize, and the waiter names the holder while it waits.
sleep 30 & a=$!
check "first run acquires" '[ "$(lock $a)" = acquired ]'
sleep 30 & b=$!
lock $b >"$tmp/b.out" 2>"$tmp/b.err" & waiter=$!
sleep 1.5
check "second run waits" 'kill -0 $waiter 2>/dev/null'
check "waiter names the holder" 'grep -q "held by ariadne (.*), make pid $a" "$tmp/b.err"'
kill $a; t0=$(date +%s)
wait $waiter
check "holder exit hands the lock over promptly" '[ $(( $(date +%s) - t0 )) -le 2 ] && [ "$(cat "$tmp/b.out")" = acquired ]'

# ^C or a hangup to make's process group reaches the watcher; make's exit, not
# the signal, releases the lock.
watcher=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["holder"])' "$WF_TEST_LOCK_FILE")
kill -INT "$watcher"; kill -HUP "$watcher"; sleep 0.2
check "watcher survives SIGINT and SIGHUP" 'kill -0 "$watcher" 2>/dev/null'

# Killing the holder process itself releases at once.
holder=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["holder"])' "$WF_TEST_LOCK_FILE")
kill -9 "$holder"; sleep 0.2
sleep 30 & c=$!
check "killed holder released the lock" '[ "$(WF_TEST_LOCK_TIMEOUT=1 lock $c)" = acquired ]'

# A nested run under the holder passes through; another run times out, reported as a wait.
check "same make is re-entrant" '[ "$(lock $c)" = reentrant ]'
sleep 30 & d=$!
check "nested make test is re-entrant" '[ "$(WF_TEST_LOCK_HELD_BY=$c lock $d)" = reentrant ]'
out=$(WF_TEST_LOCK_TIMEOUT=1 lock $d 2>"$tmp/d.err")
check "bounded wait times out" '[ "$out" = timeout ] && grep -q "a lock wait, not a test failure" "$tmp/d.err"'
check "escape hatch skips loudly" '[ "$(WF_TEST_LOCK=off lock $d 2>"$tmp/e.err")" = skipped ] && grep -q "WF_TEST_LOCK=off" "$tmp/e.err"'

# A record that isn't ours (valid JSON, wrong shape) reads as unidentified, not a crash.
printf '[1, 2]' >"$WF_TEST_LOCK_FILE"
out=$(WF_TEST_LOCK_TIMEOUT=1 lock $d 2>"$tmp/f.err")
check "foreign record reads as unidentified" '[ "$out" = timeout ] && grep -q "unidentified holder" "$tmp/f.err"'

# Makefile.workflow wiring: `make test` blocks on the held lock; other goals don't.
check "make test fails on a lock wait timeout" '! WF_TEST_LOCK_TIMEOUT=1 make -n test >/dev/null 2>"$tmp/m.err" && grep -q "full-suite lock not taken (timeout)" "$tmp/m.err"'
check "make help does not take the lock" 'WF_TEST_LOCK_TIMEOUT=1 make help >/dev/null 2>&1'
kill $c; sleep 0.3
check "make test takes the free lock" 'make -n test >/dev/null 2>&1'

[ $fails -eq 0 ] && echo "all passed" || { echo "$fails failed"; exit 1; }
