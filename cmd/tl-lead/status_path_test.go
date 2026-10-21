package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestStatusArgs(t *testing.T) {
	for _, n := range []int{1, 2, 3} {
		for count := 0; count <= 4; count++ {
			args := []string{"explicit.md", "heading", "body.md", "extra"}[:count]
			got, err := statusArgs(args, n, "/checkout/session.md")
			valid := count == n || count == n-1
			if (err == nil) != valid {
				t.Fatalf("arity %d, args %v: %v", n, args, err)
			}
			if !valid {
				continue
			}
			want := append([]string(nil), args...)
			if count == n-1 {
				want = append([]string{"/checkout/session.md"}, want...)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("arity %d, args %v: got %v, want %v", n, args, got, want)
			}
		}
	}
}

func statusCheckout(t *testing.T, hub bool) string {
	t.Helper()
	root := t.TempDir()
	if out, err := exec.Command("git", "-C", root, "init", "--quiet").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}
	if hub {
		if err := os.WriteFile(filepath.Join(root, ".hub"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Git reports the physical spelling (not /var on macOS).
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestStatusPath(t *testing.T) {
	t.Run("nested checkout with hub", func(t *testing.T) {
		root := statusCheckout(t, true)
		nested := filepath.Join(root, "nested", "deeper")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(nested)
		got, err := defaultStatusPath()
		if err != nil || got != filepath.Join(root, "session.md") {
			t.Fatalf("got %q, %v", got, err)
		}
		if _, err := os.Stat(got); !os.IsNotExist(err) {
			t.Fatalf("resolving a default created the missing file: %v", err)
		}
	})
	t.Run("checkout without hub", func(t *testing.T) {
		t.Chdir(statusCheckout(t, false))
		if _, err := defaultStatusPath(); err == nil || !strings.Contains(err.Error(), ".hub") {
			t.Fatalf("missing hub: %v", err)
		}
	})
	t.Run("outside git", func(t *testing.T) {
		t.Chdir(t.TempDir())
		if _, err := defaultStatusPath(); err == nil {
			t.Fatal("default path outside git succeeded")
		}
	})
}

func TestStatusDefaultRefusals(t *testing.T) {
	for _, tc := range []struct {
		name string
		git  bool
	}{
		{"outside git", false},
		{"checkout without hub", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.git {
				root = statusCheckout(t, false)
			}
			t.Chdir(root)
			f := &statusPathSync{fakeSessionSync: newFake(nil)}
			_, err := run(t, deps{ss: f}, "status", "seen")
			if err == nil || !strings.Contains(err.Error(), ".hub") {
				t.Fatalf("omitted FILE should explain the hub requirement: %v", err)
			}
			if len(f.paths) != 0 {
				t.Fatalf("invalid default reached session-sync: %v", f.paths)
			}
		})
	}
	t.Run("missing session remains missing", func(t *testing.T) {
		root := statusCheckout(t, true)
		t.Chdir(root)
		_, err := run(t, deps{ss: newFake(nil)}, "status", "submit")
		if !os.IsNotExist(err) {
			t.Fatalf("missing session should return its read failure: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "session.md")); !os.IsNotExist(err) {
			t.Fatalf("command created a missing session: %v", err)
		}
	})
}

// Record the paths delivered to real command behavior, keeping the stateful
// fake's reads, writes, index reconciliation, and migration planning intact.
type statusPathSync struct {
	*fakeSessionSync
	paths []string
}

func (s *statusPathSync) Peek(file string) (string, error) {
	s.paths = append(s.paths, file)
	return s.fakeSessionSync.Peek(file)
}

func (s *statusPathSync) ReadSection(file, heading string) (string, string, error) {
	s.paths = append(s.paths, file)
	return s.fakeSessionSync.ReadSection(file, heading)
}

func (s *statusPathSync) Diff(file string) (string, error) {
	s.paths = append(s.paths, file)
	return s.fakeSessionSync.Diff(file)
}

func (s *statusPathSync) Status(file string) (SyncStatus, error) {
	s.paths = append(s.paths, file)
	return s.fakeSessionSync.Status(file)
}

func (s *statusPathSync) MarkAllSeen(file string) error {
	s.paths = append(s.paths, file)
	return s.fakeSessionSync.MarkAllSeen(file)
}

func TestStatusDefaultLeaves(t *testing.T) {
	for _, tc := range []struct {
		command []string
		tail    []string
	}{
		{[]string{"submit"}, nil},
		{[]string{"section"}, []string{"Needs you", "BODY"}},
		{[]string{"thread"}, []string{"BODY"}},
		{[]string{"seen"}, nil},
		{[]string{"owed"}, nil},
		{[]string{"check"}, nil},
		{[]string{"clobbered"}, []string{"--done"}},
		{[]string{"slots"}, nil},
		{[]string{"instructions", "list"}, nil},
		{[]string{"instructions", "link"}, []string{"1", "session.side/2026-10-20-01-answer.md", "--hash", "HASH"}},
		{[]string{"index"}, nil},
		{[]string{"migrate-answers"}, nil},
	} {
		for _, explicit := range []bool{false, true} {
			name := strings.Join(tc.command, " ")
			if explicit {
				name += "/explicit outside checkout"
			}
			t.Run(name, func(t *testing.T) {
				root := statusCheckout(t, true)
				t.Chdir(root)
				file := filepath.Join(root, "session.md")
				if explicit {
					t.Chdir(t.TempDir())
					file = filepath.Join(t.TempDir(), "session.md")
				}
				f := &statusPathSync{fakeSessionSync: newFake(map[string]string{
					"Needs you": "old", "Threads": "### Existing\nold", "Slots": "old", instructionsHeading: "1. question",
				})}
				f.front += "\noperator: test"
				for _, dir := range []string{"session.side", "qa/test"} {
					if err := os.MkdirAll(filepath.Join(filepath.Dir(file), dir), 0o755); err != nil {
						t.Fatal(err)
					}
				}
				for path, body := range map[string]string{
					file: f.text(), file + ".clobbered": "recovered", filepath.Join(filepath.Dir(file), "session.side", "2026-10-20-01-answer.md"): "---\ntype: side-doc\nparent: ../session.md\n---\nanswer",
				} {
					if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				bodyFile := filepath.Join(t.TempDir(), "body.md")
				if err := os.WriteFile(bodyFile, []byte("### New\nnew body"), 0o644); err != nil {
					t.Fatal(err)
				}
				args := append([]string{"status"}, tc.command...)
				if explicit {
					args = append(args, file)
				}
				for _, a := range tc.tail {
					switch a {
					case "BODY":
						a = bodyFile
					case "HASH":
						a = fakeHash(f.sections[instructionsHeading])
					}
					args = append(args, a)
				}
				out, err := run(t, deps{ss: f, sdlc: &fakeSDLC{}}, args...)
				if err != nil {
					t.Fatalf("%v: %v\n%s", args, err, out)
				}
				if tc.command[0] == "clobbered" {
					if _, err := os.Stat(file + ".clobbered"); !os.IsNotExist(err) {
						t.Fatalf("clobbered command did not use %s: %v", file, err)
					}
					return
				}
				if len(f.paths) == 0 {
					t.Fatal("command did not reach session-sync behavior")
				}
				for _, got := range f.paths {
					if got != file {
						t.Errorf("session-sync path %q, want %q", got, file)
					}
				}
			})
		}
	}
}
