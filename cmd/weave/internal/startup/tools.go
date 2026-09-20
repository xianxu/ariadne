package startup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Tools builds each owner's optional tools target in foundation-first order.
// An empty supplemental target handles omission without hiding Makefile errors.
func Tools(fs weavefs.FS, layers []string, runner weavefs.InputRunner, dryRun bool, out io.Writer) error {
	for _, dir := range layers {
		info, err := fs.Stat(filepath.Join(dir, "Makefile"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read tools Makefile in %s: %w", dir, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("tools Makefile in %s is not a regular file", dir)
		}
		fmt.Fprintf(out, "weave: %s: make tools\n", dir)
		if dryRun {
			continue
		}
		if err := runner.RunInput(dir, []string{"make", "--no-print-directory", "-f", "Makefile", "-f", "-", "tools"}, "tools:\n"); err != nil {
			return fmt.Errorf("build tools for %s: %w", dir, err)
		}
	}
	return nil
}

// ToolEnvironment exposes owner binaries to child builds and generators only.
func ToolEnvironment(env []string, layers []string) []string {
	result := make([]string, 0, len(env)+1)
	var previous string
	for _, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			previous = strings.TrimPrefix(entry, "PATH=")
			continue
		}
		result = append(result, entry)
	}
	paths := make([]string, 0, len(layers)+1)
	// The consuming layer has the final override, like artifact composition.
	for i := len(layers) - 1; i >= 0; i-- {
		paths = append(paths, filepath.Join(layers[i], "bin"))
	}
	if previous != "" {
		paths = append(paths, previous)
	}
	return append(result, "PATH="+strings.Join(paths, string(os.PathListSeparator)))
}
