package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// This fixture publishes real immutable bytes to a private evidence store;
// current registry/policy live outside those bytes and can change independently.
func resumeSavedFixture(t *testing.T, child bool) (localRunDeps, string, string, ResumeLink) {
	t.Helper()
	d, sha, _, _ := localRunRepo(t, "skipped")
	dir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	store := suiteStoreRoot(dir)
	in := resumeCorePassed(t)
	in.Parent.Repo = d.Repo
	in.Parent.SHA = sha
	in.Parent.LogPath = "failed/aa.log"
	wall := 77.0
	in.Parent.WallSeconds = &wall
	parentLog := []byte("--- FAIL: TestOne (1.00s)\nFAIL\texample.org/p\t1.00s\n")
	in.Parent.LogSHA256 = resumeRecordDigest(parentLog)
	resumeSavedWrite(t, store, in.Parent, parentLog)
	parentRaw, _ := os.ReadFile(filepath.Join(store, "attempts", "aa.json"))
	in.ParentRecordSHA256 = resumeRecordDigest(parentRaw)
	in.ParentLogSHA256 = in.Parent.LogSHA256
	link := resumeCoreLink(in)
	resumeSavedWriteJSON(t, store, link.ID, link)
	if child {
		in.Child.Repo = d.Repo
		in.Child.SHA = sha
		in.Child.LogPath = "failed/cc.log"
		resumeSavedWrite(t, store, *in.Child, []byte(in.ChildLog))
	}
	d.CurrentSuitePolicy = func(context.Context) (flakes.Policy, error) { return flakes.PolicyAccept, nil }
	d.FlakeRegistry = gateRegistry(t, "example.org/p.TestOne")
	return d, sha, store, *link
}
func resumeSavedWriteJSON(t *testing.T, store, id string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store, "attempts", id+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}
func resumeSavedWrite(t *testing.T, store string, a SuiteAttempt, log []byte) {
	t.Helper()
	path := suiteAttemptLogFile(store, a)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, log, 0600); err != nil {
		t.Fatal(err)
	}
	resumeSavedWriteJSON(t, store, a.ID, a)
}
func TestResumeSavedCurrentComposite(t *testing.T) {
	d, sha, store, l := resumeSavedFixture(t, true)
	ctx := context.Background()
	before, err := os.ReadFile(filepath.Join(store, "attempts", "aa.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadSavedAttempt(store, d.Repo, sha, "suite", "aa", "h"); err == nil {
		t.Fatal("no-context loader gained linked authority")
	}
	rec, newest, err := loadSavedAttemptCurrent(ctx, d, store, sha, "suite", "aa", "h")
	if err != nil || !newest || !rec.passed() || rec.ResumeParent != "aa" || rec.ResumeChild != l.ChildID || rec.Duration != 77*time.Second || rec.Exit != 1 || rec.LogSHA256 != l.ParentLogSHA256 {
		t.Fatalf("composite %+v newest=%v err=%v", rec, newest, err)
	}
	if err := guardSuitePass(ctx, d, "suite", sha); err != nil {
		t.Fatal(err)
	}
	if err := guardSuitePass(ctx, d, "affected", sha); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := republishSavedPass(ctx, d, suiteCheck("false"), sha); !ok || err != nil {
		t.Fatalf("republish ok=%v err=%v", ok, err)
	}
	after, _ := os.ReadFile(filepath.Join(store, "attempts", "aa.json"))
	if string(after) != string(before) {
		t.Fatal("parent rewritten")
	}
	s, _ := readSuiteEvidence(store, testStatsRecordLimit)
	if len(s.Attempts) != 2 || s.attemptIDs["aa"].Outcome == string(flakes.KnownFlakePassed) {
		t.Fatalf("measurement mutated %+v", s.Attempts)
	}
}
func TestResumeSavedPendingRefusesFallback(t *testing.T) {
	d, sha, store, l := resumeSavedFixture(t, false)
	ctx := context.Background()
	// A later ordinary pass cannot erase the pending fence.
	pass := resumeCoreFixture(t).Parent
	pass.ID = "dd"
	pass.Repo = d.Repo
	pass.SHA = sha
	pass.Exit = 0
	pass.Outcome = ""
	pass.Failures = nil
	pass.FinishedAt = pass.FinishedAt.Add(time.Hour)
	pass.LogPath = "failed/dd.log"
	pass.LogSHA256 = resumeRecordDigest([]byte("pass\n"))
	resumeSavedWrite(t, store, pass, []byte("pass\n"))
	raw, _ := os.ReadFile(filepath.Join(store, "attempts", l.ID+".json"))
	command := resumeReconcileCommand(l.RerunOf, resumeRecordDigest(raw))
	if rec, _, err := loadSavedAttempt(store, d.Repo, sha, "suite", "dd", "h"); err == nil && rec.passed() {
		t.Fatal("legacy loader bypassed pending same-head proof through explicit later pass")
	}
	_, _, err := loadSavedAttemptCurrent(ctx, d, store, sha, "suite", "", "h")
	var unavailable *resumeProofUnavailable
	if !errors.As(err, &unavailable) || !strings.Contains(err.Error(), command) {
		t.Fatalf("missing executable recovery: %v", err)
	}
	for _, check := range []string{"suite", "affected"} {
		if err := guardSuitePass(ctx, d, check, sha); err == nil || !strings.Contains(err.Error(), command) {
			t.Fatalf("%s guard %v", check, err)
		}
	}
	if _, ok, err := republishSavedPass(ctx, d, suiteCheck("true"), sha); !ok || !errors.As(err, &unavailable) {
		t.Fatalf("fallback allowed ok=%v err=%v", ok, err)
	}
	marker := filepath.Join(t.TempDir(), "ran")
	_, err = newLocalCheckRunner(d)(ctx, suiteCheck("./marker-check.sh", marker), sha)
	if !errors.As(err, &unavailable) {
		t.Fatalf("runner refusal %v", err)
	}
	if _, e := os.Stat(marker); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("blocked saved proof caused execution")
	}
}
func TestResumeSavedRevalidatesRegistryPolicyAndBytes(t *testing.T) {
	for _, kind := range []string{"registry removed", "registry unreadable", "policy off", "policy unreadable", "child log changed", "parent record changed", "child omitted pass", "receipt", "unreserved child"} {
		t.Run(kind, func(t *testing.T) {
			d, sha, store, l := resumeSavedFixture(t, true)
			ctx := context.Background()
			if _, _, e := loadSavedAttemptCurrent(ctx, d, store, sha, "suite", "", "h"); e != nil {
				t.Fatal(e)
			}
			switch kind {
			case "registry removed":
				d.FlakeRegistry = gateRegistry(t)
			case "registry unreadable":
				d.FlakeRegistry = func(context.Context) (flakes.Registry, error) {
					return flakes.Registry{}, errors.New("registry unavailable")
				}
			case "policy off":
				d.CurrentSuitePolicy = func(context.Context) (flakes.Policy, error) { return flakes.PolicyOff, nil }
			case "policy unreadable":
				d.CurrentSuitePolicy = func(context.Context) (flakes.Policy, error) { return "", errors.New("policy unavailable") }
			case "child log changed":
				if e := os.WriteFile(filepath.Join(store, "failed", "cc.log"), []byte("changed"), 0600); e != nil {
					t.Fatal(e)
				}
			case "parent record changed":
				s, _ := readSuiteEvidence(store, testStatsRecordLimit)
				a := s.attemptIDs["aa"]
				a.Slot = ""
				a.Exit = 2
				resumeSavedWriteJSON(t, store, a.ID, a)
			case "child omitted pass":
				s, _ := readSuiteEvidence(store, testStatsRecordLimit)
				a := s.attemptIDs["cc"]
				log := []byte("{\"Action\":\"pass\",\"Package\":\"example.org/p\",\"Elapsed\":1}\n")
				a.LogSHA256 = resumeRecordDigest(log)
				resumeSavedWrite(t, store, a, log)
			case "receipt":
				raw, _ := os.ReadFile(filepath.Join(store, "attempts", "bb.json"))
				r := ResumeReconciliation{Version: 3, Kind: "resume-reconciliation", ID: "dd", Repo: d.Repo, SHA: sha, Check: "suite", RerunOf: "aa", LinkID: l.ID, LinkSHA256: resumeRecordDigest(raw), ChildID: l.ChildID, Actor: "operator", Note: "stopped", Outcome: "interrupted", FinishedAt: l.CreatedAt}
				resumeSavedWriteJSON(t, store, r.ID, r)
			case "unreserved child":
				s, _ := readSuiteEvidence(store, testStatsRecordLimit)
				a := s.attemptIDs["cc"]
				a.ID = "dd"
				a.LogPath = "failed/dd.log"
				resumeSavedWrite(t, store, a, []byte("unexpected child"))
			}
			var unavailable *resumeProofUnavailable
			if _, _, e := loadSavedAttemptCurrent(ctx, d, store, sha, "suite", "", "h"); !errors.As(e, &unavailable) {
				t.Fatalf("lookup did not refuse changed %s: %v", kind, e)
			}
			if e := guardSuitePass(ctx, d, "suite", sha); e == nil {
				t.Fatalf("guard did not refuse changed %s", kind)
			}
			if _, ok, e := republishSavedPass(ctx, d, suiteCheck("true"), sha); !ok || !errors.As(e, &unavailable) {
				t.Fatalf("changed proof triggered fallback %v %v", ok, e)
			}
		})
	}
}
func TestResumeSavedScopedAndReadsPolicyOnce(t *testing.T) {
	d, sha, store, _ := resumeSavedFixture(t, true)
	ctx := context.Background()
	policyReads, registryReads := 0, 0
	reg := d.FlakeRegistry
	d.CurrentSuitePolicy = func(context.Context) (flakes.Policy, error) { policyReads++; return flakes.PolicyRerun, nil }
	d.FlakeRegistry = func(ctx context.Context) (flakes.Registry, error) { registryReads++; return reg(ctx) }
	if _, _, err := loadSavedAttemptCurrent(ctx, d, store, sha, "suite", "", "h"); err != nil {
		t.Fatal(err)
	}
	if policyReads != 1 || registryReads != 1 {
		t.Fatalf("reads policy=%d registry=%d", policyReads, registryReads)
	}
	other := strings.Repeat("f", 40)
	if err := guardSuitePass(ctx, d, "suite", other); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadSavedAttemptCurrent(ctx, d, store, sha, "lint", "", "h"); !errors.Is(err, errNoSavedAttempt) {
		t.Fatalf("unrelated check %v", err)
	}
	if policyReads != 1 || registryReads != 1 {
		t.Fatal("unrelated subject read resume policy")
	}
}

func TestResumeSavedGuardUsesOneRegistryRead(t *testing.T) {
	d, sha, _, _ := resumeSavedFixture(t, true)
	reads := 0
	registry := d.FlakeRegistry
	d.FlakeRegistry = func(ctx context.Context) (flakes.Registry, error) { reads++; return registry(ctx) }
	if err := guardSuitePass(context.Background(), d, "suite", sha); err != nil {
		t.Fatal(err)
	}
	if reads != 1 {
		t.Fatalf("same snapshot fetched registry %d times", reads)
	}
}
func TestResumeSavedEveryParentBlocksAggregate(t *testing.T) {
	d, sha, store, first := resumeSavedFixture(t, true)
	s, _ := readSuiteEvidence(store, testStatsRecordLimit)
	p := s.attemptIDs[first.RerunOf]
	p.ID = "dd"
	p.LogPath = "failed/dd.log"
	p.FinishedAt = p.FinishedAt.Add(time.Hour)
	p.StartedAt = p.StartedAt.Add(time.Hour)
	log, e := os.ReadFile(suiteAttemptLogFile(store, s.attemptIDs[first.RerunOf]))
	if e != nil {
		t.Fatal(e)
	}
	resumeSavedWrite(t, store, p, log)
	raw, _ := os.ReadFile(filepath.Join(store, "attempts", p.ID+".json"))
	pending := first
	pending.ID = "ee"
	pending.RerunOf = p.ID
	pending.ChildID = "ff"
	pending.ParentRecordSHA256 = resumeRecordDigest(raw)
	pending.CreatedAt = p.FinishedAt
	resumeSavedWriteJSON(t, store, pending.ID, pending)
	for _, id := range []string{"", first.RerunOf, p.ID} {
		if _, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, "suite", id, "h"); err == nil || !strings.Contains(err.Error(), "parent dd") {
			t.Fatalf("lookup %q bypassed second parent: %v", id, err)
		}
	}
	if err := guardSuitePass(context.Background(), d, "affected", sha); err == nil {
		t.Fatal("affected pass bypassed pending suite parent")
	}
}

type resumePostMutationGH struct {
	*ciFakeGH
	mutate func()
	posts  int
}

func (g *resumePostMutationGH) PostStatus(ctx context.Context, repo, sha, status, state, desc string) error {
	g.posts++
	if g.posts == 1 {
		g.mutate()
		return errors.New("status reply lost")
	}
	return g.ciFakeGH.PostStatus(ctx, repo, sha, status, state, desc)
}
func TestResumeSavedRetryRevalidates(t *testing.T) {
	d, sha, store, _ := resumeSavedFixture(t, true)
	ctx := context.Background()
	rec, _, err := loadSavedAttemptCurrent(ctx, d, store, sha, "suite", "", "h")
	if err != nil {
		t.Fatal(err)
	}
	open := true
	reg := d.FlakeRegistry
	d.FlakeRegistry = func(ctx context.Context) (flakes.Registry, error) {
		if !open {
			return flakes.Registry{}, nil
		}
		return reg(ctx)
	}
	gh := &resumePostMutationGH{ciFakeGH: d.GH.(*ciFakeGH), mutate: func() { open = false }}
	d.GH = gh
	prev := localCheckPostBackoff
	localCheckPostBackoff = []time.Duration{time.Millisecond}
	defer func() { localCheckPostBackoff = prev }()
	err = postLocalCheckRecord(ctx, d, rec)
	if err == nil || gh.posts != 1 || gh.statuses[sha][localStatusContext("suite")] == "success" {
		t.Fatalf("retry bypassed changed registry posts=%d err=%v statuses=%v", gh.posts, err, gh.statuses)
	}
}

func TestResumeSavedPostBindsCompositeIdentity(t *testing.T) {
	d, sha, store, _ := resumeSavedFixture(t, true)
	ctx := context.Background()
	rec, _, err := loadSavedAttemptCurrent(ctx, d, store, sha, "suite", "", "h")
	if err != nil {
		t.Fatal(err)
	}
	rec.ResumeChild = "ff"
	if err := postLocalCheckRecord(ctx, d, rec); err == nil {
		t.Fatal("posted stale child identity despite fresh proof naming cc")
	}
	gh := d.GH.(*ciFakeGH)
	if gh.statuses[sha][localStatusContext("suite")] == "success" {
		t.Fatal("wrong composite posted success")
	}
}

func TestResumeSavedRegistryChangeBeforeRepublishNeverFallsBack(t *testing.T) {
	d, sha, _, _ := resumeSavedFixture(t, true)
	original := d.FlakeRegistry
	reads := 0
	d.FlakeRegistry = func(ctx context.Context) (flakes.Registry, error) {
		reads++
		if reads > 1 {
			return flakes.Registry{}, nil
		}
		return original(ctx)
	}
	var unavailable *resumeProofUnavailable
	if _, ok, err := republishSavedPass(context.Background(), d, suiteCheck("true"), sha); !ok || !errors.As(err, &unavailable) {
		t.Fatalf("registry changed after lookup permits replacement: handled=%v error=%v reads=%d", ok, err, reads)
	}
}

func TestResumeSavedTerminalReceiptAndLogBoundary(t *testing.T) {
	t.Run("terminal receipt is never another reconciliation or execution", func(t *testing.T) {
		d, sha, store, l := resumeSavedFixture(t, false)
		raw, _ := os.ReadFile(filepath.Join(store, "attempts", l.ID+".json"))
		receipt := ResumeReconciliation{Version: 3, Kind: "resume-reconciliation", ID: "dd", Repo: d.Repo, SHA: sha, Check: "suite", RerunOf: l.RerunOf, LinkID: l.ID, LinkSHA256: resumeRecordDigest(raw), ChildID: l.ChildID, Actor: "operator", Note: "verified stopped", Outcome: "interrupted", FinishedAt: l.CreatedAt}
		resumeSavedWriteJSON(t, store, receipt.ID, receipt)
		var unavailable *resumeProofUnavailable
		_, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, "suite", "", "h")
		if !errors.As(err, &unavailable) || !strings.Contains(err.Error(), "terminally reconciled") || strings.Contains(err.Error(), "local-check reconcile --attempt") {
			t.Fatalf("terminal receipt recovery %v", err)
		}
		if _, ok, err := republishSavedPass(context.Background(), d, suiteCheck("true"), sha); !ok || !errors.As(err, &unavailable) {
			t.Fatalf("terminal receipt fallback %v %v", ok, err)
		}
	})
	t.Run("symlink log cannot supply proof", func(t *testing.T) {
		d, sha, store, _ := resumeSavedFixture(t, true)
		path := filepath.Join(store, "failed", "cc.log")
		if e := os.Rename(path, path+".retained"); e != nil {
			t.Fatal(e)
		}
		if e := os.Symlink(path+".retained", path); e != nil {
			t.Fatal(e)
		}
		if _, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, "suite", "", "h"); err == nil {
			t.Fatal("symlink log verified")
		}
	})
	t.Run("current policy default fails closed if main cannot be read", func(t *testing.T) {
		d, sha, store, _ := resumeSavedFixture(t, true)
		d.CurrentSuitePolicy = nil
		var unavailable *resumeProofUnavailable
		if _, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, "suite", "", "h"); !errors.As(err, &unavailable) || !strings.Contains(err.Error(), "current suite policy") {
			t.Fatalf("unreadable declaration %v", err)
		}
	})
}

func TestResumeSavedRepublishUsesOneRegistryPerDecision(t *testing.T) {
	d, sha, _, _ := resumeSavedFixture(t, true)
	reads := 0
	registry := d.FlakeRegistry
	d.FlakeRegistry = func(ctx context.Context) (flakes.Registry, error) { reads++; return registry(ctx) }
	if _, ok, err := republishSavedPass(context.Background(), d, suiteCheck("true"), sha); !ok || err != nil {
		t.Fatalf("republish %v %v", ok, err)
	}
	if reads != 2 {
		t.Fatalf("want one lookup and one fresh post registry read; got %d", reads)
	}
}

// Counted multi-bundle regression follows.
func TestResumeSavedGrowthCounts(t *testing.T) {
	for _, n := range []int{4, 8} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			d, sha, store, baseLink := resumeSavedFixture(t, true)
			base, err := readSuiteEvidence(store, testStatsRecordLimit)
			if err != nil {
				t.Fatal(err)
			}
			parent, child := base.attemptIDs[baseLink.RerunOf], base.attemptIDs[baseLink.ChildID]
			parentLog, err := readResumeLog(store, parent)
			if err != nil {
				t.Fatal(err)
			}
			childLog, err := readResumeLog(store, child)
			if err != nil {
				t.Fatal(err)
			}
			for i := 1; i < n; i++ {
				p := parent
				p.ID = fmt.Sprintf("%04x", 256+3*i)
				p.LogPath = "failed/" + p.ID + ".log"
				resumeSavedWrite(t, store, p, parentLog)
				raw, err := os.ReadFile(filepath.Join(store, "attempts", p.ID+".json"))
				if err != nil {
					t.Fatal(err)
				}
				l := baseLink
				l.ID = fmt.Sprintf("%04x", 257+3*i)
				l.ChildID = fmt.Sprintf("%04x", 258+3*i)
				l.RerunOf = p.ID
				l.ParentRecordSHA256 = resumeRecordDigest(raw)
				resumeSavedWriteJSON(t, store, l.ID, l)
				c := child
				c.ID = l.ChildID
				c.RerunOf = p.ID
				c.LogPath = "failed/" + c.ID + ".log"
				resumeSavedWrite(t, store, c, childLog)
			}
			// Independently account for actual directory/record opens while building
			// the same indexed snapshot used by production projection.
			opens := map[string]int{}
			snapshot, err := readSuiteEvidenceWithOpen(store, testStatsRecordLimit, suiteAttemptReaderBudget, func(path string) (*os.File, error) { opens[path]++; return os.Open(path) })
			if err != nil || len(snapshot.Warnings) > 0 || len(snapshot.Attempts) != 2*n || len(snapshot.Resumes) != n || len(opens) != 3*n+1 {
				t.Fatalf("snapshot growth n=%d opens=%d attempts=%d links=%d warnings=%v error=%v", n, len(opens), len(snapshot.Attempts), len(snapshot.Resumes), snapshot.Warnings, err)
			}
			for path, count := range opens {
				if count != 1 {
					t.Fatalf("record rescanned %s %d times", path, count)
				}
			}
			originalRead := readResumeLog
			defer func() { readResumeLog = originalRead }()
			logReads, policyReads, registryReads := 0, 0, 0
			readResumeLog = func(root string, a SuiteAttempt) ([]byte, error) { logReads++; return originalRead(root, a) }
			registry := d.FlakeRegistry
			d.FlakeRegistry = func(ctx context.Context) (flakes.Registry, error) { registryReads++; return registry(ctx) }
			d.CurrentSuitePolicy = func(context.Context) (flakes.Policy, error) { policyReads++; return flakes.PolicyRerun, nil }
			for _, decision := range []string{"snapshot projection", "saved lookup", "success guard", "saved republish"} {
				logReads, policyReads, registryReads = 0, 0, 0
				decisions := 1
				switch decision {
				case "snapshot projection":
					records, e := resolvedResumeRecords(context.Background(), d, store, snapshot, sha, "suite", "h")
					if e != nil || len(records) != n {
						t.Fatalf("projection %d %v", len(records), e)
					}
				case "saved lookup":
					rec, newest, e := loadSavedAttemptCurrent(context.Background(), d, store, sha, "suite", "", "h")
					if e != nil || !newest || !rec.passed() {
						t.Fatalf("lookup %+v %v %v", rec, newest, e)
					}
				case "success guard":
					if e := guardSuitePass(context.Background(), d, "suite", sha); e != nil {
						t.Fatal(e)
					}
				case "saved republish":
					decisions = 2
					if _, ok, e := republishSavedPass(context.Background(), d, suiteCheck("true"), sha); !ok || e != nil {
						t.Fatalf("republish %v %v", ok, e)
					}
				}
				if logReads != 2*n*decisions || policyReads != decisions || registryReads != decisions {
					t.Fatalf("%s n=%d logreads=%d policyreads=%d registryreads=%d; want %d/%d/%d", decision, n, logReads, policyReads, registryReads, 2*n*decisions, decisions, decisions)
				}
			}
		})
	}
}
