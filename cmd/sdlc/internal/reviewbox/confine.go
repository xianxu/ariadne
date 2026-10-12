package reviewbox

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Confinement is the environment every dispatched reviewer runs in (#339), with
// or without an OS write fence: a temp dir of its own, git that can neither
// discover a repository above its root nor create one outside its directories,
// and a check afterward that nothing it shouldn't touch changed. A reviewer
// that inherited the agent session's TMPDIR once had mktemp fail, landed in
// $HOME, and ran `git init` and `git add -A` there.
//
// Lifecycle (ARCH-FUNERAL): judge.Dispatch makes it before the first attempt
// and removes it after the last. Inside a review checkout the base also dies
// with the checkout; outside one it is a small directory in the OS temp dir.
type Confinement struct {
	Root    string // the directory the reviewer owns: the review checkout or the repository it reviews
	Base    string // tmp/ (TMPDIR) and bin/ (the git shim)
	inside  bool
	realGit string
	mktemp  string // the real mktemp, shimmed on macOS; empty when not shimmed
	before  OutsideSnapshot
}

const confineBaseName = ".sdlc-review"

// Confine prepares a reviewer's confinement around root. With inside, root is
// a disposable review checkout and the base lives in it; otherwise root is
// resolved to its repository's top level and the base is a fresh directory
// under the OS temp dir. Any failure to create or write the base is an error:
// a reviewer never falls back to a directory outside.
func Confine(root string, inside bool) (*Confinement, error) {
	r, err := canonicalDir(root)
	if err != nil {
		return nil, fmt.Errorf("resolve the reviewer's root: %w", err)
	}
	if !inside {
		if top, err := gitOutput(r, "rev-parse", "--show-toplevel"); err == nil {
			if t, err := canonicalDir(top); err == nil {
				r = t
			}
		}
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("locate git for the reviewer's shim: %w", err)
	}
	if realGit, err = filepath.Abs(realGit); err != nil {
		return nil, err
	}
	base := filepath.Join(r, confineBaseName)
	if !inside {
		if base, err = os.MkdirTemp("", "sdlc-review-"); err != nil {
			return nil, fmt.Errorf("create the reviewer's temp dir: %w", err)
		}
		if base, err = canonicalDir(base); err != nil {
			return nil, fmt.Errorf("create the reviewer's temp dir: %w", err)
		}
	}
	c := &Confinement{Root: r, Base: base, inside: inside, realGit: realGit}
	if runtime.GOOS == "darwin" {
		if m, err := exec.LookPath("mktemp"); err == nil {
			c.mktemp, _ = filepath.Abs(m)
		}
	}
	if err := c.Prepare(); err != nil {
		c.Close()
		return nil, err
	}
	if c.before, err = TakeSnapshot(c.watched(), userHome()); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// Tmp is the reviewer's TMPDIR.
func (c *Confinement) Tmp() string { return filepath.Join(c.Base, "tmp") }

// BinDir holds the git shim; it goes first on the reviewer's PATH.
func (c *Confinement) BinDir() string { return filepath.Join(c.Base, "bin") }

// Prepare (re)creates the base: the temp dir, proven writable, and the shim.
// Idempotent; a retried review calls it again after the checkout was reset.
func (c *Confinement) Prepare() error {
	for _, d := range []string{c.Tmp(), c.BinDir()} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			return fmt.Errorf("create the reviewer's temp dir: %w", err)
		}
	}
	probe := filepath.Join(c.Tmp(), ".sdlc-probe")
	if err := os.WriteFile(probe, nil, 0o600); err != nil {
		return fmt.Errorf("the reviewer's temp dir %s is not writable: %w", c.Tmp(), err)
	}
	if err := os.Remove(probe); err != nil {
		return fmt.Errorf("the reviewer's temp dir %s is not writable: %w", c.Tmp(), err)
	}
	if err := os.WriteFile(filepath.Join(c.BinDir(), "git"), []byte(GitShim(c.realGit, []string{c.Root, c.Base})), 0o700); err != nil {
		return fmt.Errorf("write the reviewer's git shim: %w", err)
	}
	if c.mktemp != "" {
		if err := os.WriteFile(filepath.Join(c.BinDir(), "mktemp"), []byte(MktempShim(c.mktemp)), 0o700); err != nil {
			return fmt.Errorf("write the reviewer's mktemp shim: %w", err)
		}
	}
	if err := os.WriteFile(c.envFile(), []byte(claudeEnvFile(c.BinDir(), os.Getenv("CLAUDE_ENV_FILE"), c.Env())), 0o600); err != nil {
		return fmt.Errorf("write the reviewer's env file: %w", err)
	}
	if c.inside {
		return excludeFromCheckout(c.Root, "/"+confineBaseName+"/")
	}
	return nil
}

func (c *Confinement) envFile() string { return filepath.Join(c.Base, "env.sh") }

// claudeEnvFile is sourced by Claude Code before each Bash command
// (CLAUDE_ENV_FILE). Its Bash rebuilds PATH from the user's shell snapshot, so
// a PATH set in the process environment never reaches the reviewer's commands;
// measured on #339, the shims took effect only through this file. An env file
// the reviewer inherited is sourced first, then the confinement's own
// variables are exported again so it can't override them.
func claudeEnvFile(bin, inherited string, env []string) string {
	var b strings.Builder
	if inherited != "" {
		fmt.Fprintf(&b, "[ -f %[1]s ] && . %[1]s\n", shQuote(inherited))
	}
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		if k != "CLAUDE_ENV_FILE" {
			fmt.Fprintf(&b, "%s=%s; export %s\n", k, shQuote(v), k)
		}
	}
	fmt.Fprintf(&b, "PATH=%s:\"$PATH\"; export PATH\n", shQuote(bin))
	return b.String()
}

// Env is what the reviewer's environment overrides (PATH is prepended with
// BinDir by the dispatcher, and again by the env file for Claude Code). Claude Code's Bash takes TMPDIR from
// CLAUDE_CODE_TMPDIR, so both point into the base. The ceiling stops git's
// repository discovery above the root, and above the temp dir, so a scratch
// directory in TMPDIR is not mistaken for part of the checkout.
func (c *Confinement) Env() []string {
	tmp := c.Tmp()
	return []string{
		"TMPDIR=" + tmp, "TMP=" + tmp, "TEMP=" + tmp, "CLAUDE_CODE_TMPDIR=" + tmp,
		"GIT_CEILING_DIRECTORIES=" + filepath.Dir(c.Root) + string(os.PathListSeparator) + c.Base,
		"CLAUDE_ENV_FILE=" + c.envFile(),
	}
}

// Check reports, as one error, everything outside the reviewer's directories
// that changed since Confine: a new repository above its roots or a new entry
// in the home directory.
func (c *Confinement) Check() error {
	after, err := TakeSnapshot(c.watched(), userHome())
	if err != nil {
		return err
	}
	diff := Diff(c.before, after)
	if len(diff) == 0 {
		return nil
	}
	return fmt.Errorf("the reviewer changed paths outside its directories (inspect and remove them; if another process made them, rerun): %s", strings.Join(diff, "; "))
}

// Close removes the base (never the checkout itself). Idempotent.
func (c *Confinement) Close() error {
	if c == nil || c.Base == "" {
		return nil
	}
	return removeTree(c.Base)
}

// watched are the directories whose `.git` the check watches: every ancestor
// of the root and of the base, and the home directory.
func (c *Confinement) watched() []string {
	seen := map[string]bool{}
	var out []string
	for _, start := range []string{c.Root, c.Base} {
		for d := filepath.Dir(start); ; d = filepath.Dir(d) {
			if !seen[d] {
				seen[d] = true
				out = append(out, d)
			}
			if d == filepath.Dir(d) {
				break
			}
		}
	}
	if h := userHome(); h != "" && !seen[h] {
		out = append(out, h)
	}
	return out
}

// OutsideSnapshot is what the outside-write check compares: whether each
// watched directory holds a `.git`, and the home directory's non-hidden
// top-level entries. Hidden home entries are left out because tools touch
// dotfiles routinely; a new `~/scripts` or `~/.git` is this incident's mark.
type OutsideSnapshot struct {
	Repos map[string]bool // watched dir → <dir>/.git exists
	Home  string
	Names map[string]bool // non-hidden entries of Home
}

// TakeSnapshot reads the snapshot. A home directory it can't name or list is
// an error: a check that can't see home must not pass silently.
func TakeSnapshot(dirs []string, home string) (OutsideSnapshot, error) {
	s := OutsideSnapshot{Repos: map[string]bool{}, Home: home, Names: map[string]bool{}}
	for _, d := range dirs {
		_, err := os.Lstat(filepath.Join(d, ".git"))
		s.Repos[d] = err == nil
	}
	if home == "" {
		return s, errors.New("the outside-write check cannot locate the home directory")
	}
	entries, err := os.ReadDir(home)
	if err != nil {
		return s, fmt.Errorf("the outside-write check cannot list the home directory: %w", err)
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			s.Names[e.Name()] = true
		}
	}
	return s, nil
}

// Diff lists what appeared between two snapshots, sorted. Pure.
func Diff(before, after OutsideSnapshot) []string {
	var out []string
	for d, has := range after.Repos {
		if has && !before.Repos[d] {
			out = append(out, "new repository "+filepath.Join(d, ".git"))
		}
	}
	for n := range after.Names {
		if !before.Names[n] {
			out = append(out, "new entry "+filepath.Join(after.Home, n))
		}
	}
	sort.Strings(out)
	return out
}

// GitShim renders the `git` the reviewer's PATH finds first. It refuses `git
// init` and `git clone` whose working directory, git dir or work tree (option
// or GIT_DIR/GIT_WORK_TREE), or any path operand, relative or absolute, is
// outside allowed (a clone's existing source is read, not written). Every
// other command execs the real git unchanged. It stops an accidental
// repository in $HOME, not a reviewer that calls the real git by its path; the
// post-review check covers that.
func GitShim(realGit string, allowed []string) string {
	var pats []string
	for _, a := range allowed {
		pats = append(pats, shQuote(a), shQuote(a)+"/*")
	}
	return fmt.Sprintf(gitShimTemplate, shQuote(realGit), strings.Join(pats, "|"))
}

const gitShimTemplate = `#!/bin/sh
# sdlc reviewer confinement (#339): git may not create a repository outside
# the reviewer's own directories. Every other command runs unchanged.
real=%s
inside() { case "$1" in %s) return 0 ;; esac; return 1; }
# resolve prints the physical path $1 names (relative to $dir), or nothing when
# it climbs (..) out of its deepest existing directory.
resolve() {
	case "$1" in /*) d=$1 ;; *) d=$dir/$1 ;; esac
	rest=
	while [ ! -d "$d" ]; do rest=$(basename "$d")/$rest; d=$(dirname "$d"); done
	case "/$rest" in */../*) return 0 ;; esac
	(cd "$d" && pwd -P)
}
refuse() { echo "git $sub refused: $1 is outside the reviewer's directories (sdlc #339)" >&2; exit 128; }
check() { [ -z "$1" ] || inside "$(resolve "$1")" || refuse "$1"; }
dir=$(pwd -P); sub=; next=; gitdir=; worktree=
for a in "$@"; do
	if [ -n "$sub" ]; then
		[ "$next" = skip ] && { next=; continue; }
		case "$a" in
		--template) next=skip; continue ;;
		--template=*) continue ;;
		--*=*) p=${a#*=} ;;
		-*) continue ;;
		*) p=$a ;;
		esac
		[ "$sub" = clone ] && [ -e "$p" ] && continue # a source is read, not written
		check "$p"
		continue
	fi
	case "$next" in
	C) next=; dir=$(cd "$dir" 2>/dev/null && cd "$a" 2>/dev/null && pwd -P) || exec "$real" "$@"; continue ;;
	gitdir) next=; gitdir=$a; continue ;;
	worktree) next=; worktree=$a; continue ;;
	skip) next=; continue ;;
	esac
	case "$a" in
	-C) next=C ;;
	--git-dir) next=gitdir ;;
	--git-dir=*) gitdir=${a#*=} ;;
	--work-tree) next=worktree ;;
	--work-tree=*) worktree=${a#*=} ;;
	-c|--namespace|--config-env) next=skip ;;
	-*) ;;
	init|clone) sub=$a; check "$dir"; check "$gitdir"; check "$worktree"; check "$GIT_DIR"; check "$GIT_WORK_TREE" ;;
	*) exec "$real" "$@" ;;
	esac
done
exec "$real" "$@"
`

// MktempShim renders the macOS `mktemp` the reviewer's PATH finds first.
// Without a template, macOS mktemp writes under _CS_DARWIN_USER_TEMP_DIR
// (/var/folders/…/T) and ignores TMPDIR; an agent sandbox denies that
// directory, which is how the #339 reviewer's mktemp failed. The shim names a
// template in TMPDIR (or -p's directory) instead. Calls with a template run
// unchanged.
func MktempShim(realMktemp string) string {
	return fmt.Sprintf(mktempShimTemplate, shQuote(realMktemp))
}

const mktempShimTemplate = `#!/bin/sh
# sdlc reviewer confinement (#339): temp files go in the reviewer's TMPDIR.
real=%s
flags=; prefix=tmp; dir=${TMPDIR:-/tmp}; next=
for a in "$@"; do
	case "$next" in
	t) prefix=$a; next=; continue ;;
	p) dir=$a; next=; continue ;;
	esac
	case "$a" in
	-t) next=t ;;
	-p) next=p ;;
	--tmpdir=*) dir=${a#*=} ;;
	-[dqu]|-[dqu][dqu]|-[dqu][dqu][dqu]) flags="$flags $a" ;;
	*) exec "$real" "$@" ;;
	esac
done
exec "$real" $flags "$dir/$prefix.XXXXXXXXXX"
`

func shQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

func canonicalDir(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	c, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(c); err != nil {
		return "", err
	} else if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", c)
	}
	return c, nil
}

func userHome() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	if c, err := filepath.EvalSymlinks(h); err == nil {
		return c
	}
	return h
}

// excludeFromCheckout keeps the base out of the checkout's `git status`.
func excludeFromCheckout(dir, pattern string) error {
	path := filepath.Join(dir, ".git", "info", "exclude")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("exclude the reviewer's temp dir: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line == pattern {
			return nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("exclude the reviewer's temp dir: %w", err)
	}
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		data = append(data, '\n')
	}
	if err := os.WriteFile(path, append(data, pattern+"\n"...), 0o644); err != nil {
		return fmt.Errorf("exclude the reviewer's temp dir: %w", err)
	}
	return nil
}
