package gitx

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestParseWorktrees(t *testing.T) {
	locked := "migration in progress"
	prunable := "gitdir file points to non-existent location"
	porcelain := worktreePorcelain(
		[]string{"worktree /bare repo", "bare"},
		[]string{"worktree /repo with spaces", "HEAD aaa", "branch refs/heads/main", "locked " + locked},
		[]string{"worktree /repo/wt/detached", "HEAD bbb", "detached", "prunable " + prunable},
	)

	got, err := ParseWorktrees(porcelain)
	if err != nil {
		t.Fatal(err)
	}
	want := []Worktree{
		{Path: "/bare repo", Bare: true},
		{Path: "/repo with spaces", HEAD: "aaa", Branch: "main", Locked: &locked},
		{Path: "/repo/wt/detached", HEAD: "bbb", Detached: true, Prunable: &prunable},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseWorktrees() = %+v, want %+v", got, want)
	}
}

func TestParseWorktreesRejectsMalformedRecords(t *testing.T) {
	tests := []struct {
		name      string
		porcelain []byte
	}{
		{"missing worktree path", worktreePorcelain([]string{"worktree ", "HEAD abc", "branch refs/heads/main"})},
		{"missing head", worktreePorcelain([]string{"worktree /repo", "branch refs/heads/main"})},
		{"malformed head", worktreePorcelain([]string{"worktree /repo", "HEAD abc def", "branch refs/heads/main"})},
		{"missing checkout state", worktreePorcelain([]string{"worktree /repo", "HEAD abc"})},
		{"empty branch", worktreePorcelain([]string{"worktree /repo", "HEAD abc", "branch refs/heads/"})},
		{"branch without heads prefix", worktreePorcelain([]string{"worktree /repo", "HEAD abc", "branch refs/tags/v1"})},
		{"multiple checkout states", worktreePorcelain([]string{"worktree /repo", "HEAD abc", "branch refs/heads/main", "detached"})},
		{"bare has head", worktreePorcelain([]string{"worktree /repo", "HEAD abc", "bare"})},
		{"duplicate optional attribute", worktreePorcelain([]string{"worktree /repo", "HEAD abc", "branch refs/heads/main", "locked one", "locked two"})},
		{"unknown attribute", worktreePorcelain([]string{"worktree /repo", "HEAD abc", "branch refs/heads/main", "future thing"})},
		{"record without separator", []byte("worktree /one\x00HEAD aaa\x00branch refs/heads/one\x00worktree /two\x00HEAD bbb\x00branch refs/heads/two\x00")},
		{"attribute before worktree", []byte("HEAD aaa\x00")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := ParseWorktrees(tt.porcelain); err == nil {
				t.Fatalf("ParseWorktrees() = %+v, want error", got)
			}
		})
	}
}

func TestParseWorktreesAllowsMissingFinalSeparator(t *testing.T) {
	got, err := ParseWorktrees([]byte("worktree /repo\x00HEAD abc\x00branch refs/heads/main\x00"))
	if err != nil {
		t.Fatal(err)
	}
	if want := []Worktree{{Path: "/repo", HEAD: "abc", Branch: "main"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseWorktrees() = %+v, want %+v", got, want)
	}
}

func FuzzParseWorktrees(f *testing.F) {
	for _, seed := range []string{
		"", "worktree /repo\x00HEAD abc\x00branch refs/heads/main\x00\x00",
		"worktree /bare\x00bare\x00\x00", "worktree /repo\x00HEAD abc\x00detached\x00locked\x00\x00",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, porcelain string) {
		worktrees, err := ParseWorktrees([]byte(porcelain))
		if err != nil {
			return
		}
		roundTrip, err := ParseWorktrees(canonicalWorktrees(worktrees))
		if err != nil {
			t.Fatalf("canonical parse: %v", err)
		}
		if !reflect.DeepEqual(roundTrip, worktrees) {
			t.Fatalf("canonical round trip = %+v, want %+v", roundTrip, worktrees)
		}
	})
}

func TestParseWorktreesConformanceNewlinePath(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit())
	path := filepath.Join(t.TempDir(), "linked\nworktree")
	testfix.Git(t, repo, "worktree", "add", "-b", "linked", path, "HEAD")
	t.Cleanup(func() { testfix.Git(t, repo, "worktree", "remove", "--force", path) })
	wantPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}

	porcelain := testfix.Capture(t, repo, "worktree", "list", "--porcelain", "-z")
	worktrees, err := ParseWorktrees([]byte(porcelain))
	if err != nil {
		t.Fatal(err)
	}
	for _, worktree := range worktrees {
		if worktree.Path == wantPath {
			if worktree.Branch != "linked" {
				t.Fatalf("linked worktree branch = %q, want linked", worktree.Branch)
			}
			return
		}
	}
	t.Fatalf("newline path %q absent from %+v", wantPath, worktrees)
}

func FuzzParseWorktreesCanonicalRoundTrip(f *testing.F) {
	f.Add("/repo with spaces", "abc", "feature/x", "", "", byte(0))
	f.Add("/repo\nwith newline", "abc", "feature/x", "locked\nreason", "prunable\nreason", byte(0))
	f.Add("/repo", "abc", " ", "", "", byte(0))
	f.Add("/repo", "def", "", "locked reason", "prunable reason", byte(1))
	f.Add("/bare", "", "", "", "", byte(2))
	f.Fuzz(func(t *testing.T, path, head, branch, locked, prunable string, mode byte) {
		clean := func(s string) string {
			s = strings.ReplaceAll(s, "\x00", "")
			if s == "" {
				return "x"
			}
			return s
		}
		cleanToken := func(s string) string {
			s = strings.Map(func(r rune) rune {
				if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == 0 {
					return -1
				}
				return r
			}, s)
			if s == "" {
				return "x"
			}
			return s
		}
		w := Worktree{Path: clean(path)}
		switch mode % 3 {
		case 0:
			w.HEAD, w.Branch = cleanToken(head), cleanToken(branch)
		case 1:
			w.HEAD, w.Detached = cleanToken(head), true
		case 2:
			w.Bare = true
		}
		if mode&4 != 0 {
			value := strings.ReplaceAll(locked, "\x00", "")
			w.Locked = &value
		}
		if mode&8 != 0 {
			value := strings.ReplaceAll(prunable, "\x00", "")
			w.Prunable = &value
		}
		got, err := ParseWorktrees(canonicalWorktrees([]Worktree{w}))
		if err != nil {
			t.Fatal(err)
		}
		if want := []Worktree{w}; !reflect.DeepEqual(got, want) {
			t.Fatalf("ParseWorktrees(canonical) = %+v, want %+v", got, want)
		}
	})
}

func canonicalWorktrees(worktrees []Worktree) []byte {
	var b strings.Builder
	for i, w := range worktrees {
		if i > 0 {
			b.WriteByte(0)
		}
		b.WriteString("worktree ")
		b.WriteString(w.Path)
		b.WriteByte(0)
		if w.Bare {
			b.WriteString("bare\x00")
		} else {
			b.WriteString("HEAD ")
			b.WriteString(w.HEAD)
			b.WriteByte(0)
			switch {
			case w.Branch != "":
				b.WriteString("branch refs/heads/")
				b.WriteString(w.Branch)
				b.WriteByte(0)
			case w.Detached:
				b.WriteString("detached\x00")
			}
		}
		if w.Locked != nil {
			b.WriteString("locked")
			if *w.Locked != "" {
				b.WriteByte(' ')
				b.WriteString(*w.Locked)
			}
			b.WriteByte(0)
		}
		if w.Prunable != nil {
			b.WriteString("prunable")
			if *w.Prunable != "" {
				b.WriteByte(' ')
				b.WriteString(*w.Prunable)
			}
			b.WriteByte(0)
		}
	}
	return []byte(b.String())
}

func worktreePorcelain(records ...[]string) []byte {
	var b strings.Builder
	for _, record := range records {
		for _, line := range record {
			b.WriteString(line)
			b.WriteByte(0)
		}
		b.WriteByte(0)
	}
	return []byte(b.String())
}
