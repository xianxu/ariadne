# Boundary Review — ariadne#315 (whole-issue close)

| field | value |
|-------|-------|
| issue | 315 — Allow api.anthropic.com in the base-layer sandbox network allowlist |
| repo | ariadne |
| issue file | workshop/issues/000315-allow-api-anthropic-com-in-the-base-layer-sandbox-network-allowlist.md |
| boundary | whole-issue close |
| milestone | — |
| window | e2ef66b10bdf9a62bd2781fd016a7caef56e05d4..6c94a173b626aa192e92654489d2f1b2748bc3e6 |
| command | sdlc close --issue 315 |
| reviewer | claude |
| timestamp | 2026-10-09T20:32:35-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This change adds one line, `api.anthropic.com`, to `sandbox.network.allowedDomains` in the base-layer fragment `.claude/settings.ariadne.json`, and it meets the Spec and both `## Done when` clauses. The fragment is still valid JSON. `construct/base.manifest:76` merges the fragment into `.claude/settings.json`, and the merged file I have (generated, untracked) does list the domain at line 73. `go test ./cmd/weave/internal/settingsx/...` passes. The second sandbox allowlist, `.openshell/policy.yaml:155-168`, already allows Anthropic hosts, so no other copy of the list needs updating. Nothing blocks SHIP.

1. **Strengths**
   - The domain goes in the base-layer fragment, not the generated `settings.json`. That means it survives `weave compile` and reaches every repo that adopts the base layer, which is the point of the issue.
   - The change is as small as it can be, and the fragment still parses.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `claude -p` may also contact other Anthropic hosts. When it signs in with OAuth or a subscription and needs a fresh token, it can call `platform.claude.com` / `console.anthropic.com`, and those aren't on the list. The openshell policy allows `*.anthropic.com` and `*.claude.com`. If the "403 blocked by allowlist" error comes back after a token expires, those hosts are the next to add.
   - No test checks that the fragment contains the domain. The Done-when clauses don't ask for one, so this is just a note.

5. **Test coverage:** the `settingsx` merge tests pass. Nothing new is covered and nothing needed to be for this change.

6. **Architecture**
   - **ARCH-DRY: pass.** The value is defined once, in the fragment, and the generated file inherits it. The openshell policy is a separate sandbox system that already covers these hosts.
   - **ARCH-PURE: pass.** The change is configuration only.
   - **ARCH-PURPOSE: pass.** Because the fragment feeds the merge, every consumer gets the domain; nothing restates it by hand.

7. **Plan revision recommendations:** none.

```findings
findings:
  - id: new
    severity: Minor
    family: sandbox-allowlist-completeness
    title: |
      claude -p's OAuth token refresh may need more Anthropic hosts than api.anthropic.com
    detail: |
      .openshell/policy.yaml:155-168 allows *.anthropic.com, platform.claude.com and *.claude.com, but .claude/settings.ariadne.json now adds only api.anthropic.com. If the reviewer signs in with OAuth or a subscription, a token refresh may be blocked; add those hosts if the 403 comes back.
```
