package gitx

import (
	"context"
	"errors"
	"testing"
)

func TestRecoveryNamespaceIsolation(t *testing.T) {
	root, _ := trunkFixture(t, "source")
	ordinary, _ := NewRecoveryStore(context.Background(), root)
	correction, err := NewRecoveryStoreFor(context.Background(), root, "actual-correction")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []*RecoveryStore{ordinary, correction} {
		if _, err := s.Save("same-token", "", RecoveryData{Document: []byte("record")}); err != nil {
			t.Fatal(err)
		}
	}
	old, _ := ordinary.List()
	newer, _ := correction.List()
	if len(old) != 1 || len(newer) != 1 {
		t.Fatalf("namespace listing not isolated: %v %v", old, newer)
	}
	if err := correction.Delete(newer[0].Token, newer[0].OID); err != nil {
		t.Fatal(err)
	}
	if _, err := ordinary.Load("same-token"); err != nil {
		t.Fatal(err)
	}
	if _, err := correction.Load("same-token"); !errors.Is(err, ErrRecoveryAbsent) {
		t.Fatal(err)
	}
	if _, err := NewRecoveryStoreFor(context.Background(), root, "arbitrary"); err == nil {
		t.Fatal("accepted arbitrary namespace")
	}
}
