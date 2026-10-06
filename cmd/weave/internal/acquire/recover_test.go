package acquire

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// recoveryFleet builds a numbered environment whose host declares a sourceless
// substrate, and the fleet directory its primary checkout lives in (#296).
func recoveryFleet(t *testing.T) (Client, string, string) {
	t.Helper()
	c, env := scopedClient(t)
	fleet := canonical(t.TempDir())
	c.Policy.PrimaryRoot = filepath.Join(fleet, "host")
	put(t, filepath.Join(c.Policy.HostRoot, "construct/deps"), "substrate ../base\n")
	return c, env, fleet
}

// primarySibling is the operator's checkout of base next to the primary, with
// the remote origin a slot must clone from.
func primarySibling(t *testing.T, fleet, origin string) string {
	t.Helper()
	sibling := filepath.Join(fleet, "base")
	put(t, filepath.Join(sibling, "README"), "sibling")
	gitFixture(t, sibling, "init")
	if origin != "" {
		gitFixture(t, sibling, "remote", "add", "origin", origin)
	}
	return sibling
}

func mainRemote(t *testing.T) string {
	t.Helper()
	remote := origin(t, t.TempDir(), "base", "", true)
	gitFixture(t, remote, "branch", "main")
	gitFixture(t, remote, "symbolic-ref", "HEAD", "refs/heads/main")
	return remote
}

func TestRestoreRecoversSourcelessSubstrateFromPrimarySibling(t *testing.T) {
	c, env, fleet := recoveryFleet(t)
	c.Git = fixtureRemoteGit{remote: mainRemote(t)}
	sibling := primarySibling(t, fleet, "https://example.test/base.git")
	dest := filepath.Join(env, "base")

	r, err := c.Restore(context.Background(), c.Policy.HostRoot, true)
	if err == nil || len(r.Missing) != 1 || len(r.Recovered) != 1 {
		t.Fatalf("dry run: %+v %v", r, err)
	}
	if _, e := os.Lstat(dest); !os.IsNotExist(e) {
		t.Fatalf("dry run cloned: %v", e)
	}

	r, err = c.Restore(context.Background(), c.Policy.HostRoot, false)
	if err != nil {
		t.Fatal(err)
	}
	want := Recovery{Owner: c.Policy.HostRoot, Path: "../base", URL: "https://example.test/base.git", Sibling: sibling}
	if len(r.Recovered) != 1 || r.Recovered[0] != want {
		t.Fatalf("recovered %+v, want %+v", r.Recovered, want)
	}
	if got := gitFixture(t, dest, "config", "remote.origin.url"); got != want.URL {
		t.Fatalf("clone origin %s", got)
	}
	if _, e := os.Stat(filepath.Join(dest, "construct/base.manifest")); e != nil {
		t.Fatalf("clone is not the remote's layer: %v", e)
	}

	// Once restored, the checkout is reused and nothing is recovered again.
	if err := os.RemoveAll(sibling); err != nil {
		t.Fatal(err)
	}
	r, err = c.Restore(context.Background(), c.Policy.HostRoot, false)
	if err != nil || len(r.Recovered) != 0 {
		t.Fatalf("warm restore: %+v %v", r, err)
	}
}

func TestRestoreRefusesUnprovenSiblingSource(t *testing.T) {
	for _, tc := range []struct {
		name, reason string
		sibling      func(t *testing.T, fleet string)
	}{
		{"missing", "is not present", func(*testing.T, string) {}},
		{"not a checkout", "is not a repository checkout", func(t *testing.T, fleet string) {
			put(t, filepath.Join(fleet, "base", "README"), "plain")
		}},
		{"no origin", "has no origin", func(t *testing.T, fleet string) { primarySibling(t, fleet, "") }},
		{"local origin", "has a local origin", func(t *testing.T, fleet string) { primarySibling(t, fleet, "/elsewhere/base") }},
		{"other repository", "names repository other, not base", func(t *testing.T, fleet string) {
			primarySibling(t, fleet, "https://example.test/other.git")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, env, fleet := recoveryFleet(t)
			tc.sibling(t, fleet)
			_, err := c.Restore(context.Background(), c.Policy.HostRoot, false)
			if err == nil || !strings.Contains(err.Error(), "missing substrate") || !strings.Contains(err.Error(), "no source declared") || !strings.Contains(err.Error(), tc.reason) {
				t.Fatalf("error %v, want reason %q", err, tc.reason)
			}
			if _, e := os.Lstat(filepath.Join(env, "base")); !os.IsNotExist(e) {
				t.Fatalf("published a refused recovery: %v", e)
			}
		})
	}
}

func TestRestoreOutsideEnvironmentKeepsSourcelessError(t *testing.T) {
	root := filepath.Join(t.TempDir(), "host")
	put(t, filepath.Join(root, "construct/deps"), "substrate ../base\n")
	_, err := Restore(context.Background(), root, false)
	if err == nil || !strings.Contains(err.Error(), "record its source") || !strings.Contains(err.Error(), "no numbered environment") {
		t.Fatalf("error %v", err)
	}
}
