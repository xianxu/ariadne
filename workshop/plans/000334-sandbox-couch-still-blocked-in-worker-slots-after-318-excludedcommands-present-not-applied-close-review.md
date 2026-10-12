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
