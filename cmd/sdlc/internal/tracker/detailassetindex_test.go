package tracker

import (
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

func TestCreationRecoveryRejectsMissingBodyIndexedAssetBeforeCard(t *testing.T) {
	op, receipts, root, _, main := creationFixture(t)
	r, err := op.Start("old-prepared-generation")
	if err != nil {
		t.Fatal(err)
	}
	spec := r.Spec()
	detail, err := gitx.ReadBlob(context.Background(), root, spec.SourceBlob)
	if err != nil {
		t.Fatal(err)
	}
	detail = append(detail, []byte("\n![missing](assets/"+spec.IssueID+"/"+strings.Repeat("a", 64)+".png)\n")...)
	spec.SourceBlob, err = gitx.WriteBlob(context.Background(), root, detail)
	if err != nil {
		t.Fatal(err)
	}
	old, err := NewCreation(spec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Drive(old.Receipt(), CreationStepper, op, receipts); err == nil || !strings.Contains(err.Error(), "missing linked asset") {
		t.Fatalf("persisted generation skipped validation: %v", err)
	}
	if len(trackerCards(t, op)) != 1 {
		t.Fatal("invalid generation reserved its card")
	}
	if onMain(t, main, spec.IssueID) != "" {
		t.Fatal("invalid generation published details")
	}
	if _, err := DetailPublication(main, spec.DestinationPath, detail, "bad first details"); err == nil || !strings.Contains(err.Error(), "missing linked asset") {
		t.Fatalf("first-detail adapter skipped body validation: %v", err)
	}
}
