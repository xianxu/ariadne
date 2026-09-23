#!/usr/bin/env bash
# Real product acceptance for nested workspace source isolation.
# Inputs: WEAVE_BIN (optional; otherwise build candidate), ARIADNE_SOURCE,
# ARIADNE_REF (HEAD), ARIADNE_WORKTREE (1: include local tracked/cmd/pkg edits),
# PARLEY_SOURCE, PARLEY_REF (HEAD containing remote source metadata), PLENARY.
# RUN_PRODUCT_TESTS=0 skips the full Parley suite only; runtime probes always run.
# All clones, builds, HOME/XDG state and evidence stay in a retained /tmp tree.
# Existing Go module cache is reused with GOPROXY=off; brew is an inert fixture.
set -euo pipefail
export SLOT_TEST_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
exec python3 - <<'PY'
import hashlib,json,os,pathlib,shutil,subprocess,tempfile,time
P=pathlib.Path
repo=P(os.environ['SLOT_TEST_ROOT'])
scratch=P(tempfile.mkdtemp(prefix='slot-dependencies.',dir='/tmp')).resolve()
print('Evidence:',scratch,flush=True)
source=P(os.environ.get('ARIADNE_SOURCE',repo)).resolve()
parley=P(os.environ.get('PARLEY_SOURCE',repo.parent/'parley.nvim')).resolve()
real_git=shutil.which('git'); real_go=shutil.which('go')
env={k:v for k,v in os.environ.items() if not k.startswith(('GIT_','XDG_'))}
for key,leaf in [('HOME','home'),('XDG_CONFIG_HOME','config'),('XDG_DATA_HOME','data'),('XDG_STATE_HOME','state'),('XDG_CACHE_HOME','cache'),('TMPDIR','tmp')]:
    path=scratch/leaf;path.mkdir();env[key]=str(path)
module_cache=subprocess.check_output([real_go,'env','GOMODCACHE'],text=True).strip()
env.update(GIT_CONFIG_NOSYSTEM='1',GIT_CONFIG_GLOBAL=str(scratch/'home/.gitconfig'),GIT_TERMINAL_PROMPT='0',GOPROXY='off',GOSUMDB='off',GOMODCACHE=module_cache,GOCACHE=str(scratch/'go-cache'))
bin=scratch/'bin';bin.mkdir();env['PATH']=str(bin)+os.pathsep+os.environ['PATH']
(bin/'brew').write_text('#!/bin/sh\nprintf "brew fixture: %s\\n" "$*"\n');(bin/'brew').chmod(0o755)
records=[]
def run(args,cwd=None,log=None,check=True):
    args=list(map(str,args));start=time.monotonic()
    result=subprocess.run(args,cwd=cwd,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
    output=result.stdout.decode(errors='replace')
    if log:(scratch/log).write_text(output)
    records.append(dict(command=args,cwd=str(cwd) if cwd else None,seconds=round(time.monotonic()-start,3),status=result.returncode,log=log))
    (scratch/'commands.json').write_text(json.dumps(records,indent=2)+'\n')
    if check and result.returncode:raise RuntimeError(f'{args} failed ({result.returncode})\n{output[-12000:]}')
    return output.strip()
def git(cwd,*args):return run([real_git,'-C',cwd,*args])
def snapshot(root,exclude_owner_builds=False):
    out={}
    for base,dirs,files in os.walk(root,followlinks=False):
        dirs[:]=sorted(d for d in dirs if d!='.git')
        for name in sorted(files+dirs):
            p=P(base)/name;key=str(p.relative_to(root))
            if exclude_owner_builds and key.startswith('ariadne/bin/'):continue
            if name=='.git':continue
            if p.is_symlink():out[key]=['link',os.readlink(p)]
            elif p.is_file():out[key]=['file',hashlib.sha256(p.read_bytes()).hexdigest()]
    return out
run([real_git,'config','--global','user.name','Slot Acceptance'])
run([real_git,'config','--global','user.email','slot@example.test'])
run([real_git,'config','--global','core.hooksPath','/dev/null'])
run([real_git,'config','--global','protocol.file.allow','always'])
# Archive then overlay through a patch, never modifying the source index or refs.
owner=scratch/'source-ariadne';owner.mkdir()
base_ref=os.environ.get('ARIADNE_REF','HEAD')
archive=scratch/'ariadne.tar'
with archive.open('wb') as out:subprocess.run([real_git,'-C',str(source),'archive',base_ref],stdout=out,check=True,env=env)
run(['tar','-xf',archive,'-C',owner])
run([real_git,'init','-q','-b','main',owner])
worktree_overlay=os.environ.get('ARIADNE_WORKTREE','1')=='1'
if worktree_overlay:
    patch=subprocess.check_output([real_git,'-C',str(source),'diff','--binary',base_ref,'--','pkg/workspace','cmd/weave','cmd/sdlc'],env=env)
    if patch:
        patchfile=scratch/'candidate.patch';patchfile.write_bytes(patch);run([real_git,'apply',patchfile],cwd=owner)
    untracked=subprocess.check_output([real_git,'-C',str(source),'ls-files','--others','--exclude-standard','-z','--','cmd','pkg'],env=env)
    for raw in untracked.split(b'\0'):
        if not raw:continue
        rel=os.fsdecode(raw);dest=owner/rel;dest.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(source/rel,dest,follow_symlinks=False)
git(owner,'add','-f','.');git(owner,'commit','-qm','candidate source snapshot')
ariadne_sha=git(owner,'rev-parse','HEAD')
remote=scratch/'ariadne.git';run([real_git,'clone','--bare',owner,remote])
https='https://github.com/xianxu/ariadne.git'
run([real_git,'config','--global',f'url.file://{remote}.insteadOf',https])
fleet=scratch/'fleet';fleet.mkdir()
primary=fleet/'parley.nvim'
run([real_git,'clone','--no-hardlinks',parley,primary])
parley_ref=os.environ.get('PARLEY_REF','HEAD')
parley_sha=git(parley,'rev-parse',parley_ref)
git(primary,'checkout','-B','main',parley_sha)
assert f'substrate ../ariadne {https}' in (primary/'construct/deps').read_text(),'PARLEY_REF must contain the actual remote source-metadata commit'
# Canonical Ariadne exists as a contrasting, untouched fleet checkout.
canonical=fleet/'ariadne';run([real_git,'clone',https,canonical])
weave=P(os.environ.get('WEAVE_BIN',scratch/'weave')).resolve()
if not os.environ.get('WEAVE_BIN'):run([real_go,'build','-o',weave,'./cmd/weave'],cwd=owner,log='gateway-build.log')
primary_before=snapshot(primary);canonical_before=snapshot(canonical)
slots=[]
for number in [1,2]:
    host=fleet/'worktree'/f'parley.nvim-slot{number}'/'parley.nvim'
    git(primary,'worktree','add','-b',f'main-slot{number}',host);slots.append(host)
primary_refs=git(primary,'show-ref');canonical_refs=git(canonical,'show-ref')
for number,host in enumerate(slots,1):
    run([weave,'compile'],cwd=host,log=f'compile-{number}-initial.log')
    dep=host.parent/'ariadne'
    assert git(dep,'rev-parse','HEAD')==ariadne_sha
    assert git(dep,'rev-parse','refs/remotes/origin/main')==ariadne_sha
    assert (dep/'.git').is_dir() and not (dep/'.git').is_symlink()
    assert git(dep,'config','--get','remote.origin.url')==https
    assert (host/'Makefile.workflow').resolve()==dep/'Makefile.workflow'
    assert (host/'.agents/skills/xx-issues').resolve()==dep/'construct/local/issues'
    assert (host/'.agents/skills/xx-datatype/SKILL.md').is_file()
    for tool in ['sdlc','datatype','vocabulary','doc-review']:assert os.access(dep/'bin'/tool,os.X_OK),tool
    (scratch/f'tracked-diff-{number}.txt').write_text(git(host,'diff','--stat')+'\n'+git(host,'diff')+'\n')
    (scratch/f'owner-tracked-diff-{number}.txt').write_text(git(dep,'diff','--stat')+'\n'+git(dep,'diff')+'\n')
    # Go embeds VCS dirty/revision metadata; rebuilding after first composition
    # may change owner binary bytes. Compare composed files and links separately.
    run([real_go,'version','-m',dep/'bin/sdlc'],log=f'owner-build-{number}-initial.txt')
    first=snapshot(host.parent,True)
    run([weave,'compile'],cwd=host,log=f'compile-{number}-repeat.log')
    run([real_go,'version','-m',dep/'bin/sdlc'],log=f'owner-build-{number}-repeat.txt')
    second=snapshot(host.parent,True)
    assert first==second,f'environment {number} did not stabilize: {[k for k in first.keys()|second.keys() if first.get(k)!=second.get(k)]}'
assert snapshot(primary)==primary_before,'canonical Parley content changed'
assert snapshot(canonical)==canonical_before,'canonical Ariadne content changed'
# A real selected feature commit plus tracked and untracked edits survive reuse.
dep=slots[0].parent/'ariadne';other_before=snapshot(slots[1].parent,True)
git(dep,'checkout','-b','acceptance-feature')
(dep/'acceptance-feature.txt').write_text('unpublished feature\n');git(dep,'add','acceptance-feature.txt');git(dep,'commit','-qm','private feature')
feature_sha=git(dep,'rev-parse','HEAD')
(dep/'acceptance-feature.txt').write_text('unpublished feature\ndirty edit\n')
(dep/'acceptance-untracked.txt').write_text('private untracked\n')
refs_before=git(dep,'show-ref');content_before=snapshot(slots[0].parent,True)
run([weave,'compile'],cwd=slots[0],log='compile-1-feature.log')
assert git(dep,'rev-parse','HEAD')==feature_sha and git(dep,'branch','--show-current')=='acceptance-feature'
assert git(dep,'show-ref')==refs_before and snapshot(slots[0].parent,True)==content_before,'feature or dirty files changed'
assert snapshot(slots[1].parent,True)==other_before,'other environment changed'
assert snapshot(primary)==primary_before and snapshot(canonical)==canonical_before,'canonical fleet changed'
assert git(primary,'show-ref')==primary_refs and git(canonical,'show-ref')==canonical_refs,'canonical fleet refs changed'
# Runtime reads an archived committed Parley; product tests use the private host.
run(['sh','scripts/check-fresh-clone.sh','--runtime','--ref','HEAD'],cwd=slots[0],log='parley-runtime.log')
product_tests=os.environ.get('RUN_PRODUCT_TESTS','1')!='0'
if product_tests:
    plenary=os.environ.get('PLENARY') or os.environ.get('NVIM_TEST_PLENARY')
    if not plenary:plenary=str(P(os.environ['HOME'])/'.local/share/nvim/lazy/plenary.nvim')
    assert (P(plenary)/'lua/plenary/init.lua').is_file(),'Set PLENARY to an installed plenary.nvim checkout'
    run(['make','test',f'PLENARY={P(plenary).resolve()}',f'TEST_ENV_ROOT={scratch}/parley-tests'],cwd=slots[0],log='parley-tests.log')
report=dict(evidence=str(scratch),product_tests=product_tests,source_mode="worktree-overlay" if worktree_overlay else "committed-ref",ariadne_base=git(source,'rev-parse',base_ref),ariadne_fixture_main=ariadne_sha,parley=parley_sha,feature=feature_sha,limitations=['HTTPS transport rewritten to local fixture remote below source validation','Homebrew package installation is an inert fixture; installed toolchain used','Stable composition comparison excludes owner bin bytes: Go embeds changed vcs.modified/revision metadata; provenance is recorded',('Candidate source snapshot includes tracked workspace/Weave/SDLC changes and untracked cmd/pkg sources; no source checkout mutation' if worktree_overlay else 'Candidate source is an archive of the committed ref; fixture commit preserves its tree with synthetic history')])
(scratch/'report.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps(report,indent=2),flush=True)
print('PASS: two real environments, stable repeat composition, private feature preservation, Parley runtime'+(' and product tests' if product_tests else ' (full product suite explicitly skipped)'),flush=True)
PY
