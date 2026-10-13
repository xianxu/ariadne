#!/usr/bin/env python3
# test-lock.py — machine-wide lock so at most WF_TEST_SUITES full test suites run at
# once (#319; two by default since #341).
#
# Usage (Makefile.workflow runs this at parse time when `test` is a goal):
#   scripts/test-lock.py acquire --watch <make pid> [--slots N]
#
# Prints one word on stdout for make: acquired | reentrant | skipped | timeout.
# Waiting and skipping are reported on stderr, naming the holder.
#
# The lock is flock(2) on one of N per-user slot files: test-suite.lock (slot 0,
# the file an older copy of this script contends for, so the cap holds across
# versions) and test-suite.lock.<k>. Once held, a forked watcher keeps the
# descriptor open until the watched make exits, so the kernel releases it on any
# exit, crash or kill. Each file holds its holder's JSON record, rewritten per run.
#
# A suite's weight (#346), --slots N (WF_TEST_WEIGHT in the repo's Makefile), is
# how many slots it takes: ariadne's sharded suite takes 2, so it runs alone at
# the default cap. N is clamped to the slot count. All N are taken in one scan or
# none are, so a waiter never holds part of its set. While waiting, a run holds
# the queue file (<slot 0>.queue); only the queue's head takes slots, so light
# suites cannot starve a heavy waiter, and the head holds no slot, so two heavy
# waiters cannot deadlock.
#
# WF_TEST_SUITES=<n> sets the slot count (default 2); WF_TEST_LOCK=off skips the
# lock; WF_TEST_LOCK_TIMEOUT=<seconds> bounds the wait; WF_TEST_LOCK_FILE
# overrides slot 0's path (tests). WF_TEST_LOCK_HELD_BY marks a nested run under
# a locked make.

import argparse
import fcntl
import json
import os
import select
import signal
import subprocess
import sys
import time
from datetime import datetime

REPORT_EVERY = 30


def log(msg):
    print(f"[test-lock] {msg}", file=sys.stderr, flush=True)


def lock_paths():
    base = os.environ.get("WF_TEST_LOCK_FILE")
    if not base:
        state = os.environ.get("XDG_STATE_HOME") or os.path.expanduser("~/.local/state")
        base = os.path.join(state, "ariadne", "test-suite.lock")
    try:
        n = max(1, int(os.environ.get("WF_TEST_SUITES") or 2))
    except ValueError:
        log(f"WF_TEST_SUITES={os.environ['WF_TEST_SUITES']!r} is not a number; using 2")
        n = 2
    return [base] + [f"{base}.{k}" for k in range(1, n)]


def scan(fds, want):
    """Takes `want` free slots at once (their indices), or none, and reads the busy slots' records."""
    free, busy = [], []
    for k, fd in enumerate(fds):
        try:
            fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            busy.append(holder(fd))
            continue
        free.append(k)
    taken = free[:want] if want and len(free) >= want else []
    for k in free:
        if k not in taken:
            fcntl.flock(fds[k], fcntl.LOCK_UN)
    return taken, busy


def try_lock(fd):
    try:
        fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
        return True
    except BlockingIOError:
        return False


def holders(busy):
    """One line per holder: a weighted suite holds several slots under one record."""
    seen, lines = set(), []
    for rec in busy:
        key = json.dumps(rec, sort_keys=True)
        if key not in seen:
            seen.add(key)
            lines.append(f"\n  - {describe(rec)}")
    return "".join(lines)


def holder(fd):
    """The holder's record, or None while it is unwritten or not one of ours."""
    try:
        rec = json.loads(os.pread(fd, 65536, 0) or b"null")
    except ValueError:
        return None
    keys = ("repo", "worktree", "pid", "started")
    return rec if isinstance(rec, dict) and all(k in rec for k in keys) else None


def describe(rec):
    if not rec:
        return "an unidentified holder (record not yet written, or not one of ours)"
    return f"{rec['repo']} ({rec['worktree']}), make pid {rec['pid']}, since {rec['started']}"


def wait_for_exit(pid):
    if hasattr(select, "kqueue"):
        kq = select.kqueue()
        ev = select.kevent(pid, select.KQ_FILTER_PROC, select.KQ_EV_ADD, select.KQ_NOTE_EXIT)
        try:
            kq.control([ev], 1, None)
            return
        except OSError:  # already gone
            return
    while True:
        try:
            os.kill(pid, 0)
        except ProcessLookupError:
            return
        except PermissionError:  # alive, owned by someone else
            pass
        time.sleep(0.5)


def acquire(watch, weight=1):
    if os.environ.get("WF_TEST_LOCK", "").lower() in ("off", "0", "no"):
        log("WF_TEST_LOCK=off: NOT taking the machine-wide full-suite lock; this run may collide with another")
        return "skipped"
    paths = lock_paths()
    os.makedirs(os.path.dirname(paths[0]), exist_ok=True)
    fds = [os.open(p, os.O_RDWR | os.O_CREAT, 0o644) for p in paths]
    queue = os.open(paths[0] + ".queue", os.O_RDWR | os.O_CREAT, 0o644)
    want = min(max(1, weight), len(fds))
    top = subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True).stdout.strip() or os.getcwd()
    rec = {"repo": os.path.basename(top), "worktree": top, "pid": watch,
           "started": datetime.now().isoformat(timespec="seconds")}
    # Makefile.workflow exports the holder's make pid to its children, so a nested
    # `make test` inside a locked run (or make's own restart) passes through.
    held_by = int(os.environ.get("WF_TEST_LOCK_HELD_BY") or 0)
    limit = float(os.environ.get("WF_TEST_LOCK_TIMEOUT") or 0)
    start = last = time.time()
    first, head = True, False
    while True:
        if not head and try_lock(queue):
            head = True
            os.ftruncate(queue, 0)
            os.pwrite(queue, json.dumps(rec).encode(), 0)
        # Only the queue's head takes slots; the others just read who holds them.
        taken, busy = scan(fds, want if head else 0)
        # A slot this make (or the make it runs under) already holds makes this a
        # nested run, even with another slot free.
        if any(r and r.get("pid") in (watch, held_by) for r in busy):
            for k in taken:
                fcntl.flock(fds[k], fcntl.LOCK_UN)
            return "reentrant"
        if taken:
            break
        held = holders(busy)
        if not head:
            held += f"\n  next in line: {describe(holder(queue))}"
        now = time.time()
        wanted = "one" if want == 1 else f"{want}"
        if first or now - last >= REPORT_EVERY:
            log(f"waiting {int(now - start)}s for {wanted} of {len(fds)} full-suite slots ({paths[0]}*); held by:{held}")
            first, last = False, now
        if limit and now - start >= limit:
            log(f"gave up after {int(now - start)}s waiting for a full-suite slot — a lock wait, not a test failure; held by:{held}")
            return "timeout"
        time.sleep(1)

    os.close(queue)  # leaving the queue lets the next waiter take its slots
    held_fds = [fds[k] for k in taken]
    for k, other in enumerate(fds):
        if k not in taken:
            os.close(other)

    if first is False:
        log(f"lock acquired after {int(time.time() - start)}s")
    child = os.fork()
    if child == 0:  # watcher: hold the inherited lock until make exits
        # Make's process group gets ^C and hangups too; make's exit, not the
        # signal, is what releases the lock.
        for sig in (signal.SIGINT, signal.SIGHUP):
            signal.signal(sig, signal.SIG_IGN)
        null = os.open(os.devnull, os.O_RDWR)
        for std in (0, 1, 2):
            os.dup2(null, std)
        wait_for_exit(watch)
        os._exit(0)
    rec.update(holder=child, slot=taken[0], slots=taken, started=datetime.now().isoformat(timespec="seconds"))
    for fd in held_fds:
        os.ftruncate(fd, 0)
        os.pwrite(fd, json.dumps(rec).encode(), 0)
    return "acquired"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("cmd", choices=["acquire"])
    ap.add_argument("--watch", type=int, required=True, help="pid whose exit releases the lock (make)")
    ap.add_argument("--slots", type=int, default=1, help="this suite's weight: slots it takes at once (#346)")
    args = ap.parse_args()
    print(acquire(args.watch, args.slots))


if __name__ == "__main__":
    main()
