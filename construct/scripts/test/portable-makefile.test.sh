#!/usr/bin/env bash
# Execute owner tools and consumer aliases with no generated helper links.
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../../.." && pwd)"
python3 - "$SOURCE" <<'PY'
import os,pathlib,subprocess,sys,tempfile
source=pathlib.Path(sys.argv[1])
with tempfile.TemporaryDirectory(prefix='portable make ') as tmp:
    root=pathlib.Path(tmp).resolve(); owner=root/'owner'; owner.mkdir(); commands=root/'commands'; commands.mkdir()
    for name in ['Makefile','Makefile.workflow']:
        (owner/name).write_bytes((source/name).read_bytes())
    def executable(path,body):
        path.write_text('#!/bin/sh\nset -eu\n'+body); path.chmod(0o755)
    executable(commands/'go','printf "go:%s:%s\\n" "$PWD" "$*" >> "$FIXTURE/events"\n[ "$1" = build ]\n[ "$2" = -o ]\nmkdir -p "$(dirname "$3")"\ntouch "$3"\n')
    executable(commands/'weave','printf "weave:%s:%s\\n" "$PWD" "$*" >> "$FIXTURE/events"\nexit "${FAIL_WEAVE:-0}"\n')
    env=dict(os.environ,PATH=f'{commands}:'+os.environ['PATH'],FIXTURE=str(root))
    def make(repo,*targets,code=0):
        result=subprocess.run(['make','-s','-j4',*targets],cwd=repo,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
        assert result.returncode==code,result.stdout
    make(owner,'tools')
    rows=(root/'events').read_text().splitlines()
    assert len(rows)==4,rows
    assert {p.name for p in (owner/'bin').iterdir()}=={'sdlc','datatype','vocabulary','doc-review'},rows
    assert all(row.startswith(f'go:{owner}:build -o bin/') for row in rows),rows
    leaf=root/'leaf'; leaf.mkdir()
    # The consumer authors its own root and needs no ariadne source build rules.
    (leaf/'Makefile').write_text('include Makefile.workflow\nproduct:\n\t@echo product-ok\n')
    (leaf/'Makefile.workflow').write_bytes((source/'Makefile.workflow').read_bytes())
    (root/'events').write_text(''); make(leaf,'weave'); make(leaf,'bootstrap'); make(leaf,'product')
    assert (root/'events').read_text().splitlines()==[f'weave:{leaf}:compile']*2
    assert not (leaf/'bootstrap').exists()
    env['FAIL_WEAVE']='7'; make(leaf,'bootstrap',code=2)
    # Explicit sdlc-install still builds in the owner, with no consumer build rule.
    (leaf/'scripts').mkdir(); (leaf/'construct').mkdir(); (root/'home').mkdir()
    (leaf/'scripts/sdlc-install.sh').write_bytes((source/'scripts/sdlc-install.sh').read_bytes())
    executable(leaf/'construct/dev-aliases.sh',f'printf "sdlc\\t%s\\n" "{owner}"\n')
    (owner/'bin/sdlc').chmod(0o755)
    env.update(HOME=str(root/'home'),SHELL='/bin/bash'); del env['FAIL_WEAVE']
    result=subprocess.run(['bash','scripts/sdlc-install.sh'],cwd=leaf,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
    assert result.returncode==0,result.stdout
    assert str(owner/'bin') in (root/'home/.bashrc').read_text()
    manifest=(source/'construct/base.manifest').read_text().splitlines()
    assert not any(row.split()==['seed','Makefile'] for row in manifest),'consumer root Makefile is authored'
print('PASS owner tools independent of compilation; consumer aliases delegate CLI')
PY
