#!/usr/bin/env python3
"""Real helper/process/flock integration. Go tests supply the production provider."""
import importlib.util
import json
from types import SimpleNamespace
import os
from pathlib import Path
import signal
import subprocess
import tempfile
import time
import unittest
from unittest import mock

ROOT = Path(__file__).resolve().parents[2]
HELPER = ROOT / 'scripts/test-lock.py'


class SuiteQueuePriority(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(prefix='suite-priority-', dir=os.environ.get('TMPDIR', '/tmp'))
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.common = self.root / 'repo.git'
        self.log = self.common / 'sdlc/landing/grants.jsonl'
        self.log.parent.mkdir(parents=True)
        self.provider = os.environ.get('SUITE_PRIORITY_TEST_PROVIDER', str(ROOT / 'bin/sdlc'))
        self.env = dict(os.environ, WF_TEST_LOCK_FILE=str(self.root / 'lock'), WF_TEST_SUITES='2',
                        WF_TEST_PRIORITY_PROVIDER=self.provider, WF_TEST_LOCK_TIMEOUT='15')
        for key in ('WF_TEST_LOCK', 'WF_TEST_LOCK_HELD_BY', 'WF_TEST_LOCK_OWNER', 'WF_TEST_PRIORITY_SUBJECT'):
            self.env.pop(key, None)
        self.children = []
        self.streams = []
        self.addCleanup(self.cleanup)

    def cleanup(self):
        for child in self.children:
            if child.poll() is None:
                child.send_signal(signal.SIGCONT)
                child.kill()
        for child in self.children:
            child.wait()
        for stream in self.streams:
            stream.close()

    def grant(self, issue, expires=None, revoked=False, granted=None):
        def stamp(at):
            from datetime import datetime, timezone
            return datetime.fromtimestamp(at, timezone.utc).isoformat()
        now = time.time()
        record = dict(issue=issue, grantor=dict(address='ops:0', slot=0), reason='fixture fleet blocker',
                      granted=stamp(granted if granted is not None else now - 10),
                      expires=stamp(expires if expires is not None else now + 300), revoked=revoked)
        with self.log.open('a') as stream:
            stream.write(json.dumps(record) + '\n')

    def owner(self):
        p = subprocess.Popen(['sleep', '120'])
        self.children.append(p)
        return p

    def start(self, name, issue=0, weight=1, issues=None):
        owner = self.owner()
        out = (self.root / (name + '.out')).open('w')
        err = (self.root / (name + '.err')).open('w')
        self.streams.extend([out, err])
        subject = dict(common_dir=str(self.common), issue=issue)
        if issues is not None:
            subject['issues'] = issues
        env = dict(self.env, WF_TEST_PRIORITY_SUBJECT=json.dumps(subject))
        helper = subprocess.Popen(['python3', str(HELPER), 'acquire', '--watch', str(owner.pid), '--slots', str(weight)],
                                  env=env, stdout=out, stderr=err)
        self.children.append(helper)
        return owner, helper

    def contents(self, name, suffix):
        return (self.root / (name + suffix)).read_text()

    def await_(self, predicate, message):
        deadline = time.monotonic() + 10
        while time.monotonic() < deadline:
            if predicate():
                return
            time.sleep(.025)
        errors = '\n'.join(p.name + ': ' + p.read_text() for p in self.root.glob('*.err'))
        self.fail(message + '\n' + errors)

    def acquired(self, name):
        return self.contents(name, '.out').strip() == 'acquired'

    def wait_queued(self, name):
        self.await_(lambda: 'position ' in self.contents(name, '.err'), name + ' never queued')

    def release(self, owner):
        owner.terminate()
        owner.wait()

    def fill(self):
        a, _ = self.start('holder-a')
        self.await_(lambda: self.acquired('holder-a'), 'first holder failed')
        b, _ = self.start('holder-b')
        self.await_(lambda: self.acquired('holder-b'), 'second holder failed')
        return a, b

    def test_train_late_grant(self):
        self.train_grant_lifecycle('keep')

    def test_train_late_revoke(self):
        self.train_grant_lifecycle('revoke')

    def test_train_clock_expiry(self):
        self.train_grant_lifecycle('expire')

    def train_grant_lifecycle(self, ending):
        a, b = self.fill()
        normal, _ = self.start('normal', 1)
        self.wait_queued('normal')
        train, _ = self.start('train', issues=[2, 3])
        self.wait_queued('train')
        self.grant(99)
        self.assertFalse(self.acquired('train'))
        self.grant(3, expires=time.time() + (3 if ending == 'expire' else 300))
        self.await_(lambda: 'fleet-blocking' in self.contents('train', '.err'), 'later car grant not observed')
        if ending == 'revoke':
            self.grant(3, revoked=True)
        if ending != 'keep':
            self.await_(lambda: self.contents('train', '.err').count('normal FIFO') >= 2,
                        'train did not lose priority')
        self.release(a)
        first = 'train' if ending == 'keep' else 'normal'
        other = 'normal' if ending == 'keep' else 'train'
        self.await_(lambda: self.acquired(first) or self.acquired(other), 'no next admission')
        self.assertTrue(self.acquired(first), 'wrong admission after train grant change')
        self.assertFalse(self.acquired(other), 'both admitted into one free slot')
        self.release(train if ending == 'keep' else normal)
        self.await_(lambda: self.acquired(other), 'remaining waiter starved')
        self.release(normal if ending == 'keep' else train)
        self.release(b)


    def test_priority_takes_next_free_slot(self):
        self.grant(2)
        a, b = self.fill()
        normal, _ = self.start('normal', 1)
        self.wait_queued('normal')
        priority, _ = self.start('priority', 2)
        self.wait_queued('priority')
        self.release(a)
        self.await_(lambda: self.acquired('priority') or self.acquired('normal'), 'no next admission')
        self.assertTrue(self.acquired('priority'), 'older normal head overtook active granted issue')
        self.assertFalse(self.acquired('normal'))
        self.assertIsNone(b.poll(), 'running suite was preempted')
        self.assertIn('fleet-blocking', self.contents('priority', '.err'))
        self.release(priority)
        self.await_(lambda: self.acquired('normal'), 'normal did not follow priority')

    def test_expired_grant_returns_to_normal_fifo(self):
        self.grant(2)
        a, b = self.fill()
        _, _ = self.start('normal', 1)
        self.wait_queued('normal')
        _, _ = self.start('priority', 2)
        self.wait_queued('priority')
        self.await_(lambda: 'fleet-blocking' in self.contents('priority', '.err'), 'grant not observed')
        self.grant(2, expires=time.time() - 1)
        self.await_(lambda: 'normal FIFO' in self.contents('priority', '.err'), 'expired grant retained priority')
        self.release(a)
        self.await_(lambda: self.acquired('normal') or self.acquired('priority'), 'no admission after expiry')
        self.assertTrue(self.acquired('normal'), 'expired issue bypassed older normal waiter')
        self.assertFalse(self.acquired('priority'))

    def test_unchanged_log_expiry_unblocks_older_normal(self):
        a, b = self.fill()
        self.start('normal', 1)
        self.wait_queued('normal')
        self.start('priority', 2)
        self.wait_queued('priority')
        self.grant(2, expires=time.time() + 5)
        self.await_(lambda: 'fleet-blocking' in self.contents('priority', '.err'), 'short grant not observed')
        signature = self.log.stat().st_mtime_ns
        self.await_(lambda: self.contents('priority', '.err').count('normal FIFO') >= 2, 'clock expiry retained priority')
        self.assertEqual(self.log.stat().st_mtime_ns, signature, 'expiry test changed grant log')
        self.release(a)
        self.await_(lambda: self.acquired('normal'), 'expired grant starved older normal suite')
        self.assertFalse(self.acquired('priority'))

    def test_direct_make_uses_checkout_subject_and_supplying_provider(self):
        import shutil
        fixture = self.root / 'make-repo'
        fixture.mkdir()
        for args in (['init', '-q', '-b', '000002-priority'],
                     ['-c', 'user.name=T', '-c', 'user.email=t@e.com', 'commit', '--allow-empty', '-qm', 'fixture']):
            subprocess.run(['git', '-C', str(fixture)] + args, check=True, capture_output=True)
        self.common = (fixture / '.git').resolve()
        self.log = self.common / 'sdlc/landing/grants.jsonl'
        self.log.parent.mkdir(parents=True)
        self.grant(2)
        shutil.copy(ROOT / 'Makefile.workflow', fixture / 'Makefile.workflow')
        (fixture / 'scripts').mkdir()
        (fixture / 'scripts/test-lock.py').symlink_to(HELPER)
        (fixture / 'bin').mkdir()
        (fixture / 'bin/sdlc').symlink_to(self.provider)
        admitted = self.root / 'make-admitted'
        (fixture / 'Makefile').write_text('WF_TEST_WEIGHT := 1\ninclude Makefile.workflow\ntest:\n\t@touch "' + str(admitted) + '"\n\t@sleep 60\n')
        a, b = self.fill()
        self.start('normal', 1)
        self.wait_queued('normal')
        out = (self.root / 'make.out').open('w')
        err = (self.root / 'make.err').open('w')
        self.streams.extend([out, err])
        env = dict(self.env)
        env.pop('WF_TEST_PRIORITY_PROVIDER', None)  # Make must select its supplier.
        make = subprocess.Popen(['make', 'test'], cwd=fixture, env=env, stdout=out, stderr=err, start_new_session=True)
        self.children.append(make)
        def stop_make():
            try:
                os.killpg(make.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
        self.addCleanup(stop_make)
        self.await_(lambda: 'fleet-blocking' in self.contents('make', '.err'), 'direct make did not bind its issue')
        self.release(a)
        self.await_(admitted.exists, 'direct make did not take next slot')
        self.assertFalse(self.acquired('normal'))

    def test_revocation_and_late_grant_reorder_live_waiters(self):
        a, b = self.fill()
        normal, _ = self.start('normal', 1)
        self.wait_queued('normal')
        priority, _ = self.start('priority', 2)
        self.wait_queued('priority')
        self.grant(2)
        self.await_(lambda: 'fleet-blocking' in self.contents('priority', '.err'), 'late grant not observed')
        self.grant(2, revoked=True)
        self.await_(lambda: self.contents('priority', '.err').count('normal FIFO') >= 2, 'revoke not observed')
        self.release(a)
        self.await_(lambda: self.acquired('normal'), 'revoked grant blocked normal')
        self.assertFalse(self.acquired('priority'))

    def test_priority_fifo_stopped_waiter(self):
        self.grant(2)
        self.grant(3)
        a, b = self.fill()
        self.start('normal', 1)
        self.wait_queued('normal')
        p1, h1 = self.start('priority1', 2)
        self.wait_queued('priority1')
        p2, h2 = self.start('priority2', 3)
        self.wait_queued('priority2')
        # Diagnostics are emitted after leaving the allocation critical section.
        h1.send_signal(signal.SIGSTOP)
        self.release(a)
        time.sleep(1.2)
        self.assertFalse(self.acquired('priority2'), 'priority FIFO overtaken while first waiter stopped')
        self.assertFalse(self.acquired('normal'))
        h1.send_signal(signal.SIGCONT)
        self.await_(lambda: self.acquired('priority1'), 'oldest priority not admitted')
        self.release(p1)
        self.await_(lambda: self.acquired('priority2'), 'second priority did not follow')

    def test_weighted_priority_never_holds_partial_or_allows_backfill(self):
        import fcntl
        self.grant(2)
        a, b = self.fill()
        self.start('normal', 1)
        self.wait_queued('normal')
        priority, _ = self.start('priority', 2, weight=2)
        self.wait_queued('priority')
        self.release(a)
        time.sleep(1.2)
        self.assertFalse(self.acquired('priority'))
        self.assertFalse(self.acquired('normal'), 'normal backfilled ahead of weighted priority')
        with (self.root / 'lock').open('r+') as free:
            fcntl.flock(free.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
        self.release(b)
        self.await_(lambda: self.acquired('priority'), 'weighted priority not admitted when capacity sufficient')
        self.assertFalse(self.acquired('normal'))

    def test_dead_priority_owner_does_not_hold_queue(self):
        self.grant(2)
        a, b = self.fill()
        normal, _ = self.start('normal', 1)
        self.wait_queued('normal')
        priority, helper = self.start('priority', 2)
        self.wait_queued('priority')
        self.release(priority)
        self.await_(lambda: helper.poll() is not None, 'dead watched owner left priority helper queued')
        self.release(a)
        self.await_(lambda: self.acquired('normal'), 'dead priority ticket blocked normal')

    def test_cancelled_helper_reaps_provider_process_group(self):
        pidfile = self.root / 'provider.pid'
        provider = self.root / 'slow-provider'
        provider.write_text('#!/usr/bin/env python3\nimport os,time\nopen(' + repr(str(pidfile)) + ', "w").write(str(os.getpid()))\ntime.sleep(120)\n')
        provider.chmod(0o755)
        self.env['WF_TEST_PRIORITY_PROVIDER'] = str(provider)
        owner, helper = self.start('cancel-provider', 2)
        self.await_(lambda: pidfile.exists() and pidfile.read_text().isdigit(), 'provider never started')
        provider_pid = int(pidfile.read_text())
        def alive():
            try:
                os.kill(provider_pid, 0)
                return True
            except ProcessLookupError:
                return False
        def cleanup_provider():
            try:
                os.killpg(provider_pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
        self.addCleanup(cleanup_provider)
        helper.terminate()
        helper.wait(timeout=3)
        time.sleep(.2)
        self.assertFalse(alive(), 'cancelled helper left provider process alive')

    def test_unavailable_provider_warns_and_never_stops_suite(self):
        self.grant(2)
        self.env['WF_TEST_PRIORITY_PROVIDER'] = str(self.root / 'missing')
        a, b = self.fill()
        normal, _ = self.start('normal', 1)
        self.wait_queued('normal')
        self.start('priority', 2)
        self.wait_queued('priority')
        self.assertIn('priority unavailable', self.contents('priority', '.err'))
        self.release(a)
        self.await_(lambda: self.acquired('normal'), 'unavailable provider prevented normal admission')
        self.release(normal)
        self.await_(lambda: self.acquired('priority'), 'unavailable provider prevented known-issue admission')


spec = importlib.util.spec_from_file_location('test_lock', HELPER)
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class GrantEvidenceUnit(unittest.TestCase):
    def records(self, n):
        return [(i, {'subject': {'common_dir': '/tmp/repo' + str(i % 2), 'issue': i + 1}}) for i in range(n)]

    def test_disappearing_ticket_during_admission(self):
        with tempfile.TemporaryDirectory(prefix='ticket-release-', dir=os.environ.get('TMPDIR', '/tmp')) as directory:
            base = str(Path(directory) / 'lock')
            rec = {'repo': 'fixture', 'worktree': directory, 'pid': os.getpid(), 'started': 'now'}
            older = module.Tickets(base, rec)
            current = module.Tickets(base, rec)
            enumerate_live = module.live_tickets
            released = []
            def release_after_scan(tdir):
                snapshot = enumerate_live(tdir)
                older.release()  # Legacy/current owner can leave after the snapshot.
                released.append(True)
                return snapshot
            slots = [os.open(base, os.O_RDWR | os.O_CREAT, 0o600)]
            queue = os.open(base + '.queue', os.O_RDWR | os.O_CREAT, 0o600)
            try:
                with mock.patch.object(module, 'live_tickets', side_effect=release_after_scan):
                    taken, busy, ahead, head, deadline = current.admission(slots, queue, 1, module.GrantEvidence('unused'))
                self.assertEqual(taken, [0])
                self.assertEqual(ahead, [])
            finally:
                current.release()
                if not released:
                    older.release()
                os.close(queue)
                for fd in slots:
                    os.close(fd)

    def test_malformed_persisted_subjects_never_reach_filesystem(self):
        subjects = [
            {'common_dir': '/tmp/review\x00bad', 'issue': 1},
            {'common_dir': '/tmp/\ud800', 'issue': 1},
            {'common_dir': 'relative', 'issue': 1},
            {'common_dir': '/tmp/' + 'x' * 4096, 'issue': 1},
            {'common_dir': '/tmp/../repo', 'issue': 1},
            {'common_dir': '/tmp/repo', 'issue': True},
            {'common_dir': '/tmp/repo', 'issue': -1},
            {'common_dir': 3, 'issue': 1},
            *[{'common_dir': '/tmp/repo', 'issue': 0, 'issues': v}
              for v in (None, '2', [True], [-1], [0], [1.5], [2, '3'])], {}, [], 'bad',
        ]
        for subject in subjects:
            paths, warnings = [], []
            def stat(path):
                paths.append(path)
                return os.stat(path)
            evidence = module.GrantEvidence('unused', stat=stat)
            record = {'subject': subject}
            with mock.patch.object(module, 'log', side_effect=warnings.append):
                self.assertEqual(evidence.deadlines([(1, record)], 100), {})
            self.assertEqual(paths, [], 'invalid subject reached filesystem')
            self.assertTrue(any('priority unavailable' in w for w in warnings), 'invalid evidence silently discarded')
        # Legacy tickets and a deliberate issue-zero identity are ordinary FIFO.
        evidence = module.GrantEvidence('unused', stat=lambda path: self.fail('normal subject stat'))
        with mock.patch.object(module, 'log') as warning:
            self.assertEqual(evidence.deadlines([(1, {}), (2, {'subject': None}),
                (3, {'subject': {'common_dir': '/tmp/repo', 'issue': 0}})], 100), {})
            warning.assert_not_called()

    def test_stable_partition_and_expiry(self):
        for n in range(1, 64):
            records = self.records(n)
            deadlines = {module.subject_keys(rec)[0]: 101 if i % 3 else 99 for i, rec in records}
            got = module.effective_order(records, deadlines, 100)
            expected = [x for x in records if deadlines[module.subject_keys(x[1])[0]] > 100]
            expected += [x for x in records if deadlines[module.subject_keys(x[1])[0]] <= 100]
            self.assertEqual(got, expected)
            self.assertEqual(module.effective_order(records, deadlines, 102), records)

    def test_train_projection_and_counted_growth(self):
        for n in (8, 16):
            records = [(i, {'subject': {'common_dir': '/tmp/repo', 'issue': 0,
                        'issues': [i * 2 + 1, i * 2 + 2]}}) for i in range(n)]
            calls, stats = [], []
            def query(provider, request):
                calls.append(request)
                return {'sources': [{'common_dir': '/tmp/repo',
                        'deadlines': {str(n * 2): 103}, 'refresh_at': 103}]}
            def stat(path):
                stats.append(path)
                return SimpleNamespace(st_ino=1, st_mtime_ns=1, st_size=1)
            evidence = module.GrantEvidence('provider', query=query, stat=stat)
            ends = evidence.deadlines(records, 100)
            self.assertEqual(len(calls), 1)
            self.assertEqual(len(stats), 1)
            self.assertEqual(len(calls[0]['subjects']), n * 2)
            self.assertEqual(module.effective_order(records, ends, 100)[0], records[-1])
            self.assertEqual(module.effective_order(records, ends, 103), records)
            evidence.deadlines(records, 101)
            self.assertEqual(len(calls), 1, 'unchanged train membership bypassed cache')
        record = {'subject': {'common_dir': '/tmp/repo', 'issue': 0, 'issues': [2, 3]}}
        other_repo = {'subject': {'common_dir': '/tmp/other', 'issue': 0, 'issues': [2, 3]}}
        ends = {('/tmp/repo', 2): 101, ('/tmp/repo', 3): 105}
        self.assertEqual(module.effective_order([(1, other_repo), (2, record)], ends, 102)[0][0], 2)
        self.assertEqual(module.effective_order([(1, other_repo), (2, record)], ends, 105)[0][0], 1)

    def test_source_and_clock_boundaries_with_counted_growth(self):
        for n in (20, 40):
            stats, calls = [], []
            generation = [1]
            def stat(path):
                stats.append(path)
                return SimpleNamespace(st_ino=1, st_mtime_ns=generation[0], st_size=20)
            def query(provider, request):
                calls.append(request)
                groups = {}
                for subject in request['subjects']:
                    groups.setdefault(subject['common_dir'], {})[str(subject['issue'])] = 103
                return {'sources': [{'common_dir': common, 'deadlines': ends, 'refresh_at': 102}
                                    for common, ends in groups.items()]}
            evidence = module.GrantEvidence('provider', query=query, stat=stat)
            records = self.records(n)
            self.assertEqual(len(evidence.deadlines(records, 100)), n)
            self.assertEqual(len(stats), 2)
            self.assertEqual(len(calls), 1)
            self.assertEqual(len(calls[0]['subjects']), n)
            evidence.deadlines(records, 101)
            self.assertEqual(len(calls), 1, 'stable poll invoked a provider')
            evidence.deadlines(records, 102)
            self.assertEqual(len(calls), 2, 'clock-only activation boundary was ignored')
            generation[0] += 1
            evidence.deadlines(records, 102.1)
            self.assertEqual(len(calls), 3, 'source replacement was ignored')

    def test_invalid_and_failed_provider_evidence_is_not_cached(self):
        records = self.records(2)
        for response in ({}, {'sources': [{}]}, {'sources': [True, False]},
                         {'sources': [{'common_dir': '/tmp/repo0', 'deadlines': {'1': float('inf')}, 'refresh_at': 0}]}):
            calls = []
            def query(*args):
                calls.append(1)
                return response
            evidence = module.GrantEvidence('provider', query=query,
                stat=lambda path: SimpleNamespace(st_ino=1, st_mtime_ns=1, st_size=1))
            self.assertEqual(evidence.deadlines(records, 100), {})
            self.assertEqual(evidence.deadlines(records, 100), {})
            self.assertEqual(len(calls), 2, 'unavailable evidence cached indefinitely')

    def test_protocol_boundaries_timeout_and_output_limit(self):
        with tempfile.TemporaryDirectory(prefix='priority-protocol-', dir=os.environ.get('TMPDIR', '/tmp')) as directory:
            provider = Path(directory) / 'provider'
            original = module.PROVIDER_TIMEOUT
            module.PROVIDER_TIMEOUT = .2
            self.addCleanup(setattr, module, 'PROVIDER_TIMEOUT', original)
            for body in ('print("{}")', 'print("x" * 70000)', 'import time; time.sleep(10)', 'print("{bad")', '''print('{"version":true}')'''):
                provider.write_text('#!/usr/bin/env python3\n' + body + '\n')
                provider.chmod(0o755)
                started = time.monotonic()
                with self.assertRaises((ValueError, subprocess.SubprocessError)):
                    module.query_priority(str(provider), {'version': 1, 'subjects': []})
                self.assertLess(time.monotonic() - started, 2)


if __name__ == '__main__':
    unittest.main()
