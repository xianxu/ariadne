package gitx

import (
	"errors"
	"testing"
)

func TestTrackerPublicationStatefulSchedules(t *testing.T) {
	for _, schedule := range []string{"normal", "other-card", "same-card", "lost-ack", "receipt-failure"} {
		t.Run(schedule, func(t *testing.T) {
			f := newPublicationGit(t)
			main := f.commit("", map[string]string{"code.go": "main code"}, "main")
			f.remote = main
			base := f.commit("", map[string]string{"card.md": "open"}, "tracker")
			f.setRemote("refs/heads/issue-tracker", base)
			original := runGitInContext
			runGitInContext = f.run
			t.Cleanup(func() { runGitInContext = original })
			tf, err := NewTrunkFile("stateful-fake", "origin", "issue-tracker")
			if err != nil {
				t.Fatal(err)
			}
			receipts := map[string]bool{}
			f.beforePush = func(f *publicationGit) {
				if !receipts[f.pushCandidate] {
					t.Fatal("remote mutation attempted before candidate receipt")
				}
				if f.pushes != 1 {
					return
				}
				if schedule == "other-card" || schedule == "same-card" {
					files := copyFiles(f.files(f.remoteTip("refs/heads/issue-tracker")))
					if schedule == "other-card" {
						files["peer.md"] = "new"
					} else {
						files["card.md"] = "peer claim"
					}
					f.setRemote("refs/heads/issue-tracker", f.commit(base, files, "peer"))
				}
			}
			f.loseAck = schedule == "lost-ack"
			conflict := errors.New("card changed")
			receiptFailure := errors.New("receipt unavailable")
			err = tf.UpdateManyPrepared("claim\n\nTracker-Operation: unique-claim", func(v *TrunkView) (TrunkWrite, error) {
				raw, err := v.Read("card.md")
				if err != nil {
					return TrunkWrite{}, err
				}
				if string(raw) != "open" {
					return TrunkWrite{}, conflict
				}
				return TrunkWrite{Write: map[string][]byte{"card.md": []byte("working")}}, nil
			}, func(base, candidate string) error {
				if base == "" || candidate == "" {
					t.Fatal("receipt missing immutable identities")
				}
				if schedule == "receipt-failure" {
					return receiptFailure
				}
				receipts[candidate] = true
				return nil
			})
			if f.remote != main {
				t.Fatal("tracker operation changed main")
			}
			files := f.files(f.remoteTip("refs/heads/issue-tracker"))
			switch schedule {
			case "receipt-failure":
				if !errors.Is(err, receiptFailure) || f.pushes != 0 || files["card.md"] != "open" {
					t.Fatalf("receipt failure: pushes=%d err=%v files=%v", f.pushes, err, files)
				}
			case "same-card":
				if !errors.Is(err, conflict) || f.pushes != 1 || files["card.md"] != "peer claim" {
					t.Fatalf("peer overwritten: pushes=%d err=%v files=%v", f.pushes, err, files)
				}
			case "lost-ack":
				if !errors.Is(err, ErrPublicationUncertain) || f.pushes != 1 || files["card.md"] != "working" {
					t.Fatalf("uncertain write: pushes=%d err=%v files=%v", f.pushes, err, files)
				}
			default:
				if err != nil || files["card.md"] != "working" {
					t.Fatalf("write: err=%v files=%v", err, files)
				}
				if schedule == "other-card" && (f.pushes != 2 || files["peer.md"] != "new" || len(receipts) != 2) {
					t.Fatalf("retry lost peer/receipt: pushes=%d files=%v receipts=%v", f.pushes, files, receipts)
				}
			}
		})
	}
}

func TestTrackerBootstrapStatefulSchedules(t *testing.T) {
	for _, schedule := range []string{"normal", "already-exists", "competing-bootstrap", "lost-ack", "receipt-failure"} {
		t.Run(schedule, func(t *testing.T) {
			f := newPublicationGit(t)
			main := f.commit("", map[string]string{"code.go": "main code"}, "main")
			f.remote = main
			peer := f.commit("", map[string]string{"manifest": "peer"}, "peer bootstrap")
			if schedule == "already-exists" {
				f.setRemote("refs/heads/issue-tracker", peer)
			}
			original := runGitInContext
			runGitInContext = f.run
			t.Cleanup(func() { runGitInContext = original })
			tf, err := NewTrunkFile("stateful-fake", "origin", "issue-tracker")
			if err != nil {
				t.Fatal(err)
			}
			receipt := ""
			f.beforePush = func(f *publicationGit) {
				if receipt == "" || receipt != f.pushCandidate {
					t.Fatal("bootstrap push preceded receipt")
				}
				if schedule == "competing-bootstrap" {
					f.setRemote("refs/heads/issue-tracker", peer)
				}
			}
			f.loseAck = schedule == "lost-ack"
			failure := errors.New("receipt unavailable")
			result, err := tf.Bootstrap(map[string][]byte{"manifest": []byte("ours")}, "Tracker-Operation: unique-bootstrap", func(r BootstrapResult) error {
				if schedule == "receipt-failure" {
					return failure
				}
				receipt = r.Candidate
				return nil
			})
			if f.remote != main {
				t.Fatal("bootstrap changed main")
			}
			tip := f.remoteTip("refs/heads/issue-tracker")
			switch schedule {
			case "normal":
				if err != nil || result.Outcome != BootstrapCreated || tip != receipt || f.commits[tip].parent != "" {
					t.Fatalf("bootstrap: %+v %v", result, err)
				}
			case "already-exists":
				if !errors.Is(err, ErrBootstrapExists) || result.Outcome != BootstrapExisting || f.pushes != 0 || receipt != "" {
					t.Fatalf("existing: %+v %v", result, err)
				}
			case "competing-bootstrap":
				if !errors.Is(err, ErrBootstrapRejected) || result.Outcome != BootstrapRejected || tip != peer || f.pushes != 1 {
					t.Fatalf("race: %+v %v", result, err)
				}
			case "lost-ack":
				if !errors.Is(err, ErrPublicationUncertain) || result.Outcome != BootstrapUncertain || tip != receipt || f.pushes != 1 {
					t.Fatalf("lost ack: %+v %v", result, err)
				}
			case "receipt-failure":
				if !errors.Is(err, failure) || result.Outcome != BootstrapNotPublished || tip != "" || f.pushes != 0 {
					t.Fatalf("receipt: %+v %v", result, err)
				}
			}
		})
	}
}
