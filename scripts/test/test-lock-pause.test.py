#!/usr/bin/env python3
"""Private real-flock/process tests for the machine pause admission seam."""
import fcntl
import importlib.util
import json
import os
from pathlib import Path
import signal
import subprocess
import tempfile
import threading
import time
import unittest
from unittest import mock

ROOT = Path(__file__).resolve().parents[2]
HELPER = ROOT / 'scripts/test-lock.py'
spec = importlib.util.spec_from_file_location('pause_lock_helper', HELPER)
lock = importlib.util.module_from_spec(spec)
spec.loader.exec_module(lock)
QUEUE_READS = importlib.util.spec_from_file_location('queue_reads', ROOT / 'scripts/test/fixtures/queue_reads.py')
queue_reads = importlib.util.module_from_spec(QUEUE_READS)
QUEUE_READS.loader.exec_module(queue_reads)


class QueuePause(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(prefix='suite-pause-', dir=os.environ.get('TMPDIR') or '/tmp')
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.queue = self.root / 'queue'
        self.top = self.root / 'pair-slot5' / 'pair'
        self.top.mkdir(parents=True)
        self.env = dict(os.environ, WF_TEST_LOCK_FILE=str(self.root / 'capacity'), WF_SDLC_QUEUE_ROOT=str(self.queue),
                        WF_TEST_SUITES='1', WF_TEST_LOCK_OWNER=str(self.top), WF_TEST_LOCK_TIMEOUT='8')
        for name in ('WF_SDLC_QUEUE_ID', 'WF_SDLC_QUEUE_STOP_IF_PAUSED', 'WF_TEST_LOCK', 'WF_TEST_LOCK_HELD_BY'):
            self.env.pop(name, None)
        self.children = []
        self.streams = []
        self.descendants = []
        self.addCleanup(self.cleanup)

    def cleanup(self):
        for child in self.children:
            if child.poll() is None:
                child.kill()
        for child in self.children:
            child.wait()
        # Watchers are daemon descendants. Their slot flocks are their last
        # lifetime resources: wait for release before deleting private state.
        end = time.monotonic() + 8
        while True:
            busy = False
            for path in self.root.glob('capacity*'):
                if not path.is_file():
                    continue
                fd = os.open(path, os.O_RDWR)
                try:
                    fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
                except BlockingIOError:
                    busy = True
                finally:
                    os.close(fd)
            if not busy:
                break
            if time.monotonic() >= end:
                self.fail('private watcher did not release its lifetime resources')
            time.sleep(.02)
        for stream in self.streams:
            stream.close()
        for pid in self.descendants:
            try:
                os.kill(pid, signal.SIGKILL)
            except ProcessLookupError:
                pass

    def owner(self):
        p = subprocess.Popen(['sleep', '60'])
        self.children.append(p)
        return p

    def start(self, name, env=None):
        owner = self.owner()
        out = (self.root / (name + '.out')).open('w')
        err = (self.root / (name + '.err')).open('w')
        self.streams.extend((out, err))
        helper = subprocess.Popen(['python3', str(HELPER), 'acquire', '--watch', str(owner.pid)],
                                  env=env or self.env, stdout=out, stderr=err)
        self.children.append(helper)
        return owner, helper

    def await_(self, predicate, why):
        end = time.monotonic() + 8
        while time.monotonic() < end:
            if predicate():
                return
            time.sleep(.02)
        self.fail(why + '\n' + '\n'.join(p.read_text() for p in self.root.glob('*.err')))

    def committed(self, predicate):
        return queue_reads.committed(self.queue, predicate)

    def committed_read(self, path):
        return queue_reads.committed_read(self.await_, self.queue, path)

    def paused(self, owners):
        self.queue.mkdir(exist_ok=True)
        mutex = os.open(self.queue / 'control.lock', os.O_CREAT | os.O_RDWR, 0o600)
        try:
            fcntl.flock(mutex, fcntl.LOCK_EX)
            body = dict(schema_version=1, pause_owners=owners)
            temporary = self.queue / 'pause.tmp'
            temporary.write_text(json.dumps(body))
            temporary.replace(self.queue / 'pause.json')
        finally:
            os.close(mutex)

    def intent(self, operation_id, reason='fast fixture'):
        directory = self.queue / 'cancel'
        directory.mkdir(exist_ok=True)
        temporary = directory / 'fixture.tmp'
        temporary.write_text(json.dumps(dict(schema_version=1, id=operation_id, reason=reason)))
        temporary.replace(directory / (operation_id + '.json'))

    def managed_operation(self, operation_id='managed', **changes):
        operations = self.queue / 'operations'
        operations.mkdir(parents=True, exist_ok=True)
        op = dict(schema_version=1, id=operation_id, workspace='pair:5', kinds=['suite'], phase='waiting',
                  pid=os.getpid(), pid_started=lock.queue_process_start(os.getpid()), managed=True,
                  started_at='2026-10-20T10:00:00Z', boot_id=lock.queue_boot_id(), root=str(self.top.resolve()), detail='')
        op.update(changes)
        path = operations / (operation_id + '.json')
        fd = os.open(path, os.O_RDWR | os.O_CREAT, 0o600)
        self.addCleanup(os.close, fd)
        fcntl.flock(fd, fcntl.LOCK_EX)
        raw = json.dumps(op).encode()[:-1] + b',"future": { "n" : 1e99999 }}'
        os.write(fd, raw)
        return fd, path

    def test_pause_holds_fifo_waiter_and_resume_admits(self):
        self.paused([dict(owner='A', reason='operator data ; $(quoted)', cancel_running=False)])
        owner, helper = self.start('waiter')
        self.await_(lambda: 'queue paused' in (self.root / 'waiter.err').read_text(), 'pause was not visible')
        self.assertIsNone(helper.poll())
        self.assertEqual((self.root / 'waiter.out').read_text(), '')
        operations = list((self.queue / 'operations').glob('*.json'))
        self.assertEqual(len(operations), 1)
        self.assertEqual(json.loads(self.committed_read(operations[0]))['phase'], 'waiting')
        self.paused([])
        self.await_(lambda: helper.poll() is not None, 'resume did not admit queued operation')
        self.assertEqual(helper.returncode, 0)
        self.assertEqual((self.root / 'waiter.out').read_text().strip(), 'acquired')
        self.assertEqual(json.loads(self.committed_read(operations[0]))['phase'], 'running')
        owner.terminate()
        owner.wait()

    def test_new_suite_boundary_can_return_paused_without_admission(self):
        self.paused([dict(owner='A', reason='drain', cancel_running=False)])
        env = dict(self.env, WF_SDLC_QUEUE_STOP_IF_PAUSED='1')
        _, helper = self.start('boundary', env)
        self.await_(lambda: helper.poll() is not None, 'admitted merge did not stop at paused suite boundary')
        self.assertEqual(helper.returncode, 0)
        self.assertEqual((self.root / 'boundary.out').read_text().strip(), 'paused')
        self.assertEqual(list((self.queue / 'operations').glob('*.json')), [])

    def test_strict_raw_codec_preserves_large_unknown_values(self):
        raw = b'{"phase":"waiting","future": { "number" : 1e99999 }}'
        fields = lock.queue_strict_object(raw)
        changed = lock.queue_encode_object(fields, phase='running')
        self.assertIn(b'{ "number" : 1e99999 }', changed)
        for bad in (b'{"x":1,"x":2}', b'{"x":{"n":1,"n":2}}', b'[]', b'{} {}', b'{"x":"\xff"}'):
            with self.assertRaises(ValueError):
                lock.queue_strict_object(bad)

    def test_raw_member_scan_copies_only_linear_input(self):
        # Count real substring bytes, not wall time or a stand-in parser.
        original = lock._queue_json
        for count in (8, 16):
            copied = [0]
            class MeasuredText(str):
                def __getitem__(self, key):
                    value = super().__getitem__(key)
                    if isinstance(key, slice):
                        copied[0] += len(value)
                    return value
            def measured(data):
                text, parsed = original(data)
                return MeasuredText(text), parsed
            raw = json.dumps({f'future-{n}': n for n in range(count)}).encode()
            with mock.patch.object(lock, '_queue_json', side_effect=measured):
                self.assertEqual(len(lock.queue_strict_object(raw)), count)
            self.assertLessEqual(copied[0], 2 * len(raw), f'{count} fields recopied remaining suffixes')

    def test_kernel_incarnation_and_boot_identity_are_semantic(self):
        self.assertTrue(lock.queue_process_start(os.getpid()))
        self.assertEqual(lock.queue_process_start(os.getpid()), lock.queue_process_start(os.getpid()))
        self.assertTrue(lock.queue_boot_id())

    def test_managed_operation_is_borrowed_and_unknown_fields_survive_admission(self):
        fd, path = self.managed_operation()
        owner, helper = self.start('managed', dict(self.env, WF_SDLC_QUEUE_ID='managed'))
        self.await_(lambda: helper.poll() is not None, 'managed helper did not admit')
        self.assertEqual(helper.returncode, 0)
        raw = self.committed_read(path)
        self.assertIn(b'{ "n" : 1e99999 }', raw)
        self.assertEqual(json.loads(raw)['phase'], 'running')
        slot = json.loads((self.root / 'capacity').read_text())
        self.assertEqual(slot['queue_operation_id'], 'managed')
        self.assertEqual(slot['queue_pid'], os.getpid())
        self.assertEqual(slot['queue_root'], str(self.top.resolve()))
        self.intent('managed')
        # A managed watcher has no authority over the keeper/make; Go cancels
        # its actual owned group and retains its own interruption evidence.
        time.sleep(.25)
        self.assertIsNone(owner.poll())
        owner.terminate()
        owner.wait()
        self.assertTrue(path.exists())
        probe = os.open(path, os.O_RDWR)
        try:
            with self.assertRaises(BlockingIOError):
                fcntl.flock(probe, fcntl.LOCK_EX | fcntl.LOCK_NB)
        finally:
            os.close(probe)

    def test_managed_binding_rejects_root_and_incarnation_mismatch(self):
        for name, changes in [('root', dict(root='/outside')), ('incarnation', dict(pid_started='wrong'))]:
            _, path = self.managed_operation(name, **changes)
            _, helper = self.start(name, dict(self.env, WF_SDLC_QUEUE_ID=name))
            self.await_(lambda: helper.poll() is not None, 'invalid managed binding hung')
            self.assertNotEqual(helper.returncode, 0)
            self.assertEqual(json.loads(self.committed_read(path))['phase'], 'waiting')

    def test_fast_cancellation_preserves_unknown_child_lifetime(self):
        self.fast_cancellation(self.env)

    def test_fast_cancellation_waits_for_the_whole_retirement(self):
        # retire writes unresolved/, then interrupted/, then unlinks the
        # operation; the shim slows the unlink so an unlocked read would see
        # the step before it (#592).
        self.fast_cancellation(queue_reads.slow_writes(self.env))

    def fast_cancellation(self, env):
        child_file = self.root / 'surviving-child'
        # The pid lands by rename: the wait below must never read a partial file.
        owner = subprocess.Popen(['python3', '-c',
            'import os,pathlib,subprocess,sys,time; p=subprocess.Popen(["sleep","60"]); '
            't=pathlib.Path(sys.argv[1] + ".tmp"); t.write_text(str(p.pid)); os.replace(t, sys.argv[1]); '
            'time.sleep(60)', str(child_file)])
        self.children.append(owner)
        self.await_(child_file.exists, 'child owner failed to start')
        child_pid = int(child_file.read_text())
        self.descendants.append(child_pid)
        result = subprocess.run(['python3', str(HELPER), 'acquire', '--watch', str(owner.pid)], env=env,
                                capture_output=True, text=True, timeout=8)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.strip(), 'acquired')
        path = next((self.queue / 'operations').glob('*.json'))
        operation_id = path.stem
        self.paused([dict(owner='fast', reason='fixture', cancel_running=True)])
        self.intent(operation_id)
        self.await_(lambda: owner.poll() is not None, 'fenced owner did not receive fast cancellation')
        owner.wait()
        unresolved = self.queue / 'unresolved' / path.name
        interrupted = self.queue / 'interrupted' / path.name
        self.await_(self.committed(lambda: unresolved.exists() and interrupted.exists() and not path.exists()),
                    'unknown/interrupted evidence missing')
        os.kill(child_pid, 0)
        unknown = json.loads(unresolved.read_text())
        evidence = json.loads(interrupted.read_text())
        self.assertEqual(unknown['phase'], 'unknown')
        self.assertEqual(evidence['outcome'], 'interrupted')
        self.assertEqual(evidence['operation']['phase'], 'unknown')
        self.assertEqual(evidence['id'], operation_id)
        self.assertTrue((self.queue / 'cancel' / path.name).exists())
        self.assertFalse(path.exists())

    def test_priority_provider_and_kernel_observers_stay_outside_admission_gate(self):
        for count in (8, 16):
            with self.subTest(count=count), mock.patch.dict(os.environ, self.env, clear=True):
                gate = lock.QueueGate(str(self.queue), str(self.top), os.getpid())
                rec = dict(repo='pair', worktree=str(self.top), pid=os.getpid(), started='fixture')
                with gate.locked():
                    rec.update(gate.binding())
                line = lock.Tickets(str(self.root / ('count-' + str(count))), rec)
                slot = os.open(self.root / ('count-slot-' + str(count)), os.O_RDWR | os.O_CREAT, 0o600)
                held = os.open(self.root / ('count-slot-' + str(count)), os.O_RDWR)
                fcntl.flock(held, fcntl.LOCK_EX)
                token = os.open(self.root / ('count-token-' + str(count)), os.O_RDWR | os.O_CREAT, 0o600)
                calls = []
                def provider(records, now):
                    probe = os.open(self.queue / 'control.lock', os.O_RDWR)
                    try:
                        fcntl.flock(probe, fcntl.LOCK_EX | fcntl.LOCK_NB)
                    finally:
                        os.close(probe)
                    calls.append(len(records))
                    return {}
                try:
                    with mock.patch.object(lock, 'queue_process_start', side_effect=AssertionError('incarnation inspected during admission')), \
                         mock.patch.object(lock, 'queue_boot_id', side_effect=AssertionError('boot inspected during admission')):
                        evidence = mock.Mock(deadlines=provider)
                        for _ in range(count):
                            result = line.admission([slot], token, 1, evidence, gate)
                            self.assertFalse(result[0])
                    self.assertEqual(len(calls), count)
                finally:
                    line.release()
                    gate.retire()
                    gate.close()
                    for fd in (slot, held, token):
                        os.close(fd)

    def test_contended_allocation_never_pins_global_gate(self):
        with mock.patch.dict(os.environ, self.env, clear=True):
            gate = lock.QueueGate(str(self.queue), str(self.top), os.getpid())
            predecessor = lock.Tickets(str(self.root / 'contended'), dict(pid=os.getpid(),repo='older',worktree=str(self.top),started='fixture'))
            line = lock.Tickets(str(self.root / 'contended'), dict(pid=os.getpid(),repo='fixture',worktree=str(self.top),started='fixture'))
            original_ticket = line.path
            allocation = os.open(self.root / 'contended.tickets' / 'alloc', os.O_RDWR)
            slot = os.open(self.root / 'slot', os.O_RDWR | os.O_CREAT, 0o600)
            token = os.open(self.root / 'token', os.O_RDWR | os.O_CREAT, 0o600)
            ready, finished = threading.Event(), threading.Event()
            outcomes = []
            def provider(records, now):
                fcntl.flock(allocation, fcntl.LOCK_EX)
                ready.set()  # Hold alloc only after admission's preview read.
                return {}
            def admission():
                try:
                    outcomes.append(line.admission([slot], token, 1, mock.Mock(deadlines=provider), gate))
                finally:
                    finished.set()
            thread = threading.Thread(target=admission)
            thread.start()
            try:
                self.assertTrue(ready.wait(2), 'priority preview did not finish')
                self.assertTrue(finished.wait(2), 'allocation wait held the global gate')
                self.assertFalse(outcomes[0][0])
                self.assertIsNone(outcomes[0][2], 'unobserved order fabricated first place')
                self.paused([dict(owner='responsive', reason='contention', cancel_running=False)])
                self.assertEqual(line.path, original_ticket)
                self.assertTrue(Path(original_ticket).exists())
                fcntl.flock(allocation, fcntl.LOCK_UN)
                self.paused([])
                observed = line.admission([slot], token, 1, mock.Mock(deadlines=lambda *_: {}), gate)
                self.assertFalse(observed[0], 'waiter passed its predecessor')
                self.assertEqual(len(observed[2]), 1, 'actual second place was not restored')
                predecessor.release()
                predecessor = None
                admitted = line.admission([slot], token, 1, mock.Mock(deadlines=lambda *_: {}), gate)
                self.assertEqual(admitted[0], [0], 'same ticket did not admit in FIFO order')
                self.assertEqual(line.path, original_ticket)
            finally:
                fcntl.flock(allocation, fcntl.LOCK_UN)
                thread.join(3)
                line.release()
                if predecessor is not None:
                    predecessor.release()
                gate.retire()
                gate.close()
                for fd in (allocation, slot, token):
                    os.close(fd)

    def test_pause_and_ticket_revalidation_leave_order_unobserved(self):
        for reason in ('paused', 'pause-after-preview', 'changed-tickets'):
            with self.subTest(reason=reason), mock.patch.dict(os.environ, self.env, clear=True):
                gate = lock.QueueGate(str(self.queue), str(self.top), os.getpid())
                rec = dict(pid=os.getpid(), repo='fixture', worktree=str(self.top), started='fixture')
                base = str(self.root / reason)
                line = lock.Tickets(base, rec)
                others = []
                slot = os.open(self.root / (reason + '.slot'), os.O_CREAT | os.O_RDWR, 0o600)
                token = os.open(self.root / (reason + '.token'), os.O_CREAT | os.O_RDWR, 0o600)
                def provider(*_):
                    if reason == 'changed-tickets':
                        others.append(lock.Tickets(base, rec))
                    else:
                        self.paused([dict(owner='fixture', reason='hold', cancel_running=False)])
                    return {}
                try:
                    self.paused([dict(owner='fixture', reason='hold', cancel_running=False)] if reason == 'paused' else [])
                    evidence = mock.Mock(deadlines=mock.Mock(side_effect=provider))
                    result = line.admission([slot], token, 1, evidence, gate)
                    self.assertFalse(result[0])
                    self.assertIsNone(result[2], 'no validated order must not mean zero predecessors')
                    if reason == 'paused':
                        evidence.deadlines.assert_not_called()
                finally:
                    for other in others:
                        other.release()
                    line.release()
                    gate.retire()
                    gate.close()
                    os.close(slot)
                    os.close(token)

    def diagnostic_reports(self, observations, step=1):
        """Drive the actual acquire renderer; observations are its scheduling seam."""
        class Finished(Exception):
            pass
        messages, calls, clock = [], [], [1000]
        def admission(*_):
            result = observations[len(calls)]
            calls.append(result)
            return [], [], result, False, 0
        def sleep(_):
            clock[0] += step
            if len(calls) == len(observations):
                raise Finished()
        env = dict(self.env, WF_TEST_LOCK_TIMEOUT='0')
        with mock.patch.dict(os.environ, env, clear=True), \
             mock.patch.object(lock, 'priority_subject', return_value=None), \
             mock.patch.object(lock.Tickets, 'admission', side_effect=admission), \
             mock.patch.object(lock, 'holder', wraps=lock.holder) as holder, \
             mock.patch.object(lock, 'log', side_effect=messages.append), \
             mock.patch.object(lock.time, 'time', side_effect=lambda: clock[0]), \
             mock.patch.object(lock.time, 'sleep', side_effect=sleep):
            with self.assertRaises(Finished):
                lock.acquire(os.getpid())
            return messages, len(calls), holder.call_count

    def test_unobserved_diagnostics_never_invent_position_or_churn(self):
        older = dict(pid=123, repo='older', worktree='/fixture/older', started='fixture')
        for count in (8, 16):
            for observed in (False, True):
                with self.subTest(count=count, observed=observed):
                    sequence = [[older, older] if observed and i % 2 == 0 else None for i in range(count)]
                    messages, calls, holder_calls = self.diagnostic_reports(sequence)
                    self.assertEqual(calls, count)
                    self.assertEqual(len(messages), 1, 'unchanged observations churned the position log')
                    self.assertEqual(holder_calls, 0, 'unobserved order probed the old head token')
                    self.assertNotIn('position 1', messages[0])
                    self.assertIn('position 3 in line' if observed else 'position unknown', messages[0])

    def test_periodic_report_labels_stale_position_and_real_change_updates(self):
        older = dict(pid=123, repo='older', worktree='/fixture/older', started='fixture')
        messages, _, _ = self.diagnostic_reports([[older, older], None], step=lock.REPORT_EVERY + 1)
        self.assertEqual(len(messages), 2)
        self.assertIn('last observed position 3 in line', messages[-1])
        messages, _, _ = self.diagnostic_reports([[older, older], None, [older], None])
        self.assertEqual(len(messages), 2)
        self.assertIn('position 3 in line', messages[0])
        self.assertIn('position 2 in line', messages[1])

    def test_cancel_waiter_while_allocation_is_held(self):
        directory = self.root / 'capacity.tickets'
        directory.mkdir()
        allocation = os.open(directory / 'alloc', os.O_RDWR | os.O_CREAT, 0o600)
        try:
            fcntl.flock(allocation, fcntl.LOCK_EX)
            _, helper = self.start('contended-cancel')
            self.await_(lambda: list((self.queue / 'operations').glob('*.json')), 'waiter did not publish its lease')
            helper.terminate()
            self.await_(lambda: helper.poll() is not None, 'cancelled waiter stayed blocked on allocation')
            self.assertNotEqual(helper.returncode, 0)
            self.assertEqual(list((self.queue / 'operations').glob('*.json')), [])
            self.assertEqual((self.root / 'contended-cancel.out').read_text(), '')
        finally:
            os.close(allocation)

    def test_watcher_reopens_global_mutex_after_fork(self):
        with mock.patch.dict(os.environ, self.env, clear=True):
            gate = lock.QueueGate(str(self.queue), str(self.top), os.getpid())
            read, write = os.pipe()
            with gate.locked():
                child = os.fork()
                if child == 0:
                    os.close(read)
                    try:
                        gate.after_fork()
                        # Parent still holds the original open description;
                        # this independent mutex descriptor must contend.
                        available = lock.try_lock(gate.mutex)
                        os.write(write, b'bad' if available else b'blocked')
                    finally:
                        os._exit(0)
                os.close(write)
                self.assertEqual(os.read(read, 16), b'blocked')
            os.close(read)
            os.waitpid(child, 0)
            gate.retire()
            gate.close()

    def test_operation_id_grammar_matches_go_writer(self):
        _, path = self.managed_operation()
        fields = lock._queue_read(path)
        for value in ('one.dot:two', 'a' * 256):
            encoded = lock.queue_encode_object(fields, id=value)
            self.assertEqual(lock._queue_operation(lock.queue_strict_object(encoded))['id'], value)
        for value in ('.', '..', 'a' * 257, 'a/b'):
            with self.assertRaises(ValueError):
                lock._queue_operation(lock.queue_strict_object(lock.queue_encode_object(fields, id=value)))

    def test_unmanaged_workspace_is_unknown_without_verified_identity(self):
        with mock.patch.dict(os.environ, self.env, clear=True):
            gate = lock.QueueGate(str(self.queue), str(self.top), os.getpid())
            try:
                with gate.locked():
                    _, op = gate.operation()
                self.assertEqual(op['workspace'], '')
                self.assertIn(str(self.top), op['detail'])
            finally:
                gate.retire()
                gate.close()

    def test_recycled_kernel_identity_never_signals_owner(self):
        with mock.patch.dict(os.environ, self.env, clear=True):
            gate = lock.QueueGate(str(self.queue), str(self.top), os.getpid())
            self.assertTrue(gate.try_admit(lambda: True))
            self.intent(gate.id)
            with mock.patch.object(lock, 'process_alive', side_effect=[True, False]), \
                 mock.patch.object(lock, 'queue_process_start', return_value='recycled'), \
                 mock.patch.object(lock.os, 'kill') as kill:
                lock.watch_queue_owner(gate, os.getpid())
            kill.assert_not_called()
            unresolved = self.queue / 'unresolved' / (gate.id + '.json')
            self.assertEqual(json.loads(unresolved.read_text())['phase'], 'unknown')


if __name__ == '__main__':
    unittest.main()
