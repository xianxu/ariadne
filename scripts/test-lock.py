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
# none are, so a waiter never holds part of its set. Admission is FIFO (#357): each
# waiter takes a numbered ticket in <slot 0>.tickets/ and holds a flock on it while
# it waits, so the lowest live ticket is the oldest waiter, and a ticket whose lock
# the kernel released (its waiter died) is pruned rather than blocking the line.
# Active fleet-blocking grants take priority; each class stays FIFO. The head takes the queue
# file (<slot 0>.queue), the head token older copies of this script use, so a
# mixed fleet respects existing owners. Each failed weighted scan releases the queue
# token so admission is reelected against current grant evidence. Only the head takes slots, so light suites never
# overtake it, and the head holds no slot, so two heavy waiters cannot deadlock. A
# live waiter keeps its place even when stopped (^Z): the line waits for it, and
# the waiters behind it name it as next in line.
#
# WF_TEST_SUITES=<n> sets the slot count (default 2); WF_TEST_LOCK=off skips the
# lock; WF_TEST_LOCK_TIMEOUT=<seconds> bounds the wait; WF_TEST_LOCK_FILE
# overrides slot 0's path (tests). WF_TEST_LOCK_HELD_BY marks a nested run under
# a locked make. WF_TEST_LOCK_OWNER names the checkout a run belongs to when it
# runs elsewhere (sdlc local-check's temp worktree).

import argparse
import contextlib
import fcntl
import json
import os
import select
import signal
import selectors
import tempfile
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


def live_tickets(tdir):
    """The live tickets, oldest first, as (number, path); prunes dead waiters' tickets.

    Call under the allocation mutex: a ticket is created and locked under it, so an
    unlocked ticket seen here belongs to a waiter that has died."""
    live = []
    for name in os.listdir(tdir):
        if not name.isdigit():
            continue
        path = os.path.join(tdir, name)
        try:
            fd = os.open(path, os.O_RDWR)
        except FileNotFoundError:
            continue
        try:
            fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            live.append((int(name), path))  # its waiter holds it
        else:
            try:
                os.unlink(path)  # its waiter is gone: skip it for good
            except FileNotFoundError:
                pass
        finally:
            os.close(fd)
    return sorted(live)


class Tickets:
    """This waiter's place in the FIFO line (#357)."""

    def __init__(self, base, rec):
        self.dir = base + ".tickets"
        os.makedirs(self.dir, exist_ok=True)
        self.mutex = os.open(os.path.join(self.dir, "alloc"), os.O_RDWR | os.O_CREAT, 0o644)
        with self.locked():
            live = live_tickets(self.dir)
            self.number = (live[-1][0] if live else 0) + 1
            self.path = os.path.join(self.dir, f"{self.number:012d}")
            self.fd = os.open(self.path, os.O_RDWR | os.O_CREAT | os.O_EXCL, 0o644)
            fcntl.flock(self.fd, fcntl.LOCK_EX)
            os.pwrite(self.fd, json.dumps(rec).encode(), 0)

    @contextlib.contextmanager
    def locked(self):
        fcntl.flock(self.mutex, fcntl.LOCK_EX)
        try:
            yield
        finally:
            fcntl.flock(self.mutex, fcntl.LOCK_UN)

    def admission(self, fds, queue, want, evidence):
        """Reelect and scan atomically; a waiting head never reserves the token."""
        with self.locked():
            records = []
            for number, path in live_tickets(self.dir):
                try:
                    with open(path, "rb") as stream:
                        records.append((number, holder(stream.fileno())))
                except FileNotFoundError:
                    # Owners (including older helpers) release outside alloc.
                    # A snapshot is not a lease on another waiter's pathname.
                    continue
            now = time.time()
            deadlines = evidence.deadlines(records, now)
            order = effective_order(records, deadlines, time.time())
            index = next(i for i, (n, _) in enumerate(order) if n == self.number)
            ahead = [r for _, r in order[:index]]
            head = index == 0 and try_lock(queue)
            if head:
                try:
                    os.ftruncate(queue, 0)
                    os.pwrite(queue, json.dumps(order[0][1]).encode(), 0)
                    taken, busy = scan(fds, want)
                finally:
                    fcntl.flock(queue, fcntl.LOCK_UN)
            else:
                taken, busy = scan(fds, 0)
            deadline = deadlines.get(subject_key(order[index][1]), 0)
            return taken, busy, ahead, head, deadline

    def release(self):
        try:
            os.unlink(self.path)
        except FileNotFoundError:
            pass
        os.close(self.fd)
        os.close(self.mutex)


# The provider is selected by the caller/supplying substrate, never by a ticket.
# Tickets contain identity only. Go owns grant validity; Python owns admission.
PROTOCOL_LIMIT = 64 << 10
PROVIDER_TIMEOUT = 2


def query_priority(provider, request, describe=None):
    """Bounded read-only provider invocation; terminate/reap the whole group."""
    data = json.dumps(request).encode()
    if len(data) > PROTOCOL_LIMIT:
        raise ValueError("priority request exceeds 64 KiB")
    argv = [provider, "landing", "suite-priority"]
    if describe is not None:
        argv += ["--describe", "--repo", describe]
    # A file-backed stdin avoids deadlock if an unavailable provider won't read.
    with tempfile.TemporaryFile(dir=os.environ.get("TMPDIR") or "/tmp") as source:
        source.write(data)
        source.seek(0)
        process = subprocess.Popen(argv, stdin=source, stdout=subprocess.PIPE,
                                   stderr=subprocess.DEVNULL, start_new_session=True)
        try:
            output = bytearray()
            end = time.monotonic() + PROVIDER_TIMEOUT
            with selectors.DefaultSelector() as ready:
                ready.register(process.stdout, selectors.EVENT_READ)
                while True:
                    remaining = end - time.monotonic()
                    if remaining <= 0:
                        raise ValueError("priority provider timed out")
                    if not ready.select(remaining):
                        raise ValueError("priority provider timed out")
                    chunk = os.read(process.stdout.fileno(), 4096)
                    if not chunk:
                        break
                    output.extend(chunk)
                    if len(output) > PROTOCOL_LIMIT:
                        raise ValueError("priority response exceeds 64 KiB")
            if process.wait(timeout=max(.001, end - time.monotonic())) != 0:
                raise ValueError("priority provider failed or is too old")
            result = json.loads(output)
            if not isinstance(result, dict) or type(result.get("version")) is not int or result["version"] != 1:
                raise ValueError("unsupported priority protocol")
            return result
        finally:
            with contextlib.suppress(ProcessLookupError):
                os.killpg(process.pid, signal.SIGKILL)
            process.wait()
            process.stdout.close()


def valid_subject(subject):
    """Validate untrusted persisted identity before any filesystem/protocol use."""
    if not isinstance(subject, dict):
        return False
    common, issue = subject.get("common_dir"), subject.get("issue")
    if (not isinstance(common, str) or not os.path.isabs(common) or
            "\x00" in common or common.startswith("//") or os.path.normpath(common) != common or
            type(issue) is not int or issue < 0):
        return False
    try:
        return len(os.fsencode(common)) <= 4096
    except UnicodeError:
        return False


def subject_key(rec):
    subject = rec.get("subject") if isinstance(rec, dict) else None
    if not valid_subject(subject) or subject["issue"] == 0:
        return None
    return subject["common_dir"], subject["issue"]


def effective_order(records, deadlines, now):
    """Stable partition of numeric FIFO; expired evidence confers nothing."""
    priority, normal = [], []
    for item in records:
        (priority if deadlines.get(subject_key(item[1]), 0) > now else normal).append(item)
    return priority + normal


class GrantEvidence:
    def __init__(self, provider, query=query_priority, stat=os.stat):
        self.provider, self.query, self.stat = provider, query, stat
        self.cache = {}
        self.warned = set()

    def warn(self, reason):
        if reason not in self.warned:
            log(f"priority unavailable: {reason}; using normal FIFO")
            self.warned.add(reason)

    def deadlines(self, records, now):
        requested = {}
        for _, record in records:
            subject = record.get("subject") if isinstance(record, dict) else None
            if subject is not None and not valid_subject(subject):
                self.warn("invalid persisted suite subject")
            key = subject_key(record)
            if key:
                requested.setdefault(key[0], set()).add(key[1])
        stale, signatures = {}, {}
        result = {}
        for common, issues in requested.items():
            path = os.path.join(common, "sdlc", "landing", "grants.jsonl")
            try:
                info = self.stat(path)
                signature = (info.st_ino, info.st_mtime_ns, info.st_size)
            except FileNotFoundError:
                signature = None
            except (OSError, ValueError) as error:
                self.warn(str(error))
                self.cache.pop(common, None)
                continue
            signatures[common] = signature
            cached = self.cache.get(common)
            if (cached and cached[0] == signature and issues <= cached[1] and
                    (not cached[2] or now < cached[2])):
                result.update({(common, int(n)): end for n, end in cached[3].items()})
            else:
                stale[common] = issues
        if stale:
            subjects = [{"common_dir": common, "issue": issue}
                        for common, issues in stale.items() for issue in sorted(issues)]
            try:
                response = self.query(self.provider, {"version": 1, "subjects": subjects})
                sources = response.get("sources")
                if not isinstance(sources, list) or len(sources) != len(stale):
                    raise ValueError("invalid priority sources")
                validated = {}
                for source in sources:
                    common = source.get("common_dir")
                    ends, refresh = source.get("deadlines"), source.get("refresh_at")
                    if common not in stale or common in validated:
                        raise ValueError("unexpected priority source")
                    if source.get("error"):
                        raise ValueError(str(source["error"]))
                    if not isinstance(ends, dict) or type(refresh) not in (int, float):
                        raise ValueError("invalid priority deadlines")
                    for n, end in ends.items():
                        if (not n.isdigit() or int(n) not in stale[common] or
                                type(end) not in (int, float) or not float('-inf') < end < float('inf')):
                            raise ValueError("invalid priority deadline")
                    if not 0 <= refresh < float('inf'):
                        raise ValueError("invalid priority refresh boundary")
                    validated[common] = (ends, refresh)
                for common, (ends, refresh) in validated.items():
                    self.cache[common] = (signatures[common], stale[common], refresh, ends)
                    result.update({(common, int(n)): end for n, end in ends.items()})
            except (OSError, ValueError, subprocess.SubprocessError, TypeError, AttributeError) as error:
                self.warn(str(error))
                for common in stale:
                    self.cache.pop(common, None)
        return result


def priority_subject(provider, top, evidence):
    try:
        raw = os.environ.get("WF_TEST_PRIORITY_SUBJECT")
        subject = json.loads(raw) if raw else query_priority(provider, {"version": 1}, describe=top).get("subject")
        if not valid_subject(subject):
            raise ValueError("invalid priority subject")
        return subject
    except (OSError, ValueError, subprocess.SubprocessError, TypeError) as error:
        evidence.warn(str(error))
        return None


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


def process_alive(pid):
    try:
        os.kill(pid, 0)
    except ProcessLookupError:
        return False
    except PermissionError:  # alive, owned by someone else
        pass
    return True


def require_watch(watch):
    if not process_alive(watch):
        log(f"watched process {watch} exited; cancelling suite lock wait")
        raise SystemExit(1)


def wait_for_exit(pid):
    if hasattr(select, "kqueue"):
        kq = select.kqueue()
        ev = select.kevent(pid, select.KQ_FILTER_PROC, select.KQ_EV_ADD, select.KQ_NOTE_EXIT)
        try:
            kq.control([ev], 1, None)
            return
        except OSError:  # already gone
            return
    while process_alive(pid):
        time.sleep(0.5)


def acquire(watch, weight=1):
    if os.environ.get("WF_TEST_LOCK", "").lower() in ("off", "0", "no"):
        log("WF_TEST_LOCK=off: NOT taking the machine-wide full-suite lock; this run may collide with another")
        return "skipped"
    require_watch(watch)
    paths = lock_paths()
    os.makedirs(os.path.dirname(paths[0]), exist_ok=True)
    fds = [os.open(p, os.O_RDWR | os.O_CREAT, 0o644) for p in paths]
    queue = os.open(paths[0] + ".queue", os.O_RDWR | os.O_CREAT, 0o644)
    want = min(max(1, weight), len(fds))
    # sdlc local-check runs in a temp worktree and names its checkout here (#400).
    top = os.environ.get("WF_TEST_LOCK_OWNER") or subprocess.run(
        ["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True).stdout.strip() or os.getcwd()
    rec = {"repo": os.path.basename(top), "worktree": top, "pid": watch,
           "started": datetime.now().isoformat(timespec="seconds")}
    # Makefile.workflow exports the holder's make pid to its children, so a nested
    # `make test` inside a locked run (or make's own restart) passes through.
    held_by = int(os.environ.get("WF_TEST_LOCK_HELD_BY") or 0)
    limit = float(os.environ.get("WF_TEST_LOCK_TIMEOUT") or 0)

    def nested(busy):
        # A slot this make (or the make it runs under) already holds makes this a
        # nested run, even with another slot free.
        return any(r and r.get("pid") in (watch, held_by) for r in busy)

    # A nested run passes through without joining the line.
    if nested(scan(fds, 0)[1]):
        return "reentrant"
    provider = os.environ.get("WF_TEST_PRIORITY_PROVIDER") or os.path.join(
        os.path.dirname(os.path.realpath(__file__)), "..", "bin", "sdlc")
    evidence = GrantEvidence(provider)
    rec["subject"] = priority_subject(provider, top, evidence)
    line = Tickets(paths[0], rec)
    start = last = time.time()
    first, previous = True, None
    try:
        while True:
            require_watch(watch)
            taken, busy, ahead, head, deadline = line.admission(fds, queue, want, evidence)
            if nested(busy):
                for k in taken:
                    fcntl.flock(fds[k], fcntl.LOCK_UN)
                return "reentrant"
            if taken:
                break
            held = holders(busy)
            classification = (f"fleet-blocking issue #{rec['subject']['issue']}, until "
                              f"{datetime.fromtimestamp(deadline).isoformat(timespec='seconds')}"
                              if deadline > time.time() else "normal FIFO")
            held += f"\n  {classification}"
            if head:
                held += "\n  position 1 in line: next to run"
            else:
                position = len(ahead) + 1
                nxt = ahead[0] if ahead else holder(queue)
                held += f"\n  position {position} in line; next in line: {describe(nxt)}"
            now = time.time()
            wanted = "one" if want == 1 else f"{want}"
            position_state = (classification, len(ahead), head)
            if first or position_state != previous or now - last >= REPORT_EVERY:
                log(f"waiting {int(now - start)}s for {wanted} of {len(fds)} full-suite slots ({paths[0]}*); held by:{held}")
                first, last, previous = False, now, position_state
            if limit and now - start >= limit:
                log(f"gave up after {int(now - start)}s waiting for a full-suite slot — a lock wait, not a test failure; held by:{held}")
                return "timeout"
            time.sleep(1)
    finally:
        line.release()  # taking slots, giving up or nesting all leave the line

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
    # local-check cancellation uses SIGTERM. Unwind owned provider/ticket
    # resources before exiting, just as KeyboardInterrupt does for SIGINT.
    def terminate(signum, _frame):
        raise SystemExit(128 + signum)
    signal.signal(signal.SIGTERM, terminate)
    print(acquire(args.watch, args.slots))


if __name__ == "__main__":
    main()
