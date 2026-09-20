#!/usr/bin/env bash
# Execute the workflow's actual Bash run blocks, including bootstrap.
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../.." && pwd)"
python3 - "$SOURCE" <<'PY'
import os,pathlib,re,subprocess,sys,tempfile
source=pathlib.Path(sys.argv[1]); workflow=(source/'.github/workflows/merge-check.yml').read_text()
blocks=[''.join(line[10:] if line.startswith('          ') else line for line in m[1].splitlines(True)) for m in re.finditer(r'^        run: \|\n((?:          .*\n|\n)+)',workflow,re.M)]
with tempfile.TemporaryDirectory(prefix='portable ci ') as tmp:
    root=pathlib.Path(tmp).resolve(); leaf=root/'leaf'; leaf.mkdir(); commands=root/'commands'; commands.mkdir(); runner_temp=root/'runner temp'; runner_temp.mkdir()
    subprocess.run(['git','init','-q',str(leaf)],check=True)
    subprocess.run(['git','-C',str(leaf),'-c','user.name=Test','-c','user.email=test@test','commit','--allow-empty','-qm','test'],check=True)
    sha=subprocess.check_output(['git','-C',str(leaf),'rev-parse','HEAD'],text=True).strip()
    def executable(path,body):
        path.write_text('#!/bin/bash\nset -eu\n'+body); path.chmod(0o755)
    (leaf/'scripts/merge-checks.d').mkdir(parents=True)
    (leaf/'bootstrap.sh').write_bytes((source/'bootstrap.sh').read_bytes()); (leaf/'bootstrap.sh').chmod(0o755)
    (leaf/'Brewfile').write_text('brew "go"\nbrew "cue"\n')
    gateway='if [ "$*" = "dependencies --help" ]; then exit 0; fi\necho compile >> events\n[ "${FAIL_COMPILE:-0}" = 0 ] || exit 7\ncp "$SOURCE/scripts/run-merge-checks.sh" scripts/run-merge-checks.sh\n'
    executable(root/'candidate',gateway)
    executable(commands/'go','echo candidate-build >> events\n[ "$1" = build ] && [ "$2" = -o ]\ncp "$FIXTURE/candidate" "$3"\n')
    executable(commands/'brew','echo "brew:$*" >> events\nif [ "$1" = install ]; then mkdir -p "$FIXTURE/formula/bin"; cp "$FIXTURE/candidate" "$FIXTURE/formula/bin/weave"; fi\nif [ "$1" = --prefix ]; then echo "$FIXTURE/formula"; fi\n')
    hook=leaf/'scripts/ci-setup.sh'; executable(hook,'test -f scripts/run-merge-checks.sh\necho setup >> events\n')
    executable(leaf/'scripts/merge-checks.d/01-product.sh','echo checked >> events\n')
    env=dict(os.environ,PATH=f'{commands}:/usr/bin:/bin',SOURCE=str(source),FIXTURE=str(root),RUNNER_TEMP=str(runner_temp),GITHUB_PATH=str(root/'path'))
    def run(code=0):
        (root/'path').write_text(''); current=dict(env)
        for block in blocks:
            block=block.replace('${{ github.event.pull_request.base.sha }}',sha).replace('${{ github.event.pull_request.head.sha }}',sha)
            result=subprocess.run(['bash','-c',block],cwd=leaf,env=current,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
            if result.returncode: break
            paths=(root/'path').read_text().splitlines()
            current['PATH']=':'.join(paths+[env['PATH']])
        assert result.returncode==code,(result.returncode,result.stdout)
        return (leaf/'events').read_text().splitlines()
    rows=run(); assert rows==['brew:install xianxu/ariadne/weave','brew:--prefix xianxu/ariadne/weave','compile','setup','checked'],rows
    (leaf/'go.mod').write_text('module github.com/xianxu/ariadne\ngo 1.26.3\n'); (leaf/'events').write_text(''); (leaf/'scripts/run-merge-checks.sh').unlink()
    rows=run(); assert rows==['brew:bundle --file=Brewfile','candidate-build','compile','setup','checked'],rows
    executable(hook,'exit 9\n'); run(9)
    hook.unlink(); rows=run(); assert rows[-1]=='checked'
    env['FAIL_COMPILE']='1'; (leaf/'events').write_text(''); (leaf/'scripts/run-merge-checks.sh').unlink(); rows=run(7); assert rows[-1]=='compile' and not (leaf/'scripts/run-merge-checks.sh').exists()
assert 'Homebrew/actions/setup-homebrew@master' in workflow
assert 'actions/setup-go' not in workflow and 'BOOTSTRAP_CLONE_ONLY' not in workflow
assert '../ariadne/scripts/run-merge-checks.sh' not in workflow
print('PASS CI actual blocks: published/source gateway, compile before hooks/checks, failures')
PY
