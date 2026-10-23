package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

const (
	renameTitle = "Two long lua test names joined by a filer into one very long title that keeps going past any sensible length"
	renameSlug  = "two-long-lua-test-names-joined-by-a-filer-into-one-very-long-title-that-keeps"
)

// fileForRename files #8 with the given explicit slug as workspace slot.
func fileForRename(t *testing.T, slot, slug string) *trackerRepo {
	t.Helper()
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	slotClaimant(t, slot)
	f := newIssueFlags()
	f.Slug = slug
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, f, []string{renameTitle}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	return r
}

func (r *trackerRepo) tip(ref string) string {
	return strings.TrimSpace(testfix.Capture(r.t, r.origin, "rev-parse", ref))
}

// assertRenamed checks both refs hold only the new name, with the old bytes.
func assertRenamed(t *testing.T, r *trackerRepo, oldStem, newStem, card, detail string) {
	t.Helper()
	if got := r.card("workshop/issue-cards/" + newStem + ".md"); got != card {
		t.Fatalf("card at the new name:\n%s\nwant\n%s", got, card)
	}
	if got := r.mainFile("workshop/issues/" + newStem + ".md"); got != detail {
		t.Fatalf("details at the new name differ")
	}
	if r.card("workshop/issue-cards/"+oldStem+".md") != "" || r.mainFile("workshop/issues/"+oldStem+".md") != "" {
		t.Fatal("the old name survived")
	}
}

func TestIssueRenameMovesCardAndDetails(t *testing.T) {
	r := fileForRename(t, "ariadne:0", "first-slug")
	card, detail := r.card("workshop/issue-cards/000008-first-slug.md"), r.mainFile("workshop/issues/000008-first-slug.md")
	var out, errs bytes.Buffer
	if err := runIssueRename(context.Background(), &out, &errs, "workshop/issues", 8, "short"); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if strings.TrimSpace(out.String()) != "workshop/issues/000008-short.md" {
		t.Fatalf("stdout %q", out.String())
	}
	assertRenamed(t, r, "000008-first-slug", "000008-short", card, detail)
	// The mirror is the card's blob OID, unchanged by a path move.
	if _, err := issue.ParseCard([]byte(r.card("workshop/issue-cards/000008-short.md"))); err != nil {
		t.Fatal(err)
	}
}

// Without --slug the capped slug comes from the card's title: the one-command
// repair for a filer-made name (#566, parley.nvim#421).
func TestIssueRenameDerivesCappedSlugFromTitle(t *testing.T) {
	r := fileForRename(t, "ariadne:0", "x")
	card, detail := r.card("workshop/issue-cards/000008-x.md"), r.mainFile("workshop/issues/000008-x.md")
	var out, errs bytes.Buffer
	if err := runIssueRename(context.Background(), &out, &errs, "workshop/issues", 8, ""); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	assertRenamed(t, r, "000008-x", "000008-"+renameSlug, card, detail)
}

// A run that stopped after the card step finishes on rerun; a finished rename
// is a no-op that commits nothing.
func TestIssueRenameResumesAndIsIdempotent(t *testing.T) {
	r := fileForRename(t, "ariadne:0", "first-slug")
	card, detail := r.card("workshop/issue-cards/000008-first-slug.md"), r.mainFile("workshop/issues/000008-first-slug.md")
	if err := r.tracker.UpdateMany("half", tracker.MovePath("workshop/issue-cards/000008-first-slug.md", "workshop/issue-cards/000008-short.md")); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		var out, errs bytes.Buffer
		if err := runIssueRename(context.Background(), &out, &errs, "workshop/issues", 8, "short"); err != nil {
			t.Fatalf("run %d: %v\n%s", i, err, errs.String())
		}
	}
	assertRenamed(t, r, "000008-first-slug", "000008-short", card, detail)
	before := [2]string{r.tip("issue-tracker"), r.tip("main")}
	var out, errs bytes.Buffer
	if err := runIssueRename(context.Background(), &out, &errs, "workshop/issues", 8, "short"); err != nil {
		t.Fatal(err)
	}
	if after := [2]string{r.tip("issue-tracker"), r.tip("main")}; after != before {
		t.Fatal("a finished rename committed again")
	}
}

func TestIssueRenameRefusals(t *testing.T) {
	cases := []struct {
		name, want, slug string
		setup            func(t *testing.T, r *trackerRepo)
	}{
		{"started", "rename it before start-plan", "short", func(t *testing.T, r *trackerRepo) {
			setCardField(t, r, "workshop/issue-cards/000008-first-slug.md", "status: open", "status: working")
		}},
		{"branch on origin", "has an issue branch", "short", func(t *testing.T, r *trackerRepo) {
			testfix.Git(t, r.root, "push", "-q", "origin", "HEAD:refs/heads/000008-first-slug")
		}},
		{"another owner", "only its owner", "short", func(t *testing.T, r *trackerRepo) { slotClaimant(t, "pair:2") }},
		{"overlong slug", "at most 80 bytes", strings.Repeat("a", 81), nil},
		{"name taken", "exactly one of the old and new names", "short", func(t *testing.T, r *trackerRepo) {
			main, err := gitx.NewTrunkFile(r.root, "origin", "main")
			if err != nil {
				t.Fatal(err)
			}
			if err := main.UpdateMany("squat", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
				return gitx.TrunkWrite{Write: map[string][]byte{"workshop/issues/000008-short.md": []byte("x\n")}}, nil
			}); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := fileForRename(t, "pair:1", "first-slug")
			if c.setup != nil {
				c.setup(t, r)
			}
			before := [2]string{r.tip("issue-tracker"), r.tip("main")}
			var out, errs bytes.Buffer
			err := runIssueRename(context.Background(), &out, &errs, "workshop/issues", 8, c.slug)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
			if after := [2]string{r.tip("issue-tracker"), r.tip("main")}; after != before {
				t.Fatal("a refused rename moved a ref")
			}
		})
	}
}

// setCardField rewrites one frontmatter line of a card on origin's tracker.
func setCardField(t *testing.T, r *trackerRepo, cardPath, from, to string) {
	t.Helper()
	if err := r.tracker.UpdateMany("edit", func(v *gitx.TrunkView) (gitx.TrunkWrite, error) {
		raw, err := v.Read(cardPath)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		return gitx.TrunkWrite{Write: map[string][]byte{cardPath: bytes.Replace(raw, []byte(from), []byte(to), 1)}, ExactBytes: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
}
