#!/usr/bin/env python3
"""Pure completion evidence and adversarial native-observation contracts."""
import importlib.util
import itertools
import json
import ctypes
import os
import signal
from pathlib import Path
import tempfile
from types import SimpleNamespace
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location('decision_lock', Path(__file__).resolve().parents[1] / 'test-lock.py')
lock = importlib.util.module_from_spec(spec)
spec.loader.exec_module(lock)


class Decision(unittest.TestCase):
    def evidence(self, **changes):
        value = dict(owner_observation='exact', status_observation='observed', exit_status=0,
                     exit_observed_at='2026-10-22T05:00:00+00:00', process_group=123,
                     group_observation='empty', group_observed_at='2026-10-22T05:00:01+00:00',
                     survivor_pids=[])
        value.update(changes)
        return value

    def test_completion_requires_all_independent_positive_proofs(self):
        self.assertTrue(hasattr(lock, 'direct_retirement_decision'), 'missing completion decision')
        for exact, observed, zero, private, empty, cancel in itertools.product((False, True), repeat=6):
            evidence = self.evidence(owner_observation='exact' if exact else 'unknown',
                                     status_observation='observed' if observed else 'unavailable',
                                     exit_status=0 if zero else 7,
                                     process_group=123 if private else 9,
                                     group_observation='empty' if empty else 'survivors',
                                     survivor_pids=[] if empty else [987])
            self.assertEqual(lock.direct_retirement_decision(evidence, cancel, 123),
                             'finished' if all((exact, observed, zero, private, empty, not cancel)) else 'unknown')

    def test_missing_or_malformed_evidence_never_manufactures_success(self):
        self.assertTrue(hasattr(lock, 'decode_exit_evidence'), 'missing evidence boundary')
        self.assertEqual(lock.direct_retirement_decision({}, False, 123), 'unknown')
        for name, bad in [('exit_status', True), ('process_group', '123'), ('survivor_pids', [True]),
                          ('exit_observed_at', 'yesterday'), ('group_observation', 'gone')]:
            with self.subTest(name=name), self.assertRaises(ValueError):
                raw = {k: json.dumps(v).encode() for k, v in self.evidence(**{name: bad}).items()}
                lock.decode_exit_evidence(raw)
        raw = {k: json.dumps(v).encode() for k, v in self.evidence().items()}
        self.assertEqual(lock.decode_exit_evidence(raw), self.evidence())

    def test_noncanonical_evidence_cannot_hide_from_strict_boundary(self):
        with self.assertRaises(ValueError):
            lock.decode_exit_evidence({'Exit_Status': b'0'})


class ProcessRegistry:
    """Stateful native snapshot double: contents change between observations."""
    def __init__(self, members):
        self.members, self.calls, self.error = members, 0, 0

    def proc_listpids(self, kind, group, target, capacity):
        self.calls += 1
        ctypes.set_errno(self.error)
        if self.error:
            return -1
        array = ctypes.cast(target, ctypes.POINTER(ctypes.c_int))
        for index, pid in enumerate(self.members[:capacity // ctypes.sizeof(ctypes.c_int)]):
            array[index] = pid
        return len(self.members) * ctypes.sizeof(ctypes.c_int)


class Observation(unittest.TestCase):
    def test_retirement_crash_cuts_leave_only_committed_terminal_authority(self):
        for after_commit in (False, True):
            with self.subTest(after_commit=after_commit), tempfile.TemporaryDirectory(prefix='direct-crash-', dir=os.environ.get('TMPDIR') or '/tmp') as root:
                with mock.patch.dict(os.environ, {'WF_SDLC_QUEUE_ROOT': root, 'WF_SDLC_QUEUE_ID': ''}):
                    gate = lock.QueueGate(root, root, os.getpid())
                    try:
                        self.assertTrue(gate.try_admit(lambda: True))
                        rewrite = gate.rewrite
                        def crash(fields, **changes):
                            if after_commit:
                                rewrite(fields, **changes)
                            raise OSError('injected process death at terminal publication')
                        with mock.patch.object(gate, 'rewrite', side_effect=crash), self.assertRaises(OSError):
                            gate.retire(evidence=Decision().evidence(process_group=os.getpid()))
                        raw = json.loads(Path(gate.path).read_bytes())
                        self.assertEqual(raw['phase'], 'finished' if after_commit else 'running')
                        audit = json.loads((Path(root) / 'completion.jsonl').read_bytes())
                        self.assertEqual(audit['id'], raw['id'])
                    finally:
                        gate.close()

    def test_watcher_registration_failure_and_timeout_never_acknowledge_acquired(self):
        class Gate:
            def __init__(self, stall):
                self.stall = stall

            def after_fork(self):
                if self.stall:
                    signal.pause()
                raise ValueError('injected registration failure')

        for stall in (False, True):
            children = []
            fork = os.fork
            def capture_child():
                pid = fork()
                if pid:
                    children.append(pid)
                return pid
            with self.subTest(stall=stall), mock.patch.object(lock.os, 'fork', side_effect=capture_child), \
                 mock.patch.object(lock, 'WATCHER_READY_SECONDS', .1):
                with self.assertRaisesRegex(ValueError, 'did not acknowledge'):
                    lock.start_owner_watcher(Gate(stall), os.getpid())
            self.assertEqual(len(children), 1)
            with self.assertRaises(ChildProcessError):
                os.waitpid(children[0], os.WNOHANG)

    @unittest.skipUnless(hasattr(lock.select, 'kqueue'), 'Darwin observer conformance')
    def test_observer_requires_status_flag_and_exact_incarnation(self):
        class Registry:
            def __init__(self):
                self.started = 'original'
                self.events = []
                self.registered = False
                self.closed = False

            def control(self, changes, maximum, timeout):
                if changes:
                    self.registered = True
                    return []
                events, self.events = self.events, []
                return events

            def close(self):
                self.closed = True

        registry = Registry()
        with mock.patch.object(lock.select, 'kqueue', return_value=registry), \
             mock.patch.object(lock, 'queue_process_start', side_effect=lambda _: registry.started), \
             mock.patch.object(lock.os, 'getpgid', return_value=123), \
             mock.patch.object(lock, 'process_alive', return_value=True):
            observer = lock.OwnerExitObserver(123, 'original')
            observer.register()
            self.assertTrue(registry.registered)
            self.assertFalse(observer.poll())
            registry.events = [SimpleNamespace(ident=123, flags=0, fflags=lock.select.KQ_NOTE_EXIT, data=0)]
            self.assertTrue(observer.poll())
            self.assertNotIn('exit_status', observer.evidence)
            registry.events = [SimpleNamespace(ident=123, flags=0, fflags=lock.select.KQ_NOTE_EXIT | lock.EXIT_STATUS_NOTE, data=7 << 8)]
            self.assertTrue(observer.poll())
            self.assertEqual(observer.evidence['exit_status'], 7)
            observer.close()
            self.assertTrue(registry.closed)
            registry.started = 'recycled'
            with self.assertRaises(ValueError):
                lock.OwnerExitObserver(123, 'original').register()

    def test_admitted_owner_binding_cannot_be_rewritten_into_completion(self):
        with tempfile.TemporaryDirectory(prefix='direct-binding-', dir=os.environ.get('TMPDIR') or '/tmp') as root:
            with mock.patch.dict(os.environ, {'WF_SDLC_QUEUE_ROOT': root, 'WF_SDLC_QUEUE_ID': ''}):
                gate = lock.QueueGate(root, root, os.getpid())
                try:
                    fields, _ = gate.operation()
                    raw = lock.queue_encode_object(fields, pid_started='different-incarnation')
                    os.ftruncate(gate.fd, 0)
                    os.pwrite(gate.fd, raw, 0)
                    with self.assertRaises(ValueError):
                        gate.operation()
                finally:
                    gate.close()

    def test_failed_audit_and_prior_cancellation_retain_evidence(self):
        for mode in ('audit-failure', 'cancel', 'bad-cancel'):
            with self.subTest(mode=mode), tempfile.TemporaryDirectory(prefix='direct-retire-', dir=os.environ.get('TMPDIR') or '/tmp') as root:
                with mock.patch.dict(os.environ, {'WF_SDLC_QUEUE_ROOT': root, 'WF_SDLC_QUEUE_ID': ''}):
                    gate = lock.QueueGate(root, root, os.getpid())
                    try:
                        self.assertTrue(gate.try_admit(lambda: True))
                        evidence = Decision().evidence(process_group=os.getpid())
                        if mode != 'audit-failure':
                            path = Path(root) / 'cancel' / (gate.id + '.json')
                            path.write_text('{"torn":' if mode == 'bad-cancel' else json.dumps(dict(schema_version=1,id=gate.id,reason='private')))
                        with mock.patch.object(lock, 'append_completion_audit', side_effect=OSError('disk failure')):
                            gate.retire(evidence=evidence)
                        raw = json.loads((Path(root) / 'unresolved' / (gate.id + '.json')).read_bytes())
                        self.assertEqual(raw['phase'], 'unknown')
                        self.assertEqual(raw['exit_status'], 0)
                        self.assertEqual(raw['exit_observed_at'], evidence['exit_observed_at'])
                        self.assertFalse((Path(root) / 'completion.jsonl').exists())
                    finally:
                        gate.close()

    def test_one_group_snapshot_has_bounded_linear_work_and_unknown_on_overflow(self):
        for size in (64, 128):
            registry = ProcessRegistry(list(range(1000, 1000 + size)))
            with mock.patch.object(lock.sys, 'platform', 'darwin'), mock.patch.object(lock.ctypes, 'CDLL', return_value=registry):
                result = lock.process_group_snapshot(123, 123)
                self.assertEqual(result['survivor_pids'], registry.members)
                self.assertEqual(registry.calls, 1)
                registry.members = []
                self.assertEqual(lock.process_group_snapshot(123, 123)['group_observation'], 'empty')
                registry.error = 13
                self.assertEqual(lock.process_group_snapshot(123, 123)['group_observation'], 'unknown')
        registry = ProcessRegistry(list(range(1000, 1005)))
        with mock.patch.object(lock.sys, 'platform', 'darwin'), mock.patch.object(lock.ctypes, 'CDLL', return_value=registry), mock.patch.object(lock, 'GROUP_MAX_PROCESSES', 4):
            self.assertEqual(lock.process_group_snapshot(123, 123)['group_observation'], 'unknown')

    def test_survivor_truncation_cannot_mean_empty(self):
        registry = ProcessRegistry(list(range(1000, 1257)))
        with mock.patch.object(lock.sys, 'platform', 'darwin'), mock.patch.object(lock.ctypes, 'CDLL', return_value=registry):
            evidence = lock.process_group_snapshot(123, 123)
        self.assertEqual(len(evidence['survivor_pids']), 256)
        self.assertTrue(evidence['survivors_truncated'])
        self.assertEqual(evidence['group_observation'], 'survivors')

    def test_audit_is_bounded_durable_regular_file_and_rejects_torn_tail(self):
        with tempfile.TemporaryDirectory(prefix='direct-audit-', dir=os.environ.get('TMPDIR') or '/tmp') as root:
            operation = dict(id='private-operation', pid=123, pid_started='123.001')
            evidence = Decision().evidence()
            path = Path(root) / 'completion.jsonl'
            lock.append_completion_audit(root, operation, evidence)
            self.assertEqual(len(path.read_bytes().splitlines()), 1)
            with path.open('ab') as stream:
                stream.write(b'{"torn":')
            before = path.read_bytes()
            with self.assertRaises(ValueError):
                lock.append_completion_audit(root, operation, evidence)
            self.assertEqual(path.read_bytes(), before)
            path.unlink()
            target = Path(root) / 'unrelated'
            target.write_bytes(b'private')
            path.symlink_to(target)
            with self.assertRaises(OSError):
                lock.append_completion_audit(root, operation, evidence)
            self.assertEqual(target.read_bytes(), b'private')
            path.unlink()
            with mock.patch.object(lock.os, 'fsync', side_effect=OSError('injected disk failure')), self.assertRaises(OSError):
                lock.append_completion_audit(root, operation, evidence)
            path.unlink()
            with mock.patch.object(lock, 'COMPLETION_MAX_LINE', 8), self.assertRaises(ValueError):
                lock.append_completion_audit(root, operation, evidence)
            self.assertFalse(path.exists())


if __name__ == '__main__':
    unittest.main()
