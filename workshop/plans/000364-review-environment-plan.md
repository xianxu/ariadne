# Boundary Review Environment and Answerable Findings Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development for bounded independent work and retain integration in the main session. Steps use checkbox syntax for tracking. The SDLC binary owns the final fresh-context review.

**Goal:** Let boundary reviewers execute process-sensitive tests in prepared disposable checkouts, and require every REWORK to name a finding the author can answer.

**Architecture:** Extend the existing reviewbox fence, judge attempt lifecycle, and findings ledger. Preparation and response admission share the existing dispatch deadline and two-attempt budget. Add no alternate reviewer, general sandbox, or independent dispute system.

**Tech Stack:** Go, CUE vocabulary, Darwin Seatbelt/Linux bubblewrap, existing Git checkout and subprocess seams.

## Approved direction and measured baseline

The hub approved this direction and proceeding through change-code on 2026-10-14
(Couch delivery b9041b2d-23c0-4f77-a308-42184316e7c9). Full flow, one atomic close;
the expected code delta exceeds the quick-flow shell. No estimate until plan-quality passes.

The initial Spec assumed general signal/network/outside-write denials. The existing
fence instead denies writes to source worktrees and their shared Git directory;
the hub explicitly chose to preserve that model, not introduce broader confinement.

Measured on this host before implementation:

- Direct `/bin/ps -p 1 -o pid=,ppid=` succeeds.
- `/bin/ps` is setuid root; `sandbox-exec -p '(version 1)(allow default)'` refuses
  its execution with EPERM. Thus changing only Codex's sandbox mode is insufficient.
- Adding `(allow process-exec (literal "/bin/ps") (with no-sandbox))` succeeds.
- With that exact exception plus a protected-directory write deny, ps and scratch
  writes succeed, while a protected write remains denied. Non-setuid pgrep already works.
- pair#528's M1 sidecar labels the affected rounds `fenced`; current fenced Codex
  launch already uses `--sandbox danger-full-access` beneath the outer write fence.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| ReviewPreparation declaration/parser | `cmd/sdlc/internal/reviewprepare/config.go` | new |
| PreparationResult / renderer input | `cmd/sdlc/internal/reviewprepare/result.go` | new |
| Finding kind vocabulary | `construct/vocabulary/finding.cue`, `pkg/vocab/finding.go` | modified |
| Finding / RoundReport parsing and prior-finding rendering | `cmd/sdlc/internal/gatestate/{ledger,parse,prompt,render}.go` | modified |
| BoundaryResponse admission | `cmd/sdlc/boundaryresponse.go` | new |
| Isolation argv/profile | `cmd/sdlc/internal/reviewbox/fence.go` | modified |

One declaration belongs to a pinned review HEAD; zero or one preparation executes
per attempt. The parser consumes bytes, not a filesystem. A result contains plain
values supplied by the executor: argv, attempt, duration, exit/error, and bounded
output. No clocks, callbacks, or IO are called by the parser or renderer.

Finding kind is optional for old ledgers; absence retains ordinary defect semantics.
The modeled kinds are `defect` and `unverifiable`. Severity, stable IDs, boundary
scoping and implementer/judge authority retain their current owners. Admission is a
pure function over output and the pinned scoped ledger, reusing existing parsers and
checked disposition rules; do not import filesystem state into that function.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| Pinned preparation reader | `cmd/sdlc/reviewprepare.go` | new | Git tree entry/blob reads |
| Attempt lifecycle / preparation executor | `cmd/sdlc/internal/judge/{dispatch,prepare}.go` | modified/new | bounded process groups, confinement, existing Wrap |
| Boundary wiring and evidence | `cmd/sdlc/{milestoneclose,reviewisolation,reviewsidecar}.go` | modified | disposable checkout, ledger snapshot, sidecar IO |
| Fence conformance probe | `cmd/sdlc/internal/reviewbox/{probe,probe_test}.go` | modified | real OS fence and process inspection |

Tests reuse real temporary Git repositories and stateful fixture executables that
record attempt events and generated files. The same Wrap/process runner used in
production executes these fixtures. Live Darwin conformance verifies actual ps and
write denial; a skipped test is not evidence of that acceptance condition.

## Contracts

### Darwin process inspection

Add exactly the literal system-binary rule measured above to the Seatbelt profile.
Do not add wildcard paths, a shell escape, a caller-specified executable exception,
or permission bypass flags. Linux's existing process visibility remains unchanged.
The reviewer retains its current agent mode under a successfully probed outer fence.
An inherited parent sandbox can still refuse operations; report that limitation
honestly rather than weakening it. A reviewer unable to verify raises an
`unverifiable` finding instead of an empty REWORK.

### Preparation declaration

Use `.sdlc/review-prepare.json` (already under the existing local-gates guard):

```json
{"version":1,"command":["make","runtimebundle-generate"],"timeout_minutes":10}
```

- Missing declaration means no preparation. A present empty file, symlink, malformed
  JSON, unknown fields/version, trailing JSON, empty argv/arguments, or invalid timeout
  is an actionable setup error. Read the regular Git blob from the pinned HEAD;
  never follow a mutable working-tree declaration or a symlink outside the checkout.
- Execute argv directly (no implicit shell) at the disposable checkout root under
  the existing Wrap and confinement environment. This executes the repository's
  own build code with the same authority as its tests; it grants no new privileges.
- Default timeout 10 minutes; allowed explicit range 1–30 minutes, capped by the
  remaining dispatch deadline. These are initial operational ceilings, not measured
  runtime promises; record live pair generation time and revise if evidence requires it.
- Retain at most 64 KiB of combined preparation diagnostics per attempt, with an
  explicit truncation marker. Stream draining continues after the cap, without
  unbounded in-memory capture. No secrets belong in a declaration or its output.
- Run after confinement is ready on attempt 1; on retry run after checkout.Reset
  and conf.Prepare. Preparation failure, launch failure, cancellation, or timeout
  prevents that reviewer attempt. Reuse bounded process-group termination/reaping.
- Reject tracked-tree changes made by preparation; ignored/untracked build outputs
  are allowed and disappear with the checkout. Never run prep in the live source.
- Record command, HEAD, attempt, duration, result, and bounded output in boundary
  evidence even on failure. Keep these diagnostics separate from semantic reviewer
  text so a build log containing verdict fences cannot influence admission.
- Dry-run shows the declaration and ordering but starts no preparation or reviewer.

The operational sequence is:

`create checkout → confine → prepare → review → validate response → accept`

On one invalid response: `reset checkout → re-create confinement → prepare → review`.
After the second invalid response, halt with a protocol error. Setup errors and
cancellation halt immediately; neither is an accepted REWORK or a passing result.

### Answerable boundary response

Wire optional context-aware preparation and pure response-validation callbacks into
DispatchOptions, leaving other judge callers' behavior unchanged. Reuse the single
existing two-attempt loop; missing verdicts and invalid responses share its budget.

Boundary admission rules:

1. Parse the authoritative verdict and findings with existing precedence rules.
2. REWORK needs at least one new blocking finding with a nonempty title/detail, or
   an explicit `not-addressed` disposition of a known blocking finding in scope.
   New findings use `id: new`; the binary assigns BR-n after admission. Merely having
   an old open finding, a Minor-only finding, or an unknown ID does not justify REWORK.
3. Validate referenced dispositions against the pinned ledger before accepting any
   boundary response. Existing gate rules still decide whether a final verdict ships.
4. Each pending dispute of an `unverifiable` finding requires an explicit valid
   judge disposition with a nonempty note answering the host evidence. Scope this
   new requirement to that kind; preserve ordinary historical dispute semantics.
5. An unverifiable finding's detail names the attempted command/check, observed
   restriction/failure, and what evidence or environment change would answer it.
   Require nonempty detail mechanically; semantic specificity remains review-owned.
6. First invalid response: retain diagnostic evidence, tell the retry exactly why it
   was invalid, and reset/reprepare. Second invalid response: explicit protocol error,
   no accepted REWORK, no findings from either invalid attempt, no reviewed-head
   advancement, and no close finalization. Do not feed concatenated invalid text back
   to the verdict parser. Keep failure evidence in the existing sidecar/diagnostics.

### Diagnostic retention

Preparation/protocol evidence is permanently retained in the existing append-only
review sidecar and archived with the issue, including failed invocations that never
produce a ledger round. No expiry or automatic deletion is promised. The new
`reviewprepare.RenderDiagnostics` limits the entire added diagnostic envelope to
256 KiB per dispatch, including metadata, truncation markers and invalid-response
excerpts. The 64 KiB preparation-output cap per attempt remains a subsidiary cap;
remaining envelope space holds protocol evidence. Capture drains beyond caps.
The renderer records retained byte count and truncation; `writeReviewSidecar`
reports that count at append time, where repeated-invocation cost is measurable.
Four maximally verbose failed invocations add at most 1 MiB of diagnostics to the
working sidecar (Git history/storage overhead is additional). Existing accepted
review bodies are unchanged and outside this new diagnostic budget. Tests measure
the envelope and appended byte delta, including retries and repeated failed calls.

## Architecture and lifecycle checks

- **ARCH-DRY:** one attempt loop, process runner, ledger/dispute lifecycle, and binary-owned review.
- **ARCH-PURE:** declaration validation, response admission, and renderers consume values;
  Git/clock/process/callback effects stay in the integration rows above.
- **ARCH-PURPOSE:** prove real ps through the fence, actual generated assets, empty-REWORK
  refusal, and the full host-evidence dispute round trip; no prompt-only implementation.
- **ARCH-MOCK:** stateful fixture commands share the real executor; live ps and pair
  generation checks pin the OS/build assumptions beyond command-line assertions.
- **ARCH-CONSTRAINTS:** one prep and one reviewer at a time, maximum two attempts;
  existing overall review deadline plus prep cap, bounded output, five-second shutdown.
- **ARCH-SECURE:** pinned regular config blob, literal root-owned system ps exception,
  existing protected-path fence, no implicit shell or runtime permission broadening.
- **ARCH-ORDER:** reset/reprepare before retry; cancellation kills/reaps descendants;
  stale prepared boundary/ledger validation remains authoritative at finalization.
- **ARCH-FUNERAL:** generated artifacts stay in the existing disposable checkout;
  normal cleanup and crash sweep remove them. Bounded evidence uses existing sidecars
  archived with the issue; no new persistent cache or queue.

## Function-level verification strategy

The following strategies own the test scope; task checkboxes refer to these
rather than adding a procedural case inventory. Tests precede implementation.

| Function | Risk / adversarial input class | Strategy and mechanical guard |
|---|---|---|
| `reviewprepare.Parse` (new) | malformed or ambiguous declaration bytes | Seeded fuzzing of strict JSON validation; accepted values always satisfy version/argv/timeout invariants; remove unknown-field/EOF validation and observe a failing seed. |
| `readReviewPreparation` (new, integration) | mutable or nonregular declaration at pinned revision | Temporary real Git repository changed after pinning; assert returned bytes follow the pinned regular blob; mutation removing mode/pin guard must fail. |
| `validateBoundaryResponse` (new, pure) | malformed responses and cross-scope finding references | Seeded fuzzing plus a scoped ledger model; admitted REWORK always has a new blocker or valid maintained blocker, and every pending unverifiable dispute has an answer; remove each admission guard and observe failure. |
| `gatestate.ParseFindingsBlock` | malformed/unknown kind and legacy omission | Seeded parser fuzzing checks unknown kinds cannot enter accepted reports while historical omission retains defect semantics. |
| `gatestate.Render`, `ParseSidecar`, `RenderPriorFindingsScoped` | reordered ledger entries, legacy and new finding kinds | Round-trip property and repeatable permutation checks preserve identity/kind/scope and deterministic canonical output; mutate kind propagation and observe failure. |
| `reviewprepare.RenderDiagnostics` (new, pure) | oversized and verdict-looking diagnostic text | Property test supplies arbitrary bounded results; rendered envelope never exceeds the dispatch byte cap, remains separate from semantic output, and reports truncation and retained bytes. |
| `seatbeltProfile`, `verifyFence` | process exec alongside protected writes | Real Darwin conformance through direct and shell-child ps plus protected/scratch write probes; removing the literal rule reproduces EPERM. |
| `runPreparation` (new), `dispatchAttempts` | failed/hanging commands, descendant processes, invalid responses | Stateful executable fixture records events through the real process seam; reproducible event sequence enforces reset→confinement→prep→review with at most two attempts, strict failure stopping and descendant cleanup. Remove ordering/admission guards to prove the harness detects the fault. |
| `dispatchBoundaryReview`, `persistBoundaryRound` | invalid response after a disputed stable-ID finding | Real temporary-repo lifecycle proves host evidence reaches the next judge, invalid text cannot advance accepted ledger/head, and a valid disposition closes the loop. |
| `writeReviewSidecar` | repeated failed dispatches | Append two maximum diagnostic envelopes and measure byte deltas; each added envelope obeys the cap and reports its size even when no ledger round is accepted. |

## Tasks

### Task 1: Prove and add the literal Darwin ps exception

**Files:** `cmd/sdlc/internal/reviewbox/fence.go`, `fence_test.go`, `probe.go`,
`probe_test.go`; `cmd/sdlc/reviewisolation_test.go` as needed.

- [x] Add a failing profile test and real Darwin test: known PID read succeeds,
  scratch write succeeds, protected write fails. Test direct and shell-child ps.
- [x] Run `go test ./cmd/sdlc/internal/reviewbox -run 'Test.*(Fence|Process)' -count=1`;
  observe the ps failure without the literal rule on a fence-capable Darwin host.
- [x] Add only `(allow process-exec (literal "/bin/ps") (with no-sandbox))` to the
  generated Seatbelt profile. Keep supported-OS fallback diagnostics explicit.
- [x] Rerun, then remove the rule temporarily and prove the conformance test fails;
  restore it and commit the isolated task with an issue reference.

### Task 2: Model and validate answerable findings

**Files:** `construct/vocabulary/finding.cue`, `pkg/vocab/finding.{go,json}` and tests;
`cmd/sdlc/internal/gatestate/{ledger,parse,prompt,render}.go` and tests;
new `cmd/sdlc/boundaryresponse{,_test}.go`.

- [x] Write failing tests using the named admission, parser and renderer strategies above.
- [x] Add kind vocabulary and derive parser/rendered instructions from it; regenerate
  exports with the existing vocabulary tools, not handwritten generated JSON.
- [x] Implement one pure admission function over response + scoped ledger, using
  existing parsing and disposition validation rather than duplicating authority rules.
- [x] Run `go test ./pkg/vocab ./cmd/sdlc/internal/gatestate -count=1` and
  `go test ./cmd/sdlc -run '^TestBoundaryResponse' -count=1`; prove the rejection
  tests fail with their guards removed, restore, and commit.

### Task 3: Prepare each attempt and enforce response admission

**Files:** new `cmd/sdlc/internal/reviewprepare/{config,result}{,_test}.go`;
new `cmd/sdlc/reviewprepare{,_test}.go`; `cmd/sdlc/internal/judge/dispatch.go`,
new `prepare{,_test}.go`, `dispatchstream_test.go`, cancellation tests;
`cmd/sdlc/{milestoneclose,boundaryledger,reviewisolation,reviewsidecar}.go` and tests.

- [x] Write failing declaration, pinned-reader, diagnostic and lifecycle tests using
  the named function strategies above.
- [x] Implement declaration IO and the context-aware preparation hook inside the
  shared attempt lifecycle. Reuse process cancellation mechanics with strict prep
  exit handling; do not reuse reviewer nonzero-exit swallowing.
- [x] Add response validation to the same retry decision; no separate nested retry.
  Preserve existing missing-verdict behavior for callers without a validator.
- [x] Wire the pure boundary validator to close, milestone-close and fix-review.
  Persist prep/protocol diagnostics without accepted verdict/round/head advancement
  from invalid text; preserve stale-result checks and existing ledgers.
- [x] Execute the named boundary lifecycle strategy through the production seams.
- [x] Run `go test ./cmd/sdlc/internal/judge ./cmd/sdlc/internal/reviewprepare -count=1`
  and `go test ./cmd/sdlc -run 'Test.*(BoundaryResponse|ReviewPrepar|BoundaryReview|ReviewSidecar|Dispute)' -count=1`;
  mutation-check each new rejection/ordering branch, restore, and commit.

### Task 4: Document, prove downstream behavior, and land

**Files:** `README.md`, `atlas/workflow/sdlc-binary.md`,
`cmd/sdlc/internal/judge/{contract.go,code-review.md,boundary-tail.md}` and prompt goldens;
`cmd/sdlc/helptext/{close,milestone-close,finding-dispute}.md` where applicable;
this issue and plan. Update atlas index only if a new atlas page is introduced.

- [x] Document the literal ps exception, pinned prep declaration and limits, retry
  ordering, failure diagnostics, kind/dispute requirement, and unchanged fallback limits.
- [x] Prepare a disposable copy of pinned pair main with a tracked declaration using
  `make runtimebundle-generate`; execute through the production preparation seam and
  verify `go test ./cmd/internal/runtimebundle -count=1` succeeds in that checkout.
  Record actual revision, command, duration, and protected-source invariants.
- [x] Coordinate the real pair declaration as a separate pair-owned configuration
  issue/PR through the hub; `.sdlc/*` requires the operator's local-gates-verified label.
  Do not edit another slot or apply that label. Pair#547 independently removes ps execs.
- [x] Run live Darwin fence conformance and an installed Codex fenced-review smoke
  that calls ps; use current agent modes, not bypass-approval flags. Run Linux fence
  coverage on a suitable runner or explicitly report any missing platform evidence.
- [ ] Run full `make -k test` in the background on a committed stable tree; fix failures.
  Then `sdlc close --issue 364 --verified '<actual behavior evidence>'`; resolve findings.
- [ ] Fetch/merge main, `sdlc pr`, wait for actual CI checks, `sdlc merge --yes` on SHIP,
  and report the real PR/hash to the hub. Stop at operator-held config verification
  if such a protected path is touched; never waive CI or label it yourself.

## Execution ownership

Task 1 and the pure/model part of Task 2 are bounded independent subagent work.
The main session owns Task 3's shared lifecycle integration and Task 4. Avoid
concurrent edits to dispatch, ledger wiring, and prompt files. One final SDLC review
owns acceptance; do not add a duplicate boundary review outside the binary.

## Revisions

- 2026-10-14 — Plan-quality PQ-1: replaced procedural test inventories with named
  function-level strategies, adversarial classes, fuzz/property checks and mechanical
  guards. PQ-2: declared permanent diagnostic retention, a 256 KiB writer envelope
  per dispatch, visible retained-byte measurement and repeated-failure cost. Scope
  and approved behavior remain unchanged.

- 2026-10-14 — Implementation and live proofs complete. Full make -k suite passed
  at 74bd16708c717eeafdcf599ad9842180c09c6ed4 (1,249 SDLC tests, 53 packages,
  449.2 s). The last two administrative rows remain pending until their close and
  landing actions complete. Linux live coverage remains unavailable locally;
  the existing Linux fence behavior was preserved.

- 2026-10-14 — Boundary BR-1 (Critical, review-input-integrity): Git's normalized
  diff is not an oracle for actual tracked contents. Replace tracked-state
  comparison with a filesystem fingerprint of tracked file bytes, types, modes
  and symlink targets, retaining HEAD/index identity. The comparison must ignore
  Git text/clean normalization and assume-unchanged/skip-worktree suppression;
  generated untracked/ignored files remain allowed. Include direct production
  Dispatch regressions for each hiding mechanism and type/mode changes.

### BR-1 correction

- [x] Replace `trackedPreparationState` with a direct tracked-filesystem fingerprint
  and prove hidden modifications halt before reviewer execution. This strengthens
  the existing no-tracked-mutation contract; no new confinement model is introduced.
- [x] Rerun targeted lifecycle tests and the full suite, then re-close the finding.

### BR-2 correction

- 2026-10-14 — Boundary round 2 accepted BR-1 after an independent mutation
  check. BR-2 (Important, boundary-lifecycle-coverage) correctly identifies that
  pure admission tests do not prove the promised production dispute lifecycle.
  Add a temporary-repository close/dispute/review regression using existing
  fixtures: host evidence reaches the reviewer, unanswered and note-less replies
  cannot advance accepted ledger/head, and an explicit answering disposition
  closes the loop. Mutation-check removal of the dispute guard or ledger wiring
  (ARCH-PURPOSE). This completes existing scope, with no new production behavior.

- [x] Prove production unverifiable dispute lifecycle and kill a guard/wiring mutation.

BR-2 verification: `TestCloseUnverifiableDisputeLifecycle` passes in 16.005 s.
The real close/dispute path delivers host evidence and kind to the reviewer,
rejects both unanswered and note-less replies after two attempts without advancing
ledger, HEAD, reviewed ref, or issue state, and accepts an explicit answering
withdrawal. Bypassing the dispute guard made both invalid replies close the issue;
the new test failed (13.068 s). Restored production file exactly; final run green.
Full suite on bd486507 remains the production baseline; this correction adds only
the regression test and documentation, so targeted verification covers the delta.
