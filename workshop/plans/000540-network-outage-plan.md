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
| NetworkTarget | `cmd/sdlc/internal/netcheck/target.go` | new |
| ProbeReport | `cmd/sdlc/internal/netcheck/report.go` | new |

`GitNetworkFailure` classifies only transport-specific resolver, refused-connection and connect/transport-timeout diagnostics from fetch/push/pull/clone/ls-remote. Its Error starts `infrastructure: network`, names a validated host when observed, and gives the exact `sdlc net check --host HOST` command (plain `sdlc net check` when host unknown). `Unwrap` retains the original exit/context cause; repeated wrapping is idempotent. A cancellation or bare process deadline is not evidence of a network fault. Auth, ref/lease rejections, hooks, local IO and local-command timeout prose remain ordinary errors. It never asserts that an uncertain push failed remotely or prescribes an automatic retry.

`NetworkTarget` parses only a host from HTTPS/SSH/scp-style remote URLs or recognized transport diagnostics. URL credentials, path, query and fragment never enter new output. Handle IPv4, IPv6 and SSH aliases explicitly; local paths and helper transports cannot manufacture a host. A proxy resolver failure identifies a proxy observation rather than pretending it proves the Git origin unavailable.

`ProbeReport` renders per-path observations, including unknown/unavailable/timeout, without collapsing them into “upstream down.” System failure with direct success is evidence of a resolver-path difference, not proof of a particular cache fault. HTTP responses, including 401/403/5xx, prove HTTPS reachability only and retain the status code.

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
- ARCH-ORDER: pure classification preserves original error identity and uncertainty. Probe effects are scoped to the command context, cancelled and joined before return. Never retry a Git mutation from diagnostic code. Recovery notice follows positive evidence, not a peer's unverified claim.
- ARCH-FUNERAL: creates no runtime files/cache/records or background watcher. Child processes, HTTP bodies and transport connections end with the command. Incident bullets use existing status lifecycle, removed on recovery; evidence remains in existing issue logs.
- ARCH-EVOLVE: no persisted schema changes. Wrapped errors preserve existing errors.Is/As and stdout/stderr contracts. Existing suite verdict format and classification policy remain unchanged.
- Cost: classifier is linear in bounded diagnostic bytes and command argv, no Git/IO calls. Count probe calls for one/two hosts (maximum 3/6 DNS utility calls plus 2 HTTPS attempts) and test that repetitions do not accumulate state or fan-out. No per-byte rescan or address retries.

## Implementation and verification (one atomic review boundary)

### 1. Transport classification

- [ ] Write table-driven red tests in `cmd/sdlc/internal/gitx/network_test.go` for HTTPS/SSH/proxy DNS, connection refusal, transport timeout, secret-bearing URLs, IPv6, unknown host, duplicate wrap, preserved exit/cancellation identity, and negative auth/ref/lease/hook/local errors.
- [ ] Add the pure classifier/target parser, then run `go test ./cmd/sdlc/internal/gitx -run 'Test.*Network' -count=1`; require green after observed red.
- [ ] Wire the shared runner boundaries and fix only wrappers that would drop the typed cause. Add production-runner fixtures for fake Git fetch DNS failure and a real isolated local fetch success; assert parsed success output remains unchanged. Exercise a failed local-check preparation and assert no suite verdict/test-failure record is minted from the infrastructure error.
- [ ] Run targeted Git/runner/publication controls and commit source/tests with issue/model trailers.

### 2. Bounded net check

- [ ] Write red report/target tests and stateful fake-probe CLI tests in `cmd/sdlc/internal/netcheck/*_test.go` and `cmd/sdlc/net_test.go`: both lookups visible, disagreement, API observation, distinct fetch/push host, remote override, credentials omitted, tool absence, cancellation, bounded output, and repeat failure then recovery.
- [ ] Implement probe utilities and HTTPS adapters behind the injected seam. Add real private subprocess timeout/group cleanup controls and a local TLS conformance control that rejects a wrong hostname/invalid certificate and confirms body/connection cleanup. No test contacts production DNS/hosts.
- [ ] Register `net check` in `cmd/sdlc/main.go`; embed `cmd/sdlc/helptext/net.md`. Check exit status: zero only when all requested DNS paths and both Git-host HTTPS observations complete successfully; nonzero reports all collected failures/unavailable observations first. Help-only and malformed input perform no probes.
- [ ] Run `go test ./cmd/sdlc/internal/netcheck -count=1` and targeted command tests, race controls, and commit.

### 3. Rule, docs, live conformance and landing

- [ ] Add approved TL wording and Needs-you exception; add worker incident-report instruction. Pin required concepts in existing tl-lead/tl-worker help contract tests. These rendered rules are shared across agent kinds, so verify both binary help outputs rather than adding harness-specific prose.
- [ ] Update README discovery and `atlas/workflow/sdlc-binary.md` plus `atlas/workflow/tl-skills.md`; retain atlas index links. Record the semantic diagnostic policy in lessons; do not add a lexical lint that guesses network meaning from arbitrary error strings.
- [ ] Run `go test ./cmd/tl-lead ./cmd/tl-worker -count=1`, targeted SDLC tests, relevant `pkg/golint`/`pkg/shelllint` controls and `go vet` for touched packages. Persist raw evidence on SSD before long checks.
- [ ] Ask hub to review rendered TL rule wording and approve live conformance choice: stage build for read-only `sdlc net check` on this restored machine; no injected outage, DNS flush or system change. Supply binary revision and compare system/direct observations with the command output. Clear only our stage when done.
- [ ] Push branch, run recorded `sdlc local-check run --name suite`, and use resume only if it fails solely on registered flakes. Close with precise evidence (hub owns project), fix review findings, then enter `sdlc merge --yes` promptly. Report verified PR/merge SHA and stop.
