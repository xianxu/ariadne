# Boundary Review — ariadne#242 (whole-issue close)

| field | value |
|-------|-------|
| issue | 242 — Slots v2: workspace identity |
| repo | ariadne |
| issue file | workshop/issues/000242-slots-v2-workspace-identity.md |
| boundary | whole-issue close |
| milestone | — |
| window | 105d5ba1ed89432de9f3506a50433db0753fd8f4..b49163b742ac1c5a2605d695abe04d5f78e26c58 |
| command | sdlc close --issue 242 |
| reviewer | codex |
| timestamp | 2026-09-22T22:56:14-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The shared resolver, consumer migration, stateful fake, conformance tests, and atlas documentation are substantial and mostly align with the plan. Boundary crossing is blocked by fail-open handling of malformed Git HEAD evidence; README documentation is also missing for the new user-facing `sdlc workspace` command.

```findings
findings:
  - id: new
    severity: Critical
    family: malformed-git-evidence-fails-open
    title: |
      Malformed all-zero HEAD values are accepted as a valid unborn workspace
    detail: |
      `Classify` treats any all-zero string as unborn (`pkg/workspace/identity.go:63`) without requiring the Git OID grammar. A malformed value such as `HEAD 0` therefore resolves a primary successfully with `head: null`, contradicting the plan's fail-closed malformed-observation contract. Validate HEAD as exactly a 40/64-character zero OID for unborn state, or reject it; add a regression test that fails before the fix.
  - id: new
    severity: Important
    family: user-facing-surface-docs
    title: |
      README update is missing for the new workspace CLI surface
    detail: |
      `cmd/sdlc/main.go:112` adds `sdlc workspace [address] --json`, but README.md is unchanged and has no usage or contract entry. Add the command and its basic invocation/JSON purpose to README.md.
```

### Strengths

- `pkg/workspace` cleanly separates pure parsing/classification from Git IO.
- Real-Git and stateful-fake conformance coverage is extensive.
- Consumer migration preserves current checkout paths while deriving repository identity from Git.
- Atlas documentation covers the new workspace contract and links it from `atlas/index.md`.
- Verification commands passed: workspace tests, targeted CLI tests, `go vet`, and `git diff --check`.

### Critical findings

See machine-readable finding above: malformed all-zero HEAD evidence is accepted at `pkg/workspace/identity.go:63`.

### Important findings

See machine-readable finding above: README coverage is missing for `sdlc workspace`.

### Minor findings

None.

### Test coverage notes

The executed workspace and targeted CLI suites passed. The issue records a broader suite with the pre-existing #210 missing-plan test excluded. No regression test currently exercises malformed all-zero HEAD input.

### Architectural notes

- ARCH-DRY: Pass — shared topology and identity logic replaces duplicated parsers.
- ARCH-PURE: Pass — classification and address logic are pure.
- ARCH-PURPOSE: Pass — the consumer inventory appears substantially delivered.
- ARCH-MOCK: Pass — stateful fake and real-Git conformance are present.
- ARCH-CONSTRAINTS: Pass — linear topology reads and scale samples are documented.
- ARCH-SECURE: Flag — malformed external Git evidence is not fully rejected.
- ARCH-ORDER: Pass — resolution is treated as an observational snapshot with rechecks.
- ARCH-FUNERAL: Pass — the resolver creates no durable artifacts.

### Plan revision recommendations

- Add a `## Revisions` entry requiring strict HEAD OID validation and a regression test for malformed all-zero values.
- Add README.md to Task 5’s documentation file list and completion criteria.
