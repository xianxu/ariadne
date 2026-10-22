# Boundary Review — ariadne#540 (whole-issue close)

| field | value |
|-------|-------|
| issue | 540 — tl-lead + sdlc: surface external-service outages to the operator immediately; classify git network failures as infrastructure |
| repo | ariadne |
| issue file | workshop/issues/000540-tl-lead-sdlc-surface-external-service-outages-to-the-operator-immediately-classify-git-network-failures-as-infrastructure.md |
| boundary | whole-issue close |
| milestone | — |
| window | d2c7a015f631896fdff770f72ef0e89ff406aab8..4de897fc68ac1f36b7a119c8ad92d16a0c34c95c |
| command | sdlc close --issue 540 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T20:41:50-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The shared classifier, bounded diagnostic command, incident instructions, and documentation are largely delivered and tested. One transport coverage gap blocks shipping: native `git://` connection failures remain ordinary errors despite the issue’s infrastructure-classification contract.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: transport-classification-completeness
    title: |
      Native Git transport failures bypass infrastructure classification
    detail: |
      cmd/sdlc/internal/gitx/network.go:158 recognizes SSH and HTTP diagnostics only, although RemoteNetworkHost explicitly accepts git:// URLs at line 91. Running `git ls-remote git://127.0.0.1:1/no-repo` produced `fatal: unable to connect to 127.0.0.1:` followed by `127.0.0.1[0: 127.0.0.1]: errno=Connection refused`; neither line matches, so ClassifyNetworkFailure returns the original error without the infrastructure label, safe host, or diagnostic command. ARCH-PURPOSE: enumerate supported Git transports and cover their resolver/connect/timeout diagnostics through the shared classifier, with regression tests at the production runner boundary.
```

1. **Strengths**

   - Classification remains pure, shared across execution boundaries, and preserves underlying error identity.
   - Publication regressions verify both network causes and `ErrPublicationUncertain` survive wrapping.
   - Native probe tests exercise certificate rejection, connection cleanup, subprocess cancellation, and real CLI SIGINT/SIGTERM.
   - README, atlas, and rendered TL/worker instructions cover the new command and incident lifecycle.

2. **Critical findings**

   - `cmd/sdlc/internal/gitx/network.go:158`: the native Git transport diagnostic above falls through to the unchanged cause at line 185. Add narrowly scoped native transport recognition and regression fixtures covering DNS, refused connections, and timeouts.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed focused Git/network/CLI tests, complete TL/worker package tests, netcheck race tests, and vet for both internal packages. The explicit read-only github.com check also passed: both DNS paths, both HTTPS observations, and API-host DNS succeeded.

   Existing transport fixtures cover SSH and HTTP; they omit the reproduced native Git diagnostic. No full repository suite was run during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** one shared classifier.
   - **ARCH-PURE — pass:** classification and host validation have direct tests; probes use an injected boundary.
   - **ARCH-PURPOSE — flag:** supported native Git transport remains outside classification.
   - **ARCH-MOCK — pass:** injected probe fake, native fixtures, and live conformance exercise the production seam.
   - **ARCH-CONSTRAINTS — pass:** bounded jobs, deadlines, output, and address sampling.
   - **ARCH-SECURE — pass:** validated destinations, verified TLS, credential-free probes, and fixed error reasons.
   - **ARCH-ORDER — pass:** cancellation joins probe work; publication uncertainty remains discoverable.
   - **ARCH-FUNERAL — pass:** diagnostic resources have explicit cleanup; no new durable runtime artifacts.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.

7. **Plan revision recommendations**

   Add a `## Revisions` entry enumerating supported transport families—HTTP(S), SSH/scp, and native Git—and requiring captured resolver/connect/timeout diagnostics plus production-runner regressions for each.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b5aff8be-e40c-4d94-930b-b612935c3e78","repository":"github.com/xianxu/ariadne","issue_id":"000540","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/742526998fb5708b","reviewed_head":"4de897fc68ac1f36b7a119c8ad92d16a0c34c95c","slot":"ariadne:1","requested_at":"2026-10-22T03:39:13.796853Z"},"phase":"finished","pid":20156,"started_at":"2026-10-22T03:39:13.996669Z","finished_at":"2026-10-22T03:41:50.003412Z","session_id":"01a161b1-086c-73f2-8eb3-563bfd3f2ac8","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T20:51:30-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 540 — tl-lead + sdlc: surface external-service outages to the operator immediately; classify git network failures as infrastructure |
| repo | ariadne |
| issue file | workshop/issues/000540-tl-lead-sdlc-surface-external-service-outages-to-the-operator-immediately-classify-git-network-failures-as-infrastructure.md |
| boundary | whole-issue close |
| milestone | — |
| window | d2c7a015f631896fdff770f72ef0e89ff406aab8..0af0c6e85aa7d9eb00bb80666acae259ab918cc7 |
| command | sdlc close --issue 540 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T20:51:30-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-1 is addressed: native Git failures now classify correctly, and its regression tests fail with that recognition disabled. The diagnostic command, incident rules, and documentation are substantially complete. One cancellation ordering defect still violates the explicit contract that cancellation remains unclassified.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Native resolver/connect handling is reachable through production runners. The real loopback refusal and injected DNS/timeout regressions pass on HEAD and fail with native recognition disabled through a temporary Go overlay.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: cancellation-before-classification
    title: |
      Bounded Git cancellation retains an infrastructure classification
    detail: |
      cmd/sdlc/runner.go:57 classifies the process error before cmd/sdlc/landinggit.go:80 joins the cancelled context. A controlled Git subprocess emitting an SSH resolver diagnostic, then waiting until cancellation, returns both context.Canceled and *GitNetworkFailure. This violates the issue's cancellation exclusion. ARCH-ORDER: incorporate the final context outcome before classification across contextual runners, preserving the original cause; add production-boundary cancellation regressions.
```

### 1. Strengths

- One pure classifier serves the shared Git execution boundaries.
- Publication wrappers preserve both network causes and publication uncertainty; regression tests check their identities.
- Net-check uses independently resolved addresses, verified TLS, bounded subprocesses, and explicit partial observations.
- README and atlas document the new command and incident workflow.

### 2. Critical findings

**Cancellation is classified before its context outcome is known** — `cmd/sdlc/runner.go:57`, consumed by `cmd/sdlc/landinggit.go:76–81`.

I reproduced this with a temporary test overlay: a Git fixture emits a recognized resolver diagnostic, signals readiness, and waits. Cancelling the bounded runner produces:

```text
context canceled
infrastructure: network (host git.example); diagnose with sdlc net check --host git.example
```

Determine cancellation before applying the classifier, as the internal contextual Git boundary already does. Sweep contextual runner entry points and retain a regression asserting cancellation identity without a network classification.

### 3. Important findings

None.

### 4. Minor findings

None.

### 5. Test coverage notes

Passed targeted Git/classifier/CLI controls, the complete netcheck package, and TL/worker packages. These cover publication identity, preparation failures producing no test evidence, private TLS, and CLI signal cleanup.

BR-1’s production and injected regressions failed with native recognition disabled. The added review-only cancellation regression failed on HEAD. Temporary overlays were removed; the checkout remains unchanged. I did not run the full repository suite or repeat the external live check.

### 6. Architectural notes

| Principle | Result |
|---|---|
| ARCH-DRY | Pass — shared classification and host validation. |
| ARCH-PURE | Pass — classification/rendering separated from native effects; core concept locations match. |
| ARCH-PURPOSE | Pass — native transport coverage, incident rules, and diagnostic discovery delivered. |
| ARCH-MOCK | Pass — injected probe seam, stateful failure/recovery tests, private native controls. |
| ARCH-CONSTRAINTS | Pass — bounded jobs, output, subprocesses, and HTTPS attempts. |
| ARCH-SECURE | Pass — validated hosts/addresses, credential-safe diagnostics, verified TLS. |
| ARCH-ORDER | **Flag** — cancellation arrives after classification at the bounded main runner. |
| ARCH-FUNERAL | Pass — command-scoped resources and cleanup controls; no new runtime artifact family. |
| ARCH-EVOLVE | Pass — no persisted format change; error identities retained. |

### 7. Plan revision recommendations

Append a `## Revisions` entry requiring final context outcomes to precede classification at every contextual Git boundary, with subprocess regressions for cancellation after diagnostic emission.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"fc0d1bec-58e6-4f50-ae38-fe436b6d88f6","repository":"github.com/xianxu/ariadne","issue_id":"000540","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/2175a223c1c41d32","reviewed_head":"0af0c6e85aa7d9eb00bb80666acae259ab918cc7","slot":"ariadne:1","requested_at":"2026-10-22T03:48:51.279651Z"},"phase":"finished","pid":66875,"started_at":"2026-10-22T03:48:51.501978Z","finished_at":"2026-10-22T03:51:29.52473Z","session_id":"01a161b9-d8d7-7a91-9ced-cdd811381395","outcome":"returned"}] -->
