"""Fixture reads of the queue helper's state, shared by its private tests (#592).

The helper changes queue state in transactions under control.lock: retire
writes unresolved/, then interrupted/, then unlinks the operation record, and
rewrite truncates and rewrites that record in place (its descriptor is the
lifetime lease). An unlocked read can see a transaction half done.
"""
import fcntl
import os
from pathlib import Path

SLOW_WRITES = Path(__file__).resolve().parent / 'slow-queue-writes'


def committed(queue, predicate):
    """predicate, evaluated only while holding queue's control.lock. A busy or
    absent lock reads as not yet."""
    def check():
        try:
            fd = os.open(Path(queue) / 'control.lock', os.O_RDWR)
        except FileNotFoundError:
            return False
        try:
            try:
                fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except BlockingIOError:
                return False
            return predicate()
        finally:
            os.close(fd)
    return check


def committed_read(await_, queue, path):
    """path's bytes, read under queue's control.lock; await_ is the test's
    polling wait."""
    raw = []
    await_(committed(queue, lambda: raw.append(Path(path).read_bytes()) or True), 'control.lock stayed busy')
    return raw[0]


def slow_writes(env):
    """env with the slow-queue-writes shim first on PYTHONPATH."""
    return dict(env, PYTHONPATH=os.pathsep.join(filter(None, (str(SLOW_WRITES), env.get('PYTHONPATH')))))
