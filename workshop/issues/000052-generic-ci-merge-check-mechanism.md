---
id: 000052
status: working
deps: []
created: 2026-05-31
updated: 2026-05-31
estimate_hours: 3
---

# Generic CI merge-check mechanism (pluggable publish gate for derivatives)

## Problem

you-decide needs a server-side gate: a PR to `main` must not merge unless its
shared-substrate files are `review: passed` (you-decide#4 M3). But the *mechanism*
— "run a repo-defined gate over the PR's range in CI" — is generic: every ariadne
derivative wants a place to plug its own merge-time checks (a license check, `go
test`, a lint, a substrate gate, or nothing). Build the generic mechanism in the
base layer first; you-decide then plugs in its `review-gate.sh`.

## Spec

A three-layer split, forced by one constraint: **GitHub Actions does not follow
symlinked workflow files** (`.github/workflows/*.yml` must be real files) — but it
*does* follow symlinked scripts a workflow runs.

```
.github/workflows/merge-check.yml   ← SEED   thin, stable shim; real file per repo
scripts/run-merge-checks.sh         ← SYMLINK generic runner; propagates from ariadne
scripts/merge-checks.d/*            ← SCAFFOLD empty dir; each derivative drops its checks
```

- **Seeded shim** — PR-triggered (`pull_request` → `main`), `fetch-depth: 0`, stable
  **job name `merge-check`** (so a required-status-check can reference it generically).
  Computes the PR range and calls the runner. Seeded (not symlinked) because Actions
  ignores symlinked workflow files; kept thin so it rarely needs to change.
- **Symlinked runner** `scripts/run-merge-checks.sh <base> <head>` — discovers every
  executable in `scripts/merge-checks.d/`, runs each with `(base, head)`, aggregates:
  exit 0 iff all pass **or no checks exist** (no-op for repos with no gate). Findings
  → stderr. Propagates via the normal refresh flow, so the mechanism improves everywhere.
- **Scaffolded plugin dir** `scripts/merge-checks.d/` — repo-owned, empty by default.
  Each check: `<check> BASE_SHA HEAD_SHA` → exit 0 pass / non-zero fail. Run in
  filename order (`10-`, `20-` prefixes order them).

**One runner, two call sites (DRY):** the same `run-merge-checks.sh` is invoked by CI
*and* by a repo's local `pre-push` hook — local and server gates can't drift.

**Range correctness:** the shim passes `merge-base(base,head)` as `<base>` so a two-dot
`base..head` diff == exactly the PR's changes (no need to change individual checks to
three-dot).

**Class boundary:** these are *deterministic, server-side, repo-specific* checks —
distinct from the LLM judges (`sdlc judge`: plan/specs/lessons), which stay local at
pre-merge (an LLM in CI is flaky/costly). Two complementary tiers.

## Out of scope / later

- **M2 — `make remote-init` operator verb**: `gh repo create` (if absent) + push +
  *opt-in* `gh api … branches/main/protection` to make `merge-check` a required check.
  Needs `gh` auth (operator-run; `make` = human entrypoint). The mechanism works
  advisory-only without it (check runs + reports; doesn't block). Direct push to main
  stays open (the acknowledged escape per #51).

## Review findings (Codex A1, 2026-05-31) — enforcement hardening for M2

A cross-stack review (ariadne #53 Phase A1) confirmed the M1 mechanism is sound for
**advisory** use, and surfaced what M2 must add to make it a real *boundary* (not just a
report). These are the difference between "the gate runs" and "the gate can't be evaded":

- **Trusted gate code (the key one).** CI currently runs `run-merge-checks.sh` + the
  repo's `merge-checks.d/*` **from the PR's own tree** — so a PR can neuter the gate it's
  subject to (rewrite the check, `chmod -x` the plugin) and still go green. M2 must run the
  gate from a **trusted, pinned source** — the protected base branch's copy, or a pinned
  ariadne ref — against the PR's *content*.
- **Required-checks manifest.** "Empty `merge-checks.d/` = pass" is right for a repo with no
  gate, but for a *gated* repo, dropping/disabling the plugin silently no-ops. Add an
  opt-in `scripts/merge-checks.required` (names that MUST be present + executable, else fail).
- **`--no-verify` is unenforceable locally** — the real teeth are branch protection +
  required `merge-check` status. (Already the M2 `make remote-init` job.)
- (minor) `review-gate`-style checks that use `--diff-filter=d` get a rename false-positive
  (a file renamed *out* of a gated dir reports the old path missing-review). Fails closed
  (safe); fix with `--name-status` if it bites.

## Plan

**M1 — the mechanism (this milestone):**
- [ ] `scripts/run-merge-checks.sh` — generic runner (discover `merge-checks.d/*`, run over range, aggregate). Bash 3.2-safe.
- [ ] `scripts/merge-checks.d/` — scaffolded dir; ariadne ships a `README.md` documenting the check contract.
- [ ] `.github/workflows/merge-check.yml` — seeded shim (PR→main, fetch-depth 0, merge-base range, job `merge-check`).
- [ ] `construct/base.manifest` — `seed .github/workflows/merge-check.yml`, `symlink scripts/run-merge-checks.sh`, `scaffold scripts/merge-checks.d`.
- [ ] `atlas/workflow/` — document the mechanism (new doc or extend `pre-merge-checks.md`).
- [ ] Test: with a dummy passing + failing check in `merge-checks.d/`, the runner aggregates correctly; empty dir → pass.

**M2 — remote bootstrap (follow-up):**
- [ ] `make remote-init` (repo-create + opt-in required-check wiring).

## Notes / cross-refs

- Downstream consumer: **you-decide#4 M3** — drops `merge-checks.d/10-review-gate.sh`, refactors its `pre-push` hook to call `run-merge-checks.sh`, runs the check advisory on PRs.
- Pairs with **#51** (in-place-branch workflow): #51 makes publishing go through PRs, which is *what gives this CI check something to run on*. Local-merge-push is the acknowledged escape that bypasses it.
- Sibling tier: `scripts/parallel-checks.sh` / `sdlc judge` (LLM judges, local).
