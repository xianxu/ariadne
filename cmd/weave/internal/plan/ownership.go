package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
	"github.com/xianxu/ariadne/cmd/weave/internal/walk"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// OwnershipScope distinguishes complete artifact compilation from data-only
// reconciliation in an ancestor. Both share a single owner-local inventory.
type OwnershipScope string

const (
	ScopeArtifacts OwnershipScope = "artifacts"
	ScopeData      OwnershipScope = "data"
	InventoryPath                 = staging.RootRel + "/ownership.json"
)

type outputIdentity struct {
	Path  string         `json:"path"`
	Scope OwnershipScope `json:"scope"`
	Kind  string         `json:"kind"`
	Value string         `json:"value"`
	Mode  *os.FileMode   `json:"mode,omitempty"`
}
type inventory struct {
	Version int              `json:"version"`
	Outputs []outputIdentity `json:"outputs"`
}

func validScope(s OwnershipScope) bool { return s == ScopeArtifacts || s == ScopeData }
func safeRelative(path string) bool {
	return path != "" && path != "." && !filepath.IsAbs(path) && filepath.Clean(path) == path && path != ".." && !strings.HasPrefix(path, ".."+string(filepath.Separator)) && !strings.ContainsAny(path, "\x00\r\n")
}

// reservedOutput protects inventory storage and the separately managed ignore file.
func reservedOutput(path string) bool {
	return path == ".gitignore" || path == InventoryPath || strings.HasPrefix(InventoryPath, path+string(filepath.Separator)) || strings.HasPrefix(path, filepath.Dir(InventoryPath)+string(filepath.Separator))
}

// safeParents prevents a lexical path within the owner from writing/deleting
// through an authored parent symlink into another repository.
func safeParents(fs weavefs.FS, root, path string) error {
	if !safeRelative(path) {
		return fmt.Errorf("invalid owned path %q", path)
	}
	return weavefs.CheckParents(fs, root, filepath.Join(root, path))
}

func observe(fs weavefs.FS, root string, id outputIdentity) (bool, error) {
	if e := safeParents(fs, root, id.Path); e != nil {
		return false, e
	}
	p := filepath.Join(root, id.Path)
	fi, e := fs.Lstat(p)
	if os.IsNotExist(e) {
		return false, nil
	}
	if e != nil {
		return false, e
	}
	switch id.Kind {
	case "link":
		if fi.Mode()&os.ModeSymlink == 0 {
			return false, nil
		}
		v, e := fs.Readlink(p)
		return v == id.Value, e
	case "file":
		if id.Mode != nil && fi.Mode().Perm() != *id.Mode {
			return false, nil
		}
		if !fi.Mode().IsRegular() {
			return false, nil
		}
		v, e := fs.ReadFile(p)
		if e != nil {
			return false, e
		}
		return digest(v) == id.Value, nil
	}
	return false, fmt.Errorf("invalid identity kind %q", id.Kind)
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func readInventory(fs weavefs.FS, root string) ([]outputIdentity, error) {
	if e := safeParents(fs, root, InventoryPath); e != nil {
		return nil, e
	}
	p := filepath.Join(root, InventoryPath)
	if fi, e := fs.Lstat(p); e == nil && !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("inventory is not a regular file: %s", p)
	} else if e != nil && !os.IsNotExist(e) {
		return nil, e
	}
	b, e := fs.ReadFile(p)
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	var inv inventory
	if e = json.Unmarshal(b, &inv); e != nil {
		return nil, fmt.Errorf("read ownership inventory: %w", e)
	}
	if inv.Version != 1 {
		return nil, fmt.Errorf("unsupported ownership inventory version %d", inv.Version)
	}
	for _, id := range inv.Outputs {
		if !safeRelative(id.Path) || reservedOutput(id.Path) || !validScope(id.Scope) {
			return nil, fmt.Errorf("invalid ownership inventory entry %q", id.Path)
		}
		if id.Kind != "file" && id.Kind != "link" {
			return nil, fmt.Errorf("invalid ownership identity kind %q", id.Kind)
		}
		if id.Mode != nil && (id.Kind != "file" || *id.Mode != id.Mode.Perm()) {
			return nil, fmt.Errorf("invalid permission identity for %s", id.Path)
		}
		if id.Kind == "file" {
			b, e := hex.DecodeString(id.Value)
			if e != nil || len(b) != sha256.Size {
				return nil, fmt.Errorf("invalid file identity for %s", id.Path)
			}
		}
	}
	return inv.Outputs, nil
}
func saveInventory(fs weavefs.FS, root string, ids []outputIdentity) error {
	ids = uniqueIdentities(ids)
	b, e := json.MarshalIndent(inventory{Version: 1, Outputs: ids}, "", "  ")
	if e != nil {
		return e
	}
	p := filepath.Join(root, InventoryPath)
	if e = ensureParent(fs, p); e != nil {
		return e
	}
	mode := os.FileMode(0600)
	return weavefs.Publish(fs, root, p, append(b, '\n'), &mode)
}
func uniqueIdentities(ids []outputIdentity) []outputIdentity {
	set := map[string]bool{}
	out := make([]outputIdentity, 0, len(ids))
	for _, id := range ids {
		encoded, _ := json.Marshal(id)
		key := string(encoded)
		if !set[key] {
			set[key] = true
			out = append(out, id)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Scope != b.Scope {
			return a.Scope < b.Scope
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Value != b.Value {
			return a.Value < b.Value
		}
		if a.Mode == nil {
			return b.Mode != nil
		}
		return b.Mode != nil && *a.Mode < *b.Mode
	})
	return out
}

// ApplyManaged records matching old and intended identities before applying.
// A failed/interrupted Apply leaves that union durable; a later run can retire
// both untouched old outputs and completed new outputs without the source graph.
// Only this scope is reconciled; other scope entries remain untouched.
func ApplyManaged(fs weavefs.FS, root string, actions []Action, scope OwnershipScope) ([]string, error) {
	if !validScope(scope) {
		return nil, fmt.Errorf("invalid ownership scope %q", scope)
	}
	if err := weavefs.ReclaimPublications(fs, root); err != nil {
		return nil, err
	}
	old, e := readInventory(fs, root)
	if e != nil {
		return nil, e
	}
	strict := map[string]bool{}
	for _, action := range actions {
		if staged, ok := action.(stagedOutput); ok {
			for path := range ProducedPathSet([]Action{staged.Action}) {
				strict[path] = true
			}
		}
	}
	actions, e = materializeManaged(fs, root, actions)
	if e != nil {
		return nil, e
	}
	wanted, e := actionIdentities(root, actions, scope)
	if e != nil {
		return nil, e
	}
	produced := ProducedPathSet(actions)
	// Inspect every destination before any output writes. Known edited replacements
	// are protected; data mounts additionally refuse unowned occupied destinations.
	oldAt := map[string][]outputIdentity{}
	for _, id := range old {
		oldAt[id.Path] = append(oldAt[id.Path], id)
	}
	var previous, other []outputIdentity
	for _, id := range old {
		if id.Scope != scope {
			other = append(other, id)
			continue
		}
		match, e := observe(fs, root, id)
		if e != nil {
			return nil, e
		}
		if match {
			previous = append(previous, id)
		}
	}
	for _, id := range wanted {
		if e := safeParents(fs, root, id.Path); e != nil {
			return nil, e
		}
		for _, prev := range oldAt[id.Path] {
			if prev.Scope != scope {
				return nil, fmt.Errorf("output %s belongs to %s", id.Path, prev.Scope)
			}
		}
		p := filepath.Join(root, id.Path)
		_, statErr := fs.Lstat(p)
		if os.IsNotExist(statErr) {
			continue
		}
		if statErr != nil {
			return nil, statErr
		}
		own := false
		for _, prev := range oldAt[id.Path] {
			match, e := observe(fs, root, prev)
			if e != nil {
				return nil, e
			}
			own = own || match
		}
		same, e := observe(fs, root, id)
		if e != nil {
			return nil, e
		}
		if !own && !same && (scope == ScopeData || strict[id.Path] || len(oldAt[id.Path]) > 0) {
			return nil, fmt.Errorf("preserving authored replacement at %s", p)
		}
	}
	next := append(append([]outputIdentity{}, other...), wanted...)
	gi, e := managedIgnore(fs, root, next)
	if e != nil {
		return nil, e
	}
	prepared := append(append(append([]outputIdentity{}, other...), previous...), wanted...)
	if e = saveInventory(fs, root, prepared); e != nil {
		return nil, e
	}
	if e = Apply(fs, root, actions); e != nil {
		return nil, e
	}
	var retired []string
	for _, id := range previous {
		if produced[id.Path] {
			continue
		}
		match, e := observe(fs, root, id)
		if e != nil {
			return retired, e
		}
		if !match {
			continue
		}
		if e = fs.Remove(filepath.Join(root, id.Path)); e != nil {
			return retired, e
		}
		retired = append(retired, id.Path)
		if e = removeEmptyGeneratedParents(fs, root, id.Path); e != nil {
			return retired, e
		}
	}
	if e = writeManagedIgnore(fs, root, gi); e != nil {
		return retired, e
	}
	// Keep only exact matches for this scope; other scope state is not ours to GC.
	kept := append([]outputIdentity{}, other...)
	for _, id := range wanted {
		match, e := observe(fs, root, id)
		if e != nil {
			return retired, e
		}
		if match {
			kept = append(kept, id)
		}
	}
	if e = saveInventory(fs, root, kept); e != nil {
		return retired, e
	}
	sort.Strings(retired)
	return retired, nil
}
func actionIdentities(root string, actions []Action, scope OwnershipScope) ([]outputIdentity, error) {
	var out []outputIdentity
	for _, a := range actions {
		id := outputIdentity{Scope: scope}
		switch a := a.(type) {
		case Symlink:
			id.Path = a.Dst
			id.Kind = "link"
			rel, e := filepath.Rel(filepath.Dir(filepath.Join(root, a.Dst)), a.Src)
			if e != nil {
				return nil, e
			}
			id.Value = rel
		case WriteFile:
			id.Path = a.Path
			id.Kind = "file"
			id.Value = digest([]byte(a.Content))
			id.Mode = a.Mode
		default:
			continue
		}
		if !safeRelative(id.Path) || reservedOutput(id.Path) {
			return nil, fmt.Errorf("invalid generated output %q", id.Path)
		}
		out = append(out, id)
	}
	return out, nil
}
func materializeManaged(fs weavefs.FS, root string, actions []Action) ([]Action, error) {
	out := make([]Action, 0, len(actions))
	for _, a := range actions {
		switch a := a.(type) {
		case stagedOutput:
			out = append(out, a.Action)
		case Seed:
			if reservedOutput(filepath.Clean(a.Dst)) {
				return nil, fmt.Errorf("seed targets reserved weave state: %s", a.Dst)
			}
			out = append(out, a)
		case EnsureGitignore:
			continue // managed inventory supplies the complete block
		case MergeSettings:
			b, e := mergedSettings(fs, root, a)
			if e != nil {
				return nil, e
			}
			out = append(out, WriteFile{Path: a.Target, Content: string(b)})
		default:
			out = append(out, a)
		}
	}
	return out, nil
}

// Retired materialization directories disappear only when empty. Authored files
// are never swept as a consequence of a generator disappearing.
func removeEmptyGeneratedParents(fs weavefs.FS, root, path string) error {
	for dir := filepath.Dir(path); strings.HasPrefix(dir, walk.GeneratedRel+string(filepath.Separator)); dir = filepath.Dir(dir) {
		entries, err := fs.ReadDir(filepath.Join(root, dir))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if len(entries) != 0 {
			return nil
		}
		if err := fs.Remove(filepath.Join(root, dir)); err != nil {
			return err
		}
	}
	return nil
}
