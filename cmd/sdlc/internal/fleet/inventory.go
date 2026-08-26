package fleet

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
)

const fleetPolicyPath = ".sdlc/fleet.json"

// GitRepoPredicate distinguishes eligible Git siblings from ordinary sibling
// directories after project.FleetRepoDirs applies the shared fleet-name filter.
type GitRepoPredicate func(repoDir string) (bool, error)

// RepoIssueLookup resolves an issue ID only within the repository named by
// repoRoot. It is adapted to IssueLookup separately for every worktree.
type RepoIssueLookup func(repoRoot, id string) ([]IssueRecord, error)

// PolicyLoader loads one repository declaration. A nil loader uses the shared
// strict filesystem loader.
type PolicyLoader func(declarationPath string) PolicyCapability

// InventoryOptions contains only IO seams. Collection and ordering remain in
// CollectInventory so fake and real adapters exercise the same assembly.
type InventoryOptions struct {
	Git          GitReader
	IsGitRepo    GitRepoPredicate
	LoadPolicy   PolicyLoader
	LookupIssues RepoIssueLookup
}

// FilesystemGitRepo recognizes ordinary and linked-worktree checkouts through
// their .git directory/file marker, and bare repositories through Git's HEAD +
// objects layout. It does not execute Git; malformed candidates are reported
// later through the shared GitReader boundary.
func FilesystemGitRepo(repoDir string) (bool, error) {
	_, err := os.Stat(filepath.Join(repoDir, ".git"))
	if err == nil {
		return true, nil
	}
	if !os.IsNotExist(err) {
		return false, err
	}
	head, err := os.Stat(filepath.Join(repoDir, "HEAD"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	objects, err := os.Stat(filepath.Join(repoDir, "objects"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !head.IsDir() && objects.IsDir(), nil
}

// CollectInventory enumerates every eligible Git sibling under an already
// normalized fleet root. Repository failures are recorded and isolated; only a
// fleet-root enumeration failure prevents returning a complete observation.
func CollectInventory(fleetRoot string, options InventoryOptions) (Inventory, error) {
	inventory := Inventory{Rows: make([]TreeRow, 0), Diagnostics: make([]RepoDiagnostic, 0)}
	if options.Git == nil {
		return inventory, fmt.Errorf("collect fleet inventory: nil Git reader")
	}
	canonicalFleetRoot, err := canonicalPath(fleetRoot)
	if err != nil {
		return inventory, fmt.Errorf("collect fleet inventory root %q: %w", fleetRoot, err)
	}
	repoDirs, err := project.FleetRepoDirs(canonicalFleetRoot)
	if err != nil {
		return inventory, fmt.Errorf("enumerate fleet repositories under %q: %w", canonicalFleetRoot, err)
	}

	isGitRepo := options.IsGitRepo
	if isGitRepo == nil {
		isGitRepo = FilesystemGitRepo
	}
	loadPolicy := options.LoadPolicy
	if loadPolicy == nil {
		loadPolicy = LoadPolicyFile
	}
	lookupIssues := options.LookupIssues
	if lookupIssues == nil {
		lookupIssues = LookupRepoIssues
	}
	diagnosticKeys := make(map[string]bool)
	rowKeys := make(map[string]bool)
	repoStates := make(map[string]*inventoryRepoState)
	for _, repoDir := range repoDirs {
		collectInventoryRepo(&inventory, diagnosticKeys, rowKeys, repoStates, repoDir, options.Git, isGitRepo, loadPolicy, lookupIssues)
	}
	for _, state := range repoStates {
		if !state.complete && state.pending != nil {
			appendRepoDiagnostic(&inventory, diagnosticKeys, *state.pending)
		}
	}

	sort.Slice(inventory.Rows, func(i, j int) bool {
		if inventory.Rows[i].RepoIdentity != inventory.Rows[j].RepoIdentity {
			return inventory.Rows[i].RepoIdentity < inventory.Rows[j].RepoIdentity
		}
		return inventory.Rows[i].TreePath < inventory.Rows[j].TreePath
	})
	sort.Slice(inventory.Diagnostics, func(i, j int) bool {
		left, right := inventory.Diagnostics[i], inventory.Diagnostics[j]
		if left.RepoIdentity != right.RepoIdentity {
			return left.RepoIdentity < right.RepoIdentity
		}
		if left.RepoPath != right.RepoPath {
			return left.RepoPath < right.RepoPath
		}
		if left.Stage != right.Stage {
			return left.Stage < right.Stage
		}
		if left.TreePath != right.TreePath {
			return left.TreePath < right.TreePath
		}
		return left.Message < right.Message
	})
	return inventory, nil
}

type inventoryRepoState struct {
	complete bool
	pending  *RepoDiagnostic
}

func collectInventoryRepo(inventory *Inventory, diagnosticKeys, rowKeys map[string]bool, repoStates map[string]*inventoryRepoState, repoDir string, git GitReader, isGitRepo GitRepoPredicate, loadPolicy PolicyLoader, lookupIssues RepoIssueLookup) {
	repoPath, err := canonicalPath(repoDir)
	if err != nil {
		appendRepoDiagnostic(inventory, diagnosticKeys, RepoDiagnostic{RepoPath: repoDir, Stage: "git", Message: fmt.Sprintf("canonicalize repository: %v", err)})
		return
	}
	eligible, err := isGitRepo(repoPath)
	if err != nil {
		appendRepoDiagnostic(inventory, diagnosticKeys, RepoDiagnostic{RepoPath: repoPath, Stage: "git", Message: fmt.Sprintf("identify Git repository: %v", err)})
		return
	}
	if !eligible {
		return
	}

	commonOut, err := git.GitInDir(repoPath, "rev-parse", "--git-common-dir")
	if err != nil {
		appendRepoDiagnostic(inventory, diagnosticKeys, RepoDiagnostic{RepoPath: repoPath, Stage: "git", Message: gitFactFailure("rev-parse --git-common-dir", err, commonOut)})
		return
	}
	repoIdentity, err := canonicalGitOutputPath(repoPath, commonOut)
	if err != nil {
		appendRepoDiagnostic(inventory, diagnosticKeys, RepoDiagnostic{RepoPath: repoPath, Stage: "git", Message: fmt.Sprintf("canonicalize Git common directory: %v", err)})
		return
	}
	state := repoStates[repoIdentity]
	if state == nil {
		state = &inventoryRepoState{}
		repoStates[repoIdentity] = state
	}
	if state.complete {
		return
	}

	porcelain, err := git.GitInDir(repoPath, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		state.recordPending(RepoDiagnostic{RepoIdentity: repoIdentity, RepoPath: repoPath, Stage: "worktrees", Message: gitFactFailure("worktree list --porcelain -z", err, porcelain)})
		return
	}
	worktrees, err := gitx.ParseWorktrees(porcelain)
	if err != nil {
		state.recordPending(RepoDiagnostic{RepoIdentity: repoIdentity, RepoPath: repoPath, Stage: "worktrees", Message: fmt.Sprintf("parse git worktree list: %v", err)})
		return
	}
	if len(worktrees) == 0 {
		state.recordPending(RepoDiagnostic{RepoIdentity: repoIdentity, RepoPath: repoPath, Stage: "worktrees", Message: "git worktree list contains no worktrees"})
		return
	}
	primaryRoot, err := canonicalListedWorktreePath(repoPath, worktrees[0].Path)
	if err != nil {
		state.recordPending(RepoDiagnostic{RepoIdentity: repoIdentity, RepoPath: repoPath, Stage: "worktrees", Message: fmt.Sprintf("canonicalize primary worktree %q: %v", worktrees[0].Path, err)})
		return
	}
	state.complete = true
	state.pending = nil
	canonicalWorktrees := make([]gitx.Worktree, 0, len(worktrees))
	for _, worktree := range worktrees {
		treePath, err := canonicalListedWorktreePath(repoPath, worktree.Path)
		if err != nil {
			appendRepoDiagnostic(inventory, diagnosticKeys, RepoDiagnostic{RepoIdentity: repoIdentity, RepoPath: repoPath, TreePath: worktree.Path, Stage: "worktrees", Message: fmt.Sprintf("canonicalize worktree %q: %v", worktree.Path, err)})
			continue
		}
		worktree.Path = treePath
		canonicalWorktrees = append(canonicalWorktrees, worktree)
	}
	if len(canonicalWorktrees) == 0 {
		return
	}
	sort.Slice(canonicalWorktrees, func(i, j int) bool { return canonicalWorktrees[i].Path < canonicalWorktrees[j].Path })

	repoRoot := primaryRoot
	policy := normalizedInventoryCapability(loadPolicy(filepath.Join(repoRoot, filepath.FromSlash(fleetPolicyPath))), repoRoot)
	for _, worktree := range canonicalWorktrees {
		rowKey := repoIdentity + "\x00" + worktree.Path
		if rowKeys[rowKey] {
			continue
		}
		rowKeys[rowKey] = true

		facts := CollectFacts(git, worktree.Path)
		if facts.Error != "" {
			appendRepoDiagnostic(inventory, diagnosticKeys, RepoDiagnostic{RepoIdentity: repoIdentity, RepoPath: repoRoot, TreePath: worktree.Path, Stage: "facts", Message: facts.Error})
		} else if facts.BaseError != "" {
			appendRepoDiagnostic(inventory, diagnosticKeys, RepoDiagnostic{RepoIdentity: repoIdentity, RepoPath: repoRoot, TreePath: worktree.Path, Stage: "facts", Message: facts.BaseError})
		}

		issues, err := AssociateBranchIssue(worktree.Branch, func(id string) ([]IssueRecord, error) {
			return lookupIssues(repoRoot, id)
		})
		if err != nil {
			issues = make([]IssueAssociation, 0)
			appendRepoDiagnostic(inventory, diagnosticKeys, RepoDiagnostic{RepoIdentity: repoIdentity, RepoPath: repoRoot, TreePath: worktree.Path, Stage: "issues", Message: err.Error()})
		}

		inventory.Rows = append(inventory.Rows, TreeRow{
			RepoIdentity: repoIdentity,
			RepoRoot:     repoRoot,
			TreePath:     worktree.Path,
			Branch:       worktree.Branch,
			Detached:     worktree.Detached,
			Bare:         worktree.Bare,
			Locked:       cloneInventoryString(worktree.Locked),
			Prunable:     cloneInventoryString(worktree.Prunable),
			Facts:        facts,
			Issues:       issues,
			Policy:       policy,
		})
	}
}

func (state *inventoryRepoState) recordPending(diagnostic RepoDiagnostic) {
	if state.pending == nil {
		copy := diagnostic
		state.pending = &copy
	}
}

func normalizedInventoryCapability(capability PolicyCapability, repoRoot string) PolicyCapability {
	if err := validateEnvelope(capability.OK, capability.Value != nil, capability.Diagnostic != nil); err == nil {
		return capability
	}
	path := filepath.Join(repoRoot, filepath.FromSlash(fleetPolicyPath))
	return policyFailure(DiagnosticInvalidPolicy, path, nil, "policy loader returned an invalid capability envelope")
}

func appendRepoDiagnostic(inventory *Inventory, seen map[string]bool, diagnostic RepoDiagnostic) {
	repoKey := diagnostic.RepoIdentity
	if repoKey == "" {
		repoKey = diagnostic.RepoPath
	}
	parts := []string{repoKey, diagnostic.Stage}
	if diagnostic.TreePath != "" {
		parts = append(parts, diagnostic.TreePath)
	}
	key := strings.Join(parts, "\x00")
	if seen[key] {
		return
	}
	seen[key] = true
	inventory.Diagnostics = append(inventory.Diagnostics, diagnostic)
}

func cloneInventoryString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
