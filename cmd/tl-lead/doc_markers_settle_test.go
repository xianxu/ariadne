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

func TestWorkingdocSettlePreservesLiteralExamples(t *testing.T) {
	for _, tc := range workingdocLiteralExamples() {
		t.Run(tc.name, func(t *testing.T) {
			source, ss, repo, d := newWorkingdocFixture(t)
			before, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			before = append(before, []byte("\n"+tc.body+"\n[live](../README.md)\n")...)
			if err := os.WriteFile(source, before, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(repo.root, "README.md"), []byte("Repository"), 0600); err != nil {
				t.Fatal(err)
			}
			result, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, "")
			if err != nil || !result.Published || !result.Stamped {
				t.Fatalf("settle: %+v %v", result, err)
			}
			want := bytes.ReplaceAll(before, []byte("[live](../README.md)"), []byte("[live](../../README.md)"))
			if !bytes.Equal(repo.current[result.Target], want) {
				t.Fatalf("published code changed: got %q want %q", repo.current[result.Target], want)
			}
			copy, err := os.ReadFile(filepath.Join(repo.root, result.Target))
			if err != nil || !bytes.Equal(copy, want) {
				t.Fatalf("prepared copy differs: %q %v", copy, err)
			}
			stamped, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			meta, err := parseWorkingdoc(string(stamped))
			if err != nil || meta.Pointer == nil {
				t.Fatalf("stamp: %v", err)
			}
			if string(stamped[:meta.pointerStart])+string(stamped[meta.pointerEnd:]) != string(before) || ss.stamps != 1 || repo.publishes != 1 {
				t.Fatal("source body changed or effects duplicated")
			}
			if _, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, ""); err != nil || repo.publishes != 1 {
				t.Fatalf("identical retry: %v publishes=%d", err, repo.publishes)
			}
		})
	}
}
