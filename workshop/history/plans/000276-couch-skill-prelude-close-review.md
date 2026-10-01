# Boundary Review — ariadne#276 (whole-issue close)

| field | value |
|-------|-------|
| issue | 276 — Add Couch skill prelude |
| repo | ariadne |
| issue file | workshop/issues/000276-couch-skill-prelude.md |
| boundary | whole-issue close |
| milestone | — |
| window | c634a11d498877834858b482df99f0729da9c6dc..ce2df1ae91e4ab97567439e093042925eb52e579 |
| command | sdlc close --issue 276 |
| reviewer | claude |
| timestamp | 2026-09-30T22:16:26-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This boundary meets the issue's purpose. It adds a static pointer skill at `construct/local/couch/SKILL.md`. The skill points agents to `couch --skill`, says that loading it sends nothing, and tells the agent to stop and explain when `couch` is missing. It copies none of the Couch protocol. It follows the existing `construct/local/sdlc` pattern: the frontmatter is `name: couch`, the same way the sdlc skill uses `name: sdlc`, and the build step turns it into `xx-couch`. It reaches downstream repos through the existing `skill construct/local` line in `construct/base.manifest:103`, so no manifest change is needed. The `.gitignore` lines and the process-manual changes are generated output. `.claude/skills/xx-couch` correctly links to `../../construct/local/couch`. I found nothing that blocks shipping.

1. **Strengths**
   - `construct/local/couch/SKILL.md:7-11`: the protocol has one owner, the binary. The skill says so and tells the agent to re-read it each time instead of relying on memory.
   - `SKILL.md:19-23`: the case where `couch` is missing is handled with a clear stop. It also bans makeshift substitutes, so an agent won't invent its own channel.
   - The Plan records why a weave-generated skill was rejected: it would duplicate the protocol, and it would make `weave compile` fail on machines without Couch. That reasoning is sound.
   - The downstream check was done with a throwaway derivative repo, not just argued.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `atlas/index.md:26`: the entry links straight to a `SKILL.md` instead of an atlas page. That is acceptable for one pointer skill. The other skills appear only in the generated manual, so this entry is the only hand-written skill pointer in the index. This is a minor inconsistency, not a defect.
   - Regenerating `atlas/workflow/process-manual.md` also pulled in help-text changes from earlier work (`move`, `workspace`, and edits to `claim`, `close`, `merge`, `pr`, `set-status`, plus the lessons header). That catch-up is correct but outside this issue's scope. The Log already says so.

5. **Test coverage:** the only executable surface is a Markdown prompt plus its generated links. Evidence for the Done-when clauses:
   - **The prelude delegates to `couch --skill`:** confirmed by reading the file.
   - **Downstream discovery:** the throwaway-derivative run is logged.
   - **Distribution checks:** the Log records `go test ./cmd/weave/... ./pkg/...` and `make weave-drift-check` passing.
   - **No copied protocol:** confirmed by reading the file.
   - **Couch present vs. absent:** the prompt covers both cases. The binary isn't involved, so there's no runtime branch to test.

6. **Architecture**
   - **ARCH-DRY: pass.** The skill repeats none of the protocol and reuses the existing local-skill route.
   - **ARCH-PURE: pass.** There is no code logic in this change.
   - **ARCH-PURPOSE: pass.** The only consumer of the protocol is `couch --skill` itself. The skill is reached through normal distribution, which covers every downstream repo, so nothing is left as a hand-maintained copy.

7. **Plan revisions:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: atlas-index-entry-shape
    title: |
      atlas/index.md entry links directly to a SKILL.md rather than an atlas page
    detail: |
      Only instance in this window is atlas/index.md:26. Other skills appear only in the generated process manual, so this is the sole hand-written skill pointer in the index. Acceptable for a one-file pointer skill.
```
