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
import ctypes
import fcntl
import json
import os
import re
import select
import signal
import selectors
import stat
import tempfile
import subprocess
import sys
import time
import uuid
from datetime import datetime, timezone
from typing import TypedDict

REPORT_EVERY = 30

QUEUE_MAX_JSON = 1 << 20
QUEUE_MAX_OPERATIONS = 4096
EXIT_STATUS_NOTE = 0x04000000  # Darwin SDK NOTE_EXITSTATUS (not exposed by Python).
GROUP_MAX_PROCESSES = 65536
GROUP_MAX_SURVIVORS = 256
GROUP_SCAN_SECONDS = 5
WATCHER_READY_SECONDS = 5
COMPLETION_MAX_LINE = 1024


class DirectCompletionEvidence(TypedDict, total=False):
    owner_observation: str
    status_observation: str
    exit_status: int
    exit_observed_at: str
    process_group: int
    group_observation: str
    group_observed_at: str
    survivor_pids: list[int]
    survivors_truncated: bool


def decode_exit_evidence(fields):
    """Strict additive evidence; missing members remain unknown, never zero."""
    types = dict(owner_observation=str, status_observation=str, exit_status=int,
                 exit_observed_at=str, process_group=int, group_observation=str,
                 group_observed_at=str, survivor_pids=list, survivors_truncated=bool)
    present = {key: value for key, value in types.items() if key in fields}
    value = _queue_known(fields, present, set(types) - set(present))
    for key, choices in dict(owner_observation=('exact', 'unknown'),
                             status_observation=('observed', 'unavailable'),
                             group_observation=('empty', 'survivors', 'unknown')).items():
        if key in value and value[key] not in choices:
            raise ValueError('invalid completion evidence ' + key)
    for key in ('exit_observed_at', 'group_observed_at'):
        if key in value:
            stamp = value[key]
            if len(stamp) > 64 or not re.fullmatch(r'\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(?:\.\d+)?(?:Z|[+-]\d\d:\d\d)', stamp):
                raise ValueError('invalid completion timestamp')
            datetime.fromisoformat(stamp.replace('Z', '+00:00'))
    if 'exit_status' in value and not -255 <= value['exit_status'] <= 255:
        raise ValueError('invalid exit status')
    if 'process_group' in value and value['process_group'] <= 1:
        raise ValueError('invalid process group')
    if 'survivor_pids' in value:
        pids = value['survivor_pids']
        if len(pids) > GROUP_MAX_SURVIVORS or any(type(pid) is not int or pid <= 1 for pid in pids) or len(set(pids)) != len(pids):
            raise ValueError('invalid survivor PIDs')
    return value


def direct_retirement_decision(evidence: DirectCompletionEvidence, cancellation: bool, owner_pid: int) -> str:
    """Only independent positive observations authorize a finished lifetime."""
    return 'finished' if (not cancellation and evidence.get('owner_observation') == 'exact'
                         and evidence.get('status_observation') == 'observed'
                         and type(evidence.get('exit_status')) is int and evidence['exit_status'] == 0
                         and evidence.get('exit_observed_at') and evidence.get('group_observed_at')
                         and evidence.get('process_group') == owner_pid
                         and evidence.get('group_observation') == 'empty'
                         and evidence.get('survivor_pids') == []
                         and not evidence.get('survivors_truncated', False)) else 'unknown'


class OwnerExitObserver:
    """One exact incarnation, registered before the blocked make may continue."""
    def __init__(self, pid, started):
        self.pid, self.started, self.queue = pid, started, None
        self.evidence = dict(owner_observation='unknown', status_observation='unavailable')

    def register(self):
        if queue_process_start(self.pid) != self.started:
            raise ValueError('owner incarnation changed before exit observation')
        self.evidence['process_group'] = os.getpgid(self.pid)
        if hasattr(select, 'kqueue'):
            self.queue = select.kqueue()
            event = select.kevent(self.pid, select.KQ_FILTER_PROC, select.KQ_EV_ADD,
                                 select.KQ_NOTE_EXIT | EXIT_STATUS_NOTE)
            self.queue.control([event], 0, 0)
        if queue_process_start(self.pid) != self.started:
            raise ValueError('owner incarnation changed during exit observation')
        self.evidence['owner_observation'] = 'exact'

    def poll(self):
        if self.queue is not None:
            events = self.queue.control([], 1, 0)
            for event in events:
                if event.ident != self.pid or event.flags & select.KQ_EV_ERROR:
                    raise ValueError('invalid owner exit event')
                if event.fflags & select.KQ_NOTE_EXIT:
                    self.evidence['exit_observed_at'] = datetime.now(timezone.utc).isoformat()
                    if event.fflags & EXIT_STATUS_NOTE:
                        self.evidence.update(status_observation='observed', exit_status=os.waitstatus_to_exitcode(event.data))
                    return True
        if not process_alive(self.pid):
            self.evidence['exit_observed_at'] = datetime.now(timezone.utc).isoformat()
            return True
        return False

    def close(self):
        if self.queue is not None:
            self.queue.close()
            self.queue = None


def process_group_snapshot(pgid, owner):
    """One bounded native snapshot, only at exit; errors are never emptiness."""
    result = dict(group_observation='unknown', survivor_pids=[],
                  group_observed_at=datetime.now(timezone.utc).isoformat())
    if type(pgid) is not int or pgid <= 1 or sys.platform != 'darwin':
        return result
    start = time.monotonic()
    library = ctypes.CDLL('/usr/lib/libproc.dylib', use_errno=True)
    pids = (ctypes.c_int * (GROUP_MAX_PROCESSES + 1))()
    ctypes.set_errno(0)
    count = library.proc_listpids(2, pgid, ctypes.byref(pids), ctypes.sizeof(pids))  # PROC_PGRP_ONLY
    if count < 0 or (count == 0 and ctypes.get_errno()) or count % ctypes.sizeof(ctypes.c_int) or count >= ctypes.sizeof(pids) or time.monotonic() - start > GROUP_SCAN_SECONDS:
        return result
    # The exact owner's exit was observed; it may still be awaiting parent reap.
    survivors = sorted(set(pid for pid in pids[:count // ctypes.sizeof(ctypes.c_int)] if pid > 1 and pid != owner))
    result.update(group_observation='survivors' if survivors else 'empty',
                  survivor_pids=survivors[:GROUP_MAX_SURVIVORS])
    if len(survivors) > GROUP_MAX_SURVIVORS:
        result['survivors_truncated'] = True
    return result


def append_completion_audit(root, operation, evidence):
    """Diagnostic only; existing audit lifecycle, one bounded line per run."""
    line = json.dumps(dict(schema_version=1, id=operation['id'], pid=operation['pid'],
                           pid_started=operation['pid_started'], **evidence), separators=(',', ':')).encode() + b'\n'
    if len(line) > COMPLETION_MAX_LINE:
        raise ValueError('completion audit line exceeds its bound')
    path = os.path.join(root, 'completion.jsonl')
    fd = os.open(path, os.O_CREAT | os.O_RDWR | os.O_APPEND | os.O_NOFOLLOW, 0o600)
    try:
        if not stat.S_ISREG(os.fstat(fd).st_mode):
            raise ValueError('completion audit is not a regular file')
        size = os.fstat(fd).st_size
        if size and os.pread(fd, 1, size - 1) != b'\n':
            raise ValueError('torn completion audit; inspect before retrying')
        if os.write(fd, line) != len(line):
            raise OSError('short completion audit write')
        os.fsync(fd)
        directory = os.open(root, os.O_RDONLY)
        try:
            os.fsync(directory)
        finally:
            os.close(directory)
    finally:
        os.close(fd)


def queue_state_root():
    root = os.environ.get('WF_SDLC_QUEUE_ROOT')
    if not root and os.environ.get('WF_TEST_LOCK_FILE'):
        root = os.environ['WF_TEST_LOCK_FILE'] + '.queue-state'
    if not root:
        state = os.environ.get('XDG_STATE_HOME') or os.path.expanduser('~/.local/state')
        root = os.path.join(state, 'ariadne', 'queue')
    return os.path.abspath(root)


def _queue_pairs(pairs):
    out = {}
    for key, value in pairs:
        if key in out:
            raise ValueError('duplicate queue JSON member ' + repr(key))
        out[key] = value
    return out


def _queue_constant(value):
    raise ValueError('invalid queue JSON constant ' + value)


_QUEUE_DECODER = json.JSONDecoder(object_pairs_hook=_queue_pairs, parse_int=str,
                                 parse_float=str, parse_constant=_queue_constant)


def _queue_json(data):
    if not data or len(data) > QUEUE_MAX_JSON:
        raise ValueError('queue JSON empty or exceeds 1 MiB')
    text = data.decode('utf-8')
    try:
        parsed = _QUEUE_DECODER.decode(text)
        def depth(value, level=0):
            if level > 128:
                raise ValueError('queue JSON nesting exceeds 128')
            if isinstance(value, dict):
                for child in value.values():
                    depth(child, level + 1)
            elif isinstance(value, list):
                for child in value:
                    depth(child, level + 1)
        depth(parsed)
    except (RecursionError, UnicodeError) as error:
        raise ValueError('malformed queue JSON') from error
    return text, parsed


def _queue_skip_space(text, pos):
    while text[pos] in ' \t\r\n':
        pos += 1
    return pos


def queue_strict_object(data):
    """Strict structural boundary, returning original raw top-level values."""
    text, parsed = _queue_json(data)
    if not isinstance(parsed, dict):
        raise ValueError('queue JSON is not an object')
    fields = {}
    pos = text.index('{') + 1
    while True:
        pos = _queue_skip_space(text, pos)
        if text[pos] == '}':
            break
        key, pos = _QUEUE_DECODER.raw_decode(text, pos)
        pos = _queue_skip_space(text, pos)
        pos += 1  # colon, already structurally checked
        pos = _queue_skip_space(text, pos)
        start = pos
        _, pos = _QUEUE_DECODER.raw_decode(text, pos)
        fields[key] = text[start:pos].encode('utf-8')
        pos = _queue_skip_space(text, pos)
        if text[pos] == ',':
            pos += 1
    return fields


def queue_encode_object(fields, **changes):
    raw = dict(fields)
    for key, value in changes.items():
        raw[key] = json.dumps(value, ensure_ascii=False, allow_nan=False, separators=(',', ':')).encode('utf-8')
    result = b'{' + b','.join(json.dumps(key).encode() + b':' + value for key, value in sorted(raw.items())) + b'}'
    if len(result) > QUEUE_MAX_JSON:
        raise ValueError('queue JSON exceeds 1 MiB')
    return result


def _queue_known(fields, required, optional=()):
    names = set(required) | set(optional)
    for name in fields:
        if name not in names and any(name.casefold() == known.casefold() for known in names):
            raise ValueError('noncanonical queue member ' + name)
    result = {}
    for key, expected in required.items():
        if key not in fields:
            raise ValueError('missing queue member ' + key)
        value = json.loads(fields[key])
        if type(value) is not expected:
            raise ValueError('wrong queue member type ' + key)
        result[key] = value
    return result


def _queue_text(value, limit):
    return isinstance(value, str) and len(value.encode('utf-8')) <= limit


def _queue_read(path):
    with open(path, 'rb') as stream:
        raw = stream.read(QUEUE_MAX_JSON + 1)
    return queue_strict_object(raw)


def _queue_control(root):
    try:
        fields = _queue_read(os.path.join(root, 'pause.json'))
    except FileNotFoundError:
        return []
    control = _queue_known(fields, dict(schema_version=int, pause_owners=list))
    if control['schema_version'] != 1 or len(control['pause_owners']) > 128:
        raise ValueError('unsupported or oversized queue owner set')
    owners, seen = [], set()
    for item in control['pause_owners']:
        if not isinstance(item, dict):
            raise ValueError('invalid queue owner')
        # Owners are read only; unknown members stay opaque in the control file.
        for name, expected in dict(owner=str, reason=str, cancel_running=bool).items():
            if name not in item or type(item[name]) is not expected:
                raise ValueError('invalid queue owner member ' + name)
        if not item['owner'].strip() or not _queue_text(item['owner'], 256) or not _queue_text(item['reason'], 4096) or item['owner'] in seen:
            raise ValueError('duplicate or oversized queue owner')
        seen.add(item['owner'])
        owners.append(item)
    return owners


def _queue_operation(fields):
    op = _queue_known(fields, dict(schema_version=int, id=str, workspace=str, kinds=list, phase=str,
                     pid=int, pid_started=str, managed=bool, started_at=str, boot_id=str, root=str, detail=str), ('head',))
    if op['schema_version'] != 1 or not queue_valid_id(op['id']):
        raise ValueError('unsupported queue operation identity')
    if op['phase'] not in ('waiting', 'running', 'interrupting', 'unknown', 'finished', 'interrupted'):
        raise ValueError('invalid queue operation phase')
    if not op['kinds'] or len(op['kinds']) > 3 or any(type(kind) is not str or kind not in ('suite', 'merge', 'landing') for kind in op['kinds']) or len(set(op['kinds'])) != len(op['kinds']):
        raise ValueError('invalid queue operation kinds')
    for name, limit in dict(workspace=4096, pid_started=256, boot_id=256, root=4096, detail=4096).items():
        if not _queue_text(op[name], limit):
            raise ValueError('invalid queue operation text ' + name)
    if 'head' in fields:
        head = json.loads(fields['head'])
        if not _queue_text(head, 256):
            raise ValueError('invalid queue operation head')
    if op['pid'] < 0 or (op['managed'] and (op['pid'] <= 1 or not op['pid_started'])):
        raise ValueError('queue operation lacks process identity')
    if not re.fullmatch(r'\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(?:\.\d+)?(?:Z|[+-]\d\d:\d\d)', op['started_at']):
        raise ValueError('invalid operation timestamp')
    started = datetime.fromisoformat(op['started_at'].replace('Z', '+00:00'))
    if started == datetime.min.replace(tzinfo=timezone.utc):
        raise ValueError('zero operation timestamp')
    decode_exit_evidence(fields)
    return op


def queue_valid_id(value):
    return value not in ('.', '..') and re.fullmatch(r'[A-Za-z0-9_.:-]{1,256}', value) is not None


def queue_process_start(pid):
    """Kernel incarnation, formatted exactly as Go processgroup.StartTime."""
    if sys.platform == 'darwin':
        # Darwin SDK sys/proc_info.h struct proc_bsdinfo; libproc supplies
        # semantic start timeval fields, without ps formatting or padding.
        class BSDInfo(ctypes.Structure):
            _fields_ = [('prefix', ctypes.c_uint32 * 12), ('comm', ctypes.c_char * 16),
                        ('name', ctypes.c_char * 32), ('middle', ctypes.c_uint32 * 6),
                        ('sec', ctypes.c_uint64), ('usec', ctypes.c_uint64)]
        info = BSDInfo()
        library = ctypes.CDLL('/usr/lib/libproc.dylib', use_errno=True)
        count = library.proc_pidinfo(pid, 3, 0, ctypes.byref(info), ctypes.sizeof(info))
        if count != ctypes.sizeof(info) or info.prefix[3] != pid:
            raise OSError(ctypes.get_errno(), 'kernel process identity unavailable')
        return f'{info.sec}.{info.usec:06d}'
    if sys.platform.startswith('linux'):
        with open(f'/proc/{pid}/stat') as stream:
            raw = stream.read(8192)
        fields = raw[raw.rfind(')') + 1:].split()
        return fields[19]
    raise OSError('kernel process identity unsupported')


def queue_boot_id():
    if sys.platform == 'darwin':
        class Timeval(ctypes.Structure):
            _fields_ = [('sec', ctypes.c_long), ('usec', ctypes.c_int32)]
        value = Timeval()
        size = ctypes.c_size_t(ctypes.sizeof(value))
        library = ctypes.CDLL(None, use_errno=True)
        if library.sysctlbyname(b'kern.boottime', ctypes.byref(value), ctypes.byref(size), None, 0) != 0:
            raise OSError(ctypes.get_errno(), 'boot identity unavailable')
        return f'{value.sec}.{value.usec:06d}'
    if sys.platform.startswith('linux'):
        with open('/proc/sys/kernel/random/boot_id') as stream:
            return stream.read(256).strip()
    raise OSError('boot identity unsupported')


def _queue_atomic(path, raw):
    os.makedirs(os.path.dirname(path), mode=0o700, exist_ok=True)
    fd, temporary = tempfile.mkstemp(prefix='.queue-', dir=os.path.dirname(path))
    try:
        with os.fdopen(fd, 'wb') as stream:
            stream.write(raw)
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(temporary, path)
    finally:
        try:
            os.unlink(temporary)
        except FileNotFoundError:
            pass


class QueueGate:
    """Portable admission protocol. It never inspects fleet/process inventories."""
    def __init__(self, root, top, watch):
        self.root, self.top, self.watch = root, os.path.realpath(top), watch
        self.last_pause = ''
        os.makedirs(root, mode=0o700, exist_ok=True)
        self.mutex = os.open(os.path.join(root, 'control.lock'), os.O_RDWR | os.O_CREAT, 0o600)
        self.fd = None
        self.managed = bool(os.environ.get('WF_SDLC_QUEUE_ID'))
        self.id = os.environ.get('WF_SDLC_QUEUE_ID') or uuid.uuid4().hex
        if not queue_valid_id(self.id):
            raise ValueError('invalid queue operation ID')
        self.path = os.path.join(root, 'operations', self.id + '.json')
        if self.managed:
            with self.locked():
                self.fd = os.open(self.path, os.O_RDWR)
                fields, op = self.operation()
                if not op['managed'] or op['root'] != self.top or op['phase'] != 'waiting' or 'suite' not in op['kinds']:
                    raise ValueError('managed queue operation binding mismatch')
            if queue_process_start(op['pid']) != op['pid_started']:
                raise ValueError('managed queue controller incarnation mismatch')
        else:
            started = queue_process_start(watch)
            self.owner_started = started
            try:
                boot = queue_boot_id()
            except OSError:
                boot = ''  # unavailable boot evidence can never retire uncertainty
            # A filename layout is not Git-verified workspace identity. The
            # managed controller supplies that proof; raw make stays unknown.
            raw = dict(schema_version=1, id=self.id, workspace='', kinds=['suite'], phase='waiting',
                       pid=watch, pid_started=started, managed=False, started_at=datetime.now(timezone.utc).isoformat(),
                       boot_id=boot, root=self.top, detail='direct full suite in ' + self.top + '; workspace identity unavailable')
            self.fields = queue_strict_object(json.dumps(raw).encode())
            with self.locked():
                for family in ('operations', 'unresolved', 'cancel'):
                    directory = os.path.join(root, family)
                    os.makedirs(directory, mode=0o700, exist_ok=True)
                    if len(os.listdir(directory)) >= QUEUE_MAX_OPERATIONS:
                        raise ValueError('queue registry capacity reached')
                fd, temporary = tempfile.mkstemp(prefix='.enroll-', dir=os.path.join(root, 'operations'))
                self.fd = fd
                try:
                    fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
                    os.write(fd, queue_encode_object(self.fields))
                    os.fsync(fd)
                    if any(os.path.lexists(os.path.join(root, family, self.id + '.json')) for family in ('operations', 'unresolved', 'cancel', 'interrupted')):
                        raise ValueError('queue operation identity already exists')
                    os.rename(temporary, self.path)
                except BaseException:
                    os.close(fd)
                    self.fd = None
                    os.unlink(temporary)
                    raise

    @contextlib.contextmanager
    def locked(self):
        end = time.monotonic() + 4
        while not try_lock(self.mutex):
            if time.monotonic() >= end:
                raise ValueError('queue control mutex unavailable')
            time.sleep(.01)
        try:
            yield
        finally:
            fcntl.flock(self.mutex, fcntl.LOCK_UN)

    def operation(self):
        if self.fd is None:
            raise ValueError('queue lifetime lease is closed')
        if self.managed and try_lock(self.fd):
            fcntl.flock(self.fd, fcntl.LOCK_UN)
            raise ValueError('managed queue lifetime lease is not held')
        fields = queue_strict_object(os.pread(self.fd, QUEUE_MAX_JSON + 1, 0))
        op = _queue_operation(fields)
        if op['id'] != self.id or op['root'] != self.top or op['managed'] != self.managed:
            raise ValueError('queue lifetime identity changed')
        if not self.managed and (op['pid'] != self.watch or op['pid_started'] != self.owner_started):
            raise ValueError('watched owner incarnation changed')
        return fields, op

    def pause_reason(self):
        owners = _queue_control(self.root)
        return 'queue paused: ' + '; '.join(item['owner'] + ': ' + json.dumps(item['reason']) for item in owners) if owners else ''

    def paused(self):
        with self.locked():
            self.last_pause = self.pause_reason()
        return self.last_pause

    def rewrite(self, fields, **changes):
        raw = queue_encode_object(fields, **changes)
        _queue_operation(queue_strict_object(raw))
        os.ftruncate(self.fd, 0)
        os.pwrite(self.fd, raw, 0)
        os.fsync(self.fd)

    def try_admit(self, resource):
        with self.locked():
            self.last_pause = self.pause_reason()
            if self.last_pause:
                return False
            fields, op = self.operation()
            if op['phase'] != 'waiting':
                raise ValueError('only a waiting queue operation admits')
            if not resource():
                return False
            self.rewrite(fields, phase='running')
            return True

    def binding(self):
        _, op = self.operation()
        return dict(queue_operation_id=op['id'], queue_pid=op['pid'], queue_pid_started=op['pid_started'], queue_root=op['root'])

    def intent(self):
        try:
            fields = _queue_read(os.path.join(self.root, 'cancel', self.id + '.json'))
        except FileNotFoundError:
            return None
        value = _queue_known(fields, dict(schema_version=int, id=str, reason=str))
        if value['schema_version'] != 1 or value['id'] != self.id or not _queue_text(value['reason'], 4096):
            raise ValueError('queue cancellation intent identity mismatch')
        return value['reason']

    def retire(self, evidence=None):
        with self.locked():
            fields, op = self.operation()
            if self.managed:
                return  # Go owns both its lifetime and interruption evidence.
            if op['phase'] != 'waiting':
                evidence = evidence or {}
                decode_exit_evidence(queue_strict_object(queue_encode_object({}, **evidence)))
                try:
                    reason = self.intent()
                    cancellation = reason is not None or op['phase'] != 'running'
                except ValueError:
                    reason, cancellation = None, True
                if op['pid'] != self.watch:
                    raise ValueError('watched owner binding changed')
                if direct_retirement_decision(evidence, cancellation, self.watch) == 'finished':
                    try:
                        append_completion_audit(self.root, op, evidence)
                    except (OSError, ValueError):
                        evidence = dict(evidence, group_observation='unknown')
                    else:
                        self.rewrite(fields, phase='finished', **evidence)
                        os.unlink(self.path)
                        os.close(self.fd)
                        self.fd = None
                        return
                detail = 'unmanaged operation ' + self.id + ' in ' + self.top + ': watched owner ended; descendants require operator reconciliation'
                if evidence:
                    detail += '; exit observed ' + evidence.get('exit_observed_at', 'unavailable')
                    detail += '; status ' + str(evidence.get('exit_status', 'unavailable'))
                    detail += '; group ' + evidence.get('group_observation', 'unknown')
                    survivors = evidence.get('survivor_pids', [])
                    if survivors:
                        detail += '; survivors ' + ','.join(map(str, survivors))
                detail = detail.encode('utf-8')[:4096].decode('utf-8', errors='ignore')
                unknown = queue_encode_object(fields, phase='unknown', detail=detail, **evidence)
                destination = os.path.join(self.root, 'unresolved', self.id + '.json')
                if not os.path.exists(destination):
                    if len(os.listdir(os.path.join(self.root, 'unresolved'))) >= QUEUE_MAX_OPERATIONS:
                        raise ValueError('unresolved queue registry full; lifetime retained')
                    _queue_atomic(destination, unknown)
                try:
                    reason = self.intent()
                except ValueError:
                    reason = None  # malformed intent remains for operator repair
                if reason is not None:
                    envelope = queue_encode_object({}, schema_version=1, id=self.id, outcome='interrupted', reason=reason or 'queue cancellation', recorded_at=datetime.now(timezone.utc).isoformat())
                    raw = queue_strict_object(envelope)
                    raw['operation'] = unknown
                    path = os.path.join(self.root, 'interrupted', self.id + '.json')
                    if not os.path.exists(path):
                        os.makedirs(os.path.dirname(path), mode=0o700, exist_ok=True)
                        if len(os.listdir(os.path.dirname(path))) >= QUEUE_MAX_OPERATIONS:
                            raise ValueError('interrupted queue registry full; repair required')
                        _queue_atomic(path, queue_encode_object(raw))
                    # The unresolved operation still references this exact intent.
                    # Resume does not remove it or assert descendant completion.
            os.unlink(self.path)
            os.close(self.fd)
            self.fd = None

    def close(self):
        if self.fd is not None:
            os.close(self.fd)
            self.fd = None
        os.close(self.mutex)

    def after_fork(self):
        # A flock belongs to the open file description. The watcher must not
        # share its mutex description with its still-publishing parent.
        os.close(self.mutex)
        self.mutex = os.open(os.path.join(self.root, 'control.lock'), os.O_RDWR)


def watch_queue_owner(gate, watch, observer=None):
    """Only unmanaged owners consume intents; managed Go reaps its own group."""
    sent = False
    try:
        while process_alive(watch):
            if observer is not None and observer.poll():
                break
            if not gate.managed and not sent:
                try:
                    with gate.locked():
                        fields, op = gate.operation()
                        reason = gate.intent()
                    if reason is not None:
                        if not op['pid_started'] or queue_process_start(watch) != op['pid_started'] or op['pid'] != watch:
                            raise ValueError('queue cancellation kernel incarnation mismatch')
                        with gate.locked():
                            fields, current = gate.operation()
                            if current['pid_started'] != op['pid_started'] or gate.intent() is None:
                                raise ValueError('queue cancellation identity changed')
                            gate.rewrite(fields, phase='interrupting')
                        os.kill(watch, signal.SIGTERM)  # one fenced make, never its shared group
                        sent = True
                except (OSError, ValueError):
                    # No cancellation authority: keep the lifetime visible.
                    pass
            time.sleep(.1)
        if not gate.managed:
            evidence = {}
            if observer is not None:
                observer.poll()  # collect a queued event even if the PID was reaped first
                evidence = dict(observer.evidence)
                try:
                    # Reused PID/group names never supply completion authority.
                    if queue_process_start(watch) != observer.started:
                        evidence['owner_observation'] = 'unknown'
                except ProcessLookupError:
                    pass
                except OSError:
                    if process_alive(watch):
                        evidence['owner_observation'] = 'unknown'
                try:
                    evidence.update(process_group_snapshot(evidence.get('process_group'), watch))
                except (OSError, ValueError):
                    evidence.update(group_observation='unknown', survivor_pids=[])
            gate.retire(evidence=evidence)
    finally:
        if observer is not None:
            observer.close()
        gate.close()


def start_owner_watcher(gate, watch):
    """Do not let make run before the post-fork exit observer is registered."""
    read, write = os.pipe()
    child = os.fork()
    if child == 0:
        os.close(read)
        observer = None
        try:
            os.setsid()  # the monitor must not be counted in make's job group
            for sig in (signal.SIGINT, signal.SIGHUP):
                signal.signal(sig, signal.SIG_IGN)
            null = os.open(os.devnull, os.O_RDWR)
            for std in (0, 1, 2):
                os.dup2(null, std)
            os.close(null)
            gate.after_fork()
            if not gate.managed:
                _, op = gate.operation()
                observer = OwnerExitObserver(watch, op['pid_started'])
                observer.register()
            os.write(write, b'1')
            os.close(write)
            write = None
            watch_queue_owner(gate, watch, observer)
        except BaseException:
            os._exit(1)  # incomplete lifetime is recovered conservatively by Go
        os._exit(0)
    os.close(write)
    try:
        ready, _, _ = select.select([read], [], [], WATCHER_READY_SECONDS)
        if not ready or os.read(read, 1) != b'1':
            # This exact child was just forked and remains ours until waitpid.
            try:
                os.kill(child, signal.SIGKILL)
            except ProcessLookupError:
                pass
            os.waitpid(child, 0)
            raise ValueError('suite watcher did not acknowledge exit observation')
        return child
    finally:
        os.close(read)


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
    def locked(self, blocking=True):
        if blocking:
            fcntl.flock(self.mutex, fcntl.LOCK_EX)
        elif not try_lock(self.mutex):
            yield False
            return
        try:
            yield True
        finally:
            fcntl.flock(self.mutex, fcntl.LOCK_UN)

    def records(self):
        records = []
        for number, path in live_tickets(self.dir):
            try:
                with open(path, "rb") as stream:
                    records.append((number, holder(stream.fileno())))
            except FileNotFoundError:
                continue
        return records

    def elect(self, records, deadlines, fds, queue, want):
        order = effective_order(records, deadlines, time.time())
        index = next(i for i, (n, _) in enumerate(order) if n == self.number)
        ahead = [r for _, r in order[:index]]
        head = index == 0 and try_lock(queue)
        if head:
            try:
                os.ftruncate(queue, 0)
                os.pwrite(queue, json.dumps(order[0][1]).encode(), 0)
                taken, busy = scan(fds, want)
                if taken:
                    rec = dict(order[0][1], slot=taken[0], slots=taken)
                    raw = json.dumps(rec).encode()
                    try:
                        for slot in taken:
                            os.ftruncate(fds[slot], 0)
                            os.pwrite(fds[slot], raw, 0)
                    except BaseException:
                        for slot in taken:
                            fcntl.flock(fds[slot], fcntl.LOCK_UN)
                        raise
            finally:
                fcntl.flock(queue, fcntl.LOCK_UN)
        else:
            taken, busy = scan(fds, 0)
        deadline = deadlines.get(subject_key(order[index][1]), 0)
        return taken, busy, ahead, head, deadline

    def admission(self, fds, queue, want, evidence, gate=None):
        """Prepare slow priority evidence outside global gate; revalidate inside."""
        if gate is None:  # old direct unit callers share the same scheduler
            with self.locked():
                records = self.records()
                deadlines = evidence.deadlines(records, time.time())
                return self.elect(records, deadlines, fds, queue, want)
        if gate.paused():
            return [], scan(fds, 0)[1], None, False, 0
        with self.locked():
            records = self.records()
        # Never call a provider while holding the machine gate. New/vanished
        # tickets invalidate this snapshot and retry without taking any slot.
        deadlines = evidence.deadlines(records, time.time())
        # None means no validated order, unlike an observed empty ahead list.
        result = ([], scan(fds, 0)[1], None, False, 0)
        def attempt():
            nonlocal result
            with self.locked(blocking=False) as acquired:
                if not acquired:
                    return False  # Release the global gate before retrying.
                current = self.records()
                if current != records:
                    return False
                result = self.elect(current, deadlines, fds, queue, want)
                return bool(result[0])
        gate.try_admit(attempt)
        return result

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
    top = os.environ.get("WF_TEST_LOCK_OWNER") or subprocess.run(
        ["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True).stdout.strip() or os.getcwd()
    rec = {"repo": os.path.basename(top), "worktree": top, "pid": watch,
           "started": datetime.now().isoformat(timespec="seconds")}
    held_by = int(os.environ.get("WF_TEST_LOCK_HELD_BY") or 0)
    limit = float(os.environ.get("WF_TEST_LOCK_TIMEOUT") or 0)

    def nested(busy):
        return any(r and r.get("pid") in (watch, held_by) for r in busy)

    gate, detached = None, False
    try:
        if nested(scan(fds, 0)[1]):
            return "reentrant"
        gate = QueueGate(queue_state_root(), top, watch)
        with gate.locked():
            rec.update(gate.binding())
        if os.environ.get('WF_SDLC_QUEUE_STOP_IF_PAUSED') == '1' and gate.paused():
            log(gate.last_pause + '; admitted merge stops before a new suite')
            return 'paused'
        provider = os.environ.get("WF_TEST_PRIORITY_PROVIDER") or os.path.join(
            os.path.dirname(os.path.realpath(__file__)), "..", "bin", "sdlc")
        evidence = GrantEvidence(provider)
        rec["subject"] = priority_subject(provider, top, evidence)
        line = Tickets(paths[0], rec)
        start = last = time.time()
        first, previous = True, None
        last_order = None
        try:
            while True:
                require_watch(watch)
                taken, busy, ahead, head, deadline = line.admission(fds, queue, want, evidence, gate)
                if nested(busy):
                    for slot in taken:
                        fcntl.flock(fds[slot], fcntl.LOCK_UN)
                    return "reentrant"
                if taken:
                    break
                if gate.last_pause and os.environ.get('WF_SDLC_QUEUE_STOP_IF_PAUSED') == '1':
                    log(gate.last_pause + '; admitted merge stops before a new suite')
                    return 'paused'
                held = holders(busy)
                if ahead is not None:
                    classification = (f"fleet-blocking issue #{rec['subject']['issue']}, until "
                                      f"{datetime.fromtimestamp(deadline).isoformat(timespec='seconds')}"
                                      if deadline > time.time() else "normal FIFO")
                    if head:
                        position = "position 1 in line: next to run"
                    else:
                        nxt = ahead[0] if ahead else holder(queue)
                        position = f"position {len(ahead) + 1} in line; next in line: {describe(nxt)}"
                    last_order = classification, position, (classification, len(ahead), head)
                if last_order is None:
                    classification, position, order_state = "queue order unavailable", "position unknown", None
                else:
                    classification, position, order_state = last_order
                    if ahead is None:
                        classification = "last observed " + classification
                        position = "last observed " + position
                held += f"\n  {classification}"
                if gate.last_pause:
                    held += "\n  " + gate.last_pause
                held += "\n  " + position
                now = time.time()
                wanted = "one" if want == 1 else str(want)
                position_state = (order_state, gate.last_pause)
                if first or position_state != previous or now - last >= REPORT_EVERY:
                    log(f"waiting {int(now - start)}s for {wanted} of {len(fds)} full-suite slots ({paths[0]}*); held by:{held}")
                    first, last, previous = False, now, position_state
                if limit and now - start >= limit:
                    log(f"gave up after {int(now - start)}s waiting for a full-suite slot — a lock wait, not a test failure; held by:{held}")
                    return "timeout"
                time.sleep(1)
        finally:
            line.release()
        os.close(queue)
        queue = None
        for slot, other in enumerate(fds):
            if slot not in taken:
                os.close(other)
        if not first:
            log(f"lock acquired after {int(time.time() - start)}s")
        child = start_owner_watcher(gate, watch)
        detached = True
        rec.update(holder=child, slot=taken[0], slots=taken, started=datetime.now().isoformat(timespec="seconds"))
        with gate.locked():
            for slot in taken:
                os.ftruncate(fds[slot], 0)
                os.pwrite(fds[slot], json.dumps(rec).encode(), 0)
        return "acquired"
    finally:
        if gate is not None:
            if not detached and not gate.managed:
                gate.retire()
            gate.close()
        if queue is not None:
            os.close(queue)
        for fd in fds:
            try:
                os.close(fd)
            except OSError:
                pass


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
