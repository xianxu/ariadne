package gitx

import (
	"context"
	"testing"
)

func TestReviewUsageNamespaceIsolated(t *testing.T) {
	root, _ := trunkFixture(t, "source")
	ctx := context.Background()
	usage, err := NewRecoveryStoreFor(ctx, root, "review-usage")
	if err != nil {
		t.Fatal(err)
	}
	defaultStore, _ := NewRecoveryStore(ctx, root)
	correction, _ := NewRecoveryStoreFor(ctx, root, "actual-correction")
	if _, err := usage.Save("review-one", "", RecoveryData{Document: []byte(`{"version":1}`)}); err != nil {
		t.Fatal(err)
	}
	for _, store := range []*RecoveryStore{defaultStore, correction} {
		entries, err := store.List()
		if err != nil || len(entries) != 0 {
			t.Fatal("usage leaked into other namespace")
		}
	}
	entries, err := usage.List()
	if err != nil || len(entries) != 1 {
		t.Fatal("usage missing")
	}
}
