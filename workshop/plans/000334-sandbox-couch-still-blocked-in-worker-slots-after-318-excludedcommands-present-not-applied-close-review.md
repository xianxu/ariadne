# Boundary Review — ariadne#334 (whole-issue close)

| field | value |
|-------|-------|
| issue | 334 — sandbox: couch still blocked in worker slots after #318 (excludedCommands present, not applied) |
| repo | ariadne |
| issue file | workshop/issues/000334-sandbox-couch-still-blocked-in-worker-slots-after-318-excludedcommands-present-not-applied.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8f9ea913af4e10c119ecf8950e6a1efc1895cc0f..9a96dc9cd874d328d94a3d011ab161bb73760f90 |
| command | sdlc close --issue 334 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T23:53:22-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I'm checking that the guard test fails against the pre-fix settings, then I'll write the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Verdict: fix then ship.** The cause and the fix are right. Claude Code reads `excludedCommands` entries as `Bash(...)` rules, so `"couch"` only exempts `couch` typed with no arguments. Writing the entries as `X *` fixes that at the source. The new guard test is real: I swapped in the base version of `.claude/settings.ariadne.json` and it failed with three errors, then I restored the file and it passed again. The live check in the Log includes negative controls (`cd` into another directory, full path, `$(...)`), which match the docs. One cheap thing should be fixed first. Making `sdlc` run unsandboxed makes a statement in another atlas page out of date, and that page describes how the review sandbox degrades. The other findings are minor: the test only covers two verbs, the atlas overstates one live check, bypass-permissions mode wasn't checked live, and two descriptions are stale.

1. **Strengths**
   - The diagnosis rules out the cheap hypotheses with evidence before fixing anything. The Log also records that edited settings apply in a running session with no restart, so no `couch --prepare` restart change was needed.
   - The test (`cmd/weave/internal/settingsx/base_settings_test.go:17-63`) checks the shipped settings file as plain data, with no mocks. It ties the rule to `permissions.allow`, so a verb that is pre-allowed with arguments but excluded only by its bare name fails. It goes red without the fix (checked).
   - `atlas/workflow/sandbox.md:16-25` gives agents a usable rule: call the binary as a plain leading command, with a list of forms that don't match. It also removes the old "treat the exclusion as unreliable" caveat, which this fix now explains.
   - Leaving `git`, `gh`, `docker`, `npm` and `npx` bare is called out as a separate decision rather than widened quietly.

2. **Critical:** none.

3. **Important**
   - **`atlas/workflow/sdlc-binary.md:963-965` is now out of date** (ARCH-PURPOSE shadow-sweep). It says "Inside Claude Code's Bash sandbox, where slot agents run sdlc, `sandbox_apply` is refused, so those rounds run clone-only." That held only because the `"sdlc"` exclusion never matched. With `sdlc *`, a plain leading `sdlc …` runs unsandboxed, so the reviewbox fence probe can now succeed. Fix: reword it to say sdlc normally runs outside the sandbox (#334), and that the clone-only fallback applies only when sdlc is called in a form that stays sandboxed.

4. **Minor**
   - The test hard-codes only `sdlc` and `couch` (`base_settings_test.go:58`). If `weave`, `datatype`, `doc-review`, `vocabulary` or the `pair*` entries go back to bare names, nothing fails, because none of them are in `permissions.allow`. A sturdier check: list `cmd/*` and require `X *` for each binary.
   - `atlas/workflow/sandbox.md:19-23` says "verified live" for the whole list. The Log only verified `cd <other>`, full path and `$(...)`; redirects and "chained after another command" weren't tested.
   - Bypass-permissions mode wasn't tested live, though Done-when asks for both modes. The Log says so and argues from the docs; acceptable, but the close note should say this was inferred, not observed.
   - The `$comment` in `.claude/settings.ariadne.json:2` still names only sdlc and couch as unsandboxed.
   - The atlas sentence "they are in `permissions.allow` as `Bash(sdlc:*)`/`Bash(couch:*)`" reads as if it covers the whole list. Only sdlc and couch are pre-allowed, so the others still go through the auto-mode classifier.

5. **Test coverage:** the guard pins the class of bug this issue shipped (a bare-name exact match for a verb used with arguments), for the verbs that are pre-allowed. How Claude Code itself matches the entries can't be tested in this repo; the documented live check in the Log covers it.

6. **Architecture**
   - **ARCH-DRY:** pass. One source (`settings.ariadne.json`) is merged by weave.
   - **ARCH-PURE:** pass. The test is a pure read of a file in the repo.
   - **ARCH-PURPOSE:** flagged. `sdlc-binary.md` still assumes the old behaviour (Important above). The binary list is maintained by hand, which only partly derives from `cmd/` (Minor above).
   - **ARCH-MOCK:** N/A. No new external calls; matching behaviour is checked live, as recorded.
   - **ARCH-CONSTRAINTS:** N/A. This is a config change and adds no runtime load.
   - **ARCH-SECURE:** pass, with a note. `weave *` and `pair *` now run unsandboxed, so `weave compile` can rewrite `.claude/settings.json`, which the sandbox otherwise denies. They aren't in `permissions.allow`, so the permission flow and classifier still gate them. The operator approved the wider scope (964c47a1). Worth one sentence in the atlas.
   - **ARCH-ORDER:** N/A. Nothing holds state between events.
   - **ARCH-FUNERAL:** N/A. Nothing new and durable is created; it is a test file plus edits to existing files.

7. **Plan revisions:** add a Revisions entry noting that bypass-permissions mode was verified from the docs, not live, so the "both modes" line in Done-when is recorded honestly.

```findings
findings:
  - id: new
    severity: Important
    family: consumer-docs-stale-after-behavior-change
    title: |
      atlas/workflow/sdlc-binary.md:963 still says slot agents run sdlc inside the Bash sandbox (fence refused, clone-only)
    detail: |
      With "sdlc *" excluded, a plain leading sdlc call is unsandboxed, so the reviewbox fence probe can now succeed. Reword it to cover only sdlc calls that stay sandboxed (#334).
  - id: new
    severity: Minor
    family: guard-test-covers-subset
    title: |
      TestBaseSettingsExcludedVerbsMatchArguments hard-codes only sdlc/couch; weave/datatype/doc-review/vocabulary/pair* can revert to bare names unnoticed
    detail: |
      Derive the ariadne list from cmd/* (and keep the pair list explicit) so every exempted workbench binary must appear as "X *".
  - id: new
    severity: Minor
    family: doc-overclaims-verification
    title: |
      sandbox.md says "verified live" for redirect and chained forms, but the Log only tested cd-other-dir, full path and $(...)
  - id: new
    severity: Minor
    family: done-when-partially-inferred
    title: |
      Done-when asks for bypass-permissions mode, but that mode was argued from the docs, not checked live
  - id: new
    severity: Minor
    family: stale-inline-comment
    title: |
      settings.ariadne.json $comment, and the atlas line on permissions.allow, still describe only sdlc/couch
```

---

## Re-review — 2026-10-11T23:59:55-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 334 — sandbox: couch still blocked in worker slots after #318 (excludedCommands present, not applied) |
| repo | ariadne |
| issue file | workshop/issues/000334-sandbox-couch-still-blocked-in-worker-slots-after-318-excludedcommands-present-not-applied.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8f9ea913af4e10c119ecf8950e6a1efc1895cc0f..eb685d1185db420c0d754dfa20d5a9042d16fa84 |
| command | sdlc close --issue 334 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T23:59:55-07:00 |
| verdict | SHIP |

## Review

I'm checking the guard test by reverting one entry to see it fail, then I'll write up the verdict.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** All five round-1 findings are fixed. The root cause is right: Claude Code reads `excludedCommands` entries as `Bash(...)` rules, where a bare word matches only that word with no arguments, so `"couch"` never covered `couch --send-to …`. The fix writes every ariadne and pair binary as `X *` in `.claude/settings.ariadne.json`. The guard test passes at the head commit. One thing I couldn't run: the read-only mode refused my revert-and-retest of the settings file. The test code clearly fails for a bare entry, and the Log reports it was red on the pre-fix base. I found one new Minor issue: the atlas sentence that describes the guard test still covers only its first half. That doesn't block.

**1. Strengths**
- The cause is backed by the docs' exact-match rule and checked live. The Log has positive calls and three failing negative controls (`cd` into another directory, the full path, `$(...)`). It also explains why a `cd` into the current directory passed.
- The guard test takes ariadne's binaries from the `cmd/` directories, so a new binary has to be listed as `X *` (`cmd/weave/internal/settingsx/base_settings_test.go:58-74`). It also checks every pre-allowed verb against its exclusion entry.
- `atlas/workflow/sandbox.md` gives agents a concrete calling rule (plain leading command; no `cd &&`, pipe, `$(...)`, redirect or full path). It says honestly which forms were tested live and which come only from the docs.
- The trust note is good (ARCH-SECURE): `weave *` runs unsandboxed but isn't pre-allowed, so the permission flow still gates `weave compile` rewriting `.claude/settings.json`.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- `atlas/workflow/sandbox.md:25` says the guard test only checks verbs that `permissions.allow` pre-allows. Since the BR-2 fix, it also requires `X *` for every `cmd/*` binary and the listed pair binaries. The test's own header comment (`base_settings_test.go:11-17`) has the same gap.
  - This is the 2nd finding in family `consumer-docs-stale-after-behavior-change`, so I'm stating the rule rather than asking for an instance fix: when a fix round changes a test's or binary's behavior, grep its name across `atlas/` and code comments and update every description in that same round.
  - The sweep here is small: the test name appears in one atlas line plus the test's header comment. The issue file and gate ledger only cite it.

**5. Test coverage**
- I read the test and ran it at the head commit: it passes.
- It reads the real shipped `settings.ariadne.json`, with no mocks.
- Reverting any entry to its bare name, for example `"weave"`, would leave `weave *` missing and trip the second loop. I confirmed this by reading the code; the edit-and-rerun was refused, so it wasn't run here.
- The live check used `couch --broadcast-list` rather than `--send-to`. Both dial the same socket, so it's adequate evidence.

**6. Architecture (ARCH-*)**
- **ARCH-DRY:** pass. One source of truth (`settings.ariadne.json`), and the test derives the ariadne list from `cmd/`.
- **ARCH-PURE:** pass. Config plus a test that reads a file; no logic tangled with IO.
- **ARCH-PURPOSE:** pass.
  - The scope covers the whole class: all workbench binaries, not just sdlc and couch.
  - I swept the docs that consume this: only the `atlas/workflow/sdlc-binary.md` fence note made a claim about how sdlc runs, and it's fixed. `docflow.md`'s `.git/config` note is about docflow's own design and still correct.
  - `git`, `gh`, `docker`, `npm` and `npx` stay bare on purpose, recorded as a separate decision.
- **ARCH-MOCK:** pass. No new external call. The live check against the real harness is a one-off Log entry, which is acceptable for a settings change.
- **ARCH-CONSTRAINTS:** not applicable; static policy only.
- **ARCH-SECURE:** pass. The trust boundary is widened explicitly, approved by the operator, and the unsandboxed settings rewrite is called out as gated.
- **ARCH-ORDER:** not applicable; nothing holds state between events. The Log confirms settings apply live without a restart.
- **ARCH-FUNERAL:** not applicable; creates nothing durable beyond the config entries.

**7. Plan revisions:** none needed. The existing Revisions entry already records that bypass-permissions mode was inferred from the docs rather than tested, and the operator-approved scope widening.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      sdlc-binary.md:963-967 now says the fence can apply for plain sdlc calls, clone-only only for forms that stay sandboxed; matches sandbox.md rule.
  - id: BR-2
    disposition: addressed
    note: |
      base_settings_test.go:58-74 requires "X *" for every cmd/ dir plus explicit pair list; passes at head; bare entry fails by inspection (revert not runnable in read-only mode).
  - id: BR-3
    disposition: addressed
    note: |
      sandbox.md:23 now lists cd-other-dir, full path, $(...), pipe as live and says redirect is from the docs; Log records the pipe observation.
  - id: BR-4
    disposition: addressed
    note: |
      Issue Revisions entry records that bypass-permissions was inferred from the docs, not observed.
  - id: BR-5
    disposition: addressed
    note: |
      $comment and sandbox.md:13 now name all ariadne cmd/ and pair binaries; sdlc/couch singled out only as the pre-allowed pair.
findings:
  - id: new
    severity: Minor
    family: consumer-docs-stale-after-behavior-change
    title: |
      sandbox.md:25 and the test header describe TestBaseSettingsExcludedVerbsMatchArguments as checking only pre-allowed verbs
    detail: |
      The BR-2 fix added a second check (every cmd/ dir and pair binary must be "X *"), but the atlas line and the test's header comment (base_settings_test.go:11-17) still cover only the first. 2nd finding in this family. Rule: when a fix round changes a test's or binary's behavior, grep its name across atlas/ and code comments and update every description in the same round. Measured: 1 atlas line plus 1 header comment.
```

{"type":"system","subtype":"permission_denied","tool_name":"Bash","tool_use_id":"toolu_01WG853MK9PKAPhEiHVHMAHe","decision_reason_type":"mode","message":"Permission to use Bash has been denied because Claude Code is running in don't ask mode. IMPORTANT: You *may* attempt to accomplish this action using other tools that might naturally be used to accomplish this goal, e.g. using head instead of cat. But you *should not* attempt to work around this denial in malicious ways, e.g. do not use your ability to run tests to execute non-test actions. You should only try to work around this restriction in reasonable ways that do not attempt to bypass the intent behind this denial. If you believe this capability is essential to complete the user's request, STOP and explain to the user what you were trying to do and why you need this permission. Let the user decide how to proceed.","uuid":"67580499-a78c-42b9-906a-938da1871f1c","session_id":"f6742821-dcb1-47cb-bc12-e6738e5eac13"}

---

## Re-review — 2026-10-12T00:02:24-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 334 — sandbox: couch still blocked in worker slots after #318 (excludedCommands present, not applied) |
| repo | ariadne |
| issue file | workshop/issues/000334-sandbox-couch-still-blocked-in-worker-slots-after-318-excludedcommands-present-not-applied.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0a1a791aa0080c7d83a55979d30cff91f8f441af..280f2768da11124a555909332d1a571fddfafbb8 |
| command | sdlc close --issue 334 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T00:02:24-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-6 is fixed. `atlas/workflow/sandbox.md:25` now describes both checks the guard test makes, the test's header comment (`base_settings_test.go:18-19`) does the same, and the test passes at head. What blocks shipping is commit `280f2768`, which changes `"gh"` to `"gh *"` in `.claude/settings.ariadne.json:100`. That exempts `gh` from the sandbox when it's called with arguments, which widens the trust boundary. No prior finding asked for it, and the issue's own Log and atlas text call widening `git`/`gh`/`docker`/`npm`/`npx` "a separate decision". The fix-review rules say a scope change is not a fix, so this needs a full `sdlc close`. The same commit also leaves the atlas stale: `sandbox.md:25` still says `gh` "is still listed by bare name … so [it runs] sandboxed with arguments".

**1. Strengths**
- The BR-6 fix covers every description of the test: the atlas line and the header comment. A grep finds no third one.
- The new atlas note about `$(` inside quoted arguments records a real failure seen on #334's close. It is useful guidance for agents, not new behavior.
- `go test ./internal/settingsx/ -run TestBaseSettings` passes at head.

**2. Critical:** none.

**3. Important**
- **The scope changed after close** (`.claude/settings.ariadne.json:100`, commit 280f2768). Exempting `gh *` from the sandbox is new behavior, not a fix. It lets `gh` run unsandboxed with any arguments. The issue's Spec, Revisions and Log don't record the decision; the Log still says "`git`/`gh`/`docker`/`npm`/`npx` are left bare, which is a separate decision". Run a full `sdlc close`, or move it to its own issue, and record the operator's approval in Revisions.
- **This is the 3rd finding in family `consumer-docs-stale-after-behavior-change`.** `atlas/workflow/sandbox.md:25` still lists `gh` among the commands "still listed by bare name … so they run sandboxed with arguments". The same commit made that false in both `settings.ariadne.json` and `~/.claude/settings.json:77`. The rule covering all three findings: a commit that changes a setting or test must, in the same commit, grep for every place the docs describe it (the atlas, the `$comment`, the issue Log) and update each one. Here that means `sandbox.md:25` and the Log line about leaving `gh` bare. Doing that sweep per commit, rather than per review round, is what stops this family recurring.

**4. Minor:** none.

**5. Test coverage**
- `TestBaseSettingsExcludedVerbsMatchArguments` passes. The `gh *` change is not guarded, which is fine because nothing requires it. If it stays, decide whether the test should pin it.

**6. Architecture (interdiff only)**
- **ARCH-DRY:** pass.
- **ARCH-PURE:** pass. Config and comment changes only.
- **ARCH-PURPOSE:** flagged. `gh *` goes beyond the issue's purpose, and the atlas sweep missed the `gh` mention.

**7. Plan revisions**
- If `gh *` stays, add a `## Revisions` entry: "Scope widened after close: `gh` exempted with arguments (`gh *`), operator-approved; supersedes the Log's 'left bare' note." Then run a full close.

```findings
dispose:
  - id: BR-6
    disposition: addressed
    note: |
      sandbox.md:25 and base_settings_test.go:18-19 now describe both checks (pre-allowed verbs, and X * for every cmd/ and pair binary); test passes at head.
findings:
  - id: new
    severity: Important
    family: scope-change-after-close
    title: |
      scope changed after close — run a full sdlc close (280f2768 exempts gh with arguments, which no finding asked for)
    detail: |
      settings.ariadne.json:100 changes "gh" to "gh *", widening the sandbox trust boundary. The issue Log and the atlas call widening git/gh/docker/npm/npx a separate decision, and no Revisions entry records it.
  - id: new
    severity: Important
    family: consumer-docs-stale-after-behavior-change
    title: |
      sandbox.md:25 still says gh is listed by bare name and runs sandboxed, after 280f2768 made it "gh *"
    detail: |
      3rd finding in this family. Rule: any commit that changes a setting or test must, in that same commit, grep every place that describes it (atlas, the $comment, the issue Log) and update each one. Measured: 1 atlas sentence plus 1 issue Log line ("git/gh/docker/npm/npx are left bare").
```

---

## Re-review — 2026-10-12T00:03:54-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 334 — sandbox: couch still blocked in worker slots after #318 (excludedCommands present, not applied) |
| repo | ariadne |
| issue file | workshop/issues/000334-sandbox-couch-still-blocked-in-worker-slots-after-318-excludedcommands-present-not-applied.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8f9ea913af4e10c119ecf8950e6a1efc1895cc0f..0741bbdeb22667b2f1fc3882d664cb708e514e28 |
| command | sdlc close --issue 334 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T00:03:54-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both open findings are fixed, and nothing new blocks shipping. For BR-7, the `gh *` widening now has a dated `## Revisions` entry. It cites 280f2768, the operator's approval and the fact that `git`, `docker`, `npm` and `npx` stay exact matches, and it says it replaces the Log's earlier "left bare" note. This close gate is the full re-close that BR-7 asked for. For BR-8, commit 0741bbde fixed every description of `gh` in one pass: the sentence at `atlas/workflow/sandbox.md:25`, the `$comment` in `.claude/settings.ariadne.json:2`, and the Log's "Cause" line. A grep across `atlas/`, `cmd/`, `.claude/`, AGENTS.md and README.md finds no other text calling `gh` sandboxed or listed by bare name. `go test ./cmd/weave/internal/settingsx/` passes.

1. **Strengths**
   - The Revisions entry is precise. It names the commit, the approval, what stayed excluded and which Log note it replaces.
   - BR-8 was fixed as a class, not as the one instance named. A new lesson in `workshop/lessons.md:456-463` states the rule ("grep the name across atlas/, comments and the issue in the commit that changes the behavior"). This answers the repeat-family escalation.
   - The Log line was amended, not deleted ("were left bare at first … `gh` was later widened"), so the history still reads true.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - No test pins `gh *`. If it were reverted to `"gh"`, `TestBaseSettingsExcludedVerbsMatchArguments` would still pass, because `gh` is neither a workbench binary nor pre-allowed with arguments. It's an operator policy choice, so I'm not raising it as a finding, but a one-line assertion would lock it in.
   - The issue frontmatter still says `updated: 2026-10-11` after the 10-12 edits.

5. **Test coverage:** the guard test covers the class of bug this issue was about: bare-word entries for workbench binaries and pre-allowed verbs. `gh` falls outside it (see Minor).

6. **Architecture**
   - ARCH-DRY: pass.
   - ARCH-PURE: pass. The test is a pure read and parse of the settings file.
   - ARCH-PURPOSE: pass. The fix was applied across the whole class.
   - ARCH-MOCK: N/A. These are config and docs changes, with no new external calls.
   - ARCH-CONSTRAINTS: N/A.
   - ARCH-SECURE: pass, with a note. `gh *` running outside the sandbox gives `gh` network and credential access without the sandbox fence. That widening was recorded and approved by the operator, and the permission flow still gates each call.
   - ARCH-ORDER: N/A. Nothing here holds state between events; it's static settings.
   - ARCH-FUNERAL: N/A. Nothing durable is created besides the review sidecars, which follow the existing plans lifecycle.

7. **Plan revisions:** none needed. Revisions now matches the code.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      Revisions entry (2026-10-12) records gh * (280f2768, operator-approved) and supersedes the Log note; this gate is the re-close.
  - id: BR-8
    disposition: addressed
    note: |
      0741bbde updates sandbox.md:25, the settings $comment and the Log Cause line; repo grep finds no stale gh-bare/sandboxed text; lesson states the class rule.
```
