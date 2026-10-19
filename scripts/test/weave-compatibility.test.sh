#!/usr/bin/env bash
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
python3 - "$repo_root" <<'PY'
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import time
import signal
import sys
import tempfile
sys.dont_write_bytecode = True
import unittest

root = Path(sys.argv[1])
helper = root / 'scripts/ci/weave-compatibility.py'

class CompatibilityTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        assert helper.is_file(), 'missing compatibility guard implementation'
        spec = importlib.util.spec_from_file_location('compatibility', helper)
        cls.m = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(cls.m)

    def test_ref_grammar_and_membership(self):
        sha = 'a' * 40
        self.assertEqual(self.m.parse_selected_ref((sha+'\n').encode()), sha)
        for raw in (b'', b'a'*40, b'A'*40+b'\n', b'a'*39+b'\n', b'a'*40+b'\r\n', b'a'*40+b'\n\n', b'x'*5000):
            with self.subTest(raw=raw[:50]), self.assertRaises(ValueError):
                self.m.parse_selected_ref(raw)
        self.assertEqual(self.m.parse_supported_refs((sha+'\n'+'b'*40+'\n').encode()), (sha, 'b'*40))
        for raw in ((sha+'\n')*2, (sha+'\n')*3, '', 'x'*5000):
            with self.assertRaises(ValueError):
                self.m.parse_supported_refs(raw.encode())
        self.m.require_supported(sha, (sha,))
        with self.assertRaises(ValueError):
            self.m.require_supported('b'*40, (sha,))

    def test_actual_declaration_bytes_are_preserved(self):
        row = b'export\t skill  construct/local   # candidate producer\n'
        self.assertEqual(self.m.select_vocabulary_declarations(b'# unrelated\nseed bootstrap.sh\n'+row), row)
        for manifest in (b'skill construct/adapted\n', b'# skill construct/local\n', b'skill construct/renamed\n'):
            with self.assertRaises(ValueError):
                self.m.select_vocabulary_declarations(manifest)
        # Visibility/verbs remain the real compiler's responsibility.
        changed = b'internal skill construct/local\n'
        self.assertEqual(self.m.select_vocabulary_declarations(changed), changed)

    def test_exact_corpus_and_independent_digests(self):
        expected = {'issue.json': b'{"changed":true}\n', 'new.json': b'{}\n'}
        side = {p: (hashlib.sha256(b).hexdigest()+'\n').encode() for p,b in expected.items()}
        self.assertEqual(self.m.compare_corpus(expected, dict(expected), side), ())
        for actual, digests in (({'issue.json': b'old'}, side), (expected, {}), ({**expected, 'extra.json': b'{}'}, side), (expected, {p:b'0'*64+b'\n' for p in expected})):
            with self.subTest(actual=actual):
                self.assertTrue(self.m.compare_corpus(expected, actual, digests))
        class Counted(dict):
            def __init__(self, *a):
                super().__init__(*a)
                self.visits = 0
            def items(self):
                for p,b in super().items():
                    self.visits += 1
                    yield p,b
        visits = []
        for n in (8,16):
            corpus = Counted({str(i): b'{}' for i in range(n)})
            digests = {p: (hashlib.sha256(b).hexdigest()+'\n').encode() for p,b in dict(corpus).items()}
            self.assertEqual(self.m.compare_corpus(corpus, corpus, digests), ())
            visits.append(corpus.visits)
        self.assertEqual(visits[1], visits[0]*2)

    def test_process_exit_and_timeout_are_not_success(self):
        with tempfile.TemporaryDirectory(prefix='weave-compat-process-', dir=os.environ.get('TMPDIR', '/tmp')) as d:
            d = Path(d)
            binary = d/'compiler'
            binary.write_text('#!/bin/sh\nprintf "actual executable\\n"\nexit 7\n')
            binary.chmod(0o755)
            result = self.m.run_compiler(binary, d, 5)
            self.assertEqual(result.returncode, 7)
            self.assertIn('actual executable', result.stdout)
            binary.write_text('#!/bin/sh\nsleep 20 &\nwait\n')
            with self.assertRaises(TimeoutError):
                self.m.run_compiler(binary, d, 0.05)

    def test_stateful_cold_protocol_and_failures(self):
        with tempfile.TemporaryDirectory(prefix='weave-compat-fixture-', dir=os.environ.get('TMPDIR', '/tmp')) as d:
            d = Path(d)
            candidate, tools = d/'candidate', d/'tools'
            (candidate/'construct/local/vocabulary').mkdir(parents=True)
            (candidate/'construct/vocabulary/testdata').mkdir(parents=True)
            (candidate/'construct/base.manifest').write_text('skill construct/local\n')
            (candidate/'.sdlc').mkdir()
            (candidate/'.sdlc/weave-compiler-ref').write_text('a'*40+'\n')
            (candidate/'.sdlc/weave-compiler-refs').write_text('a'*40+'\n'+'b'*40+'\n')
            (candidate/'construct/vocabulary/issue.cue').write_text('package vocabulary\n')
            (candidate/'construct/vocabulary/testdata/invalid.cue').write_text('deliberately invalid')
            marker = candidate/'construct/local/vocabulary/.dynamic-skill'
            marker.write_text('#!/bin/sh\n# weave-output: argv1\nvocabulary export --output "$1"\n')
            marker.chmod(0o755)
            tools.mkdir()
            vocabulary = tools/'vocabulary'
            vocabulary.write_text('''#!/usr/bin/env python3
from pathlib import Path
import sys
p=Path(sys.argv[sys.argv.index('--output')+1]);p.mkdir(parents=True,exist_ok=True)
assert not (Path.cwd()/'construct/vocabulary/testdata').exists()
(p/'issue.json').write_text('{"candidate":true}\\n')
(p/'SKILL.md').write_text('---\\nname: vocabulary\\ndescription: actual fixture generator\\n---\\n')
''')
            vocabulary.chmod(0o755)
            compiler = tools/'private-compiler'
            compiler.write_text('''#!/usr/bin/env python3
import hashlib, os
from pathlib import Path
import subprocess, tempfile
leaf=Path.cwd()
assert os.environ.get('WEAVE_FAKE_ASSERT') == 'identity'
with (leaf.parent/'compiler-calls').open('a') as f:f.write(str(Path(__file__).resolve())+' '+leaf.name+'\\n')
owner=(leaf/'construct/deps').read_text().split()[1]
with tempfile.TemporaryDirectory(dir=leaf.parent) as d:
    out=Path(d)
    subprocess.run([str((leaf/owner/'construct/local/vocabulary/.dynamic-skill').resolve()),str(out)],check=True)
    manifest=(leaf/'construct/base.manifest').read_text()
    for src in out.iterdir():
        dst=leaf/'construct/generated/vocabulary'/src.name
        side=Path(str(dst)+'.weave-sha256')
        if dst.exists() and dst.read_bytes()!=src.read_bytes():
            valid=not side.is_symlink() and side.is_file() and side.read_bytes()==(hashlib.sha256(dst.read_bytes()).hexdigest()+'\\n').encode()
            if not valid or ('ship '+str(dst.relative_to(leaf))) not in manifest:
                print('occupied by something weave did not write');sys.exit(1)
    for src in out.iterdir():
        dst=leaf/'construct/generated/vocabulary'/src.name
        dst.parent.mkdir(parents=True,exist_ok=True);dst.write_bytes(src.read_bytes())
        Path(str(dst)+'.weave-sha256').write_text(hashlib.sha256(src.read_bytes()).hexdigest()+'\\n')
'''.replace('import hashlib, os', 'import hashlib, os, sys'))
            compiler.chmod(0o755)
            saved = dict(os.environ)
            try:
                os.environ['PATH'] = str(tools)+os.pathsep+saved['PATH']
                os.environ['WEAVE_FAKE_ASSERT'] = 'identity'
                scratch = d/'scratch'
                scratch.mkdir()
                self.m.check_pin(compiler, candidate, scratch)
                self.assertEqual((scratch/'changed-cold/.sdlc/weave-compiler-ref').read_text(), 'b'*40+'\n')
                calls = (scratch/'compiler-calls').read_text().splitlines()
                self.assertGreaterEqual(len(calls), 6)
                self.assertTrue(all(line.startswith(str(compiler.resolve())+' ') for line in calls))
                self.assertFalse(any(p.name=='testdata' for p in (scratch/'owner/construct/vocabulary').iterdir()))
                # A compiler that exits zero without generating cannot pass.
                broken = tools/'broken'
                broken.write_text('#!/bin/sh\nexit 0\n');broken.chmod(0o755)
                second=d/'broken-scratch';second.mkdir()
                with self.assertRaises(RuntimeError):
                    self.m.check_pin(broken, candidate, second)
                # A candidate export failure must stop before any external compiler
                # invocation, so it cannot be mistaken for an old-compiler refusal.
                generator_source = vocabulary.read_text()
                vocabulary.write_text('#!/bin/sh\nprintf "candidate export rejected\\n" >&2\nexit 9\n')
                failed_export = d/'failed-export-scratch'
                failed_export.mkdir()
                try:
                    with self.assertRaisesRegex(RuntimeError, 'candidate export rejected'):
                        self.m.check_pin(compiler, candidate, failed_export)
                    self.assertFalse((failed_export/'compiler-calls').exists())
                finally:
                    vocabulary.write_text(generator_source)
                # SIGTERM must join the active compiler before CLI scratch cleanup.
                sleeper = tools/'sleeper'
                sleeper.write_text('#!/usr/bin/env python3\nimport os, time\nfrom pathlib import Path\nPath(os.environ["FAKE_PID_FILE"]).write_text(str(os.getpid()))\ntime.sleep(20)\n')
                sleeper.chmod(0o755)
                pid_file = d/'compiler.pid'
                os.environ['FAKE_PID_FILE'] = str(pid_file)
                process = subprocess.Popen([sys.executable, str(helper), 'check',
                    '--compiler', str(sleeper), '--candidate', str(candidate)],
                    stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
                compiler_pid = None
                try:
                    deadline = time.monotonic()+5
                    while not pid_file.exists() and time.monotonic() < deadline:
                        time.sleep(0.01)
                    self.assertTrue(pid_file.exists(), 'compiler did not start')
                    compiler_pid = int(pid_file.read_text())
                    process.terminate()
                    output, _ = process.communicate(timeout=5)
                    self.assertNotEqual(process.returncode, 0, output)
                    with self.assertRaises(ProcessLookupError):
                        os.kill(compiler_pid, 0)
                    compiler_pid = None
                finally:
                    if process.poll() is None:
                        process.kill();process.communicate()
                    if compiler_pid is not None:
                        try:os.kill(compiler_pid, signal.SIGKILL)
                        except ProcessLookupError:pass
            finally:
                os.environ.clear();os.environ.update(saved)

unittest.main(argv=['weave-compatibility'], verbosity=2)
PY
