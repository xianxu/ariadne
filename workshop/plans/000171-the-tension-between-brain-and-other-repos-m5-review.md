# Boundary Review — ariadne#171 (milestone M5)

| field | value |
|-------|-------|
| issue | 171 — the tension between brain and other repos |
| repo | ariadne |
| issue file | workshop/issues/000171-the-tension-between-brain-and-other-repos.md |
| boundary | milestone M5 |
| milestone | M5 |
| window | b8484a7fcdd965ef2c0f12cae15d538b066036af..HEAD |
| command | sdlc milestone-close --issue 171 --milestone M5 |
| reviewer | claude |
| timestamp | 2026-07-17T21:30:26-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

All verification is done. Writing up the review now.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M5 delivers what it claims: the residency model (projects live in the center-of-gravity repo's `workshop/projects/`, never in brain; brain is capture/measurement only) now reads consistently across `AGENTS.base.md`, all three woven faces (verified on disk at lines 20/62 of `AGENTS.md`/`CLAUDE.md`/`GEMINI.md` — propagate-base demonstrably ran), the atlas, and the project datatype doc, and the plan's Revisions entry honestly records the execution deltas. What keeps this from SHIP is the ARCH-PURPOSE shadow-sweep: this was *the* docs boundary of the lift, and the embedded `--help` text — user-facing surface, the exact #142 class the checklist names — still describes the brain-era project gate in three places, plus one same-file contradiction inside the datatype doc M5 edited. All fixes are cheap prose edits; none block at the gate.

### 1. Strengths

- **Propagate actually propagated.** All four constitution faces carry the new §Peer-Repo and §8 text verbatim — the "compiled to consumers" promise is real, not claimed (`AGENTS.md:20,62`, `CLAUDE.md:20,62`, `GEMINI.md:20,62`).
- **The Revisions entry is a model of honest bookkeeping** (`workshop/plans/000171-cross-repo-project-lift-plan.md:843-863`): gitignored woven faces amending the commit list, the two dirty dependents deliberately skipped, the sandbox constraint on `make weave`, and the M3 peer-write guarantee folded into the brain line — deltas recorded, not papered over.
- **The atlas caveat is a real map entry, not filler** (`atlas/workflow/sdlc-binary.md:191-194`): extending the sibling-discovery-model caveat to project discovery was exactly the M4 reviewer's handoff note, and it landed.
- **`construct/datatype/project.md:217-226` residency item is well-shaped**: soft rule vs. address distinction, live-move vs. frozen-verbatim-move distinction, and the discovery/navigation tooling named — the doc teaches the model, not just the path.
- **§8 rewrite keeps discovery as tooling's job** ("Discovery is tooling's job: the close gate finds and ticks…") — the prose derives from what the binary enforces rather than restating a manual discipline (ARCH-PURPOSE done right on that consumer).

### 2. Critical findings

None. Docs-only window; no behavior drift, no correctness risk.

### 3. Important findings

1. **`cmd/sdlc/helptext/close.md:76-77` — the close gate's `--help` still describes the brain-era project gate** ("project file (if any, under `<brain>/data/project/*.md` referencing…"). The gate has discovered fleet-wide since M2 and ticks *every* match. Same file line 137: `--brain-dir <path>  project-file lookup root (default ../brain)` — wrong since M2; the Go flag registration was already corrected to "for the calibration ledger; project files are now discovered across the fleet, #171" (`close.go:140`), so the embedded helptext now contradicts the cobra flag help one screen apart. `cmd/sdlc/helptext/milestone-close.md:69` has the identical stale flag line. Fix: reword all three to the fleet-discovery model + legacy-home deprecation. This is the #142 docs-gate class the checklist names, and an ARCH-PURPOSE shadow-sweep miss — the helptext is a hand-maintained restatement of the residency model that didn't get swept at the sweep boundary (ARCH-DRY is the root cause: three prose restatements of one flag's meaning).
2. **`construct/datatype/project.md:88` (and the example at `:164`) contradicts the residency charter shipped 130 lines below.** The ref-grammar section says issue refs work uniformly for "shared brain repos (`brain-team`, `brain-family`, etc.)" — i.e., brain repos with issue trackers — while item 7 (`:222-223`, added this milestone) states brain holds **no** SDLC process artifacts, and #176's guard refuses spine verbs in any brain repo. A reader of this one file gets both "brain repos host issues" and "never in brain." Fix: drop or reword the brain-repo example (e.g., "any peer repo with a `workshop/issues/` tracker").
3. **The Revisions entry claims a Log note that doesn't exist.** Delta 2 says the propagate-base re-run (for the skipped dirty `42shots`/`pair` dependents) is "noted in the issue Log for a re-run at issue close" — but the issue file contains no such note (grep for propagate/42shots/re-run: only the estimate footnote matches). Since the Done-when includes "downstream repos re-woven," the re-run reminder needs to live somewhere the close will actually see. Fix: fold it into the M5 `closed M5` Log line being written at this boundary (or add a standalone Log note now).

### 4. Minor findings

- `cmd/sdlc/helptext/state.md:46-47`: the deferred-feature note still names "BRAIN_DIR resolution" as the future mechanism for the project-tick drift check; the mechanism is now fleet discovery.
- `scripts/close-issue.py:15` still documents `$BRAIN_DIR/data/project/` and is distributed fleet-wide via `construct/base.manifest:118` despite being fully ported to `sdlc close` (`close.go:1`) — retire it from the manifest or annotate it as superseded.
- `construct/datatype/project.md:224` cites "#171 M6's precedent" for a milestone that hasn't executed yet — a forward reference reading as settled precedent; harmless once M6 lands, momentarily misleading now.
- The brain-peer line's "(`roadmap` remains until it too lifts — #171)" points at #171, whose scope doesn't include the roadmap lift; consider filing the roadmap-lift follow-up and pointing there instead.

### 5. Test coverage notes

Docs-only window — nothing testable shipped, appropriately. Plan Step 4's `go test ./cmd/sdlc/ -run Propagate` could not be re-run in this review (the Bash tool failed at the harness level — `mkdir ~/.claude/session-env/...` EPERM, even with sandbox disabled); I verified the propagate outcome by direct inspection of all four faces instead, which is the stronger check for this boundary. The stale-helptext class (finding 1) has no automated guard — helptext markdown is hand-maintained prose with no drift check against flag registrations; that's the M1-noted "generated-face staleness depends on reviewer catch" pattern in a second wardrobe, worth an eventual lint.

### 6. Architectural notes

- **ARCH-DRY — flag with a pass on the shipped files.** The deliberate four-surface restatement (base constitution → woven faces, datatype, atlas) is the designed compilation model and the faces verifiably derive. But the helptext markdowns are a *fifth, underived* restatement layer, and finding 1 is the drift it produces. Upcoming M6 shouldn't add prose to helptext without asking whether the line restates something `sdlc` already knows.
- **ARCH-PURE — pass (vacuous).** No code in the window; propagate-base's pure-weave/thin-IO split untouched.
- **ARCH-PURPOSE — flag (findings 1–2).** The shadow-sweep across residency-model consumers: constitution ✓, faces ✓, atlas ✓ (`sdlc-binary.md`, `ledger-landscape.md:17`, `vocabulary.md:21`), README ✓ (`README.md:17`, from M4), datatype ✓ — but helptext ✗ and one intra-datatype contradiction ✗. The purpose of a docs milestone *is* the sweep; the misses are the findings above, all cheap.
- For M6: the migrate helptext example (`helptext/migrate.md:9`, `data/project/metis-v2.md`) is currently correct-by-accident (metis-v2 legitimately still lives in brain); after M6 + metis-v2's eventual move, that example becomes the last brain-path teaching text — sweep it then.

### 7. Plan revision recommendations

One entry needed, piggybacking on finding 3: amend the 2026-07-17 M5 Revisions delta 2 (or satisfy it) so "noted in the issue Log" is true — either the M5 close Log line carries the propagate-base re-run reminder for `42shots`/`pair`, or the revision text is corrected to name where the reminder actually lives. Optionally append the helptext files to the same entry's sweep record once fixed, so the plan's account of the M5 sweep matches what was actually swept.
