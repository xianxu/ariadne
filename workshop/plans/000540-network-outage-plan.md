---
collaboration-mode: discussion
---

# External dependency outage reporting and network diagnosis

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Tell the operator promptly when external infrastructure blocks the fleet, and distinguish Git transport failures from code/test failures with a bounded diagnostic command.

**Architecture:** One pure Git transport classifier wraps existing execution errors without probing or retrying. A read-only `sdlc net check` owns all active DNS/HTTPS observations. TL and worker help define incident escalation and recovery using the existing status file and Couch notification protocol.

**Tech Stack:** Go, Cobra, existing Git execution/process-group seams, OS resolver utilities, dig, Go HTTPS transport, embedded Markdown help.

## Approved scope and wording

TL approval: Couch delivery `047b5c29-ca18-4ce0-9dba-841fbfc49d8a` on 2026-10-21. The Git error path stays pure and immediate; probing exists only in the explicit command. This supersedes the original Spec sentence requesting API resolution inside each Git error.

Proposed addition to `cmd/tl-lead/help.md`, TALKING TO THE OPERATOR:

> The moment you learn that an external dependency is unavailable (Git host/DNS, GitHub API, model API, or tunnel), put the incident first in Needs you and send one couch notify immediately; do not wait for root-cause diagnosis. Name the service/host, affected work, and whether evidence points to a local, upstream, or still unknown cause. Say what the operator can do now, or explicitly that this is awareness-only. When the known remedy requires the operator, include the exact command for them to run. Update the same incident rather than repeating notifications for each blocked slot. On verified recovery, send a recovery notification, remove the active incident, and check/resume the affected slots through their existing recovery paths.

Adjust the existing Needs-you section's “only asks/no status” rule to admit these immediate incident notices explicitly. Add the worker side to `cmd/tl-worker/help.md`: report external dependency failures to the hub immediately as incidents, with observed evidence and affected operation, not as a code/test failure; preserve the runner/recovery state and do not blindly retry uncertain writes. These are shared agent instructions for Claude, Codex and other supported workers, not harness-specific adapters.

## Core concepts

### PURE

| Name | Lives in | Status |
|------|----------|--------|
| GitNetworkFailure | `cmd/sdlc/internal/gitx/network.go` | new |
| NetworkHost validation / remote extraction | `cmd/sdlc/internal/gitx/network.go` | new |
| Report | `cmd/sdlc/internal/netcheck/check.go` | new |

`GitNetworkFailure` classifies only transport-specific resolver, refused-connection and connect/transport-timeout diagnostics from fetch/push/pull/clone/ls-remote. Its Error starts `infrastructure: network`, names a validated host when observed, and gives the exact `sdlc net check --host HOST` command (plain `sdlc net check` when host unknown). `Unwrap` retains the original exit/context cause; repeated wrapping is idempotent. A cancellation or bare process deadline is not evidence of a network fault. Auth, ref/lease rejections, hooks, local IO and local-command timeout prose remain ordinary errors. It never asserts that an uncertain push failed remotely or prescribes an automatic retry.

`NormalizeNetworkHost` and `RemoteNetworkHost` parse only a host from explicit input, HTTPS/SSH/scp-style remote URLs or recognized transport diagnostics. URL credentials, path, query and fragment never enter new output. Handle IPv4, IPv6 and SSH aliases explicitly; local paths and helper transports cannot manufacture a host. A proxy resolver failure identifies a proxy observation rather than pretending it proves the Git origin unavailable.

`Report` renders per-path observations, including unknown/unavailable/timeout, without collapsing them into “upstream down.” System failure with direct success is evidence of a resolver-path difference, not proof of a particular cache fault. HTTP responses, including 401/403/5xx, prove HTTPS reachability only and retain the status code.

### INTEGRATION

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| Git error adapters | `cmd/sdlc/runner.go`, `internal/gitx/lifetime.go`, `internal/gitx/trunkfile.go`, `weeklystats_inputs.go` | modified | existing Git process seams |
| NetProbe | `cmd/sdlc/internal/netcheck/probe.go` | new | OS resolver, direct DNS, HTTPS |
| Net command | `cmd/sdlc/net.go` | new | Cobra, local Git configuration, NetProbe |

Use the same pure classifier at all four existing Git execution boundaries. `runGitCmd` also covers bounded durable-landing Git. Internal bounded-input calls share `executeGitInputContext`; capture diagnostics there without contaminating stdout, and avoid classifying twice. Audit caller wrappers that use `%v` so publication sentinels and the network cause both remain discoverable with `errors.Is/As`. Do not redesign publication uncertainty or outcome ledgers.

The command accepts `--host HOST` to work outside a repository and diagnose the exact host printed in a failure. Otherwise resolve the configured publication remote locally (explicit `--remote NAME` overrides it), and its effective fetch URL using Git's get-url semantics. If fetch and push hosts differ, name that fact and the explicit host command for the other host; do not silently claim to have checked both. Reject local-only/ambiguous remote configuration with an actionable `--host` hint. For github.com also report api.github.com DNS; other hosts have no guessed API endpoint (an explicit `--api-host` may name one).

System DNS uses the native resolver utility (`dscacheutil` on macOS, `getent ahosts` on Linux). Direct DNS uses `dig` against its configured default DNS server, bypassing the OS resolver cache, with A and AAAA queries. Do not invent a public DNS server, install tools, alter caches, or use another service as a DNS oracle. Missing tools or unsupported platforms produce explicit unavailable observations. Execute through a context-bounded, output-capped subprocess seam using existing process-group ownership. Parse only validated IP addresses; preserve no arbitrary tool output in the report.

For the Git host, attempt HTTPS using addresses independently obtained on each resolver path. Preserve TLS verification and original hostname/SNI, disable redirects and environmental proxies, and label that this is direct HTTPS without proxy so it cannot be mistaken for Git's authenticated/proxied route. Choose at most one validated address per path; report that sampling limit. Close response bodies and transport idle connections; do not send credentials or download bodies. API-host checks are DNS observations only. No GitHub API calls or repository writes are made.

## Visible states and actions

Actions: report incident; run the printed net check; perform a known operator remedy; verify/reconcile the existing operation; announce recovery.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Recognized Git DNS/connect failure | infrastructure: network, safe host, exact probe command | run net check | remote write outcome may be unknown; original recovery state retained |
| Ordinary Git failure or cancellation | existing error | existing operation recovery | unchanged |
| System lookup fails, direct succeeds | resolver-path difference, separate observations | report incident with evidence | no work changed |
| Both fail / HTTPS fails | observed DNS/HTTPS failures, cause unknown | report incident | no work changed |
| Missing tool, timeout, unsupported platform | observation unavailable, never fabricated success | report incomplete diagnosis | diagnostic evidence incomplete |
| Incident learned | top Needs-you item + one immediate notify | operator remedy if known, otherwise awareness | blocked operations retain existing state |
| Dependency confirmed recovered | recovery notify, active incident removed | verify/reconcile affected operations before resuming | no automatic replay of uncertain writes |

No new durable incident state machine: the existing TL status file is the incident record and its established edit/notification paths remain authoritative. This issue adds no autonomous outage watcher or retry loop.

## Operating envelope and architecture

- ARCH-DRY/PURE: one classifier and one probe/report seam; no per-call-site error-string matching or resolver IO in Git boundaries.
- ARCH-PURPOSE: sweep all shared transport runners; add discovery in README and atlas as well as help. Test actual fetch failure through the production runner, not just the classifier.
- ARCH-MOCK: a stateful fake probe stores DNS responses and ordered effects, checks that HTTPS receives only addresses returned on that path, and models failure/recovery across repeated checks. Production CLI uses the same interface. Native subprocess fixture tests model missing/hung utilities; a live read-only check compares native observations after staging.
- ARCH-CONSTRAINTS: explicit light online command, one total eight-second budget, native DNS subprocesses each at most two seconds, HTTPS at most three seconds within remaining total. At most four concurrent effects, at most two diagnostic hostnames, two direct queries per hostname, one HTTPS attempt per Git-host resolver path. Cap utility output at 64 KiB each; no unbounded address retry or body reads. Test budgets with controllable barriers rather than a slow public endpoint. No probes on successful Git or ordinary errors.
- ARCH-SECURE: untrusted argv/config/stderr/DNS utility output must become validated hosts/IPs before output or dialing; do not expose URL credentials, invoke a shell, disable certificate verification, follow redirects, or change system configuration. Explicit check intentionally contacts the requested/configured host and local configured DNS server.
- ARCH-ORDER: pure classification preserves original error identity and uncertainty. Extend `cliOwnsReviewSignals` in `cmd/sdlc/main.go` to admit only `check` under parent `net`, retaining its existing command allowlist. The existing executeCLI signal owner then supplies SIGINT/SIGTERM cancellation to the whole probe. Probe effects are scoped to that command context, cancelled and joined before return. A real CLI signal control in `cmd/sdlc/net_signal_unix_test.go`, reusing the subprocess pattern from `cli_signal_unix_test.go`, must wait for a native utility's child-ready barrier, signal the CLI and prove its utility group is gone before exit; injected context cancellation alone is insufficient. Never retry a Git mutation from diagnostic code. Recovery notice follows positive evidence, not a peer's unverified claim.
- ARCH-FUNERAL: creates no runtime files/cache/records or background watcher. Child processes, HTTP bodies and transport connections end with the command. Incident bullets use existing status lifecycle, removed on recovery; evidence remains in existing issue logs.
- ARCH-EVOLVE: no persisted schema changes. Wrapped errors preserve existing errors.Is/As and stdout/stderr contracts. Existing suite verdict format and classification policy remain unchanged.
- Cost: classifier is linear in bounded diagnostic bytes and command argv, no Git/IO calls. Count probe calls for one/two hosts (maximum 3/6 DNS utility calls plus 2 HTTPS attempts) and test that repetitions do not accumulate state or fan-out. No per-byte rescan or address retries.

## Implementation and verification (one atomic review boundary)

### 1. Transport classification

- [ ] Write red tests in `cmd/sdlc/internal/gitx/network_test.go`: `ClassifyNetworkFailure` gets transport-signature property tests and negative nontransport/cancellation controls; `NetworkHost` gets fuzzed arbitrary diagnostics with the invariant that returned hosts are valid and secrets never reach rendered errors; `GitNetworkFailure.Unwrap` retains errors.Is/As identity and wrapping is idempotent. Keep concrete cases in executable fixtures.
- [ ] Add the pure classifier/target parser, then run `go test ./cmd/sdlc/internal/gitx -run 'Test.*Network' -count=1`; require green after observed red.
- [ ] Wire the shared runner boundaries and fix only wrappers that would drop the typed cause. Add production-runner fixtures for fake Git fetch DNS failure and a real isolated local fetch success; assert parsed success output remains unchanged. Exercise a failed local-check preparation and assert no suite verdict/test-failure record is minted from the infrastructure error.
- [ ] Run targeted Git/runner/publication controls and commit source/tests with issue/model trailers.

### 2. Bounded net check

- [ ] Write red tests in `cmd/sdlc/internal/netcheck/*_test.go`: `ParseTarget` and `ParseAddresses` get fuzzed arbitrary config/utility bytes and must return only validated hosts/IPs; `ProbeReport.Render` preserves independently supplied observations without secrets or inferred success; `Check` gets a stateful fake and barriers proving bounded call counts, path-specific address provenance, cancellation/join and no retained state across checks. `NewNetCmd` tests exercise the real Cobra tree with that seam, pinning input/exit/output and remote-selection contracts. Keep concrete cases in executable fixtures.
- [ ] Implement probe utilities and HTTPS adapters behind the injected seam. Add real private subprocess timeout/group cleanup controls and a local TLS conformance control that rejects a wrong hostname/invalid certificate and confirms body/connection cleanup. No test contacts production DNS/hosts.
- [ ] Register `net check` in `cmd/sdlc/main.go`; embed `cmd/sdlc/helptext/net.md`. Check exit status: zero only when all requested DNS paths and both Git-host HTTPS observations complete successfully; nonzero reports all collected failures/unavailable observations first. Help-only and malformed input perform no probes.
- [ ] Run `go test ./cmd/sdlc/internal/netcheck -count=1` and targeted command tests, race controls, and commit.

### 3. Rule, docs, live conformance and landing

- [ ] Add approved TL wording and Needs-you exception; add worker incident-report instruction. Pin required concepts in existing tl-lead/tl-worker help contract tests. These rendered rules are shared across agent kinds, so verify both binary help outputs rather than adding harness-specific prose.
- [ ] Update README discovery and `atlas/workflow/sdlc-binary.md` plus `atlas/workflow/tl-skills.md`; retain atlas index links. Record the semantic diagnostic policy in lessons; do not add a lexical lint that guesses network meaning from arbitrary error strings.
- [ ] Run `go test ./cmd/tl-lead ./cmd/tl-worker -count=1`, targeted SDLC tests, relevant `pkg/golint`/`pkg/shelllint` controls and `go vet` for touched packages. Persist raw evidence on SSD before long checks.
- [ ] Ask hub to review rendered TL rule wording and approve live conformance choice: stage build for read-only `sdlc net check` on this restored machine; no injected outage, DNS flush or system change. Supply binary revision and compare system/direct observations with the command output. Clear only our stage when done.
- [ ] Push branch, run recorded `sdlc local-check run --name suite`, and use resume only if it fails solely on registered flakes. Close with precise evidence (hub owns project), fix review findings, then enter `sdlc merge --yes` promptly. Report verified PR/merge SHA and stop.

## Revisions

### 2026-10-21 — plan-quality PQ-1/PQ-2

PQ-1: explicitly connect net-check to executeCLI's existing signal owner and require real CLI SIGINT/SIGTERM process-lifetime proof, covering the whole utility family rather than only injected cancellation. PQ-2: replace prose case inventories with named risky functions and adversarial/property strategies; executable fixtures will own individual cases. No scope or user-visible contract change. Hub additionally confirmed the staged read-only github.com conformance in delivery `25352ce0-3206-4fc0-90a1-1070b29317c5`.

### 2026-10-21 — implementation organization

Shared host validation lives alongside the pure Git classifier in gitx, so both error formatting and net-check derive the same validated host without a package cycle. Report and bounded orchestration share `internal/netcheck/check.go`; native effects live in `probe.go`. `Report.WriteTo` uses the standard `io.WriterTo` signature `(int64, error)`, as required by go vet, rather than a lookalike signature. These are naming/file-layout refinements; no new mechanism, scope or persisted state. Native HTTPS completes its pinned TLS dial synchronously before HTTP framing and closes the owned socket explicitly, avoiding net/http's detached-dial cancellation lifetime.

### 2026-10-21 — BR-1 transport completeness

The supported network transport families are HTTP(S), SSH/scp and native Git (`git://`); accepting their remote host forms commits the classifier to their resolver/connect/timeout diagnostics. Round-1 close review reproduced a native Git loopback refusal that bypassed classification. `ClassifyNetworkFailure` must cover transport-specific native diagnostics while retaining negative hook/local prose controls and original error identity. Capture the real local refusal through the production runner, and use hermetic diagnostic fixtures/property controls for resolver/timeout variants. This completes the existing Git-network contract; it adds no probe, retry, dependency or protocol. The stale queued source suite is cancelled by its exact lease, and the corrected source will get a fresh recorded suite.

### 2026-10-21 — BR-2 final context precedes classification

Round-2 review disposed BR-1 but reproduced a main bounded-runner ordering defect: a resolver diagnostic was classified before its cancelled context was joined. Keep execution raw until the final existing context owner decides the outcome. The main runner shares a raw stdout/stderr helper; bounded landing performs process-group cleanup and context joining before the pure classifier. Internal raw execution/input/injectable shims similarly defer classification to `RunGitInContext`, `TrunkFile.runIn`, or the bounded-output owner after their existing context/output checks; local history consumers route through the public contextual seam. Preserve the existing cancelled-success regressions unchanged. Add subprocess barriers for cancellation after diagnostic emission and check cancellation/process identities without a network type. A bare deadline stays ordinary; explicit transport evidence may retain a deadline cause as already specified. No error-graph surgery, new persisted state, retry or policy is introduced. The prior queued suite is cancelled by its exact lease and replaced only after the correction.
