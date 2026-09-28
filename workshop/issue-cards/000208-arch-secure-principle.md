---
id: '000208'
status: done
started: 2026-09-02T15:19:59-07:00
created: 2026-09-02
updated: 2026-09-02
estimate_hours: 1.13
actual_hours: 1.72
---

# ARCH-SECURE: a security lens in the principle registry

## Problem

The ARCH-* registry has five entries — DRY, PURE, PURPOSE, MOCK, CONSTRAINTS —
and **zero security-ish words across all of them**: no threat, secret,
credential, untrusted, injection, privilege. Every other quality axis assumes
indifferent inputs; nothing in the registry asks what happens when an input is
chosen by someone who wants the system to fail.

The fleet is dev tooling that runs on a host behind firewalls, which removes the
remote attacker and makes the gap easy to live with. But the threat model that
*survives* a firewall is credentials and local state, and that is exactly what
recent work kept hitting. From parley.nvim#205 alone, three security-shaped
defects were caught only incidentally, each filed under a non-security marker:

- a spec run overwrote the operator's **live** cliproxy config with a test API
  key; the running proxy reloaded it and began rejecting their own bearer
  (landed as a test-hygiene lesson)
- opening the agent picker **minted a 0600 `management.key`** as a side effect of
  gathering host/port (landed as ARCH-PURPOSE)
- a persisted `catalog.json`, readable across sessions and versions, crashed the
  picker on a malformed row (landed as a crash bug)

None was found by a security lens, because there is none. Each was found by a
reviewer noticing something adjacent.

There is a second, higher-level concern — authority gradients and blast radius —
which is deliberately NOT in scope here; see Spec.
