---
id: 000552
status: open
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
github_issue:
---

# flake: TestDocRepositoryCancellation

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/tl-lead.TestDocRepositoryCancellation: Recorded suite attempt 4638dd482de33746116a1788f13ba510 failed TestDocRepositoryCancellation on exact head 0ae0016dbbe6634bfbf59132045e683bfa76862a: strconv.Atoi parsing empty readiness content. Afterward, a clean detached worktree at that exact Git-verified head ran the whole top-level test: go test ./cmd/tl-lead -run ^TestDocRepositoryCancellation$ -count=1 -v; PASS, test 0.23s, command 0.390s. Worktree status clean. Source polls file existence before reading an echo-created PID file, so creation can be observed before PID content. The suite also had a separate reproducible context-guard failure, being fixed on #551; this registration covers only the independent tl-lead fixture.
