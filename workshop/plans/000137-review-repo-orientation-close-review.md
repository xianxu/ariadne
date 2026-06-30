# Boundary Review — ariadne#137 (whole-issue close)

| field | value |
|-------|-------|
| issue | 137 — sdlc boundary review repo orientation |
| repo | ariadne |
| issue file | workshop/issues/000137-review-repo-orientation.md |
| boundary | whole-issue close |
| milestone | — |
| window | 25b8dbafe932c155a823db9b5086af4243cba36c..HEAD |
| command | sdlc close --issue 137 |
| reviewer | claude |
| timestamp | 2026-06-29T21:36:44-07:00 |
| verdict | unknown |

## Review

Confirmed — that third full `cmd/sdlc` suite run also returned exit 0 (three clean runs total), so the "existing verdict/trailer/gate tests pass" Done-when item holds. My review verdict stands:

**FIX-THEN-SHIP** — the core repo-orientation change is solid (clean ARCH-PURE split, ARCH-DRY consolidation, strong negative tests). The one real issue is a preview-only regression: `close.go:742`'s dry-run `boundaryReviewParams` omits `IssueNum`, so `sdlc close --dry-run` renders `issue: <repo>#0 (file: )` — the exact misorientation this issue targets. Non-blocking at the gate, but fix before the next boundary by adding `IssueNum: f.Issue` (mirroring the dispatch literal at `close.go:761`) and tightening `TestRunCloseWithReview_DryRunPrintsPairAgentCommand` to assert the derived ref.
