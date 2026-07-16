---
id: 000152
status: done
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-01
estimate_hours: 0.47
started: 2026-07-01T10:40:28-07:00
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

- [x] `sdlc claim` / `sdlc pr` / `sdlc merge` complete their git push/pull steps
      with the sandbox **enabled** (no `dangerouslyDisableSandbox`, no approval).
      *(Dogfooded: `sdlc claim #152` + `sdlc pr` (PR #74) both pushed over HTTPS
      with the Claude Code sandbox ON.)*
- [x] The chosen mechanism is applied and documented (host `insteadOf` +
      openshell overlay fix), covering ariadne + all `git@github.com:` peer repos.
      **brain/brain-family (`gcrypt::ssh://`) are deliberately scoped OUT** — a
      sensitive, separate follow-up (see Revisions), not covered here.
- [x] Non-sandbox workflow still works (push/pull/signing unaffected). *(HTTPS
      `git ls-remote origin` works host-side with the same gh creds; no commit
      signing is configured (`git config` shows none), so the transport switch
      can't affect signing.)*
- [x] The durability decision is recorded: one-time operator step vs. baked into
      the ariadne setup path (with a pointer from the sandbox atlas doc).
      *(Recorded in Revisions: host = manual for blast-radius; container = the
      openshell overlay; atlas documents both.)*

## Plan

- [x] Apply the global `insteadOf` (both forms, `--add`) to the **host**
      `~/.gitconfig` — mirroring openshell's container config, minus
      `http.sslVerify=false` (host TLS is real). [done during implementation]
- [x] Fix the `--add` bug in `.openshell/overlay/setup.sh` so the OpenShell
      container actually rewrites `git@github.com:` (today the 2nd line clobbers
      the 1st, leaving only `ssh://` — ARCH-DRY: one multi-valued key).
- [x] Verify `sdlc claim`/`pr`/`push` git ops run sandbox-ON (dogfooded: this
      issue's claim pushed to origin over HTTPS in-sandbox).
- [x] Document the git-transport story in `atlas/workflow/openshell-sandbox.md`:
      container (overlay) + host (one-time) + the brain/gcrypt caveat.
- [x] Note the brain/`gcrypt::ssh://` remotes are NOT switched (separate, sensitive
      — encrypted; needs tested `gcrypt::https://`).

## Revisions

### 2026-07-01 — design refined after finding the openshell precedent

Implementing revealed the mechanism is **already established** (and the gap is
narrower than the Spec assumed):

- `.openshell/overlay/setup.sh:11-14` **already** runs the HTTPS `insteadOf`
  (both `git@github.com:` and `ssh://git@github.com/`) + `http.sslVerify=false`
  for the OpenShell **container**. So that sandbox path is (nominally) covered.
- The real gap is the **host** `~/.gitconfig` — which the **Claude Code** sandbox
  (sandbox-exec on the host) uses, and which had no `insteadOf`. Applied there
  (with `--add`, both forms; no `sslVerify` change — host TLS is real). Decisive
  test: `git push --dry-run origin main` (SSH URL → rewritten HTTPS) now succeeds
  **in-sandbox**, and this issue's `sdlc claim` pushed over HTTPS with the sandbox
  ON.
- **Bug found:** those two openshell lines lack `--add`. `git config url.X.insteadOf`
  is single-valued without `--add`, so line 12 **overwrites** line 11 — the
  container only ever gets the `ssh://` rewrite, NOT `git@github.com:` (the form
  real origins use). Fixing that is now part of this issue.
- **Durability decision (recorded):** container = automated via the openshell
  overlay (once the `--add` bug is fixed); **host = one-time `git config --global`**,
  documented in atlas. The load-bearing reason to keep the host step *manual* is
  **blast radius** — auto-applying a *global* transport rewrite from a base-layer
  script (`bootstrap-peers.sh` et al.) would silently flip git transport for every
  downstream user, sandbox or not. (There's also no pre-clone host hook to bake it
  into, but blast radius is why we wouldn't even if there were.) Documented
  one-time step it is; migrate it into a host post-clone phase if one ever exists.
- **brain / brain-family** (`gcrypt::ssh://git@github.com/…`) are NOT switched:
  the `git@github.com:` prefix doesn't match `gcrypt::ssh://…`, so they're
  untouched (confirmed). Switching an encrypted gcrypt remote's transport is a
  separate, sensitive change (needs a tested `gcrypt::https://` + GPG-in-sandbox).
  Left as a documented follow-up.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.* Design pre-resolved (mechanism established by
openshell; +15% buffer); impl at v3.1's 40%-of-v2 unit.

- atlas-docs: document the git-transport story (container + host + gcrypt caveat)
  — design 0.1 + impl 0.1.
- smaller-go-module: the `--add` fix in `.openshell/overlay/setup.sh` + applying +
  verifying the host config — impl 0.1.
- milestone-review: the one boundary review auto-dispatched at `sdlc close` —
  impl 0.15.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: atlas-docs         design=0.1 impl=0.1
item: smaller-go-module  design=0.0 impl=0.1
item: milestone-review   design=0.0 impl=0.15
total: 0.47
```

## Log

### 2026-07-01
- 2026-07-01: closed — Host ~/.gitconfig insteadOf applied (both git@github.com: and ssh:// forms, via --add); git push --dry-run origin main → HTTPS rewrite → "Everything up-to-date" with the Claude Code sandbox ON (pre-fix it errored on SSH). Dogfooded: sdlc claim #152 pushed to origin over HTTPS sandbox-ON. Fixed the missing-*--add* bug in .openshell/overlay/setup.sh (bash -n clean; git-config block idempotent — exactly 2 insteadOf values on re-run). brain/gcrypt::ssh remotes confirmed unmatched/untouched; atlas documents the container+host+gcrypt story.; review verdict: SHIP

- Filed while landing #140/#141: every `sdlc` push in this sandbox session failed
  on the SSH remote, so pr/merge/claim ran with the sandbox disabled. Verified
  HTTPS git transport (fetch + push --dry-run) works with the sandbox ON, and
  that gh already provides HTTPS credentials — so switching transport removes the
  need to disable the sandbox at all. See the session's #140/#141 landing for the
  concrete failures and the HTTPS verification.
