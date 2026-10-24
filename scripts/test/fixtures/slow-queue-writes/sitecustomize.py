"""Slows the queue helper's transaction steps, for #592's regressions.

On PYTHONPATH, it delays each in-place record write (os.pwrite, after the
truncate) and each removal of an operation record. An unlocked test read then
reliably sees the transaction half done; a read under control.lock never does.
"""
import os
import time

_pwrite = os.pwrite
_unlink = os.unlink


def pwrite(fd, data, offset):
    time.sleep(.3)
    return _pwrite(fd, data, offset)


def unlink(path, *args, **kwargs):
    if os.sep + 'operations' + os.sep in str(path) and str(path).endswith('.json'):
        time.sleep(.5)
    return _unlink(path, *args, **kwargs)


os.pwrite = pwrite
os.unlink = unlink
