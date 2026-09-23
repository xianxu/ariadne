package acquire

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Transport translation happens after production source validation.
type fixtureRemoteGit struct {
	ExecGit
	remote string
}

func (g fixtureRemoteGit) RunOwned(ctx context.Context, dir, stage string, args ...string) (string, error) {
	a := append([]string(nil), args...)
	for i, v := range a {
		if v == "https://example.test/base.git" {
			a[i] = g.remote
		}
	}
	out, err := g.ExecGit.RunOwned(ctx, dir, stage, a...)
	if err == nil {
		_, err = g.ExecGit.Run(ctx, a[len(a)-1], "remote", "set-url", "origin", "https://example.test/base.git")
	}
	return out, err
}
func scopedClient(t *testing.T) (Client, string) {
	t.Helper()
	env := canonical(t.TempDir())
	host := filepath.Join(env, "host")
	put(t, filepath.Join(host, "construct/base.manifest"), "")
	return Client{Policy: &Policy{EnvironmentRoot: env, HostRoot: host, HostCommonDir: filepath.Join(host, ".git")}}, env
}
func TestPolicyConfinement(t *testing.T) {
	c, env := scopedClient(t)
	for _, p := range []string{env, filepath.Dir(env), filepath.Join(env, "host"), filepath.Join(env, "deep", "base")} {
		if err := c.Ensure(context.Background(), p, "https://example.test/base.git", true); err == nil {
			t.Fatalf("accepted %s", p)
		}
	}
	for _, source := range []string{"../base", "/tmp/base", "file:///tmp/base"} {
		if err := c.Ensure(context.Background(), filepath.Join(env, "base"), source, true); err == nil || !strings.Contains(err.Error(), "remote") {
			t.Fatalf("%s: %v", source, err)
		}
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(env, "alias")); err != nil {
		t.Fatal(err)
	}
	if err := c.Ensure(context.Background(), filepath.Join(env, "alias"), "", true); err == nil {
		t.Fatal("accepted outward alias")
	}
}
func TestPolicyMainAndWarmRealGit(t *testing.T) {
	for _, tagOnly := range []bool{false, true} {
		t.Run(map[bool]string{false: "branch", true: "tag-only"}[tagOnly], func(t *testing.T) {
			c, env := scopedClient(t)
			remote := origin(t, t.TempDir(), "base", "", true)
			if tagOnly {
				gitFixture(t, remote, "tag", "main")
			} else {
				gitFixture(t, remote, "branch", "main")
				gitFixture(t, remote, "symbolic-ref", "HEAD", "refs/heads/master")
			}
			c.Git = fixtureRemoteGit{remote: remote}
			dest := filepath.Join(env, "base")
			err := c.Ensure(context.Background(), dest, "https://example.test/base.git", true)
			if tagOnly {
				if err == nil {
					t.Fatal("published tag-only main")
				}
				if _, e := os.Stat(dest); !os.IsNotExist(e) {
					t.Fatalf("published failed checkout: %v", e)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := gitFixture(t, dest, "branch", "--show-current"); got != "main" {
				t.Fatal(got)
			}
			gitFixture(t, dest, "checkout", "-b", "chosen")
			put(t, filepath.Join(dest, "README"), "private")
			gitFixture(t, dest, "add", "README")
			gitFixture(t, dest, "commit", "-m", "unpublished")
			put(t, filepath.Join(dest, "README"), "dirty")
			put(t, filepath.Join(dest, "untracked"), "keep")
			before := gitFixture(t, dest, "show-ref")
			if err := os.Rename(remote, remote+"-offline"); err != nil {
				t.Fatal(err)
			}
			for _, source := range []string{"https://example.test/base.git", ""} {
				if err := c.Ensure(context.Background(), dest, source, true); err != nil {
					t.Fatal(err)
				}
			}
			if after := gitFixture(t, dest, "show-ref"); before != after {
				t.Fatal("changed refs")
			}
			if b, _ := os.ReadFile(filepath.Join(dest, "README")); string(b) != "dirty" {
				t.Fatal("changed dirty state")
			}
			if err := c.Ensure(context.Background(), dest, "https://example.test/other.git", true); err == nil {
				t.Fatal("accepted wrong origin")
			}
		})
	}
}
func TestPolicyOrdinaryOnlyAndDryRun(t *testing.T) {
	c, env := scopedClient(t)
	dest := filepath.Join(env, "base")
	put(t, filepath.Join(dest, "construct/base.manifest"), "")
	if err := c.Ensure(context.Background(), dest, "", true); err == nil {
		t.Fatal("accepted non-repository")
	}
	gitFixture(t, dest, "init")
	gitFixture(t, dest, "add", ".")
	gitFixture(t, dest, "commit", "-m", "init")
	if err := c.Ensure(context.Background(), dest, "", true); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(env, "linked")
	gitFixture(t, dest, "worktree", "add", "-b", "linked", linked)
	if err := c.Ensure(context.Background(), linked, "", true); err == nil {
		t.Fatal("accepted linked dependency")
	}
	for _, row := range []string{"substrate ../../escape https://example.test/base.git\n", "substrate ../missing ../local\n", "data https://example.test/host.git mount\n", "substrate ../base\ndata https://example.test/base.git mount\n", "data https://example.test/base.git mount\nsubstrate ../base\n"} {
		put(t, filepath.Join(c.Policy.HostRoot, "construct/deps"), row)
		if _, err := c.Restore(context.Background(), c.Policy.HostRoot, true); err == nil {
			t.Fatalf("accepted %s", row)
		}
	}
}

func TestPolicyPublicationConflictRetainsRecoveryEvidence(t *testing.T) {
	c, env := scopedClient(t)
	dest := filepath.Join(env, "base")
	c.Git = &stateGit{run: func(_ int, dir string, args []string) (string, error) {
		if args[0] == "clone" {
			put(t, filepath.Join(args[len(args)-1], "construct/base.manifest"), "")
			put(t, filepath.Join(dest, "foreign"), "keep")
			return "", nil
		}
		return "same-sha", nil
	}}
	if err := c.Ensure(context.Background(), dest, "https://example.test/base.git", true); err == nil {
		t.Fatal("expected destination conflict")
	}
	stages, _ := filepath.Glob(filepath.Join(env, ".base-weave-*"))
	if len(stages) != 1 {
		t.Fatalf("missing recovery evidence: %v", stages)
	}
	if b, _ := os.ReadFile(filepath.Join(dest, "foreign")); string(b) != "keep" {
		t.Fatal("foreign destination changed")
	}
}

func TestPolicyTransitiveEscapeStopsBeforeClone(t *testing.T) {
	c, env := scopedClient(t)
	source := "https://example.test/base.git"
	g := &remoteModel{t: t, remotes: map[string]fakeRemote{source: {main: "main", head: "default", deps: "substrate ../../escape https://example.test/escape.git\n"}}}
	c.Git = g
	put(t, filepath.Join(c.Policy.HostRoot, "construct/deps"), "substrate ../base "+source+"\n")
	if _, err := c.Restore(context.Background(), c.Policy.HostRoot, false); err == nil {
		t.Fatal("accepted transitive escape")
	}
	if g.clones != 1 {
		t.Fatalf("attempted escaping clone: %d", g.clones)
	}
	if _, err := os.Stat(filepath.Join(env, "base/construct/base.manifest")); err != nil {
		t.Fatal("lost completed clone")
	}
}
