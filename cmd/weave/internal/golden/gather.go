package golden

import (
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// gather.go is the golden harness's IO seam: it observes the live filesystem to
// fill the classifier's Observed snapshot, and collects the deferred-verb
// intents to ledger. It is STRICTLY read-only on the live repo — Lstat /
// Readlink / Stat / ReadFile only, never a mutation. The pure classifier
// (golden.go) does the reasoning; this seam only looks.

// DeferredIntents collects, across all walked layers, the verbs weave does NOT
// lower to a filesystem Action yet (now just seed — see IsDeferred; tool lowers
// to a ToolDep, merge to a MergeSettings), so the classifier can ledger each as
// EXPECTED rather than silently dropping it.
// De-duplicated by target: a verb declared in multiple layers (or repeated on a
// self-walk) ledgers once. Order-stable (first occurrence wins) so the ledger
// is deterministic.
func DeferredIntents(layers []layer.Layer) []intent.Intent {
	var out []intent.Intent
	seen := map[string]bool{}
	for _, l := range layers {
		for _, in := range l.Intents {
			if !IsDeferred(in.Kind) {
				continue
			}
			if seen[in.Target] {
				continue
			}
			seen[in.Target] = true
			out = append(out, in)
		}
	}
	return out
}

// Gather assembles the classifier Input for one live repo: it walks weave's
// planned Actions + the deferred Intents, observing the live FS state at each
// target's ABSOLUTE path (root-joined, matching how classifyAction/Deferred look
// up Observed). Read-only on the live tree.
func Gather(fs weavefs.FS, root string, actions []plan.Action, deferred []intent.Intent) Input {
	obs := map[string]Observed{}

	observe := func(rel string, readContent bool) {
		abs := filepath.Join(root, rel)
		if _, dup := obs[abs]; dup {
			return
		}
		obs[abs] = observePath(fs, abs, readContent)
	}

	for _, a := range actions {
		switch act := a.(type) {
		case plan.Symlink:
			observe(act.Dst, false)
		case plan.Mkdir:
			observe(act.Path, false)
		case plan.Touch:
			observe(act.Path, false) // existence is enough for create-if-missing
		case plan.WriteFile:
			observe(act.Path, true) // content compared for a WriteFile
		case plan.ToolDep:
			// The probe is bimodal (matching classifyToolDep): a derivative
			// checks construct/deps for the substrate row; the owner self-walk
			// checks go.mod for the tool directive. Both need CONTENT.
			if act.Owner == root {
				observe("go.mod", true)
			} else {
				observe(filepath.Join("construct", "deps"), true)
			}
		case plan.MergeSettings:
			// The probe is THREE files (matching classifyMergeSettings): the base
			// (Source), the optional sibling settings.local.json, and the live
			// target (Target = setup.sh's output). All need CONTENT — the
			// classifier recomputes the merge from base+local and semantically
			// compares it to the target. The local path mirrors Apply/the bash:
			// <dir(Target)>/settings.local.json.
			observe(act.Source, true)
			observe(act.Target, true)
			localRel := filepath.Join(filepath.Dir(act.Target), "settings.local.json")
			observe(localRel, true)
		}
	}
	for _, in := range deferred {
		observe(in.Target, false) // presence is enough for the EXPECTED ledger
	}

	return Input{RepoRoot: root, Actions: actions, Deferred: deferred, Observed: obs}
}

// observePath snapshots one absolute path via the read-only FS ops. Lstat (NOT
// Stat) so a symlink is seen AS a symlink (matching setup.sh's [[ -L ]] and
// Apply's idempotency probe). When the path is a symlink, Readlink captures its
// target; when readContent is set and the path is a regular file, its bytes are
// read for the WriteFile content comparison.
func observePath(fs weavefs.FS, abs string, readContent bool) Observed {
	fi, err := fs.Lstat(abs)
	if err != nil {
		return Observed{Exists: false}
	}
	o := Observed{Exists: true}
	switch {
	case fi.Mode()&os.ModeSymlink != 0:
		o.IsSymlink = true
		if tgt, rerr := fs.Readlink(abs); rerr == nil {
			o.LinkTarget = tgt
		}
	case fi.IsDir():
		o.IsDir = true
	default:
		if readContent {
			if data, rerr := fs.ReadFile(abs); rerr == nil {
				o.Content = string(data)
			}
		}
	}
	return o
}
