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

func TestRecoveryRefPinsObjectsWithoutChangingCaller(t *testing.T) {
	repo, _ := trunkFixture(t, "source")
	// The pinned source and candidate are deliberately unreachable elsewhere.
	testfix.Git(t, repo, "commit", "--allow-empty", "-qm", "candidate")
	candidate := strings.TrimSpace(testfix.Capture(t, repo, "rev-parse", "HEAD"))
	testfix.Git(t, repo, "reset", "--soft", "HEAD^")
	sourcePath := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(sourcePath, []byte("exact source\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	blob := strings.TrimSpace(testfix.Capture(t, repo, "hash-object", "-w", "--no-filters", sourcePath))
	before := testfix.Capture(t, repo, "rev-parse", "HEAD")
	index := testfix.Capture(t, repo, "write-tree")
	store, err := NewRecoveryStore(context.Background(), repo)
	if err != nil {
		t.Fatal(err)
	}
	doc := []byte("{\"operation\":\"test\"}\n")
	oid, err := store.Save("op-one", "", RecoveryData{Document: doc, Objects: []string{candidate, blob}})
	if err != nil {
		t.Fatal(err)
	}
	// Delete reflog reachability and prune; refs, not JSON strings, retain objects.
	testfix.Git(t, repo, "reflog", "expire", "--expire=now", "--all")
	testfix.Git(t, repo, "gc", "--prune=now")
	for _, obj := range []string{candidate, blob} {
		testfix.Git(t, repo, "cat-file", "-e", obj)
	}
	got, err := store.Load("op-one")
	if err != nil || got.OID != oid || string(got.Data.Document) != string(doc) || len(got.Data.Objects) != 2 {
		t.Fatalf("recovery read %+v: %v", got, err)
	}
	entries, err := store.List()
	if err != nil || len(entries) != 1 || entries[0].Token != "op-one" || entries[0].OID != oid {
		t.Fatalf("list %+v: %v", entries, err)
	}
	if testfix.Capture(t, repo, "rev-parse", "HEAD") != before || testfix.Capture(t, repo, "write-tree") != index {
		t.Fatal("changed caller")
	}
}

func TestRecoveryRefCASAndScopedDelete(t *testing.T) {
	repo, _ := trunkFixture(t, "source")
	store, _ := NewRecoveryStore(context.Background(), repo)
	first, err := store.Save("op", "", RecoveryData{Document: []byte("first")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save("op", "", RecoveryData{Document: []byte("other")}); !errors.Is(err, ErrRecoveryChanged) {
		t.Fatalf("creation overwrote: %v", err)
	}
	second, err := store.Save("op", first, RecoveryData{Document: []byte("second")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save("op", first, RecoveryData{Document: []byte("stale")}); !errors.Is(err, ErrRecoveryChanged) {
		t.Fatalf("stale update: %v", err)
	}
	if err := store.Delete("op", first); !errors.Is(err, ErrRecoveryChanged) {
		t.Fatalf("stale delete: %v", err)
	}
	if err := store.Delete("op", ""); err == nil {
		t.Fatal("unconditional delete")
	}
	if err := store.Delete("op", second); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load("op"); !errors.Is(err, ErrRecoveryAbsent) {
		t.Fatalf("deleted load: %v", err)
	}
}

func TestRecoveryRefInputSafetyAndCancellation(t *testing.T) {
	repo, _ := trunkFixture(t, "source")
	store, _ := NewRecoveryStore(context.Background(), repo)
	for _, token := range []string{"", "../main", "op/ref", "op\n", strings.Repeat("a", 129)} {
		if _, err := store.Save(token, "", RecoveryData{Document: []byte("x")}); err == nil {
			t.Fatalf("accepted token %q", token)
		}
	}
	for _, data := range []RecoveryData{{}, {Document: make([]byte, SnapshotBlobLimit+1)}, {Document: []byte("x"), Objects: []string{"HEAD"}}, {Document: []byte("x"), Objects: []string{strings.Repeat("1", 40)}}} {
		if _, err := store.Save("op", "", data); err == nil {
			t.Fatalf("accepted bad recovery data")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cancelled, _ := NewRecoveryStore(ctx, repo)
	if _, err := cancelled.Save("op", "", RecoveryData{Document: []byte("x")}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}
