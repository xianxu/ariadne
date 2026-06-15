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
// lower to a filesystem Action yet (as of #95 M5 NONE — merge lowers to a
// MergeSettings, seed to a Seed; the `tool` verb is RETIRED, not deferred), so
// the classifier can ledger each as EXPECTED rather than silently dropping it.
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

	// observeMerge records a merge-probe file's RESOLVED content (Stat + ReadFile
	// FOLLOW symlinks, unlike observePath's Lstat), so a symlinked base/target
	// (the derivative case) yields real bytes. It merges Content into any existing
	// Observed (a path may also be a Symlink-action probe, observed as a symlink
	// with no content) rather than clobbering its symlink fields. Absent ⇒ leave
	// the existing record (or an Exists:false) so the classifier sees it missing.
	observeMerge := func(rel string) {
		abs := filepath.Join(root, rel)
		cur, had := obs[abs]
		if _, err := fs.Stat(abs); err != nil {
			if !had {
				obs[abs] = Observed{Exists: false}
			}
			return // absent (Stat follows the link; a dangling link is also absent)
		}
		cur.Exists = true
		if data, rerr := fs.ReadFile(abs); rerr == nil {
			cur.Content = string(data)
		}
		obs[abs] = cur
	}

	// observeAbs observes a path given ALREADY-ABSOLUTE (not root-joined) — for a
	// Seed's upstream source, which lives at the layer's abs path, potentially
	// OUTSIDE the consuming repo root. Content is always read (the seed compares
	// the target to the source bytes).
	observeAbs := func(abs string) {
		if _, dup := obs[abs]; dup {
			return
		}
		obs[abs] = observePath(fs, abs, true)
	}

	for _, a := range actions {
		switch act := a.(type) {
		case plan.Symlink:
			observe(act.Dst, false)
		case plan.Mkdir:
			observe(act.Path, false)
		case plan.Seed:
			// Two probes (matching classifyAction's Seed case): the target (Dst,
			// root-relative) and the upstream source (Src, absolute). Both need
			// CONTENT — the classifier compares the live target to the source
			// bytes. The source is FOLLOWED to its real content (a layer's
			// upstream file is a regular file, but observeAbs reads whatever the
			// path resolves to).
			observe(act.Dst, true)
			observeAbs(act.Src)
		case plan.Touch:
			observe(act.Path, false) // existence is enough for create-if-missing
		case plan.WriteFile:
			observe(act.Path, true) // content compared for a WriteFile
		case plan.MergeSettings:
			// The probe is THREE files (matching classifyMergeSettings): the base
			// (Source), the optional sibling settings.local.json, and the live
			// target (Target = setup.sh's output). All need CONTENT — the
			// classifier recomputes the merge from base+local and semantically
			// compares it to the target. The local path mirrors Apply/the bash:
			// <dir(Target)>/settings.local.json.
			//
			// Crucially the merge probe reads content by FOLLOWING symlinks: in a
			// derivative repo .claude/settings.ariadne.json is itself a symlink to
			// the ariadne base, and merge-settings.sh (json.load(open(...))) follows
			// it. The default observePath uses Lstat + a regular-file content read,
			// so a symlinked base would carry an empty Content and the merge would
			// spuriously fail to parse — a harness bug, not a port gap. observeMerge
			// records the resolved content alongside any existing symlink fields.
			observeMerge(act.Source)
			observeMerge(act.Target)
			localRel := filepath.Join(filepath.Dir(act.Target), "settings.local.json")
			observeMerge(localRel)
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
