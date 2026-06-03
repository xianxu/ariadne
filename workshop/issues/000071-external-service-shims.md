---
id: 000071
status: open
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
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
