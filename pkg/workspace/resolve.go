package workspace

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Resolve observes an existing workspace. Its result is a snapshot, never a
// reservation or authority for a later mutation. Ref occupancy is observed once
// with topology; subsequent operations must revalidate under their own lock.
// Selected paths, HEAD/branch and resting commit are checked again to reject
// conflicting evidence, without claiming an atomic snapshot of concurrent Git.
func Resolve(git GitReader, dir, address string) (Identity, error) {
	v, trees, err := loadVantage(git, dir)
	if err != nil {
		return Identity{}, err
	}
	if address != "" {
		a, e := ParseAddress(address)
		if e != nil {
			return Identity{}, e
		}
		if a.Repo == "" {
			a.Repo = filepath.Base(v.PrimaryRoot)
		}
		if a.Repo != filepath.Base(v.PrimaryRoot) {
			peer := filepath.Join(v.FleetRoot, a.Repo)
			pv, pt, e := loadVantage(git, peer)
			if e != nil {
				return Identity{}, fmt.Errorf("resolve repository %q: %w", a.Repo, e)
			}
			if pv.PrimaryRoot != peer || pv.FleetRoot != v.FleetRoot {
				return Identity{}, fmt.Errorf("repository %q is not an exact fleet primary", a.Repo)
			}
			v, trees = pv, pt
		}
		v.WorktreeRoot = v.PrimaryRoot
		if a.Slot > 0 {
			expected, e := SlotPath(v.FleetRoot, a.Repo, a.Slot)
			if e != nil {
				return Identity{}, e
			}
			actual, e := CanonicalPath(expected)
			if e != nil {
				return Identity{}, e
			}
			if actual != expected {
				return Identity{}, fmt.Errorf("canonical slot path %q redirects to %q", expected, actual)
			}
			v.WorktreeRoot = actual
		}
	}
	// Missing unrelated prunable worktrees remain ordinary fleet topology. Their
	// original path is retained for ref occupancy, but cannot select a workspace.
	commandDir, e := CanonicalPath(dir)
	if e != nil {
		return Identity{}, e
	}
	for i := range trees {
		p, e := CanonicalListedWorktreePath(commandDir, trees[i].Path)
		if e == nil {
			trees[i].Path = p
		}
	}
	refs := map[string]string{}
	n := slotNumber(v)
	if n > 0 {
		resting := fmt.Sprintf("main-slot%d", n)
		out, e := git.GitInDir(v.WorktreeRoot, "rev-parse", "--verify", "--end-of-options", "refs/heads/"+resting+"^{commit}")
		if e != nil {
			return Identity{}, gitPathError(v.WorktreeRoot, "resolve resting commit", e, out)
		}
		oid := strings.TrimSuffix(string(out), "\n")
		if !validOID(oid) {
			return Identity{}, fmt.Errorf("malformed resting commit OID")
		}
		refs[resting] = oid
	}
	id, e := Classify(v, trees, refs)
	if e != nil {
		return Identity{}, e
	}
	if id.Head != nil && !validOID(*id.Head) {
		return Identity{}, fmt.Errorf("malformed worktree HEAD OID")
	}
	if e := recheck(git, v, id); e != nil {
		return Identity{}, e
	}
	if n > 0 {
		resting := *id.RestingBranch
		out, e := git.GitInDir(v.WorktreeRoot, "rev-parse", "--verify", "--end-of-options", "refs/heads/"+resting+"^{commit}")
		if e != nil {
			return Identity{}, gitPathError(v.WorktreeRoot, "recheck resting commit", e, out)
		}
		if strings.TrimSuffix(string(out), "\n") != refs[resting] {
			return Identity{}, fmt.Errorf("resting ref changed during workspace resolution")
		}
	}
	return id, nil
}
func validOID(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func recheck(git GitReader, v Vantage, id Identity) error {
	for _, q := range []struct{ arg, want string }{{"--show-toplevel", v.WorktreeRoot}, {"--git-common-dir", v.RepoIdentity}} {
		out, e := git.GitInDir(v.WorktreeRoot, "rev-parse", q.arg)
		if e != nil {
			return gitPathError(v.WorktreeRoot, q.arg, e, out)
		}
		got, e := CanonicalGitOutputPath(v.WorktreeRoot, out)
		if e != nil {
			return e
		}
		if got != q.want {
			return fmt.Errorf("workspace changed during resolution: %s", q.arg)
		}
	}
	if id.Head != nil {
		out, e := git.GitInDir(v.WorktreeRoot, "rev-parse", "--verify", "--end-of-options", "HEAD^{commit}")
		if e != nil {
			return gitPathError(v.WorktreeRoot, "resolve HEAD", e, out)
		}
		if strings.TrimSuffix(string(out), "\n") != *id.Head {
			return fmt.Errorf("HEAD changed during workspace resolution")
		}
	} else {
		if id.Branch == nil {
			return fmt.Errorf("unborn HEAD has no branch")
		}
		out, e := git.GitInDir(v.WorktreeRoot, "for-each-ref", "--format=%(refname)", "refs/heads/"+*id.Branch)
		if e != nil {
			return gitPathError(v.WorktreeRoot, "recheck unborn branch", e, out)
		}
		for _, ref := range strings.Split(strings.TrimSuffix(string(out), "\n"), "\n") {
			if ref == "refs/heads/"+*id.Branch {
				return fmt.Errorf("unborn HEAD changed during workspace resolution")
			}
		}
	}
	out, e := git.GitInDir(v.WorktreeRoot, "symbolic-ref", "-q", "HEAD")
	if id.Branch != nil {
		if e != nil {
			return gitPathError(v.WorktreeRoot, "symbolic-ref HEAD", e, out)
		}
		if strings.TrimSuffix(string(out), "\n") != "refs/heads/"+*id.Branch {
			return fmt.Errorf("branch changed during workspace resolution")
		}
	} else if e == nil {
		return fmt.Errorf("detached HEAD changed during workspace resolution")
	} else {
		var exit interface{ ExitCode() int }
		if !errors.As(e, &exit) || exit.ExitCode() != 1 {
			return gitPathError(v.WorktreeRoot, "symbolic-ref HEAD", e, out)
		}
	}
	return nil
}
