---
id: 000152
status: open
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-01
estimate_hours:
---

# Switch git remotes from SSH to HTTPS so sdlc verbs run in-sandbox

## Problem

Sandboxed agents (Claude Code, and Codex in its sandbox) cannot run the git
operations inside `sdlc` verbs when `origin` is an **SSH** remote
(`git@github.com:…`). The sandbox network layer proxies HTTP(S) only; raw SSH is
routed through that proxy and fails the handshake:

```
nc: authentication method negotiation failed
Connection closed by UNKNOWN port 65535
```

Concretely, this breaks the push/pull steps of:

- `sdlc claim` — the issue-sync push to main (observed failing every claim this
  session; the issue flip lands locally but never broadcasts to origin);
- `sdlc pr` — `git push -u origin <branch>`;
- `sdlc merge` — `git pull` + the archive `git push` (the `gh pr merge` HTTPS call
  itself works);
- `sdlc issue new` — the auto-sync push (this very ticket's creation push failed).

Today the only workaround is disabling the sandbox for every network op, which
defeats the point of the sandbox and forces per-command approval.

**The transport is incidental, not required.** Git itself works fine over HTTPS
inside the sandbox — verified this session: `git ls-remote` and
`git push --dry-run` against `https://github.com/xianxu/ariadne.git` both
succeeded (exit 0) with the sandbox ON. `github.com` is already on the sandbox
host-allowlist (that is why HTTPS works and SSH does not). Auth is already wired:
`gh config get git_protocol` = `https` and gh is the HTTPS credential helper (the
`failed to store: 100001` line is only the macOS keychain *cache* write being
blocked — the request still authenticates).

So the fix is to stop using SSH transport, not to try to open SSH in the sandbox
(which the host-allowlist cannot grant).

## Spec

Make every ariadne-ecosystem repo's git transport HTTPS so `sdlc`'s
push/pull/claim/pr/merge run **fully inside the sandbox** with no disabling and
no per-command approval.

Preferred mechanism — a single global rewrite (keeps SSH URLs in every repo's
config, rewrites to HTTPS at runtime, covers ariadne + brain + all peer repos and
any future clone at once):

```
git config --global url."https://github.com/".insteadOf "git@github.com:"
```

Alternatives considered:
- Per-repo `git remote set-url origin https://github.com/<owner>/<repo>.git` —
  explicit and visible in `git remote -v`, but must be repeated per repo and per
  new clone; misses the multi-repo goal.
- Opening SSH in the sandbox — rejected: the Claude Code sandbox network is an
  HTTP(S) proxy (`/sandbox` + `settings.json` host-allowlist); it cannot grant
  raw SSH. Codex has its own `~/.codex/config.toml` sandbox, same limitation.

Open design questions to resolve while implementing:
- **Where does this live so it is durable + reproducible?** A one-time documented
  operator step vs. baking it into the ariadne setup path (`openshell-sandbox` /
  bootstrap / weave) so a fresh sandbox/clone gets it automatically. `insteadOf`
  is per-machine git config, not repo-tracked — decide whether setup applies it.
- **Blast radius:** the global rewrite affects non-sandbox use too (push/pull go
  over HTTPS + gh token instead of SSH keys). This is seamless with gh, but
  confirm it does not disturb SSH-based commit signing or any key-only flow.
- **Interaction with the sandbox's blocked keychain write** (`failed to store:
  100001`): confirm credentials resolve without the keychain cache on repeat
  operations (it did this session, but document it).

## Done when

- [ ] `sdlc claim` / `sdlc pr` / `sdlc merge` complete their git push/pull steps
      with the sandbox **enabled** (no `dangerouslyDisableSandbox`, no approval).
- [ ] The chosen mechanism is applied and documented (global `insteadOf` and/or a
      setup step), covering ariadne + brain + peer repos.
- [ ] Non-sandbox workflow still works (push/pull/signing unaffected).
- [ ] The durability decision is recorded: one-time operator step vs. baked into
      the ariadne setup path (with a pointer from the sandbox atlas doc).

## Plan

- [ ] Decide mechanism: global `insteadOf` (recommended) vs. per-repo set-url vs.
      setup-path automation. Capture the call + rationale.
- [ ] Apply it; verify `sdlc pr`/`merge`/`claim` push/pull run sandbox-ON against
      a throwaway branch.
- [ ] Confirm non-sandbox git (push/pull, any commit signing) is unaffected.
- [ ] Document in `atlas/workflow/openshell-sandbox.md` (and, if automated, wire
      into the setup/bootstrap path with a note in `atlas/workflow/base-layer.md`).

## Log

### 2026-07-01

- Filed while landing #140/#141: every `sdlc` push in this sandbox session failed
  on the SSH remote, so pr/merge/claim ran with the sandbox disabled. Verified
  HTTPS git transport (fetch + push --dry-run) works with the sandbox ON, and
  that gh already provides HTTPS credentials — so switching transport removes the
  need to disable the sandbox at all. See the session's #140/#141 landing for the
  concrete failures and the HTTPS verification.
