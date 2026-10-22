# Boundary Review — ariadne#545 (whole-issue close)

| field | value |
|-------|-------|
| issue | 545 — local-check: export the caller root and dependency checkout paths to suites (seeded workflow harness assumes a sibling ../ariadne) |
| repo | ariadne |
| issue file | workshop/issues/000545-local-check-export-the-caller-root-and-dependency-checkout-paths-to-suites-seeded-workflow-harness-assumes-a-sibling-ariadne.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6f22f46410834970c44578d49b910617416e8086..7b1c8d1892a3bf05379ca8aaa58cf0d422b334ff |
| command | sdlc close --issue 545 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T00:43:58-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range delivers the revised issue contract. Dependency declarations come from the tested tree, paths resolve against the initiating checkout, and environment preparation runs before managed command admission. No blocking defects or documentation gaps found.

```findings
{}
```

1. **Strengths**
   - Shared integration at `cmd/sdlc/localcheck_run.go:353` covers ordinary checks, runners, affected checks, and rerun batches.
   - Reuses layergraph’s bounded reader, parser, and substrate resolver.
   - Deterministic projection clears stale reserved variables and refuses conflicting paths (`cmd/sdlc/localcheck_dependency_environment.go:49`).
   - Detached wrapper fixtures exercise explicit selection, direct fallback, missing-path skip, and dependency preservation. Help, README, and atlas document the contract.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** All seven new tests, six adjacent affected-check controls, and the layergraph package passed. `git diff --check` passed; checkout remains clean. Full suite was not rerun during this review.

6. **Architecture:** ARCH-DRY passes through shared preparation and existing resolution. ARCH-PURE passes with deterministic projection separated from filesystem/Git reads. ARCH-PURPOSE passes against the recorded scope revision; Ariadne’s existing portable harness is self-rooted. ARCH-EVOLVE passes: no persisted format changes.

7. **Plan revisions:** None required.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"bbba570f-3021-45c6-b8ef-5583a1b03074","repository":"github.com/xianxu/ariadne","issue_id":"000545","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/9aaaabe007b40592","reviewed_head":"7b1c8d1892a3bf05379ca8aaa58cf0d422b334ff","slot":"ariadne:2","requested_at":"2026-10-22T07:42:26.900941Z"},"phase":"finished","pid":33448,"started_at":"2026-10-22T07:42:27.062267Z","finished_at":"2026-10-22T07:43:57.202542Z","session_id":"01a1628f-b489-7dc2-95ec-ce15e30c79e6","outcome":"returned"}] -->
