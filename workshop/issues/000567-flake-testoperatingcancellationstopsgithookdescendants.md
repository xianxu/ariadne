---
id: 000567
status: codecomplete
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9f2216dbc3a04a3aa2a6b32d42ceb28879ce32b1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T05:17:26-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/tl-lead.TestOperatingCancellationStopsGitHookDescendants]
flow: {kind: quick, provenance: inferred, spec: "cbf89813", done: "c4b06df2"}
actual_hours: 0.35
---

# flake: TestOperatingCancellationStopsGitHookDescendants

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/tl-lead.TestOperatingCancellationStopsGitHookDescendants: Recorded #563 suite attempt 2cb913014a018f68422fdf9ef52180cf on Git-verified head 6a0b35faea2b351b5b1ec19a310f6ea1d19f1948 failed at doc_operating_test.go:419: Git descendant remains runnable: ?E (0.87s). Unchanged exact detached checkout targeted command go test ./cmd/tl-lead -run ^TestOperatingCancellationStopsGitHookDescendants$ -count=1 -timeout 2m returned exit0 PASS in0.986s. Captured original JSON events and targeted result are retained in common test-stats/evidence/000563/cancellation-failure.jsonl and cancellation-targeted-pass.txt. Full original suite is still running; no full retry or source edit. Investigate whether transient Darwin exiting state E is misclassified as runnable, without weakening cleanup proof.

## Spec

The cancellation assertion races asynchronous kernel teardown. The publisher
already kills its private process group on cancellation and return. Its Git
leader and output pipes can finish while a killed descendant is still visible
to Darwin `ps` as `?E` (the native manual defines `E` as trying to exit).
The current one-shot oracle accepts only absence or a leading `Z`, so it calls
this terminal transition runnable. An unchanged 30-run targeted reproduction
on the planning base produced 29 passes and the identical `?E` failure once.

Fix the test's completion boundary, not the publisher: capture the descendant's
owned group before cancellation, reject the test's own group, and wait until
`procgroup.Kill(pgid, 0)` reports `os.ErrProcessDone`, under the existing test
deadline. This shared Unix boundary already handles Darwin exiting/zombie-only
groups and absent groups. A live group remains a failure when the deadline
expires, and unexpected probe errors fail immediately. This is a condition
wait for the original cancellation, never a retry of the operation or test.
The hook child itself publishes its PID and readiness before blocking, so the
assertion covers an executing descendant rather than only the parent's fork.

ARCH-DRY/ORDER: reuse the same group-lifetime semantics as the neighboring
repository-cancellation fixture; do not invent a `ps` state parser or accept
`E` text as proof. A process group is the existing owner-lifetime resource;
no new lease or production state is needed. ARCH-PURE/SECURE: the only new
shipping logic is a pure AST lint rule. Native fixtures use private temporary
Git repositories/FIFOs and signal only the captured owned group. ARCH-PURPOSE:
sweep found this is the only direct `ps stat=` test query in the repository.
ARCH-MOCK: keep real local Git and kernel conformance; no service adapter is
added. ARCH-CONSTRAINTS: preserve all existing command/test deadlines and add
no production hot-path work; lint visits source calls once, O(AST nodes + argv
literals), with counted N/2N checks. ARCH-FUNERAL: immediate fixture cleanup
cancels/joins the publisher and kills its own blocked child on failures;
`t.TempDir` removes readiness files and repositories. ARCH-EVOLVE: no persisted
format, public state or lifecycle contract changes.

Alternatives rejected: accepting an `E` suffix merely adds another textual
state guess; a fixed sleep or longer timeout masks scheduling. Changing the
production group runner would widen scope despite successful group signalling.

## Done when

- The captured native failure is retained with provenance; the unchanged original test reproduces the same `?E` rejection.
- Cancellation waits on the existing owned-group terminal condition, uses child-origin readiness, and preserves the existing deadline and error/WaitDelay assertions.
- A native leader-only-kill mutation fails the descendant guard; original/corrected lint source fixtures reject the one-shot `ps` oracle without banning unrelated commands or production diagnostics.
- Focused normal and race checks plus a recorded full suite pass; closing this issue unregisters its flake.

## Plan

- [x] Retain the original JSON capture and test-source excerpt; make the narrow `test-ps-liveness` catalog rule red on that source, with import-alias/context/absolute-binary and unrelated-command controls plus counted N/2N visits.
- [x] Fix `TestOperatingCancellationStopsGitHookDescendants` to obtain child-origin readiness, capture and validate its group before cancel, and observe the shared terminal condition under its existing guard; register cleanup before failure-prone waits.
- [x] Verify the real Git fixture normally and with race instrumentation; temporarily change only the operating command kill to leader-only, observe the negative control fail, restore it, and retain the result. The test must not accept a surviving owned group or an unexpected probe error.
- [x] Run the tracked-file catalog and relevant cancellation controls, record the lesson and atlas rule, and push the tested correction for the recorded acceptance gates.

Risky function strategies: `TestOperatingCancellationStopsGitHookDescendants`
uses a ready blocked native child and the independent leader-only-kill mutation;
`psLivenessQueries` receives the recorded old oracle and adversarial command
syntax, checked through both Check and CheckSources. Finite case inventories
live in executable tests. No new user-visible state or production state machine
is introduced.

## Log

### 2026-10-22

### 2026-10-23 — diagnosis and proposed quick-flow fix
- 2026-10-23: closed — BR-1 native ps format family verified and corrected through one shared predicate: all Check/CheckSources separator and command-form controls pass, counted visits remain linear, full golint PASS 0.813s, race controls PASS 1.157s. Native cancellation correction remains unchanged and passed 30/30 plus five race repetitions and broader operating checks; leader-only negative control refuses surviving work. Old queued suite canceled because superseded; fresh recorded suite runs on pushed correction before landing. No production cancellation/deadline change; atlas and lessons updated. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

Claimed and started on the fresh main base. Native original repetition:
`go test ./cmd/tl-lead -run '^TestOperatingCancellationStopsGitHookDescendants$' -count=30 -timeout 3m -json`
returned exit 1 in 26.605s: 29 pass, one failure with descendant `?E`, exactly
matching the original #563 capture. Raw results are under the Git common dir,
`sdlc/test-stats/evidence/000567/original-repeat.jsonl`; no code was edited.
The native `ps(1)` state section documents `E` as trying to exit; existing
`pkg/procgroup.Kill` explicitly covers XNU's exiting/zombie-only group result.
Production `operatingCommandGroup` already installs Setpgid, group SIGKILL and
one-second WaitDelay; adjacent the repository-cancellation regression uses a bounded
group completion probe instead of an immediate `ps` snapshot. The proposed
patch changes the fixture/oracle and adds the mechanical lint backstop only.

### 2026-10-23 — correction and independent negative control

The catalog regression first failed because `test-ps-liveness` was absent;
after adding the rule, its captured-source and syntax controls passed. The
tracked-file catalog then rejected the actual old oracle at its original
source line. The corrected fixture and catalog passed together (tl-lead
1.847s, golint 0.851s). The rule has no production process observer or state
parser; it visits each source call once and examines literal command/format
arguments, with N=16/32 counted predicate calls. Dynamic argv and diagnostics
are deliberately outside this narrow lexical guard.

Temporary native negative control replaced only the production group kill
with leader kill and bounded the observation *after publisher return* to one
second. The corrected test failed explicitly: `Git descendant group ...
remains runnable` (2.81s test; 2.965s package). Both temporary mutations were
restored byte-for-byte; Git confirms no production runner diff. An earlier
diagnostic started its short guard immediately after cancel and stopped at
`cancellation did not return`, because it raced the existing one-second
WaitDelay; that unsuccessful diagnostic is retained separately. No shipped
timeout changed. Evidence files live in common
`sdlc/test-stats/evidence/000567/leader-only-negative{,-group}.txt`.

### 2026-10-23 — focused verification complete

Corrected original fixture: 30/30 pass, package 22.099s. Race-enabled operating
cancellation, neighboring repository cancellation and lint controls repeated
five times: PASS (tl-lead 10.935s, golint 1.205s). Broader operating-publication
and repository-cancellation tests: PASS 18.555s. Full Go lint catalog and
native procgroup package: PASS 0.739s / 0.081s. The first complete lint run
correctly required the new rule's atlas catalog entry; that entry is now
present and verified. The captured source oracle, native failure and successful
leader-only negative control are committed under
`pkg/golint/testdata/operating-cancellation/`. Remaining test logs stay in the
common evidence directory. Central SDLC reports the primary path, current
28c40fd5 revision and modified=false. Next: push PR, recorded full suite, close
review, then central merge; no staging or live fleet process check is needed
for this test-only cancellation correction and pure lint guard.

### 2026-10-23 — recorded suite queued; close overlap

PR #414 is open. Central `sdlc local-check run --name suite` is retained on
pushed implementation e2b301d63319abda81e91213f74b15e916a6b1b5, run
`lease-2313690157`, private exact-head worktree
`/Volumes/TestScratch/local-checks/sdlc-local-1970967993`. Its first admission
report was position seven behind two Pair suites; no grant or live queue
mutation. The implementation plan is complete; remaining suite, close and
landing outcomes are tracked here rather than checked off before they happen.
Per tl-worker's explicit overlap rule, the mandatory close review may run
while this suite waits. No landing proceeds without a recorded pass on the
final code; a code change from review requires fresh verification. This is an
acceptance checkpoint, not a suite-pass claim.

### 2026-10-23 — boundary round one; BR-1 corrected

Round one returned FIX-THEN-SHIP with one Important finding, but the ledger
correctly refused finalization while BR-1 remained open. The reviewer reran
30 native cancellation repetitions, race controls and broader operating tests
successfully; the defect was incomplete literal ps format recognition in the
new lint. Independently verified `/bin/ps` accepts `pid stat=`, `pid, stat=`,
`pid, state=` and tab separators, including attached -o and -O forms. Added a
format-family matrix across Command/CommandContext and both Check/CheckSources;
it failed on the original matcher. One shared predicate now normalizes comma
and whitespace fields; full golint PASS 0.813s and expanded race controls PASS.

Canceled only this slot's still-waiting recorded suite through
`sdlc local-check cancel --name suite`; it had not acquired capacity. Its runner
exited interrupted/canceled, and the CLI confirmed two other checkouts' runs
were left running. The old queued code is superseded, so a fresh recorded
suite on the pushed correction is required. No main integration or waiter
reorder. Next: fresh push/suite, then ordinary close again as the gate names
(not a bypass or an unfinalized fix-only close).

### 2026-10-23 — SHIP and retained final-code suite

Round two returned SHIP, disposed BR-1 and raised no new findings. The reviewer
independently repeated 30 native cancellation cases, race and operating checks,
matched the captured source to the pinned base, and reproduced the leader-only
negative control. Close finalized codecomplete with evidence commit
843549fcb391e21e27b8cd37ee66e4b1830c6b90 and mirror
3a57c1470686b8fbb8d423d335ea9ecb6b20f5af, automatically measured actual 0.35h.
CI is green on that pushed close head.

The sole retained branch-suite controller is the central local-check on
839fc18ae652343c0777a8a515ac943476fc32d3, run `lease-2618842037`, exact-head
worktree `/Volumes/TestScratch/local-checks/sdlc-local-4107304242`, inflight log
`61fe807ebaf6fc547a0cd20c7e4c708e/run.log`. It is now next in line after over
one hour of normal FIFO waiting, with two Echo suites running. Nothing has
started in this attempt yet; preserve it. The old e2b301d6 attempt was canceled
and is not acceptance evidence. No code has changed since the final reviewed
and queued 839fc18a head. After this retained suite passes, push the final
bookkeeping if needed and run central `sdlc merge --yes`; report only after
Git, PR and tracker landing evidence agree. Do not manually chase newer main
or restart a retained waiter.

### 2026-10-23 — recorded branch suite passed

Retained `lease-2618842037` acquired both slots after 3760s. The full suite on
839fc18ae652343c0777a8a515ac943476fc32d3 returned exit 0: 2374 cmd/sdlc tests
over eight shards plus 76 other packages, zero failures and no reruns, 956.0s
(15m56s). The corrected cancellation case passed under suite load (0.98s).
Saved proof: common `sdlc/local-checks/839fc18ae652343c0777a8a515ac943476fc32d3/suite.log`,
SHA-256 prefix `035c9478f411`, recorded `sdlc-local/suite` success. All code is
unchanged since this tested and reviewed head; subsequent commits contain
close evidence/card mirror and this Log only. Round-two SHIP has no open
findings. Next: central `sdlc merge --yes` owns integration, CI, any required
integration suite, archive and final cleanup. No manual integration or fresh
pre-close suite is needed.

## Revisions

### 2026-10-23 — quick-flow implementation detail and verification

The small fix plan was posted to the hub and its delivery was verified before
change-code inferred quick flow. Named the lint's pure per-call predicate
`psLivenessQuery` and one-pass visitor `psLivenessCalls` for counted tests;
`psLivenessQueries` remains the catalog entry point. The native negative
control's diagnostic observation starts after publisher return (rather than
racing WaitDelay); all such instrumentation is restored before verification.
The atlas catalog entry is included alongside the lesson. No change to the
approved test-only boundary or production cancellation/deadline semantics.

### 2026-10-23 — plan bookkeeping revision

The final Plan checkbox now describes completed implementation preparation;
recorded-suite, review and landing outcomes remain explicit Log commitments.
This removes a circular checkbox asking for landing before the close gate,
without changing Done-when or claiming unfinished acceptance.

### 2026-10-23 — Revisions: BR-1 literal-format family

Extended the existing lint test inventory to native format separators and
attached -O alongside -o and Linux --format=. The command/format boundary and
Done-when remain unchanged. Concrete spelling inventory stays in the matrix;
source-call visits remain N/2N linear. The lesson records the shared dependency
field grammar. This is the first design correction to the lint mechanism.
