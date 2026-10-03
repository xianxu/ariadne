package fleet

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// #289: slot readiness end to end on a real fleet: `prod` declares
// `substrate ../dep`; each numbered slot holds prod's linked worktree and its
// own clone of dep, with one readiness cause placed in the host or the clone.
// The tracked `dep` holds a claim made in slot 1's clone.
func TestFleetInventorySlotReadiness(t *testing.T) {
	fleetRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	env := func(n int) string { return filepath.Join(fleetRoot, "worktree", "prod-slot"+strconv.Itoa(n)) }

	full := issue.Render(issue.ScaffoldSpec{ID: "000001", Title: "Claimed", Today: "2026-10-01"})
	card, _, err := issue.SplitCardWithFormat([]byte(full), "sha1")
	if err != nil {
		t.Fatal(err)
	}
	card = []byte(strings.Replace(string(card), "status: open", "status: working", 1))
	if card, err = issue.SetCardClaimant(card, issue.Claimant{Operator: "op", Machine: meFP, MachineName: "m",
		Worktree: filepath.Join(env(1), "dep"), Repository: "r"}); err != nil {
		t.Fatal(err)
	}
	dep := trackedRepo(t, fleetRoot, "dep", map[string][]byte{tracker.CardPath("000001", "claimed"): card})
	depOrigin := strings.TrimSpace(testfix.Capture(t, dep, "config", "--get", "remote.origin.url"))

	prod := filepath.Join(fleetRoot, "prod")
	testfix.Git(t, "", "init", "-q", "-b", "main", prod)
	configure(t, prod)
	for p, body := range map[string]string{
		"construct/deps":                 "substrate ../dep\n",
		"f.txt":                          "base\n",
		"workshop/issues/000007-open.md": "---\nid: 000007\nstatus: working\n---\n\n# Open\n",
	} {
		writeFile(t, filepath.Join(prod, p), body)
	}
	testfix.Git(t, prod, "add", "-A")
	testfix.Git(t, prod, "commit", "-qm", "base")

	host := func(n int) string { return filepath.Join(env(n), "prod") }
	clone := func(n int) string { return filepath.Join(env(n), "dep") }
	for n := 1; n <= 11; n++ {
		testfix.Git(t, prod, "worktree", "add", "-q", "-b", workspace.RestingBranch(n), host(n))
		if n != 9 { // slot 9's clone is missing
			testfix.Git(t, "", "clone", "-q", depOrigin, clone(n))
		}
	}
	writeFile(t, filepath.Join(host(3), "g.txt"), "unlanded\n") // 3: unlanded commit
	testfix.Git(t, host(3), "add", "g.txt")
	testfix.Git(t, host(3), "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "unlanded")
	testfix.Git(t, host(4), "switch", "-q", "-c", "000007-open")               // 4: zero-commit open-issue branch
	writeFile(t, filepath.Join(clone(5), "stray.txt"), "untracked\n")          // 5: dirty clone
	writeFile(t, filepath.Join(host(6), "f.txt"), "modified\n")                // 6: dirty host
	gitDir, err := workspace.WorktreeGitDir(host(7), workspace.ReadGitPointer) // 7: merge in progress
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(gitDir, "MERGE_HEAD"), strings.TrimSpace(testfix.Capture(t, host(7), "rev-parse", "HEAD"))+"\n")
	testfix.Git(t, clone(8), "switch", "-q", "--detach") // 8: detached clone
	extra := filepath.Join(env(10), "extra")             // 10: undeclared dirty sibling
	testfix.Git(t, "", "init", "-q", extra)
	writeFile(t, filepath.Join(extra, "junk.txt"), "x\n")
	// 11: a declaration that cannot be parsed, committed on the slot's branch
	// (so the checkout is clean, one commit ahead of main).
	writeFile(t, filepath.Join(host(11), "construct", "deps"), "substrate\n")
	testfix.Git(t, host(11), "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qam", "malformed deps")

	freshRecords(t)
	read := map[string]bool{}
	opts := InventoryOptions{Git: execGitReader{},
		Machine: func() (MachineIdentity, error) { return MachineIdentity{Fingerprint: meFP, Name: "here"}, nil },
		LookupIssues: func(root, id string) ([]IssueRecord, error) {
			read[root] = true
			return LookupRepoIssues(context.Background(), root, id)
		},
		LookupClaims: func(root string) RepoClaims { read[root] = true; return LookupRepoClaims(context.Background(), root) },
	}
	inv, err := CollectInventory(context.Background(), fleetRoot, opts)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(inv)
	if err != nil {
		t.Fatal(err)
	}
	var back Inventory
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("not contract JSON: %v", err)
	}

	got := map[string]string{}
	for _, s := range back.Slots {
		var ms []string
		for _, m := range s.Members {
			ms = append(ms, m.Role+":"+string(m.Verdict)+"("+strings.Join(m.Reasons, ",")+")")
		}
		got[s.Address] = string(s.Verdict) + " " + strings.Join(ms, " ")
	}
	for addr, want := range map[string]string{
		"prod:1":  "holds-work host:ready() dependency:holds-work(claimed:dep#000001)",
		"prod:2":  "ready host:ready() dependency:ready()",
		"prod:3":  "holds-work host:holds-work(unlanded-commits) dependency:ready()",
		"prod:4":  "holds-work host:holds-work(open-issue:prod#000007) dependency:ready()",
		"prod:5":  "needs-recovery host:ready() dependency:needs-recovery(dirty)",
		"prod:6":  "needs-recovery host:needs-recovery(dirty) dependency:ready()",
		"prod:7":  "needs-recovery host:needs-recovery(operation:MERGE_HEAD) dependency:ready()",
		"prod:8":  "needs-recovery host:ready() dependency:needs-recovery(detached)",
		"prod:9":  "missing host:ready() dependency:missing(missing)",
		"prod:10": "ready host:ready() dependency:ready()",
		"prod:11": "unknown host:unknown(unlanded-commits,probe:deps)",
		"prod:0":  "ready host:ready()",
		"dep:0":   "ready host:ready()",
	} {
		if got[addr] != want {
			t.Errorf("%s: got %q\n        want %q", addr, got[addr], want)
		}
	}
	if len(back.DanglingClaims) != 0 {
		t.Errorf("the clone's claim must be placed, not dangling: %+v", back.DanglingClaims)
	}
	for root := range read {
		if strings.Contains(root, string(filepath.Separator)+"worktree"+string(filepath.Separator)) {
			t.Errorf("a dependency clone read its own tracker (%s); it must reuse the fleet primary's", root)
		}
	}
}

func writeFile(t *testing.T, p, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// worktreeListFails fails `git worktree list` in one directory, after every
// earlier probe of it succeeded — the failure inventory records as pending.
type worktreeListFails struct {
	GitReader
	dir string
}

func (r worktreeListFails) GitInDir(dir string, args ...string) ([]byte, error) {
	if dir == r.dir && len(args) > 0 && args[0] == "worktree" {
		return nil, errors.New("worktree list failed")
	}
	return r.GitReader.GitInDir(dir, args...)
}

// #289: a declared clone whose worktree list fails surfaces as a repository
// diagnostic (recorded as pending, so it must be flushed after dependency
// collection) and an unknown member — never dropped, never ready.
func TestBrokenDependencyCloneIsReported(t *testing.T) {
	fleetRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	prod := filepath.Join(fleetRoot, "prod")
	testfix.Git(t, "", "init", "-q", "-b", "main", prod)
	configure(t, prod)
	writeFile(t, filepath.Join(prod, "construct", "deps"), "substrate ../dep\n")
	testfix.Git(t, prod, "add", "-A")
	testfix.Git(t, prod, "commit", "-qm", "base")
	host := filepath.Join(fleetRoot, "worktree", "prod-slot1", "prod")
	testfix.Git(t, prod, "worktree", "add", "-q", "-b", workspace.RestingBranch(1), host)
	clone := filepath.Join(fleetRoot, "worktree", "prod-slot1", "dep")
	testfix.Git(t, "", "init", "-q", "-b", "main", clone)
	configure(t, clone)
	testfix.Git(t, clone, "commit", "-q", "--allow-empty", "-m", "base")
	freshRecords(t)
	inv, err := CollectInventory(context.Background(), fleetRoot, InventoryOptions{Git: worktreeListFails{GitReader: execGitReader{}, dir: clone}})
	if err != nil {
		t.Fatal(err)
	}
	reported := false
	for _, d := range inv.Diagnostics {
		reported = reported || (d.RepoPath == clone && d.Stage == "worktrees")
	}
	var dep SlotMember
	for _, s := range inv.Slots {
		if s.Address == "prod:1" && len(s.Members) == 2 {
			dep = s.Members[1]
		}
	}
	if !reported || dep.Verdict != VerdictUnknown {
		t.Fatalf("broken clone: diagnostic %v, member %+v; diagnostics %+v", reported, dep, inv.Diagnostics)
	}
}
