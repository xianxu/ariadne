#!/usr/bin/env bash
# test-lock.test.sh — drives scripts/test-lock.py and its Makefile.workflow wiring (#319, #341).
# Each case uses a private lock file, so it never touches the real machine-wide lock.
# Run:  bash scripts/test-lock.test.sh   (exit 0 = all pass)
set -u
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d "${TMPDIR:-/tmp}/test-lock.XXXXXX") || exit 1; trap 'kill $(jobs -p) 2>/dev/null; rm -rf "$tmp"' EXIT
export WF_TEST_LOCK_FILE="$tmp/lock"
unset WF_TEST_LOCK WF_TEST_LOCK_TIMEOUT WF_TEST_LOCK_HELD_BY
# The single-slot cases first; the two-slot default has its own block below.
export WF_TEST_SUITES=1
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
check "waiter names the holder" 'grep -q "  - $(basename "$(git rev-parse --show-toplevel)") (.*), make pid $a" "$tmp/b.err"'
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

# #341: by default two suites run at once; a third waits naming both holders and
# takes the slot freed first. A nested run is re-entrant even with a slot free.
export WF_TEST_LOCK_FILE="$tmp/two"; unset WF_TEST_SUITES
sleep 30 & p=$!; sleep 30 & q=$!; sleep 30 & r=$!; sleep 30 & s=$!
slot() { python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["slot"])' "$1"; }
check "first suite takes slot 0" '[ "$(lock $p)" = acquired ] && [ "$(slot "$tmp/two")" = 0 ]'
check "nested run with a slot free is re-entrant" '[ "$(WF_TEST_LOCK_HELD_BY=$p lock $s)" = reentrant ]'
check "second suite runs alongside on slot 1" '[ "$(WF_TEST_LOCK_TIMEOUT=1 lock $q)" = acquired ] && [ "$(slot "$tmp/two.1")" = 1 ]'
check "nested run under slot 1's holder is re-entrant" '[ "$(WF_TEST_LOCK_HELD_BY=$q lock $s)" = reentrant ]'
lock $r >"$tmp/r.out" 2>"$tmp/r.err" & third=$!
sleep 1.5
check "third suite waits" 'kill -0 $third 2>/dev/null'
check "the wait names both holders" 'grep -q "one of 2 full-suite slots" "$tmp/r.err" && grep -q "make pid $p" "$tmp/r.err" && grep -q "make pid $q" "$tmp/r.err"'
kill $p; wait $third
check "third takes the freed slot 0" '[ "$(cat "$tmp/r.out")" = acquired ] && [ "$(slot "$tmp/two")" = 0 ]'
check "a bad WF_TEST_SUITES says so and uses 2" '[ "$(WF_TEST_SUITES=x WF_TEST_LOCK_TIMEOUT=1 lock $s 2>"$tmp/x.err")" = timeout ] && grep -q "not a number; using 2" "$tmp/x.err"'

# #346: a weighted suite takes its slots all at once or not at all, waits in the
# queue holding none, and a heavy waiter is not starved by light newcomers.
export WF_TEST_LOCK_FILE="$tmp/w"; unset WF_TEST_SUITES
heavy() { python3 scripts/test-lock.py acquire --watch "$1" --slots 2; }
free_slot() { python3 -c 'import fcntl,os,sys; fd=os.open(sys.argv[1], os.O_RDWR); fcntl.flock(fd, fcntl.LOCK_EX|fcntl.LOCK_NB)' "$1" 2>/dev/null; }
release() { kill "$1"; sleep 0.3; }
sleep 60 & h1=$!; sleep 60 & l1=$!; sleep 60 & h2=$!; sleep 60 & l2=$!; sleep 60 & h3=$!
check "a weight-2 suite takes both slots" '[ "$(heavy $h1)" = acquired ] && grep -q "\"slots\": \[0, 1\]" "$tmp/w" && grep -q "\"slots\": \[0, 1\]" "$tmp/w.1"'
check "a light suite waits behind it, naming it once" '[ "$(WF_TEST_LOCK_TIMEOUT=1 lock $l1 2>"$tmp/w1.err")" = timeout ] && [ "$(grep -c "make pid $h1" "$tmp/w1.err")" = 2 ]'
release $h1
check "a light suite takes one slot" '[ "$(lock $l1)" = acquired ]'
heavy $h2 >"$tmp/h2.out" 2>"$tmp/h2.err" & hw=$!
sleep 1.5
check "a heavy suite waits for a light one" 'kill -0 $hw 2>/dev/null && grep -q "for 2 of 2 full-suite slots" "$tmp/h2.err"'
check "the waiter holds no slot (no partial hold)" 'free_slot "$tmp/w.1"'
check "a light newcomer queues behind the heavy waiter" '[ "$(WF_TEST_LOCK_TIMEOUT=1 lock $l2 2>"$tmp/l2.err")" = timeout ] && grep -q "next in line: .*make pid $h2" "$tmp/l2.err"'
heavy $h3 >"$tmp/h3.out" 2>/dev/null & hw3=$!
sleep 1.5
release $l1; wait $hw
check "the heavy waiter takes both slots once free" '[ "$(cat "$tmp/h2.out")" = acquired ] && kill -0 $hw3 2>/dev/null'
release $h2; wait $hw3
check "a second heavy waiter follows (no deadlock)" '[ "$(cat "$tmp/h3.out")" = acquired ]'
release $h3
check "the weight is clamped to the slot count" '[ "$(WF_TEST_SUITES=1 WF_TEST_LOCK_TIMEOUT=1 python3 scripts/test-lock.py acquire --watch $l2 --slots 5)" = acquired ]'
release $l2
# Makefile.workflow passes WF_TEST_WEIGHT; ariadne's Makefile declares 2, so its
# suite waits while a light suite holds one of the two slots.
sleep 60 & l3=$!
lock $l3 >/dev/null
check "make test waits for both slots (ariadne weighs 2)" '! WF_TEST_LOCK_TIMEOUT=1 make -n test >/dev/null 2>"$tmp/mw.err" && grep -q "for 2 of 2" "$tmp/mw.err"'
check "a weight-1 make test runs alongside" 'WF_TEST_WEIGHT=1 WF_TEST_LOCK_TIMEOUT=1 make -n test >/dev/null 2>&1'

[ $fails -eq 0 ] && echo "all passed" || { echo "$fails failed"; exit 1; }
