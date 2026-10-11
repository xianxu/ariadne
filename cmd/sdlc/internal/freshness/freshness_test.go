package freshness

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func mustZone(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func TestCutoff(t *testing.T) {
	la := mustZone(t, "America/Los_Angeles")
	at := func(s string) time.Time {
		v, err := time.ParseInLocation("2006-01-02 15:04", s, la)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	for _, c := range []struct{ now, want string }{
		{"2026-10-10 13:00", "2026-10-10 12:00"}, // afternoon: today's noon
		{"2026-10-10 12:00", "2026-10-10 12:00"}, // exactly noon counts
		{"2026-10-10 11:59", "2026-10-09 12:00"}, // morning: yesterday's
		{"2026-10-10 00:30", "2026-10-09 12:00"}, // after midnight
		{"2026-11-01 10:00", "2026-10-31 12:00"}, // DST ends 11-01 02:00: still local noon
		{"2026-03-08 13:00", "2026-03-08 12:00"}, // DST starts 03-08
		{"2026-01-01 09:00", "2025-12-31 12:00"}, // across a year
	} {
		if got := Cutoff(at(c.now), la); !got.Equal(at(c.want)) {
			t.Errorf("Cutoff(%s) = %s, want %s", c.now, got.In(la), c.want)
		}
	}
	// The zone decides: 20:00 UTC is 13:00 in LA (today's noon passed) but 05:00
	// next day in Tokyo (yesterday's noon).
	now := time.Date(2026, 10, 10, 20, 0, 0, 0, time.UTC)
	if got := Cutoff(now, la); !got.Equal(at("2026-10-10 12:00")) {
		t.Errorf("LA cutoff = %s", got)
	}
	tokyo := mustZone(t, "Asia/Tokyo")
	if got := Cutoff(now, tokyo); !got.Equal(time.Date(2026, 10, 10, 12, 0, 0, 0, tokyo)) {
		t.Errorf("Tokyo cutoff = %s", got)
	}
}

func TestLoadLocation(t *testing.T) {
	root := t.TempDir()
	loc, src, err := LoadLocation(root)
	if err != nil || loc != time.Local || !strings.Contains(src, "local zone") {
		t.Fatalf("absent config: %v %q %v", loc, src, err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".sdlc"), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(s string) {
		if err := os.WriteFile(filepath.Join(root, ConfigPath), []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"version":1,"timezone":"America/Los_Angeles"}`)
	loc, src, err = LoadLocation(root)
	if err != nil || loc.String() != "America/Los_Angeles" || src != ConfigPath {
		t.Fatalf("valid config: %v %q %v", loc, src, err)
	}
	for _, bad := range []string{`{`, `{"version":2,"timezone":"UTC"}`, `{"version":1}`, `{"version":1,"timezone":"Mars/Olympus"}`} {
		write(bad)
		if _, _, err := LoadLocation(root); err == nil || !strings.Contains(err.Error(), ConfigPath) {
			t.Errorf("config %s: want an error naming %s, got %v", bad, ConfigPath, err)
		}
	}
}

// fixture is a real repository: main is a chain of commits with controlled
// committer dates, and a feature branch forks from its first commit.
type fixture struct {
	t   *testing.T
	dir string
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{t: t, dir: t.TempDir()}
	f.git("init", "-q", "-b", "main")
	return f
}

func (f *fixture) git(args ...string) string {
	f.t.Helper()
	out, err := f.run(args...)
	if err != nil {
		f.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func (f *fixture) run(args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = f.dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	return cmd.Output()
}

func (f *fixture) commit(path string, when time.Time) string {
	f.t.Helper()
	full := filepath.Join(f.dir, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(when.String()), 0o644); err != nil {
		f.t.Fatal(err)
	}
	f.git("add", path)
	date := when.Format(time.RFC3339)
	cmd := exec.Command("git", "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "commit", "-q", "-m", path)
	cmd.Dir = f.dir
	cmd.Env = append(os.Environ(), "GIT_COMMITTER_DATE="+date, "GIT_AUTHOR_DATE="+date, "GIT_CONFIG_GLOBAL=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		f.t.Fatalf("commit: %v\n%s", err, out)
	}
	return f.git("rev-parse", "HEAD")
}

func isCode(paths []string) bool {
	for _, p := range paths {
		if !strings.HasSuffix(p, ".md") {
			return true
		}
	}
	return false
}

func TestCheck(t *testing.T) {
	cutoff := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	before := func(h int) time.Time { return cutoff.Add(-time.Duration(h) * time.Hour) }

	// main: root (-48h) → docs (-3h) → code (-2h) → late code (+1h).
	// feature forks from root.
	setup := func(t *testing.T) (*fixture, map[string]string) {
		f := newFixture(t)
		c := map[string]string{"root": f.commit("a.go", before(48))}
		f.git("branch", "feature")
		c["docs"] = f.commit("notes.md", before(3))
		c["code"] = f.commit("b.go", before(2))
		c["late"] = f.commit("c.go", cutoff.Add(time.Hour))
		return f, c
	}

	t.Run("stale: main gained code before the cutoff", func(t *testing.T) {
		f, c := setup(t)
		r := Check(f.run, "main", "feature", cutoff, isCode)
		if r.State != Stale || r.Required != c["code"] || strings.Join(r.Code, ",") != "b.go" || r.Docs != 1 {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("fresh after integrating main as of the cutoff; a later landing is not required", func(t *testing.T) {
		f, c := setup(t)
		f.git("checkout", "-q", "feature")
		f.git("merge", "-q", "--no-edit", c["code"])
		if r := Check(f.run, "main", "HEAD", cutoff, isCode); r.State != Fresh {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("docs-only: the missing landings carry no code", func(t *testing.T) {
		f, c := setup(t)
		r := Check(f.run, "main", "feature", before(2).Add(-time.Minute), isCode)
		if r.State != DocsOnly || r.Required != c["docs"] || r.Docs != 1 || len(r.Code) != 0 {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("fresh when nothing on main is old enough", func(t *testing.T) {
		f, _ := setup(t)
		if r := Check(f.run, "main", "feature", before(100), isCode); r.State != Fresh || r.Required != "" {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("first parent is the landing time: an old commit merged late is not required yet", func(t *testing.T) {
		f, c := setup(t)
		// A side branch commit authored long ago lands on main after the cutoff.
		f.git("checkout", "-q", "-b", "side", c["root"])
		f.commit("old.go", before(30))
		f.git("checkout", "-q", "main")
		cmd := exec.Command("git", "-c", "user.name=t", "-c", "user.email=t@t", "merge", "-q", "--no-ff", "--no-edit", "side")
		cmd.Dir = f.dir
		cmd.Env = append(os.Environ(), "GIT_COMMITTER_DATE="+cutoff.Add(2*time.Hour).Format(time.RFC3339), "GIT_CONFIG_GLOBAL=/dev/null")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("merge: %v\n%s", err, out)
		}
		f.git("checkout", "-q", "feature")
		f.git("merge", "-q", "--no-edit", c["code"])
		if r := Check(f.run, "main", "HEAD", cutoff, isCode); r.State != Fresh {
			t.Fatalf("old.go landed after the cutoff, so it is not required: %+v", r)
		}
	})
	t.Run("inconclusive without a main ref", func(t *testing.T) {
		f, _ := setup(t)
		r := Check(f.run, "origin/main", "feature", cutoff, isCode)
		if r.State != Inconclusive || r.Why == "" {
			t.Fatalf("got %+v", r)
		}
	})
}
