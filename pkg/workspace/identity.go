package workspace

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

type Identity struct {
	EnvironmentRoot string           `json:"environment_root"`
	EnvironmentHost *EnvironmentHost `json:"environment_host,omitempty"`
	SchemaVersion   int              `json:"schema_version"`
	Repo            string           `json:"repo"`
	RepoIdentity    string           `json:"repo_identity"`
	PrimaryRoot     string           `json:"primary_root"`
	FleetRoot       string           `json:"fleet_root"`
	WorktreeRoot    string           `json:"worktree_root"`
	Kind            string           `json:"kind"`
	Address         *string          `json:"address"`
	Slot            *int             `json:"slot"`
	Branch          *string          `json:"branch"`
	Head            *string          `json:"head"`
	RestingBranch   *string          `json:"resting_branch"`
}

func pointer[T any](v T) *T { return &v }
func slotNumber(v Vantage) int {
	repo := filepath.Base(v.PrimaryRoot)
	prefix := repo + "-slot"
	base := filepath.Base(filepath.Dir(v.WorktreeRoot))
	if filepath.Base(v.WorktreeRoot) != repo || filepath.Dir(filepath.Dir(v.WorktreeRoot)) != filepath.Join(v.FleetRoot, "worktree") || !strings.HasPrefix(base, prefix) {
		return -1
	}
	s := strings.TrimPrefix(base, prefix)
	n, e := strconv.Atoi(s)
	if e != nil || n <= 0 || strconv.Itoa(n) != s {
		return -1
	}
	return n
}

// Classify consumes canonical paths and observed commit refs, and performs no IO.
func Classify(v Vantage, trees []Worktree, refs map[string]string) (Identity, error) {
	id := Identity{SchemaVersion: 2, EnvironmentRoot: v.EnvironmentRoot, EnvironmentHost: v.EnvironmentHost, Repo: filepath.Base(v.PrimaryRoot), RepoIdentity: v.RepoIdentity, PrimaryRoot: v.PrimaryRoot, FleetRoot: v.FleetRoot, WorktreeRoot: v.WorktreeRoot, Kind: "worktree"}
	var selected Worktree
	count := 0
	seen := map[string]bool{}
	for _, w := range trees {
		if !w.Bare && !validOID(w.HEAD) {
			return Identity{}, fmt.Errorf("malformed worktree HEAD OID at %q", w.Path)
		}
		if seen[w.Path] {
			return Identity{}, fmt.Errorf("duplicate worktree membership %q", w.Path)
		}
		seen[w.Path] = true
		if w.Path == v.WorktreeRoot {
			selected = w
			count++
		}
	}
	if count != 1 || selected.Bare {
		return Identity{}, fmt.Errorf("ambiguous or missing worktree membership %q", v.WorktreeRoot)
	}
	if selected.Branch != "" {
		id.Branch = pointer(selected.Branch)
	}
	if strings.Trim(selected.HEAD, "0") != "" {
		id.Head = pointer(selected.HEAD)
	}
	if v.EnvironmentHost != nil && v.RepoIdentity != v.EnvironmentHost.RepoIdentity && v.WorktreeRoot == v.PrimaryRoot {
		id.Kind = "dependency"
		return id, nil
	}
	n := slotNumber(v)
	if v.WorktreeRoot == v.PrimaryRoot {
		n = 0
		id.Kind = "primary"
	} else if n > 0 {
		id.Kind = "slot"
	} else {
		return id, nil
	}
	if !validRepoName(id.Repo) {
		return Identity{}, fmt.Errorf("repository basename %q cannot form a workspace address", id.Repo)
	}
	id.Slot = pointer(n)
	id.Address = pointer((Address{Repo: id.Repo, Slot: n}).String())
	rest := "main"
	if n > 0 {
		rest += "-slot" + strconv.Itoa(n)
	}
	id.RestingBranch = pointer(rest)
	if n == 0 {
		return id, nil
	}
	if !validOID(refs[rest]) || strings.Trim(refs[rest], "0") == "" {
		return Identity{}, fmt.Errorf("slot resting branch %q does not resolve to a commit", rest)
	}
	if selected.Branch == "main" || (strings.HasPrefix(selected.Branch, "main-slot") && reservedSlot(selected.Branch) && selected.Branch != rest) {
		return Identity{}, fmt.Errorf("slot is on mismatched resting branch %q", selected.Branch)
	}
	for _, w := range trees {
		if w.Path != selected.Path && w.Branch == rest {
			return Identity{}, fmt.Errorf("resting branch %q is checked out at %q", rest, w.Path)
		}
	}
	if id.Head == nil {
		return Identity{}, fmt.Errorf("slot has unborn HEAD")
	}
	return id, nil
}
func reservedSlot(branch string) bool {
	s := strings.TrimPrefix(branch, "main-slot")
	n, e := strconv.Atoi(s)
	return e == nil && n > 0 && strconv.Itoa(n) == s
}

// validOID accepts full Git SHA-1 or SHA-256 object IDs, including the
// corresponding zero sentinel. Callers must reject zero for committed refs.
func validOID(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
