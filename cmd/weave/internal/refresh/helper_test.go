package refresh

import (
	"context"
	"errors"
	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelperRefreshEligibility(t *testing.T) {
	for _, c := range []struct {
		branch, head, target string
		ancestor             bool
		want                 string
	}{
		{"refs/heads/main", "a", "b", true, "refresh"}, {"refs/heads/main", "a", "a", true, "current"},
		{"refs/heads/main", "a", "b", false, "diverged"}, {"refs/heads/issue", "a", "b", true, "off-main"},
		{"", "a", "b", true, "off-main"},
	} {
		if got := helperAction(c.branch, c.head, c.target, c.ancestor); got != c.want {
			t.Fatalf("%+v: %s", c, got)
		}
	}
}

func TestHelperRefreshRevalidatesBeforeUpdate(t *testing.T) {
	for _, mutation := range []string{"branch", "head", "dirty", "origin", "declarations"} {
		t.Run(mutation, func(t *testing.T) {
			f := repo(t, t.TempDir(), "supplier")
			before := git(t, f.root, "rev-parse", "HEAD")
			remoteAdvance(t, f)
			changed, updates := false, 0
			g := &gitModel{base: client().Git}
			g.after = func(dir string, args []string, out string, err error) (string, error) {
				if strings.Contains(strings.Join(args, " "), "merge --ff-only") {
					updates++
				}
				if !changed && len(args) > 0 && args[0] == "fetch" && err == nil {
					changed = true
					switch mutation {
					case "branch":
						git(t, f.root, "switch", "-c", "other")
					case "head":
						commit(t, f.root, "local", "keep")
					case "dirty":
						write(t, filepath.Join(f.root, "untracked"), "keep")
					case "origin":
						git(t, f.root, "remote", "set-url", "origin", f.seed)
					case "declarations":
						write(t, filepath.Join(f.root, "construct/deps"), "# changed")
					}
				}
				return out, err
			}
			r := RefreshHelper(context.Background(), f.root, g)
			if r.State != "skipped" || !changed || updates != 0 {
				t.Fatalf("race not refused: %+v changed=%t updates=%d", r, changed, updates)
			}
			if mutation != "head" && git(t, f.root, "rev-parse", "HEAD") != before {
				t.Fatal("own update after race")
			}
		})
	}
}

func TestHelperRefreshUncertainEffectRetainsProgress(t *testing.T) {
	f := repo(t, t.TempDir(), "supplier")
	target := remoteAdvance(t, f)
	g := &gitModel{base: client().Git, after: func(_ string, args []string, out string, err error) (string, error) {
		if strings.Contains(strings.Join(args, " "), "merge --ff-only") && err == nil {
			return out, errors.New("lost reply")
		}
		return out, err
	}}
	r := RefreshHelper(context.Background(), f.root, g)
	if r.State != "uncertain" || r.After != target || git(t, f.root, "rev-parse", "HEAD") != target {
		t.Fatalf("lost effect or fabricated confirmation: %+v", r)
	}
}

func TestHelperRefreshCancellationBeforeAndAfterEffect(t *testing.T) {
	for _, after := range []bool{false, true} {
		t.Run(map[bool]string{false: "before", true: "after"}[after], func(t *testing.T) {
			f := repo(t, t.TempDir(), "supplier")
			before := git(t, f.root, "rev-parse", "HEAD")
			target := remoteAdvance(t, f)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if !after {
				cancel()
			}
			g := &gitModel{base: client().Git, after: func(_ string, args []string, out string, err error) (string, error) {
				if after && strings.Contains(strings.Join(args, " "), "merge --ff-only") && err == nil {
					cancel()
				}
				return out, err
			}}
			r := RefreshHelper(ctx, f.root, g)
			want := before
			state := "skipped"
			if after {
				want = target
				state = "uncertain"
			}
			if r.State != state || git(t, f.root, "rev-parse", "HEAD") != want {
				t.Fatalf("cancellation: %+v", r)
			}
		})
	}
}

func TestHelperRefreshProtectsIgnoredFiles(t *testing.T) {
	f := repo(t, t.TempDir(), "supplier")
	commit(t, f.root, ".gitignore", "precious\n")
	git(t, f.root, "push", "origin", "main")
	git(t, f.seed, "pull", "--ff-only")
	write(t, filepath.Join(f.seed, "precious"), "remote")
	git(t, f.seed, "add", "-f", "precious")
	git(t, f.seed, "commit", "-m", "collision")
	git(t, f.seed, "push", "origin", "main")
	before := git(t, f.root, "rev-parse", "HEAD")
	write(t, filepath.Join(f.root, "precious"), "local")
	r := RefreshHelper(context.Background(), f.root, acquire.ExecGit{Raw: true})
	b, err := os.ReadFile(filepath.Join(f.root, "precious"))
	if r.State != "skipped" || err != nil || string(b) != "local" || git(t, f.root, "rev-parse", "HEAD") != before {
		t.Fatalf("ignored file lost: %+v %q %v", r, b, err)
	}
}
