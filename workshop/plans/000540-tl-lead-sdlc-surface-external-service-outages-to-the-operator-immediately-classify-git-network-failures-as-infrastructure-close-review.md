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

---

## Re-review — 2026-10-21T21:03:56-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 540 — tl-lead + sdlc: surface external-service outages to the operator immediately; classify git network failures as infrastructure |
| repo | ariadne |
| issue file | workshop/issues/000540-tl-lead-sdlc-surface-external-service-outages-to-the-operator-immediately-classify-git-network-failures-as-infrastructure.md |
| boundary | whole-issue close |
| milestone | — |
| window | d2c7a015f631896fdff770f72ef0e89ff406aab8..7cf22ce3c771dd2cc80d40799468bb1c5f065f64 |
| command | sdlc close --issue 540 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T21:03:56-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-2 is addressed: cancellation now precedes classification, preserves process-error identity, and has regressions that fail against the pre-fix code. One separate correctness bug blocks shipment: macOS IPv6 DNS answers are discarded, producing misleading network observations.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      Production-boundary and outer-context regressions pass at head and fail with pre-fix runner code restored in a disposable checkout, reproducing retained network classification and lost process-error identity. The temporary checkout was removed.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: native-observation-format-completeness
    title: |
      macOS system DNS silently discards IPv6 answers
    detail: |
      cmd/sdlc/internal/netcheck/probe.go:87 recognizes ip_address: but omits dscacheutil's ipv6_address: field. A local dscacheutil lookup confirmed that format, and an isolated regression using "ipv6_address: ::1" fails with an empty result. For IPv6-only answers, check.go:155 marks successful DNS unknown and skips HTTPS, violating the diagnostic contract (ARCH-PURPOSE). Accept both native address fields and add parser plus production-adapter regressions covering IPv6-only and mixed answers.
```

1. **Strengths**

   - Context owners finalize cleanup and cancellation before classification; BR-2’s correction covers main and internal runners.
   - Publication wrappers preserve both infrastructure causes and uncertainty sentinels.
   - HTTPS retains certificate validation, disables redirects/proxies, and explicitly owns connection cleanup.
   - README, atlas, and rendered TL/worker instructions cover the new command and incident workflow.

2. **Critical findings**

   - `cmd/sdlc/internal/netcheck/probe.go:87`: handle macOS `ipv6_address:` records as described above. The current tests cover bare IPv6 output but miss the native macOS format.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Targeted Git/CLI network controls, the full netcheck package, TL/worker packages, race controls, vet, and diff whitespace checks passed. The isolated IPv6 regression failed as expected. BR-2 regressions failed against restored pre-fix code. The full repository suite and public-host conformance were not rerun during this review.

6. **Architectural notes**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | Pass — shared classifier and host validation. |
   | ARCH-PURE | Pass — classification remains free of probes; effects use injected boundaries. |
   | ARCH-PURPOSE | **Flag** — native macOS IPv6 observations are incomplete. |
   | ARCH-MOCK | Pass — stateful probe fake, private subprocess/TLS controls, and documented live conformance. |
   | ARCH-CONSTRAINTS | Pass — explicit deadlines, output caps, bounded jobs and address sampling. |
   | ARCH-SECURE | Pass — validated hosts/IPs, safe diagnostic rendering, verified TLS. |
   | ARCH-ORDER | Pass — final context outcome precedes classification; signal controls exercise cleanup. |
   | ARCH-FUNERAL | Pass — no new durable runtime artifacts; subprocesses and connections have owners. |
   | ARCH-EVOLVE | Pass — no persisted-format change. |

7. **Plan revision recommendations**

   Add a `## Revisions` entry recording native resolver-format coverage: both macOS address fields, IPv6-only and mixed answers, and regression evidence through the production lookup adapter.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"290e3cdb-0c93-4215-8094-ff5205553e18","repository":"github.com/xianxu/ariadne","issue_id":"000540","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/12083b1229c4995a","reviewed_head":"7cf22ce3c771dd2cc80d40799468bb1c5f065f64","slot":"ariadne:1","requested_at":"2026-10-22T04:01:19.270993Z"},"phase":"finished","pid":65431,"started_at":"2026-10-22T04:01:19.378846Z","finished_at":"2026-10-22T04:03:55.888904Z","session_id":"01a161c5-412e-78a0-9ac0-b185779411d8","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T21:11:33-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 540 — tl-lead + sdlc: surface external-service outages to the operator immediately; classify git network failures as infrastructure |
| repo | ariadne |
| issue file | workshop/issues/000540-tl-lead-sdlc-surface-external-service-outages-to-the-operator-immediately-classify-git-network-failures-as-infrastructure.md |
| boundary | whole-issue close |
| milestone | — |
| window | d2c7a015f631896fdff770f72ef0e89ff406aab8..f64e748c5510c35bd99d6b5ba1ef403117c17246 |
| command | sdlc close --issue 540 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T21:11:33-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range matches the revised Spec and Plan. BR-3 is fixed with meaningful parser and production-adapter regressions: both fail when the IPv6 condition is removed and pass on HEAD. No new blocking findings. The checkout remains unchanged.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      probe.go:87 accepts both native macOS address fields. TestParseAddressesNativeFormats and TestNativeDarwinFormatsEligibleForHTTPS cover IPv6-only and mixed answers; both failed under a temporary overlay removing the fix and pass on HEAD. A read-only native localhost lookup independently matched the recorded fixture.
```

1. **Strengths**

   - Shared classification remains pure and preserves original causes and publication uncertainty (`cmd/sdlc/internal/gitx/network.go:177`).
   - BR-3 tests exercise native lookup results through HTTPS eligibility, including malformed-field controls (`cmd/sdlc/internal/netcheck/native_format_test.go:15`, `:55`).
   - Probe work has explicit budgets, bounded output, verified TLS, and cancellation ownership.
   - README and both relevant atlas pages document the new command and incident workflow.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed targeted Git/CLI network tests, full netcheck race tests, targeted Git/CLI race tests, full TL/worker packages, and vet for touched packages. BR-3’s mutation control failed as expected. Core-concept entities exist at the documented locations; pure classification tests require no external effects.

   I did not rerun the full repository suite or public-host conformance. The fixture’s trailing blank line triggers `git diff --check`; it faithfully preserves native output.

6. **Architectural notes**

   - **ARCH-DRY — pass:** one classifier serves the shared Git boundaries.
   - **ARCH-PURE — pass:** classification/rendering remain separate from native effects.
   - **ARCH-PURPOSE — pass:** operator escalation, transport classification, and explicit diagnosis are delivered.
   - **ARCH-MOCK — pass:** injected probe tests, native fixtures, and private TLS/process controls exercise production seams.
   - **ARCH-CONSTRAINTS — pass:** deadlines, output caps, concurrency, and address sampling are bounded.
   - **ARCH-SECURE — pass:** validated destinations, fixed diagnostic reasons, certificate verification, and credential omission.
   - **ARCH-ORDER — pass:** final cancellation precedes classification; signal tests verify probe cleanup.
   - **ARCH-FUNERAL — pass:** no new durable runtime artifacts; connections and subprocesses have scoped cleanup.
   - **ARCH-EVOLVE — pass:** no persisted-format changes; error identities remain discoverable.

7. **Plan revision recommendations:** None. Existing revisions describe the implementation corrections accurately.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"15c23c50-1d3f-4a0e-860a-81d0234dcf9a","repository":"github.com/xianxu/ariadne","issue_id":"000540","recipe":"milestone-review","round":4,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/5cfc37e60d716dab","reviewed_head":"f64e748c5510c35bd99d6b5ba1ef403117c17246","slot":"ariadne:1","requested_at":"2026-10-22T04:09:27.641983Z"},"phase":"finished","pid":10223,"started_at":"2026-10-22T04:09:27.787648Z","finished_at":"2026-10-22T04:11:33.422458Z","session_id":"01a161cc-b55b-7271-b90c-12f7636505f3","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T21:55:10-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 540 — tl-lead + sdlc: surface external-service outages to the operator immediately; classify git network failures as infrastructure |
| repo | ariadne |
| issue file | workshop/issues/000540-tl-lead-sdlc-surface-external-service-outages-to-the-operator-immediately-classify-git-network-failures-as-infrastructure.md |
| boundary | whole-issue close |
| milestone | — |
| window | d2c7a015f631896fdff770f72ef0e89ff406aab8..cb46c379614209911fbebf8281d39297e415a295 |
| command | sdlc close --issue 540 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T21:55:10-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation matches the revised Spec and Plan. Git classification stays pure, diagnostic probes are bounded and explicitly invoked, and publication uncertainty remains intact. All three prior fixes have verified regression evidence. No new blocking findings emerged.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The real native-Git loopback-refusal regression passes at head and fails with the pre-fix classifier supplied through a temporary Go overlay.
  - id: BR-2
    disposition: addressed
    note: |
      Production cancellation regressions pass at head. Restoring the pre-fix bounded runner through an overlay reproduces the unwanted network classification after cancellation.
  - id: BR-3
    disposition: addressed
    note: |
      Captured macOS resolver output and IPv6-only adapter regressions pass at head; both fail with the pre-fix probe implementation.
```

1. **Strengths**

   - Shared classification preserves original error identities and publication sentinels across runner boundaries.
   - `cmd/sdlc/internal/netcheck/check.go:124` bounds concurrency, preserves independent observations, and joins effects before returning.
   - Private TLS and real CLI signal tests exercise certificate validation, cancellation, and descendant cleanup.
   - README, atlas, and rendered TL/worker instructions cover the new command and incident workflow.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed targeted Git/network/CLI regressions, full netcheck tests, TL/worker packages, Go and shell lint catalogs, touched-package vet, selected publication/lifetime controls, and race checks. Each prior correction was independently tested red using its pre-fix implementation. Temporary overlays were removed; the checkout remains unchanged. The complete repository suite and staged live conformance were not repeated during this review.

6. **Architectural notes**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | Pass — shared classifier and host validation serve the changed consumers. |
   | ARCH-PURE | Pass — classification and rendering remain separate from native effects. |
   | ARCH-PURPOSE | Pass — transport runners, diagnostic command, incident rules, and discovery docs are covered. |
   | ARCH-MOCK | Pass — injected probes, stateful publication schedules, and native fixtures exercise production seams. |
   | ARCH-CONSTRAINTS | Pass — time, output, concurrency, and address-sampling limits are enforced. |
   | ARCH-SECURE | Pass — validated hosts/IPs, verified TLS, fixed error classes, and credential-free probes. |
   | ARCH-ORDER | Pass — final cancellation precedes classification; uncertainty and process ownership are preserved. |
   | ARCH-FUNERAL | Pass — command-owned processes and connections terminate; no new durable runtime artifacts. |
   | ARCH-EVOLVE | Pass — no persisted-format changes; existing error identities remain discoverable. |

7. **Plan revision recommendations:** None. Existing revisions describe the final implementation accurately.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7bf90837-6062-44a1-a718-b7ac5415f4ca","repository":"github.com/xianxu/ariadne","issue_id":"000540","recipe":"milestone-review","round":5,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/5bbde5d87471ab78","reviewed_head":"cb46c379614209911fbebf8281d39297e415a295","slot":"ariadne:1","requested_at":"2026-10-22T04:52:54.31291Z"},"phase":"finished","pid":93573,"started_at":"2026-10-22T04:52:54.431609Z","finished_at":"2026-10-22T04:55:10.348736Z","session_id":"01a161f4-7b4a-7dd1-bf35-6e79a0e03d7c","outcome":"returned"}] -->
