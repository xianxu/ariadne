package main

import (
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Init and migration need every publication-branch category. These assertions
// exercise observation and conservative presence independently of CLI recovery
// exceptions for an already verified tracker.
func TestIssueInitObservesEveryPublicationRefCategory(t *testing.T) {
	categories := []struct{ name, ref string }{
		{"local-main", "refs/heads/main"},
		{"local-resting", "refs/heads/main-slot3"},
		{"local-tracker", "refs/heads/issue-tracker"},
		{"remote-main", "refs/remotes/origin/main"},
		{"remote-tracker", "refs/remotes/origin/issue-tracker"},
		{"remote-head-alias", "refs/remotes/origin/HEAD"},
	}
	for _, root := range []string{"workshop/issues", "workshop/history", "workshop/history/issues"} {
		for _, category := range categories {
			t.Run(root+"/"+category.name, func(t *testing.T) {
				r := emptyInitRepo(t)
				const seed = "temporary-ref-evidence-seed"
				issuePath := root + "/000001-category-evidence.md"
				r.git("switch", "-qc", seed)
				fields := "status: open\n"
				if root != "workshop/issues" {
					fields = "status: done\nactual_hours: 1\n"
				}
				writeRepoFile(t, r.root, issuePath, "---\nid: 000001\n"+fields+"created: 2026-10-21\n---\n\n# Preserved category evidence\n")
				r.git("add", issuePath)
				r.git("commit", "-qm", "legacy evidence on excluded ref category")
				tip := r.git("rev-parse", "HEAD")
				r.git("switch", "-q", "main")
				switch category.name {
				case "local-main":
					r.git("merge", "-q", "--ff-only", seed)
				case "local-resting":
					r.git("branch", "main-slot3", tip)
				case "local-tracker":
					r.git("branch", "issue-tracker", tip)
				case "remote-main", "remote-head-alias":
					r.git("push", "-q", "origin", seed+":main")
					if category.name == "remote-head-alias" {
						r.git("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
					}
				case "remote-tracker":
					r.git("push", "-q", "origin", seed+":issue-tracker")
				}
				r.git("branch", "-D", seed)
				if got := r.git("for-each-ref", "--format=%(refname)", "refs/heads/"+seed); got != "" {
					t.Fatalf("temporary seed could mask omission: %s", got)
				}
				env, err := openMigrate(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if category.name == "local-resting" {
					env.resting = "main-slot3"
				}
				main, err := env.main.Snapshot()
				if err != nil {
					t.Fatal(err)
				}
				beforeMain := r.originMain()
				beforeHead := r.git("rev-parse", "HEAD")
				beforeStatus := r.git("status", "--porcelain", "--untracked-files=all")
				beforeIndex := r.git("write-tree")
				beforeLocalRefs := r.git("for-each-ref", "--format=%(refname) %(objectname)", "refs/heads")
				beforeRemoteRefs := strings.TrimSpace(testfix.Capture(t, r.origin, "for-each-ref", "--format=%(refname) %(objectname)", "refs/heads"))
				observed, err := publicationBranchTips(env, main.Ref())
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for _, ref := range observed {
					if ref.Name == category.ref {
						found = true
						if ref.OID != tip {
							t.Errorf("%s pinned %s, want exact %s", ref.Name, ref.OID, tip)
						}
					}
					if ref.Name == "refs/heads/"+seed {
						t.Fatal("temporary seed survived fixture setup")
					}
				}
				if !found {
					t.Errorf("complete observer omitted %s from %#v", category.ref, observed)
				}
				if category.name == "remote-head-alias" {
					if got := r.git("symbolic-ref", category.ref); got != "refs/remotes/origin/main" {
						t.Fatalf("fixture lost symbolic alias: %s", got)
					}
				}
				if err = verifyInitializationEmpty(env, main.Ref()); err == nil || !strings.Contains(err.Error(), "issue migrate") {
					t.Errorf("presence gate did not refuse %s issue evidence: %v", category.ref, err)
				}
				// Both consumers use this complete observation before their own
				// presence/import policies; no category disappears in a wrapper.
				inventory, err := observeIssueInventory(env, main.Ref())
				if err != nil {
					t.Fatal(err)
				}
				included := false
				for _, ref := range inventory.Tips {
					if ref.Name == category.ref && ref.OID == tip {
						included = true
					}
				}
				if !included {
					t.Errorf("shared inventory omitted %s at %s", category.ref, tip)
				}
				if r.originMain() != beforeMain {
					t.Fatal("observation/presence gate changed published main")
				}
				assertCheckoutUntouched(t, r, beforeHead, beforeStatus)
				if r.git("write-tree") != beforeIndex {
					t.Fatal("observation/presence gate changed index")
				}
				if got := r.git("for-each-ref", "--format=%(refname) %(objectname)", "refs/heads"); got != beforeLocalRefs {
					t.Fatalf("observation changed local branch authority:\n%s\nwas:\n%s", got, beforeLocalRefs)
				}
				if got := strings.TrimSpace(testfix.Capture(t, r.origin, "for-each-ref", "--format=%(refname) %(objectname)", "refs/heads")); got != beforeRemoteRefs {
					t.Fatalf("observation changed remote branch authority:\n%s\nwas:\n%s", got, beforeRemoteRefs)
				}
			})
		}
	}
}
