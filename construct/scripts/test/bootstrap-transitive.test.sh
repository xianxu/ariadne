#!/usr/bin/env bash
# Real bootstrap entrypoint, with isolated PATH executables (never installs).
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../../.." && pwd)"
python3 - "$SOURCE" <<'PY'
import os,pathlib,subprocess,sys,tempfile
source=pathlib.Path(sys.argv[1])
with tempfile.TemporaryDirectory(prefix='bootstrap gateway ') as tmp:
    root=pathlib.Path(tmp).resolve(); repo=root/'repo with spaces'; repo.mkdir(); commands=root/'commands'; commands.mkdir(); formula=root/'formula'; (formula/'bin').mkdir(parents=True)
    script=repo/'bootstrap.sh'; script.write_bytes((source/'bootstrap.sh').read_bytes())
    for name in ['bash','dirname']:
        import shutil
        (commands/name).symlink_to(shutil.which(name))
    env=dict(os.environ,PATH=str(commands),FIXTURE=str(root))
    def executable(path,body):
        path.write_text('#!/bin/bash\nset -eu\n'+body); path.chmod(0o755)
    gateway='if [ "$*" = "dependencies --help" ]; then exit 0; fi\nprintf "compile:%s:%s\\n" "$PWD" "$*" >> "$FIXTURE/events"\nexit "${COMPILE_EXIT:-0}"\n'
    def run(code=0):
        result=subprocess.run(['/bin/bash',str(script)],cwd=root,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
        assert result.returncode==code,(result.returncode,result.stdout)
        return result.stdout
    executable(commands/'weave',gateway)
    run(); assert (root/'events').read_text()==f'compile:{repo}:compile\n'
    env['COMPILE_EXIT']='7'; run(7); del env['COMPILE_EXIT']
    (commands/'weave').unlink()
    assert 'Homebrew' in run(1)
    executable(formula/'bin/weave',gateway)
    executable(commands/'brew','printf "brew:%s\\n" "$*" >> "$FIXTURE/events"\nif [ "$1" = install ]; then exit "${INSTALL_EXIT:-0}"; fi\nprintf "%s/formula\\n" "$FIXTURE"\n')
    (root/'events').write_text(''); run()
    assert (root/'events').read_text().splitlines()==['brew:install xianxu/ariadne/weave','brew:--prefix xianxu/ariadne/weave',f'compile:{repo}:compile']
    # An old source binary shadows brew; execute the formula binary explicitly.
    executable(commands/'weave','exit 2\n'); (root/'events').write_text(''); run()
    assert (root/'events').read_text().splitlines()[-1]==f'compile:{repo}:compile'
    env['INSTALL_EXIT']='9'; (root/'events').write_text(''); run(9)
    assert (root/'events').read_text()=='brew:install xianxu/ariadne/weave\n'
print('PASS bootstrap gateway: cwd/spaces, reuse, missing brew, install/shadowing/failure')
PY
