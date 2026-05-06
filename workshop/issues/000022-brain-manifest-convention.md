---
id: 000022
status: done
deps: []
created: 2026-05-05
updated: 2026-05-05
estimate_hours: 1
actual_hours: 0.4
---

# brain manifest convention in the constitution

## Done when

- `AGENTS.md` §1 (Peer Repo, "brain is a special peer..." paragraph) is extended with the **brain manifest convention**: a repo is a brain iff it contains a `.brain/config.md` manifest at its root, and the manifest declares mode + identity + operational knobs.
- The convention is described tightly enough that any agent reading the constitution can answer "is this a brain?" and "what mode is it in?" without consulting other docs.
- A pointer to the full schema + security posture lives in the constitution (referencing `brain/atlas/threat-model-shared-brain.md`), so agents who need depth can find it.
- After landing in ariadne, the change propagates to downstream peers (nous, brain, charon, parley) via `make refresh` so all ariadne-styled repos see the same convention.

## Spec

The shared-brain effort (`nous#3`, `nous#6`, the threat model in `brain/atlas/threat-model-shared-brain.md`) introduces the **`.brain/config.md` manifest** as the canonical way to identify a repo as a brain and to declare its security mode. Identification by directory or repo name is brittle (real names will vary: `brain`, `family-brain`, `brain-private`, `xianxu-brain`, etc.) and conflating naming with security mode is an anti-pattern.

The constitution is the right place for this convention because:

1. Multiple downstream tools need to agree on it: the passphrase wrapper in `nous#3`, the cross-brain resolver in `nous#6`, the sync daemon in `nous#4`, the lifecycle integration in `charon#21`. Encoding it in each repo independently is parallel-mechanism failure.
2. Any agent reading any ariadne-styled repo's AGENTS.md should immediately understand "is this a brain, and if so what mode?" — that's a constitutional-level question, not a per-repo one.
3. Brain is already named in §1 as a special peer; this issue is the natural extension of that paragraph rather than a new concept.

What lands in AGENTS.md (proposed; iterate during M1):

- **Identification rule.** A repo is a brain iff `.brain/config.md` exists at its root. `test -d .brain` is the predicate. Naming carries no security weight.
- **Manifest schema** (declared, not exhaustively documented in the constitution):
  - `mode: private | shared`
  - `name: <slug>` — brain identity for cross-brain references, decoupled from directory and remote name
  - `recipients: [<gpg-fingerprint>, ...]` — for `shared` mode
  - `passphrase_source: keychain | op | tty | env` — for `private` mode
  - `sync_substrate: syncthing | git-daemon | none` — for `shared` mode
- **Behavior implication.** A repo without `.brain/config.md` is not a brain and does not receive brain-aware behavior (no encryption, no sync daemon attachment, no cross-brain reference resolution against it).
- **Pointer.** Full schema, security rationale, and threat-model context live in `brain/atlas/threat-model-shared-brain.md`. The constitution names the convention; the threat model carries the depth.

Out of scope for this issue:

- Implementing the manifest reader / writer (lives in `nous#3` M1).
- Implementing the cross-brain resolver (lives in `nous#6` M2).
- The threat model itself (already shipped in `nous#8` M1).
- Atlas entries beyond the constitutional change (downstream repos may add atlas pointers as needed, not gating).

## Plan

### M1 — extend AGENTS.md §1 with the brain manifest convention

- [ ] Edit `AGENTS.md` in ariadne, in the "brain is a special peer" paragraph of §1 (Peer Repo). Add the identification rule, the manifest schema, the behavior implication, and the threat-model pointer.
- [ ] Keep the addition tight — constitution prose, not the threat model. ~1 paragraph net add.
- [ ] Commit + push ariadne.

### M2 — propagate to downstream peers

- [ ] Run `make refresh` in nous, brain (and charon, parley as time permits) to pull the updated AGENTS.md.
- [ ] Verify each downstream's AGENTS.md now carries the convention.
- [ ] Commit + push the propagated files in each downstream.

## Log



- 2026-05-05: closed M1 — AGENTS.md §1 extended (commit 472139c); atlas/brain-manifest.md pointer added with index entry
- 2026-05-05: closed M2 — make refresh ran in nous and charon; AGENTS.md updated in both; brain follows via symlink
### 2026-05-05

- Issue spec'd. Carved out from `nous#3` after recognizing that the manifest convention is a constitutional-level change affecting multiple downstream tools, not a `nous#3`-internal convention.
