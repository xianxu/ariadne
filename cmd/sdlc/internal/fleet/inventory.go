package fleet

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// GitRepoPredicate distinguishes eligible Git siblings from ordinary sibling
// directories after project.FleetRepoDirs applies the shared fleet-name filter.
type GitRepoPredicate func(repoDir string) (bool, error)

// RepoIssueLookup resolves an issue ID only within the repository named by
// repoRoot. It is adapted to IssueLookup separately for every worktree.
type RepoIssueLookup func(repoRoot, id string) ([]IssueRecord, error)

// RepoClaimsLookup reads one repository's claims (#288), from the same single
// tracker load as the branch-prefix lookup.
type RepoClaimsLookup func(repoRoot string) RepoClaims

// MachineSource names this machine as claims record it.
type MachineSource func() (MachineIdentity, error)

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
	LookupClaims RepoClaimsLookup
	// Machine is required for claims to be judged; nil leaves the machine
	// unknown, so no claim is ever reported present without it.
	Machine MachineSource
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
func CollectInventory(ctx context.Context, fleetRoot string, options InventoryOptions) (Inventory, error) {
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
		lookupIssues = func(repoRoot, id string) ([]IssueRecord, error) { return LookupRepoIssues(ctx, repoRoot, id) }
	}
	if options.LookupIssues == nil && options.LookupClaims == nil {
		// #290: tracker reads are network round trips; load every tracked
		// repository's records concurrently up front, so the row walk below
		// reads them from the cache.
		warmRecords(ctx, trackedRoots(repoDirs), recordsReadConcurrency, repoRecords)
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
	// #289: slots. Numbered slots' declared dependency clones join the rows
	// (the fleet walk never sees independent clones), reading their tracker
	// through the fleet primary of the same repository.
	hosts := discoverSlots(inventory.Rows, canonicalFleetRoot)
	aliases := map[string]string{}
	decls := collectDependencyRows(&inventory, hosts, canonicalFleetRoot, aliases, func(repoDir string) {
		collectInventoryRepo(&inventory, diagnosticKeys, rowKeys, repoStates, repoDir, options.Git, isGitRepo, loadPolicy, func(repoRoot, id string) ([]IssueRecord, error) {
			return lookupIssues(aliasOf(aliases, repoRoot), id)
		})
	}, options.Git)
	collectClaims(ctx, &inventory, options, aliases)
	inventory.Slots = AssembleSlots(hosts, decls, inventory.Rows)

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

// collectClaims judges every repository's claims against this machine: one
// claims read per repository with rows, the identity once per inventory.
func collectClaims(ctx context.Context, inventory *Inventory, options InventoryOptions, aliases map[string]string) {
	if options.Machine == nil {
		inventory.Machine = MachineFrom(MachineIdentity{}, errors.New("no machine identity source"))
	} else {
		inventory.Machine = MachineFrom(options.Machine())
	}
	lookup := options.LookupClaims
	if lookup == nil {
		lookup = func(repoRoot string) RepoClaims { return LookupRepoClaims(ctx, repoRoot) }
	}
	byRepo := map[string]RepoClaims{}
	for _, row := range inventory.Rows {
		if _, done := byRepo[row.RepoIdentity]; !done {
			byRepo[row.RepoIdentity] = lookup(aliasOf(aliases, row.RepoRoot))
		}
	}
	inventory.Rows, inventory.DanglingClaims = PlaceClaims(inventory.Rows, byRepo, inventory.Machine)
}

// recordsReadConcurrency bounds concurrent tracker reads (#290).
const recordsReadConcurrency = 8

// trackedRoots are the canonical fleet repositories that use the issue tracker
// (cutover marker or fetched tracker; local checks only). A spelling that
// differs from the walk's key only costs that repository a sequential read.
func trackedRoots(repoDirs []string) []string {
	var roots []string
	for _, dir := range repoDirs {
		root, err := canonicalPath(dir)
		if err != nil {
			continue
		}
		if cut, err := tracker.CutOver(root); err == nil && cut {
			roots = append(roots, root)
		}
	}
	return roots
}

// collectDependencyRows walks every numbered slot's declared dependencies and
// collects each present clone that is not already a row (through collect),
// recording an alias to the fleet primary of the same repository when both
// share an origin URL, so the clone reuses that primary's tracker read. It
// returns each host's declaration for AssembleSlots.
func collectDependencyRows(inventory *Inventory, hosts []SlotHost, fleetRoot string, aliases map[string]string, collect func(string), git GitReader) map[string]SlotDeclaration {
	decls := map[string]SlotDeclaration{}
	known := map[string]bool{}
	for _, row := range inventory.Rows {
		known[row.TreePath] = true
	}
	for _, h := range hosts {
		if h.Slot == 0 {
			continue // :0 peers are shared, not this slot's
		}
		members, errs := DeclaredMembers(h.HostPath, h.EnvRoot, readDeclaration, statPath)
		decls[h.HostPath] = SlotDeclaration{Members: members, Errors: errs}
		for _, m := range members {
			if m.State != MemberPresent || known[m.Path] {
				continue
			}
			known[m.Path] = true
			if primary := filepath.Join(fleetRoot, filepath.Base(m.Path)); sameOrigin(git, m.Path, primary) {
				aliases[m.Path] = primary
			}
			collect(m.Path)
		}
	}
	return decls
}

// sameOrigin reports whether two checkouts' origins are one repository —
// compared as publication identities, so GitHub's SSH and HTTPS spellings
// match (one local config read each; any failure means "not the same").
func sameOrigin(git GitReader, a, b string) bool {
	identity := func(dir string) string {
		out, err := git.GitInDir(dir, "config", "--get", "remote.origin.url")
		if err != nil {
			return ""
		}
		id, err := gitx.PublicationRepository(dir, strings.TrimSpace(string(out)))
		if err != nil {
			return ""
		}
		return id
	}
	ia := identity(a)
	return ia != "" && ia == identity(b)
}

// aliasOf is the root whose tracker records stand for repoRoot.
func aliasOf(aliases map[string]string, repoRoot string) string {
	if a, ok := aliases[repoRoot]; ok {
		return a
	}
	return repoRoot
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
	policy := normalizedInventoryCapability(loadPolicy(PolicyDeclarationPath(repoRoot)), repoRoot)
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
		issuesError := ""
		if err != nil {
			issues, issuesError = make([]IssueAssociation, 0), err.Error()
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
			IssuesError:  issuesError,
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
	if err := validatePolicyCapability(capability); err == nil {
		return capability
	}
	path := PolicyDeclarationPath(repoRoot)
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
