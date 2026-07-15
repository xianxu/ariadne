Move a markdown artifact to a peer repo, rewriting repo-relative refs so
they resolve identically from the destination (#179). Deterministic, no
LLM — the fix for the silent trap where a moved file's bare `#NNN` refs
re-resolve against the destination's issue numbering.

USAGE

  sdlc migrate <file> <dest-repo-dir>
  sdlc migrate data/project/metis-v2.md ../kbench

REWRITE RULES (the ref grammar authority is `sdlc resolve`'s parser)

  #N            → <source>#N     bare refs are source-relative; qualify them
  <dest>#M      → #M             dest-qualified refs become local
  <other>#K     unchanged        source-qualified / third-repo refs still resolve
  gh#N          reported only    github refs are repo-relative; verify by hand

  Fenced code blocks (``` … ```) pass through verbatim — documents about
  the ref grammar quote it literally. An inline `code span` is rewritten
  only when its WHOLE content is a single ref (`#171`); spans quoting
  commands or grep patterns are skipped and listed in the report.

  Every rewrite is VERIFIED from the destination's vantage before anything
  is written: the new form must resolve from the dest repo (this also
  catches a dest that is not a sibling of the source). A non-resolving
  rewrite aborts the migration — nothing half-moved.

WHAT IT DOES

  1. Guards: source file tracked + unmodified, inside the current repo;
     not an id-keyed issue-family artifact (those need dest-side
     renumbering — v2); dest is a git repo, not this repo, NOT a brain
     (SDLC process artifacts don't live in brain, #171); dest path free;
     dest repo clean (--no-clean-check to override).
  2. Rewrites + verifies (rules above), printing every rewrite as
     `line N: old → new` plus a skipped-report.
  3. Writes the file at the destination (same repo-relative path, or
     --dest-path), removes the source, stages EXPLICIT paths on both
     sides, and commits each side with a scoped `migrate:` subject —
     or stages only, with --no-commit.
  4. Sweeps sibling repos for inbound references to the OLD path
     (report-only: issue refs survive moves; path references don't).

  Deliberately runnable IN a brain repo: migrating an artifact OUT of
  brain is the #171 use case, so this verb is not in the spine guard set;
  the brain check applies to the DESTINATION instead.

  Lock note: the .git/sdlc.lock transaction lock covers the CURRENT
  (source) repo; the destination write is a cross-repo operation outside
  that lock — same posture as close's project-file write.

KNOWN LIMITS (v1)

  - "PR #96" prose scans as issue 96 and, if it exists, rewrites — review
    the printed summary (--no-commit + git diff is the review path).
  - A hex color like #123456 scans as a ref and fails verification →
    the migration refuses; fence or reword it.
  - The space-separated form `ariadne #15` is scanned as bare `#15` and
    re-qualified (producing `ariadne src#15`) — use the attached form.
  - gh#N refs and mixed-content code spans are reported, not rewritten.
  - Issue-family files (workshop/{issues,plans,history}/NNNNNN-*) refuse:
    issue IDs are per-repo sequences; renumbering migration is v2.

FLAGS

  --dest-path <rel>    repo-root-relative path at the destination
                       (default: same as the source's)
  --no-commit          stage both sides; print the two commit commands
  --no-clean-check     proceed even if the destination repo is dirty

EXIT CODES

  0   moved (or staged with --no-commit)
  1   any guard refusal or verification failure — nothing moved

RELATED

  sdlc resolve    the ref grammar + fleet-wide artifact resolution
