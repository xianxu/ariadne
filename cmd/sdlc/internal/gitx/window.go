// Package gitx wraps the small set of git invocations that the sdlc binary's
// checkpoint guards need: commit-window discovery (subject-anchored grep for
// #N references), peer-issue discovery, and changed-file listing.
//
// Ported from scripts/close-issue.py — semantics preserved including the
// commit-window cap (see WindowCapDays; raised to 61 in #68) and the "parent of
// first match" window-start trick.
package gitx

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueref"
)

// run is the package-level command runner. Test code in this package
// (and downstream packages once we propagate the pattern) can override
// it to drive fixture-based scenarios without spawning real git
// processes. Production path defaults to exec.Command(...).Output().
//
// All new git callers in this package should use run; M3+ will migrate
// the legacy direct exec.Command calls below when those code paths are
// touched again.
var run = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// RunGit runs `git <args>` via the package-level `run` shim and returns
// the raw stdout bytes. Use this when you need the full output (newlines
// preserved) or need to distinguish empty-but-OK from error — Capture
// flattens both into "".
func RunGit(args ...string) ([]byte, error) {
	return run("git", args...)
}

// Capture runs `git <args>` and returns trimmed stdout. Empty string on
// any error (caller decides whether to refuse or degrade). Uses the
// package-level `run` shim so tests can override.
//
// Suitable for one-shot queries like `git rev-parse --show-toplevel`,
// `git branch --show-current`, `git worktree list --porcelain`. Not
// suitable for queries where you must distinguish "ran but empty" from
// "errored" — use run() directly for those.
func Capture(args ...string) string {
	out, err := run("git", args...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// DiffBase returns the git ref to compare against for "what's new on
// this branch." Mirrors scripts/lib.sh's git_diff_base():
//
//  1. If <repo-root>/COMPARE-SHA exists and points to a verified ref,
//     use that. Lets callers override the default for ad-hoc reviews.
//  2. If current branch is main, return origin/main (HEAD~10 fallback).
//  3. Otherwise (feature branch), return the merge-base of main and HEAD.
//
// Used by `sdlc judge` to determine the diff window for principle checks.
func DiffBase() string {
	root := Capture("rev-parse", "--show-toplevel")
	if root != "" {
		path := root + "/COMPARE-SHA"
		if data, err := os.ReadFile(path); err == nil {
			sha := strings.TrimSpace(strings.SplitN(string(data), "\n", 2)[0])
			if sha != "" && Capture("rev-parse", "--verify", sha) != "" {
				return sha
			}
		}
	}
	branch := Capture("branch", "--show-current")
	if branch == "main" {
		if ref := Capture("rev-parse", "origin/main"); ref != "" {
			return "origin/main"
		}
		return "HEAD~10"
	}
	if base := Capture("merge-base", "main", "HEAD"); base != "" {
		return base
	}
	return "HEAD~10"
}

// MergeBaseWithMain returns the branch point — `git merge-base main HEAD` — or
// "" when HEAD has not diverged from main (merge-base == HEAD: on main, or a
// branch with no new commits) or merge-base is unavailable. The empty result is
// a deliberate signal: it lets the caller fall back to an *issue-specific* anchor
// rather than a generic one.
//
// Distinct from DiffBase above — same `merge-base main HEAD` core, different
// fallback contract. DiffBase is the `sdlc judge` diff window: it layers in a
// COMPARE-SHA override and generic on-main fallbacks (origin/main / HEAD~10) and
// never returns "". MergeBaseWithMain is the close-window branch point: on no
// divergence it returns "" so boundaryWindowBase picks the issue's own branch
// start (the first `#N` commit's parent) for the direct-on-main flow (#77).
func MergeBaseWithMain() string {
	base := Capture("merge-base", "main", "HEAD")
	if base == "" {
		return ""
	}
	if base == Capture("rev-parse", "HEAD") {
		return "" // no divergence (on main / no new commits) — caller falls back
	}
	return base
}

// MainRef returns the ref that stands for "the published main line":
// origin/main if it verifies, else local main, else "" (fresh repo or a
// master-only checkout). Single source for the origin/main→main→"" fallback
// shared by recentCommits (state.go) and ShippedWorkOnMain below, so the
// resolution lives in one place rather than being copy-pasted.
func MainRef() string {
	if Capture("rev-parse", "--verify", "origin/main") != "" {
		return "origin/main"
	}
	if Capture("rev-parse", "--verify", "main") != "" {
		return "main"
	}
	return ""
}

// bookkeepingVerbs are subject lead-ins that *do* anchor #N but represent
// workflow bookkeeping — filing, claiming, or closing an issue — not shipped
// implementation. (issue-sync commits never anchor #N in their subject —
// they read "issue-sync: update issues" — so they're excluded for free by the
// subject anchor and need no entry here.)
var bookkeepingVerbs = []string{"file issue", "ticket", "claim", "close"}

// IsShippedWorkSubject reports whether a commit subject is *implementation*
// work for issueNum: either `#N ...` or the documented `<area>: #N ...`
// convention, and not a bookkeeping lead-in. Loose refs later in the subject
// (for example `docs: mention #N`) do not anchor ownership. Pure — no git,
// table-tested directly.
//
// The discriminator that keeps a bare `--grep #N` count honest: it separates a
// shipped `#51 M1-M3: …` / `#80: archive stages …` from a `#76: file issue …`
// or `#51: close …`.
func IsShippedWorkSubject(issueNum, subject string) bool {
	rest, ok := issueSubjectDescriptor(issueNum, subject, false)
	if !ok {
		return false
	}
	lower := strings.ToLower(rest)
	for _, v := range bookkeepingVerbs {
		// Whole-token match: "close" is bookkeeping, "close-off" is not.
		if strings.HasPrefix(lower, v) {
			after := lower[len(v):]
			if after == "" || !isWordByte(after[0]) {
				return false
			}
		}
	}
	return true
}

func issueSubjectDescriptor(issueNum, subject string, allowClosePrefix bool) (string, bool) {
	s := strings.TrimSpace(subject)
	if allowClosePrefix {
		s = strings.TrimSpace(strings.TrimPrefix(s, "close "))
	}
	if rest, ok := descriptorAfterIssuePrefix(issueNum, s); ok {
		return rest, true
	}
	if i := strings.Index(s, ":"); i >= 0 {
		return descriptorAfterIssuePrefix(issueNum, strings.TrimSpace(s[i+1:]))
	}
	return "", false
}

func descriptorAfterIssuePrefix(issueNum, s string) (string, bool) {
	prefix := "#" + issueNum
	if !strings.HasPrefix(s, prefix) {
		return "", false
	}
	if len(s) > len(prefix) {
		next := s[len(prefix)]
		if next >= '0' && next <= '9' {
			return "", false
		}
	}
	rest := strings.TrimSpace(strings.TrimPrefix(s, prefix))
	rest = strings.TrimSpace(strings.TrimLeft(rest, ":"))
	return rest, true
}

// isWordByte reports whether b continues a word (letter or '-'), used to keep
// bookkeepingVerbs matching on whole tokens ("close" ≠ "close-off").
func isWordByte(b byte) bool {
	return b == '-' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// subjectAnchorRE is the single source for the "commit subject opens with #N"
// anchor: #N followed by end-of-subject or a non-digit, so #51 never matches
// #510 and a parenthetical "(see #51)" never anchors. RE2 has no negative
// lookahead, so "(?!\d)" is rendered "($|[^0-9])". allowClosePrefix permits a
// leading "close " — CommitWindow counts the close commit, the #76 ship probe
// does not.
func subjectAnchorRE(issueNum string, allowClosePrefix bool) *regexp.Regexp {
	prefix := ""
	if allowClosePrefix {
		prefix = `(close\s+)?`
	}
	return regexp.MustCompile(`^` + prefix + `#` + regexp.QuoteMeta(issueNum) + `($|[^0-9])`)
}

// ShippedWorkOnMain reports whether implementation work for issueNum has
// landed on the published main line — the close-off signal for `sdlc state`.
// Returns (firstWorkSHA, itsSubject, true) for the first subject-anchored
// non-bookkeeping commit reachable from MainRef, or ("","",false) when none
// exists or there is no main ref (degrades by construction — no network, never
// hard-fails). A merged PR lands the work-convention commits on main, so this
// gh-free scan *is* the merged-work signal.
//
// Mirrors CommitWindow's cheap `--grep #N` prefilter, then re-filters in Go
// with IsShippedWorkSubject (git's POSIX grep can't express the subject anchor
// + denylist reliably across platforms).
func ShippedWorkOnMain(issueNum string) (sha, subject string, shipped bool) {
	mainRef := MainRef()
	if mainRef == "" {
		return "", "", false
	}
	out, err := run("git", "log", mainRef, "--grep=#"+issueNum, "--pretty=%H%x00%s")
	if err != nil {
		return "", "", false
	}
	text := strings.TrimRight(string(out), "\n")
	if text == "" {
		return "", "", false
	}
	for _, line := range strings.Split(text, "\n") {
		parts := strings.SplitN(line, "\x00", 2)
		if len(parts) != 2 {
			continue
		}
		if IsShippedWorkSubject(issueNum, parts[1]) {
			return parts[0], parts[1], true
		}
	}
	return "", "", false
}

// WindowCapDays is the sanity cap on how far back the commit window can
// reach. Anything older is almost certainly a fork-upstream collision
// (the forked repo's history reusing #N for a different historical issue),
// not legitimate ancient work.
//
// 61 days (was 31, #68): long-running issues legitimately span more than a
// month — the project datatype is built around month-scale work (e.g. #16 ran
// ~34 days), and a 31-day cap truncated their windows, dropping the early work
// and contributing to empty/under-counted actuals.
const WindowCapDays = 61

// CommitWindow returns (firstSHA, firstISO, lastISO) for commits whose
// *subject* is owned by issueNum (`#N ...` or `<area>: #N ...`),
// capped at WindowCapDays in the past.
//
// firstISO is the *parent* of the first matching commit (the v3 segment-
// start trick: v3 segments span [parent_commit_time, this_commit_time], so
// using the parent lets the first segment extend backward and capture
// pre-commit work like typing and thinking). Falls back to the first
// match's own ISO if the parent lookup fails (initial-commit edge case) or
// the parent is outside the cap.
//
// Returns all-empty (no error) if no in-window subject anchor exists.
//
// Subject-anchored, not whole-message: forked-upstream history may contain
// commits referencing the same number in their *body* (e.g., "docs: setup
// snippet (issue: #123)" from a 2-year-old upstream commit) but not the
// subject. Whole-message --grep would pull those in and stretch the window
// by years.
func CommitWindow(issueNum string) (firstSHA, firstISO, lastISO string, err error) {
	// Loose --grep first to narrow candidates; precise subject-anchor
	// check happens below. Git's POSIX regex doesn't reliably support \b
	// for word boundaries across platforms, so we filter subjects in Go.
	cmd := exec.Command("git", "log",
		"--grep=#"+issueNum, "--reverse",
		"--pretty=%aI%x00%H%x00%s",
	)
	out, err := cmd.Output()
	if err != nil {
		// non-zero exit (e.g., not a git repo) → no window, no error
		// (matches close-issue.py's CalledProcessError swallow)
		return "", "", "", nil
	}
	text := strings.TrimRight(string(out), "\n")
	if text == "" {
		return "", "", "", nil
	}
	type match struct{ iso, sha string }
	var matches []match
	for _, line := range strings.Split(text, "\n") {
		parts := strings.SplitN(line, "\x00", 3)
		if len(parts) != 3 {
			continue
		}
		iso, sha, subject := parts[0], parts[1], parts[2]
		if _, ok := issueSubjectDescriptor(issueNum, subject, true); ok {
			matches = append(matches, match{iso, sha})
		}
	}
	if len(matches) == 0 {
		return "", "", "", nil
	}
	capISO := windowCapISO()
	var recent []match
	for _, m := range matches {
		if m.iso >= capISO {
			recent = append(recent, m)
		}
	}
	if len(recent) == 0 {
		return "", "", "", nil
	}
	firstSHA = recent[0].sha
	firstISO = recent[0].iso
	lastISO = recent[len(recent)-1].iso

	// v3 segment-start: parent of first match (still bounded by cap).
	parentOut, perr := exec.Command(
		"git", "log", "-1", "--pretty=%aI", firstSHA+"^",
	).Output()
	if perr == nil {
		parentISO := strings.TrimSpace(string(parentOut))
		if parentISO != "" && parentISO >= capISO {
			return firstSHA, parentISO, lastISO, nil
		}
	}
	return firstSHA, firstISO, lastISO, nil
}

// WorkingTransitionISO returns the author-date ISO timestamp of the EARLIEST
// commit that flipped issueFile to `status: working` — i.e. the claim /
// start-work commit. It anchors the active-time window at engagement start
// (#113), so DESIGN attention between the claim and the first `#N` code commit
// (brainstorm / spec / plan / reviews) lands in-window instead of being cut off
// at the parent-of-first-commit. Returns ("", false) when no such commit exists
// or the earliest one is older than WindowCapDays (a stale or fork-collision
// claim must not anchor an ancient window).
//
// Mechanism: `git log -G'^status: *working' --reverse -- <issueFile>` lists
// commits whose diff added/removed a `status: working` line; --reverse yields
// the oldest first — the open→working flip (a later working→X removal, or a
// reopen, can only come after it). `%aI` mirrors CommitWindow's author-date
// choice and format, so the WindowCapDays string compare is apples-to-apples.
//
// The -G (diff-grep) form is deliberate, not -S (pickaxe count): the flip
// changes `status: open` → `status: working`, so the added line matches; a
// later body edit that leaves status untouched carries no such line in its diff
// and is skipped. Uses the package `run` shim so tests can drive it.
func WorkingTransitionISO(issueFile string) (string, bool) {
	out, err := run("git", "log", "-G^status: *working", "--reverse",
		"--format=%aI", "--", issueFile)
	if err != nil {
		return "", false
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return "", false
	}
	iso := strings.SplitN(text, "\n", 2)[0]
	if iso < windowCapISO() {
		return "", false
	}
	return iso, true
}

// windowCapISO is the oldest ISO timestamp a window edge may take — now minus
// WindowCapDays. The single source for the cap shared by CommitWindow and
// WorkingTransitionISO (ARCH-DRY); both compare `%aI` author dates against it.
func windowCapISO() string {
	return time.Now().UTC().
		Add(-time.Duration(WindowCapDays) * 24 * time.Hour).
		Format("2006-01-02T15:04:05-07:00")
}

// DiscoverWindowIssues returns every distinct LOCAL issue number referenced in commit
// subjects within [since, until]. `primary` is always included, even if no commits match it.
//
// Why all of them: the active-time-v3 algorithm anchors segments by commit-subject issue ref;
// unrecognized refs fall into mention-fallback and inflate the closing issue's share by
// 3-10x.
//
// selfRepo is this repo's name — the qualifier that counts as LOCAL, so `ariadne#180` inside
// ariadne is kept while `pair#127` is not. It is a PARAMETER rather than a RepoTopLevel() call
// here for two reasons: RepoTopLevel shells out via exec.Command directly, bypassing the `run`
// shim, so a self-qualifier resolved internally would be untestable — the self-qualified case
// would ship with no guard at all; and the single caller already holds the repo root, so
// passing it makes the dependency explicit rather than implicit in the process cwd. "" means
// unknown, which keeps only bare refs (ariadne#190).
//
// This is the entry point of the #190 chain: whatever lands here becomes activetime's tracked
// issue set, and therefore its mention pattern.
func DiscoverWindowIssues(sinceISO, untilISO, primary, selfRepo string) ([]string, error) {
	// Via the package `run` shim, not exec.Command directly — the shim is the documented path
	// for callers in this package, and it is what makes the ref filtering testable.
	out, err := run("git", "log",
		"--since="+sinceISO, "--until="+untilISO, "--pretty=%s",
	)
	if err != nil {
		return []string{primary}, nil
	}
	text := strings.TrimRight(string(out), "\n")
	seen := map[string]struct{}{}
	for _, line := range strings.Split(text, "\n") {
		for _, num := range issueref.LocalNums(line, selfRepo) {
			seen[num] = struct{}{}
		}
	}
	if _, ok := seen[primary]; !ok {
		seen[primary] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ai, _ := strconv.Atoi(keys[i])
		aj, _ := strconv.Atoi(keys[j])
		return ai < aj
	})
	// close-issue.py: sorted set keyed by int, then primary appended if
	// not present. Our sort already gives the numerically-sorted set;
	// primary lands wherever its number sorts.
	return keys, nil
}

// DiffNames returns the list of file paths changed between sinceRef and
// untilRef (`git diff --name-only sinceRef untilRef`). Empty slice + nil
// error if there are no changes; non-nil error only on hard git failures.
func DiffNames(sinceRef, untilRef string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", sinceRef, untilRef)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --name-only %s %s: %w", sinceRef, untilRef, err)
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return nil, nil
	}
	return strings.Split(text, "\n"), nil
}

// FileChange is one entry of `git diff --name-status`: a single-letter status
// (A added, M modified, D deleted, R renamed, C copied) and the file's CURRENT
// path (for a rename, the destination path).
type FileChange struct {
	Status string
	Path   string
}

// DiffNameStatus returns the changed files between sinceRef and untilRef with their
// status code (`git diff --name-status`). untilRef "" compares against the working
// tree (mirrors collectDiff's omit-when-empty). Empty slice + nil on no changes;
// error only on hard git failures. Used by the #124 instance-conformance gate to tell
// newly-ADDED issue files (section-checked) from MODIFIED/renamed ones (grandfathered).
func DiffNameStatus(sinceRef, untilRef string) ([]FileChange, error) {
	args := []string{"diff", "--name-status", sinceRef}
	if untilRef != "" {
		args = append(args, untilRef)
	}
	out, err := run("git", args...) // via the shim so the A/M/R/D parser is unit-testable
	if err != nil {
		return nil, fmt.Errorf("git diff --name-status %s %s: %w", sinceRef, untilRef, err)
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return nil, nil
	}
	var changes []FileChange
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 2 || fields[0] == "" {
			continue
		}
		changes = append(changes, FileChange{
			Status: fields[0][:1],         // "R100" → "R", "A" → "A"
			Path:   fields[len(fields)-1], // rename: destination path is last
		})
	}
	return changes, nil
}

// LogEntry is one line of `git log --reverse --format=%H %ci %s`.
type LogEntry struct {
	SHA, Date, Subject string
}

// LogReverse returns the full commit log in reverse-chronological order
// (oldest first), one LogEntry per commit. Used by close-issue.py's
// "first commit referencing '#N M4'" scan.
func LogReverse() ([]LogEntry, error) {
	cmd := exec.Command("git", "log", "--reverse", "--format=%H %ci %s")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	text := strings.TrimRight(string(out), "\n")
	if text == "" {
		return nil, nil
	}
	var entries []LogEntry
	for _, line := range strings.Split(text, "\n") {
		// Format: "<sha> <YYYY-MM-DD HH:MM:SS ±zzzz> <subject>"
		// SHA is 40 hex chars; date is fixed 25 chars; subject is the rest.
		// We split on the first space (sha), then need to peel off the date.
		if len(line) < 41 {
			continue
		}
		sha := line[:40]
		rest := line[41:]
		// Date is exactly 25 chars: "2006-01-02 15:04:05 -0700"
		if len(rest) < 26 {
			continue
		}
		date := rest[:25]
		subject := rest[26:]
		entries = append(entries, LogEntry{sha, date, subject})
	}
	return entries, nil
}

// RepoTopLevel returns the path of the git repo root (`git rev-parse
// --show-toplevel`).
func RepoTopLevel() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ErrNoMatches is returned by helpers when nothing matched (so callers can
// distinguish "ran fine, no result" from "git failed").
var ErrNoMatches = errors.New("no matches")
