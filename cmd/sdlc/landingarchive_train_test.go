package main

import (
	"context"
	"errors"
	"io"
	"path"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func TestTrainLandingArchiveProofControls(t *testing.T) {
	// Use the actual done-card sweep, whose archive deliberately has no
	// Landing-PR trailer. All negative candidates share this one landed close.
	r, cardPath, detailPath := closedAndLanded(t, 494)
	ctx := context.Background()
	env, err := openTrackerAt(ctx, r.root)
	if err != nil {
		t.Fatal(err)
	}
	pr := landingPR{
		Number: 494, State: "MERGED", Repo: "test/repo",
		HeadRef: "000494-e2e", HeadOID: r.git("rev-parse", "HEAD^2"),
		BaseRef: "main", BaseOID: r.git("rev-parse", "HEAD^1"),
		MergeOID: r.git("rev-parse", "HEAD"), LandedOID: r.git("rev-parse", "HEAD"),
	}
	if err := completeLandingPR(ctx, r.root, "workshop/issues", pr); err != nil {
		t.Fatal(err)
	}
	if err := finishLandedLeftovers(env, io.Discard, archiveDirs{
		Issues: "workshop/issues", Plans: "workshop/plans", History: "workshop/history",
	}); err != nil {
		t.Fatal(err)
	}
	r.git("fetch", "-q", env.target.Remote, "main")
	archived := r.originMain()
	destination := path.Join("workshop/history/issues", path.Base(detailPath))
	body := r.git("show", archived+":"+destination) + "\n"
	if trailers := r.git("show", "-s", "--format=%(trailers:key=Landing-PR,valueonly)", archived); trailers != "" {
		t.Fatalf("fixture unexpectedly has per-PR provenance: %s", trailers)
	}
	proof := func(t *testing.T, tip string, wantError string) {
		t.Helper()
		head, index := r.git("rev-parse", "HEAD"), r.git("ls-files", "--stage")
		remoteMain := r.originMain()
		done, err := confirmTrainLandingArchive(ctx, r.root, tip, pr.Repo, pr,
			"workshop/issues", "workshop/plans", "workshop/history")
		if wantError == "" && (err != nil || !done) {
			t.Fatalf("valid generic archive refused: complete=%v err=%v", done, err)
		}
		if wantError != "" && (err == nil || done || !strings.Contains(err.Error(), wantError)) {
			t.Fatalf("unsafe generic archive accepted: complete=%v err=%v", done, err)
		}
		if r.git("rev-parse", "HEAD") != head || r.git("ls-files", "--stage") != index || r.originMain() != remoteMain {
			t.Fatal("read-only archive proof changed HEAD, index, or remote main")
		}
	}
	t.Run("generic archive requires train proof", func(t *testing.T) {
		done, err := laProof(r.root, archived, pr)
		var missing *landingArchiveWithoutProvenance
		if done || !errors.As(err, &missing) {
			t.Fatalf("ordinary proof must reject generic archive with missing provenance: %v %v", done, err)
		}
		proof(t, archived, "")
	})
	t.Run("no archive", func(t *testing.T) { proof(t, pr.LandedOID, "requires one archive commit") })
	for _, kind := range []string{"missing destination", "tampered destination", "malformed destination", "active issue", "active sidecar", "ambiguous add history"} {
		t.Run(kind, func(t *testing.T) {
			r.git("checkout", "-q", "--detach", archived)
			switch kind {
			case "missing destination":
				r.git("rm", "-q", destination)
			case "tampered destination":
				laWrite(t, r.root, destination, body+"\nChanged archived details.\n")
			case "malformed destination":
				laWrite(t, r.root, destination, "---\nid: [broken\n---\n")
			case "active issue":
				laWrite(t, r.root, detailPath, r.git("show", pr.LandedOID+":"+detailPath)+"\n")
			case "active sidecar":
				laWrite(t, r.root, "workshop/plans/000494-extra.md", "residual plan\n")
			case "ambiguous add history":
				r.git("rm", "-q", destination)
				laCommit(t, r.root, "remove archive")
				laWrite(t, r.root, destination, body)
			}
			candidate := laCommit(t, r.root, kind)
			wantError := "archived generation changed"
			switch kind {
			case "active issue":
				wantError = "active archive source remains"
			case "active sidecar":
				wantError = "active or conflicting archive artifact"
			case "ambiguous add history":
				wantError = "requires one archive commit"
			}
			proof(t, candidate, wantError)
		})
	}
	t.Run("malformed archive transition", func(t *testing.T) {
		// Alter the introducing commit itself, rather than only checking later
		// corruption. The transition verifier must reject its invalid date.
		r.git("checkout", "-q", "--detach", archived)
		fm, _, err := issue.Parse(body)
		if err != nil {
			t.Fatal(err)
		}
		date, _ := issue.GetField(fm, "updated")
		if date == "" {
			t.Fatal("fixture has no archive date")
		}
		laWrite(t, r.root, destination, strings.Replace(body, "updated: "+date, "updated: invalid", 1))
		r.git("add", "--", destination)
		r.git("commit", "-q", "--amend", "--no-edit")
		proof(t, r.git("rev-parse", "HEAD"), "archive generation has invalid updated date")
	})
	t.Run("different completion token", func(t *testing.T) {
		r.git("checkout", "-q", "--detach", archived)
		original := []byte(r.card(cardPath))
		binding, ok, err := issue.CardCompletion(original)
		if err != nil || !ok {
			t.Fatalf("fixture completion: %v %v", ok, err)
		}
		binding.Token += "-other-close"
		changed, err := issue.SetCardCompletion(original, binding)
		if err != nil {
			t.Fatal(err)
		}
		setCard := func(raw []byte, token string) error {
			return env.repo.ChangeCard("000494", cardPath, "test completion", token,
				func([]byte) ([]byte, error) { return raw, nil })
		}
		if err := setCard(changed, "train-proof-other-token"); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := setCard(original, "train-proof-restore-token"); err != nil {
				t.Errorf("restore completion: %v", err)
			}
		})
		proof(t, archived, "not this close's done card")
	})
}

func TestTrainLandingArchiveProofRejectsUntrackedIssue(t *testing.T) {
	root, origin, pr := laFixture(t)
	if err := laArchive(root, pr); err != nil {
		t.Fatal(err)
	}
	tip := laGit(t, origin, "rev-parse", "main")
	if done, err := laProof(root, tip, pr); err != nil || !done {
		t.Fatalf("fixture ordinary archive proof: %v %v", done, err)
	}
	if done, err := confirmTrainLandingArchive(context.Background(), root, tip, pr.Repo, pr,
		"workshop/issues", "workshop/plans", "workshop/history"); err == nil || done {
		t.Fatalf("train proof accepted untracked issue: %v %v", done, err)
	}
}
