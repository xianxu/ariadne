package gitx

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestSnapshotPinsTreeAndBatchReadsLiteralBytes(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	want := "literal $Format:%H$\n\x00tail\n"
	err = tf.UpdateMany("snapshot", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{
			"cards/one.md": []byte(want), ".gitattributes": []byte("cards/* export-ignore export-subst\n"),
		}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	view, err := tf.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	all, err := view.Files("")
	if err != nil || len(all) < 2 {
		t.Fatalf("root snapshot = %#v, %v", all, err)
	}
	err = tf.Update("cards/one.md", "later", func([]byte) ([]byte, error) { return []byte("changed"), nil })
	if err != nil {
		t.Fatal(err)
	}
	files, err := view.Files("cards")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != "cards/one.md" || string(files[0].Content) != want || files[0].Mode != "100644" {
		t.Fatalf("snapshot = %#v", files)
	}
	blob, err := view.ReadBlob(files[0].OID)
	if err != nil || string(blob) != want {
		t.Fatalf("blob = %q, %v", blob, err)
	}
}

func FuzzSnapshotFrames(f *testing.F) {
	oid := strings.Repeat("a", 40)
	f.Add([]byte("100644 blob "+oid+"\tcards/one.md\x00"), []byte(oid+" blob 1\nx\n"))
	f.Add([]byte("garbage"), []byte("garbage"))
	f.Fuzz(func(t *testing.T, tree, blobs []byte) {
		files, err := parseSnapshotTree(tree)
		if err != nil {
			return
		}
		result, err := parseSnapshotBlobs(blobs, files)
		if err != nil {
			return
		}
		if len(result) != len(files) {
			t.Fatal("changed snapshot cardinality")
		}
		for i := range files {
			if result[i].OID != files[i].OID || result[i].Path != files[i].Path || len(result[i].Content) > 1<<20 {
				t.Fatal("accepted mismatched or oversized snapshot")
			}
		}
	})
}

func TestSnapshotBatchRejectsMalformedFrames(t *testing.T) {
	for _, raw := range []string{"", "abc missing\n", "abc blob -1\n", "abc blob 5\nxx", "abc blob 1\nxX"} {
		if _, err := parseSnapshotBlobs([]byte(raw), []TreeFile{{OID: "abc"}}); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
}

// #290: a snapshot right after an unchanged presence probe reads the tip the
// probe saw without fetching; a moved remote is fetched; the probe's tip is used
// once; and tracker fetches never trigger automatic maintenance.
func TestSnapshotSkipsTheFetchWhenTheTipIsUnchanged(t *testing.T) {
	repo, origin := trunkFixture(t, "seed\n")
	trace := filepath.Join(t.TempDir(), "trace")
	t.Setenv("GIT_TRACE", trace)
	fetches := func() (n int, noMaint bool) {
		raw, _ := os.ReadFile(trace)
		noMaint = true
		for _, l := range strings.Split(string(raw), "\n") {
			if strings.Contains(l, "built-in: git fetch") {
				n++
				noMaint = noMaint && strings.Contains(l, "--no-auto-maintenance")
			}
		}
		return n, noMaint
	}
	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	tip := strings.TrimSpace(testfix.Capture(t, origin, "rev-parse", "main"))
	if ok, err := tf.RemoteExists(); !ok || err != nil {
		t.Fatalf("presence: %v %v", ok, err)
	}
	view, err := tf.Snapshot()
	if n, _ := fetches(); err != nil || n != 0 || view.Ref() != tip {
		t.Fatalf("unchanged tip: %d fetches, ref %s want %s, %v", n, view.Ref(), tip, err)
	}

	other := filepath.Join(t.TempDir(), "other")
	testfix.Git(t, "", "clone", "-q", origin, other)
	testfix.Git(t, other, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "moved")
	testfix.Git(t, other, "push", "-q", "origin", "main")
	moved := strings.TrimSpace(testfix.Capture(t, origin, "rev-parse", "main"))
	if _, err := tf.RemoteExists(); err != nil {
		t.Fatal(err)
	}
	view, err = tf.Snapshot()
	if n, noMaint := fetches(); err != nil || n != 1 || !noMaint || view.Ref() != moved {
		t.Fatalf("moved tip: %d fetches (no-maintenance %v), ref %s want %s, %v", n, noMaint, view.Ref(), moved, err)
	}

	if _, err := tf.Snapshot(); err != nil {
		t.Fatal(err)
	}
	if n, _ := fetches(); n != 2 {
		t.Fatalf("a probe's tip was reused by a later snapshot: %d fetches, want 2", n)
	}

	// Any fetch consumes the probe: a snapshot after probe → fetch fetches.
	if _, err := tf.RemoteExists(); err != nil {
		t.Fatal(err)
	}
	if _, err := tf.Read("note.md"); err != nil {
		t.Fatal(err)
	}
	if _, err := tf.Snapshot(); err != nil {
		t.Fatal(err)
	}
	if n, _ := fetches(); n != 4 {
		t.Fatalf("a probe survived an intervening fetch: %d fetches, want 4", n)
	}
}

func TestFilesSelectedFiltersBeforeReadingBlobs(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	writes := map[string][]byte{"issues/000001-one.md": []byte("issue document")}
	blob := []byte(strings.Repeat("x", 1<<20))
	for i := 0; i < 33; i++ {
		writes[fmt.Sprintf("issues/assets/%06d/a.bin", i)] = blob
	}
	if err := tf.UpdateMany("assets", func(*TrunkView) (TrunkWrite, error) { return TrunkWrite{Write: writes}, nil }); err != nil {
		t.Fatal(err)
	}
	view, err := tf.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	files, err := view.FilesSelected("issues", func(f TreeFile) bool {
		seen++
		if len(f.Content) != 0 || f.OID == "" || f.Mode == "" {
			t.Fatal("selector must receive metadata only")
		}
		return path.Dir(f.Path) == "issues"
	})
	if err != nil {
		t.Fatal(err)
	}
	if seen != 34 || len(files) != 1 || string(files[0].Content) != "issue document" {
		t.Fatalf("selected=%d visited=%d", len(files), seen)
	}
	empty, err := view.FilesSelected("issues", func(TreeFile) bool { return false })
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty selection = %v, %v", empty, err)
	}
	if _, err := view.Files("issues"); err == nil {
		t.Fatal("unfiltered aggregate must still exceed existing output limit")
	}
}

func TestFilesSelectedPreservesBlobLimit(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	if err := tf.UpdateMany("oversized", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{"issues/one.md": []byte(strings.Repeat("x", SnapshotBlobLimit+1))}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	view, err := tf.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := view.FilesSelected("issues", func(TreeFile) bool { return true }); err == nil || !strings.Contains(err.Error(), "limit 1 MiB") {
		t.Fatalf("oversized selection: %v", err)
	}
}

// #598: TreeFilesSelected reads any commit under the caller's context, and refuses a cwd-defaulting dir, an option
// revision and a non-canonical prefix before running git.
func TestTreeFilesSelected(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	if err := tf.UpdateMany("sidecars", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{"r/a.json": []byte("a"), "r/b.txt": []byte("b")}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	view, err := tf.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	files, err := TreeFilesSelected(context.Background(), repo, view.ref, "r", func(f TreeFile) bool { return strings.HasSuffix(f.Path, ".json") })
	if err != nil || len(files) != 1 || files[0].Path != "r/a.json" || string(files[0].Content) != "a" {
		t.Fatalf("files = %#v, %v", files, err)
	}
	if none, err := TreeFilesSelected(context.Background(), repo, view.ref, "absent", nil); err != nil || len(none) != 0 {
		t.Fatalf("absent prefix = %#v, %v", none, err)
	}
	for name, call := range map[string]func() error{
		"empty dir": func() error { _, err := TreeFilesSelected(context.Background(), " ", view.ref, "r", nil); return err },
		"option rev": func() error {
			_, err := TreeFilesSelected(context.Background(), repo, "--output=x", "r", nil)
			return err
		},
		"empty rev": func() error { _, err := TreeFilesSelected(context.Background(), repo, "", "r", nil); return err },
		"traversal": func() error {
			_, err := TreeFilesSelected(context.Background(), repo, view.ref, "../r", nil)
			return err
		},
		"cancelled": func() error {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_, err := TreeFilesSelected(ctx, repo, view.ref, "r", nil)
			return err
		},
	} {
		if call() == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// TreeFilesAt reads exact paths, at the root or nested, in one listing; a missing path is absent, never an error.
func TestTreeFilesAt(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	if err := tf.UpdateMany("traceability", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{"t.yaml": []byte("root"), "atlas/t.yaml": []byte("nested"), "atlas/t.yaml.bak": []byte("no")}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	view, err := tf.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	files, err := TreeFilesAt(context.Background(), repo, view.ref, []string{"t.yaml", "atlas/t.yaml", "absent.yaml"})
	got := map[string]string{}
	for _, f := range files {
		got[f.Path] = string(f.Content)
	}
	if err != nil || len(got) != 2 || got["t.yaml"] != "root" || got["atlas/t.yaml"] != "nested" {
		t.Fatalf("files = %v, %v", got, err)
	}
	if none, err := TreeFilesAt(context.Background(), repo, view.ref, nil); err != nil || none != nil {
		t.Fatalf("no paths = %v, %v", none, err)
	}
	for _, bad := range []string{"", ".", "..", "../t.yaml", "/t.yaml", "a//t.yaml", "a\\t.yaml"} {
		if _, err := TreeFilesAt(context.Background(), repo, view.ref, []string{bad}); err == nil {
			t.Errorf("%q: accepted", bad)
		}
	}
}
