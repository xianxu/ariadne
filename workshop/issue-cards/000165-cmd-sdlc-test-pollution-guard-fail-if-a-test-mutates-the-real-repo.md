---
id: '000165'
status: punt
created: 2026-07-05
updated: 2026-07-05
---

# cmd/sdlc test-pollution guard: fail if a test mutates the real repo

## Problem

`cmd/sdlc`'s tests build throwaway git repos (merge e2e, issue e2e, close, repolock,
activetime, …). They isolate via `cmd.Dir=<temp>` / `-C <temp>` and/or `os.Chdir`
into a `t.TempDir()`. But there is **no backstop**: a test that forgets to isolate
(a `git(t, "", …)` with empty dir → process cwd, a `chdir`-without-restore, a
relative path, or a git op that finds the enclosing repo) silently operates on the
**live ariadne repo** — and `go test ./cmd/sdlc/` runs from `cmd/sdlc/`, which is
*inside* it.

This is not hypothetical. During #148 (2026-07-05) a test-harness git sequence
(`add -A && commit -m seed` → `switch -c feature` → `commit -m work` → `switch main`
→ `merge feature`) leaked into the real repo: it committed a stray `f` + a bare
`origin.git/` + swept the untracked `workshop/` files, then fast-forwarded `main`
onto the junk — corrupting **local main AND `origin/main`** (recovery required a
force-push + cherry-pick salvage of the real #148 commits). The committed code was
later proven clean (a severed-clone re-run of the full suite produced zero
pollution), so the trigger was a transient intermediate state — but nothing would
have *caught* it, and nothing prevents the next one.
