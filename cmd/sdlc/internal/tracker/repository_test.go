package tracker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func fixture(t *testing.T) (*Repository, string, *gitx.TrunkFile) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	root := testfix.Repo(t, testfix.InitialCommit())
	testfix.Git(t, root, "config", "core.hooksPath", t.TempDir())
	origin := filepath.Join(t.TempDir(), "remote.git")
	testfix.Git(t, root, "init", "--bare", origin)
	testfix.Git(t, root, "remote", "add", "publication", origin)
	tf, err := gitx.NewTrunkFileContext(context.Background(), root, "publication", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	_, err = tf.Bootstrap(map[string][]byte{ManifestPath: ManifestBytes(), testPath: []byte(testCard)}, "bootstrap unique-op", func(gitx.BootstrapResult) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRepository(context.Background(), root, "publication")
	if err != nil {
		t.Fatal(err)
	}
	return r, root, tf
}

func TestRepositoryFreshSnapshotAndConditionalWrite(t *testing.T) {
	r, root, _ := fixture(t)
	before := testfix.Capture(t, root, "rev-parse", "HEAD")
	snap, err := r.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	old, _ := snap.Card("000252")
	changed := []byte(strings.Replace(testCard, "status: open", "status: working", 1))
	receipts := 0
	err = r.UpdateCard(old, changed, "test-claim-one", func(base, candidate string) error {
		receipts++
		if base != snap.Ref() || candidate == base {
			t.Error("unbound receipt")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if receipts != 1 {
		t.Fatalf("receipts %d", receipts)
	}
	current, err := r.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	now, _ := current.Card("000252")
	if now.BlobOID == old.BlobOID || string(now.Raw) != string(changed) {
		t.Fatal("write not visible")
	}
	if oldAgain, _ := snap.Card("000252"); string(oldAgain.Raw) != testCard {
		t.Fatal("old snapshot changed")
	}
	if err := r.UpdateCard(old, changed, "test-claim-two", func(string, string) error { return nil }); !errors.Is(err, ErrCardChanged) {
		t.Fatalf("stale write: %v", err)
	}
	if testfix.Capture(t, root, "rev-parse", "HEAD") != before || testfix.Capture(t, root, "status", "--porcelain") != "" {
		t.Fatal("tracker write changed caller checkout")
	}
	if log := testfix.Capture(t, root, "log", "-1", "--format=%B", "refs/remotes/publication/issue-tracker"); !strings.Contains(log, "Tracker-Operation: test-claim-one") {
		t.Fatalf("missing provenance: %s", log)
	}
}

func TestRepositoryReceiptFailurePreventsPublication(t *testing.T) {
	r, _, _ := fixture(t)
	snap, err := r.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	old, _ := snap.Card("000252")
	want := errors.New("receipt unavailable")
	err = r.UpdateCard(old, []byte(strings.Replace(testCard, "open", "working", 1)), "receipt-failure", func(string, string) error { return want })
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
	after, err := r.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if after.Ref() != snap.Ref() {
		t.Fatal("published without receipt")
	}
}

func TestRepositoryNoopDoesNotClaimOwnership(t *testing.T) {
	r, _, _ := fixture(t)
	snapshot, err := r.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	card, _ := snapshot.Card("000252")
	called := false
	err = r.UpdateCard(card, card.Raw, "noop-operation", func(string, string) error { called = true; return nil })
	if !errors.Is(err, ErrNoChange) || called {
		t.Fatalf("noop ownership: err=%v receipt=%v", err, called)
	}
}

func TestRepositoryRejectsUnreadableReplacement(t *testing.T) {
	r, _, _ := fixture(t)
	snapshot, err := r.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	card, _ := snapshot.Card("000252")
	receipt := false
	err = r.UpdateCard(card, []byte(testCard+strings.Repeat("x", 1<<20)), "oversize", func(string, string) error { receipt = true; return nil })
	if !errors.Is(err, gitx.ErrOutputLimit) || receipt {
		t.Errorf("oversize published: err=%v receipt=%v", err, receipt)
	}
	view, err := r.trunk.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if view.Ref() != snapshot.Ref() {
		t.Error("oversize update moved remote")
	}
}

func TestRepositoryHonorsCancellation(t *testing.T) {
	_, root, _ := fixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	r, err := NewRepository(ctx, root, "publication")
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := r.Snapshot(); !errors.Is(err, context.Canceled) {
		t.Fatalf("snapshot cancellation: %v", err)
	}
}

func TestRepositoryOfflineDoesNotUseCachedAuthority(t *testing.T) {
	r, root, _ := fixture(t)
	if _, err := r.Snapshot(); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, root, "remote", "set-url", "publication", filepath.Join(t.TempDir(), "missing.git"))
	if _, err := r.Snapshot(); err == nil {
		t.Fatal("offline snapshot silently used cached tracker")
	}
}

func TestRepositoryRevalidatesAfterPeerWrite(t *testing.T) {
	for _, sameCard := range []bool{false, true} {
		t.Run(map[bool]string{false: "other card", true: "same card"}[sameCard], func(t *testing.T) {
			r, _, peer := fixture(t)
			snapshot, err := r.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			old, _ := snapshot.Card("000252")
			calls := 0
			err = r.UpdateCard(old, []byte(strings.Replace(testCard, "open", "working", 1)), "race-op", func(string, string) error {
				calls++
				if calls != 1 {
					return nil
				}
				path, raw := "workshop/issue-cards/000253-peer.md", strings.ReplaceAll(testCard, "000252", "000253")
				if sameCard {
					path = testPath
					raw = strings.Replace(testCard, "Tracker title", "Peer title", 1)
				}
				return peer.UpdateMany("peer write", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
					return gitx.TrunkWrite{Write: map[string][]byte{path: []byte(raw)}}, nil
				})
			})
			if sameCard {
				if !errors.Is(err, ErrCardChanged) {
					t.Fatalf("same-card contention: %v", err)
				}
			} else if err != nil || calls != 2 {
				t.Fatalf("unrelated contention: calls=%d err=%v", calls, err)
			}
		})
	}
}
