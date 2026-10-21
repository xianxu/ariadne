package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// The real ref push finishes, then a drop wins before PR creation. The
// durable record must distinguish this from an uncertain PR create reply.
type trainDropPushRunner struct {
	gitRunner
	afterPush func() error
}

func (r trainDropPushRunner) GitInDir(root string, args ...string) ([]byte, error) {
	out, err := r.gitRunner.GitInDir(root, args...)
	if err == nil && len(args) > 0 && args[0] == "push" && strings.Contains(strings.Join(args, " "), "refs/heads/landing-train/") && r.afterPush != nil {
		err = r.afterPush()
	}
	return out, err
}

func TestLandingDropAfterPushBeforeCreate(t *testing.T) {
	s := newTrainSetup(t, "train")
	s.blocker.Release()
	h, err := s.line.Enqueue(landingline.Ticket{Issue: 1, Branch: landingTestBranch})
	if err != nil {
		t.Fatal(err)
	}
	defer h.Release()
	if err := s.line.Wait(context.Background(), h); err != nil {
		t.Fatal(err)
	}
	car := landingline.Car{Seq: h.Ticket().Seq, Issue: 1, Branch: landingTestBranch, Head: s.heads[landingTestBranch], PR: 1}
	c := &conductor{ctx: context.Background(), r: landingTransportRunner{}, gh: s.gh, ci: s.gh, line: s.line, handle: h, id: car.Seq, self: car.Seq,
		t: landingTarget{Root: s.first, Remote: "upstream", Repo: "test/repo"}, f: landingFlags(), stderr: io.Discard, heads: []string{car.Head}}
	if err := c.beginTrainControl([]landingline.Car{car}); err != nil {
		t.Fatal(err)
	}
	tr := landingline.NewTrain(c.id)
	tr.Cars = []landingline.Car{car}
	tr.Phase = landingline.PhaseTesting
	control, err := s.line.UpdateTrain(h, tr)
	if err != nil {
		t.Fatal(err)
	}
	c.r = trainDropPushRunner{gitRunner: landingTransportRunner{}, afterPush: func() error {
		_, err := s.line.RequestTrainDrop(c.id, control.Members[0], "before PR create")
		return err
	}}
	if err := c.publish(1); !errors.Is(err, landingline.ErrTrainDropPending) {
		t.Fatalf("publication result %v", err)
	}
	if len(s.gh.created) != 0 {
		t.Fatalf("drop won before create, but PRs created %v", s.gh.created)
	}
	control, err = s.line.ReadTrainControl()
	if err != nil {
		t.Fatal(err)
	}
	if control.Cleanup.Publication.Started {
		t.Fatal("record claims PR creation began before its fenced boundary")
	}
	if err := closeDroppedTrain(context.Background(), s.gh, c.git, c.locked, control.Cleanup); err != nil {
		t.Fatal(err)
	}
	wantNoTrainBranch(t, s.gh.remote)
}
