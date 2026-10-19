# Require a confirmation guard on stored destructive actions

## Spec

This is a self-contained proposed change in a small, otherwise unspecified Go
filesystem utility. All relevant code context is defined here, rather than asserted
about the review checkout. There are no external dependencies or other workstreams.

`action.go` defines `DecodeAction([]byte) (Action, error)`,
`EncodeAction(Action) ([]byte, error)`, and `Execute(Action, Confirmer) error`.
Stored JSON has required `version` (integer 1), `path` (nonempty string), and
`operation` (the string `delete`). The decoder rejects unsupported versions before
executing any action. Its shared recursive scanner rejects duplicate keys; known
fields retain strict type and presence checks. Unknown members at every object
level are captured as raw JSON values and merged back unchanged on every rewrite.
Version-1 execution deletes the selected path without asking for confirmation.

Add required boolean `require_confirmation` to newly created version-2 actions.
When true, `Execute` must obtain affirmative confirmation before deletion. A
version-1 executor would ignore this field and delete without confirmation,
violating a safety condition on a destructive action. Therefore version 2 is a
necessary incompatibility boundary: old readers already reject it before acting.
The new reader accepts version 1 with its existing unguarded semantics and
version 2 only with a correctly typed, present guard. Never upgrade old records
implicitly or infer consent. Explicitly creating a guarded action writes version 2.

## Done when

A version-2 guarded action cannot delete without affirmative confirmation; an
old reader refuses it before action execution. Version-1 behavior is unchanged.
Both accepted formats preserve unknown members through rewrites. Required fields,
known types, and duplicate keys remain strict at every object level.

## Plan

- In `action.go`, add the version-2 branch to `DecodeAction` and `EncodeAction`,
  reuse their raw-member retention and shared scanner, and enforce the decoded
  guard at the existing `Execute` seam before the filesystem operation.
- In `action_test.go`, test `Execute` through an in-memory filesystem and fake
  `Confirmer` across denied/error/affirmative responses; test both codec functions
  with round-trip properties over nested future members and fuzz malformed input
  for duplicate/type/required-field rejection. Verify the old codec refuses the
  new format before any execution using the same stateful filesystem fake.
- Run `go test ./...` and a bounded decoder fuzz run. Document the explicit
  safety reason for version-2 refusal in `README.md`. Land as one atomic change.

Non-goals: retroactively guarding version-1 actions, migrations, network access,
new storage abstractions, and unrelated filesystem cleanup.
