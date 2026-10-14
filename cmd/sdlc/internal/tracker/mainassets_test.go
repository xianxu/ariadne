package tracker

import (
	"context"
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainCandidateReplayBindsBodyAndCompleteBundle(t *testing.T) {
	for _, operation := range []string{"creation", "transfer"} {
		t.Run(operation, func(t *testing.T) {
			creation, _, root, _, main := creationFixture(t)
			receipt, err := creation.Start("candidate-proof")
			if err != nil {
				t.Fatal(err)
			}
			spec := receipt.Spec()
			raw, err := main.ViewOf(spec.MainBase).ReadBlob(spec.SourceBlob)
			if err != nil {
				t.Fatal(err)
			}
			data := []byte("exact evidence\x00\xff")
			name, _ := issueassets.Name(data, "png")
			rel := path.Join("assets", spec.IssueID, name)
			asset := path.Join(path.Dir(spec.DestinationPath), rel)
			for _, variant := range []string{"changed-body", "orphan", "missing-linked", "valid-unpinned", "wrong-bytes", "oversize-pin"} {
				t.Run(variant, func(t *testing.T) {
					source := append([]byte{}, raw...)
					candidate := append([]byte{}, raw...)
					writes := map[string][]byte{}
					next := spec
					switch variant {
					case "changed-body":
						candidate = append(candidate, []byte("\nchanged candidate\n")...)
					case "orphan":
						writes[asset] = data
					case "missing-linked", "valid-unpinned", "wrong-bytes", "oversize-pin":
						source = append(source, []byte("\n![evidence]("+rel+")\n")...)
						candidate = source
						if variant != "missing-linked" {
							writes[asset] = data
						}
						if variant == "wrong-bytes" || variant == "oversize-pin" {
							pin := data
							if variant == "oversize-pin" {
								pin = []byte(strings.Repeat("x", issueassets.MaxFileBytes+1))
							}
							var blob string
							var err error
							if variant == "oversize-pin" {
								file := filepath.Join(t.TempDir(), "large")
								if err := os.WriteFile(file, pin, 0600); err != nil {
									t.Fatal(err)
								}
								blob = strings.TrimSpace(testfix.Capture(t, root, "hash-object", "-w", "--", file))
							} else {
								blob, err = gitx.WriteBlob(context.Background(), root, pin)
							}
							if err != nil {
								t.Fatal(err)
							}
							next.SourceAssets = FormatEvidenceEntries([]EvidenceEntry{{Path: asset, Blob: blob}})
							if variant == "wrong-bytes" {
								writes[asset] = []byte("corrupt")
							}
						}
					}
					next.SourceBlob, err = gitx.WriteBlob(context.Background(), root, source)
					if err != nil {
						t.Fatal(err)
					}
					// Creation receipts cannot contain asset pins. The transfer cases cover
					// bounded pin reads and exact linked member comparison.
					if operation == "creation" && next.SourceAssets != "" {
						return
					}
					var r Receipt
					if operation == "creation" {
						state, e := NewCreation(next)
						err = e
						r = state.Receipt()
					} else {
						state, e := NewTransfer(next)
						err = e
						r = state.Receipt()
					}
					if err != nil {
						t.Fatal(err)
					}
					writes[next.DestinationPath] = candidate
					c, err := main.PrepareCandidate("old unvalidated candidate", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
						return gitx.TrunkWrite{Write: writes, ExactBytes: true}, nil
					})
					if err != nil {
						t.Fatal(err)
					}
					before, err := main.Snapshot()
					if err != nil {
						t.Fatal(err)
					}
					transfer := NewTransferOp(context.Background(), creation.repo, main, root, next.SourceBranch, nil)
					for _, probe := range []bool{false, true} {
						effect := Effect{Stage: operation + ".main", Kind: PublishMain, Expected: r.binding(), CandidateOID: c.OID}
						effect.Expected.MainBase = c.Base
						if probe {
							effect.Kind = ProbePublication
						}
						var got error
						if operation == "creation" {
							if probe {
								_, got = creation.Probe(effect, r)
							} else {
								_, got = creation.Apply(effect, r)
							}
						} else {
							if probe {
								_, got = transfer.Probe(effect, r)
							} else {
								_, got = transfer.Apply(effect, r)
							}
						}
						if got == nil {
							t.Fatalf("probe=%v accepted %s candidate", probe, variant)
						}
					}
					after, err := main.Snapshot()
					if err != nil || after.Ref() != before.Ref() {
						t.Fatalf("refusal changed main: %v", err)
					}
				})
			}
		})
	}
}

func TestMainCandidateReplayPolicyOnlyBeforeNewPush(t *testing.T) {
	for _, landed := range []bool{false, true} {
		for _, probe := range []bool{false, true} {
			name := "apply"
			if probe {
				name = "probe"
			}
			if landed {
				name += "-landed"
			} else {
				name += "-absent"
			}
			t.Run(name, func(t *testing.T) {
				creation, _, root, _, main := creationFixture(t)
				r, err := creation.Start("policy-replay")
				if err != nil {
					t.Fatal(err)
				}
				spec := r.Spec()
				raw, err := main.ViewOf(spec.MainBase).ReadBlob(spec.SourceBlob)
				if err != nil {
					t.Fatal(err)
				}
				data := []byte("asset")
				assetName, _ := issueassets.Name(data, "png")
				rel := path.Join("assets", spec.IssueID, assetName)
				asset := path.Join(path.Dir(spec.DestinationPath), rel)
				raw = append(raw, []byte("\n![evidence]("+rel+")\n")...)
				spec.SourceBlob, err = gitx.WriteBlob(context.Background(), root, raw)
				if err != nil {
					t.Fatal(err)
				}
				blob, err := gitx.WriteBlob(context.Background(), root, data)
				if err != nil {
					t.Fatal(err)
				}
				spec.SourceAssets = FormatEvidenceEntries([]EvidenceEntry{{Path: asset, Blob: blob}})
				state, err := NewTransfer(spec)
				if err != nil {
					t.Fatal(err)
				}
				r = state.Receipt()
				c, err := DetailBundlePublication(main, spec.DestinationPath, raw, map[string][]byte{asset: data}, "valid old candidate")
				if err != nil {
					t.Fatal(err)
				}
				if landed {
					out, err := main.PushCandidate(c)
					if err != nil || out != gitx.PushAccepted {
						t.Fatalf("fixture push: %v %v", out, err)
					}
				}
				calls := 0
				op := NewTransferOp(context.Background(), creation.repo, main, root, spec.SourceBranch, nil)
				op.BeforePublish = func(_ ReceiptSpec, view *gitx.TrunkView) error {
					calls++
					if view.Ref() != c.Base {
						t.Fatal("policy not pinned to publication base")
					}
					return errors.New("destination now text-only")
				}
				e := Effect{Stage: "transfer.main", Kind: PublishMain, Expected: r.binding(), CandidateOID: c.OID}
				e.Expected.MainBase = c.Base
				var event Event
				if probe {
					e.Kind = ProbePublication
					event, err = op.Probe(e, r)
				} else {
					event, err = op.Apply(e, r)
				}
				if err != nil {
					t.Fatal(err)
				}
				if landed {
					if event.Kind != EventConfirmed || calls != 0 {
						t.Fatalf("landed proof incorrectly admitted again: %v, calls=%d", event.Kind, calls)
					}
				} else {
					if event.Kind == EventConfirmed || calls != 1 || op.Diagnostic() == nil || !strings.Contains(op.Diagnostic().Error(), "text-only") {
						t.Fatalf("new push bypassed policy: %v calls=%d diagnostic=%v", event.Kind, calls, op.Diagnostic())
					}
					view, err := main.Snapshot()
					if err != nil || view.Ref() != c.Base {
						t.Fatalf("refusal changed main: %v", err)
					}
				}
			})
		}
	}
}

func TestCreationLandedCandidateStillRequiresBundleProof(t *testing.T) {
	op, _, root, _, main := creationFixture(t)
	initial, err := op.Start("old-landed-creation")
	if err != nil {
		t.Fatal(err)
	}
	spec := initial.Spec()
	raw, err := main.ViewOf(spec.MainBase).ReadBlob(spec.SourceBlob)
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, []byte("\n![missing](assets/"+spec.IssueID+"/"+strings.Repeat("a", 64)+".png)\n")...)
	spec.SourceBlob, err = gitx.WriteBlob(context.Background(), root, raw)
	if err != nil {
		t.Fatal(err)
	}
	state, err := NewCreation(spec)
	if err != nil {
		t.Fatal(err)
	}
	r := state.Receipt()
	c, err := main.PrepareCandidate("old landed candidate", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: map[string][]byte{spec.DestinationPath: raw}, ExactBytes: true}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := main.PushCandidate(c)
	if err != nil || out != gitx.PushAccepted {
		t.Fatalf("fixture push: %v %v", out, err)
	}
	e := Effect{Stage: "creation.main", Kind: PublishMain, Expected: r.binding(), CandidateOID: c.OID}
	e.Expected.MainBase = c.Base
	for _, probe := range []bool{false, true} {
		if probe {
			e.Kind = ProbePublication
			_, err = op.Probe(e, r)
		} else {
			_, err = op.Apply(e, r)
		}
		if err == nil || !strings.Contains(err.Error(), "missing linked asset") || !strings.Contains(err.Error(), "refs/sdlc/recovery/"+spec.Token+":document") {
			t.Fatalf("invalid landed creation proof: %v", err)
		}
	}
}
