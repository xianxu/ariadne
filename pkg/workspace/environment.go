package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// EnvironmentHost is a Git-proved numbered host, independent of dependency identity.
type EnvironmentHost struct {
	Repo         string `json:"repo"`
	Slot         int    `json:"slot"`
	RepoIdentity string `json:"repo_identity"`
	PrimaryRoot  string `json:"primary_root"`
	WorktreeRoot string `json:"worktree_root"`
}
type Environment struct {
	Root string
	Host EnvironmentHost
}

// environmentCandidate observes only the spelling of a direct child. It conveys
// no Git authority. A malformed numbered spelling must not downgrade to generic.
func environmentCandidate(root string) (*Environment, error) {
	env := filepath.Dir(root)
	if filepath.Base(filepath.Dir(env)) != "worktree" {
		return nil, nil
	}
	name := filepath.Base(env)
	i := strings.LastIndex(name, "-slot")
	if i < 0 {
		return nil, nil
	}
	repo, num := name[:i], name[i+5:]
	n, e := strconv.Atoi(num)
	if e != nil || n < 1 || strconv.Itoa(n) != num || !validRepoName(repo) {
		return nil, fmt.Errorf("invalid numbered environment %q", env)
	}
	fleet := filepath.Dir(filepath.Dir(env))
	host, _ := SlotPath(fleet, repo, n)
	return &Environment{Root: env, Host: EnvironmentHost{Repo: repo, Slot: n, PrimaryRoot: filepath.Join(fleet, repo), WorktreeRoot: host}}, nil
}

// DiscoverEnvironment is optional for archive/fixture callers: non-Git inputs
// outside numbered paths need no Git executable. Linked worktrees are followed
// through their verified primary, including those placed outside the environment.
func DiscoverEnvironment(git GitReader, dir string) (*Environment, error) {
	abs, e := filepath.Abs(dir)
	if e != nil {
		return nil, e
	}
	candidate := false
	for p := abs; ; p = filepath.Dir(p) {
		c, err := environmentCandidate(p)
		if err != nil {
			return nil, err
		}
		if c != nil {
			candidate = true
			break
		}
		if _, err := os.Lstat(filepath.Join(p, ".git")); err == nil {
			candidate = true
			break
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	if !candidate {
		return nil, nil
	}
	v, _, e := loadVantage(git, dir)
	if e != nil {
		return nil, e
	}
	if v.EnvironmentHost == nil {
		return nil, nil
	}
	return &Environment{Root: v.EnvironmentRoot, Host: *v.EnvironmentHost}, nil
}

func environmentForVantage(git GitReader, v Vantage) (*Environment, error) {
	c, e := environmentCandidate(v.WorktreeRoot)
	if e != nil {
		return nil, e
	}
	selected := v
	if c == nil {
		c, e = environmentCandidate(v.PrimaryRoot)
		if e != nil || c == nil {
			return c, e
		}
		var primaryTrees []Worktree
		selected, primaryTrees, e = loadTopology(git, v.PrimaryRoot)
		if e != nil {
			return nil, e
		}
		if selected.WorktreeRoot != v.PrimaryRoot || selected.PrimaryRoot != v.PrimaryRoot || selected.RepoIdentity != v.RepoIdentity {
			return nil, fmt.Errorf("feature worktree primary identity conflicts with environment clone")
		}
		if err := exactRegistration(primaryTrees, v.WorktreeRoot); err != nil {
			return nil, err
		}
	}
	host, trees, e := loadTopology(git, c.Host.WorktreeRoot)
	if e != nil {
		return nil, fmt.Errorf("verify environment host %q: %w", c.Host.WorktreeRoot, e)
	}
	if host.WorktreeRoot != c.Host.WorktreeRoot || host.PrimaryRoot != c.Host.PrimaryRoot || host.PrimaryRoot == host.WorktreeRoot {
		return nil, fmt.Errorf("environment host is not the exact canonical linked worktree")
	}
	// Probe the actual primary: the host's own porcelain is insufficient authority.
	primary, primaryTrees, e := loadTopology(git, c.Host.PrimaryRoot)
	if e != nil {
		return nil, e
	}
	if primary.WorktreeRoot != c.Host.PrimaryRoot || primary.PrimaryRoot != c.Host.PrimaryRoot || primary.RepoIdentity != host.RepoIdentity {
		return nil, fmt.Errorf("environment host primary/common directory conflict")
	}
	if err := exactRegistration(primaryTrees, host.WorktreeRoot); err != nil {
		return nil, err
	}
	if err := exactRegistration(trees, host.WorktreeRoot); err != nil {
		return nil, err
	}
	if selected.WorktreeRoot == host.WorktreeRoot {
		if selected.RepoIdentity != host.RepoIdentity || selected.PrimaryRoot != host.PrimaryRoot {
			return nil, fmt.Errorf("conflicting environment host identity")
		}
	} else {
		if selected.WorktreeRoot != selected.PrimaryRoot || selected.RepoIdentity == host.RepoIdentity || selected.RepoIdentity != filepath.Join(selected.PrimaryRoot, ".git") {
			return nil, fmt.Errorf("environment dependency must be an independent ordinary clone")
		}
		physical, e := CanonicalPath(filepath.Join(c.Root, filepath.Base(selected.PrimaryRoot)))
		if e != nil || physical != selected.PrimaryRoot {
			return nil, fmt.Errorf("redirected environment dependency")
		}
		info, e := os.Lstat(filepath.Join(selected.PrimaryRoot, ".git"))
		if e != nil || !info.IsDir() {
			return nil, fmt.Errorf("environment dependency must have its own ordinary .git directory")
		}
	}
	// Canonical expected paths reject symlinked host/env/primary authority.
	for _, p := range []string{c.Root, c.Host.WorktreeRoot, c.Host.PrimaryRoot} {
		actual, e := CanonicalPath(p)
		if e != nil || actual != p {
			return nil, fmt.Errorf("redirected environment path %q", p)
		}
	}
	for _, q := range []struct{ arg, want string }{{"--show-toplevel", host.WorktreeRoot}, {"--git-common-dir", host.RepoIdentity}} {
		out, e := git.GitInDir(host.WorktreeRoot, "rev-parse", q.arg)
		if e != nil {
			return nil, e
		}
		actual, e := CanonicalGitOutputPath(host.WorktreeRoot, out)
		if e != nil || actual != q.want {
			return nil, fmt.Errorf("environment host changed during discovery")
		}
	}
	c.Host.RepoIdentity = host.RepoIdentity
	return c, nil
}

func exactRegistration(trees []Worktree, path string) error {
	count := 0
	for _, w := range trees {
		p, e := CanonicalPath(w.Path)
		if e == nil && p == path {
			if w.Path != path || w.Bare {
				return fmt.Errorf("redirected worktree registration %q", path)
			}
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("missing or ambiguous primary registration of %q", path)
	}
	return nil
}
