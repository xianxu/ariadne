package refresh

import (
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/pkg/layergraph"
)

// Recorded incident: echo's main changed a substrate transport from SSH to
// HTTPS while its clean resting slot still declared SSH. Both transports below
// are synthetic; the transport seam fetches from the private bare fixture.
func TestDeclarationTransportRecovery(t *testing.T) {
	for _, mode := range []string{"behind", "divergent", "substrate", "credential"} {
		t.Run(mode, func(t *testing.T) {
			base := t.TempDir()
			host, dep := repo(t, base, "host"), repo(t, base, "dep")
			const ssh = "git@github.com:fixture/dep.git"
			git(t, dep.root, "remote", "set-url", "origin", ssh)
			c := client()
			c.Git = localTransportGit{GitRunner: c.Git, checkout: dep.root, remote: dep.remote}
			commit(t, host.seed, "construct/deps", "substrate ../dep "+ssh+"\n")
			git(t, host.seed, "push", "origin", "main")
			git(t, host.root, "pull", "--ff-only")
			subject := host
			if mode == "substrate" {
				subject = dep
				// A data declaration changes inside the substrate being refreshed.
				commit(t, dep.seed, "construct/deps", "data "+ssh+" content\n")
				git(t, dep.seed, "push", "origin", "main")
				git(t, dep.root, "fetch", dep.remote, "main")
				git(t, dep.root, "merge", "--ff-only", "FETCH_HEAD")
			}
			content := "substrate ../dep https://github.com/fixture/dep.git\n"
			label := `substrate "../dep"`
			if mode == "credential" {
				content = "substrate ../dep https://user:secret@example.invalid/dep?token=hidden\n"
			}
			if mode == "substrate" {
				content = "data https://github.com/fixture/dep.git content\n"
				label = `data "content"`
			}
			commit(t, subject.seed, "construct/deps", content)
			git(t, subject.seed, "push", "origin", "main")
			if mode == "divergent" {
				commit(t, host.root, "local", "retain local work")
			}
			before := git(t, subject.root, "rev-parse", "HEAD")
			hostBefore := git(t, host.root, "rev-parse", "HEAD")
			var err error
			var n int
			if mode == "substrate" {
				_, err, n = runSubstrate(t, host, c)
			} else {
				_, err, n = run(t, host, c, false)
			}
			if err == nil {
				t.Fatal("declaration change accepted")
			}
			message := err.Error()
			for _, want := range []string{subject.root, "construct/deps", "HEAD", "origin/main", label, "source", "./bootstrap.sh", "retry"} {
				if !strings.Contains(message, want) {
					t.Errorf("missing %q in %s", want, message)
				}
			}
			if mode == "divergent" {
				if !strings.Contains(message, "reconcile this branch first") || strings.Contains(message, "git merge --ff-only origin/main") {
					t.Errorf("unsafe divergence guidance: %s", message)
				}
			} else if !strings.Contains(message, "git merge --ff-only origin/main") {
				t.Errorf("missing recovery: %s", message)
			}
			if strings.Contains(message, ssh) || strings.Contains(message, "https://") || strings.Contains(message, "secret") || strings.Contains(message, "hidden") {
				t.Errorf("source printed: %s", message)
			}
			if n != 0 || git(t, subject.root, "rev-parse", "HEAD") != before || git(t, host.root, "rev-parse", "HEAD") != hostBefore {
				t.Fatal("refusal changed checkout or compiled")
			}
		})
	}
}

// Keep the observed SSH origin while using real local Git state and transport.
type localTransportGit struct {
	acquire.GitRunner
	checkout, remote string
}

func (g localTransportGit) Run(ctx context.Context, dir string, args ...string) (string, error) {
	if dir == g.checkout && len(args) > 3 && args[0] == "fetch" {
		args = append([]string(nil), args...)
		args[3] = g.remote
	}
	return g.GitRunner.Run(ctx, dir, args...)
}

func TestDeclarationChanges(t *testing.T) {
	for _, tt := range []struct {
		name, before, after string
		want                []string
		absent              []string
	}{
		{"source", "substrate ../dep git@github.com:fixture/dep.git", "substrate ../dep https://user:secret@example.invalid/dep?token=hidden", []string{`row 1`, `substrate "../dep"`, "source"}, []string{"secret", "hidden", "https://", "git@"}},
		{"path", "substrate ../old", "substrate ../new", []string{`"../old"`, `"../new"`, "path"}, nil},
		{"mount", "data ../data.git old", "data ../data.git new", []string{`data "old"`, `data "new"`, "mount"}, []string{"../data.git"}},
		{"kind", "substrate ../dep", "data ../data.git content", []string{"kind", "path", "source", "mount"}, []string{"../data.git"}},
		{"add", "", "substrate ../new", []string{`substrate "../new"`, "added"}, nil},
		{"remove", "data ../data.git content", "", []string{`data "content"`, "removed"}, []string{"../data.git"}},
		{"reorder", "substrate ../a\nsubstrate ../b", "substrate ../b\nsubstrate ../a", []string{"row 1", "row 2", `"../a"`, `"../b"`, "path"}, nil},
		{"duplicate", "substrate ../a\nsubstrate ../a", "substrate ../a\nsubstrate ../a ../secret.git", []string{"row 2", "source"}, []string{"row 1", "secret"}},
		{"same", "substrate ../a", "substrate ../a", nil, []string{"row"}},
		{"url label", "", "substrate https://user:secret@example.invalid/dep?token=hidden", []string{"row 1", "substrate", "added"}, []string{"secret", "hidden", "https://"}},
		{"scp label", "data ../data.git content", "data ../data.git user:secret@example.invalid/path", []string{"row 1", "mount"}, []string{"secret", "example.invalid"}},
		{"query label", "", "data ../data.git //example.invalid/path?token=secret", []string{"row 1", "added"}, []string{"secret", "example.invalid"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			before, err := layergraph.ParseRows(tt.before)
			if err != nil {
				t.Fatal(err)
			}
			after, err := layergraph.ParseRows(tt.after)
			if err != nil {
				t.Fatal(err)
			}
			got := declarationChanges(before, after)
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("missing %q: %s", want, got)
				}
			}
			for _, absent := range tt.absent {
				if strings.Contains(got, absent) {
					t.Errorf("unexpected %q: %s", absent, got)
				}
			}
			if tt.name == "same" && got != "" {
				t.Errorf("unchanged rows: %q", got)
			}
		})
	}
}
