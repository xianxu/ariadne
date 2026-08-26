Resolve the normalized admission key and tagged capacity for one prospective
path. `--path` selects that target and defaults to `.`. The path may be a file or
directory that does not exist yet: its deepest existing ancestor is resolved
through symlinks before the missing suffix is preserved.

The declaration is always loaded from the containing repository's primary
worktree. The default output is the human rendering of the typed policy result;
`--json` emits the same result contract.

Missing, invalid, and outside-scope declarations print their typed diagnostic to
stdout and return a non-zero exit without usage text. No repository name is used
to infer policy.
