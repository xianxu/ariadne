---
id: 000422
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1a1c3909605f049622573e5ab56dd3a402c1b353' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T01:34:29-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
---

# A downstream repo's committed generated vocabulary goes stale on every ariadne vocabulary change and turns its CI red

## Problem

## Spec

Found 2026-10-16 by parley.nvim:3. parley.nvim commits a copy of the generated `construct/generated/vocabulary/issue.json`. When ariadne's issue vocabulary changes, weave's apply refuses to overwrite the committed copy ("preserving authored replacement"). parley's main merge-check then goes red until someone refreshes it by hand. That has happened three times: parley #321, #326, and PR 285 after ariadne#415, which turned parley's main red from 22:32 to about 23:50 and blocked every parley landing.

Preserving a committed file that differs from weave's output is the right default for authored files. The gap is a generated file that a repo tracks on purpose. Options:
1. Repos stop committing generated vocabulary; it is built at bootstrap. Find out why parley committed it first.
2. A repo declares tracked generated paths, and apply refreshes those, showing the diff, instead of preserving them.
3. An ariadne vocabulary change ships with a fleet check that names each downstream repo whose committed copy it makes stale, so the refresh lands with the change.

## Done when

- An ariadne vocabulary change can no longer turn a downstream repo's main CI red silently: a test or check reproduces the parley case and shows the chosen fix handling it.
- parley's issue.json is handled by the chosen rule, and the reason is recorded.

## Plan

Proposed, waiting for the TL's approval (a change to weave's ownership rule).

**Why parley commits issue.json (parley#208, 3e2daf45):** parley is a Neovim plugin that users install straight from
GitHub through a plugin manager, so they never run weave or bootstrap. It reads its issue vocabulary from the installed
tree ("shipped vocabulary available without CUE"). So the commit is a distribution artifact, and **option 1 is out**:
it would break every plugin install.

**Mechanism (parley#326's class answer, confirmed in `cmd/weave/internal/plan/ownership.go` `prepareManaged`):** weave
proves it wrote a file through its ownership inventory, `construct/generated/weave/`, which is gitignored. A cold CI
checkout has no inventory. An occupied staged output whose bytes differ from the new output, with no recorded identity,
is refused as "something weave did not write". So any upstream vocabulary change turns parley's main CI red until
someone regenerates by hand.

**Overlap:** ariadne#368 (filed from parley#326, 10-14; claimed by parley.nvim:2, never started) is this same class. Its
spec asks to keep authored-replacement protection intact (ARCH-SECURE) and to validate committed exports against
recorded provenance. Proposal: #422 implements it and #368 closes as superseded (the TL's call).

**Recommended: option 2 with committed provenance, not a blanket overwrite.** A repository declares the generated
outputs it ships in its own `construct/base.manifest`, as a new row `internal ship <path>` (parley:
`internal ship construct/generated/vocabulary/issue.json`). For each declared output, weave also writes a committed digest sidecar, `<path>.weave-sha256` (the sha256 of
what weave wrote). In `prepareManaged`, an occupied declared output whose bytes match its committed sidecar counts as
weave's own, even with no inventory, so a cold checkout regenerates it. Apply prints `refreshed tracked output <path>:
commit it` so the refreshed copy reaches the shipped plugin.
- A hand edit to issue.json changes its bytes without the sidecar, so it is still refused: protection stays intact.
- An undeclared output behaves exactly as today.
- Option 3 (a fleet check that names stale consumers) would only report the breakage, not prevent it, so it doesn't
  meet Done-when.

**parley's side** (cross-repo, routed through the TL): declare the path, un-ignore and commit the sidecar (one
`.gitignore` line beside the existing `!/construct/generated/vocabulary/issue.json`), and record the reason.

**Durable state (ARCH-CONSTRAINTS at-plan):** one sidecar per declared output: weave writes it, the consumer's git
tracks it, and it changes only when the output changes. It is removed when the declaration goes (weave's retirement).

- [ ] `ship` manifest verb; sidecar write; ownership by committed digest in `prepareManaged`.
- [ ] Tests (cmd/weave plan): the parley case — a cold checkout (no inventory), a declared output plus a matching
      sidecar, and a changed upstream output → regenerated with the report line; a hand-edited declared output → still
      refused; an undeclared output → refused as today; a warm checkout → unchanged.
- [ ] Atlas: weave ownership and tracked outputs.
- [ ] parley follow-up proposal to the TL (declaration and sidecar commit).

## Log

### 2026-10-16
