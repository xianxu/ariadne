# Operating cancellation oracle capture (#567)

`failure.jsonl` is the original unmodified Go event excerpt from #563 recorded
suite attempt 2cb913014a018f68422fdf9ef52180cf, head
6a0b35faea2b351b5b1ec19a310f6ea1d19f1948, 2026-10-22 23:25 PDT.
`original.go.txt` preserves the affected test function and its imports from
main 28c40fd5dc3485dce3ec4936485238a56743f4ab; no fixture-specific rewrite.
It demonstrates the unsupported one-shot `ps stat=` completion oracle.

A fresh unchanged 30-run reproduction returned 29 passes and the same `?E`
rejection once (26.605s). Full raw repeat output stays in the Git common dir
at `sdlc/test-stats/evidence/000567/original-repeat.jsonl`.
The local Darwin ps(1) manual defines E as "The process is trying to exit."
Production cancellation already signals the owned process group. The fix
waits on that group's existing kernel terminal condition rather than parsing
or approving the observed state string. The catalog replays the old source;
the native Git fixture independently proves owned-descendant cancellation.

`leader-only-negative.log` retains the native independent control: production
kill was temporarily changed to leader-only, and only after publisher return
the test observation was shortened to one second. The fixed group oracle
refused the still-runnable group. Source was restored byte-for-byte afterward;
no production runner or shipped deadline change is part of this fix.

Corrected repetition: 30/30 passes in 22.099s; race-enabled cancellation and
lint controls repeated five times passed (tl-lead 10.935s, golint 1.205s).
These timings are observed host evidence, not runtime budgets.
