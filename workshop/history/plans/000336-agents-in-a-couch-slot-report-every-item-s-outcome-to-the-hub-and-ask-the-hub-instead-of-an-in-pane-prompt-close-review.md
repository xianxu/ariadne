# Boundary Review — ariadne#336 (whole-issue close)

| field | value |
|-------|-------|
| issue | 336 — agents in a Couch slot report every item's outcome to the hub, and ask the hub instead of an in-pane prompt |
| repo | ariadne |
| issue file | workshop/issues/000336-agents-in-a-couch-slot-report-every-item-s-outcome-to-the-hub-and-ask-the-hub-instead-of-an-in-pane-prompt.md |
| boundary | whole-issue close |
| milestone | — |
| window | 505a771abe91e2393c9033ceefc5809c9041ca59..59b3c4cb8de2d4dbcd5508e93a154719fc74ad61 |
| command | sdlc close --issue 336 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T08:50:26-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

The review is done: the edit that would have reverted `gh *` to confirm the guard test goes red was denied in this mode. Reading the test, it plainly fails without the entry. Verdict below.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This change is mostly prose, and it does what the issue asks. The new §15 Couch Slots rule in `AGENTS.base.md` covers all four Spec bullets: the hub's address through `couch --send-to hub` with a fallback to the slot that dispatched the work, reporting every outcome, asking the hub instead of an in-pane prompt, and the measured form a call must take to stay outside the sandbox. The Couch skill points at §15, the atlas is updated, and the guard test now pins `gh *`. That test passes at head. I couldn't revert the entry to watch it go red because the edit was denied in this mode, but the check is a direct `excluded["gh *"]` lookup, so removing the entry fails it.

What keeps this from a clean SHIP is that older text in `atlas/workflow/sandbox.md` now contradicts what this issue measured. The same page gives two incompatible versions of the rule. The Couch skill's shortened copy of the rule also leaves cases out. Both are cheap prose fixes.

**1. Strengths**
- §15 (`AGENTS.base.md:96-102`) states the measured rule precisely. It names exactly which substitution openers break the sandbox exemption, says `2>&1` is the one trailing form that passes, and lists the punctuation that is safe. That's backed by the Log's probe-per-character-class method.
- The hub's address is resolved live by `couch --send-to hub` (pair#463), not by a `COUCH_HUB` environment variable that would go stale. The Log records the reason, and the rule falls back to the dispatching slot.
- The guard test pin (`base_settings_test.go:72-74`) is minimal and reuses the existing loop over binaries, so no logic is duplicated.
- The deferred live check is honestly recorded under `## Revisions`, with the TL as owner and the classifier refusal noted rather than worked around.

**2. Critical findings:** none.

**3. Important findings**
- **`atlas/workflow/sandbox.md:16-22` contradicts #336's measurements and line 24 of the same page.**
  - The bullet "chained or piped with a command the list doesn't cover" implies a pipe to a covered command is fine. The measurements say any pipe, `;` chain or file redirect breaks the exemption.
  - "inside `$(...)` or with a redirect to a file" doesn't mention that `2>&1` passes.
  - "The redirect form comes from the docs and wasn't tested" is now false: #336 measured it.
  - "Verified live on 2026-10-11" doesn't mention the new substitution openers (backticks, `$((`, `<(`, `>(`).
  - **Fix:** rewrite the "Calling an excluded binary" subsection to match §15, or reduce it to a pointer to §15 plus a list of what was measured. (ARCH-DRY)

**4. Minor findings**
- **`construct/local/couch/SKILL.md:19-22` repeats §15 incompletely.** It says "nothing piped or chained after it" but leaves out "nothing before it" (`cd x &&`, a full path). It also omits `$((` and `>(`, the `2>&1` exception, and `sdlc`. Either reduce it to a pure pointer to §15 or make it complete. (ARCH-DRY)
- **The issue still has one unchecked Plan item** (part 2 of the live check). It's covered by the Revision, but the TL needs to actually follow up after the merge and a weave refresh.

**5. Test coverage notes**
- The `gh *` pin is real logic against the real settings file. The rest of the diff is prose; the behaviour it documents is backed by the live probes recorded in the Log.

**6. Architectural notes**
- **ARCH-DRY: flag.** The sandbox rule is now restated in three places (§15, `sandbox.md`, `SKILL.md`), and two of them have drifted from §15. Keep §15 as the only source and make the others point to it.
- **ARCH-PURE: pass.** The only code change is a test.
- **ARCH-PURPOSE: pass.** The purpose is delivered: the rule is in the constitution that weave compiles into every repo, and the hub's address resolves live. The deferred check is a separable, TL-owned verification, not the point of the issue.

**7. Plan revision recommendations:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Important
    family: single-source-restatement-drift
    title: |
      atlas/workflow/sandbox.md agent-rule subsection contradicts the #336 measurements and §15
    detail: |
      Lines 16-22 still say a pipe is only a problem when piped to a command the list doesn't cover, and that the redirect form was untested. They also omit the 2>&1 exception and the backtick, $((, <( and >( substitution openers. Line 24 of the same page and AGENTS.base.md §15 state the measured rule: any pipe, ; chain or file redirect breaks the exemption, and only 2>&1 passes. Rewrite the subsection to match §15, or reduce it to a pointer.
  - id: new
    severity: Minor
    family: single-source-restatement-drift
    title: |
      construct/local/couch/SKILL.md partially restates §15 and omits cases
    detail: |
      The summary leaves out 'nothing before it' (cd x &&, a full path), the $(( and >( openers, the 2>&1 exception, and sdlc. Make it a pure pointer to §15 or make it complete. Instances of this family in the window: sandbox.md lines 16-22 and SKILL.md lines 19-22.
```

{"type":"system","subtype":"permission_denied","tool_name":"Bash","tool_use_id":"toolu_01TGa3Zmc7vx657XTLYz6PLr","decision_reason_type":"mode","message":"Permission to use Bash has been denied because Claude Code is running in don't ask mode. IMPORTANT: You *may* attempt to accomplish this action using other tools that might naturally be used to accomplish this goal, e.g. using head instead of cat. But you *should not* attempt to work around this denial in malicious ways, e.g. do not use your ability to run tests to execute non-test actions. You should only try to work around this restriction in reasonable ways that do not attempt to bypass the intent behind this denial. If you believe this capability is essential to complete the user's request, STOP and explain to the user what you were trying to do and why you need this permission. Let the user decide how to proceed.","uuid":"6cfd9db7-6638-49c6-a1da-e8a81619b911","session_id":"a54880bc-12a0-42b7-b5a7-70067226f988"}

---

## Re-review — 2026-10-12T08:51:58-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 336 — agents in a Couch slot report every item's outcome to the hub, and ask the hub instead of an in-pane prompt |
| repo | ariadne |
| issue file | workshop/issues/000336-agents-in-a-couch-slot-report-every-item-s-outcome-to-the-hub-and-ask-the-hub-instead-of-an-in-pane-prompt.md |
| boundary | whole-issue close |
| milestone | — |
| window | 505a771abe91e2393c9033ceefc5809c9041ca59..b8c75fb0e4dc45ec7e4c93204c7c7d38b8c3eb15 |
| command | sdlc close --issue 336 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T08:51:58-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both findings from the last round are fixed and nothing new turned up, so the verdict is SHIP. The constitution's new §15 (Couch Slots, in `AGENTS.base.md`) is now the only place that states the agent rule. `atlas/workflow/sandbox.md` and `construct/local/couch/SKILL.md` now just point to it, and `sandbox.md` keeps a dated log of what was measured. That log matches §15 and the issue's Log. `TestBaseSettingsExcludedVerbsMatchArguments` now pins `gh *`, and the test passes at the pinned head. The fourth Done-when clause is still open: an operator-started slot reporting on its own. The issue's Revisions entry explicitly hands that to the TL after the merge, so it is a recorded split, not a hidden gap.

1. **Strengths**
   - `AGENTS.base.md:96-102`: §15 states the measured rule completely. Nothing may come before or after the command, `2>&1` is the only exception, and the arguments may not contain any of the five substitution openers. It also lists the punctuation that is safe to use.
   - `construct/local/couch/SKILL.md:19-21`: this is now only a pointer, which removes the second copy of the rule that could drift.
   - `atlas/workflow/sandbox.md:17-24`: the measurements are kept as dated evidence instead of being restated as a rule. The old "redirect form wasn't tested" caveat is gone because the redirect case has now been measured.
   - `base_settings_test.go:72-74`: pinning `gh` reuses the existing `excluded[verb+" *"]` check, so there is no second test path.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none worth raising. The atlas log's list of passing punctuation (which also names `${VAR}` and backslashes) is a record of what was measured, not a restatement of the rule, so it is not drift.

5. **Test coverage**
   - The settingsx test passes at the head commit.
   - I could not run the revert-and-run-red check: the sandbox refused the command that edits the settings file. Reading the code instead, the test asserts that `excluded["gh *"]` exists, and `.claude/settings.ariadne.json:100` holds `"gh *"`. Changing that entry back to bare `gh` would therefore fail the test, which matches the Plan's claim.

6. **Architecture**
   - **ARCH-DRY: pass.** The rule exists once (§15), and both other places that use it point to it.
   - **ARCH-PURE: pass.** The change is documentation and configuration plus one data-only test addition.
   - **ARCH-PURPOSE: pass.** I checked every place that consumes the rule: the constitution, the couch skill, `sandbox.md` and `index.md`. All of them now defer to §15. The deferred live check is the separable operator-started test, not the point of the issue.

7. **Plan revisions:** none. The plan matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      sandbox.md:17-24 now points at §15; the 10-12 measurement log matches §15 (any pipe, ; or file redirect breaks the exemption, 2>&1 passes, all five openers); the untested-redirect caveat is gone.
  - id: BR-2
    disposition: addressed
    note: |
      SKILL.md:19-21 is now a pure pointer to §15 ("it is not restated here"), with no partial summary left.
```

{"type":"system","subtype":"permission_denied","tool_name":"Bash","tool_use_id":"toolu_01W9WCEsnAV1FxEVbaNiT1eA","decision_reason_type":"mode","message":"Permission to use Bash has been denied because Claude Code is running in don't ask mode. IMPORTANT: You *may* attempt to accomplish this action using other tools that might naturally be used to accomplish this goal, e.g. using head instead of cat. But you *should not* attempt to work around this denial in malicious ways, e.g. do not use your ability to run tests to execute non-test actions. You should only try to work around this restriction in reasonable ways that do not attempt to bypass the intent behind this denial. If you believe this capability is essential to complete the user's request, STOP and explain to the user what you were trying to do and why you need this permission. Let the user decide how to proceed.","uuid":"c0db3f18-fd4d-469c-97e2-00e8425296b2","session_id":"e4df5ede-626a-417a-9e62-b96d1f17e526"}
