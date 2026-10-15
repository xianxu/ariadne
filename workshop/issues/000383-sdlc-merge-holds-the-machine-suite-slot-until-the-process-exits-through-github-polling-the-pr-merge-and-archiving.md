---
id: 000383
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: 'fac27596dccb9cc329d61acc38a123dff6aeee8f' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc merge holds the machine suite slot until the process exits, through GitHub polling, the PR merge and archiving

## Problem

On 2026-10-14 (21:56–22:08) ariadne#377's `sdlc merge --yes` held the machine-wide suite slot for ~12 minutes with no
test process running, while parley.nvim#330's suite waited 2nd in the queue (reported by parley.nvim:2). Cause:
`acquireSuiteSlot` (`cmd/sdlc/localcheck_run.go:99`, `:202–212`) runs `scripts/test-lock.py acquire --watch
<os.Getpid()>`, so the slot is tied to the `sdlc merge` process and released only when it exits; `test-lock.py` has
no release. After the suite finishes, the slot stays held through:
- GitHub status/comment writes;
- the CI re-read and the 30 s pending-poll loop (`cigate_wait.go:190–198`);
- possible re-integration pushes;
- the PR merge (`landing.go:508–529` / `merge.go:641`), archive commits and pushes, and branch cleanup.

Every landing therefore blocks the fleet's suites for minutes beyond its own test run.

## Spec

- The suite slot covers **test execution only:** acquire right before the suite process, release right after it exits
  (e.g. a short-lived child process as the `--watch` pid, or an explicit release in `test-lock.py`).
- A reentrant rerun in the same merge (re-integration, baseline diagnosis) acquires again, in ticket order like anyone
  else.
- No change to admission order, slot counts or the suite command.

## Done when

- A test with the real `test-lock.py` shows the slot is free while `sdlc merge` polls GitHub and merges (a fake GitHub
  client that blocks), and a second waiter is admitted then.
- A rerun inside the same merge re-acquires correctly.
- The local-check docs say what the slot covers.

## Plan

- [ ]

## Log

### 2026-10-14
