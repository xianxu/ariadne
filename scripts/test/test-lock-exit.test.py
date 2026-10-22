#!/usr/bin/env python3
"""Private real-make regression tests for direct-suite completion."""
import fcntl
from datetime import datetime, timezone
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time
import unittest

ROOT = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location('direct_exit_helper', ROOT / 'scripts/test-lock.py')
HELPER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(HELPER)


class DirectSuiteExit(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix='direct-suite-exit-', dir=os.environ.get('TMPDIR') or '/tmp')
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.queue = self.root / 'queue'
        self.capacity = self.root / 'capacity'
        self.children = []
        self.detached_groups = []
        self.streams = []
        self.addCleanup(self.cleanup)

    def await_(self, predicate, message):
        deadline = time.monotonic() + 10
        while time.monotonic() < deadline:
            if predicate():
                return
            time.sleep(.02)
        self.fail(message)

    def cleanup(self):
        for group in self.detached_groups:
            try:
                os.killpg(group, signal.SIGKILL)
            except ProcessLookupError:
                pass
        for child in self.children:
            try:
                os.killpg(child.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            child.wait(timeout=10)
        # The existing watcher owns the capacity until it observes make's exit.
        self.await_(self.slot_free, 'private watcher did not release capacity')
        for stream in self.streams:
            stream.close()

    def slot_free(self):
        if not self.capacity.exists():
            return True
        fd = os.open(self.capacity, os.O_RDWR)
        try:
            try:
                fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
                return True
            except BlockingIOError:
                return False
        finally:
            os.close(fd)

    def start_make(self, exit_status=0, survivor=False, shared_group=False, supervisor=False):
        top = self.root / ('make-' + str(len(self.children)))
        top.mkdir()
        recipe = top / 'recipe.py'
        recipe.write_text('import pathlib,time,sys,subprocess\n'
                          + ('child=subprocess.Popen(["sleep","60"], stdin=subprocess.DEVNULL, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)\n'
                             'pathlib.Path("survivor").write_text(str(child.pid))\n' if survivor else '') +
                          'pathlib.Path("ready").write_text("ready")\n'
                          'while not pathlib.Path("release").exists(): time.sleep(.02)\n'
                          f'sys.exit({exit_status})\n')
        (top / 'Makefile').write_text(f'include {ROOT / "Makefile.workflow"}\n'
                                     f'.PHONY: test\ntest:\n\t@{sys.executable} recipe.py\n')
        env = dict(os.environ, WF_TEST_LOCK_FILE=str(self.capacity), WF_SDLC_QUEUE_ROOT=str(self.queue),
                   WF_TEST_SUITES='1', WF_TEST_WEIGHT='1', WF_TEST_LOCK_OWNER=str(top), WF_TEST_LOCK_TIMEOUT='8')
        for key in ('WF_SDLC_QUEUE_ID', 'WF_SDLC_QUEUE_STOP_IF_PAUSED', 'WF_TEST_LOCK', 'WF_TEST_LOCK_HELD_BY', 'MAKEFLAGS', 'MFLAGS'):
            env.pop(key, None)
        out = (top / 'stdout').open('w')
        err = (top / 'stderr').open('w')
        self.streams.extend((out, err))
        command = ['make', '--no-print-directory', 'test']
        if supervisor:
            # Simulate a restartable launcher's lifetime, not live Couch.
            launcher = ('import pathlib,subprocess,sys,time\n'
                        'child=subprocess.Popen(sys.argv[1:], start_new_session=True)\n'
                        'pathlib.Path("make.pid").write_text(str(child.pid))\n'
                        'while not pathlib.Path("supervisor-release").exists(): time.sleep(.02)\n')
            command = [sys.executable, '-c', launcher] + command
        if shared_group:
            command = [sys.executable, '-c', 'import subprocess,sys; sys.exit(subprocess.call(sys.argv[1:]))'] + command
        owner = subprocess.Popen(command, cwd=top, env=env,
                                 stdout=out, stderr=err, start_new_session=True)
        self.children.append(owner)
        self.await_(lambda: (top / 'ready').exists() or owner.poll() is not None, 'make never reached recipe barrier')
        if supervisor and (top / 'make.pid').exists():
            self.detached_groups.append(int((top / 'make.pid').read_text()))
        self.assertIsNone(owner.poll(), (top / 'stderr').read_text())
        operations = list((self.queue / 'operations').glob('*.json'))
        self.assertEqual(len(operations), 1)
        operation = json.loads(operations[0].read_bytes())
        if not shared_group and not supervisor:
            self.assertEqual(operation['pid'], owner.pid)
        self.assertEqual(operation['phase'], 'running')
        self.assertEqual(os.getpgid(operation['pid']), operation['pid'] if supervisor else owner.pid)
        if shared_group:
            self.assertNotEqual(operation['pid'], owner.pid)
        self.assertFalse(self.slot_free(), 'recipe ran without capacity ownership')
        return owner, top, operation

    def finish_make(self, owner, top):
        (top / 'release').touch()
        status = owner.wait(timeout=10)
        self.await_(self.slot_free, 'completed make did not release capacity')
        return status

    def unknown(self, operation):
        path = self.queue / 'unresolved' / (operation['id'] + '.json')
        self.await_(path.exists, 'uncertain ending lost unresolved lifetime')
        unknown = json.loads(path.read_bytes())
        self.assertEqual(unknown['phase'], 'unknown')
        self.assertEqual(unknown['pid_started'], operation['pid_started'])
        self.assertFalse((self.queue / 'completion.jsonl').exists(), 'uncertain ending produced a completion claim')
        return unknown

    @unittest.skipUnless(sys.platform == 'darwin', 'native nonparent exit status currently requires Darwin')
    def test_clean_make_exit_retires_without_unresolved_tombstone(self):
        owner, top, operation = self.start_make()
        released_at = datetime.now(timezone.utc)
        self.assertEqual(self.finish_make(owner, top), 0, (top / 'stderr').read_text())
        self.assertEqual(list((self.queue / 'operations').glob('*.json')), [])
        unresolved = list((self.queue / 'unresolved').glob('*.json'))
        self.assertEqual(unresolved, [], 'clean private-group make exit left unresolved lifetime: ' +
                         '\n'.join(path.read_text() for path in unresolved))
        lines = (self.queue / 'completion.jsonl').read_bytes().splitlines(keepends=True)
        self.assertEqual(len(lines), 1)
        self.assertTrue(lines[0].endswith(b'\n'))
        self.assertLessEqual(len(lines[0]), 1024)
        records = [json.loads(line) for line in lines]
        self.assertEqual([record['id'] for record in records], [operation['id']])
        record = records[0]
        self.assertEqual(record['schema_version'], 1)
        self.assertEqual(record['pid'], owner.pid)
        self.assertEqual(record['pid_started'], operation['pid_started'])
        self.assertEqual(record['exit_status'], 0)
        self.assertEqual(record['process_group'], owner.pid)
        self.assertEqual(record['survivor_pids'], [])
        exit_at = datetime.fromisoformat(record['exit_observed_at'].replace('Z', '+00:00'))
        group_at = datetime.fromisoformat(record['group_observed_at'].replace('Z', '+00:00'))
        self.assertEqual(exit_at.utcoffset().total_seconds(), 0)
        self.assertEqual(group_at.utcoffset().total_seconds(), 0)
        self.assertLessEqual(released_at, exit_at)
        self.assertLessEqual(exit_at, group_at)
        self.assertLessEqual(group_at, datetime.now(timezone.utc))

    def test_nonzero_make_exit_remains_unknown(self):
        owner, top, operation = self.start_make(exit_status=7)
        self.assertNotEqual(self.finish_make(owner, top), 0)
        unknown = self.unknown(operation)
        self.assertNotEqual(unknown.get('exit_status'), 0)

    def test_background_survivor_prevents_retirement(self):
        owner, top, operation = self.start_make(survivor=True)
        survivor = int((top / 'survivor').read_text())
        self.assertEqual(self.finish_make(owner, top), 0)
        os.kill(survivor, 0)
        unknown = self.unknown(operation)
        if sys.platform == 'darwin':
            self.assertIn('survivor_pids', unknown, 'surviving process group was not observed')
            self.assertIn(survivor, unknown['survivor_pids'])
            self.assertEqual(unknown['exit_status'], 0)

    def test_signal_exit_retains_observed_status(self):
        owner, top, operation = self.start_make()
        signalled_at = datetime.now(timezone.utc)
        owner.kill()
        self.assertEqual(owner.wait(timeout=10), -signal.SIGKILL)
        (top / 'release').touch()
        self.await_(self.slot_free, 'signal exit did not release capacity')
        unknown = self.unknown(operation)
        if sys.platform == 'darwin':
            self.assertEqual(unknown['status_observation'], 'observed')
            self.assertEqual(unknown['exit_status'], -signal.SIGKILL)
            observed_at = datetime.fromisoformat(unknown['exit_observed_at'].replace('Z', '+00:00'))
            self.assertLessEqual(signalled_at, observed_at)
            self.assertLessEqual(observed_at, datetime.now(timezone.utc))
        else:
            self.assertEqual(unknown['status_observation'], 'unavailable')
            self.assertNotIn('exit_status', unknown)

    @unittest.skipUnless(sys.platform == 'darwin', 'native nonparent exit status currently requires Darwin')
    def test_simulated_supervisor_exit_does_not_end_make_or_watcher(self):
        supervisor, top, operation = self.start_make(supervisor=True)
        make_pid = operation['pid']
        watcher = json.loads(self.capacity.read_bytes())['holder']
        (top / 'supervisor-release').touch()
        self.assertEqual(supervisor.wait(timeout=10), 0)
        os.kill(make_pid, 0)
        os.kill(watcher, 0)
        self.assertFalse(self.slot_free(), 'launcher exit released the still-running make')
        path = self.queue / 'operations' / (operation['id'] + '.json')
        self.assertEqual(json.loads(path.read_bytes())['phase'], 'running')
        self.assertEqual(list((self.queue / 'unresolved').glob('*.json')), [])
        (top / 'release').touch()
        self.await_(self.slot_free, 'independent watcher did not finish after launcher exit')
        self.assertFalse(path.exists())
        self.assertEqual(list((self.queue / 'unresolved').glob('*.json')), [])
        records = [json.loads(line) for line in (self.queue / 'completion.jsonl').read_text().splitlines()]
        self.assertEqual(len(records), 1)
        self.assertEqual(records[0]['id'], operation['id'])
        self.assertEqual(records[0]['pid'], make_pid)
        self.assertEqual(records[0]['exit_status'], 0)

    @unittest.skipUnless(sys.platform == 'darwin', 'native nonparent exit status currently requires Darwin')
    def test_rapid_recipe_release_keeps_registered_exit_observation(self):
        expected_ids = []
        for attempt in range(5):
            with self.subTest(attempt=attempt):
                owner, top, operation = self.start_make()
                self.assertEqual(self.finish_make(owner, top), 0)
                self.assertEqual(list((self.queue / 'operations').glob('*.json')), [])
                self.assertEqual(list((self.queue / 'unresolved').glob('*.json')), [])
                expected_ids.append(operation['id'])
        lines = (self.queue / 'completion.jsonl').read_bytes().splitlines(keepends=True)
        self.assertEqual([json.loads(line)['id'] for line in lines], expected_ids)
        self.assertTrue(all(line.endswith(b'\n') and len(line) <= 1024 for line in lines))

    def test_shared_make_group_prevents_retirement(self):
        owner, top, operation = self.start_make(shared_group=True)
        self.assertEqual(self.finish_make(owner, top), 0)
        self.unknown(operation)

    def test_watcher_crash_does_not_fabricate_exit_evidence(self):
        owner, top, operation = self.start_make()
        watcher = json.loads(self.capacity.read_bytes())['holder']
        os.kill(watcher, signal.SIGKILL)
        self.await_(self.slot_free, 'killed watcher retained capacity')
        path = self.queue / 'operations' / (operation['id'] + '.json')
        raw = json.loads(path.read_bytes())
        self.assertEqual(raw['phase'], 'running')
        self.assertNotIn('exit_status', raw)
        self.assertNotIn('exit_observed_at', raw)
        self.assertFalse((self.queue / 'completion.jsonl').exists())
        self.assertEqual(self.finish_make(owner, top), 0)
        self.assertTrue(path.exists(), 'watcher crash erased unknown completion')

    def test_cancellation_wins_before_recipe_completion(self):
        owner, top, operation = self.start_make()
        cancel = self.queue / 'cancel' / (operation['id'] + '.json')
        mutex = os.open(self.queue / 'control.lock', os.O_RDWR)
        try:
            fcntl.flock(mutex, fcntl.LOCK_EX)
            temporary = cancel.with_suffix('.tmp')
            temporary.write_text(json.dumps(dict(schema_version=1, id=operation['id'], reason='private cancellation fixture')))
            temporary.replace(cancel)
        finally:
            os.close(mutex)
        path = self.queue / 'operations' / (operation['id'] + '.json')
        self.await_(lambda: not path.exists() or json.loads(path.read_bytes())['phase'] == 'interrupting',
                    'watcher did not acknowledge cancellation')
        self.finish_make(owner, top)
        self.unknown(operation)
        interrupted = self.queue / 'interrupted' / (operation['id'] + '.json')
        self.assertEqual(json.loads(interrupted.read_bytes())['outcome'], 'interrupted')

    def test_original_incidents_have_no_positive_completion_evidence(self):
        directory = ROOT / 'scripts/test/fixtures/direct-suite-completion'
        provenance = json.loads((directory / 'provenance.json').read_bytes())
        self.assertEqual(len(provenance['records']), 3)
        for record in provenance['records']:
            raw = (directory / (record['id'] + '.json')).read_bytes()
            self.assertEqual(hashlib.sha256(raw).hexdigest(), record['sha256'])
            operation = HELPER._queue_operation(HELPER.queue_strict_object(raw))
            self.assertEqual(operation['phase'], 'unknown')
            self.assertNotIn('exit_status', operation)
            self.assertNotIn('exit_observed_at', operation)


if __name__ == '__main__':
    unittest.main()
