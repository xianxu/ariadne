---
gate: boundary-review
issue: 315
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-09T20:32:35-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: claude -p's OAuth token refresh may need more Anthropic hosts than api.anthropic.com
          detail: .openshell/policy.yaml:155-168 allows *.anthropic.com, platform.claude.com and *.claude.com, but .claude/settings.ariadne.json now adds only api.anthropic.com. If the reviewer signs in with OAuth or a subscription, a token refresh may be blocked; add those hosts if the 403 comes back.
          family: sandbox-allowlist-completeness
          round: 1
      recipe: small-diff-review
      blocked: false
---

# Gate ledger — ariadne#315 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-09T20:32:35-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `sandbox-allowlist-completeness` claude -p's OAuth token refresh may need more Anthropic hosts than api.anthropic.com
  .openshell/policy.yaml:155-168 allows *.anthropic.com, platform.claude.com and *.claude.com, but .claude/settings.ariadne.json now adds only api.anthropic.com. If the reviewer signs in with OAuth or a subscription, a token refresh may be blocked; add those hosts if the 403 comes back.

## Open findings

- **BR-1** [Minor] `sandbox-allowlist-completeness` claude -p's OAuth token refresh may need more Anthropic hosts than api.anthropic.com
