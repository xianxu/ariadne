# Boundary Review — ariadne#354 (whole-issue close)

| field | value |
|-------|-------|
| issue | 354 — add tl-worker * and tl-lead * to .claude/settings.ariadne.json sandbox.excludedCommands (local gates; unblocks #350) |
| repo | ariadne |
| issue file | workshop/issues/000354-sandbox-exempt-tl-worker-and-tl-lead.md |
| boundary | whole-issue close |
| milestone | — |
| window | 482f760b343d1ee934f8a9e7e24a6c0273837b4f..729a91747d61fe6c9514d7c707571aa1a6302e05 |
| command | sdlc close --issue 354 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T18:21:08-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

I re-ran the check myself and the change does what the issue asks. It adds `"tl-worker *"` and `"tl-lead *"` to `sandbox.excludedCommands` in `.claude/settings.ariadne.json` (lines 106–107), using the same `X *` form as the other cmd binaries. I made a scratch copy of the head tree with empty `cmd/tl-worker` and `cmd/tl-lead` folders. There, `TestBaseSettingsExcludedVerbsMatchArguments` **passes**. With the base settings file swapped in, the same tree **fails**, with `lacks "tl-lead *"` and `lacks "tl-worker *"`. So the test passes with the fix and fails without it. The JSON parses. On the real tree (no such folders yet) the test passes too, because extra entries do no harm. Nothing blocks shipping. The one step left is outside the code: the operator's `local-gates-verified` label on the PR.

1. **Strengths**
   - The change is two lines, written exactly like the existing entries, and placed next to the other ariadne cmd binaries (`.claude/settings.ariadne.json:106-107`).
   - The plan was right not to edit the `$comment`. It describes the binaries as a group ("ariadne's cmd/*") rather than naming each one, so there is no list to keep in sync.
   - The Log records both a passing and a failing run, and I reproduced both on my own.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none.

5. **Test coverage notes:** The existing test reads the `cmd/` folders and the settings file, and already covers this kind of change. Checking "every `cmd/` folder has an entry" is what forces this file to be updated whenever a binary is added, so no new test is needed. Both states in `## Done when` were checked: with the folders it passes, and with the old settings it fails.

6. **Architecture notes**
   - **ARCH-DRY: pass.** Nothing is duplicated. The `$comment` describes the binaries as a group instead of restating the list.
   - **ARCH-PURE: pass.** The change is configuration only, with no logic in it.
   - **ARCH-PURPOSE: pass.** The only consumers are the test and the base settings file, and the latter is symlinked or merged into downstream repos by weave. Both pick up the change from this one file. The atlas and README do not list excluded commands one by one: `atlas/workflow/sandbox.md` is cited for the rule, not for a list. The atlas mentions of binary names (for example `base-layer.md:75`) are about building and distribution, which belongs to #350, not to this issue.

7. **Plan revision recommendations:** none. The plan matches the code.

```findings
```
