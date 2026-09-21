#!/usr/bin/env bash
# Build real release artifacts and execute the formula's native composition test.
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../.." && pwd)"
python3 - "$SOURCE" "${1:-}" "${2:-weave-v1.2.3}" <<'PY'
import hashlib,os,pathlib,platform,re,shutil,signal,subprocess,sys,tarfile,tempfile,time
source=pathlib.Path(sys.argv[1]); destination=sys.argv[2]; release_tag=sys.argv[3]; release_version=release_tag.removeprefix("weave-v")
with tempfile.TemporaryDirectory(prefix='weave-release-test.') as tmp:
    root=pathlib.Path(tmp).resolve(); release=pathlib.Path(destination).resolve() if destination else root/'release output'
    script=source/'scripts/release-weave.sh'
    def package(tag,out,env=None,ok=True):
        result=subprocess.run(['bash',str(script),tag,os.path.relpath(out,root)],cwd=root,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
        assert (result.returncode==0)==ok,(result.returncode,result.stdout)
        return result
    for tag in ['v1.2.3','weave-v1.2','weave-v01.2.3','weave-v1.2.3;touch bad','weave-v1.2.3\n']:
        result=package(tag,root/'invalid',ok=False)
        assert 'invalid release tag' in result.stdout,result.stdout
        assert not (root/'invalid').exists()
    # Real process failure through PATH must not expose completed artifacts.
    commands=root/'commands'; commands.mkdir(); fake=commands/'go'
    fake.write_text('#!/bin/sh\nif [ "$1" = -C ] || [ "$1" = run ]; then exec "$RELEASE_TEST_GO" "$@"; fi\necho intentional-build-failure >&2\nexit 17\n'); fake.chmod(0o755)
    result=package('weave-v1.2.3',root/'failed',dict(os.environ,PATH=str(commands)+':'+os.environ['PATH'],RELEASE_TEST_GO=shutil.which('go')),ok=False)
    assert 'intentional-build-failure' in result.stdout,result.stdout
    assert not (root/'failed').exists()
    # Fail the second target after the first archive has been prepared privately.
    fake.write_text('#!/bin/sh\nif [ "$1" = -C ] || [ "$1" = run ]; then exec "$RELEASE_TEST_GO" "$@"; fi\nif [ -f "$RELEASE_TEST_COUNT" ]; then echo intentional-second-build-failure >&2; exit 17; fi\ntouch "$RELEASE_TEST_COUNT"\nexec "$RELEASE_TEST_GO" "$@"\n')
    failed_env=dict(os.environ,PATH=str(commands)+':'+os.environ['PATH'],RELEASE_TEST_COUNT=str(root/'build-count'),RELEASE_TEST_GO=shutil.which('go'))
    result=package('weave-v1.2.3',root/'partial',failed_env,ok=False)
    assert 'intentional-second-build-failure' in result.stdout,result.stdout
    assert not (root/'partial').exists()
    # Exercise the public go-run launcher under process-group cancellation too.
    fake.write_text('#!/bin/sh\nif [ "$1" = -C ] || [ "$1" = run ]; then exec "$RELEASE_TEST_GO" "$@"; fi\ntouch "$RELEASE_TEST_STARTED"\nexec sleep 60\n')
    cancelled=root/'cancelled'; started=root/'cancel-started'
    cancel_env=dict(os.environ,PATH=str(commands)+':'+os.environ['PATH'],RELEASE_TEST_GO=shutil.which('go'),RELEASE_TEST_STARTED=str(started))
    with (root/'cancel.log').open('w') as log:
        process=subprocess.Popen(['bash',str(script),release_tag,str(cancelled)],cwd=root,env=cancel_env,stdout=log,stderr=log,start_new_session=True)
        try:
            deadline=time.monotonic()+20
            while not started.exists() and time.monotonic()<deadline: time.sleep(0.02)
            assert started.exists(), 'public launcher did not reach producer'
            os.killpg(process.pid,signal.SIGTERM); assert process.wait(timeout=10)!=0
            deadline=time.monotonic()+10
            while list(root.glob('.cancelled-weave-*')) and time.monotonic()<deadline: time.sleep(0.02)
            assert not list(root.glob('.cancelled-weave-*')), 'public cancellation leaked owned stage'
            assert not cancelled.exists()
        finally:
            if process.poll() is None: os.killpg(process.pid,signal.SIGKILL); process.wait()
    package(release_tag,release)
    assert {p.name for p in release.iterdir()}=={'SHA256SUMS','weave.rb'}|{f'weave_{release_version}_{o}_{a}.tar.gz' for o in ['darwin','linux'] for a in ['arm64','amd64']}
    sums={line.split()[1]:line.split()[0] for line in (release/'SHA256SUMS').read_text().splitlines()}
    assert len(sums)==4,sums
    formula=(release/'weave.rb').read_text()
    assert '@WEAVE_' not in formula
    assert 'license ' not in formula and 'depends_on ' not in formula
    for name,digest in sums.items():
        archive=release/name
        assert hashlib.sha256(archive.read_bytes()).hexdigest()==digest,name
        assert digest in formula and f'/{release_tag}/{name}' in formula,name
        with tarfile.open(archive) as tar:
            entries=tar.getmembers(); assert len(entries)==1 and entries[0].name=='weave',entries
            entry=entries[0]; assert entry.isfile() and entry.mode & 0o111==0o111
            if name==f'weave_{release_version}_{platform.system().lower()}_{dict(aarch64="arm64",arm64="arm64",x86_64="amd64")[platform.machine()]}.tar.gz':
                native=root/'native'; native.mkdir()
                executable=native/'weave'; executable.write_bytes(tar.extractfile(entry).read()); executable.chmod(entry.mode)
    version=subprocess.check_output([str(native/'weave'),'--version'],text=True).strip()
    assert version==f'weave version {release_version}',version
    # Build metadata verifies every binary's target and disabled CGO, even cross builds.
    for name in sums:
        with tarfile.open(release/name) as tar:
            binary=root/'inspect'; binary.write_bytes(tar.extractfile('weave').read())
        metadata=subprocess.check_output(['go','version','-m',str(binary)],text=True)
        osname,arch=name.removesuffix('.tar.gz').split('_')[-2:]
        assert f'GOOS={osname}' in metadata and f'GOARCH={arch}' in metadata and 'CGO_ENABLED=0' in metadata,metadata
    # Existing output is never overwritten, even by a failing retry.
    before={p.name:p.read_bytes() for p in release.iterdir()}
    package('weave-v1.2.4',release,ok=False)
    assert before=={p.name:p.read_bytes() for p in release.iterdir()}
    ruby=[shutil.which('ruby') or subprocess.check_output(['brew','ruby','-e','print RbConfig.ruby'],text=True).strip()]
    subprocess.run(ruby+['-c',str(release/'weave.rb')],check=True)
    harness=root/'formula-test.rb'; harness.write_text(r'''
require "pathname"
require "open3"
class Formula
  class << self
    attr_reader :test_block, :release_version
    def desc(*); end
    def homepage(*); end
    def url(*); end
    def sha256(*); end
    def version(value); @release_version=value; end
    def on_macos; yield if RUBY_PLATFORM.include?("darwin"); end
    def on_linux; yield if RUBY_PLATFORM.include?("linux"); end
    def on_arm; yield if RUBY_PLATFORM.match?(/arm64|aarch64/); end
    def on_intel; yield unless RUBY_PLATFORM.match?(/arm64|aarch64/); end
    def test(&block); @test_block=block; end
  end
  def initialize(binary, path); @bin=Pathname(binary); @path=Pathname(path); @path.mkpath; end
  def cd(path, &block); Dir.chdir(path, &block); end
  def bin; @bin; end
  def testpath; @path; end
  def version; self.class.release_version; end
  def assert_match(expected, actual); raise "#{expected.inspect} missing in #{actual.inspect}" unless actual.include?(expected.to_s); end
  def shell_output(command)
    out,status=Open3.capture2e(command); raise out unless status.success?; out
  end
  def system(*args); raise "failed: #{args.inspect}" unless Kernel.system(*args.map(&:to_s)); end
end
load ARGV[0]
formula=Weave.new(ARGV[1],ARGV[2])
formula.instance_eval(&Weave.test_block)
''')
    subprocess.run(ruby+[str(harness),str(release/'weave.rb'),str(native),str(root/'formula-test')],check=True)
    assert not list(release.parent.glob('.'+release.name+'-weave-*')),'leaked preparation stages'
    print('PASS release: four real archives, targets/CGO/version/layout/checksums, formula composition, failures')
PY
