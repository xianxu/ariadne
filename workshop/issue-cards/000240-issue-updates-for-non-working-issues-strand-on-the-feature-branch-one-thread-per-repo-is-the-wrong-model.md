---
id: 000240
status: open
created: 2026-09-20
updated: 2026-09-21
estimate_hours:
github_issue:
---

# Issue updates for non-working issues strand on the feature branch — one thread per repo is the wrong model

## Problem

Working #239 on a feature branch, three unrelated issues (#236, #237, #238) had
in-flight edits in the working tree. There was no correct place to put them.

**The asymmetry, measured this session:**

| Verb | From a feature branch | Result |
|---|---|---|
| `sdlc issue new` | **publishes to the trunk** ("Issue changes are on the trunk") | ✅ filing works |
| `sdlc issue sync --issue N` | commits **locally, to the current branch** ("not pushed") | ❌ update strands |

So *filing* an issue from anywhere lands on main, which is right. *Updating* one
lands on whatever branch you happen to be standing on. For the issue you are
working, that is fine — it merges with the work. For any **other** issue it is
wrong twice over:

1. The update is invisible to peer agents until an unrelated feature branch
   merges — the exact thing `claim`/`sync` exist to prevent (AGENTS.md §2:
   "Issue files are workflow state… They need to land on `main` quickly so peer
   agents see status changes without waiting for the feature branch to merge").
2. It couples unrelated content to that branch's fate. If #239 is abandoned or
   rebased, #236–#238's edits go with it.

**The workaround is worse than the gap.** With nowhere to put them, the files sit
dirty across the whole session, and every `git add -A` / `git add <dir>` sweeps
them into an unrelated commit. That happened **twice** in this session; one of
them was `.claude/settings.ariadne.json`, a **fleet-propagated** egress
allowlist, which would have reached every derivative inside a #239 commit,
undeclared. Both were caught and backed out, but only by inspection. See
`workshop/lessons.md` ("Stage by path, never by directory…").

There is also no verb at all for a **non-issue** edit that belongs on main — the
settings file above had to be committed onto the #239 branch with a `side-quest:`
message explaining why, which is a workaround, not a home.

**Secondary friction:** `sdlc issue sync` takes exactly one `--issue N` ("the
commit message names it"), so N dirty issues need N invocations. Reasonable given
the one-commit-per-issue contract, but it compounds the above.
