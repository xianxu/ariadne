package plan

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
)

// A shipped output is a generated file the repository commits on purpose
// (#422): parley's issue vocabulary, which its plugin reads from the installed
// tree with no weave run. Its committed sidecar, SidecarPath(p), holds the
// digest of the bytes weave wrote, so a cold checkout (no inventory) can still
// tell weave's output from a hand edit.

// SidecarPath names a shipped output's digest sidecar.
func SidecarPath(p string) string { return p + ".weave-sha256" }

// ShippedPaths lists the selected `ship` rows' paths: every layer's exports and
// the leaf's internals, foundation-first. Pure.
func ShippedPaths(layers []layer.Layer) ([]string, error) {
	var out []string
	leafIdx := len(layers) - 1
	for i, l := range layers {
		for _, in := range l.Intents {
			if in.Kind != intent.Ship || !participates(in.Visibility, i, leafIdx) {
				continue
			}
			p := filepath.Clean(in.Source)
			if !safeRelative(p) {
				return nil, fmt.Errorf("ship %s: not a safe repository-relative path", in.Source)
			}
			out = append(out, p)
		}
	}
	return out, nil
}

// ShipStaged marks each staged file a `ship` row declares and adds its sidecar,
// holding the digest of the content weave is about to write. It returns the
// declared paths no staged file produces, for a warning. Pure.
func ShipStaged(actions []Action, shipped []string) ([]Action, []string) {
	want := map[string]bool{}
	for _, p := range shipped {
		want[p] = true
	}
	var sidecars []Action
	out := make([]Action, 0, len(actions))
	for _, a := range actions {
		if s, ok := a.(stagedOutput); ok {
			if w, ok := s.Action.(WriteFile); ok && want[filepath.Clean(w.Path)] {
				delete(want, filepath.Clean(w.Path))
				s.Sidecar = SidecarPath(w.Path)
				mode := sidecarMode
				sidecars = append(sidecars, WriteFile{Path: s.Sidecar, Content: digest([]byte(w.Content)) + "\n", Mode: &mode})
				a = s
			}
		}
		out = append(out, a)
	}
	var unmatched []string
	for _, p := range shipped {
		if want[p] {
			unmatched = append(unmatched, p)
		}
	}
	return append(out, sidecars...), unmatched
}

const sidecarMode os.FileMode = 0o644
