package main

import (
	"context"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetBodyOnlyMissingReferenceRefused(t *testing.T) {
	for _, mode := range []string{"legacy-guard", "legacy-no-upstream", "legacy-archive", "legacy-publish", "tracker-guard", "tracker-publish", "first-publish", "abandon"} {
		t.Run(mode, func(t *testing.T) {
			var r *trackerRepo
			id, detail := "000001", "workshop/issues/000001-one.md"
			switch mode {
			case "tracker-guard", "tracker-publish":
				r, _ = claimSetRepo(t)
				id, detail = "000009", "workshop/issues/000009-s09.md"
			case "first-publish":
				r = spinOffOnCodeBranch(t)
				id, detail = "000008", spinOffDetails
			case "abandon":
				r, _, _ = startedHere(t)
				id, detail = "000009", handoffDetail
			default:
				r = legacyRepo(t)
			}
			if mode == "legacy-archive" {
				id, detail = "000003", "workshop/history/issues/000003-gone.md"
			}
			appendDetail(t, r, detail, "![missing](assets/"+id+"/"+strings.Repeat("a", 64)+".png)")
			r.git("add", detail)
			r.git("commit", "-qm", "body-only asset link")
			if _, err := os.Stat(filepath.Join(r.root, filepath.Dir(detail), "assets")); !os.IsNotExist(err) {
				t.Fatalf("fixture unexpectedly has assets: %v", err)
			}
			before := r.originMain()
			if mode == "legacy-no-upstream" {
				r.git("branch", "--unset-upstream", "main")
			}
			var err error
			switch mode {
			case "legacy-publish":
				err = publishIssueCommit(context.Background(), io.Discard, io.Discard, r.root, "HEAD", "workshop/issues", "workshop/history")
			case "tracker-publish":
				_, err = publish(t, 9)
			case "first-publish":
				_, err = publish(t, 8)
			case "abandon":
				err = abandon9("punt", "later")
			default:
				err = guardTransferredDetails(context.Background())
			}
			if err == nil || !strings.Contains(err.Error(), "missing linked asset") {
				t.Fatalf("body-only missing reference escaped %s: %v", mode, err)
			}
			if r.originMain() != before {
				t.Fatal("missing asset publication changed main")
			}
		})
	}
}

func TestAssetBodyOnlyReferenceRefusedByLandingEntryPoints(t *testing.T) {
	r := legacyRepo(t)
	r.git("switch", "-q", "-c", "evidence")
	appendDetail(t, r, "workshop/issues/000001-one.md", "![missing](assets/000001/"+strings.Repeat("a", 64)+".png)")
	r.git("commit", "-qam", "body-only missing evidence")
	head, base := r.git("rev-parse", "HEAD"), r.git("rev-parse", "origin/main")
	checks := map[string]func() error{
		"pr": func() error {
			return runPR(io.Discard, io.Discard, &prFlags{IssuesDir: "workshop/issues", DryRun: true})
		},
		"push": func() error { return runPublishGate(context.Background(), base, "workshop/issues", io.Discard) },
		"merge": func() error {
			return runLandingPublishGate(context.Background(), landingPR{HeadOID: head, BaseOID: base}, "workshop/issues", io.Discard)
		},
	}
	for name, check := range checks {
		t.Run(name, func(t *testing.T) {
			if err := check(); err == nil || !strings.Contains(err.Error(), "missing linked asset") {
				t.Fatalf("%s accepted body-only reference: %v", name, err)
			}
		})
	}
}

func TestAssetBodyRemovalJudgedAgainstNewMainAssets(t *testing.T) {
	r := legacyRepo(t)
	detail := "workshop/issues/000001-one.md"
	original := readRepoFileOr(t, r.root, detail)
	name, _ := issueassets.Name([]byte("evidence"), "png")
	appendDetail(t, r, detail, "![evidence](assets/000001/"+name+")")
	r.git("commit", "-qam", "historical body reference before asset publication")
	r.git("push", "-q", "origin", "HEAD:main")
	r.git("switch", "-q", "-c", "remove-reference")
	writeRepoFile(t, r.root, detail, original)
	r.git("commit", "-qam", "remove body reference")
	peerAdd(t, r, "workshop/issues/assets/000001/"+name, "evidence")
	if err := guardTransferredDetails(context.Background()); err == nil || !strings.Contains(err.Error(), "not linked from the body") {
		t.Fatalf("main asset orphaned by branch body edit: %v", err)
	}
}

func TestAssetBodyIndexFallbackWithoutPublicationTarget(t *testing.T) {
	for _, mode := range []string{"text-only", "external", "example", "invalid-reference", "valid-bundle"} {
		t.Run(mode, func(t *testing.T) {
			r := legacyRepo(t)
			detail := "workshop/issues/000001-one.md"
			switch mode {
			case "external":
				appendDetail(t, r, detail, "![external](https://example.com/assets/photo.png)")
			case "example":
				appendDetail(t, r, detail, "`![example](assets/000001/no-file.png)`")
			case "invalid-reference":
				appendDetail(t, r, detail, "![invalid](assets/000001/no-file.png)")
			case "valid-bundle":
				assetFixture(t, r, "000001", detail)
			}
			if mode != "text-only" {
				r.git("add", "workshop/issues")
				r.git("commit", "-qm", "body fixture")
			}
			r.git("remote", "remove", "origin")
			err := guardTransferredDetails(context.Background())
			switch mode {
			case "invalid-reference":
				if err == nil || !strings.Contains(err.Error(), "invalid asset link") {
					t.Fatalf("invalid reference skipped: %v", err)
				}
			case "valid-bundle":
				if err == nil || !strings.Contains(err.Error(), "configure and fetch") {
					t.Fatalf("missing target silently accepted: %v", err)
				}
			default:
				if err != nil {
					t.Fatalf("text-only legacy work needs no publication target: %v", err)
				}
			}
		})
	}
}

func TestAssetBodyOnlyIssueCreationRefusedBeforeEffects(t *testing.T) {
	for _, tracked := range []bool{false, true} {
		for _, invalid := range []bool{false, true} {
			label := "legacy"
			if tracked {
				label = "tracker"
			}
			if invalid {
				label += "-invalid"
			}
			t.Run(label, func(t *testing.T) {
				var r *trackerRepo
				id := "000004"
				if tracked {
					r = newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
					id = "000008"
				} else {
					r = legacyRepo(t)
				}
				name := strings.Repeat("a", 64) + ".png"
				want := "missing linked asset"
				if invalid {
					name = "not-canonical.png"
					want = "invalid asset link"
				}
				previousGH, previousRepo := ghClient, detectRepo
				ghClient = stubGH{title: "Imported evidence", body: "![missing](assets/" + id + "/" + name + ")"}
				detectRepo = func() (string, error) { return "example/issues", nil }
				t.Cleanup(func() { ghClient, detectRepo = previousGH, previousRepo })
				refs := r.git("ls-remote", "origin", "refs/heads/main", "refs/heads/issue-tracker")
				head, status := r.git("rev-parse", "HEAD"), r.git("status", "--porcelain", "--untracked-files=all")
				flags := newIssueFlags()
				flags.FromGitHub = 99
				err := runIssueNew(context.Background(), io.Discard, io.Discard, flags, nil)
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("creation accepted missing/invalid body-indexed asset: %v", err)
				}
				if got := r.git("ls-remote", "origin", "refs/heads/main", "refs/heads/issue-tracker"); got != refs {
					t.Fatal("invalid creation changed card or main")
				}
				assertCheckoutUntouched(t, r, head, status)
			})
		}
	}
}

func TestAssetBodyOnlyLegacyReservationRefused(t *testing.T) {
	r := legacyRepo(t)
	p := "workshop/issues/000004-new.md"
	writeRepoFile(t, r.root, p, "---\nid: 000004\nstatus: open\ncreated: 2026-10-13\n---\n\n![missing](assets/000004/"+strings.Repeat("a", 64)+".png)\n")
	before := r.originMain()
	f := &claimFlags{Issue: 4, IssuesDir: "workshop/issues", HistoryDir: "workshop/history", FirstPublication: true}
	if err := syncIssuesToMain(io.Discard, io.Discard, f, execGitRunner{}, "new issue"); err == nil || !strings.Contains(err.Error(), "missing linked asset") {
		t.Fatalf("reservation skipped body index: %v", err)
	}
	if r.originMain() != before {
		t.Fatal("invalid reservation changed main")
	}
}

func TestAssetBodyIssueCreationAcceptsExternalImagesAndCodeExamples(t *testing.T) {
	for _, tracked := range []bool{false, true} {
		name := "legacy"
		if tracked {
			name = "tracker"
		}
		t.Run(name, func(t *testing.T) {
			if tracked {
				newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
			} else {
				legacyRepo(t)
			}
			previousGH, previousRepo := ghClient, detectRepo
			ghClient = stubGH{title: "External evidence", body: "![external](https://example.com/assets/photo.png)\n\n`![example](assets/000001/no-file.png)`"}
			detectRepo = func() (string, error) { return "example/issues", nil }
			t.Cleanup(func() { ghClient, detectRepo = previousGH, previousRepo })
			flags := newIssueFlags()
			flags.FromGitHub = 99
			if err := runIssueNew(context.Background(), io.Discard, io.Discard, flags, nil); err != nil {
				t.Fatalf("ordinary imported body refused: %v", err)
			}
		})
	}
}
