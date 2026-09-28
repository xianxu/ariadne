---
id: '000152'
status: done
started: 2026-07-01T10:40:28-07:00
created: 2026-07-01
updated: 2026-07-01
estimate_hours: 0.47
actual_hours: 0.42
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
