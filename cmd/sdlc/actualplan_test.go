package main

import (
	"context"
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestSelectFreshSplitRoundsRetiredBoundaryProperties(t *testing.T) {
	random := rand.New(rand.NewSource(531))
	for run := 0; run < 500; run++ {
		var rounds []activetime.BoundaryRound
		var members []string
		final := map[string]bool{}
		for n := 0; n < random.Intn(40)+1; n++ {
			name := []string{"", "M1", "M2", "M3", "M4b"}[random.Intn(5)]
			done := random.Intn(4) == 0
			rounds = append(rounds, activetime.BoundaryRound{Boundary: name, Finalizing: done, Finalized: time.Unix(int64(n), 0)})
			final[name] = final[name] || done
			if random.Intn(3) == 0 {
				members = append(members, name)
			}
		}
		got, retired := selectFreshSplitRounds(rounds, members)
		var want []activetime.BoundaryRound
		for _, r := range rounds {
			if r.Boundary == "" || final[r.Boundary] || slices.Contains(members, r.Boundary) {
				want = append(want, r)
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("selection changed preserved history: got=%+v want=%+v", got, want)
		}
		for _, name := range retired {
			if name == "" || final[name] || slices.Contains(members, name) {
				t.Fatalf("invalid retirement %q", name)
			}
		}
		again, removed := selectFreshSplitRounds(got, members)
		if !reflect.DeepEqual(again, got) || len(removed) != 0 {
			t.Fatal("selection is not idempotent")
		}
	}
}

func TestPinnedActualPlanRetiredBoundaryAuthority(t *testing.T) {
	for _, tc := range []struct {
		name, body, other, root, identity string
		want                              string
	}{
		{name: "active", body: "## Plan\n- [x] M1 — done\n- M2 — historical\n- [ ] M3 — work\n"},
		{name: "fenced examples", body: "## Plan\n```md\n- [ ] M2 — example\n```\n- [ ] M3 — work\n"},
		{name: "archive", root: "workshop/history/issues", body: "## Plan\n- [ ] M3 — work\n"},
		{name: "custom root", root: "custom/issues", body: "## Plan\n- [ ] M3 — work\n"},
		{name: "wrong identity", identity: "000010", body: "## Plan\n- [ ] M3\n", want: "identity"},
		{name: "duplicate identity", identity: "000009\nid: 000009", body: "## Plan\n- [ ] M3\n", want: "frontmatter"},
		{name: "quoted identity", identity: "\"000009\"", body: "## Plan\n- [ ] M3\n"},
		{name: "absent", body: "## Log\nretired M2\n", want: "Plan section"},
		{name: "duplicate section", body: "## Plan\n- [ ] M3\n## Log\nx\n## Plan\n- [ ] M2\n", want: "Plan section"},
		{name: "unclosed fence", body: "## Plan\n```md\n- [ ] M2\n", want: "unclosed"},
		{name: "duplicate files", body: "## Plan\n- [ ] M3\n", other: "workshop/history/issues/000009-y.md", want: "found 2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit())
			dir := tc.root
			if dir == "" {
				dir = "workshop/issues"
			}
			if tc.name == "custom root" {
				t.Setenv("WF_ISSUES_DIR", filepath.Join(root, dir))
			}
			path := filepath.Join(root, dir, "000009-x.md")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			identity := tc.identity
			if identity == "" {
				identity = "000009"
			}
			raw := "---\nid: " + identity + "\n---\n# fixture\n\n" + tc.body
			if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			if tc.other != "" {
				other := filepath.Join(root, tc.other)
				if err := os.MkdirAll(filepath.Dir(other), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(other, []byte(raw), 0600); err != nil {
					t.Fatal(err)
				}
			}
			testfix.Git(t, root, "add", ".")
			testfix.Git(t, root, "commit", "-qm", "#9: pinned plan")
			pin := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
			// Both a later committed plan and dirty changes must be irrelevant.
			if err := os.WriteFile(path, []byte("---\nid: 000009\n---\n## Plan\n- [ ] M2\n"), 0600); err != nil {
				t.Fatal(err)
			}
			testfix.Git(t, root, "add", ".")
			testfix.Git(t, root, "commit", "-qm", "#9: later plan")
			if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := pinnedActualPlan(context.Background(), root, pin, "9")
			if tc.want != "" {
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("got=%v err=%v want=%q", got, err, tc.want)
				}
				return
			}
			if err != nil || !slices.Contains(got, "M3") || slices.Contains(got, "M2") {
				t.Fatalf("unpinned plan: %v %v", got, err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err := pinnedActualPlan(ctx, root, pin, "9"); err == nil {
				t.Fatal("cancellation ignored")
			}
		})
	}
}

func TestFreshActualRoundsRetiredBoundaryRefusals(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	pin := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	// No named open rounds need no new Plan authority.
	if _, _, err := freshActualRounds(context.Background(), root, pin, "9", "", nil); err != nil {
		t.Fatal(err)
	}
	rounds := []activetime.BoundaryRound{{Boundary: "M2"}}
	_, _, err := freshActualRounds(context.Background(), root, pin, "000009", "M3", rounds)
	if err == nil || !strings.Contains(err.Error(), "retry sdlc actual --issue 9") {
		t.Fatalf("missing actionable refusal: %v", err)
	}
	path := filepath.Join(root, "workshop/issues/000009-x.md")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("---\nid: 000009\n---\n## Plan\n- [ ] M2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "#9: executable M2")
	pin = strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	_, _, err = freshActualRounds(context.Background(), root, pin, "000009", "M3", rounds)
	if err == nil || !strings.Contains(err.Error(), "select its replacement") || !strings.Contains(err.Error(), "--issue 9 --milestone") {
		t.Fatalf("absent current accepted: %v", err)
	}
	// Both executable rows preserve the open M2 and expose why M3 is empty.
	raw := []byte("---\nid: 000009\n---\n## Plan\n- [ ] M2\n- [ ] M3\n")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "#9: both boundaries executable")
	pin = strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	kept, notes, err := freshActualRounds(context.Background(), root, pin, "000009", "M3", rounds)
	if err != nil || len(kept) != 1 || !strings.Contains(strings.Join(notes, " "), "retained executable unfinalized boundary M2") {
		t.Fatalf("open M2 silently retired: %v %v %v", kept, notes, err)
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	shim := t.TempDir()
	log := filepath.Join(shim, "calls")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$ACTUAL_PLAN_GIT_LOG\"\nexec '" + strings.ReplaceAll(realGit, "'", "'\"'\"'") + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(shim, "git"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ACTUAL_PLAN_GIT_LOG", log)
	t.Setenv("PATH", shim+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, n := range []int{100, 200} {
		if err := os.WriteFile(log, nil, 0600); err != nil {
			t.Fatal(err)
		}
		rs := make([]activetime.BoundaryRound, n)
		for i := range rs {
			rs[i] = activetime.BoundaryRound{Boundary: "M2", Finalized: time.Unix(int64(i), 0)}
		}
		got, notes, err := freshActualRounds(context.Background(), root, pin, "9", "M2", rs)
		if err != nil || len(got) != n {
			t.Fatalf("N=%d: %v %v", n, notes, err)
		}
		calls, err := os.ReadFile(log)
		if err != nil {
			t.Fatal(err)
		}
		if len(strings.Split(strings.TrimSpace(string(calls)), "\n")) != 2 || strings.Count(string(calls), "show ") != 1 {
			t.Fatalf("N=%d unbounded Plan observation: %s", n, calls)
		}
	}
}

func TestCapturedPair737RetiredBoundaryReplay(t *testing.T) {
	var capture struct {
		Since, Until, Current, Plan string
		ObservedActivityStart       string `json:"observed_activity_start"`
		ObservedActivityEnd         string `json:"observed_activity_end"`
		Rounds                      []struct {
			Boundary, Dispatched, Finalized string
			Finalizing                      bool
		}
	}
	raw, err := os.ReadFile("testdata/actual/pair737-retired-boundary.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &capture); err != nil {
		t.Fatal(err)
	}
	parse := func(s string) time.Time {
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	var rounds []activetime.BoundaryRound
	for _, r := range capture.Rounds {
		rounds = append(rounds, activetime.BoundaryRound{Boundary: r.Boundary, Dispatched: parse(r.Dispatched), Finalized: parse(r.Finalized), Finalizing: r.Finalizing})
	}
	start, end := parse(capture.ObservedActivityStart), parse(capture.ObservedActivityEnd)
	segments := []activetime.Segment{{Start: start, End: end, Active: 38, Alloc: map[string]float64{"737": 38}}}
	old := activetime.Split(segments, "737", activetime.BuildTimeline(rounds, nil, capture.Current, parse(capture.Since), parse(capture.Until)).For(capture.Current))
	if old.Build+old.Converge != 0 {
		t.Fatalf("capture no longer demonstrates original zero: %+v", old)
	}
	plan, _ := issue.PlanItemsBody(capture.Plan)
	selected, retired := selectFreshSplitRounds(rounds, issue.MilestonesInPlanOrder(plan))
	tl := activetime.BuildTimeline(selected, nil, capture.Current, parse(capture.Since), parse(capture.Until)).For(capture.Current)
	got := activetime.Split(segments, "737", tl)
	if math.Abs(got.Build-38) > 1e-9 || got.Converge != 0 || got.Tail != 0 || !reflect.DeepEqual(retired, []string{"M2"}) || !tl.Phases[0].Start.Equal(parse("2026-10-20T21:23:47-07:00")) {
		t.Fatalf("capture lost M3: split=%+v phases=%+v retired=%v", got, tl.Phases, retired)
	}
}

// BR-1: requested live milestone authority is independent of open review cuts.
func TestFreshActualRoundsRequestedMilestoneAuthority(t *testing.T) {
	for _, ledger := range []struct {
		name   string
		rounds []activetime.BoundaryRound
	}{
		{"empty", nil},
		{"whole issue", []activetime.BoundaryRound{{Boundary: ""}}},
		{"finalized only", []activetime.BoundaryRound{{Boundary: "M1", Finalizing: true}}},
	} {
		for _, plan := range []struct{ name, body, want string }{
			{"missing", "## Log\nNo Plan\n", "Plan unavailable"},
			{"ambiguous", "## Plan\n- [ ] M3\n## Plan\n- [ ] M3\n", "Plan unavailable"},
			{"absent target", "## Plan\n- [x] M1\n", "absent from the pinned executable Plan"},
			{"executable", "## Plan\n- [ ] M3\n", ""},
		} {
			t.Run(ledger.name+"/"+plan.name, func(t *testing.T) {
				root := testfix.Repo(t, testfix.InitialCommit())
				path := filepath.Join(root, "workshop/issues/000009-x.md")
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("---\nid: 000009\n---\n"+plan.body), 0600); err != nil {
					t.Fatal(err)
				}
				testfix.Git(t, root, "add", ".")
				testfix.Git(t, root, "commit", "-qm", "#9: pinned membership")
				pin := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
				_, _, err := freshActualRounds(context.Background(), root, pin, "9", "M3", ledger.rounds)
				if plan.want == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else if err == nil || !strings.Contains(err.Error(), plan.want) {
					t.Fatalf("want %q, got %v", plan.want, err)
				}
				// Whole-issue accounting and finalized historical membership need no new authority.
				if _, _, err := freshActualRounds(context.Background(), root, pin, "9", "", ledger.rounds); err != nil {
					t.Fatal(err)
				}
				if ledger.name == "finalized only" {
					if _, _, err := freshActualRounds(context.Background(), root, pin, "9", "M1", ledger.rounds); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}
