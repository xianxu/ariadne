package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
	"github.com/xianxu/ariadne/pkg/weaveownership"
)

func TestPlanDeferredSeedsMixedActions(t *testing.T) {
	rng := rand.New(rand.NewSource(557))
	for trial := 0; trial < 100; trial++ {
		var actions []plan.Action
		for i := 0; i < rng.Intn(100)+1; i++ {
			p := fmt.Sprintf("output-%d", i)
			switch rng.Intn(6) {
			case 0:
				actions = append(actions, plan.Seed{Src: "source", Dst: p})
			case 1:
				actions = append(actions, plan.SeedOnce{Src: "source", Dst: p})
			case 2:
				actions = append(actions, plan.WriteFile{Path: p, Content: "managed"})
			case 3:
				actions = append(actions, plan.Touch{Path: p})
			case 4:
				actions = append(actions, plan.Mkdir{Path: p})
			default:
				actions = append(actions, plan.DeferredOutput{Path: p})
			}
		}
		before := append([]plan.Action(nil), actions...)
		got := planDeferredSeeds(actions)
		if !reflect.DeepEqual(actions, before) || len(got.actions) != len(actions) {
			t.Fatal("changed input or action count")
		}
		seed := 0
		for i, a := range actions {
			var dst string
			switch a := a.(type) {
			case plan.Seed:
				dst = a.Dst
			case plan.SeedOnce:
				dst = a.Dst
			default:
				if !reflect.DeepEqual(a, got.actions[i]) {
					t.Fatalf("nonseed %d changed", i)
				}
				continue
			}
			if !reflect.DeepEqual(got.actions[i], plan.DeferredOutput{Path: dst}) || !reflect.DeepEqual(got.seeds[seed], deferredSeed{path: dst, action: a}) {
				t.Fatalf("seed %d not deferred", i)
			}
			seed++
		}
		if seed != len(got.seeds) || !reflect.DeepEqual(planDeferredSeeds(got.actions).actions, got.actions) {
			t.Fatal("unstable classification")
		}
	}
}

func TestDeferredSeedsPreserveDestinationAndEvidence(t *testing.T) {
	for _, once := range []bool{false, true} {
		for _, state := range []string{"missing", "current", "edited", "mode", "symlink", "owned", "missing source"} {
			t.Run(fmt.Sprintf("once=%v/%s", once, state), func(t *testing.T) {
				root := t.TempDir()
				src := filepath.Join(t.TempDir(), "seed")
				desired := "-include Makefile.workflow\n"
				mkfile(t, src, desired)
				if err := os.Chmod(src, 0755); err != nil {
					t.Fatal(err)
				}
				const target = "entrypoint"
				dst := filepath.Join(root, target)
				backup := filepath.Join(root, plan.DisplacedPath(target))
				if state == "owned" {
					if _, err := plan.ApplyManaged(weavefs.OSFS{}, root, []plan.Action{plan.WriteFile{Path: target, Content: "old owned"}}, plan.ScopeArtifacts); err != nil {
						t.Fatal(err)
					}
					mkfile(t, backup, "prior recovery")
				}
				switch state {
				case "current", "mode":
					mkfile(t, dst, desired)
				case "edited", "owned", "missing source":
					mkfile(t, dst, "consumer edit")
				case "symlink":
					mkfile(t, filepath.Join(root, "victim"), "victim")
					if err := os.Symlink("victim", dst); err != nil {
						t.Fatal(err)
					}
				}
				if state == "current" {
					if err := os.Chmod(dst, 0755); err != nil {
						t.Fatal(err)
					}
				}
				if state == "missing source" {
					if err := os.Remove(src); err != nil {
						t.Fatal(err)
					}
				}
				before, _ := os.ReadFile(dst)
				info, statErr := os.Lstat(dst)
				oldInventory, err := weaveownership.Read(weavefs.OSFS{}, root)
				if err != nil {
					t.Fatal(err)
				}
				attempts := 0
				fs := protectedSettingsFS{targets: map[string]bool{dst: true, backup: true}, attempts: &attempts}
				var action plan.Action = plan.Seed{Src: src, Dst: target}
				if once {
					action = plan.SeedOnce{Src: src, Dst: target}
				}
				var out bytes.Buffer
				// Duplicate candidates must not duplicate the operator notice.
				actions, err := deferredSeedOutputs(fs, root, []plan.Action{action, action}, &out)
				if err != nil {
					t.Fatal(err)
				}
				wantNotice := state != "current" && state != "missing source" && !(once && (state == "mode" || state == "symlink"))
				if got := strings.Count(out.String(), "deferred during planning"); got != boolCount(wantNotice) {
					t.Fatalf("notice=%d want=%v: %s", got, wantNotice, &out)
				}
				if _, err := plan.ApplyManaged(fs, root, actions, plan.ScopeArtifacts); err != nil {
					t.Fatal(err)
				}
				if attempts != 0 {
					t.Fatalf("%d protected mutation attempts", attempts)
				}
				after, _ := os.ReadFile(dst)
				afterInfo, afterErr := os.Lstat(dst)
				if !bytes.Equal(before, after) || os.IsNotExist(statErr) != os.IsNotExist(afterErr) {
					t.Fatal("destination content/existence changed")
				}
				if statErr == nil && (afterErr != nil || info.Mode() != afterInfo.Mode()) {
					t.Fatal("destination mode changed")
				}
				if state == "symlink" {
					if link, err := os.Readlink(dst); err != nil || link != "victim" {
						t.Fatal("link changed", link, err)
					}
				}
				inventory, err := weaveownership.Read(weavefs.OSFS{}, root)
				if err != nil {
					t.Fatal(err)
				}
				if len(oldInventory) != len(inventory) || (len(inventory) > 0 && !reflect.DeepEqual(oldInventory, inventory)) {
					t.Fatalf("ownership changed: before=%+v after=%+v", oldInventory, inventory)
				}
				if state == "owned" {
					if b, err := os.ReadFile(backup); err != nil || string(b) != "prior recovery" {
						t.Fatal("lost recovery evidence", string(b), err)
					}
				}
			})
		}
	}
}

func boolCount(b bool) int {
	if b {
		return 1
	}
	return 0
}

type seedObservationFS struct {
	weavefs.OSFS
	reads, stats, lstats int
	statError            error
}

func (f *seedObservationFS) ReadFile(p string) ([]byte, error) { f.reads++; return f.OSFS.ReadFile(p) }
func (f *seedObservationFS) Stat(p string) (os.FileInfo, error) {
	f.stats++
	if f.statError != nil {
		return nil, f.statError
	}
	return f.OSFS.Stat(p)
}
func (f *seedObservationFS) Lstat(p string) (os.FileInfo, error) { f.lstats++; return f.OSFS.Lstat(p) }

type seedErrorWriter struct{ err error }

func (w seedErrorWriter) Write([]byte) (int, error) { return 0, w.err }

func TestDeferredSeedsObservationBoundAndErrors(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			root := t.TempDir()
			src := filepath.Join(t.TempDir(), "seed")
			mkfile(t, src, "desired")
			var actions []plan.Action
			for i := 0; i < n; i++ {
				actions = append(actions, plan.Seed{Src: src, Dst: fmt.Sprint(i)})
			}
			fs := &seedObservationFS{}
			if _, err := deferredSeedOutputs(fs, root, actions, io.Discard); err != nil {
				t.Fatal(err)
			}
			if fs.reads != n || fs.stats != n || fs.lstats != n {
				t.Fatalf("observations at N=%d: reads=%d stats=%d lstats=%d", n, fs.reads, fs.stats, fs.lstats)
			}
		})
	}
	root := t.TempDir()
	src := filepath.Join(t.TempDir(), "seed")
	mkfile(t, src, "desired")
	actions := []plan.Action{plan.Seed{Src: src, Dst: "target"}}
	sentinel := errors.New("observation failed")
	if _, err := deferredSeedOutputs(&seedObservationFS{statError: sentinel}, root, actions, io.Discard); !errors.Is(err, sentinel) || !strings.Contains(err.Error(), "target") {
		t.Fatalf("observation failure: %v", err)
	}
	if _, err := deferredSeedOutputs(weavefs.OSFS{}, root, actions, seedErrorWriter{sentinel}); !errors.Is(err, sentinel) {
		t.Fatalf("diagnostic failure: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(root, "target")); !os.IsNotExist(err) {
		t.Fatal("failure published seed", err)
	}
}

// Fail after preflight, at the actual managed publication boundary. Suppression
// must survive partial nonseed publication and cancellation without rollback.
type interruptedSeedPublicationFS struct {
	protectedSettingsFS
	failPath string
	stop     func() error
}

func (f interruptedSeedPublicationFS) Rename(a, b string) error {
	if b == f.failPath {
		return f.stop()
	}
	return f.protectedSettingsFS.Rename(a, b)
}
func TestPlanningCompileInterruptedPublicationKeepsSeeds(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(fmt.Sprintf("cancelled=%v", cancelled), func(t *testing.T) {
			_, root := declineFixture(t, "")
			compileOK(t, root, false)
			const target = ".github/workflows/merge-check.yml"
			dst := filepath.Join(root, target)
			mkfile(t, dst, "consumer workflow")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var out bytes.Buffer
			client, runner, closeSetup, err := prepareSetup(ctx, root, false, &out, &out)
			if err != nil {
				t.Fatal(err)
			}
			defer closeSetup()
			attempts := 0
			sentinel := errors.New("publication failed")
			fs := interruptedSeedPublicationFS{
				protectedSettingsFS: protectedSettingsFS{targets: map[string]bool{dst: true}, attempts: &attempts},
				failPath:            filepath.Join(root, "AGENTS.md"),
				stop: func() error {
					if cancelled {
						cancel()
						return ctx.Err()
					}
					return sentinel
				},
			}
			err = compilePreparedWithPolicy(ctx, fs, root, plan.TargetAll, false, &out, client, runner, deferPlanningOutputs)
			want := sentinel
			if cancelled {
				want = context.Canceled
			}
			if !errors.Is(err, want) {
				t.Fatalf("publication error: %v, output: %s", err, &out)
			}
			if attempts != 0 {
				t.Fatalf("attempted %d seed writes", attempts)
			}
			if b, err := os.ReadFile(dst); err != nil || string(b) != "consumer workflow" {
				t.Fatal("interrupted compile changed seed", string(b), err)
			}
		})
	}
}
