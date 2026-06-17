package main

import (
	"flag"
	"fmt"
	"os"
)

// main parses --output <dir> (required) and --datatype-dir (default
// construct/datatype), then writes <output>/SKILL.md. Invoked by the datatype
// package's `.dynamic-skill` (`go run ../../../cmd/datatype --output .`) at
// `weave compile` time, with cwd = the package dir — so the default
// construct/datatype resolves from the repo root the marker is run under.
func main() {
	output := flag.String("output", "", "directory to write SKILL.md into (required)")
	datatypeDir := flag.String("datatype-dir", "construct/datatype", "directory of datatype prototypes to enumerate")
	flag.Parse()

	if *output == "" {
		fmt.Fprintln(os.Stderr, "datatype: --output <dir> is required")
		os.Exit(2)
	}
	if err := writeSkill(*datatypeDir, *output); err != nil {
		fmt.Fprintln(os.Stderr, "datatype:", err)
		os.Exit(1)
	}
}
