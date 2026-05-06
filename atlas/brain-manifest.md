# Brain Manifest

A pointer for agents and humans encountering the brain manifest convention in this repo's `AGENTS.md` §1. The convention is the canonical answer to "is this repo a brain?" and "what mode is it in?" It is constitutional (lives in AGENTS.md) because multiple downstream tools depend on it converging.

## What it is

A repo is a **brain** iff it contains `.brain/config.md` at its root. The manifest declares:

- `mode: private | shared`
- `name: <slug>` — brain identity for cross-brain references, decoupled from directory and remote name
- `recipients:` — for shared mode, the GPG fingerprints admitted to the brain
- `passphrase_source:` — for private mode, where the gcrypt passphrase is fetched from (`keychain`, `op`, `tty`, `env`)
- `sync_substrate:` — for shared mode, the sync mechanism (`syncthing`, `git-daemon`, `none`)

A repo without `.brain/config.md` is **not** a brain — agents apply brain-aware behavior (encryption, sync, cross-brain reference resolution) only to repos that declare themselves.

## Why constitutional

The convention is in AGENTS.md (not just downstream-repo docs) because multiple tools across repos must converge:

- **`nous#3`** — the gcrypt passphrase wrapper reads `passphrase_source:` from the manifest.
- **`nous#4`** — the sync daemon attaches only to repos with `sync_substrate:` declared.
- **`nous#6`** — the cross-brain reference resolver discovers candidate brains by scanning for `.brain/config.md` manifests, mapping `manifest.name` → checkout path. Resolution is by manifest name, not by directory name.
- **`charon#21`** — the gpg-agent lifecycle integration only pre-warms keys for brains the user is actually a recipient on (read from manifests).

Encoding the convention in each tool independently would be parallel-mechanism failure. The constitution is the one source of truth.

## Why not by directory name

Real repo names will vary: `brain`, `family-brain`, `brain-private`, `xianxu-brain`, etc. Naming is a UX hint, not a security primitive. Conflating them is brittle — a renamed checkout would silently lose its brain-ness, or a non-brain repo named `brain` would accidentally get brain-aware behavior. The manifest is explicit, auditable, and decoupled from on-disk path.

## Where the depth lives

Full schema rationale, security posture, threat boundaries, and per-mode implications live in `brain/atlas/threat-model-shared-brain.md`. This atlas entry is the constitutional pointer; the threat model carries the load-bearing analysis.

## History

- 2026-05-05 — convention added to `AGENTS.md` §1 under `ariadne#22` M1, propagated downstream via `make refresh` in `ariadne#22` M2. Spec'd from `brain/atlas/threat-model-shared-brain.md` (authored under `nous#8` M1).
