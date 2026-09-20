// Package weaveownership reads weave's generated-output inventory and proves
// current ownership. Compilation and Git migration share this schema and proof;
// this package never publishes, removes, or changes any filesystem entry.
package weaveownership

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// RootRel is weave's private durable ownership and staging storage.
	RootRel = "construct/generated/weave"
	// Version is the persisted inventory format, including optional mode proof.
	Version = 1
)

// Reader is the narrow observation boundary needed to prove file/link identity.
// Reads use absolute paths; Lstat never follows the final link.
type Reader interface {
	ReadFile(string) ([]byte, error)
	Lstat(string) (os.FileInfo, error)
	Readlink(string) (string, error)
}

type OSReader struct{}

func (OSReader) ReadFile(path string) ([]byte, error)   { return os.ReadFile(path) }
func (OSReader) Lstat(path string) (os.FileInfo, error) { return os.Lstat(path) }
func (OSReader) Readlink(path string) (string, error)   { return os.Readlink(path) }

// Scope distinguishes complete artifact compilation from data-only
// reconciliation in an ancestor. Both share a single owner-local inventory.
type Scope string

const (
	ScopeArtifacts Scope = "artifacts"
	ScopeData      Scope = "data"
	InventoryPath        = RootRel + "/ownership.json"
)

type Identity struct {
	Path  string       `json:"path"`
	Scope Scope        `json:"scope"`
	Kind  string       `json:"kind"`
	Value string       `json:"value"`
	Mode  *os.FileMode `json:"mode,omitempty"`
}
type Inventory struct {
	Version int        `json:"version"`
	Outputs []Identity `json:"outputs"`
}

func ValidScope(s Scope) bool { return s == ScopeArtifacts || s == ScopeData }
func SafeRelative(path string) bool {
	return path != "" && path != "." && !filepath.IsAbs(path) && filepath.Clean(path) == path && path != ".." && !strings.HasPrefix(path, ".."+string(filepath.Separator)) && !strings.ContainsAny(path, "\x00\r\n")
}

// ReservedOutput protects inventory storage and the separately managed ignore file.
func ReservedOutput(path string) bool {
	return path == ".gitignore" || path == InventoryPath || strings.HasPrefix(InventoryPath, path+string(filepath.Separator)) || strings.HasPrefix(path, filepath.Dir(InventoryPath)+string(filepath.Separator))
}

// CheckRelativeParents prevents a lexical path within the owner from writing/deleting
// through an authored parent symlink into another repository.
func CheckRelativeParents(fs Reader, root, path string) error {
	if !SafeRelative(path) {
		return fmt.Errorf("invalid owned path %q", path)
	}
	return CheckParents(fs, root, filepath.Join(root, path))
}

func Matches(fs Reader, root string, id Identity) (bool, error) {
	if e := CheckRelativeParents(fs, root, id.Path); e != nil {
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
		return Digest(v) == id.Value, nil
	}
	return false, fmt.Errorf("invalid identity kind %q", id.Kind)
}
func Digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func Read(fs Reader, root string) ([]Identity, error) {
	if e := CheckRelativeParents(fs, root, InventoryPath); e != nil {
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
	var inv Inventory
	if e = json.Unmarshal(b, &inv); e != nil {
		return nil, fmt.Errorf("read ownership inventory: %w", e)
	}
	if inv.Version != Version {
		return nil, fmt.Errorf("unsupported ownership inventory version %d", inv.Version)
	}
	for _, id := range inv.Outputs {
		if !SafeRelative(id.Path) || ReservedOutput(id.Path) || !ValidScope(id.Scope) {
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

// CheckParents rejects traversal and existing symlink parents before any
// staging/final directory writes. A final symlink may be atomically replaced.
func CheckParents(fs Reader, root, path string) error {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("publication path %s is outside root %s", path, root)
	}
	for dir := filepath.Dir(rel); dir != "."; dir = filepath.Dir(dir) {
		info, err := fs.Lstat(filepath.Join(root, dir))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("publication path %s has non-directory parent %s", path, dir)
		}
	}
	return nil
}

// MatchingPaths returns sorted, unique owner-relative paths whose persisted
// identities still match current files or links, across both scopes. Missing
// inventory means no ownership proof; malformed or unreadable state is an error.
func MatchingPaths(fs Reader, root string) ([]string, error) {
	identities, err := Read(fs, root)
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for _, id := range identities {
		matches, err := Matches(fs, root, id)
		if err != nil {
			return nil, err
		}
		if matches {
			set[id.Path] = true
		}
	}
	paths := make([]string, 0, len(set))
	for path := range set {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}
