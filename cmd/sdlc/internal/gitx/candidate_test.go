package gitx

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func trim(s string) string { return strings.TrimSpace(s) }

func TestCandidateStepsProveOwnershipByExactCommit(t *testing.T) {
	repo, origin := trunkFixture(t, "seed\n")
	head, index := trim(testfix.Capture(t, repo, "rev-parse", "HEAD")), trim(testfix.Capture(t, repo, "write-tree"))
	tf, _ := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	c, err := tf.PrepareCandidate("add card\n\nTracker-Operation: op-1", func(v *TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{"cards/1.md": []byte("one\n")}, ExactBytes: true}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Base != trim(testfix.Capture(t, origin, "rev-parse", "main")) {
		t.Fatal("candidate base is not the fetched tip")
	}
	if probe, _, err := tf.ProbeCandidate(c); err != nil || probe != ProbeAbsentSame {
		t.Fatalf("unpushed candidate probe = %v, %v", probe, err)
	}
	if out, err := tf.PushCandidate(c); err != nil || out != PushAccepted {
		t.Fatalf("push = %v, %v", out, err)
	}
	if probe, now, err := tf.ProbeCandidate(c); err != nil || probe != ProbeReachable || now != c.OID {
		t.Fatalf("pushed probe = %v %s, %v", probe, now, err)
	}
	// Re-pushing the identical candidate is idempotent: never a second commit.
	if out, err := tf.PushCandidate(c); err != nil || out != PushRejected {
		t.Fatalf("repush = %v, %v", out, err)
	}
	if trim(testfix.Capture(t, origin, "rev-parse", "main")) != c.OID {
		t.Fatal("identical repush changed the tip")
	}
	if blob, err := tf.BlobAt(c.OID, "cards/1.md"); err != nil || blob != trim(testfix.Capture(t, origin, "rev-parse", "main:cards/1.md")) {
		t.Fatalf("blob at candidate = %s, %v", blob, err)
	}
	if blob, err := tf.BlobAt(c.OID, "cards/absent.md"); err != nil || blob != "" {
		t.Fatalf("absent blob = %q, %v", blob, err)
	}
	if trim(testfix.Capture(t, repo, "rev-parse", "HEAD")) != head || trim(testfix.Capture(t, repo, "write-tree")) != index {
		t.Fatal("candidate steps changed the caller checkout")
	}
}

func TestCandidateRejectedByPeerIsProvenNotApplied(t *testing.T) {
	repo, origin := trunkFixture(t, "seed\n")
	tf, _ := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	c, err := tf.PrepareCandidate("mine\n\nTracker-Operation: op-2", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{"cards/1.md": []byte("mine\n")}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	peer := testfix.Repo(t, testfix.InitialCommit())
	testfix.Git(t, peer, "fetch", "-q", origin, "main")
	testfix.Git(t, peer, "reset", "-q", "--hard", "FETCH_HEAD")
	if err := os.WriteFile(filepath.Join(peer, "peer.md"), []byte("peer\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, peer, "add", "peer.md")
	testfix.Git(t, peer, "commit", "-qm", "peer")
	testfix.Git(t, peer, "push", "-q", origin, "HEAD:main")
	if out, err := tf.PushCandidate(c); err != nil || out != PushRejected {
		t.Fatalf("stale lease push = %v, %v", out, err)
	}
	if probe, now, err := tf.ProbeCandidate(c); err != nil || probe != ProbeAbsentMoved || now == c.Base {
		t.Fatalf("probe after peer = %v, %v", probe, err)
	}
	if _, err := tf.PrepareCandidate("noop\n\nTracker-Operation: op-3", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{"peer.md": []byte("peer\n")}}, nil
	}); !errors.Is(err, ErrCandidateNoChange) {
		t.Fatalf("no-op candidate: %v", err)
	}
}

func TestCandidateCancellationIsUnknownNotRejected(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	live, _ := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	c, err := live.PrepareCandidate("x\n\nTracker-Operation: op-4", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{"x.md": []byte("x\n")}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	dead, _ := NewTrunkFileContext(ctx, repo, "origin", "main")
	if out, err := dead.PushCandidate(c); out != PushUnknown || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled push = %v, %v", out, err)
	}
	if probe, _, err := dead.ProbeCandidate(c); probe != ProbeUnknown || err == nil {
		t.Fatalf("cancelled probe = %v, %v", probe, err)
	}
	for _, bad := range []Candidate{{}, {Base: "HEAD", OID: c.OID}, {Base: c.Base, OID: "main"}} {
		if out, err := live.PushCandidate(bad); out != PushUnknown || err == nil {
			t.Fatalf("accepted malformed candidate %+v", bad)
		}
	}
}
