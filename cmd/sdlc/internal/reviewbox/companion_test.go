package reviewbox

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCheckoutCompanionLifetime(t *testing.T) {
	for _, action := range []string{"reset", "close", "sweep-live", "sweep-dead"} {
		t.Run(action, func(t *testing.T) {
			src, _, head := sourceRepo(t)
			root := t.TempDir()
			co, err := CreateContext(t.Context(), root, src, head, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer co.Close()
			base := co.Dir + ".sdlc-review"
			if err := os.MkdirAll(filepath.Join(base, "tmp"), 0700); err != nil {
				t.Fatal(err)
			}
			sentinel := filepath.Join(base, "tmp", "attempt-one")
			write(t, sentinel, "stale")
			switch action {
			case "reset":
				err = co.ResetContext(t.Context())
			case "close":
				err = co.Close()
			case "sweep-dead":
				// Releasing the actual flock models owner death, without a PID heuristic.
				co.lock.Close()
				co.lock = nil
				fallthrough
			case "sweep-live":
				_, problems := Sweep(root)
				if len(problems) != 0 {
					t.Fatal(problems)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			_, err = os.Stat(sentinel)
			if action == "sweep-live" {
				if err != nil {
					t.Fatalf("live owner's scratch removed: %v", err)
				}
				if _, err := os.Stat(filepath.Join(co.Dir, "a.txt")); err != nil {
					t.Fatal(err)
				}
			} else if !os.IsNotExist(err) {
				t.Fatalf("%s retained prior scratch: %v", action, err)
			}
		})
	}
}

func TestConfineRefusesInvalidCompanionBeforeWrites(t *testing.T) {
	for _, kind := range []string{"symlink", "file"} {
		t.Run(kind, func(t *testing.T) {
			src, _, head := sourceRepo(t)
			co, err := CreateContext(t.Context(), t.TempDir(), src, head, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer co.Close()
			outside := t.TempDir()
			base := co.Dir + ".sdlc-review"
			if kind == "symlink" {
				err = os.Symlink(outside, base)
			} else {
				err = os.WriteFile(base, []byte("foreign"), 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			c, err := Confine(t.Context(), co.Dir, true)
			if c != nil {
				c.Close()
			}
			if err == nil || !strings.Contains(err.Error(), base) {
				t.Fatalf("invalid companion accepted: %v", err)
			}
			entries, err := os.ReadDir(outside)
			if err != nil || len(entries) != 0 {
				t.Fatalf("wrote through companion: %v, %v", entries, err)
			}
			if _, err := os.Lstat(base); err != nil {
				t.Fatalf("refusal changed collision: %v", err)
			}
		})
	}
}

func TestSweepCompanionOrphans(t *testing.T) {
	for _, checkoutPresent := range []bool{false, true} {
		t.Run(map[bool]string{false: "alone", true: "with-checkout"}[checkoutPresent], func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "abandoned")
			if checkoutPresent {
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Mkdir(dir+".sdlc-review", 0700); err != nil {
				t.Fatal(err)
			}
			removed, problems := Sweep(root)
			if removed != 1 || len(problems) != 0 {
				t.Fatalf("sweep: %d removed, %v", removed, problems)
			}
			entries, err := os.ReadDir(root)
			if err != nil || len(entries) != 0 {
				t.Fatalf("orphan residue: %v, %v", entries, err)
			}
		})
	}
}

func TestConfinedScratchHasNoCheckoutProjectAncestor(t *testing.T) {
	// Replay the actual failure's ancestor shape, rather than trusting TMPDIR's
	// string alone. Parley's discovery walks project markers independently of Git.
	raw, err := os.ReadFile("testdata/parley422-review.log")
	if err != nil {
		t.Fatal(err)
	}
	captured := regexp.MustCompile(`chat outside default dir: (\S+)/workshop/parley/`).FindSubmatch(raw)
	if len(captured) != 2 || !strings.Contains(string(raw), string(captured[1])+"/.sdlc-review/tmp/") {
		t.Fatal("capture does not establish in-checkout scratch and escaped chat")
	}
	src, _, head := sourceRepo(t)
	co, err := CreateContext(t.Context(), t.TempDir(), src, head, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer co.Close()
	// The incident's tracked .parley marker, independent of Git's discovery ceiling.
	write(t, filepath.Join(co.Dir, ".parley"), "")
	c, err := Confine(t.Context(), co.Dir, true)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	scratch, err := os.MkdirTemp(c.Tmp(), "project-")
	if err != nil {
		t.Fatal(err)
	}
	for dir := scratch; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, ".parley")); err == nil {
			t.Fatalf("scratch %s discovers checkout project %s", scratch, dir)
		}
		if dir == filepath.Dir(dir) {
			break
		}
	}
}
