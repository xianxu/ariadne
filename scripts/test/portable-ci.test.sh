#!/usr/bin/env bash
# Run the workflow's actual Bash against stateful command/filesystem fixtures.
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../.." && pwd)"
python3 - "$SOURCE" <<'PY'
import os,pathlib,re,shutil,subprocess,sys,tempfile
source=pathlib.Path(sys.argv[1]); workflow=(source/'.github/workflows/merge-check.yml').read_text()
# Every step, `uses:` ones included (block None), so the runner model below
# decides which run on a GitHub-hosted and which on a self-hosted runner (#641).
steps=[]
for raw in re.split(r'^      - ',workflow,flags=re.M)[1:]:
    condition=re.search(r'^        if: (.+)$',raw,re.M)
    condition=condition[1] if condition else ''
    uses=re.search(r'^(?:        )?uses: (\S+)$',raw,re.M)
    match=re.search(r'^        run: \|\n((?:          .*\n|\n)+)',raw,re.M)
    if not match:
        steps.append((raw.splitlines()[0],None,condition,{},uses[1] if uses else ''))
        continue
    block=''.join(line[10:] if line.startswith('          ') else line for line in match[1].splitlines(True))
    environment=dict(re.findall(r'^          ([A-Z_]+): (.+)$',raw[:match.start()],re.M))
    steps.append((raw.splitlines()[0],block,condition,environment,''))
assert len([step for step in steps if step[1] is not None])>=8,steps
HOSTED="runner.environment == 'github-hosted'"; NOT_HOSTED="runner.environment != 'github-hosted'"
def holds(condition,runner,outputs):
    """The step conditions the seeds use: conjunctions of == / != over the runner and gateway outputs."""
    for term in filter(None,condition.split(' && ')):
        m=re.fullmatch(r"(runner\.environment|steps\.gateway\.outputs\.(\w+)) (==|!=) '([^']*)'",term)
        assert m,('a step condition the runner model cannot evaluate',condition)
        value=runner if m[2] is None else outputs.get(m[2],'')
        if (value==m[4])!=(m[3]=='=='):return False
    return True
# The runner is a per-repository choice; unset is GitHub-hosted as before (#641).
RUNS_ON="runs-on: ${{ vars.CI_RUNS_ON && fromJSON(vars.CI_RUNS_ON) || 'ubuntu-latest' }}"
for name in ('merge-check.yml','local-gates-guard.yml'):
    text=(source/'.github/workflows'/name).read_text()
    assert re.findall(r'^    runs-on: .*$',text,re.M)==['    '+RUNS_ON],(name,'runs-on')
    assert not re.search(r'^\s+ref:',text,re.M),(name,'checks out something other than the event commit')
for name,block,condition,environment,uses in steps:
    if uses.startswith(('Homebrew/actions/setup-homebrew@','actions/setup-go@')):
        assert HOSTED in condition.split(' && '),('a host-installing step can run on a self-hosted runner',name)
self_hosted=[step for step in steps if NOT_HOSTED in step[2].split(' && ')]
assert len(self_hosted)==2 and all(step[1] for step in self_hosted),self_hosted
for name,block,*_ in self_hosted:
    assert not re.search(r'\./|scripts/|bash |source ',block),('a self-hosted check runs a path from the tree',name)
pin='63d6bfd86509d97649e10d7a5dd1e9a06db17d21'; other='a'*40; old='4de31b2de32c8143d61abe17bbe7f7654503cd42'
assert 'scripts/ci/*' not in workflow  # no path filter can skip conformance
assert 'go-version-file: ${{ runner.temp }}/weave-ci/pin1/go.mod' in workflow
assert 'go-version-file: ${{ runner.temp }}/weave-ci/pin2/go.mod' in workflow
assert workflow.index('Select candidate toolchain') > workflow.index('Build second pinned compiler')
assert 'contents: read' in workflow
assert '${{' not in ''.join(step[1] or '' for step in steps), 'event values reach scripts only through env'
assert 'brew tap' not in workflow and 'BOOTSTRAP_CLONE_ONLY' not in workflow
# Exercise the exact template pipelines under pipefail, without paying for a
# compiler build per iteration. The large valid input forces an early reader
# exit to race a writer even when the three-line incident happens to pass.
metadata_checks='\n'.join(line.strip() for line in workflow.splitlines()
                          if 'grep ' in line and ('vcs.revision=' in line or 'vcs.modified=' in line))
assert len(metadata_checks.splitlines())==2,metadata_checks
metadata=f'binary: go1.26.3\n  build vcs.revision={pin}\n  build vcs.modified=false'
# Env carries fixture data into Bash; the real workflow assigns a local variable.
metadata_prelude='set -euo pipefail\nexport -n metadata\n'
for name,count in [('incident',1000),('large',20)]:
    # Expand inside Bash to exceed pipe buffers without hitting env/argv limits.
    padding="metadata+=$(printf '\\n  build irrelevant=%1048576s\\n' x)\n" if name=='large' else ''
    loop=metadata_prelude+padding+'for ((iteration=0; iteration<'+str(count)+'; iteration++)); do\n'+metadata_checks+'\ndone\n'
    result=subprocess.run(['bash','-c',loop],env=dict(os.environ,metadata=metadata,PIN_REF=pin),
                          text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=60)
    assert result.returncode==0,('metadata loop',name,result.returncode,result.stdout)
for name,value in [('wrong revision',metadata.replace(pin,other)),
                   ('dirty source',metadata.replace('modified=false','modified=true')),
                   ('missing revision','binary: go1.26.3\n  build vcs.modified=false'),
                   ('missing modified',f'binary: go1.26.3\n  build vcs.revision={pin}'),
                   ('empty','')]:
    result=subprocess.run(['bash','-c',metadata_prelude+metadata_checks],
                          env=dict(os.environ,metadata=value,PIN_REF=pin),
                          text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=10)
    assert result.returncode==1,('metadata refusal',name,result.returncode,result.stdout)
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
    executable(root/'gateway','if [ "$*" = "dependencies --help" ]; then [ "${NO_DEPENDENCIES:-0}" = 0 ]; exit 0; fi\necho "compile:$(cat "$0.identity")" >> "$FIXTURE/leaf/events"\n[ "${FAIL_COMPILE:-0}" = 0 ] || exit 7\ncp "$SOURCE/scripts/run-merge-checks.sh" scripts/run-merge-checks.sh\n')
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
    env=dict(os.environ,PATH=f'{commands}:/usr/bin:/bin',SOURCE=str(source),FIXTURE=str(root),RUNNER_TEMP=str(runner),GITHUB_WORKSPACE=str(leaf),GITHUB_PATH=str(root/'path'),GITHUB_OUTPUT=str(root/'output'),GITHUB_ENV=str(root/'github env'),REAL_GIT=realgit)
    # The same commands without go: a self-hosted runner missing a tool.
    nogo=root/'commands without go'; nogo.mkdir()
    for tool in ('git','brew','weave'):(nogo/tool).symlink_to(commands/tool)
    last={}
    def run(code=0,source_ci=False,selected=pin+'\n',supported=None,event=None,extra=None,module='github.com/xianxu/ariadne',machine='github-hosted',tools=commands):
        if runner.exists():shutil.rmtree(runner)
        runner.mkdir();(root/'path').write_text('');(root/'output').write_text('');(root/'github env').write_text('');(leaf/'events').write_text('')
        if selected is None:(leaf/'.sdlc/weave-compiler-ref').unlink(missing_ok=True)
        else:(leaf/'.sdlc/weave-compiler-ref').write_bytes(selected.encode())
        (leaf/'.sdlc/weave-compiler-refs').write_text(supported or pin+'\n')
        if source_ci:(leaf/'go.mod').write_text('module '+module+'\ngo 1.26.3\n')
        else:(leaf/'go.mod').unlink(missing_ok=True)
        current=dict(env,PATH=env['PATH'].replace(str(commands),str(tools)),RUNNER_ENVIRONMENT=machine,CI_REPOSITORY='xianxu/ariadne' if source_ci else 'xianxu/consumer',**(event or {'EVENT':'pull_request','PR_BASE':sha,'PR_HEAD':sha}),**(extra or {}));outputs={}
        taken=[]
        for name,block,condition,environment,uses in steps:
            if not holds(condition,machine,outputs):continue
            if block is None:
                taken.append(uses);continue
            step_env=dict(current)
            for key,value in environment.items():
                match=re.fullmatch(r'\$\{\{ steps.gateway.outputs.(\w+) \}\}',value)
                if match:value=outputs.get(match[1],'')
                elif '${{' in value:continue  # GitHub event env is provided by the fixture above.
                step_env[key]=value
            result=subprocess.run(['bash','-c',block],cwd=leaf,env=step_env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
            if result.returncode:break
            outputs=dict(line.split('=',1) for line in (root/'output').read_text().splitlines())
            paths=(root/'path').read_text().splitlines();current['PATH']=':'.join(reversed(paths))+':'+env['PATH'].replace(str(commands),str(tools))
            current.update(line.split('=',1) for line in (root/'github env').read_text().splitlines())
        assert result.returncode==code,(name,result.returncode,result.stdout)
        rows=(leaf/'events').read_text().splitlines();assert 'WRONG-PATH-COMPILER' not in rows,rows
        last.update(uses=taken,github_env=(root/'github env').read_text())
        return rows,result.stdout
    rows,_=run();assert rows==[f'fetch:{pin}',f'build:{pin}:./cmd/weave',f'compile:{pin}','setup','checked'],rows
    assert 'Homebrew/actions/setup-homebrew@main' in last['uses'] and 'actions/setup-go@v5' in last['uses'] and not last['github_env'],last
    # A self-hosted runner (#641) runs the same blocks with the host's tools:
    # nothing installs there, toolchains never download, and weave is checked
    # before bootstrap.sh so its brew-install branch cannot run.
    rows,_=run(machine='self-hosted');assert rows==[f'fetch:{pin}',f'build:{pin}:./cmd/weave',f'compile:{pin}','setup','checked'],rows
    assert last['uses']==['actions/checkout@v4'],last
    assert last['github_env'].splitlines()==['GOTOOLCHAIN=local','UV_PYTHON_DOWNLOADS=never'],last
    rows,out=run(1,machine='self-hosted',tools=nogo);assert not rows and 'lacks go' in out,(rows,out)
    rows,out=run(1,machine='self-hosted',extra={'NO_DEPENDENCIES':'1'})
    assert rows==[f'fetch:{pin}',f'build:{pin}:./cmd/weave'] and 'no weave with dependencies' in out,(rows,out)
    rows,out=run(1,source_ci=True,machine='self-hosted');assert not rows and 'GitHub-hosted' in out,(rows,out)
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
