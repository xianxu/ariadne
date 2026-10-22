package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitInit(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
}

func TestTeamNewCreates(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ops-echo")
	var out bytes.Buffer
	if err := teamNew(&out, root, []string{"echo", "pair:7"}); err != nil {
		t.Fatal(err)
	}
	hubRaw, _ := os.ReadFile(filepath.Join(root, ".hub"))
	if string(hubRaw) != "echo\npair:7\n" {
		t.Fatalf(".hub = %q", hubRaw)
	}
	md, _ := os.ReadFile(filepath.Join(root, "team.md"))
	if !strings.Contains(string(md), "type: team") || !strings.Contains(string(md), "ops-echo:0") || !strings.Contains(string(md), "[.hub](.hub)") {
		t.Fatalf("team.md:\n%s", md)
	}
	if _, err := git(root, "rev-parse", "--show-toplevel"); err != nil {
		t.Fatalf("not a git repo: %v", err)
	}
	// .hub carries claim lines only, which Couch's grammar (mirrored) accepts.
	if _, err := parseHub(hubRaw); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "next (yours, not tl-lead's)") {
		t.Fatalf("no next steps: %s", out.String())
	}
}

func TestTeamNewAdoptsAndResumes(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ops")
	gitInit(t, root)
	if err := os.WriteFile(filepath.Join(root, "notes.md"), []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// An interrupted earlier run left .hub only.
	if err := os.WriteFile(filepath.Join(root, ".hub"), []byte("ariadne\npair\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := teamNew(&out, root, []string{"ariadne", "pair"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "wrote team.md") {
		t.Fatalf("out: %s", out.String())
	}
	if b, _ := os.ReadFile(filepath.Join(root, "notes.md")); string(b) != "keep\n" {
		t.Fatal("touched another file")
	}
	// A rerun with everything in place writes nothing and succeeds.
	out.Reset()
	if err := teamNew(&out, root, []string{"ariadne", "pair"}); err != nil || !strings.Contains(out.String(), "wrote none") {
		t.Fatalf("rerun: %v %s", err, out.String())
	}
}

func TestTeamNewRefuses(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ops")
	gitInit(t, root)
	if err := os.WriteFile(filepath.Join(root, ".hub"), []byte("pair\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := teamNew(&out, root, []string{"echo"})
	if err == nil || !strings.Contains(err.Error(), "other content") {
		t.Fatalf("differing .hub: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "team.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("wrote team.md despite refusing")
	}
	for _, tc := range []struct {
		claims []string
		want   string
	}{
		{nil, "at least one --claim"},
		{[]string{"../x"}, "invalid"},
	} {
		if err := teamNew(&out, filepath.Join(t.TempDir(), "x"), tc.claims); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("claims %v: %v", tc.claims, err)
		}
	}
	plain := filepath.Join(t.TempDir(), "plain")
	if err := os.Mkdir(plain, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := teamNew(&out, plain, []string{"echo"}); err == nil || !strings.Contains(err.Error(), "not a git checkout") {
		t.Errorf("a non-git folder: %v", err)
	}
	inner := filepath.Join(t.TempDir(), "outer")
	gitInit(t, inner)
	if err := os.Mkdir(filepath.Join(inner, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := teamNew(&out, filepath.Join(inner, "nested"), []string{"echo"}); err == nil || !strings.Contains(err.Error(), "inside the checkout") {
		t.Errorf("a folder inside a checkout: %v", err)
	}
}

func TestRenderTeams(t *testing.T) {
	hubs := []hub{
		{Name: "ops", Claims: []hubClaim{{"ariadne", -1}, {"pair", -1}}},
		{Name: "ops-bad", Err: errors.New("line 1: invalid")},
		{Name: "ops-echo", Claims: []hubClaim{{"echo", -1}, {"pair", -1}, {"pair", 7}}},
	}
	var out bytes.Buffer
	renderTeams(&out, hubs, "")
	for _, want := range []string{"ops:0  ariadne pair", "ops-bad:0  ?  .hub unreadable", "ops-echo:0  echo pair pair:7", "warning: pair is claimed by ops and ops-echo"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("list lacks %q:\n%s", want, out.String())
		}
	}
	out.Reset()
	renderTeams(&out, nil, "")
	if !strings.Contains(out.String(), "no teams") {
		t.Errorf("an empty workspace printed %q", out.String())
	}
	for repo, want := range map[string]string{
		"ariadne": "ariadne: ops:0",
		"pair":    "pair: tied between ops:0, ops-echo:0",
		"brain":   "brain: unclaimed",
	} {
		out.Reset()
		renderTeams(&out, hubs, repo)
		if !strings.Contains(out.String(), want) {
			t.Errorf("%s: %s", repo, out.String())
		}
	}
	out.Reset()
	renderTeams(&out, hubs, "pair")
	if !strings.Contains(out.String(), "pair:7: ops-echo:0 (an exact slot claim)") {
		t.Errorf("exact claim not shown:\n%s", out.String())
	}
}

func TestUnclaimedIssues(t *testing.T) {
	issues := []RepoIssue{
		{ID: "000700", Status: "open", Title: "old"},
		{ID: "000702", Status: "open", Title: "newest"},
		{ID: "000701", Status: "open", Title: "mine", Owner: "pair:3"},
		{ID: "000703", Status: "working", Title: "started"},
		{ID: "000699", Status: "open", Title: "?", OwnerErr: "tracker unreadable"},
		{ID: "000698", Status: "open", Title: "older"},
	}
	open, unknown := unclaimedIssues(issues, 2)
	if len(open) != 2 || open[0].ID != "000702" || open[1].ID != "000700" || len(unknown) != 1 {
		t.Fatalf("open %+v unknown %+v", open, unknown)
	}
}

func TestTeamIssues(t *testing.T) {
	f := &fakeSDLC{
		fleet: Fleet{Slots: []FleetSlot{{Repo: "pair", Slot: 0, HostPath: "/w/pair"}, {Repo: "echo", Slot: 0, HostPath: "/w/echo"}}},
		issues: map[string][]RepoIssue{
			"/w/pair": {{ID: "000702", Status: "open", Title: "crash on start"}, {ID: "000699", Status: "open", Title: "x", OwnerErr: "unreadable"}},
		},
		issueErr: map[string]error{"/w/echo": errors.New("tracker down")},
	}
	h := hub{Name: "ops", Claims: []hubClaim{{"pair", -1}, {"echo", -1}, {"ariadne", 3}, {"gone", -1}}}
	var out bytes.Buffer
	teamIssues(&out, f, "/w/ops", h, 10)
	for _, want := range []string{"pair: 1 unclaimed", "pair#702  crash on start", "pair#699  owner unreadable (unreadable)", "echo: ? (tracker down)", "gone: ? ("} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("lacks %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "ariadne") {
		t.Errorf("a slot claim is not a repo the team owns:\n%s", out.String())
	}
}

func TestDecodeRepoIssues(t *testing.T) {
	raw := []byte(`{"workspace":{"x":1},"issues":[{"id":"000468","status":"open","title":"t","owner":{"operator":"o","workspace":"ariadne:4","worktree":"/w"},"plan_total":3,"new_field":true},{"id":"000469","status":"open","title":"u","owner_error":"bad card"}],"drift":[]}`)
	got, err := decodeRepoIssues(raw)
	if err != nil || len(got) != 2 || got[0].Owner != "ariadne:4" || got[1].OwnerErr != "bad card" || got[1].Owner != "" {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := decodeRepoIssues([]byte("not json")); err == nil {
		t.Fatal("accepted garbage")
	}
}

// Live: the real sdlc's state JSON decodes (TL_LEAD_LIVE=1, sdlc on PATH).
func TestRepoIssuesLive(t *testing.T) {
	if os.Getenv("TL_LEAD_LIVE") != "1" {
		t.Skip("set TL_LEAD_LIVE=1 to read the real sdlc state")
	}
	top, err := git(".", "rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatal(err)
	}
	issues, err := execSDLC{bin: "sdlc"}.RepoIssues(top)
	if err != nil || len(issues) == 0 || issues[0].Status == "" {
		t.Fatalf("%d issues, err %v", len(issues), err)
	}
}

// #543: the delivered operator/TL setup procedure must end with explicit init.
// Merely documenting the command in sdlc help leaves new team repos legacy.
func TestTeamNewSetupEndsWithExplicitTrackerInit(t *testing.T) {
	var out bytes.Buffer
	root := filepath.Join(t.TempDir(), "ops-new")
	if err := teamNew(&out, root, []string{"product"}); err != nil {
		t.Fatal(err)
	}
	for _, procedure := range []string{out.String(), strings.Join(strings.Fields(helpOutput(t, "--help")), " ")} {
		link, compile := strings.Index(procedure, "weave link github.com/xianxu/ariadne"), strings.Index(procedure, "weave compile")
		push, init := strings.Index(procedure, "push main"), strings.Index(procedure, "sdlc issue init")
		if link < 0 || compile <= link || push <= compile || init <= push {
			t.Fatalf("new-team setup needs ordered base setup, pushed main, then explicit init: %s", procedure)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "workshop/issue-tracker.json")); !os.IsNotExist(err) {
		t.Fatalf("team new must only instruct initialization, marker error=%v", err)
	}
}
