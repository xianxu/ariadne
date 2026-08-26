package fleet

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

func TestInventory_MultiRepoMutationAndFaultIsolation(t *testing.T) {
	fleetRoot := t.TempDir()
	fake := NewFakeGit()
	alpha := addInventoryFakeRepo(t, fake, fleetRoot, "alpha", "000123-alpha")
	zeta := addInventoryFakeRepo(t, fake, fleetRoot, "zeta", "main")
	if err := os.Mkdir(filepath.Join(fleetRoot, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}

	validPolicy := DecodePolicy("alpha-policy", []byte(`{
		"version":1,
		"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"bounded","limit":1},"onCapacity":"reject"}
	}`))
	invalidPolicy := DecodePolicy("zeta-policy", []byte(`{"version":1,"admission":{}}`))
	lookups := map[string]map[string][]IssueRecord{
		alpha.primary: {"000123": {{Ref: "alpha#123", DeclaredStatus: "working"}}},
	}
	options := InventoryOptions{
		Git: fake,
		LoadPolicy: func(path string) PolicyCapability {
			switch filepath.Dir(filepath.Dir(path)) {
			case alpha.primary:
				return validPolicy
			case zeta.primary:
				return invalidPolicy
			default:
				return policyFailure(DiagnosticMissingPolicy, path, nil, "missing")
			}
		},
		LookupIssues: func(repoRoot, id string) ([]IssueRecord, error) {
			return append([]IssueRecord(nil), lookups[repoRoot][id]...), nil
		},
	}

	first, err := CollectInventory(fleetRoot, options)
	if err != nil {
		t.Fatal(err)
	}
	assertInventoryCollectionsNonNil(t, first)
	if got, want := inventoryRowKeys(first), []string{
		alpha.common + "\x00" + alpha.primary,
		zeta.common + "\x00" + zeta.primary,
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("initial inventory row order = %#v, want %#v", got, want)
	}
	if got := first.Rows[0].Issues; !reflect.DeepEqual(got, []IssueAssociation{{Ref: "alpha#123", DeclaredStatus: "working", Provenance: IssueProvenanceBranchPrefix}}) {
		t.Fatalf("alpha issues = %#v, want branch-prefix provenance", got)
	}
	if !first.Rows[0].Policy.OK || first.Rows[0].Policy.Value == nil || first.Rows[0].Policy.Value.KeyKind != "repo" {
		t.Fatalf("alpha policy = %+v, want declaration capability", first.Rows[0].Policy)
	}
	if first.Rows[1].Policy.OK || first.Rows[1].Policy.Diagnostic == nil || first.Rows[1].Policy.Diagnostic.Code != DiagnosticInvalidPolicy {
		t.Fatalf("zeta policy = %+v, want row-scoped invalid-policy diagnostic", first.Rows[1].Policy)
	}
	if len(first.Diagnostics) != 0 {
		t.Fatalf("policy failure leaked into top-level repo diagnostics: %+v", first.Diagnostics)
	}
	assertInventoryMeasuredOnly(t, first)

	linked := filepath.Join(fleetRoot, "worktree", "alpha-linked")
	if err := os.MkdirAll(linked, 0o755); err != nil {
		t.Fatal(err)
	}
	mustFakeMutation(t, fake.SetRef(alpha.common, "refs/heads/000124-linked", alpha.head))
	mustFakeMutation(t, fake.AddWorktree(alpha.common, gitx.Worktree{Path: linked, HEAD: alpha.head, Branch: "000124-linked"}))
	mustFakeMutation(t, fake.SetDirty(linked, []FakeGitStatusEntry{{Code: "??", Path: "new.txt"}}))
	lookups[alpha.primary]["000124"] = []IssueRecord{{Ref: "alpha#124", DeclaredStatus: "open"}}

	second, err := CollectInventory(fleetRoot, options)
	if err != nil {
		t.Fatal(err)
	}
	assertInventoryCollectionsNonNil(t, second)
	if len(second.Rows) != 3 {
		t.Fatalf("inventory after mutation has %d rows, want 3: %+v", len(second.Rows), second.Rows)
	}
	linkedPath := canonicalContractPath(t, linked)
	linkedRow := inventoryRowByPath(t, second, linkedPath)
	if linkedRow.Facts.DirtyCount == nil || *linkedRow.Facts.DirtyCount != 1 {
		t.Fatalf("linked dirty facts = %+v, want one measured entry", linkedRow.Facts)
	}
	if got := linkedRow.Issues; len(got) != 1 || got[0].Ref != "alpha#124" {
		t.Fatalf("linked issues = %+v, want alpha#124", got)
	}

	faults := &inventoryFaultGit{GitReader: fake, failures: map[string]error{
		zeta.primary + "\x00worktree list --porcelain -z":                    errors.New("corrupt worktree metadata"),
		alpha.primary + "\x00status --porcelain=v1 -z --untracked-files=all": errors.New("index locked"),
	}}
	options.Git = faults
	options.LookupIssues = func(repoRoot, id string) ([]IssueRecord, error) {
		if repoRoot == alpha.primary && id == "000124" {
			return nil, errors.New("issue store unavailable")
		}
		return append([]IssueRecord(nil), lookups[repoRoot][id]...), nil
	}
	faulted, err := CollectInventory(fleetRoot, options)
	if err != nil {
		t.Fatal(err)
	}
	assertInventoryCollectionsNonNil(t, faulted)
	if len(faulted.Rows) != 2 {
		t.Fatalf("faulted inventory has %d rows, want unaffected alpha rows: %+v", len(faulted.Rows), faulted.Rows)
	}
	if got, want := diagnosticStages(faulted), []string{"facts", "issues", "worktrees"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("diagnostic stages = %#v, want %#v; diagnostics=%+v", got, want, faulted.Diagnostics)
	}
	for _, stage := range []string{"facts", "issues", "worktrees"} {
		if countInventoryDiagnostics(faulted, stage) != 1 {
			t.Fatalf("stage %q diagnostics = %+v, want exactly one", stage, faulted.Diagnostics)
		}
	}
	if inventoryRowByPath(t, faulted, alpha.primary).Facts.Available {
		t.Fatal("failed facts were collapsed to an available/empty measurement")
	}
	if got := inventoryRowByPath(t, faulted, linkedPath).Issues; got == nil || len(got) != 0 {
		t.Fatalf("failed issue lookup issues = %#v, want non-nil empty", got)
	}
}

func TestInventory_RejectsMissingRequiredDependenciesAndPreservesEmptyArrays(t *testing.T) {
	if _, err := CollectInventory(t.TempDir(), InventoryOptions{}); err == nil {
		t.Fatal("CollectInventory accepted a nil GitReader")
	}

	empty, err := CollectInventory(t.TempDir(), InventoryOptions{Git: NewFakeGit()})
	if err != nil {
		t.Fatal(err)
	}
	assertInventoryCollectionsNonNil(t, empty)
	if len(empty.Rows) != 0 || len(empty.Diagnostics) != 0 {
		t.Fatalf("empty inventory = %+v", empty)
	}
}

func TestInventory_LinkedSiblingUsesPrimaryPolicyAndDoesNotDuplicateRepo(t *testing.T) {
	fleetRoot := t.TempDir()
	fake := NewFakeGit()
	repo := addInventoryFakeRepo(t, fake, fleetRoot, "z-primary", "main")
	linked := filepath.Join(fleetRoot, "a-linked")
	if err := os.MkdirAll(linked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(linked, ".git"), []byte("gitdir: elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustFakeMutation(t, fake.SetRef(repo.common, "refs/heads/linked", repo.head))
	mustFakeMutation(t, fake.AddWorktree(repo.common, gitx.Worktree{Path: linked, HEAD: repo.head, Branch: "linked"}))

	var loaded []string
	inventory, err := CollectInventory(fleetRoot, InventoryOptions{
		Git: fake,
		LoadPolicy: func(path string) PolicyCapability {
			loaded = append(loaded, path)
			return DecodePolicy(path, []byte(`{"version":1,"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"unbounded"}}}`))
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(inventory.Rows) != 2 {
		t.Fatalf("linked sibling produced %d rows, want one canonical repo's two trees: %+v", len(inventory.Rows), inventory.Rows)
	}
	wantPolicyPath := filepath.Join(repo.primary, ".sdlc", "fleet.json")
	if !reflect.DeepEqual(loaded, []string{wantPolicyPath}) {
		t.Fatalf("policy loads = %#v, want primary-only %#v", loaded, []string{wantPolicyPath})
	}
	for _, row := range inventory.Rows {
		if row.RepoRoot != repo.primary || !row.Policy.OK {
			t.Fatalf("linked sibling row = %+v, want primary root and valid policy", row)
		}
	}
}

func TestInventory_LinkedSiblingWorktreeFailureIsReportedOncePerRepo(t *testing.T) {
	fleetRoot := t.TempDir()
	fake := NewFakeGit()
	repo := addInventoryFakeRepo(t, fake, fleetRoot, "z-primary", "main")
	linked := filepath.Join(fleetRoot, "a-linked")
	if err := os.MkdirAll(linked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(linked, ".git"), []byte("gitdir: elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustFakeMutation(t, fake.SetRef(repo.common, "refs/heads/linked", repo.head))
	mustFakeMutation(t, fake.AddWorktree(repo.common, gitx.Worktree{Path: linked, HEAD: repo.head, Branch: "linked"}))

	command := "worktree list --porcelain -z"
	faults := &inventoryFaultGit{GitReader: fake, failures: map[string]error{
		canonicalContractPath(t, linked) + "\x00" + command: errors.New("linked list failure"),
		repo.primary + "\x00" + command:                     errors.New("primary list failure"),
	}}
	inventory, err := CollectInventory(fleetRoot, InventoryOptions{Git: faults})
	if err != nil {
		t.Fatal(err)
	}
	if len(inventory.Rows) != 0 {
		t.Fatalf("failed worktree listing produced rows: %+v", inventory.Rows)
	}
	if len(inventory.Diagnostics) != 1 || inventory.Diagnostics[0].RepoIdentity != repo.common || inventory.Diagnostics[0].Stage != "worktrees" {
		t.Fatalf("shared-repo failure diagnostics = %+v, want exactly one canonical worktrees diagnostic", inventory.Diagnostics)
	}
	if got := faults.commandCalls(command); got != 2 {
		t.Fatalf("shared-repo worktree list calls = %d, want both aliases tried", got)
	}
}

func TestInventory_LinkedAliasFailureRetriesHealthyPrimary(t *testing.T) {
	fleetRoot := t.TempDir()
	fake := NewFakeGit()
	repo := addInventoryFakeRepo(t, fake, fleetRoot, "z-primary", "main")
	linked := filepath.Join(fleetRoot, "a-linked")
	if err := os.MkdirAll(linked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(linked, ".git"), []byte("gitdir: elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustFakeMutation(t, fake.SetRef(repo.common, "refs/heads/linked", repo.head))
	mustFakeMutation(t, fake.AddWorktree(repo.common, gitx.Worktree{Path: linked, HEAD: repo.head, Branch: "linked"}))

	command := "worktree list --porcelain -z"
	faults := &inventoryFaultGit{GitReader: fake, failures: map[string]error{
		canonicalContractPath(t, linked) + "\x00" + command: errors.New("linked list failure"),
	}}
	inventory, err := CollectInventory(fleetRoot, InventoryOptions{Git: faults})
	if err != nil {
		t.Fatal(err)
	}
	if len(inventory.Rows) != 2 {
		t.Fatalf("healthy primary retry produced %d rows, want 2: %+v", len(inventory.Rows), inventory.Rows)
	}
	if countInventoryDiagnostics(inventory, "worktrees") != 0 {
		t.Fatalf("recovered alias failure left a false diagnostic: %+v", inventory.Diagnostics)
	}
	if got := faults.commandCalls(command); got != 2 {
		t.Fatalf("worktree list calls = %d, want failed alias plus healthy primary", got)
	}
}

func TestInventory_TreeFailuresArePerCanonicalTreeNotAlias(t *testing.T) {
	fleetRoot := t.TempDir()
	fake := NewFakeGit()
	repo := addInventoryFakeRepo(t, fake, fleetRoot, "z-primary", "000123-primary")
	linked := filepath.Join(fleetRoot, "a-linked")
	if err := os.MkdirAll(linked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(linked, ".git"), []byte("gitdir: elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustFakeMutation(t, fake.SetRef(repo.common, "refs/heads/000124-linked", repo.head))
	mustFakeMutation(t, fake.AddWorktree(repo.common, gitx.Worktree{Path: linked, HEAD: repo.head, Branch: "000124-linked"}))
	linked = canonicalContractPath(t, linked)

	statusCommand := "status --porcelain=v1 -z --untracked-files=all"
	faults := &inventoryFaultGit{GitReader: fake, failures: map[string]error{
		repo.primary + "\x00" + statusCommand: errors.New("primary status failure"),
		linked + "\x00" + statusCommand:       errors.New("linked status failure"),
	}}
	inventory, err := CollectInventory(fleetRoot, InventoryOptions{
		Git: faults,
		LookupIssues: func(_ string, id string) ([]IssueRecord, error) {
			return nil, fmt.Errorf("issue %s unavailable", id)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(inventory.Rows) != 2 {
		t.Fatalf("tree failures erased rows: %+v", inventory.Rows)
	}
	for _, stage := range []string{"facts", "issues"} {
		if got := countInventoryDiagnostics(inventory, stage); got != 2 {
			t.Fatalf("%s diagnostics = %d, want one per canonical tree: %+v", stage, got, inventory.Diagnostics)
		}
		paths := map[string]bool{}
		for _, diagnostic := range inventory.Diagnostics {
			if diagnostic.Stage == stage {
				paths[diagnostic.TreePath] = true
			}
		}
		if !paths[repo.primary] || !paths[linked] || len(paths) != 2 {
			t.Fatalf("%s diagnostic tree paths = %#v, want primary and linked exactly once", stage, paths)
		}
	}
}

func TestInventory_WorktreeCanonicalizationFailuresArePerListedTreeNotAlias(t *testing.T) {
	fleetRoot := t.TempDir()
	fake := NewFakeGit()
	repo := addInventoryFakeRepo(t, fake, fleetRoot, "z-primary", "main")
	alias := filepath.Join(fleetRoot, "a-linked")
	if err := os.MkdirAll(alias, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(alias, ".git"), []byte("gitdir: elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustFakeMutation(t, fake.SetRef(repo.common, "refs/heads/alias", repo.head))
	mustFakeMutation(t, fake.AddWorktree(repo.common, gitx.Worktree{Path: alias, HEAD: repo.head, Branch: "alias"}))

	missingA := filepath.Join(fleetRoot, "missing-a")
	missingB := filepath.Join(fleetRoot, "missing-b")
	reason := "gitdir file points to non-existent location"
	for i, path := range []string{missingB, missingA} {
		branch := fmt.Sprintf("stale-%d", i)
		mustFakeMutation(t, fake.SetRef(repo.common, "refs/heads/"+branch, repo.head))
		mustFakeMutation(t, fake.AddWorktree(repo.common, gitx.Worktree{Path: path, HEAD: repo.head, Branch: branch, Prunable: &reason}))
	}

	inventory, err := CollectInventory(fleetRoot, InventoryOptions{Git: fake})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, diagnostic := range inventory.Diagnostics {
		if diagnostic.Stage == "worktrees" {
			got = append(got, diagnostic.TreePath)
		}
	}
	want := []string{
		filepath.Join(filepath.Dir(repo.primary), filepath.Base(missingA)),
		filepath.Join(filepath.Dir(repo.primary), filepath.Base(missingB)),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("per-tree canonicalization diagnostics = %#v, want stable %#v; all=%+v", got, want, inventory.Diagnostics)
	}
}

func TestInventory_HeadExistsOnlyInMeasuredFacts(t *testing.T) {
	if _, duplicate := reflect.TypeOf(TreeRow{}).FieldByName("HEAD"); duplicate {
		t.Fatal("TreeRow duplicates measured Facts.Head")
	}
	fleetRoot := t.TempDir()
	fake := NewFakeGit()
	addInventoryFakeRepo(t, fake, fleetRoot, "repo", "main")
	inventory, err := CollectInventory(fleetRoot, InventoryOptions{Git: fake})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(inventory.Rows[0])
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(raw), `"head"`); got != 1 {
		t.Fatalf("row JSON contains %d head authorities, want facts.head only: %s", got, raw)
	}
}

type inventoryFakeRepo struct {
	primary string
	common  string
	head    string
}

func addInventoryFakeRepo(t *testing.T, fake *FakeGit, fleetRoot, name, branch string) inventoryFakeRepo {
	t.Helper()
	primary := filepath.Join(fleetRoot, name)
	common := filepath.Join(primary, ".git")
	if err := os.MkdirAll(common, 0o755); err != nil {
		t.Fatal(err)
	}
	primary = canonicalContractPath(t, primary)
	common = canonicalContractPath(t, common)
	head := strings.Repeat(fmt.Sprintf("%x", len(name)), 40)[:40]
	repo := &FakeGitRepo{
		CommonDir: common, PrimaryRoot: primary,
		Worktrees: []gitx.Worktree{{Path: primary, HEAD: head, Branch: branch}},
		Refs: map[string]string{
			"refs/heads/" + branch:     head,
			"refs/remotes/origin/main": head,
		},
		Commits: map[string]FakeGitCommit{head: {CommittedAt: time.Date(2026, 8, len(name), 12, 0, 0, 0, time.UTC)}},
		Dirty:   map[string][]FakeGitStatusEntry{},
	}
	if branch != "main" {
		repo.Refs["refs/heads/main"] = head
	}
	mustFakeMutation(t, fake.AddRepo(repo))
	return inventoryFakeRepo{primary: primary, common: common, head: head}
}

type inventoryFaultGit struct {
	GitReader
	failures map[string]error
	calls    []string
}

func (r *inventoryFaultGit) GitInDir(dir string, args ...string) ([]byte, error) {
	canonical, err := canonicalPath(dir)
	if err == nil {
		call := canonical + "\x00" + strings.Join(args, " ")
		r.calls = append(r.calls, call)
		if failure := r.failures[call]; failure != nil {
			return []byte("fault output"), failure
		}
	}
	return r.GitReader.GitInDir(dir, args...)
}

func (r *inventoryFaultGit) commandCalls(command string) int {
	count := 0
	for _, call := range r.calls {
		if strings.HasSuffix(call, "\x00"+command) {
			count++
		}
	}
	return count
}

func assertInventoryCollectionsNonNil(t *testing.T, inventory Inventory) {
	t.Helper()
	if inventory.Rows == nil || inventory.Diagnostics == nil {
		t.Fatalf("inventory collections must be non-nil: %+v", inventory)
	}
	for _, row := range inventory.Rows {
		if row.Issues == nil {
			t.Fatalf("row issues must be non-nil: %+v", row)
		}
	}
}

func assertInventoryMeasuredOnly(t *testing.T, inventory Inventory) {
	t.Helper()
	text := fmt.Sprintf("%#v", inventory)
	for _, forbidden := range []string{"Cold", "Drift", "AdmissionKey"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("inventory contains derived/resolved field %q: %s", forbidden, text)
		}
	}
}

func inventoryRowKeys(inventory Inventory) []string {
	keys := make([]string, len(inventory.Rows))
	for i, row := range inventory.Rows {
		keys[i] = row.RepoIdentity + "\x00" + row.TreePath
	}
	return keys
}

func inventoryRowByPath(t *testing.T, inventory Inventory, path string) TreeRow {
	t.Helper()
	for _, row := range inventory.Rows {
		if row.TreePath == path {
			return row
		}
	}
	t.Fatalf("inventory has no row for %q: %+v", path, inventory.Rows)
	return TreeRow{}
}

func diagnosticStages(inventory Inventory) []string {
	stages := make([]string, len(inventory.Diagnostics))
	for i, diagnostic := range inventory.Diagnostics {
		stages[i] = diagnostic.Stage
	}
	return stages
}

func countInventoryDiagnostics(inventory Inventory, stage string) int {
	count := 0
	for _, diagnostic := range inventory.Diagnostics {
		if diagnostic.Stage == stage {
			count++
		}
	}
	return count
}
