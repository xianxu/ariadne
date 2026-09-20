#!/usr/bin/env bash
# Prepare local artifacts only. Publication and tap updates belong to #241.
set -euo pipefail
repo_root="$(cd "$(dirname "$0")/.." && pwd -P)"
if [ "$#" -ne 2 ]; then
    echo 'usage: scripts/release-weave.sh weave-vVERSION NEW_OUTPUT_DIRECTORY' >&2
    exit 2
fi
python3 - "$repo_root" "$1" "$2" <<'PY'
import gzip,hashlib,io,os,pathlib,re,subprocess,sys,tarfile,tempfile
root=pathlib.Path(sys.argv[1]); tag=sys.argv[2]; output=pathlib.Path(sys.argv[3]).absolute()
match=re.fullmatch(r'weave-v((?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*))',tag)
if not match:
    sys.exit('invalid release tag: expected weave-vMAJOR.MINOR.PATCH')
version=match[1]
if output.exists() or output.is_symlink():
    sys.exit(f'output already exists: {output}')
output.parent.mkdir(parents=True,exist_ok=True)
formula=(root/'packaging/homebrew/Formula/weave.rb').read_text().replace('@WEAVE_VERSION@',version)
try:
    with tempfile.TemporaryDirectory(prefix='.weave-release-',dir=output.parent) as stage_path:
        stage=pathlib.Path(stage_path); artifacts=stage/'artifacts'; artifacts.mkdir(); checksums=[]
        for osname in ['darwin','linux']:
            for arch in ['arm64','amd64']:
                binary=stage/'weave'
                subprocess.run(['go','build','-trimpath','-ldflags',f'-s -w -X main.version={version}','-o',str(binary),'./cmd/weave'],cwd=root,env=dict(os.environ,CGO_ENABLED='0',GOOS=osname,GOARCH=arch),check=True)
                name=f'weave_{version}_{osname}_{arch}.tar.gz'; archive=artifacts/name
                # Fixed metadata makes packaging repeatable for identical binaries.
                with archive.open('wb') as target:
                    with gzip.GzipFile(filename='',mode='wb',fileobj=target,mtime=0) as zipped:
                        with tarfile.open(fileobj=zipped,mode='w') as tar:
                            data=binary.read_bytes(); entry=tarfile.TarInfo('weave'); entry.size=len(data); entry.mode=0o755; entry.mtime=0
                            tar.addfile(entry,io.BytesIO(data))
                digest=hashlib.sha256(archive.read_bytes()).hexdigest(); checksums.append(f'{digest}  {name}\n')
                key=f'@WEAVE_{osname.upper()}_{arch.upper()}'
                formula=formula.replace(key+'_URL@',f'https://github.com/xianxu/ariadne/releases/download/{tag}/{name}').replace(key+'_SHA256@',digest)
        if '@WEAVE_' in formula:
            raise ValueError('unresolved formula metadata')
        (artifacts/'SHA256SUMS').write_text(''.join(checksums))
        (artifacts/'weave.rb').write_text(formula)
        # A failed build cannot leave a directory that looks like a completed release.
        artifacts.rename(output)
except (OSError,ValueError,subprocess.CalledProcessError) as error:
    sys.exit(f'release preparation failed: {error}')
print(f'Prepared {tag}: {output}')
PY
