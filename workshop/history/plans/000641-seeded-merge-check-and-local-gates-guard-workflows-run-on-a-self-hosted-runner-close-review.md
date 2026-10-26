# Boundary Review — ariadne#641 (whole-issue close)

| field | value |
|-------|-------|
| issue | 641 — Seeded merge-check and local-gates-guard workflows run on a self-hosted runner (CI_RUNS_ON), split from #637 for echo#51 |
| repo | ariadne |
| issue file | workshop/issues/000641-seeded-merge-check-and-local-gates-guard-workflows-run-on-a-self-hosted-runner.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4bd1ef6e2c0acffa462c2b83543b8c52393e7bac..f8e342c2f3147c45f0f86876ed64f06c7b244b2f |
| command | sdlc close --issue 641 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T22:54:39-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Tests pass; checking one last thing — whether the consumer Brewfile install path runs on self-hosted.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The window delivers S1–S5 as specified. Both seeds take `runs-on` from `vars.CI_RUNS_ON` and fall back to `ubuntu-latest` when it is unset. setup-homebrew and all three setup-go steps carry the `runner.environment == 'github-hosted'` guard. The S3 tool-check step and the S4 weave-before-bootstrap step exist and run only when the runner is not GitHub-hosted. ariadne source CI refuses a non-hosted runner and names the reason. I ran `portable-ci.test.sh` and `local-gates-guard.test.sh` at the pinned head and both pass. The harness change is real: it parses every step, including `uses:` ones, and runs a strict condition evaluator. That evaluator fails the harness on any condition it cannot model, instead of treating it as true. It then runs the actual blocks under both runner models. Two things stand before a clean SHIP. README does not mention the new `CI_RUNS_ON` variable, though the README section on seeded workflows already tells operators which `gh` setup to do. And S6 says `local-gates-guard.test.sh` asserts its runs-on, but that file is unchanged; the assertion lives in `portable-ci.test.sh` instead.

1. **Strengths**
   - `scripts/test/portable-ci.test.sh:24-31`: `holds()` only accepts `==`/`!=` conjunctions over `runner.environment` and gateway outputs, and anything else trips an assertion. The old exact-string skip list could go wrong without failing; this cannot.
   - `portable-ci.test.sh:38-44`: structural guards are derived from the workflow, not listed by hand. Every `setup-homebrew@`/`setup-go@` step must carry the hosted guard, exactly two steps may be not-hosted, and those two may not run paths from the tree. The `ref:` ban covers both seeds.
   - Self-hosted cases cover the main paths:
     - the same event rows as the hosted run;
     - only `actions/checkout` is taken;
     - the exact `GITHUB_ENV` contents;
     - a missing `go` refuses with its name;
     - a weave without `dependencies` refuses before bootstrap;
     - source CI refuses.
   - The hosted run now asserts that `GITHUB_ENV` stays empty and that both setup actions are taken (`:181`), so unchanged hosted behaviour is pinned rather than assumed.
   - `merge-check.yml:56-61`: refusing source CI off hosted runners, instead of silently skipping `brew bundle`, follows the spec's reasoning and fails loudly.

2. **Critical findings:** none.

3. **Important findings**
   - **README update appears missing for `CI_RUNS_ON`** (README.md:582-588). `CI_RUNS_ON` is a new per-repository config key that the operator sets with `gh variable set`. README's seeded-workflow section already carries operator setup of this kind (`gh label create local-gates-verified`), but nothing says the runner can be chosen. The only mention is in `atlas/workflow/ci-merge-check.md:115`. Fix: add one sentence near README:575-588 naming `CI_RUNS_ON`, giving the JSON example (`["self-hosted","local-ci"]`), and saying that a self-hosted runner installs nothing, so go, git and python3 must already be on the host. Link the atlas section. Family instances in this window: just this key. No other new user-facing surface was added.

4. **Minor findings**
   - S6 says "`scripts/test/local-gates-guard.test.sh` asserts its runs-on", but that file is untouched in the window. The assertion for both seeds is at `portable-ci.test.sh:35-37`. The behaviour is covered, but the Spec claim is wrong: either amend the Spec or move or duplicate the assertion. Family instances: only this clause.
   - The `NO_DEPENDENCIES` case checks `'weave' in out` (`portable-ci.test.sh:189`), which is a weak match because many messages contain "weave". Matching `'no weave with dependencies'` would pin the S4 message itself.
   - Neither the atlas nor the workflow comments warn that `merge-check` on `pull_request` runs PR code on the operator's machine once `CI_RUNS_ON` is set. That is fine for private repositories like echo, but it is a known risk for any public consumer that accepts fork PRs, so a one-line caution in the atlas section would help.

5. **Test coverage notes:** Hosted and self-hosted are both exercised for every Done-when clause that names two modes: runs-on (structural), the host-mutating guards (structural plus the `uses` taken in each mode), the tool checks (pass and fail), and the source-CI refusal (self-hosted refuses; existing hosted source-CI cases still pass). A malformed `CI_RUNS_ON` failing the run is GitHub's own `fromJSON` behaviour and cannot be tested locally. That is acceptable. The implementor's mutation claims in the Log (dropping the refusal or the homebrew guard turns the test red) match what the asserts at `:39-41` and `:191` would catch.

6. **Architectural notes**
   - **ARCH-DRY: pass.** The runs-on expression appears in both workflows because GitHub Actions has no shared-job include for seeded single files. The harness pins both against one constant (`RUNS_ON`, `:34`), which serves as the single source.
   - **ARCH-PURE: pass.** `holds()` is a pure evaluator, and the subprocess runs stay in the harness shell.
   - **ARCH-PURPOSE: pass.** Both consumers, the two seeds, are delivered; nothing that is the point of the issue was deferred.
   - **ARCH-EVOLVE: pass.** No persisted formats changed. `GITHUB_ENV` is per-run, and bootstrap.sh is unchanged as S4 requires.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry: "S6: the runs-on assertion for local-gates-guard.yml lives in portable-ci.test.sh (both seeds against one constant); local-gates-guard.test.sh is unchanged." Alternatively, add the assertion to that test.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: readme-missing-new-config-surface
    title: |
      README does not mention the new CI_RUNS_ON repository variable for the seeded workflows
    detail: |
      README.md:575-588 covers seeded-workflow operator setup (gh label create) but not CI_RUNS_ON. It appears only in atlas/workflow/ci-merge-check.md:115. Add one sentence: set it with gh variable set to a JSON runs-on value; a self-hosted runner installs nothing, so go, git and python3 must already be on the host. Link the atlas section. It is the only new config key in the window.
  - id: new
    severity: Minor
    kind: defect
    family: spec-claim-matches-delivered-test-site
    title: |
      S6 says local-gates-guard.test.sh asserts its runs-on, but that file is unchanged
    detail: |
      The assertion for both seeds is in scripts/test/portable-ci.test.sh:35-37. Behaviour is covered; amend the Spec through a Revisions entry or move the assertion. It is the only such clause.
  - id: new
    severity: Minor
    kind: defect
    family: weak-failure-message-assert
    title: |
      The NO_DEPENDENCIES self-hosted case matches only the word weave in output
    detail: |
      portable-ci.test.sh:189 checks 'weave' in out; match the S4 message ('no weave with dependencies') so the test pins that step's refusal rather than any weave-mentioning output.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a00f25a6-1bc8-4826-aa62-61ec6eda5302","repository":"github.com/xianxu/ariadne","issue_id":"000641","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/73565c3c01965a74","reviewed_head":"f8e342c2f3147c45f0f86876ed64f06c7b244b2f","slot":"ariadne:2","requested_at":"2026-10-26T05:52:52.181061Z"},"phase":"finished","pid":73512,"started_at":"2026-10-26T05:52:52.28769Z","finished_at":"2026-10-26T05:54:37.664069Z","session_id":"3c60f241-e293-4bf4-acb8-ff86a5d1630c","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T23:00:17-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 641 — Seeded merge-check and local-gates-guard workflows run on a self-hosted runner (CI_RUNS_ON), split from #637 for echo#51 |
| repo | ariadne |
| issue file | workshop/issues/000641-seeded-merge-check-and-local-gates-guard-workflows-run-on-a-self-hosted-runner.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4bd1ef6e2c0acffa462c2b83543b8c52393e7bac..1994b7b73ae79590f15202f3d8dd8f52c6b945d3 |
| command | sdlc close --issue 641 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T23:00:17-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This change does what #641 asks. Both seeded workflows read `runs-on` from `vars.CI_RUNS_ON` and default to `ubuntu-latest`. Every setup-homebrew and setup-go step runs only on GitHub-hosted runners. The two self-hosted checks (S3 for host tools, S4 for the pinned weave) are in place. ariadne's own source CI refuses to run off a GitHub-hosted runner and says why. At the pinned head, `scripts/test/portable-ci.test.sh` passes (exit 0) and `scripts/test/local-gates-guard.test.sh` passes (exit 0). The portable-ci test now evaluates step conditions over a runner model and runs the real blocks under both runner types. All three round-1 findings are fixed. The only new findings are two Minor doc inaccuracies, mostly in the README paragraph added for BR-1. Neither blocks shipping.

Done-when: current — the first two clauses (runs-on from the variable with the default, hosted-only setup steps, self-hosted checks, the source-CI refusal, both harnesses green) are exactly what this diff builds and tests. The third clause (landed with the label, merge commit sent to echo:2) happens after merge and can't be checked in this window. The Revisions entry for S6 doesn't change what Done-when needs to cover.

1. **Strengths**
   - `scripts/test/portable-ci.test.sh:24-31`: `holds()` replaces the old exact-string condition matches with a small evaluator. Any condition it can't parse fails the harness instead of being skipped silently, so a future guard can't get past the runner model unnoticed.
   - `portable-ci.test.sh:38-40`: there is one structural rule — any `setup-homebrew@`/`setup-go@` step must carry the hosted guard. A new setup step therefore can't be added without it.
   - The self-hosted run checks that only `actions/checkout@v4` is taken and that `GITHUB_ENV` holds exactly `GOTOOLCHAIN=local` and `UV_PYTHON_DOWNLOADS=never`. The failure cases cover a missing go (using a PATH without go), a weave that lacks `dependencies` (the run stops after the pinned build and before bootstrap), and source CI refusing to run.
   - `merge-check.yml:56-61`: rather than silently skipping `brew bundle`, source CI refuses to run on a non-hosted runner and names why. Skipping it would have built the candidate without its tools.

2. **Critical findings** — none.

3. **Important findings** — none.

4. **Minor findings**
   - **`README.md:593` says the host needs "the pinned weave".** It doesn't: the job fetches and builds the pinned weave itself (`merge-check.yml:126-135`, "Build first pinned compiler", added to `GITHUB_PATH`). The host only needs a Go new enough for the pin's go.mod. As written, the sentence could lead an operator to `brew install` weave on the host, the very kind of host install this issue avoids. The atlas wording ("requires the pinned weave to answer") is correct.
     - Same family, also in this window: the comment at `merge-check.yml:16-18` and the atlas text both say "a malformed value fails the run, never falls back". That's true for invalid JSON. But valid JSON that evaluates as falsy (`0`, `false`, `""`, `null`) makes `&& … ||` fall back to `ubuntu-latest`. Either narrow the claim to "invalid JSON" or accept the edge case knowingly.
     - The rule covering both: every operator-facing claim about what the workflow does or requires should be checked against the step that implements it.
   - `atlas/workflow/ci-merge-check.md:118`: "(echo: `[…]`" reads like a typo. "e.g., echo uses" would be clearer.

5. **Test coverage notes**
   - BR-3 is fixed: `portable-ci.test.sh:190` now matches `'no weave with dependencies'`, the exact S4 message, and also checks that the rows stop after the pinned build.
   - The Done-when items that name two modes are exercised in both: hosted runs check that both setup actions are taken and `GITHUB_ENV` stays empty; self-hosted runs check the opposite.
   - Not covered: falsy-JSON values of `CI_RUNS_ON`. `runs-on` is evaluated by GitHub, so the harness can't test it. That's acceptable once the doc claim is narrowed.

6. **Architectural notes**
   - ARCH-DRY: pass. Both YAMLs have to repeat the runs-on expression, and the test checks both against one `RUNS_ON` constant rather than keeping a second copy in `local-gates-guard.test.sh`. The S4 check repeats bootstrap.sh's own weave test because the Spec calls for exactly that.
   - ARCH-PURE: pass. The condition evaluator `holds()` is a pure function, and IO stays in the step runner.
   - ARCH-PURPOSE: pass. Both seeds and all four host-mutating steps are guarded, and the source-CI case is refused rather than left out.
   - ARCH-EVOLVE: pass. No persisted format changes. `CI_RUNS_ON` is an additive config key with the old behaviour as its default.

7. **Plan revision recommendations** — none. The round-1 Revisions entry already makes S6 match where its runs-on assertion lives.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:590-596 now documents CI_RUNS_ON with gh variable set, host tools, the fork-PR caution, and links the atlas section; one inaccuracy in it is raised separately as Minor.
  - id: BR-2
    disposition: addressed
    note: |
      The Revisions entry records that the runs-on assertion for both seeds lives in portable-ci.test.sh:33-37 against one RUNS_ON constant; local-gates-guard.test.sh is unchanged and green.
  - id: BR-3
    disposition: addressed
    note: |
      portable-ci.test.sh:190 now asserts 'no weave with dependencies' in out plus rows stopping after the pinned build; test passes at head.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: operator-doc-claim-matches-workflow-step
    title: |
      README says a self-hosted host must already have the pinned weave; the job builds it itself
    detail: |
      README.md:593 "along with the pinned weave" conflicts with merge-check.yml:126-135, which fetches and builds the pin and adds it to GITHUB_PATH; the host needs only a Go new enough for the pin's go.mod, and the sentence invites a host brew install of weave. Same family in this window: merge-check.yml:16-18 and atlas ci-merge-check.md:119 say a malformed CI_RUNS_ON "never falls back", but falsy valid JSON (0, false, "", null) falls back to ubuntu-latest through the && || expression. Rule: check each operator-facing claim against the step that implements it.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e503ea9c-aa4e-43e4-bf8a-86f868e2b744","repository":"github.com/xianxu/ariadne","issue_id":"000641","recipe":"small-diff-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/772e84f519c51598","reviewed_head":"1994b7b73ae79590f15202f3d8dd8f52c6b945d3","slot":"ariadne:2","requested_at":"2026-10-26T05:58:50.878494Z"},"phase":"finished","pid":2805,"started_at":"2026-10-26T05:58:50.982434Z","finished_at":"2026-10-26T06:00:16.626316Z","session_id":"bb993921-c09d-4d30-9f1a-13419c1b0746","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T23:14:47-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 641 — Seeded merge-check and local-gates-guard workflows run on a self-hosted runner (CI_RUNS_ON), split from #637 for echo#51 |
| repo | ariadne |
| issue file | workshop/issues/000641-seeded-merge-check-and-local-gates-guard-workflows-run-on-a-self-hosted-runner.md |
| boundary | whole-issue close |
| milestone | — |
| window | 03240ca21cfbe668d9327166e002bc62e08904fd..b67a69bf9e31ddbaecffdfb30c2aa36a349c778f |
| command | sdlc close --issue 641 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T23:14:47-07:00 |
| verdict | SHIP |

## Review

I've checked the fix diff and the sibling claims across the workflows, README and atlas. Writing the verdict now.

```verdict
verdict: SHIP
confidence: high
```

This window changes only one workflow comment, the README and the atlas, plus the issue log, frontmatter and the review/gate plan files. Both parts of BR-4 are fixed. The README no longer says the host needs the pinned weave. It now says the host needs a Go new enough for the pin's go.mod and that the job builds weave itself. That matches `merge-check.yml`, where `build.sh` builds with `GOTOOLCHAIN=local` and the build step appends `bin-pin1` to `GITHUB_PATH`. The "never falls back" claim is gone from both the `merge-check.yml` comment and atlas `ci-merge-check.md`. Both now say invalid JSON fails and a falsy value (`0`, `false`, `""`, `null`) reads as unset. That is how `vars.CI_RUNS_ON && fromJSON(vars.CI_RUNS_ON) || 'ubuntu-latest'` behaves: the variable is a non-empty string, so it passes the first test, then `fromJSON` returns a falsy value and the `||` falls through to `ubuntu-latest`.

I swept `.github`, README, atlas and scripts for other claims in this family (`CI_RUNS_ON`, "falls back", "malformed", "pinned weave"). The comment at `local-gates-guard.yml:42-44` makes no fallback claim. The atlas sentence "requires the pinned weave to answer" describes a step that runs after the job has built weave, so it is accurate. The class is covered, not just the site BR-4 named. The window does not touch `## Spec`, `## Revisions` or `## Done when`, so there is no scope change.

Done-when: current — the window changes no behavior; it fixes docs and a comment, so the existing Done-when still covers what was built.

1. **Strengths**
   - The README host-requirements sentence (`README.md:593-594`) now matches the build step and names the real constraint: a Go new enough for the pin's go.mod.
   - The falsy-value wording now appears identically in the workflow comment and the atlas, so they can't drift apart.
   - The fix covers the whole family (the weave claim and the fallback claim), not only the cited line.
2. **Critical:** none.
3. **Important:** none.
4. **Minor:** `README.md:594` is longer than the surrounding lines because it wasn't rewrapped. Cosmetic only.
5. **Test coverage:** these are prose and comment changes, checked against the steps they describe. The existing `portable-ci.test.sh` behavior tests still cover what the docs describe. No new test is needed.
6. **Architecture** (on this window only):
   - ARCH-DRY: pass. The two docs now state the same fact the same way.
   - ARCH-PURE: n/a, no code changed.
   - ARCH-PURPOSE: pass. Every site in the family was fixed, not just the named one.
   - ARCH-EVOLVE: n/a, no persisted format changed.
7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      README.md:593-594 now says the job builds the pinned weave and the host needs Go new enough for the pin's go.mod, which matches merge-check.yml build.sh (GOTOOLCHAIN=local) and the GITHUB_PATH append. merge-check.yml:18-19 and atlas ci-merge-check.md:119-120 now say invalid JSON fails and falsy values read as unset, which matches the && || expression. A sweep of .github, README, atlas and scripts found no remaining sibling claims; the local-gates-guard.yml comment makes no fallback claim.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a6fa6b0e-f466-4fb6-aacd-a3de5af33758","repository":"github.com/xianxu/ariadne","issue_id":"000641","recipe":"fix-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/b2297b5bb521d0cf","reviewed_head":"b67a69bf9e31ddbaecffdfb30c2aa36a349c778f","slot":"ariadne:2","requested_at":"2026-10-26T06:14:15.730055Z"},"phase":"finished","pid":46378,"started_at":"2026-10-26T06:14:15.828408Z","finished_at":"2026-10-26T06:14:46.763231Z","session_id":"fec59edb-6eaa-4bb1-991e-961491cef9f0","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T23:22:23-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 641 — Seeded merge-check and local-gates-guard workflows run on a self-hosted runner (CI_RUNS_ON), split from #637 for echo#51 |
| repo | ariadne |
| issue file | workshop/issues/000641-seeded-merge-check-and-local-gates-guard-workflows-run-on-a-self-hosted-runner.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9be56ece84983c4be5c627be73b4f81960c79ca3..756fbd92d1782e9f5ef0e993ae46843ae85ead65 |
| command | sdlc close --issue 641 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T23:22:23-07:00 |
| verdict | SHIP |

## Review

Wrapping up: the fix test passes locally; I'm confirming the stray glob-named file is gone and then writing the verdict.

```verdict
verdict: SHIP
confidence: high
```

The fix stays inside the test harness. `scripts/test/portable-ci.test.sh` builds the missing-go PATH from three parts: the fixture commands without go, symlinks of every entry in `/usr/bin` and `/bin` except `go`, and nothing else. The run harness now takes a `path=` argument, so it no longer string-replaces inside `PATH`. Each step rebuilds `PATH` from `GITHUB_PATH` followed by `base_path`, so the self-hosted case can't pick up a system go. That was the cause of the hosted merge refusal recorded in the Log. The test passes at the pinned head on this host (`PASS CI actual blocks…`, rc=0). The stray issue file whose name was an unexpanded glob, created by commit f085a20cf, is removed in commit 756fbd92 and doesn't exist at head. The window doesn't change `## Spec`, `## Revisions` or `## Done when`; the issue diff is front-matter, one Log entry and the closed line.

Done-when: current. The interdiff only hardens one harness case so it actually models a host without go. It adds no behavior and no contract surface.

1. **Strengths**
   - Replacing `env['PATH'].replace(str(commands),str(tools))` with an explicit `base_path` (`portable-ci.test.sh:164`, `:180`) removes a fragile substring substitution. The PATH used for the case is now explicit rather than derived.
   - The case now models what it claims, a host without go, instead of depending on the image not having one. That makes it portable across the macOS and Ubuntu hosts it runs on.
   - The Log records the root cause and why it passed locally (macOS has no system go).

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `portable-ci.test.sh:151-153`: the dedupe uses `Path.exists()`, which follows symlinks. On a usrmerge Linux host `/bin` is the same directory as `/usr/bin`. There, a dangling symlink in `/usr/bin` would first be linked as a dangling link and then re-attempted on the `/bin` pass. That second `symlink_to` raises `FileExistsError`. Use `os.path.lexists` or `is_symlink() or exists()` instead. This host has no dangling links and no name overlaps, so I could not show it failing.
   - The window has no record of a hosted run going green after the fix. I couldn't revert the fix and see the test go red, because this macOS host has no system go. The hosted CI rerun is the real evidence.

5. **Test coverage notes:** the change is to the test itself. All existing cases still pass at the pinned head, including the hosted and self-hosted rows, `NO_DEPENDENCIES` and the source-CI refusal.

6. **Architectural notes**
   - ARCH-DRY: pass. The ad-hoc `tools=` substitution became one `path=` parameter, with no duplication.
   - ARCH-PURE: pass. The change is harness fixture setup, with no product logic.
   - ARCH-PURPOSE: pass. It fixes the class (any system go leaking into the missing-go case), not only the Ubuntu instance, because every system directory is filtered.
   - ARCH-EVOLVE: pass. No persisted format changes.

7. **Plan revision recommendations:** none.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: test-fixture-host-portability
    title: |
      The missing-go PATH dedupe uses Path.exists(), which follows symlinks, so a dangling /usr/bin link crashes on usrmerge hosts
    detail: |
      portable-ci.test.sh:151-153 links /usr/bin, then /bin. Where /bin is the same directory as /usr/bin, a dangling entry is linked once, exists() then returns False, and the second symlink_to raises FileExistsError. Use os.path.lexists (or is_symlink() or exists()) for the dedupe check.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9afa2bef-4531-4cd4-a843-63d43e3a93a2","repository":"github.com/xianxu/ariadne","issue_id":"000641","recipe":"fix-review","round":4,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/44d5fb1140a731ab","reviewed_head":"756fbd92d1782e9f5ef0e993ae46843ae85ead65","slot":"ariadne:2","requested_at":"2026-10-26T06:21:18.963087Z"},"phase":"finished","pid":75166,"started_at":"2026-10-26T06:21:19.054009Z","finished_at":"2026-10-26T06:22:22.327857Z","session_id":"53ef5a6f-9e1e-4577-ae6a-0136f3f1bc0f","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T23:26:59-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 641 — Seeded merge-check and local-gates-guard workflows run on a self-hosted runner (CI_RUNS_ON), split from #637 for echo#51 |
| repo | ariadne |
| issue file | workshop/issues/000641-seeded-merge-check-and-local-gates-guard-workflows-run-on-a-self-hosted-runner.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2c262cf88862e1cf3a1670c7e92128f2f90d067b..fd6e197055255b92c22b1468e30ae3d38afb9b27 |
| command | sdlc close --issue 641 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T23:26:59-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This window makes one code change: a one-line fix to a test fixture at `scripts/test/portable-ci.test.sh:153`. The other changes are gate and review ledger records in `workshop/plans/` and a close-log entry in the issue file. The fix replaces `(nogo/tool).exists()` with `os.path.lexists(nogo/tool)`. That fully addresses BR-5. `os` is already imported in the embedded Python, and the full harness passes on the pinned head (`PASS CI actual blocks: …`, rc=0). Nothing blocks SHIP.

Done-when: current — the window changes no Spec, Revisions or Done-when text. The issue-file diff only adds a close-log entry and a Log line about this fix, so the Done-when written for the self-hosted CI_RUNS_ON workflows still covers the build.

1. **Strengths**
   - `scripts/test/portable-ci.test.sh:153`: the fix covers the whole class. `lexists` treats any entry that is already present as present, whether it is a dangling link, a valid link or a file. That covers merged-/usr hosts and partial overlaps between `/usr/bin` and `/bin`, not just the one case the finding named.
   - I reproduced the scenario in a scratch directory: `bin` was a symlink to `usr`, and `usr` held a dangling link. The old `exists()` check raised `FileExistsError`; `lexists` completed. So the fix is reachable and it is what fixes the bug.
   - The change is minimal and matches the style of the surrounding fixture code.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none.

5. **Test coverage notes**
   - This is a fix to test-harness code, so the harness running cleanly on a normal host is the evidence. I didn't ask for a dedicated test that fails without the fix, because that would need a merged-/usr host with a dangling system link. The scratch reproduction above shows the old code fails and the new code doesn't.

6. **Architectural notes (interdiff only)**
   - ARCH-DRY: pass. No duplicated logic was added.
   - ARCH-PURE: pass. The change only touches a test fixture; no business logic moved.
   - ARCH-PURPOSE: pass. The fix covers the whole class of "entry already present", not only the instance BR-5 named.
   - ARCH-EVOLVE: pass. No persisted format changed (the ledger files are append-only records).

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      portable-ci.test.sh:153 now uses os.path.lexists; a scratch merged-usr repro raises FileExistsError with exists() and passes with lexists; the full harness passes on fd6e1970.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b317fa68-6fc0-43ec-bc0c-cbf65d2eff0e","repository":"github.com/xianxu/ariadne","issue_id":"000641","recipe":"fix-review","round":5,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/ea314fd290f76ff6","reviewed_head":"fd6e197055255b92c22b1468e30ae3d38afb9b27","slot":"ariadne:2","requested_at":"2026-10-26T06:26:19.197925Z"},"phase":"finished","pid":2439,"started_at":"2026-10-26T06:26:19.307431Z","finished_at":"2026-10-26T06:26:57.903795Z","session_id":"113a7028-93e6-44df-bbc2-1552c85b84f6","outcome":"returned"}] -->
