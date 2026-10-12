# Boundary Review — ariadne#309 (milestone M2)

| field | value |
|-------|-------|
| issue | 309 — Split local gates and server CI checks |
| repo | ariadne |
| issue file | workshop/issues/000309-split-local-gates-and-server-ci-checks.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 311977d89393ac8c4dfe1f077b95e390e021383b..3982c4d65ad7a843841d27294115ed9c5e0cf10e |
| command | sdlc milestone-close --issue 309 --milestone M2 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T13:13:44-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Diff inspected; now checking test wiring, #052 status, and the path-quoting behavior of the guard.

I accidentally ran the scratch commands in the home directory when `mktemp` failed. I need to check what was created there before I clean anything up.

Tests pass. Now checking whether the guard misses protected paths that git prints in quoted form, using a scratch repo under /tmp/claude.

Confirmed: the guard misses protected paths that git prints in quoted form. Checking what the manifest's `seed` verb means for propagation.

Running the classifier test this window added.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2 delivers most of what the plan promised. The guard runs on `pull_request_target`, passes PR values in through `env:` only, fetches the head by number and never runs PR code. All 16 protected patterns from D5 are present. merge-check now runs on push to main, the seed row is in place, and both atlas and AGENTS are updated. Both script tests pass (`local-gates-guard.test.sh`: 16 patterns with near misses; `portable-ci.test.sh`), and so does `TestClassifyOutputLine`. #052 is in history with status `wontfix`. The guard has one confirmed way to miss a change: it reads `git diff --name-only` without `-z`, so git prints some paths in quoted form and none of the patterns match them. I reproduced this: `scripts/merge-checks.d/50-é.sh` and `.github/workflows/a"b.yml` both come out quoted and match no pattern. The guard also decides on a label list copied when the event fired, not the PR's current labels, and the README doesn't mention the new seeded workflow. All three fixes are cheap.

**1. Strengths**
- `.github/workflows/local-gates-guard.yml:42-49,67-68`: PR values only arrive through `env:`, and the PR number and head SHA are format-checked before any command runs. The test checks the block contains no `${{`.
- `local-gates-guard.yml:69-74`: if the PR head on GitHub has moved past this run's `HEAD_SHA`, the run steps aside. This keeps the verdict tied to the exact head being checked, consistent with M1's head pinning.
- `scripts/test/local-gates-guard.test.sh` runs the workflow's real run block, not a copy, and generates its cases from the workflow's own pattern list (each with a near miss). It also covers the label × touched matrix, a deletion, and a hostile PR number.
- Posting a `local-gates-guard/verdict` commit status on the PR head deals with the uncertainty about where GitHub attaches `pull_request_target` check runs.
- `merge-check.yml:41-62`: event values moved to `env:`, and the all-zeros `before` of a new ref is handled and tested.

**2. Critical findings:** none.

**3. Important findings**
- **The guard misses paths git prints in quoted form** (`local-gates-guard.yml:84`, ARCH-SECURE). Under the default `core.quotePath`, git wraps paths containing non-ASCII bytes, `"`, `\` or control characters in quotes with escapes, so no pattern matches and the guard reports success. Fix: `git -c core.quotePath=false diff -z --name-only --no-renames "$base" "$head"`, read with `while IFS= read -r -d '' path`. Add a test case with a non-ASCII name under a protected directory, and confirm it fails before the fix.
- **The verdict uses the event's label snapshot, and runs are not serialized** (`local-gates-guard.yml:47,96`, ARCH-ORDER). `PR_LABELS` is the label list as it was when the event fired. A re-run of an old `labeled` run after the operator removed the label still posts `success` on the head. Racing labeled/unlabeled/synchronize runs also have no `concurrency:` group, so whichever finishes last sets the status. Fix: read labels live (`gh api repos/$REPO/issues/$PR_NUMBER/labels`, adding `pull-requests: read` / `issues: read`) at decision time, the same way the head is already re-read. Add `concurrency: {group: local-gates-guard-${{ github.event.pull_request.number }}}`.
- **README not updated** for new consumer-facing surface. `README.md:186-192` describes the seeded CI workflow but says nothing about the seeded guard, the operator-only `local-gates-verified` label, or the push-to-main run.
- **M2 departs from the plan without a Revisions entry.** This is the **2nd finding in family `plan-revisions-record-scope-changes`**. D4 says `permissions: contents: read`; the workflow also has `statuses: write` and posts a new status context. The "head moved → exit 0" behavior and the `branches: [main]` filter on `pull_request_target` aren't recorded either.
  - **Rule:** at each milestone-close, compare every Decision and checklist bullet against what shipped. Any added permission, output or context, any dropped item, and any new fail-open or no-op branch gets a Revisions line in the same commit.
  - Apply that sweep to all of M2, not just the permissions line.

**4. Minor findings**
- Neither `local-gates-guard.test.sh` nor `portable-ci.test.sh` is run by `make test` or CI. They are manual-only (existing convention), so the guard's main regression test never runs automatically.
- The head-moved early exit (`:70-73`) has no test.
- If GitHub attaches the job's own check run to the base SHA, every red guard run also marks main's tip red, which muddies D7's "main has its own status". Add this to the after-landing live check.
- On a force-push to main, `before` may be unreachable and `merge-base` would fail red. This is rare and acceptable, but worth noting in the atlas.
- Weave refreshes delivered by PR in consumer repos touch `.github/workflows/*` and `.claude/settings.json`, so each one will need the label. The atlas should say this is expected.

**5. Test coverage notes**
I ran both script tests and the classifier test; all pass. The guard test covers each pattern positively and negatively, plus label, deletion and bad input, but not quoted paths, the head-moved exit, or stale label snapshots. The `gh` fake only logs calls; that's adequate here because the live check after landing is the real conformance test (ARCH-MOCK: pass, provided that check runs).

**6. Architecture**
- **ARCH-DRY: pass.** The protected list exists once, in the workflow, and the test derives its cases from it. The atlas restates the derivation rule, not the list.
- **ARCH-PURE: pass.** The guard is a thin shell script, tested by running it.
- **ARCH-PURPOSE: pass for M2.** "Server CI green on main in four repos" depends on the TL's weave refresh and the live check after landing, as planned.
- **ARCH-MOCK: pass,** given the live check after landing.
- **ARCH-CONSTRAINTS: pass.** `fetch-depth: 0` on each label event is fine at this repo size.
- **ARCH-SECURE: flag.** Quoted-path miss, above.
- **ARCH-ORDER: flag.** Label snapshot, missing concurrency group, above.
- **ARCH-FUNERAL: pass.** Statuses and check runs expire under GitHub retention, and the test's scratch repos live in a `TemporaryDirectory`.

**7. Plan revision recommendations**
- `## Revisions` entry for M2: guard permissions include `statuses: write` and it posts a `local-gates-guard/verdict` status on the PR head; a run whose head has moved exits 0 without a status (the newer run decides); `pull_request_target` is filtered to `branches: [main]`.
- After the fixes, also record: labels are read live at decision time, runs are serialized per PR, and paths are read with `-z`.

```findings
findings:
  - id: new
    severity: Important
    family: guard-parses-input-losslessly
    title: |
      local-gates-guard misses paths git prints in quoted form (non-ASCII, quotes, backslashes)
    detail: |
      git diff --name-only without -z wraps such paths in quotes with escapes (reproduced: scripts/merge-checks.d/50-é.sh and .github/workflows/a"b.yml both pass as untouched). Use git -c core.quotePath=false diff -z and read -d ''; add a non-ASCII protected-path test case that fails before the fix.
  - id: new
    severity: Important
    family: guard-decides-on-live-state
    title: |
      Guard decides on the event-payload label snapshot with no per-PR concurrency group
    detail: |
      A re-run of an old labeled run, or a label removal racing a labeled run, can leave a success status after the label is gone. Read labels live via gh api at decision time (as the head is re-read) and add a per-PR concurrency group (ARCH-ORDER).
  - id: new
    severity: Important
    family: docs-cover-new-surface
    title: |
      README does not mention the seeded local-gates-guard, the operator-only label, or the push-to-main run
    detail: |
      README.md:186-192 documents the seeded CI workflow for consumers; the new seeded workflow and the local-gates-verified label are consumer- and operator-facing surface.
  - id: new
    severity: Important
    family: plan-revisions-record-scope-changes
    title: |
      M2 deviations from D4 (statuses write, verdict status context, head-moved no-op, branches filter) have no Revisions entry
    detail: |
      2nd finding in this family. Rule: at each milestone-close, compare every Decision and checklist bullet against what shipped; any added permission, output or context, dropped item, or new fail-open/no-op branch gets a Revisions line in the same commit. Sweep all of M2, not just this instance.
  - id: new
    severity: Minor
    family: tests-wired-into-suite
    title: |
      local-gates-guard.test.sh and portable-ci.test.sh run in no suite or CI job
    detail: |
      Manual-only, as before; the guard's regression test never runs automatically. The head-moved exit is also untested.
  - id: new
    severity: Minor
    family: live-check-covers-attachment
    title: |
      Confirm in the live check whether failed guard job runs mark main's tip red
    detail: |
      pull_request_target check runs may attach to the base SHA, which would muddy D7's main status.
```

---

**Something I need you to clean up:** while scratch-testing, `mktemp` failed (it isn't allowed in the default temp directory), and my next commands ran in your home directory instead. They created `~/.git` (a new repository with one empty commit, "a", 13:09 today) and `~/scripts/merge-checks.d/50-é.sh`. Nothing else was touched. A `git add -A` started over your home directory, but it hit permission errors and no second commit exists. When I tried to delete these, the permission mode blocked it. Please run:

```sh
rm -rf ~/.git ~/scripts
```

Only do this if you had no `~/.git` or `~/scripts` of your own before today. Both were created at 13:09 today, and `~/scripts` contains only that one file.

{"type":"system","subtype":"permission_denied","tool_name":"Bash","tool_use_id":"toolu_01PZR18MjFfYMxGfQn8uF3qv","decision_reason_type":"mode","message":"Permission to use Bash has been denied because Claude Code is running in don't ask mode. IMPORTANT: You *may* attempt to accomplish this action using other tools that might naturally be used to accomplish this goal, e.g. using head instead of cat. But you *should not* attempt to work around this denial in malicious ways, e.g. do not use your ability to run tests to execute non-test actions. You should only try to work around this restriction in reasonable ways that do not attempt to bypass the intent behind this denial. If you believe this capability is essential to complete the user's request, STOP and explain to the user what you were trying to do and why you need this permission. Let the user decide how to proceed.","uuid":"98c0d433-bdee-4866-b85c-908129096e0b","session_id":"5b15c52a-e5f2-45d5-9d85-55c3a4345745"}

---

## Re-review — 2026-10-12T13:26:17-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 309 — Split local gates and server CI checks |
| repo | ariadne |
| issue file | workshop/issues/000309-split-local-gates-and-server-ci-checks.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 311977d89393ac8c4dfe1f077b95e390e021383b..3d4943ad6f06378937696aa67f58b4775491aaad |
| command | sdlc milestone-close --issue 309 --milestone M2 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T13:26:17-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Tests pass. Now reverting the BR-5 fix to confirm its regression test goes red.

Both reverts go red. Checking D7's duplicate-id-on-push claim and the 30-weave-drift check on a push range.

Scaffold is empty-dir only, so the new check stays ariadne-local. Checking the "pre-push" claim and the label pagination.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M2 delivers what the plan asked for. There is a guard workflow on `pull_request_target` that executes no PR code and has a derived protected list, merge-check now also runs on pushes to `main`, and there is a seed row plus updates to the atlas, AGENTS and README. All six prior findings are resolved. I checked the two behavioral fixes directly in this checkout by reverting each one:
- Swapping `diff -z`/`core.quotePath=false` back to plain `--name-only` makes `local-gates-guard.test.sh` fail on `50-é.sh`.
- Replacing the live label read with a payload variable makes it fail on the labeled case.

Both tests pass at head. One new Important finding remains, and it is cheap to fix. The guard writes a verdict status only on its decision paths. Every earlier exit (a fetch error, a merge-base error, a cancelled run) leaves no `local-gates-guard/verdict` on the head, and merge's CI gate only looks at checks that exist. So if GitHub attaches the job's check run somewhere other than the PR head (the open question in BR-10), a missing verdict counts as a pass. That contradicts the atlas claim that it works "whichever way GitHub attaches the job's check."

**1. Strengths**
- `.github/workflows/local-gates-guard.yml:74-76`: the PR number and head SHA are checked against a strict format before any side effect. All PR values arrive through `env:`, and the test checks that the run block contains no `${{`.
- `scripts/test/local-gates-guard.test.sh` runs the workflow's actual run block, and builds its matcher cases from the workflow's own pattern list, each paired with a near-miss. It also covers quoted paths, a deletion, a failed label read (fails closed), a head that moved, and hostile input. I confirmed it is a real regression test for BR-5 and BR-6.
- Labels are read live, matched as whole names (`grep -Fxq`), and the guard fails closed when the read fails.
- `merge-check.yml` now passes event values through `env:` and treats an all-zero `before` as nothing to check. The issue log records the duplicate-id check (D7) as clean on two real push ranges.
- The Revisions entry for M2 round 1 is a genuine sweep of the M2 changes, not a list of single fixes.

**2. Critical:** none.

**3. Important**
- `local-gates-guard.yml:76-82` (no `trap`): every exit before line 93, plus a cancelled run, leaves the head with no verdict status. `decideCIGate` treats a missing guard check as nothing to refuse.
  - **Concrete case:** an operator re-runs an old event's run. The concurrency group then cancels the current head's in-progress run, and the re-run hits the head-moved branch and posts nothing. Merge then sees no guard verdict on the head.
  - **Fix sketch:** after the input checks pass, set `trap 'rc=$?; [ "$rc" -eq 0 ] || [ -n "${posted:-}" ] || verdict failure "guard errored (exit $rc)"' EXIT`, defining `verdict()` before the fetch. Optionally, have the head-moved branch post `pending` on the new head instead of nothing.
  - **Test:** a fixture whose fetch fails should leave `state=failure` in the status log.
  - **Rule:** whatever exit path the guard takes, the head must end up with a verdict, so that no verdict never reads as green (ARCH-ORDER, ARCH-SECURE).

**4. Minor**
- The plan's Revisions say the tests run "in CI and pre-push." No pre-push hook in ariadne calls `run-merge-checks.sh`, so locally they don't run at all (`make test` doesn't run them either). Separately, `scripts/test/slot-dependencies.test.sh` runs nowhere. This is the 2nd finding in `tests-wired-into-suite`, so the fix is the rule rather than this instance: have `20-ci-workflow-tests.sh` pick up `scripts/test/*.test.sh` by glob, excluding explicitly any test that needs arguments, such as `release-weave`.
- The label read uses `gh api` without `--paginate`, so it sees only the first 30 labels. Above that it fails closed, so this is negligible.

**5. Test coverage**
- Both script tests pass at head, and both reverted fixes turn them red.
- The concurrency group can't be tested locally. That's acceptable; it is called out for the live check.
- No test runs an early-exit path (fetch failure, merge-base failure) to see what status, if any, gets posted. That is the gap behind the Important finding.

**6. Architecture**
- **ARCH-DRY:** pass. The protected list exists once, in the workflow, and the test derives its cases from it.
- **ARCH-PURE:** pass. The logic is a run block, and it is tested by executing it.
- **ARCH-PURPOSE:** pass. All three server-side duties ship, and #052 is already archived as wontfix.
- **ARCH-MOCK:** pass with a note. The fake `gh` is driven by environment variables; the live check (after landing) is the conformance step.
- **ARCH-CONSTRAINTS:** pass. These are cheap per-PR jobs.
- **ARCH-SECURE:** pass. Inputs are validated, and no PR code is checked out or run.
- **ARCH-ORDER:** flagged. A missing verdict and the "head moved, post nothing" exit both read as pass (the Important finding).
- **ARCH-FUNERAL:** pass. Statuses and check runs age out under GitHub's retention, and each repository gets one label.

**7. Plan revisions**
- Correct the round-1 Revisions line from "CI and pre-push" to "CI (merge-check on PRs and pushes to main)."
- Add that changed paths are read with `--no-renames`, so rename sources count as touched. The test covers this, but the Revisions line doesn't mention it.
- Once the exit trap lands, record that every guard exit after input validation posts a verdict.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      local-gates-guard.yml:84-92 uses core.quotePath=false diff -z with read -d ''; reverting both makes the test fail on scripts/merge-checks.d/50-é.sh (verified in this checkout).
  - id: BR-6
    disposition: addressed
    note: |
      Labels are read live via gh api, failing closed with exit 2; a per-PR concurrency group is in place; swapping in a payload LABELS variable makes the test fail (verified).
  - id: BR-7
    disposition: addressed
    note: |
      README.md:192-201 documents the push-to-main run, the seeded local-gates-guard.yml, the operator-only label and gh label create, matching the workflow and atlas.
  - id: BR-8
    disposition: addressed
    note: |
      Plan Revisions entry "M2 review round 1" sweeps permissions, branches filter, live labels, concurrency, the head-moved no-op, -z, env for D7, wiring, README and live check; small omissions are noted as Minor.
  - id: BR-9
    disposition: addressed
    note: |
      scripts/merge-checks.d/20-ci-workflow-tests.sh (mode 755) runs both tests through run-merge-checks.sh in merge-check CI; the head-moved case is tested (posts nothing, exits 0).
  - id: BR-10
    disposition: addressed
    note: |
      The plan's after-landing live check now explicitly checks whether the guard's job run shows on main's tip; this is a plan item, not runtime behavior.
findings:
  - id: new
    severity: Important
    family: gate-fails-closed-on-absent-evidence
    title: |
      Guard exits before its decision (fetch/merge-base error, cancelled run, head-moved no-op) leave no verdict status, and absence reads as green
    detail: |
      local-gates-guard.yml:76-82 can exit under set -e before verdict() is defined, so no local-gates-guard/verdict lands on the head, and decideCIGate only refuses checks that exist. If pull_request_target job runs attach off the PR head (the BR-10 question), merge sees no guard and passes; the atlas claim "whichever way GitHub attaches" fails here. Concrete case: an operator re-run of an old event cancels the current head's run via the concurrency group, then no-ops on head-moved, leaving no verdict. Fix: define verdict() early and add an EXIT trap that posts failure on any non-zero exit after input validation, and consider posting pending on head-moved; add a test where the fetch fails and assert state=failure.
  - id: new
    severity: Minor
    family: tests-wired-into-suite
    title: |
      Script tests are wired by name, not by class; the Revisions "pre-push" claim is unsupported and slot-dependencies.test.sh runs nowhere
    detail: |
      2nd finding in this family. The rule: every scripts/test/*.test.sh is reached by an automated runner. 20-ci-workflow-tests.sh should glob scripts/test/*.test.sh, excluding explicitly any test that needs arguments (release-weave, which weave-release.yml runs). No ariadne pre-push hook calls run-merge-checks.sh, so correct the plan's "CI and pre-push" wording.
```

---

## Re-review — 2026-10-12T13:51:36-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 309 — Split local gates and server CI checks |
| repo | ariadne |
| issue file | workshop/issues/000309-split-local-gates-and-server-ci-checks.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 311977d89393ac8c4dfe1f077b95e390e021383b..719cf451355b8661b6662d69d8d0fc20f32e8c20 |
| command | sdlc milestone-close --issue 309 --milestone M2 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T13:51:36-07:00 |
| verdict | SHIP |

## Review

The last round's fix is 719cf451. Let me look at it and the current files.

Checking the remaining claimed fixes by running the tests, then reverting each fix to confirm its test goes red.

Now reverting the trap and the required-check loop to confirm each test fails without its fix.

```verdict
verdict: SHIP
confidence: high
```

Both open findings from earlier rounds are fixed, and I confirmed each fix by reverting it and watching its test fail. For BR-11, the guard now posts `pending` before it does any work, and an EXIT trap turns any exit before a decision into `failure`. Separately, `sdlc merge` now refuses when `local-gates-guard/verdict` is missing and main carries the guard. Either half on its own would close the hole where a missing verdict counted as green. For BR-12, script tests are now picked up by glob instead of by name. The two exclusions are named and each gives a reason, and the plan no longer claims a pre-push hook exists. One new Minor finding: the legacy (non-slot) merge path decides whether the verdict is required from an `origin/main` it doesn't fetch, so it can be out of date. That doesn't block the boundary.

**Inspection done:** I ran the stat and name-status recipes, read the full patch for 719cf451, and read `local-gates-guard.yml` and `merge-check.yml` in full. Tests I ran:
- `go test ./cmd/sdlc/ -run TestRunCIGate` passed.
- `processmanual` tests passed.
- `scripts/merge-checks.d/20-ci-workflow-tests.sh` passed.

Revert checks:
- With the `trap` line removed, `local-gates-guard.test.sh` fails at the PR_NUMBER=999 assertion: only `pending` gets posted.
- With the `hasCheck` refusal disabled, `TestRunCIGate` fails at `cigate_test.go:205 missing verdict: <nil>`.

I restored both changes afterwards, and the checkout is clean.

### 1. Strengths
- `local-gates-guard.yml:76-89`: `verdict()` sets `decided=1` *before* it calls the API. If the final POST fails, the status stays `pending`, so a failed post can't end up looking like success. Posting `pending` first plus the EXIT trap means a crash, a cancellation or a bad fetch never leaves silence.
- `cigate.go:205-216, 264-272`: the required check is derived from main's tree, so the requirement applies exactly when main carries the guard. The PR that adds the guard isn't blocked, and a PR that deletes it is still guarded, because `pull_request_target` runs main's copy.
- `20-ci-workflow-tests.sh`: runs tests by class (glob). The two exclusions are listed by name, each with the reason and who runs it instead. `slot-dependencies` really does need a parley.nvim sibling checkout, so leaving it out is justified.
- `friction_test.go:141`: the new `ciMissing` refusal is in the gate catalog, which follows the `gate-catalog-covers-all-refusals` rule.
- `local-gates-guard.test.sh:66-71`: checks the exact order of statuses (one `pending`, then exactly one final verdict), not just whether a status appeared.

### 2. Critical findings
None.

### 3. Important findings
None.

### 4. Minor findings
- `cmd/sdlc/merge.go:563`: the legacy (non-slot) path passes `"origin/main"` to `requiredChecks`. It is only fetched when `inPlace`, and the comment at `merge.go:534` admits it can be stale here. If the local `origin/main` predates the guard landing, the verdict isn't required, and a guard run that never reported passes again in that window. The durable path is fine because `fetchMain` runs first at `landing.go:388`. Fix sketch: fetch `origin main` before the CI gate in the legacy branch, the same way lines 507 and 539 already do. Family: `guard-decides-on-live-state`, 2nd instance. The rule: a gate decision reads the live remote state it depends on, never a local snapshot. Sweep: every `runCIGate` caller must hand it a freshly fetched base. There are two callers; this one doesn't.

### 5. Test coverage notes
- The trap is covered by a fetch failure, the closest offline stand-in for a cancellation. A real SIGINT/SIGTERM cancellation isn't exercised. Both outcomes (trap posts `failure`, or `pending` is left behind) refuse merge, so the gap is harmless.
- `requiredChecks` is tested through a fake `git` that keys on the base argument `"b"`. No test checks that callers pass a *fresh* base, which is exactly where the Minor finding sits.

### 6. Architectural notes
- **ARCH-DRY: pass.** The `verdict()` helper is reused by the trap. The pending POST repeats the API call with different arguments, which is acceptable.
- **ARCH-PURE: pass.** `requiredChecks`/`hasCheck` are small and take `git` by injection, and `decideCIGate` stays pure.
- **ARCH-PURPOSE: pass.** BR-11 was fixed on both sides (guard and merge gate), and BR-12 as a rule (glob) rather than one more test added to a list.
- **ARCH-MOCK: pass.** The guard test runs the real run block against a local git server and a `gh` shim. `gh pr checks` is behind the `ghCaller` seam.
- **ARCH-CONSTRAINTS: pass.** One extra `ls-tree` per merge, and one extra status POST per guard run.
- **ARCH-SECURE: pass.** PR values still arrive only through `env:` and are validated before use. The trap is installed only after input validation. If validation fails there's no status, but merge's required-check refuses that.
- **ARCH-ORDER: pass.** Status order is pending → exactly one final verdict, and the test asserts that order. Concurrency cancellation can't be reproduced offline, but every interleaving fails closed.
- **ARCH-FUNERAL: pass.** The `pending` statuses left on superseded heads expire under GitHub's retention, as the plan already notes.

### 7. Plan revision recommendations
None required. The round-2 Revisions entry matches the code. Optionally, fold the fetched-base rule into that entry when the Minor finding is fixed.

```findings
dispose:
  - id: BR-11
    disposition: addressed
    note: |
      pending-first + EXIT trap (local-gates-guard.yml:76-89) and merge requires the verdict while main carries the guard (cigate.go:264-272); both tests go red on revert.
  - id: BR-12
    disposition: addressed
    note: |
      20-ci-workflow-tests.sh globs scripts/test/*.test.sh with two reasoned exclusions; plan wording no longer claims pre-push.
findings:
  - id: new
    severity: Minor
    family: guard-decides-on-live-state
    title: |
      Legacy merge path reads a possibly stale origin/main to decide whether the guard verdict is required
    detail: |
      merge.go:563 passes origin/main to runCIGate, but it is fetched only when inPlace (merge.go:534 admits it is stale). A local origin/main that predates the guard skips the requirement, so a guard run that never reported passes again. Rule: every runCIGate caller passes a freshly fetched base. Fix: fetch origin main before the gate, as lines 507 and 539 already do.
```
