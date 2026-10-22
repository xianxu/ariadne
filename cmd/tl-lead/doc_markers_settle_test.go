package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkingdocBlockBoundaryMarkersRefuseBeforeEffects(t *testing.T) {
	for _, tc := range docMarkerBlockBoundaries() {
		t.Run(tc.name, func(t *testing.T) {
			source, ss, repo, d := newWorkingdocFixture(t)
			before, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			before = append(before, []byte(tc.body)...)
			if err := os.WriteFile(source, before, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, ""); err == nil {
				t.Fatal("accepted unresolved commentary outside actual code")
			}
			after, err := os.ReadFile(source)
			if err != nil || !bytes.Equal(before, after) || ss.stamps != 0 || repo.publishes != 0 {
				t.Fatalf("refusal changed source or published: stamps=%d publishes=%d error=%v", ss.stamps, repo.publishes, err)
			}
			if _, err := os.Stat(filepath.Join(repo.root, "workshop")); !os.IsNotExist(err) {
				t.Fatalf("refusal created destination tree: %v", err)
			}
		})
	}
}
