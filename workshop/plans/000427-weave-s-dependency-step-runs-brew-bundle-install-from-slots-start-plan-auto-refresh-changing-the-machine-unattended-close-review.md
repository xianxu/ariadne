# Boundary Review — ariadne#427 (whole-issue close)

| field | value |
|-------|-------|
| issue | 427 — weave's dependency step runs brew bundle install from slots (start-plan auto-refresh), changing the machine unattended |
| repo | ariadne |
| issue file | workshop/issues/000427-weave-s-dependency-step-runs-brew-bundle-install-from-slots-start-plan-auto-refresh-changing-the-machine-unattended.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0408d048433e0073f5914787eb3f093480cfbf64..c1173b2725c429a2b1a59824eef06a05b99588e7 |
| command | sdlc close --issue 427 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T09:29:30-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned changes satisfy the issue’s dependency policy. Automated contexts check packages, hand runs retain installation, failures stop compilation, and brew calls suppress metadata updates. No blocking defects found.

```findings
{}
```

1. **Strengths**
   - Both weave dependency callers reuse `dependencyChecksOnly` and `startup.Dependencies`.
   - Both SDLC subprocess paths override the child context without changing the parent environment.
   - Stateful tests verify package state, metadata state, failure propagation, and executable recovery commands containing apostrophes.
   - Help, README, and atlas describe the changed policy.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Focused dependency tests passed with `TMPDIR=/tmp`; SDLC subprocess-context tests and workspace discovery tests also passed. Review-local temporary directories caused non-Git fixtures to encounter the enclosing checkout’s `.git`; external temporary directories resolved that environmental interference. The recorded full-suite result was not independently rerun; its referenced commit differs from the reviewed head only in issue evidence.

6. **Architecture:**
   - **ARCH-DRY — pass:** shared selection and execution cover dependencies, compile, and refresh; both SDLC paths share environment construction.
   - **ARCH-PURE — pass:** effects remain in the existing injected filesystem/runner seam, with small boundary policy selection.
   - **ARCH-PURPOSE — pass:** numbered-slot discovery, Couch context, and SDLC context all restrict package effects. Existing discovery covers private dependency checkouts, and positive hand controls preserve provisioning.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"5d9eefd6-cd09-4658-8740-f52db97c002e","repository":"github.com/xianxu/ariadne","issue_id":"000427","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/7a0c99d1f2236f0b","reviewed_head":"c1173b2725c429a2b1a59824eef06a05b99588e7","slot":"ariadne:5","requested_at":"2026-10-17T16:28:05.528898Z"},"phase":"finished","pid":46091,"started_at":"2026-10-17T16:28:05.631155Z","finished_at":"2026-10-17T16:29:30.472295Z","session_id":"01a14ab1-257f-7612-bd6f-e5a68f136d63","outcome":"returned"}] -->
