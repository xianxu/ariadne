# Boundary Review — ariadne#318 (whole-issue close)

| field | value |
|-------|-------|
| issue | 318 — sandbox: allow couch in base-layer settings so agents can message peers |
| repo | ariadne |
| issue file | workshop/issues/000318-sandbox-allow-couch-in-base-layer-settings-so-agents-can-message-peers.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1c1144a43928df983f8af4c5d69bd0ce6d69fd19..3ac15bafe3d8c4094cd9660a223d992a85e1980d |
| command | sdlc close --issue 318 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T15:22:14-07:00 |
| verdict | SHIP |

## Review

I've read the full diff and the issue; writing up the verdict now.

```verdict
verdict: SHIP
confidence: high
```

This is a small config change, and it does what the Spec and its Revisions describe. `couch` is added to `sandbox.excludedCommands`, next to `sdlc`, `git` and `gh`. `Bash(sdlc:*)` and `Bash(couch:*)` are added to `permissions.allow`. Both keys are listed in `$merge_keys`, so each downstream repo's `settings.local.json` merges its own entries with these instead of replacing them. Codex's lack of a counterpart is explained in `.codex/config.toml`, and that explanation is accurate: `network_access = true` is present at `.codex/config.toml:25` and `approval_policy = "on-request"`. Gemini being left alone is correct: `construct/base.manifest` has no `.gemini` row. The atlas was updated in the same range (`atlas/workflow/sandbox.md:13`). Nothing blocks shipping. Two Plan rows are still open, the live `couch --send-to` / `sdlc merge` check and the `weave refresh` in ops, but the plan itself defines both as post-merge checks.

Done-when: current. The three clauses (`sdlc merge` runs in auto mode without a denial, a sandboxed couch send delivers with a receipt, the change reaches peer repos through `weave refresh`) cover everything the diff builds. Revision items (4) and (5) were both settled with no code change, so they add nothing to check.

1. **Strengths**
   - `.claude/settings.ariadne.json:21-22,102`: the change sits in the single base-layer source and relies on the existing `$merge_keys` union, so no restatement is needed per repo.
   - `.codex/config.toml:13-16`: the "no counterpart needed" decision is recorded next to the setting it depends on, and the reasoning checks out against the file.
   - Revision (5) deliberately keeps the `.claude/settings.json` rewrite as an operator step. This is a sound choice: it doesn't let an agent rewrite its own permissions.
   - The new allow entries use the same `Bash(x:*)` form as the existing entries in the same list.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The atlas states two things as fact that nothing in the window establishes:
     - `atlas/workflow/sandbox.md:13` says "The exclusion matches only when the verb is the leading command". Revision (4) itself says kink 52 "most likely" came from a compound command, so this is a hypothesis.
     - The same line says sdlc's spawned git "can write `.git/config`" because of the exclusion. That depends on the exclusion applying to child processes, which this issue's own `sdlc merge` was meant to confirm and which hasn't happened yet.
     - Either soften both claims or confirm them in the post-merge live check.
   - The Plan's Gemini row says "record why in `## Revisions`", but the reason is only in the Plan row. Revisions just says "(see Plan)". Plan-only wording, no effect on behavior.
   - The base-layer `$comment` and the atlas paragraph overlap in what they say. That's acceptable: the comment is a short pointer and the atlas has the full explanation.

5. **Test coverage notes:** This change is pure configuration (JSON and TOML that control sandbox and permission behavior). No unit test asserts which commands are allowed in the base-layer policy. The issue's real verification is the live Done-when checks, and two of them are still open in the Plan, so close should wait for them. The Log records `go test ./cmd/weave/...` exiting 0 after the merge with origin/main. That covers the merge mechanism, not these entries.

6. **Architecture notes**
   - **ARCH-DRY: pass.** The policy is single-sourced in `settings.ariadne.json`, and the generated `.claude/settings.json` is derived from it by weave. Nothing is copied by hand into other repos.
   - **ARCH-PURE: pass.** No code changed; only declarative configuration.
   - **ARCH-PURPOSE: pass.** The diff covers the whole purpose: the couch exclusion, the pre-allows that keep the classifier from re-judging these commands, and a check of every other agent tool. For Codex and Gemini, the reason no change is needed is written down. Rolling out to ops depends on the operator running `weave refresh`. That's the delivery mechanism the plan defines, not a deferral of the point of the issue.

7. **Plan revision recommendations:** Optionally add a one-line Revisions entry saying why Gemini is excluded (not woven, ariadne-local only), so the Plan row's pointer to Revisions is accurate.

```findings
findings:
  - id: new
    severity: Minor
    family: doc-claim-unverified
    title: |
      atlas sandbox.md states unverified exclusion semantics as fact
    detail: |
      atlas/workflow/sandbox.md:13 says the exclusion "matches only when the verb is the leading command" and that sdlc's git "can write .git/config"; Revision (4) calls the former "most likely" and the latter awaits this issue's own sdlc merge. Instances in window: these two clauses only. Soften the wording or confirm in the post-merge live check.
  - id: new
    severity: Minor
    family: plan-pointer-mismatch
    title: |
      Gemini row says "record why in Revisions" but the reason only lives in the Plan
    detail: |
      The Revisions entry says "(see Plan)" and the Plan row says "record why in Revisions"; each points at the other. Add a one-line Revisions note or reword the Plan row.
```
