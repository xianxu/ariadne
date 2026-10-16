package plan

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/walk"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
	"github.com/xianxu/ariadne/pkg/weaveownership"
)

// Compiler aliases preserve the existing internal action/ownership API while
// the persisted schema and its proof have one owner shared with Git migration.
type OwnershipScope = weaveownership.Scope
type outputIdentity = weaveownership.Identity
type inventory = weaveownership.Inventory

const (
	ScopeArtifacts = weaveownership.ScopeArtifacts
	ScopeData      = weaveownership.ScopeData
	InventoryPath  = weaveownership.InventoryPath
)

func validScope(scope OwnershipScope) bool { return weaveownership.ValidScope(scope) }
func safeRelative(path string) bool        { return weaveownership.SafeRelative(path) }
func reservedOutput(path string) bool      { return weaveownership.ReservedOutput(path) }
func safeParents(fs weavefs.FS, root, path string) error {
	return weaveownership.CheckRelativeParents(fs, root, path)
}
func observe(fs weavefs.FS, root string, id outputIdentity) (bool, error) {
	return weaveownership.Matches(fs, root, id)
}
func digest(data []byte) string { return weaveownership.Digest(data) }

// DisplacedPath is where a compile keeps the latest displaced copy of an output.
func DisplacedPath(path string) string { return weaveownership.DisplacedPath(path) }
func readInventory(fs weavefs.FS, root string) ([]outputIdentity, error) {
	return weaveownership.Read(fs, root)
}

func saveInventory(fs weavefs.FS, root string, ids []outputIdentity) error {
	ids = uniqueIdentities(ids)
	b, e := json.MarshalIndent(inventory{Version: weaveownership.Version, Outputs: ids}, "", "  ")
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

// ManagedReport is what a managed apply did beyond writing its outputs.
type ManagedReport struct {
	Retired     []string // unchanged outputs no longer generated, removed
	Displaced   []string // recorded outputs changed since weave wrote them, moved to DisplacedPath
	Abandoned   []string // outputs no longer generated that changed since weave wrote them, left in place
	Regenerated []string // outputs that existed with other content and were rewritten
}

// EditHome names where an edit to a generated output belongs, so a warning
// about a displaced or abandoned output gives the next action. Pure.
func EditHome(path string) string {
	for _, face := range TargetAll.Faces() {
		if path == face.EntryFile {
			return "AGENTS.local.md"
		}
		if strings.HasPrefix(path, face.SkillDir+"/") {
			return "the skill's source in its owning layer"
		}
	}
	switch {
	case filepath.Base(path) == "settings.json":
		return filepath.Join(filepath.Dir(path), "settings.local.json")
	case strings.HasPrefix(path, walk.GeneratedRel+"/"):
		return "the layer source it is generated from"
	}
	return "the owning layer's manifest source, or a local overlay"
}

// managedPass is a prepared managed apply: every inspection and refusal has
// run, and nothing has been written.
type managedPass struct {
	scope                   OwnershipScope
	actions                 []Action
	wanted, previous, other []outputIdentity
	retained                []outputIdentity
	deferred                map[string]bool
	produced, displace      map[string]bool
	ignore                  string
	report                  ManagedReport
}

// within reports the first of paths that equals path or is its ancestor.
func within(path string, paths map[string]bool) (string, bool) {
	for p := path; p != "." && p != string(filepath.Separator); p = filepath.Dir(p) {
		if paths[p] {
			return p, true
		}
	}
	return "", false
}

// overlap reports a path the pass produces at, above or below path: its write
// there would make path's destination or a parent something weave refuses.
func (pass *managedPass) overlap(path string) (string, bool) {
	if p, ok := within(path, pass.produced); ok {
		return p, true
	}
	for p := range pass.produced {
		if strings.HasPrefix(p, path+string(filepath.Separator)) {
			return p, true
		}
	}
	return "", false
}

// retiring is what the pass removes: outputs it wrote last time, unchanged
// since, that it no longer produces. Commit retires these; a later pass's
// check takes them as gone.
func (pass *managedPass) retiring() []outputIdentity {
	var out []outputIdentity
	for _, id := range pass.previous {
		if !pass.produced[id.Path] {
			out = append(out, id)
		}
	}
	return out
}

// CheckManaged runs the refusals of ApplyManaged for each pass, writing
// nothing (publication stages are reclaimed only by an apply). Passes share
// one root and apply in order, so each is checked against the inventory and
// outputs the passes before it leave (#265).
func CheckManaged(fs weavefs.FS, root string, passes ...ManagedPass) error {
	var after []*managedPass
	for _, p := range passes {
		pass, err := prepareManaged(fs, root, p.Actions, p.Scope, after)
		if err != nil {
			return err
		}
		if pass != nil {
			after = append(after, pass)
		}
	}
	return nil
}

// ManagedPass is one scope's actions for CheckManaged.
type ManagedPass struct {
	Actions []Action
	Scope   OwnershipScope
}

// ApplyManaged records matching old and intended identities before applying.
// A failed/interrupted Apply leaves that union durable; a later run can retire
// both untouched old outputs and completed new outputs without the source graph.
// Only this scope is reconciled; other scope entries remain untouched.
// Recorded outputs always converge: a changed one is displaced, not refused.
func ApplyManaged(fs weavefs.FS, root string, actions []Action, scope OwnershipScope) (ManagedReport, error) {
	if err := weavefs.ReclaimPublications(fs, root); err != nil {
		return ManagedReport{}, err
	}
	pass, err := prepareManaged(fs, root, actions, scope, nil)
	if err != nil || pass == nil {
		return ManagedReport{}, err
	}
	return commitManaged(fs, root, pass)
}

// prepareManaged inspects without writing. A nil pass means there is nothing
// to do: nothing to apply and nothing of this scope to retire, so a mount-less
// data pass leaves an unmigrated repo untouched (#264). Each pass in after is
// taken as applied first: its records replace its scope's, its outputs occupy
// their paths, and the outputs it retires are gone.
func prepareManaged(fs weavefs.FS, root string, actions []Action, scope OwnershipScope, after []*managedPass) (*managedPass, error) {
	var report ManagedReport
	if !validScope(scope) {
		return nil, fmt.Errorf("invalid ownership scope %q", scope)
	}
	old, e := readInventory(fs, root)
	if e != nil {
		return nil, e
	}
	retired := map[string]bool{}
	for _, prior := range after {
		old = append(slices.DeleteFunc(old, func(id outputIdentity) bool { return id.Scope == prior.scope }), prior.wanted...)
		old = append(old, prior.retained...)
		for _, id := range prior.retiring() {
			retired[id.Path] = true
		}
		for path := range prior.produced {
			retired[path] = false
		}
	}
	if len(actions) == 0 && !slices.ContainsFunc(old, func(id outputIdentity) bool { return id.Scope == scope }) {
		return nil, nil
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
	deferred := map[string]bool{}
	for _, action := range actions {
		if a, ok := action.(DeferredOutput); ok {
			if !safeRelative(a.Path) || reservedOutput(a.Path) {
				return nil, fmt.Errorf("invalid deferred output %q", a.Path)
			}
			if e := safeParents(fs, root, a.Path); e != nil {
				return nil, e
			}
			deferred[a.Path] = true
		}
	}
	for _, action := range actions {
		if _, ok := action.(DeferredOutput); ok {
			continue
		}
		for p := range ProducedPathSet([]Action{action}) {
			for d := range deferred {
				if _, scaffold := action.(Mkdir); scaffold && strings.HasPrefix(d, p+string(filepath.Separator)) {
					continue // Creating a parent directory does not write its deferred child.
				}
				if p == d || strings.HasPrefix(p, d+string(filepath.Separator)) || strings.HasPrefix(d, p+string(filepath.Separator)) {
					return nil, fmt.Errorf("output %s overlaps deferred output %s", p, d)
				}
			}
		}
	}
	wanted, e := actionIdentities(root, actions, scope)
	if e != nil {
		return nil, e
	}
	produced := ProducedPathSet(actions)
	// Inspect every destination before any output writes. Changed recorded
	// outputs are displaced; data mounts and staged outputs refuse unowned
	// occupied destinations, as does a directory at a recorded path.
	oldAt := map[string][]outputIdentity{}
	for _, id := range old {
		oldAt[id.Path] = append(oldAt[id.Path], id)
	}
	var previous, other, retained []outputIdentity
	for _, id := range old {
		if deferred[id.Path] {
			if id.Scope != scope {
				return nil, fmt.Errorf("output %s belongs to %s", id.Path, id.Scope)
			}
			retained = append(retained, id)
			continue
		}
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
	displace := map[string]bool{}
	for _, id := range wanted {
		for _, prior := range after {
			if p, ok := prior.overlap(id.Path); ok && prior.scope != scope {
				return nil, fmt.Errorf("%s output %s overlaps %s output %s", scope, id.Path, prior.scope, p)
			}
		}
		// A retired earlier output leaves nothing here; its own parents still count.
		if gone, ok := within(id.Path, retired); ok {
			if e := safeParents(fs, root, gone); e != nil {
				return nil, e
			}
			continue
		}
		if e := safeParents(fs, root, id.Path); e != nil {
			return nil, e
		}
		for _, prev := range oldAt[id.Path] {
			if prev.Scope != scope {
				return nil, fmt.Errorf("output %s belongs to %s", id.Path, prev.Scope)
			}
		}
		p := filepath.Join(root, id.Path)
		info, statErr := fs.Lstat(p)
		if os.IsNotExist(statErr) {
			continue
		}
		if statErr != nil {
			return nil, statErr
		}
		same, e := observe(fs, root, id)
		if e != nil {
			return nil, e
		}
		if same {
			continue
		}
		report.Regenerated = append(report.Regenerated, id.Path)
		own := false
		for _, prev := range oldAt[id.Path] {
			match, e := observe(fs, root, prev)
			if e != nil {
				return nil, e
			}
			own = own || match
		}
		recorded := len(oldAt[id.Path]) > 0
		switch {
		case own:
		case recorded && !info.IsDir():
			displace[id.Path] = true
		case recorded || scope == ScopeData || strict[id.Path]:
			return nil, fmt.Errorf("%s is occupied by something weave did not write: move it aside and re-run weave compile (edits belong in %s)", p, EditHome(id.Path))
		}
	}
	matched := map[string]bool{}
	for _, id := range previous {
		matched[id.Path] = true
	}
	for _, id := range old {
		if id.Scope != scope || produced[id.Path] || matched[id.Path] || slices.Contains(report.Abandoned, id.Path) {
			continue
		}
		_, e := fs.Lstat(filepath.Join(root, id.Path))
		if e == nil {
			report.Abandoned = append(report.Abandoned, id.Path)
		} else if !os.IsNotExist(e) {
			return nil, e
		}
	}
	next := append(append([]outputIdentity{}, other...), wanted...)
	next = append(next, retained...)
	gi, e := managedIgnore(fs, root, old, next, scope == ScopeArtifacts)
	if e != nil {
		return nil, e
	}
	return &managedPass{scope: scope, actions: actions, wanted: wanted, previous: previous, other: other,
		produced: produced, displace: displace, ignore: gi, report: report, retained: retained, deferred: deferred}, nil
}

func commitManaged(fs weavefs.FS, root string, pass *managedPass) (ManagedReport, error) {
	actions, wanted, previous, other := pass.actions, pass.wanted, pass.previous, pass.other
	displace, gi, report := pass.displace, pass.ignore, pass.report
	prepared := append(append(append([]outputIdentity{}, other...), previous...), wanted...)
	prepared = append(prepared, pass.retained...)
	if e := saveInventory(fs, root, prepared); e != nil {
		return report, e
	}
	var e error
	// Once an output is displaced, every failure names where it went, so no
	// caller can drop the move on its error path.
	fail := func(e error) (ManagedReport, error) {
		if len(report.Displaced) == 0 {
			return report, e
		}
		copies := make([]string, len(report.Displaced))
		for i, path := range report.Displaced {
			copies[i] = filepath.Join(root, DisplacedPath(path))
		}
		return report, fmt.Errorf("%w (already displaced: %s)", e, strings.Join(copies, ", "))
	}
	// A rename keeps the displaced bytes, mode and link target exactly.
	for _, path := range slices.Sorted(maps.Keys(displace)) {
		dst := filepath.Join(root, DisplacedPath(path))
		if e = fs.RemoveAll(dst); e != nil {
			return fail(e)
		}
		if e = ensureParent(fs, dst); e != nil {
			return fail(e)
		}
		if e = fs.Rename(filepath.Join(root, path), dst); e != nil {
			return fail(e)
		}
		report.Displaced = append(report.Displaced, path)
	}
	if e = Apply(fs, root, actions); e != nil {
		return fail(e)
	}
	for _, id := range pass.retiring() {
		match, e := observe(fs, root, id)
		if e != nil {
			return fail(e)
		}
		if !match {
			continue
		}
		if e = fs.Remove(filepath.Join(root, id.Path)); e != nil {
			return fail(e)
		}
		report.Retired = append(report.Retired, id.Path)
		if e = removeEmptyGeneratedParents(fs, root, id.Path); e != nil {
			return fail(e)
		}
	}
	if e = writeManagedIgnore(fs, root, gi); e != nil {
		return fail(e)
	}
	// Keep only exact matches for this scope; other scope state is not ours to GC.
	kept := append([]outputIdentity{}, other...)
	kept = append(kept, pass.retained...)
	for _, id := range wanted {
		match, e := observe(fs, root, id)
		if e != nil {
			return fail(e)
		}
		if match {
			kept = append(kept, id)
		}
	}
	if e = sweepDisplaced(fs, root, kept, pass.deferred); e != nil {
		return fail(e)
	}
	if e = saveInventory(fs, root, kept); e != nil {
		return fail(e)
	}
	sort.Strings(report.Retired)
	sort.Strings(report.Abandoned)
	report.Regenerated = slices.Compact(slices.Sorted(slices.Values(report.Regenerated)))
	return report, nil
}

// sweepDisplaced removes the displaced copy of every output that left the
// inventory: a copy lives as long as the output it backs up (ARCH-FUNERAL).
func sweepDisplaced(fs weavefs.FS, root string, live []outputIdentity, deferred map[string]bool) error {
	dir := filepath.Join(root, weaveownership.DisplacedRel)
	entries, err := fs.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	keep := map[string]bool{}
	for path := range deferred {
		keep[filepath.Base(DisplacedPath(path))] = true
	}
	for _, id := range live {
		keep[filepath.Base(DisplacedPath(id.Path))] = true
	}
	for _, entry := range entries {
		if !keep[entry.Name()] {
			if err := fs.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
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
		case SeedOnce:
			if reservedOutput(filepath.Clean(a.Dst)) {
				return nil, fmt.Errorf("seed-once targets reserved weave state: %s", a.Dst)
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
