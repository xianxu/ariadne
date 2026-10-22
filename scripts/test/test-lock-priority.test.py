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


    def test_acquire_watcher_retains_immutable_enqueue_and_unknown_metadata(self):
        import fcntl
        a, b = self.fill()
        owner, _ = self.start('queued-metadata', 1)
        self.wait_queued('queued-metadata')
        allocation = os.open(str(self.root / 'lock') + '.tickets/alloc', os.O_RDWR)
        try:
            fcntl.flock(allocation, fcntl.LOCK_EX)
            for path in (self.root / 'lock.tickets').iterdir():
                if not path.name.isdigit():
                    continue
                queued = json.loads(path.read_text())
                if queued['pid'] == owner.pid:
                    queued['future_metadata'] = {'nested': ['keep-through-watcher-rewrite']}
                    path.write_text(json.dumps(queued))
                    break
            else:
                self.fail('missing private metadata ticket')
        finally:
            os.close(allocation)
        self.release(a)
        self.await_(lambda: self.acquired('queued-metadata'), 'metadata waiter did not acquire')
        held = [json.loads((self.root / name).read_text()) for name in ('lock', 'lock.1')]
        actual = next(record for record in held if record['pid'] == owner.pid)
        self.assertIn('enqueue_monotonic', actual, 'final watcher rewrite erased ticket age')
        self.assertEqual(actual['enqueue_monotonic'], queued['enqueue_monotonic'])
        self.assertEqual(actual['future_metadata'], queued['future_metadata'])
        self.assertEqual(actual['subject'], queued['subject'])
        self.assertIn('holder', actual, 'slot assertion must inspect the final watcher-owned record')

    def test_aged_weighted_waiter_survives_multiple_renewed_train_grants(self):
        import fcntl
        a, b = self.fill()
        old, _ = self.start('aged-heavy', 1, weight=2)
        self.wait_queued('aged-heavy')
        self.start('young-grant', 2)
        self.wait_queued('young-grant')
        self.start('young-train', issues=[3, 4])
        self.wait_queued('young-train')
        # Private fixture clocks only: emulate an hour's elapsed enqueue age.
        allocation = os.open(str(self.root / 'lock') + '.tickets/alloc', os.O_RDWR)
        try:
            fcntl.flock(allocation, fcntl.LOCK_EX)
            for path in (self.root / 'lock.tickets').iterdir():
                if not path.name.isdigit():
                    continue
                record = json.loads(path.read_text())
                if record['pid'] == old.pid:
                    record['enqueue_monotonic'] = time.monotonic() - 3601
                    path.write_text(json.dumps(record))
                    break
            else:
                self.fail('missing private aged ticket')
        finally:
            os.close(allocation)
        for _ in range(2):
            self.grant(2)
            self.grant(4)
        observed = subprocess.run(['python3', str(HELPER), 'inspect'], env=self.env,
                                  capture_output=True, text=True, check=True)
        projection = json.loads(observed.stdout)
        self.assertTrue(projection['available'], projection)
        self.assertEqual(projection['tickets'][0]['record']['pid'], old.pid)
        self.assertTrue(projection['priority_overridden'])
        self.release(a)
        self.await_(lambda: 'FIFO protected' in self.contents('aged-heavy', '.err'), 'aged reason not rendered')
        self.assertFalse(self.acquired('aged-heavy'))
        self.assertFalse(self.acquired('young-grant'), 'younger grant backfilled one free slot')
        self.assertFalse(self.acquired('young-train'), 'younger train backfilled one free slot')
        self.assertIsNone(b.poll(), 'running suite was preempted')
        self.release(b)
        self.await_(lambda: self.acquired('aged-heavy'), 'protected heavy head did not admit')
        self.assertFalse(self.acquired('young-grant'))
        self.assertFalse(self.acquired('young-train'))
        self.release(old)
        self.await_(lambda: self.acquired('young-grant') and self.acquired('young-train'), 'priority suffix did not progress')

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
        return [(i + 1, {'enqueue_monotonic': time.monotonic(), 'subject': {'common_dir': '/tmp/repo' + str(i % 2), 'issue': i + 1}}) for i in range(n)]

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
            records = [(i + 1, {'enqueue_monotonic': time.monotonic(), 'subject': {'common_dir': '/tmp/repo', 'issue': 0,
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
        record = {'enqueue_monotonic': time.monotonic(), 'subject': {'common_dir': '/tmp/repo', 'issue': 0, 'issues': [2, 3]}}
        other_repo = {'enqueue_monotonic': time.monotonic(), 'subject': {'common_dir': '/tmp/other', 'issue': 0, 'issues': [2, 3]}}
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



class SuiteAgingPolicy(unittest.TestCase):
    def record(self, issue, enqueued):
        return dict(repo='fixture', worktree='/fixture', pid=os.getpid(), started='fixture',
                    enqueue_monotonic=enqueued, subject={'common_dir': '/fixture.git', 'issue': issue})

    def test_old_waiter_stops_continuous_grant_overtaking(self):
        records = [(1, self.record(1, 0)), (2, self.record(2, 3599))]
        deadlines = {('/fixture.git', 2): 20000}
        with mock.patch.object(module.time, 'monotonic', return_value=3600):
            self.assertEqual(module.effective_order(records, deadlines, 10000), records,
                             '60-minute ordinary waiter must precede younger granted suite')

    def test_threshold_renewals_and_independent_clocks(self):
        records = [(1, self.record(1, 100)), (2, self.record(2, 3699))]
        for wall, end in ((1000, 2000), (-5000, 20000), (1000000, 2000000)):
            for mono, numbers in ((3699.99, [2, 1]), (3700, [1, 2]), (5000, [1, 2])):
                got = module.suite_order(records, {('/fixture.git', 2): end}, wall, mono)
                self.assertEqual([t['number'] for t in got['tickets']], numbers, (wall, end, mono, got))
                self.assertEqual(got['protected_through'], 1 if mono >= 3700 else 0)
                self.assertEqual(got['priority_overridden'], mono >= 3700)
        self.assertEqual(records[0][1]['enqueue_monotonic'], 100)

    def test_unknown_age_protects_entire_fifo_prefix(self):
        invalid = [None, True, False, 'old', float('nan'), float('inf'), -1, 5001]
        for value in invalid:
            records = [(1, self.record(1, 4900)), (2, self.record(2, value)), (3, self.record(3, 4900))]
            got = module.suite_order(records, {('/fixture.git', 3): 200}, 100, 5000)
            self.assertEqual([r['number'] for r in got['tickets']], [1, 2, 3])
            self.assertEqual(got['protected_through'], 2)
            self.assertIsNone(got['tickets'][1]['age_seconds'])
            self.assertEqual(got['tickets'][0]['reason'], 'protected_prefix')
            self.assertEqual(got['tickets'][1]['reason'], 'unknown_age_fifo')
        self.assertIsNone(module.ticket_age({}, 5000))

    def test_multiple_grants_and_counted_linear_work(self):
        for count in (32, 64):
            records = [(i, self.record(i, 0 if i <= 3 else 4999)) for i in range(1, count + 1)]
            ends = {('/fixture.git', i): 1000 for i in range(2, count + 1, 2)}
            with mock.patch.object(module, 'ticket_age', wraps=module.ticket_age) as ages, \
                 mock.patch.object(module, 'subject_deadline', wraps=module.subject_deadline) as subjects:
                got = module.suite_order(records, ends, 100, 5000)
            self.assertEqual(ages.call_count, count)
            self.assertEqual(subjects.call_count, count)
            self.assertEqual([r['number'] for r in got['tickets']],
                             [1, 2, 3] + list(range(4, count + 1, 2)) + list(range(5, count + 1, 2)))
            self.assertEqual(got['tickets'][1]['position'], 2)
            self.assertTrue(got['tickets'][1]['priority_overridden'])

    def test_wait_diagnostic_names_limit_even_before_protection(self):
        records = [(1, self.record(1, 4999))]
        row = module.suite_order(records, {}, 100, 5000)['tickets'][0]
        text = module.format_suite_wait(row)
        self.assertIn('wait age 1s', text)
        self.assertIn('grant overtaking limit 3600s', text)

    def test_all_granted_fifo_and_regrant_cannot_reset_age(self):
        records = [(1, self.record(1, 1)), (2, self.record(2, 4999)), (3, self.record(3, 4999))]
        for ends in ({}, {('/fixture.git', 2): 999}, {('/fixture.git', 3): 999},
                     {('/fixture.git', 1): 999, ('/fixture.git', 2): 999, ('/fixture.git', 3): 999}):
            got = module.suite_order(records, ends, 100, 5000)
            self.assertEqual(got['tickets'][0]['number'], 1)
            self.assertEqual(got['tickets'][0]['age_seconds'], 4999)



class SuiteAgingResources(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(prefix='suite-aging-', dir=os.environ.get('TMPDIR', '/tmp'))
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.base = str(self.root / 'lock')
        self.records = []
        self.descriptors = []
        self.addCleanup(self.cleanup)

    def cleanup(self):
        for ticket in self.records:
            ticket.release()
        for fd in self.descriptors:
            os.close(fd)

    def ticket(self, issue, clock=5000):
        rec = dict(repo='fixture', worktree='/fixture', pid=os.getpid(), started='fixture',
                   enqueue_monotonic=-123, future={'nested': ['retained']},
                   subject={'common_dir': '/fixture.git', 'issue': issue})
        with mock.patch.object(module.time, 'monotonic', return_value=clock):
            ticket = module.Tickets(self.base, rec)
        self.records.append(ticket)
        return ticket

    def fd(self, suffix):
        fd = os.open(self.base + suffix, os.O_RDWR | os.O_CREAT, 0o600)
        self.descriptors.append(fd)
        return fd

    def evidence(self, ends=None):
        return SimpleNamespace(deadlines=lambda *_: ends or {}, warned=set())

    def test_enqueue_stamped_under_alloc_and_preserved_into_slots(self):
        calls = []
        def clock():
            import fcntl
            probe = os.open(self.base + '.tickets/alloc', os.O_RDWR)
            try:
                with self.assertRaises(BlockingIOError):
                    fcntl.flock(probe, fcntl.LOCK_EX | fcntl.LOCK_NB)
            finally:
                os.close(probe)
            calls.append(True)
            return 1000
        with mock.patch.object(module.time, 'monotonic', side_effect=clock):
            ticket = module.Tickets(self.base, dict(repo='fixture',worktree='/fixture',pid=os.getpid(),started='fixture',
                                                  enqueue_monotonic=-1,future={'nested':[1]}))
        self.records.append(ticket)
        self.assertEqual(len(calls), 1)
        rec = json.loads(Path(ticket.path).read_text())
        self.assertEqual(rec['enqueue_monotonic'], 1000)
        self.assertEqual(rec['future'], {'nested': [1]})
        slot, token = self.fd(''), self.fd('.queue')
        with mock.patch.object(module.time, 'monotonic', return_value=5000):
            taken, _, ahead, _, _ = ticket.admission([slot], token, 1, self.evidence())
        self.assertEqual(taken, [0])
        self.assertEqual(ahead, [])
        held = module.holder(slot)
        self.assertEqual(held['enqueue_monotonic'], 1000)
        self.assertEqual(held['future'], {'nested': [1]})
        self.assertIn('FIFO protected', module.format_suite_wait(ticket.observation))

    def test_aged_heavy_head_blocks_young_grant_without_partial_slots(self):
        import fcntl
        old, young = self.ticket(1, 100), self.ticket(2, 4999)
        slots, token = [self.fd(''), self.fd('.1')], self.fd('.queue')
        occupied = os.open(self.base + '.1', os.O_RDWR)
        self.descriptors.append(occupied)
        fcntl.flock(occupied, fcntl.LOCK_EX)
        evidence = self.evidence({('/fixture.git', 2): time.time() + 300})
        with mock.patch.object(module.time, 'monotonic', return_value=5000):
            self.assertEqual(old.admission(slots, token, 2, evidence)[0], [])
            self.assertEqual(young.admission(slots, token, 1, evidence)[0], [])
            probe = os.open(self.base, os.O_RDWR)
            try:
                fcntl.flock(probe, fcntl.LOCK_EX | fcntl.LOCK_NB)
            finally:
                os.close(probe)
            fcntl.flock(occupied, fcntl.LOCK_UN)
            self.assertEqual(old.admission(slots, token, 2, evidence)[0], [0, 1])
        self.assertTrue(young.observation['priority_overridden'])
        self.assertEqual(old.observation['reason'], 'aged_fifo')

    def test_inspection_no_creation_or_dead_ticket_pruning(self):
        missing = str(self.root / 'absent')
        self.assertTrue(module.inspect_queue(missing, self.evidence())['available'])
        self.assertFalse(Path(missing + '.tickets').exists())
        ticket = self.ticket(1, 100)
        dead = Path(self.base + '.tickets/000000000002')
        dead.write_text('dead legacy data')
        before = {p.name: (p.stat().st_ino, p.read_bytes()) for p in Path(ticket.dir).iterdir()}
        got = module.inspect_queue(self.base, self.evidence(), now=10000, monotonic_now=5000)
        self.assertTrue(got['available'])
        self.assertEqual(len(got['tickets']), 1)
        self.assertEqual(got['tickets'][0]['age_seconds'], 4900)
        self.assertEqual(got['tickets'][0]['record']['future'], {'nested':['retained']})
        self.assertEqual(before, {p.name: (p.stat().st_ino, p.read_bytes()) for p in Path(ticket.dir).iterdir()})

    def test_inspection_revalidates_snapshot_and_provider_outside_alloc(self):
        import fcntl
        ticket = self.ticket(1)
        def changed(*_):
            probe = os.open(Path(ticket.dir) / 'alloc', os.O_RDWR)
            try:
                fcntl.flock(probe, fcntl.LOCK_EX | fcntl.LOCK_NB)
            finally:
                os.close(probe)
            os.pwrite(ticket.fd, b' ', os.fstat(ticket.fd).st_size)
            return {}
        got = module.inspect_queue(self.base, SimpleNamespace(deadlines=changed, warned=set()))
        self.assertFalse(got['available'])
        self.assertEqual(got['reason'], 'snapshot_changed')

    def test_inspection_busy_bounds_and_provider_failure(self):
        import fcntl
        ticket = self.ticket(1)
        allocation = os.open(Path(ticket.dir) / 'alloc', os.O_RDWR)
        try:
            fcntl.flock(allocation, fcntl.LOCK_EX)
            self.assertEqual(module.inspect_queue(self.base, self.evidence())['reason'], 'allocation_busy')
        finally:
            os.close(allocation)
        failed = SimpleNamespace(deadlines=lambda *_: {}, warned={'provider failed'})
        self.assertEqual(module.inspect_queue(self.base, failed)['reason'], 'provider_unavailable')
        with mock.patch.object(module, 'INSPECT_MAX_RECORDS', 0):
            self.assertEqual(module.inspect_queue(self.base, self.evidence())['reason'], 'record_limit')
        with mock.patch.object(module, 'INSPECT_MAX_RECORD_BYTES', 8):
            self.assertEqual(module.inspect_queue(self.base, self.evidence())['reason'], 'record_limit')
        with mock.patch.object(module, 'INSPECT_MAX_OUTPUT_BYTES', 8):
            self.assertEqual(module.inspect_queue(self.base, self.evidence())['reason'], 'output_limit')

    def test_inspection_hostile_files_are_nonblocking_and_read_only(self):
        self.ticket(1)
        outside = self.root / 'outside'
        outside.write_text('untouched')
        path = Path(self.base + '.tickets/000000000099')
        path.symlink_to(outside)
        self.assertFalse(module.inspect_queue(self.base, self.evidence())['available'])
        self.assertEqual(outside.read_text(), 'untouched')
        path.unlink()
        os.mkfifo(path)
        self.assertEqual(module.inspect_queue(self.base, self.evidence())['reason'], 'record_invalid')
        path.unlink()
        alloc = Path(self.base + '.tickets/alloc')
        alloc.unlink()
        os.mkfifo(alloc)
        self.assertEqual(module.inspect_queue(self.base, self.evidence())['reason'], 'allocation_unavailable')

    def test_inspection_inode_replacement_invalidates_same_bytes(self):
        import fcntl
        ticket = self.ticket(1)
        replacements = []
        def replace(*_):
            raw = Path(ticket.path).read_bytes()
            Path(ticket.path).unlink()
            fd = os.open(ticket.path, os.O_RDWR | os.O_CREAT, 0o600)
            replacements.append(fd)
            fcntl.flock(fd, fcntl.LOCK_EX)
            os.write(fd, raw)
            return {}
        try:
            got = module.inspect_queue(self.base, SimpleNamespace(deadlines=replace, warned=set()))
            self.assertEqual(got['reason'], 'snapshot_changed')
        finally:
            for fd in replacements:
                os.close(fd)

    def test_nonfinite_persisted_json_is_explicitly_unavailable(self):
        ticket = self.ticket(1)
        for raw_age in ('NaN', 'Infinity', '-Infinity', '1e999'):
            raw = Path(ticket.path).read_bytes()
            record = json.loads(raw)
            record['enqueue_monotonic'] = 'replace-age'
            replacement = json.dumps(record).replace('"replace-age"', raw_age).encode()
            os.ftruncate(ticket.fd, 0)
            os.pwrite(ticket.fd, replacement, 0)
            got = module.inspect_queue(self.base, self.evidence())
            self.assertFalse(got['available'])
            self.assertEqual(got['reason'], 'record_invalid')
            json.dumps(got, allow_nan=False)
            self.assertEqual(Path(ticket.path).read_bytes(), replacement)

    def test_inspection_cli_and_acquire_watch_contract(self):
        env = dict(os.environ, WF_TEST_LOCK_FILE=self.base)
        result = subprocess.run(['python3', str(HELPER), 'inspect'], env=env, capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(json.loads(result.stdout)['tickets'], [])
        missing = subprocess.run(['python3', str(HELPER), 'acquire'], capture_output=True, text=True)
        self.assertNotEqual(missing.returncode, 0)
        self.assertIn('--watch', missing.stderr)


if __name__ == '__main__':
    unittest.main()
