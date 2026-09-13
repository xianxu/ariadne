#!/usr/bin/env bash
# Execute the workflow's actual Bash blocks against scratch Git repositories.
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../.." && pwd)"
python3 - "$SOURCE" <<'PY'
import os,pathlib,re,subprocess,sys,tempfile
source=pathlib.Path(sys.argv[1]); workflow=(source/'.github/workflows/merge-check.yml').read_text()
# Only run: | steps are evaluated; action wiring is checked explicitly below.
blocks=[]
for match in re.finditer(r'^        run: \|\n((?:          .*\n|\n)+)',workflow,re.M):
    blocks.append(''.join(line[10:] if line.startswith('          ') else line for line in match[1].splitlines(True)))
assert blocks, 'no workflow execution surface'
with tempfile.TemporaryDirectory(prefix='portable-ci.') as tmp:
    root=pathlib.Path(tmp); leaf=root/'leaf'; leaf.mkdir(); peer=root/'ariadne'; (peer/'scripts').mkdir(parents=True)
    (peer/'go.mod').write_text('module example.com/base\ngo 1.26.3\n')
    subprocess.run(['git','init','-q',str(leaf)],check=True)
    subprocess.run(['git','-C',str(leaf),'-c','user.name=Test','-c','user.email=test@test','commit','--allow-empty','-qm','test'],check=True)
    sha=subprocess.check_output(['git','-C',str(leaf),'rev-parse','HEAD'],text=True).strip()
    (leaf/'scripts/merge-checks.d').mkdir(parents=True)
    def executable(file,body):
        file.write_text('#!/bin/sh\nset -eu\n'+body); file.chmod(0o755)
    # Real generic runner, not a canned success; its check observes hook state.
    (peer/'scripts/run-merge-checks.sh').write_bytes((source/'scripts/run-merge-checks.sh').read_bytes())
    hook=leaf/'scripts/ci-setup.sh'; check=leaf/'scripts/merge-checks.d/01-product.sh'
    executable(hook,'echo setup >> events\ntouch ready\n')
    executable(check,'test -f ready\necho checked >> events\n')
    env=dict(os.environ,HOME=str(root/'home'),GITHUB_OUTPUT=str(root/'output'))
    def run(expected=0):
        result=None
        for block in blocks:
            block=block.replace('${{ github.event.pull_request.base.sha }}',sha).replace('${{ github.event.pull_request.head.sha }}',sha)
            result=subprocess.run(['bash','-c',block],cwd=leaf,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
            if result.returncode: break
        assert (result.returncode==0)==(expected==0),result.stdout
    run(); assert (leaf/'events').read_text().splitlines()==['setup','checked']
    assert 'path=../ariadne/go.mod' in (root/'output').read_text()
    # Consumer runner takes precedence over fallback; setup still comes first.
    executable(leaf/'scripts/run-merge-checks.sh','test -f ready\necho local-runner >> events\n')
    (leaf/'go.mod').write_text('module example.com/leaf\ngo 1.26.3\n')
    run(); assert (leaf/'events').read_text().splitlines()[-2:]==['setup','local-runner']
    assert 'path=go.mod' in (root/'output').read_text()
    executable(hook,'exit 9\n'); before=(leaf/'events').read_text(); run(expected=9); assert (leaf/'events').read_text()==before
    hook.unlink(); run(); assert (leaf/'events').read_text().splitlines()[-1]=='local-runner'
    hook.write_text('exit 9\n'); hook.chmod(0o644); run() # only executable hook is opted in
    (leaf/'scripts/run-merge-checks.sh').unlink(); (peer/'scripts/run-merge-checks.sh').unlink(); run(expected=1)
assert 'go-version-file: ${{ steps.go-module.outputs.path }}' in workflow
assert workflow.index('Clone base-layer peers') < workflow.index('uses: actions/setup-go')
print('PASS portable CI actual blocks: source selection, hook ordering and failures')
PY
