# #560 supervised inspection conformance

2026-10-22 PDT. Supersedes the earlier capture for the final runtime source after BR-1. Trailing whitespace normalized. No live queue/grant mutation or staged cutover occurred.

Clean build metadata and Git pins (HEAD, helper blob, observer blob, queuecmd blob):

```text
bin/sdlc: go1.27.1
	path	github.com/xianxu/ariadne/cmd/sdlc
	mod	github.com/xianxu/ariadne	v0.0.0-20261022224353-c66fa30e12d1
	dep	github.com/spf13/cobra	v1.10.2	h1:DMTTonx5m65Ic0GOoRY2c16WCbHxOOw6xxezuLaBpcU=
	dep	github.com/spf13/pflag	v1.0.9	h1:9exaQaMOCwffKiiiYk6/BndUBv+iRViNW+4lEMi0PvY=
	dep	github.com/yuin/goldmark	v1.7.13	h1:GPddIs617DnBLFFVJFgpo1aBfe/4xcvMc3SB5t/D0pA=
	dep	go.yaml.in/yaml/v3	v3.0.4	h1:tfq32ie2Jv2UxXFdLJdh3jXuOzWiL1fo0bu/FbuKpbc=
	dep	golang.org/x/sys	v0.47.0	h1:o7XGOvZQCADBQQ4Y7VNq2dRWQR7JmOUW8Kxx4ZsNgWs=
	build	-buildmode=exe
	build	-compiler=gc
	build	-trimpath=true
	build	DefaultGODEBUG=tracebacklabels=0,x509sslcertoverrideplatform=0
	build	CGO_ENABLED=1
	build	GOARCH=arm64
	build	GOOS=darwin
	build	GOARM64=v8.0
	build	vcs=git
	build	vcs.revision=c66fa30e12d1d12dd3897446f20266f71d444322
	build	vcs.time=2026-10-22T22:43:53Z
	build	vcs.modified=false
c66fa30e12d1d12dd3897446f20266f71d444322
4223c92f23b7cf5a5d1a290d9de6077d7086a4ee
967a0b225ebd6ca38924d961b3a4bd2f283279ee
f5bd1976aae64ff53a4558a5138f57c67e161b79
```

Read-only `bin/sdlc queue status`, exit 0:

```text
WARNING: sdlc is not the fleet :0 binary: path=/Users/xianxu/workspace/worktree/ariadne-slot1/ariadne/bin/sdlc revision=c66fa30e12d1d12dd3897446f20266f71d444322 modified=false; expected=/Users/xianxu/workspace/ariadne/bin/sdlc. Refresh aliases or relaunch.
queue: running; complete=true
  suite op-004a63c6dc6f72661757290ffa08437f: running (pair:4)
  merge op-072c4d23d7d638686f8517e8576b9b7a: running (parley.nvim:3)
  landing op-072c4d23d7d638686f8517e8576b9b7a: running (parley.nvim:3)
  merge op-26cc562a45ba687042a9e890b08fcbaf: running (ariadne:2)
  landing op-26cc562a45ba687042a9e890b08fcbaf: running (ariadne:2)
  suite op-8934ea467e075d1741062f6487c741b9: running (pair:5)
suite queue: GRANT_MAX_WAIT_SECONDS=3600 (60 min grant overtaking limit); protected through ticket 131
  ticket 103: position 1; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "parley.nvim", worktree "/Users/xianxu/workspace/worktree/parley.nvim-slot2/parley.nvim"
  ticket 105: position 2; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot6/pair"
  ticket 109: position 3; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot1/pair"
  ticket 110: position 4; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "parley.nvim", worktree "/Users/xianxu/workspace/worktree/parley.nvim-slot4/parley.nvim"
  ticket 111: position 5; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot3/pair"
  ticket 112: position 6; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot4/pair"
  ticket 113: position 7; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot3/pair"
  ticket 117: position 8; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot8/pair"
  ticket 123: position 9; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "ariadne", worktree "/Users/xianxu/workspace/worktree/ariadne-slot5/ariadne"
  ticket 124: position 10; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot9/pair"
  ticket 125: position 11; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "ariadne", worktree "/Users/xianxu/workspace/worktree/ariadne-slot3/ariadne"
  ticket 126: position 12; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "ariadne", worktree "/Users/xianxu/workspace/worktree/ariadne-slot2/ariadne"
  ticket 128: position 13; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "echo", worktree "/Users/xianxu/workspace/echo"
  ticket 129: position 14; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot7/pair"
  ticket 130: position 15; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "ops", worktree "/Users/xianxu/workspace/worktree/ops-slot1/ops"
  ticket 131: position 16; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "parley.nvim", worktree "/Users/xianxu/workspace/worktree/parley.nvim-slot3/parley.nvim"
```

The same binary was used as SUITE_PRIORITY_TEST_PROVIDER for the private real-helper weighted/renewed-grant cohort and watcher metadata regression. Final output chunk, exit 0:

```text
ok
test_acquire_watcher_retains_immutable_enqueue_and_unknown_metadata (__main__.SuiteQueuePriority.test_acquire_watcher_retains_immutable_enqueue_and_unknown_metadata) ... ok

----------------------------------------------------------------------
Ran 2 tests in 9.283s

OK
```

BR-1 deterministic real-inspector/provider-tree regression: cancellation and deadline both left provider+descendant alive before the supervision fix (red test exit1,7.063s). The fixed actual-inspector regression covers cancel, deadline, successful completion and provider timeout in the broader Go queue/suite/helper run (exit0,77.260s). Python23 ownership/core/resource/evidence tests passed0.689s; pause18 passed4.222s; golint passed1.694s. This deterministic production-path fixture is the cancellation proof; no live provider was disrupted.
