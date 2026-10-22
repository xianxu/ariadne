# #555 retained incident evidence

Captured read-only from owner-named sources on 2026-10-22. Timestamps below establish sequence; they do not reconstruct missing lock acquisition/release events.

## Pair #803 suite receipt

Source: pair common Git directory, sdlc/test-stats/attempts/f5e320381ff102cd8180f600780f0ccc.json. Selected fields copied without changing values:

```json
{
  "id": "f5e320381ff102cd8180f600780f0ccc",
  "repo": "xianxu/pair",
  "sha": "37007dc788c16cc63f47bfd8f3a43d646cbf9170",
  "check": "suite",
  "slot": "pair:6",
  "queued_at": "2026-10-22T17:42:17.469306Z",
  "acquired_at": "2026-10-22T17:51:21.387607Z",
  "finished_at": "2026-10-22T18:03:30.681045Z",
  "wall_seconds": 729.266438,
  "exit": 0,
  "timed_out": false,
  "log_sha256": "7e7f7430b18768e8787d24cf72d06c757995befe690327bd3c611453c4eeb4a2",
  "outcome": "passed"
}
```

## Exact stage diagnostic and observed bounds

Pair:9 supplied this retained session96254 output:

```text
waiting for sdlc repo lock held by pid 17156: sdlc merge: on Xians-MacBook-Pro.local (long-running review/ship transaction)
```

The clear followed authorization at 11:07:33.679 PDT; its successful completion report was accepted11:11:04.226. At about11:09, ps reported pid17156 PPID46243 elapsed29:32. This is process age, not lock-held time. The stage wait occurred after the suite receipt's 18:03:30Z /11:03:30 PDT finish.

## Landing identity

Read from pair .git/sdlc/landing/events.jsonl lines852–856. Wait identity binds issue803 / seq274, pid17156, pid_started1792690767.793984, pair:6, operation op-6a1bb7bf0f02b3eaa656cb5dcdaffd05, source head37007dc788c16cc63f47bfd8f3a43d646cbf9170. Enqueued10:39:29.733512 PDT; admitted10:39:29.800737.

## Owner-reported buffered output

Pair:6 retained tool receipts show CI waiting by11:07:09; by11:08:12 CI4green followed by a repo-lock wait on issue set-status pid39100; by11:09:02 a wait on stage pid54025. These tool receipt times are upper bounds for buffered output, not lock transitions. Its merge61b60a8f has committer time11:08:20 PDT; landing event11:10:12.643; final successful output by11:10:45. No continuous lock hold across suite admission/execution follows from those records.

## Evidence limits

No exact launch build revision or lock-transition timestamps were retained. The full executable path and prior10:27 clean binary observation do not establish the10:39 runner's revision. Do not substitute the current binary version or infer a stale executable. TL accepted the corrected interpretation: short legitimate post-suite transaction; retain scoped locking and pin command-level regressions.

## Command-level oracle mutation

The temporary mutation set the test root's actual merge command to automatic whole-command locking and re-applied the root wrapper; no production file changed. Targeted command:

```text
go test ./cmd/sdlc -run '^TestMergeCommandStageDuringSuiteWaits/admission/clear$' -count=1 -timeout=45s
```

Observed failure excerpt (mutation restored afterward):

```text
--- FAIL: TestMergeCommandStageDuringSuiteWaits (26.30s)
    --- FAIL: TestMergeCommandStageDuringSuiteWaits/admission/clear (26.30s)
        merge_stage_lock_test.go:190: waiting for stage clear during admission exhausted 18.852276646s; goroutine stacks:
FAIL
FAIL	github.com/xianxu/ariadne/cmd/sdlc	26.862s
FAIL
```

The independent stage-completion condition failed while merge admission remained held, establishing that this command-level regression detects whole-command locking. Final passing verification is recorded in the issue Log.
