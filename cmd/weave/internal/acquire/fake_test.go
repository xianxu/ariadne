package acquire

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRemote struct {
	main, head, deps string
	tagOnly, noLayer bool
}

// checkout state is stored in the checkout itself so real staging rename and
// cleanup exercise the same publication boundary as the real Git fixture.
type remoteModel struct {
	t       *testing.T
	remotes map[string]fakeRemote
	clones  int
}

func (g *remoteModel) RunOwned(ctx context.Context, dir, stage string, args ...string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	g.clones++
	src, dest := args[len(args)-2], args[len(args)-1]
	r, ok := g.remotes[src]
	if !ok {
		return "", fmt.Errorf("remote unavailable")
	}
	selected := r.head
	if len(args) > 2 && args[1] == "--branch" {
		selected = r.main
	}
	if selected == "" {
		return "", fmt.Errorf("missing branch main")
	}
	put(g.t, filepath.Join(dest, ".git/origin"), src)
	put(g.t, filepath.Join(dest, ".git/head"), selected)
	if !r.tagOnly {
		put(g.t, filepath.Join(dest, ".git/main"), r.main)
	}
	if !r.noLayer {
		put(g.t, filepath.Join(dest, "construct/base.manifest"), "")
	}
	if r.deps != "" {
		put(g.t, filepath.Join(dest, "construct/deps"), r.deps)
	}
	return "", nil
}
func (g *remoteModel) Run(ctx context.Context, dir string, args ...string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		return "", err
	}
	key := strings.Join(args, " ")
	switch key {
	case "rev-parse --show-toplevel":
		return dir, nil
	case "rev-parse --path-format=absolute --git-common-dir", "rev-parse --absolute-git-dir":
		return filepath.Join(dir, ".git"), nil
	case "config --local --list":
		return "", nil
	}
	files := map[string]string{"config --local --get remote.origin.url": "origin", "rev-parse --verify HEAD^{commit}": "head", "rev-parse --verify refs/remotes/origin/main^{commit}": "main"}
	if file, ok := files[key]; ok {
		b, err := os.ReadFile(filepath.Join(dir, ".git", file))
		return string(b), err
	}
	return "", fmt.Errorf("unsupported git operation: %s", key)
}
func TestPolicyFakeFailuresAndRetry(t *testing.T) {
	for _, failure := range []string{"main", "tag", "manifest", "later", "cancel"} {
		t.Run(failure, func(t *testing.T) {
			c, env := scopedClient(t)
			source := "https://example.test/base.git"
			r := fakeRemote{main: "main-sha", head: "default-sha"}
			switch failure {
			case "main":
				r.main = ""
			case "tag":
				r.tagOnly = true
			case "manifest":
				r.noLayer = true
			case "later":
				r.deps = "substrate ../later https://example.test/later.git\n"
			}
			model := &remoteModel{t: t, remotes: map[string]fakeRemote{source: r}}
			c.Git = model
			put(t, filepath.Join(c.Policy.HostRoot, "construct/deps"), "substrate ../base "+source+"\n")
			ctx := context.Background()
			if failure == "cancel" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			if _, err := c.Restore(ctx, c.Policy.HostRoot, false); err == nil {
				t.Fatal("expected failure")
			}
			dest := filepath.Join(env, "base")
			_, err := os.Stat(dest)
			if failure == "later" {
				if err != nil {
					t.Fatal("lost completed dependency")
				}
				put(t, filepath.Join(dest, ".git/head"), "chosen-sha")
				put(t, filepath.Join(dest, "private"), "keep")
			} else if !os.IsNotExist(err) {
				t.Fatal("published incomplete dependency")
			}
			model.remotes[source] = fakeRemote{main: "new-main-sha", head: "default-sha"}
			model.remotes["https://example.test/later.git"] = fakeRemote{main: "later-sha", head: "default-sha"}
			if _, err := c.Restore(context.Background(), c.Policy.HostRoot, false); err != nil {
				t.Fatal(err)
			}
			if failure == "later" {
				b, _ := os.ReadFile(filepath.Join(dest, ".git/head"))
				if string(b) != "chosen-sha" {
					t.Fatal("retry reset chosen revision")
				}
			}
			matches, _ := filepath.Glob(filepath.Join(env, ".base-weave-*"))
			if len(matches) != 0 {
				t.Fatal(matches)
			}
		})
	}
}
func TestPolicyGeneratedPathsAndTransports(t *testing.T) {
	p := Policy{EnvironmentRoot: "/env", HostRoot: "/env/host"}
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("repo%d", i)
		good := "/env/" + name
		for _, prefix := range []string{"/", "/env/deeper/", "/elsewhere/"} {
			if err := p.validate(prefix+name, good, Source{}); err == nil {
				t.Fatal("lexical path bypass")
			}
		}
		if err := p.validate(good, good, Source{Identity: "uri:https://host/" + name}); err != nil {
			t.Fatal(err)
		}
		if err := p.validate(good, good, Source{Identity: "file:" + name}); err == nil {
			t.Fatal("local source accepted")
		}
	}
}

func TestPolicyGraphScale(t *testing.T) {
	for _, n := range []int{1, 10, 100} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			c, _ := scopedClient(t)
			model := &remoteModel{t: t, remotes: map[string]fakeRemote{}}
			c.Git = model
			for i := 0; i < n; i++ {
				deps := ""
				if i+1 < n {
					deps = fmt.Sprintf("substrate ../repo%d https://example.test/repo%d.git\n", i+1, i+1)
				}
				model.remotes[fmt.Sprintf("https://example.test/repo%d.git", i)] = fakeRemote{main: "main", head: "default", deps: deps}
			}
			put(t, filepath.Join(c.Policy.HostRoot, "construct/deps"), "substrate ../repo0 https://example.test/repo0.git\n")
			result, err := c.Restore(context.Background(), c.Policy.HostRoot, false)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Layers) != n+1 || model.clones != n {
				t.Fatalf("layers=%d clones=%d expected %d dependencies plus host", len(result.Layers), model.clones, n)
			}
			if _, err := c.Restore(context.Background(), c.Policy.HostRoot, false); err != nil {
				t.Fatal(err)
			}
			if model.clones != n {
				t.Fatal("warm graph cloned sources")
			}
		})
	}
}
