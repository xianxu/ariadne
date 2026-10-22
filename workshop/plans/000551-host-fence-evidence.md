# Host OS-fence conformance — ariadne#551

Captured at 2026-10-22 13:16:38 UTC from ariadne:1, outside the reviewer's nested sandbox. Platform: macOS (Seatbelt).

Command:

```sh
env WF_LIVE_REVIEW_DEPENDENCIES=1 go test ./cmd/sdlc -run '^TestReviewDependencyLiveFence$' -count=1 -v
```

Exit status: 0. Raw combined command output:

```text
=== RUN   TestReviewDependencyLiveFence
--- PASS: TestReviewDependencyLiveFence (0.91s)
PASS
ok  	github.com/xianxu/ariadne/cmd/sdlc	1.391s
```

Tested HEAD and immutable source blob identities:

```text
80e095147c447abf1a3c41799029a050eca5817f
100644 blob 4c225de22dd36aab658acd2e268d1c3b3c1650eb	cmd/sdlc/internal/gitx/boundedoutput.go
100644 blob aa244082bd7f0f588afbd7de90041eecf0db26b2	cmd/sdlc/internal/reviewbox/checkout.go
100644 blob 80de2b7bb9c674056ba7cea9ec2d8eeb4c4e6d61	cmd/sdlc/internal/reviewbox/fence.go
100644 blob db81ace7cc0bd1585020b500d95f6773654a7a19	cmd/sdlc/reviewdependencies.go
100644 blob c84690442c1b605d2ad3d3441e679af2453f04dc	cmd/sdlc/reviewdependencies_live_test.go
100644 blob 3e4e9797b54b977af10faa785a1971999233e731	cmd/sdlc/reviewisolation.go
```

At execution, the only tracked working-tree changes were the generated close-gate and close-review Markdown sidecars. Executable source matched this HEAD. Subsequent evidence/documentation commits retain these source blob identities; compare with Git when reviewing.

The production openReviewIsolation path creates a disposable consumer and dependency. The test verifies reading the pinned fixture, writing the primary checkout, and rejection of writes in the dependency clone, source checkout and source Git common directory, including ref updates. Its assertions run against the actual platform fence. All repositories and attempted writes are disposable fixtures; no live peer checkout is used.

