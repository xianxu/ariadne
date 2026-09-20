// Package startup prepares layer-owned prerequisites before artifact composition.
package startup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Dependencies delegates package state to Homebrew, in the graph's existing
// foundation-first order. A layer with no Brewfile needs no package operation.
func Dependencies(fs weavefs.FS, layers []string, runner weavefs.Runner, dryRun bool, out io.Writer) error {
	for _, dir := range layers {
		bundle := filepath.Join(dir, "Brewfile")
		info, err := fs.Stat(bundle)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read layer bundle %s: %w", bundle, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("layer bundle %s is not a regular file", bundle)
		}
		fmt.Fprintf(out, "weave: %s: brew bundle install --no-upgrade --file=Brewfile\n", dir)
		if dryRun {
			continue
		}
		if err := runner.Run(dir, []string{"brew", "bundle", "install", "--no-upgrade", "--file=Brewfile"}); err != nil {
			return fmt.Errorf("dependencies for %s (Homebrew must be installed and on PATH): %w", dir, err)
		}
	}
	return nil
}
