#!/usr/bin/env bash
# Run the workflow's actual Bash against stateful command/filesystem fixtures.
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../.." && pwd)"
python3 - "$SOURCE" <<'PY'
import os,pathlib,re,shutil,subprocess,sys,tempfile
source=pathlib.Path(sys.argv[1]); workflow=(source/'.github/workflows/merge-check.yml').read_text()
steps=[]
for raw in re.split(r'^      - ',workflow,flags=re.M)[1:]:
    match=re.search(r'^        run: \|\n((?:          .*\n|\n)+)',raw,re.M)
    if not match: continue
    block=''.join(line[10:] if line.startswith('          ') else line for line in match[1].splitlines(True))
    condition=re.search(r'^        if: (.+)$',raw,re.M)
    environment=dict(re.findall(r'^          ([A-Z_]+): (.+)$',raw[:match.start()],re.M))
    steps.append((raw.splitlines()[0],block,condition[1] if condition else '',environment))
assert len(steps)>=8,steps
pin='63d6bfd86509d97649e10d7a5dd1e9a06db17d21'; other='a'*40; old='4de31b2de32c8143d61abe17bbe7f7654503cd42'
assert 'scripts/ci/*' not in workflow  # no path filter can skip conformance
assert 'go-version-file: ${{ runner.temp }}/weave-ci/pin1/go.mod' in workflow
assert 'go-version-file: ${{ runner.temp }}/weave-ci/pin2/go.mod' in workflow
assert workflow.index('Select candidate toolchain') > workflow.index('Build second pinned compiler')
assert 'contents: read' in workflow
assert '${{' not in ''.join(step[1] for step in steps), 'event values reach scripts only through env'
assert 'brew tap' not in workflow and 'BOOTSTRAP_CLONE_ONLY' not in workflow
with tempfile.TemporaryDirectory(prefix='portable ci ') as tmp:
    root=pathlib.Path(tmp).resolve(); leaf=root/'leaf'; leaf.mkdir(); commands=root/'commands'; commands.mkdir(); runner=root/'runner temp'
    realgit=shutil.which('git')
    subprocess.run([realgit,'init','-q',str(leaf)],check=True)
    subprocess.run([realgit,'-C',str(leaf),'-c','user.name=Test','-c','user.email=test@test','commit','--allow-empty','-qm','test'],check=True)
    sha=subprocess.check_output([realgit,'-C',str(leaf),'rev-parse','HEAD'],text=True).strip()
    def executable(path,body):
        path.parent.mkdir(parents=True,exist_ok=True);path.write_text('#!/bin/bash\nset -eu\n'+body);path.chmod(0o755)
    (leaf/'scripts/merge-checks.d').mkdir(parents=True);(leaf/'.sdlc').mkdir()
    (leaf/'bootstrap.sh').write_bytes((source/'bootstrap.sh').read_bytes());(leaf/'bootstrap.sh').chmod(0o755)
    (leaf/'Brewfile').write_text('brew "go"\nbrew "cue"\n')
    executable(root/'gateway','if [ "$*" = "dependencies --help" ]; then exit 0; fi\necho "compile:$(cat "$0.identity")" >> "$FIXTURE/leaf/events"\n[ "${FAIL_COMPILE:-0}" = 0 ] || exit 7\ncp "$SOURCE/scripts/run-merge-checks.sh" scripts/run-merge-checks.sh\n')
    executable(commands/'weave','echo WRONG-PATH-COMPILER >> "$FIXTURE/leaf/events"; exit 99\n')
    executable(commands/'git',r'''
if [ "${1:-}" = -C ]; then
  dir="$2"; shift 2
  if [ "$1" = fetch ]; then
    echo "fetch:${!#}" >> "$FIXTURE/leaf/events"
    [ "${FAIL_FETCH:-0}" = 0 ] || exit 11
    printf '%s\n' "${!#}" > "$dir/.test-pin"
    printf 'module github.com/xianxu/ariadne\ngo 1.26.3\n' > "$dir/go.mod"
    touch "$dir/go.sum"
    exit 0
  fi
  if [ "$1" = checkout ] && [ -f "$dir/.test-pin" ]; then exit 0; fi
  if [ "$1 $2" = 'rev-parse HEAD' ] && [ -f "$dir/.test-pin" ]; then
    if [ "${FAIL_IDENTITY:-0}" = 1 ]; then echo bad; else cat "$dir/.test-pin"; fi
    exit 0
  fi
  exec "$REAL_GIT" -C "$dir" "$@"
fi
exec "$REAL_GIT" "$@"
''')
    executable(commands/'go',r'''
if [ "$1" = version ]; then
  if [ "${2:-}" = -m ]; then
    revision="$(cat "$3.identity")"
    [ "${FAIL_METADATA:-0}" = 0 ] || revision=wrong
    modified=false; [ "${FAIL_MODIFIED:-0}" = 0 ] || modified=true
    printf '%s\n' "$3: go1.26.3" "  build vcs.revision=$revision" "  build vcs.modified=$modified"
  else echo 'go version go1.26.3 fixture'; fi
  exit 0
fi
[ "$1" = build ] && [ "$2" = -trimpath ] && [ "$3" = -o ]
[ "${GOTOOLCHAIN:-}" = local ] && [ "${CGO_ENABLED:-}" = 0 ]
revision=candidate; [ ! -f .test-pin ] || revision="$(cat .test-pin)"
echo "build:$revision:${!#}" >> "$FIXTURE/leaf/events"
[ "${FAIL_BUILD:-0}" = 0 ] || exit 12
cp "$FIXTURE/gateway" "$4"
printf '%s\n' "$revision" > "$4.identity"
''')
    executable(commands/'brew','echo "brew:$*" >> "$FIXTURE/leaf/events"\n[ "$1" = bundle ] || exit 88\n')
    hook=leaf/'scripts/ci-setup.sh'; executable(hook,'test -f scripts/run-merge-checks.sh\necho setup >> events\n')
    executable(leaf/'scripts/merge-checks.d/01-product.sh','echo checked >> events\n')
    guard=leaf/'scripts/ci/weave-compatibility.py';guard.parent.mkdir()
    guard.write_text('''import os,pathlib,sys
args=sys.argv[1:]
if args[0]=='refs':
    refs=pathlib.Path('.sdlc/weave-compiler-refs').read_text();print(refs,end='')
else:
    compiler=pathlib.Path(args[args.index('--compiler')+1]);identity=pathlib.Path(str(compiler)+'.identity').read_text().strip()
    with open('events','a') as f:f.write('check:'+identity+'\\n')
    if identity=='4de31b2de32c8143d61abe17bbe7f7654503cd42':
        print('preserving authored replacement' if not os.getenv('WRONG_OLD_FAILURE') else 'unrelated failure')
        if not os.getenv('MISSING_PRESERVATION_MARKER'):print('compatibility: COLD-REFUSAL-PRESERVED')
        sys.exit(1)
    if os.getenv('FAIL_CHECK'):sys.exit(14)
''')
    env=dict(os.environ,PATH=f'{commands}:/usr/bin:/bin',SOURCE=str(source),FIXTURE=str(root),RUNNER_TEMP=str(runner),GITHUB_WORKSPACE=str(leaf),GITHUB_PATH=str(root/'path'),GITHUB_OUTPUT=str(root/'output'),REAL_GIT=realgit)
    def run(code=0,source_ci=False,selected=pin+'\n',supported=None,event=None,extra=None,module='github.com/xianxu/ariadne'):
        if runner.exists():shutil.rmtree(runner)
        runner.mkdir();(root/'path').write_text('');(root/'output').write_text('');(leaf/'events').write_text('')
        if selected is None:(leaf/'.sdlc/weave-compiler-ref').unlink(missing_ok=True)
        else:(leaf/'.sdlc/weave-compiler-ref').write_bytes(selected.encode())
        (leaf/'.sdlc/weave-compiler-refs').write_text(supported or pin+'\n')
        if source_ci:(leaf/'go.mod').write_text('module '+module+'\ngo 1.26.3\n')
        else:(leaf/'go.mod').unlink(missing_ok=True)
        current=dict(env,CI_REPOSITORY='xianxu/ariadne' if source_ci else 'xianxu/consumer',**(event or {'EVENT':'pull_request','PR_BASE':sha,'PR_HEAD':sha}),**(extra or {}));outputs={}
        for name,block,condition,environment in steps:
            if condition=="steps.gateway.outputs.pin2 != ''" and not outputs.get('pin2'):continue
            if condition=="steps.gateway.outputs.source_ci == 'true'" and outputs.get('source_ci')!='true':continue
            step_env=dict(current)
            for key,value in environment.items():
                match=re.fullmatch(r'\$\{\{ steps.gateway.outputs.(\w+) \}\}',value)
                if match:value=outputs.get(match[1],'')
                elif '${{' in value:continue  # GitHub event env is provided by the fixture above.
                step_env[key]=value
            result=subprocess.run(['bash','-c',block],cwd=leaf,env=step_env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
            if result.returncode:break
            outputs=dict(line.split('=',1) for line in (root/'output').read_text().splitlines())
            paths=(root/'path').read_text().splitlines();current['PATH']=':'.join(reversed(paths))+':'+env['PATH']
        assert result.returncode==code,(name,result.returncode,result.stdout)
        rows=(leaf/'events').read_text().splitlines();assert 'WRONG-PATH-COMPILER' not in rows,rows
        return rows,result.stdout
    rows,_=run();assert rows==[f'fetch:{pin}',f'build:{pin}:./cmd/weave',f'compile:{pin}','setup','checked'],rows
    # A second declared compiler remains independent; source CI uses candidate for bootstrap.
    rows,_=run(source_ci=True,supported=pin+'\n'+other+'\n')
    assert rows.count(f'check:{pin}')==1 and rows.count(f'check:{other}')==1 and rows.count(f'check:{old}')==1,rows
    assert 'compile:candidate' in rows and f'compile:{pin}' not in rows,rows
    assert rows.index('compile:candidate') < rows.index(f'check:{pin}') < rows.index('setup') < rows.index('checked'),rows
    assert rows.count('build:candidate:./cmd/vocabulary')==1,rows
    assert 'CI_REPOSITORY' in workflow, 'source guard identity must not depend on candidate go.mod'
    # Pin grammar is checked before any fetch/build/bootstrap effect.
    for bad in (None,'main\n',pin,pin.upper()+'\n',pin+'\n'+other+'\n',pin+'\n; touch BAD\n','x'*5000):
        rows,out=run(1,selected=bad);assert not rows,(bad,rows,out)
    for flag,code in [('FAIL_FETCH',11),('FAIL_IDENTITY',1),('FAIL_BUILD',12),('FAIL_METADATA',1),('FAIL_MODIFIED',1)]:
        rows,_=run(code,extra={flag:'1'});assert not any(x.startswith('compile:') for x in rows),rows
    rows,_=run(1,source_ci=True,module='example.invalid/renamed');assert not rows,rows
    rows,_=run(7,extra={'FAIL_COMPILE':'1'});assert rows[-1]==f'compile:{pin}',rows
    run(14,source_ci=True,extra={'FAIL_CHECK':'1'})
    run(1,source_ci=True,extra={'WRONG_OLD_FAILURE':'1'})
    run(1,source_ci=True,extra={'MISSING_PRESERVATION_MARKER':'1'})
    executable(hook,'exit 9\n');run(9);hook.unlink();rows,_=run();assert rows[-1]=='checked'
    rows,_=run(event={'EVENT':'push','PUSH_BEFORE':sha,'PUSH_AFTER':sha});assert rows[-1]=='checked'
    rows,_=run(source_ci=True,event={'EVENT':'push','PUSH_BEFORE':'0'*40,'PUSH_AFTER':sha});assert 'checked' not in rows and f'check:{pin}' in rows,rows
for name in ('merge-check.yml','weave-release.yml'):
    assert 'Homebrew/actions/setup-homebrew@main' in (source/'.github/workflows'/name).read_text()
assert '.sdlc/weave-compiler-ref' not in (source/'construct/base.manifest').read_text(), 'authored pins must not be generated'
print('PASS CI actual blocks: exact source pins, provenance, old PATH exclusion, source candidate separation, cold control, failure propagation and ranges')
PY
