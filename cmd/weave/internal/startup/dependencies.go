// Package startup prepares layer-owned prerequisites before artifact composition.
package startup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Dependencies delegates package state to Homebrew, in the graph's existing
// foundation-first order. Check-only callers never provision packages; all calls
// suppress automatic metadata updates. A layer without a Brewfile needs no call.
func Dependencies(fs weavefs.FS, layers []string, runner weavefs.Runner, checkOnly, dryRun bool, out io.Writer) error {
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
		mode := []string{"install", "--no-upgrade", "--file=Brewfile"}
		if checkOnly {
			mode = []string{"check", "--no-upgrade", "--verbose", "--file=Brewfile"}
		}
		argv := append([]string{"env", "HOMEBREW_NO_AUTO_UPDATE=1", "brew", "bundle"}, mode...)
		fmt.Fprintf(out, "weave: %s: %s\n", dir, strings.Join(argv, " "))
		if dryRun {
			continue
		}
		if err := runner.Run(dir, argv); err != nil {
			if checkOnly {
				bundle, pathErr := filepath.Abs(bundle)
				if pathErr != nil {
					return fmt.Errorf("dependency check for %s: %w (resolve operator Brewfile: %v)", dir, err, pathErr)
				}
				quoted := "'" + strings.ReplaceAll(bundle, "'", "'\\''") + "'"
				return fmt.Errorf("dependency check for %s failed: %w\nask the operator to run: HOMEBREW_NO_AUTO_UPDATE=1 brew bundle install --no-upgrade --file=%s", dir, err, quoted)
			}
			return fmt.Errorf("dependencies for %s (Homebrew must be installed and on PATH): %w", dir, err)
		}
	}
	return nil
}
