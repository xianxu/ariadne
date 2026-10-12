#!/usr/bin/env bash
# Execute local-gates-guard.yml's actual run block against scratch repositories (#309).
# Matcher cases are generated from the workflow's own pattern list, plus a
# near-miss mutation of each; the decision is touched/untouched x labeled/unlabeled.
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../.." && pwd)"
python3 - "$SOURCE" <<'PY'
import os,pathlib,re,subprocess,sys,tempfile,json
source=pathlib.Path(sys.argv[1]); workflow=(source/'.github/workflows/local-gates-guard.yml').read_text()
blocks=[''.join(line[10:] if line.startswith('          ') else line for line in m[1].splitlines(True)) for m in re.finditer(r'^        run: \|\n((?:          .*\n|\n)+)',workflow,re.M)]
assert len(blocks)==1,len(blocks)
block=blocks[0]
assert '${{' not in block, 'PR values must reach the script through env:, never interpolation'
assert 'pull_request_target:' in workflow and 'contents: read' in workflow
patterns=re.findall(r"^  '([^']+)'$",block,re.M)
assert len(patterns)>=10,patterns

def sample(p):   # a concrete path the pattern must match
    return p.replace('*','x/y') if p.endswith('/*') else p.replace('*','.local')
def near_miss(p):  # a sibling the pattern must not match
    stem=p.rstrip('/*')
    return stem+'.md' if p.endswith('/*') else sample(p)+'.bak'

with tempfile.TemporaryDirectory(prefix='guard ') as tmp:
    root=pathlib.Path(tmp); server=root/'server.git'; work=root/'work'; bin_=root/'bin'; bin_.mkdir()
    g=lambda *a,cwd=None: subprocess.check_output(['git',*a],cwd=cwd,text=True).strip()
    g('init','-q','--bare','-b','main',str(server))
    g('init','-q','-b','main',str(work))
    for k,v in (('user.name','T'),('user.email','t@t')): g('config',k,v,cwd=work)
    (work/'README').write_text('x\n'); g('add','.',cwd=work); g('commit','-qm','base',cwd=work)
    g('remote','add','origin',str(server),cwd=work); g('push','-q','origin','main',cwd=work)
    base=g('rev-parse','HEAD',cwd=work)
    # gh: labels are read live (GH_LABELS, or GH_LABELS_FAIL); statuses are logged.
    (bin_/'gh').write_text('#!/bin/bash\nif [ "$1 $2" = "api repos/$REPO/issues/$PR_NUMBER/labels" ]; then [ -z "${GH_LABELS_FAIL:-}" ] || exit 1; printf "%s" "${GH_LABELS:-}"; exit 0; fi\nprintf "%s\\n" "$*" >> "$STATUS_LOG"\n'); (bin_/'gh').chmod(0o755)
    statuses=root/'statuses'
    n=[0]
    def pr(paths,labels,extra=None,head_override=None):
        n[0]+=1; num=n[0]
        g('checkout','-q','-B',f'pr{num}',base,cwd=work)
        for p in paths:
            f=work/p; f.parent.mkdir(parents=True,exist_ok=True); f.write_text(f'{num}\n')
        g('add','-f','--',*paths,cwd=work); g('commit','-qm',f'pr{num}',cwd=work)
        head=g('rev-parse','HEAD',cwd=work)
        g('push','-q','origin',f'HEAD:refs/pull/{num}/head',cwd=work)
        guard=root/f'guard{num}'; g('clone','-q',str(server),str(guard))
        statuses.write_text('')
        env=dict(os.environ,PATH=f'{bin_}:{os.environ["PATH"]}',STATUS_LOG=str(statuses),GH_TOKEN='t',REPO='o/r',
                 PR_NUMBER=str(num),BASE_SHA=base,HEAD_SHA=head_override or head,GH_LABELS='\n'.join(labels)+'\n')
        env.update(extra or {})
        r=subprocess.run(['bash','-c',block],cwd=guard,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
        return r.returncode,r.stdout,statuses.read_text()

    # Matcher: every pattern's sample is caught (and named); its near miss is not.
    for p in patterns:
        code,out,st=pr([sample(p)],[])
        assert code==1 and sample(p) in out and 'state=failure' in st, (p,code,out,st)
        code,out,st=pr([near_miss(p)],[])
        assert code==0 and 'state=success' in st, (p,near_miss(p),code,out)
    # Paths git would quote without -z (non-ASCII, quotes, backslashes) are still caught.
    for odd in ['scripts/merge-checks.d/50-é.sh','.github/workflows/a"b.yml','.sdlc/back\\slash.json']:
        code,out,st=pr([odd],[]); assert code==1 and odd in out and 'state=failure' in st,(odd,code,out)
    # A label is matched whole, not as a substring.
    code,out,st=pr(['bootstrap.sh'],['local-gates-verified-not']); assert code==1,out
    # Labels unreadable: refuse and say so (fail closed).
    code,out,st=pr(['bootstrap.sh'],['local-gates-verified'],extra={'GH_LABELS_FAIL':'1'}); assert code==2 and 'state=failure' in st,(code,out)
    # The PR head moved since the event: this run decides nothing and posts nothing.
    code,out,st=pr(['bootstrap.sh'],[],head_override='0'*40); assert code==0 and 'moved' in out and st.count('state=')==1 and 'state=pending' in st,(code,out,st)
    # Every decided run starts pending and ends with exactly one final verdict.
    code,out,st=pr(['README'],[]); assert st.splitlines()[0].count('state=pending')==1 and st.count('state=')==2 and 'state=success' in st.splitlines()[-1],st
    # A run that stops before deciding (the PR ref cannot be fetched) leaves a failure, never silence.
    code,out,st=pr(['README'],[],extra={'PR_NUMBER':'999'}); assert code!=0 and 'state=failure' in st.splitlines()[-1] and 'stopped before deciding' in st,(code,out,st)
    # Gate code is not protected.
    code,out,st=pr(['cmd/sdlc/merge.go'],[]); assert code==0,out
    # Decision: touched x labeled.
    code,out,st=pr(['scripts/merge-checks.d/40-x.sh'],['local-gates-verified']); assert code==0 and 'operator-verified' in out and 'state=success' in st,out
    code,out,st=pr(['README'],['local-gates-verified']); assert code==0,out
    code,out,st=pr(['.github/workflows/merge-check.yml','README'],['other']); assert code==1 and 'README' not in out.split('configuration:')[1].split('::error')[0],out
    # The status is posted on the PR head, under one context.
    assert 'context=local-gates-guard/verdict' in st and '/statuses/' in st
    # A deletion is a touch (rename sources count too).
    g('checkout','-q','-B','seed',base,cwd=work); (work/'bootstrap.sh').write_text('x\n'); g('add','.',cwd=work); g('commit','-qm','seed',cwd=work); g('push','-q','origin','HEAD:main',cwd=work)
    base=g('rev-parse','HEAD',cwd=work)
    n[0]+=1; num=n[0]; g('rm','-q','bootstrap.sh',cwd=work); g('commit','-qm','rm',cwd=work); head=g('rev-parse','HEAD',cwd=work)
    g('push','-q','origin',f'HEAD:refs/pull/{num}/head',cwd=work); guard=root/'guardrm'; g('clone','-q',str(server),str(guard))
    env=dict(os.environ,PATH=f'{bin_}:{os.environ["PATH"]}',STATUS_LOG=str(statuses),GH_TOKEN='t',REPO='o/r',PR_NUMBER=str(num),BASE_SHA=base,HEAD_SHA=head,GH_LABELS='')
    r=subprocess.run(['bash','-c',block],cwd=guard,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
    assert r.returncode==1 and 'bootstrap.sh' in r.stdout, r.stdout
    # Hostile inputs refuse before any effect.
    env2=dict(env,PR_NUMBER='1; rm -rf /'); r=subprocess.run(['bash','-c',block],cwd=guard,env=env2,text=True,capture_output=True); assert r.returncode==2,r
print(f'PASS local-gates-guard: {len(patterns)} patterns matched with near misses; quoted paths, live labels, head moved, deletion and input checks')
PY
