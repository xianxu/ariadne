---
id: 000071
status: open
deps: [nous#42, nous#44, nous#45]
github_issue:
created: 2026-06-02
updated: 2026-06-06
estimate_hours:
---

# construct a testable shim for every external service (gh/github first): the deterministic-shell mock pattern

## Problem

AGENTS.md §5 mandates process-level fakes ("external-service features ship a process-level
fake — function-call mocks miss interaction bugs"), but in practice we don't have them, and
it bites:

- `nous#41 #11` (re-invite hard-error fix) shipped with **zero automated coverage** —
  `gh` shells out to the CLI with no injectable seam, so it's only "dogfood-verified."
- `nous#26` (GitHub-mediated onboarding) shipped "build+vet pass" per milestone, then a
  manual run against real GitHub caught **five** control-plane bugs in succession (404 on
  the *validation* lookup not the add; `MinimalRepository` omitting `ssh_url`; stuck
  collaborator-but-unpublished state; discovery filter excluding provisioned-for-sharing
  brains; missing operator-pubkey publish). Every one was on the GitHub **control plane**,
  which our `file://` bare-repo integration tests model the **data plane** of but not.

Function-call mocks ("every `gh.AddCollaborator` returns nil") cover the trivial cases and
miss exactly these interaction bugs — the ones that only emerge from real-shaped responses
and multi-call state.

This is already an operator vision, not a new idea:
`brain/docs/vision/2026-05-19-01-pensive-auto-mocking-external-services.md` (and the earlier
`2026-05-04-01-pensive-auto-mock.md`), built on the deterministic-shell pensive
(`brain/docs/vision/2026-05-12-01-pensive-book-4-deterministic-shell.md`). This issue
promotes that vision to buildable work.

## Spec

### Generic framing (the pattern, load-bearing — solve gh *as an instance of this*)

For **any** external binary/service **X** we integrate against, ariadne constructs:

1. **`shim(X)`** — a code seam in front of X. *All* of our access to X goes through
   `shim(X)`; nothing calls the raw binary/API directly. (For `gh`: a `lib/gh`-style wrapper
   that is the only thing that execs `gh`.)
2. **`shim'(X)`** — a **testable fake** behind the same seam: a process-/in-memory model of
   X that mimics X's *internal state to the fidelity we need to reproduce observed
   behavior* — real-shaped responses, multi-call/multi-user state, tear-down between tests —
   and that all real-flow code paths go through **unchanged**.

Make this a **standing ariadne coding convention**: every external service we integrate
(`gh`/GitHub, `gmail`, Google OAuth, gpg/gpg-agent, …) gets a `shim(X)` + `shim'(X)` pair,
documented once (AGENTS.md and/or `construct/`) so it's the default, not a per-feature
afterthought. The fidelity bar is "X-enough that our integration test exercises every call
we actually make," not "reimplement X."

### First instance: gh / GitHub

- `shim(gh)`: consolidate all `gh` access behind one seam (extend the existing `lib/gh`
  wrapper so nothing else execs `gh`).
- `shim'(gh)`: a local GitHub-control-plane fake the `gh` CLI (or our seam) talks to, backed
  by ordinary git/file ops on tmpdir bare repos:
  - `PUT collaborators/<login>` mutates internal state; the *validation lookup* path is
    modeled (the nous#26 bug #1 surface);
  - `user/repository_invitations` returns the **`MinimalRepository`** shape (no
    `ssh_url`/`clone_url`) — reproduces bug #2;
  - `PATCH user/repository_invitations/<id>` transitions to accepted;
  - multi-user token contexts (switch operator ↔ joiner);
  - repo content underneath is real git on bare repos in tmpdir.
- Then: rewrite the nous#26 / nous#41 GitHub-layer integration tests to run hermetically
  through `shim'(gh)` — all five nous#26 bugs + the `nous#41 #11` re-invite hard-error
  should be catchable in-process.

## Done when

- A documented **shim(X)/shim'(X) convention** exists (AGENTS.md / `construct/`), naming it
  the default for every external integration.
- `gh`/GitHub is the first built instance: `shim(gh)` seam + `shim'(gh)` fake; the
  nous#26/#41 GitHub-control-plane flows run as **hermetic** integration tests (no
  network, no real GitHub), and reproduce the historical bugs as regression coverage.

## Plan

- [ ] Write the generic shim convention (AGENTS.md §5 extension + a `construct/` note).
- [ ] `shim(gh)`: make `lib/gh` the sole seam to the `gh` binary.
- [ ] `shim'(gh)`: control-plane fake (invitations, collaborators, MinimalRepository shape,
      multi-user tokens) over tmpdir git.
- [ ] Hermetic integration tests for the nous#26 + nous#41 GitHub flows; pin the historical
      bugs.
- [ ] Capture the pattern so the next external (gmail / Google OAuth) follows it by default.

## Log

### 2026-06-02

Filed from the sdlc tooling retro
(`workshop/pensive/2026-06-02-01-pensive-sdlc-tooling-retro.md`, finding F5). Operator: make
gh/github the first instance, but solve it as the **generic pattern** — always construct a
shim for any external service (gmail, Google OAuth, …), systematically. Anchored in the
existing auto-mocking vision in brain.

## Revisions

### 2026-06-05 — repurpose #71 as the architecture-promotion step; gh instance → nous#42

Brainstormed the scope with operator. Outcome reshapes this issue:

- **The gh instance is built in nous, not here.** `shim(gh)`/`shim'(gh)` and the hermetic
  nous#26/#41 regression tests are now **nous#42** (the spec lives there). gh is only used in
  nous; building it here would split spec-from-code across repos. `deps: [nous#42]` is the
  stable, machine-checkable cross-repo link (prose links rot) — **#71 cannot close until
  nous#42 lands.**
- **#71 is scoped down to the final step:** *promote the proven shim(X)/shim'(X) pattern to an
  ariadne architecture choice.* That means, only **after** the gh instance (nous#42) and the
  planned second instance (Google OAuth, also nous) succeed: amend AGENTS.md §5 (replace
  "process-level fake" with "stateful fake behind a provider-neutral port," distinguishing it
  from the per-call stubs §5 actually warned against) and add a pattern entry to the `ARCH-*`
  registry / architecture.md. **No ariadne files change before then** — the convention is not
  generalized from n=1.
- **Design decisions fixed for the pattern** (full rationale in nous#42): provider-neutral
  domain port (Ports & Adapters) with documented extension points for vendor peculiarities;
  in-process library-call transport with a *stateful* fake (not bridge/RPC/channel);
  wire-fidelity owned by a periodic **dual-backend contract test** (fake always + real `gh`
  build-tagged) — the "make the assumptions explicit," two-step grounding model; uniform
  `New(Conf)`/`NewFake(Conf)` constructor convention with opaque per-service `Conf` and **no**
  shared cross-service framework.
- Originally this issue's "Done when" bundled both the convention *and* the gh instance; that
  is superseded by the split above. The gh-instance bullets now belong to nous#42; this
  issue's completion = the §5 + architecture.md promotion, gated on two successful instances.

### 2026-06-06 — gate on the second instance + the full-surface demo (deps expanded)

nous#42 (gh) is now done+merged AND certified against real GitHub. But `deps:`
only listed nous#42, so once it landed #71 was no longer actually blocked — the
"promote only after ≥2 instances + the surface is proven" rule was prose, not an
enforced gate. Fixed by expanding `deps:` to the work that must precede promotion:
- **nous#44** — `shim(google-oauth)` (instance #2; proves the pattern generalizes
  past gh, incl. the async-callback case).
- **nous#45** — shim *every* nous external dependency + an integrated end-to-end
  mock harness (the deterministic-shell demo; the evidence the convention is the
  default across the whole surface, not just n=2).

`deps: [nous#42, nous#44, nous#45]` now machine-enforces "don't promote to an
architecture choice until the pattern is proven, generalized, AND demonstrated
end-to-end." Only then does #71's own work run (AGENTS.md §5 amendment + ARCH
registry / architecture.md entry).

### 2026-06-08 — n=2-real cross-provider evidence (nous#48: OAuth at Google + Microsoft)

Promotion evidence the convention demands: the OAuth shim now has **two real
backends** behind one port, not one. nous#48 added Microsoft/Entra as the second
real OAuth adapter, and the design held under genuine cross-provider stress — a
single real backend (Google) could have let provider-quirks masquerade as the
abstraction; two cannot. Concretely, the pattern's stated design decisions all
survived n=2-real:

- **Provider-neutral port + opaque per-service `Conf`, no shared cross-service
  framework:** confirmed. One generic `OIDCProvider` drives both; the per-provider
  variation is an injected in-package `dialect` value (identity extractor, required
  scopes, auth-URL params, PKCE flag, revoke mechanism) — sibling provider files
  over a shared core, **not** a second adapter and **not** a transport abstraction
  spanning services. The ariadne#71 "no shared cross-service framework" rule was
  the right constraint: the reuse is *within* the oauth shim, not across shims.
- **Documented extension points for vendor peculiarities:** exactly what absorbed
  Microsoft's real differences — PKCE/public-client (vs Google's secret), the
  `preferred_username` identity claim with **no** `email_verified` guard (the guard
  proved Google-specific and moved into `googleIdentity`), `offline_access` scope
  (vs `access_type=offline`), and **no token-revoke endpoint at all**
  (`ErrRevokeUnsupported` — the port method generalizes, the mechanism doesn't).
- **Stateful fake behind the port + dual-backend contract test:** the dialect-aware
  fake certifies the **same** `runOAuthContract` body under both Google and
  Microsoft dialects (always-on, hermetic); real-MS grounding is `conformance`-
  tagged like Google's. A wire quirk the fake-as-stateful-model surfaced that
  per-call stubs would miss: Microsoft's **single-use refresh-token rotation**
  forces the grounding harness to *persist the rotated token back* — opposite
  persistence discipline from Google under the identical contract body.

Full write-up: `nous/workshop/targets/oauth-credential-lifecycle.md` `## Revisions`
(2026-06-08, nous#48). This is the n=2-real leg of the promotion gate's evidence
base (alongside gh n=2 in nous#46); #71's own promotion work still waits on
`deps: [nous#42, nous#44, nous#45]`.
