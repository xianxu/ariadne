package layergraph

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// #295: a present substrate without construct/base.manifest is a typed
// NotLayerError — from Walk and DeclaredSubstrates alike — carrying the
// substrate and its declaring layer, with the message unchanged.
func TestNotLayerError(t *testing.T) {
	p := canon(t, t.TempDir())
	layer(t, filepath.Join(p, "d"), "substrate ../notalayer\n")
	writeFile(t, filepath.Join(p, "notalayer", "README.md"), "present, no manifest")
	root, dep := filepath.Join(p, "d"), filepath.Join(p, "notalayer")
	want := "substrate " + dep + " (declared in " + filepath.Join(root, "construct", "deps") + ") is present but not a compilable layer: " +
		"missing " + filepath.Join(dep, "construct", "base.manifest") + " — seed its base.manifest (`weave link` does this) or fix the construct/deps path"
	walk := func() error { _, err := Walk(OSFS{}, root); return err }
	declared := func() error { _, err := DeclaredSubstrates(OSFS{}, root); return err }
	for name, run := range map[string]func() error{"Walk": walk, "DeclaredSubstrates": declared} {
		err := run()
		var nle *NotLayerError
		if !errors.As(err, &nle) || nle.Path != dep || nle.Owner != root {
			t.Fatalf("%s: %v (%+v), want NotLayerError{%s, %s}", name, err, nle, dep, root)
		}
		if err.Error() != want {
			t.Fatalf("%s: message changed:\n got %q\nwant %q", name, err.Error(), want)
		}
	}
}

// #295: other failures are not NotLayerError.
func TestOtherFailuresAreNotNotLayerError(t *testing.T) {
	malformed := canon(t, t.TempDir())
	layer(t, filepath.Join(malformed, "d"), "substrate\n")
	unreadable := canon(t, t.TempDir())
	layer(t, filepath.Join(unreadable, "d"), "")
	if err := os.MkdirAll(filepath.Join(unreadable, "d", "construct", "deps"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, root := range map[string]string{"malformed row": malformed, "unreadable construct/deps": unreadable} {
		for view, run := range map[string]func(string) error{
			"Walk":               func(r string) error { _, err := Walk(OSFS{}, r); return err },
			"DeclaredSubstrates": func(r string) error { _, err := DeclaredSubstrates(OSFS{}, r); return err },
		} {
			err := run(filepath.Join(root, "d"))
			var nle *NotLayerError
			if err == nil || errors.As(err, &nle) {
				t.Errorf("%s via %s: %v must fail, and not as NotLayerError", name, view, err)
			}
		}
	}
}
