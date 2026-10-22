package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkingdocStatusBatchesBaselines(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			source, _, repo, _ := newWorkingdocFixture(t)
			os.Remove(source)
			for i := 0; i < n; i++ {
				target := fmt.Sprintf("workshop/workingdoc/topic-%d.md", i)
				sha := strings.Repeat("b", 40)
				clean := fmt.Sprintf("---\ntype: workingdoc\n---\nTopic %d\n", i)
				stamped, err := stampWorkingdoc(clean, settlementPointer{target, sha})
				if err != nil {
					t.Fatal(err)
				}
				os.WriteFile(filepath.Join(repo.root, fmt.Sprintf("session.side/%02d.md", i)), []byte(stamped), 0600)
				os.MkdirAll(filepath.Join(repo.root, "workshop/workingdoc"), 0700)
				os.WriteFile(filepath.Join(repo.root, target), []byte(clean), 0600)
				repo.blobs[settlementPointer{target, sha}] = []byte(clean)
			}
			reads := 0
			rows, err := readWorkingdocStatus(context.Background(), repo, repo.root, func(root, rel string) ([]byte, bool, error) { reads++; return readWorkingdoc(root, rel) })
			if err != nil || len(rows) != n {
				t.Fatalf("rows %d err %v", len(rows), err)
			}
			for _, r := range rows {
				if r.State != "settled" {
					t.Fatalf("row %+v", r)
				}
			}
			if reads != 2*n || repo.baselineReads != 1 {
				t.Fatalf("N=%d file reads=%d baseline calls=%d", n, reads, repo.baselineReads)
			}
		})
	}
}

func TestWorkingdocStatusDoesNotHideUnknownOrMarkers(t *testing.T) {
	source, _, repo, d := newWorkingdocFixture(t)
	result, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, "")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(source)
	os.WriteFile(source, append(b, []byte("🤖[open]\n")...), 0600)
	rows, err := readWorkingdocStatus(context.Background(), repo, repo.root, readWorkingdoc)
	if err != nil || len(rows) != 1 || rows[0].State != "drift" {
		t.Fatalf("rows %+v err %v", rows, err)
	}
	delete(repo.blobs, settlementPointer{result.Target, result.Commit})
	rows, err = readWorkingdocStatus(context.Background(), repo, repo.root, readWorkingdoc)
	if err == nil || rows[0].State != "unverifiable" {
		t.Fatalf("unknown became success: %+v %v", rows, err)
	}
}
