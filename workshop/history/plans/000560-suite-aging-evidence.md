# #560 suite-aging conformance

Captured 2026-10-22 PDT; trailing whitespace normalized. This is host evidence for the approved read-only live queue observation and disposable private queue cohort. No live grants or tickets were edited, no running suite was interrupted, and no stage/cutover was initiated. Existing helpers continue their original algorithm until next start; unknown legacy ages in this observation are conservatively projected as protected.

Source pins (HEAD, then helper/observer/queuecmd blob IDs):

```text
d7136fdce2100bff36eb271c9a4f00bd455c6232
fde8d8d61320819ba8a158b55742b85d7096f1d3
41bcf50727b1702e29b380f7d9f5d57a35a413b1
f5bd1976aae64ff53a4558a5138f57c67e161b79
```

Clean build (`make sdlc-build`, then `go version -m bin/sdlc`):

```text
bin/sdlc: go1.27.1
	path	github.com/xianxu/ariadne/cmd/sdlc
	mod	github.com/xianxu/ariadne	v0.0.0-20261022223257-d7136fdce210
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
	build	vcs.revision=d7136fdce2100bff36eb271c9a4f00bd455c6232
	build	vcs.time=2026-10-22T22:32:57Z
	build	vcs.modified=false
```

Read-only branch-binary `bin/sdlc queue status`, exit 0 (warning is expected for the explicitly approved branch-binary conformance):

```text
WARNING: sdlc is not the fleet :0 binary: path=/Users/xianxu/workspace/worktree/ariadne-slot1/ariadne/bin/sdlc revision=d7136fdce2100bff36eb271c9a4f00bd455c6232 modified=false; expected=/Users/xianxu/workspace/ariadne/bin/sdlc. Refresh aliases or relaunch.
queue: running; complete=true
  suite op-a8322735be465367ca1ac212341902b3: running (parley.nvim:1)
  merge op-a0c6739d543d454d59aaeca45d571991: running (parley.nvim:1)
  landing op-a0c6739d543d454d59aaeca45d571991: running (parley.nvim:1)
  merge op-26cc562a45ba687042a9e890b08fcbaf: running (ariadne:2)
  landing op-26cc562a45ba687042a9e890b08fcbaf: running (ariadne:2)
  suite op-8eb8220f8db54ece5403db9f021e02ae: running (pair:4)
suite queue: GRANT_MAX_WAIT_SECONDS=3600 (60 min grant overtaking limit); protected through ticket 130
  ticket 99: position 1; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot5/pair"
  ticket 103: position 2; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "parley.nvim", worktree "/Users/xianxu/workspace/worktree/parley.nvim-slot2/parley.nvim"
  ticket 105: position 3; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot6/pair"
  ticket 109: position 4; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot1/pair"
  ticket 110: position 5; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "parley.nvim", worktree "/Users/xianxu/workspace/worktree/parley.nvim-slot4/parley.nvim"
  ticket 111: position 6; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot3/pair"
  ticket 112: position 7; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot4/pair"
  ticket 113: position 8; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot3/pair"
  ticket 117: position 9; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot8/pair"
  ticket 123: position 10; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "ariadne", worktree "/Users/xianxu/workspace/worktree/ariadne-slot5/ariadne"
  ticket 124: position 11; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot9/pair"
  ticket 125: position 12; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "ariadne", worktree "/Users/xianxu/workspace/worktree/ariadne-slot3/ariadne"
  ticket 126: position 13; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "ariadne", worktree "/Users/xianxu/workspace/worktree/ariadne-slot2/ariadne"
  ticket 128: position 14; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "echo", worktree "/Users/xianxu/workspace/echo"
  ticket 129: position 15; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "pair", worktree "/Users/xianxu/workspace/worktree/pair-slot7/pair"
  ticket 130: position 16; age unknown; "unknown_age_fifo"; protected=true; granted=false; repo "ops", worktree "/Users/xianxu/workspace/worktree/ops-slot1/ops"
```

Private real helper/process/flock cohort with the same clean binary as SUITE_PRIORITY_TEST_PROVIDER:

```text
env SUITE_PRIORITY_TEST_PROVIDER=/Users/xianxu/workspace/worktree/ariadne-slot1/ariadne/bin/sdlc python3 scripts/test/test-lock-priority.test.py SuiteQueuePriority.test_aged_weighted_waiter_survives_multiple_renewed_train_grants SuiteQueuePriority.test_acquire_watcher_retains_immutable_enqueue_and_unknown_metadata -v
```

Final completion chunk (exit 0):

```text
ok
test_acquire_watcher_retains_immutable_enqueue_and_unknown_metadata (__main__.SuiteQueuePriority.test_acquire_watcher_retains_immutable_enqueue_and_unknown_metadata) ... ok

----------------------------------------------------------------------
Ran 2 tests in 8.314s

OK
```

The first case independently observes weighted slot acquisition: an aged ordinary heavy head precedes renewed young issue/train grants without partial slots or preempting existing holders. It compares read-only inspection to that effective order. The second observes final watcher-owned slot bytes, including exact enqueue age and unknown nested metadata. These cases use private queue roots and flock-owner subprocesses; their actual assertion code is in the pinned source commit.
