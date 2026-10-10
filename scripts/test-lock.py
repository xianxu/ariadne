#!/usr/bin/env python3
# test-lock.py — machine-wide lock so only one full test suite runs at a time (#319).
#
# Usage (Makefile.workflow runs this at parse time when `test` is a goal):
#   scripts/test-lock.py acquire --watch <make pid>
#
# Prints one word on stdout for make: acquired | reentrant | skipped | timeout.
# Waiting and skipping are reported on stderr, naming the holder.
#
# The lock is flock(2) on one per-user file. Once held, a forked watcher keeps the
# descriptor open until the watched make exits, so the kernel releases it on any
# exit, crash or kill. The file holds the holder's JSON record, rewritten per run.
#
# WF_TEST_LOCK=off skips the lock; WF_TEST_LOCK_TIMEOUT=<seconds> bounds the wait;
# WF_TEST_LOCK_FILE overrides the path (tests). WF_TEST_LOCK_HELD_BY marks a
# nested run under a locked make.

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


def lock_path():
    if os.environ.get("WF_TEST_LOCK_FILE"):
        return os.environ["WF_TEST_LOCK_FILE"]
    state = os.environ.get("XDG_STATE_HOME") or os.path.expanduser("~/.local/state")
    return os.path.join(state, "ariadne", "test-suite.lock")


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


def acquire(watch):
    if os.environ.get("WF_TEST_LOCK", "").lower() in ("off", "0", "no"):
        log("WF_TEST_LOCK=off: NOT taking the machine-wide full-suite lock; this run may collide with another")
        return "skipped"
    path = lock_path()
    os.makedirs(os.path.dirname(path), exist_ok=True)
    fd = os.open(path, os.O_RDWR | os.O_CREAT, 0o644)
    # Makefile.workflow exports the holder's make pid to its children, so a nested
    # `make test` inside a locked run (or make's own restart) passes through.
    held_by = int(os.environ.get("WF_TEST_LOCK_HELD_BY") or 0)
    limit = float(os.environ.get("WF_TEST_LOCK_TIMEOUT") or 0)
    start = last = time.time()
    first = True
    while True:
        try:
            fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
            break
        except BlockingIOError:
            pass
        rec = holder(fd)
        if rec and rec.get("pid") in (watch, held_by):
            return "reentrant"
        now = time.time()
        if first or now - last >= REPORT_EVERY:
            log(f"waiting {int(now - start)}s for the full-suite lock ({path}); held by {describe(rec)}")
            first, last = False, now
        if limit and now - start >= limit:
            log(f"gave up after {int(now - start)}s waiting for the lock held by {describe(rec)} — a lock wait, not a test failure")
            return "timeout"
        time.sleep(1)

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
    top = subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True).stdout.strip() or os.getcwd()
    rec = {"repo": os.path.basename(top), "worktree": top, "pid": watch, "holder": child,
           "started": datetime.now().isoformat(timespec="seconds")}
    os.ftruncate(fd, 0)
    os.pwrite(fd, json.dumps(rec).encode(), 0)
    return "acquired"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("cmd", choices=["acquire"])
    ap.add_argument("--watch", type=int, required=True, help="pid whose exit releases the lock (make)")
    print(acquire(ap.parse_args().watch))


if __name__ == "__main__":
    main()
