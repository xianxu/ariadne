---
id: '000124'
status: done
started: 2026-06-25T12:14:37-07:00
created: 2026-06-24
updated: 2026-06-25
estimate_hours: 4.55
actual_hours: 2.02
---

# Instance-conformance: validate typed-markdown artifacts against their datatype schema (extract then cue vet)

## Problem

#122 makes `construct/vocabulary/issue.cue` the formal type of an issue (frontmatter
enum/fields + lifecycle), and wires `sdlc`'s *verbs* to it. But nothing validates that
a real `workshop/issues/NNNNNN-*.md` **file** conforms to that type. A direct edit —
an LLM "tidying" frontmatter to `status: in-progress` (the enum says `working`), a
bulk migration, a mangled merge, a `nous-resolve` AI-merge — bypasses the verb guards
and lands an ill-formed instance that no tool flags. Symptoms are silent: `claim`
assumes already-claimed, `close` clobbers, `state` mis-categorizes. Because issue files
publish to `main` (the bulletin board), the bad value propagates to peer agents.

The shell around a typed artifact's *shape* is today only **prose** (`AGENTS.md`: "do
NOT hand-edit `status:`") — a soft shell an LLM violates by default, since the Edit
tool knows nothing about the lifecycle. The vocabulary layer makes a **hard** shell
possible for the first time (there's now a formal `#Issue` to vet against); this issue
builds it. Generalizes beyond `issue`: it's the **instance-conformance** closed-loop
for *all* typed markdown (datatype = prose skill + schema). Design captured in
`workshop/pensive/2026-06-24-01-pensive-cue-schema-layer-nouns-verbs.md` ("Typing
markdown — the general frame").
