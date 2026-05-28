# Construct: Adaptation is Ariadne-Only

The construct system imports external skill sources, adapts them via intent
transcripts, and deploys them. There is exactly **one** adaptation target:
ariadne itself. Derivative repos inherit the rendered output verbatim — they
never run `/construct adapt`.

## Why single-target

Ariadne is an opinionated stack on how to leverage AI in software development.
The premise of an opinionated stack is that downstream consumers inherit the
opinions, they don't redo them. An adaptation pipeline with per-derivative
intents is more machinery than the actual use case warrants — in practice,
every derivative (nous, parley.nvim, ...) converges to ariadne's adaptation.
Historical context lives in issue [#33](../../workshop/history/000033-adaptation-system-narrow-to-ariadne-only.md).

## How inheritance works

Ariadne's promote step writes the rendered skills to two places at once:

- `$REPO_ROOT/.claude/skills/<source>-<skill>/` — ariadne's own live skill set
- `$REPO_ROOT/construct/adapted/<source>-<skill>/` — the inheritable copy

Derivative repos pick `construct/adapted/` up through `construct/base.manifest`:

```
# in construct/base.manifest
symlink construct/adapted
```

`construct/setup.sh` resolves that line into a symlink (default mode) or a
vendored copy (`--vendor`) in the derivative's tree. Then the derivative's
`.claude/skills/<source>-<skill>/` directories are themselves symlinks pointing
at `../../construct/adapted/<source>-<skill>/`. The chain is:

```
<derivative>/.claude/skills/superpowers-brainstorming/
    → ../../construct/adapted/superpowers-brainstorming/
    → (in symlink mode) ariadne/construct/adapted/superpowers-brainstorming/
```

When ariadne promotes a new version, derivatives pick up the change the next
time they refresh — no per-derivative adaptation step exists.

## What this rules out

- **Per-derivative intent transcripts.** There is one intent file per source
  (`construct/intents/<source>.md`), not per source+target pair. A change to
  ariadne's adaptation is a change to *everyone's* adaptation.
- **Personal-scope skills.** The construct system does not deploy to
  `~/.claude/skills/`. Personal skills are managed outside construct.
- **Cross-derivative customization through construct.** If a single derivative
  needs different behavior from a skill, the options are (a) propose the
  change upstream to ariadne's intent, accepting that everyone gets it, or
  (b) override locally outside construct's pipeline (e.g. a sibling skill at
  the derivative's `.claude/skills/`).

## Where to look

- `construct/skill/SKILL.md` — full command reference for `/construct adapt`,
  `/construct promote`, etc.
- `construct/intents/superpowers.md` — the live ariadne adaptation transcript.
- `construct/base.manifest` — declares which base-layer files derivatives inherit.
- [`setup-and-replication.md`](setup-and-replication.md) — how `construct/setup.sh`
  wires the symlinks/copies into a derivative.
