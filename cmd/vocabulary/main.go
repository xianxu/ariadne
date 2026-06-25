// Command vocabulary is the DAG-aware compiler for the formal vocabulary layer
// (ariadne#122): the system's nouns + lifecycles declared as CUE in
// construct/vocabulary/*.cue, the single source consumers derive from.
//
//	vocabulary vet                      — cue vet the DAG-merged set
//	vocabulary export --output <dir>    — cue export each noun → <dir>/<noun>.json (+ a freshness stamp)
//	vocabulary export --noun <name>     — print one noun's JSON to stdout (the make/embed path)
//	vocabulary check  --output <dir>    — fail if <dir> is stale vs the current merged source
//	vocabulary validate-instance --type <noun> <file>
//	                                    — cue-vet a typed-markdown FILE's frontmatter against
//	                                      #<Noun> (instance-conformance, #124)
//
// The export --output form is invoked by the vocabulary skill's `.dynamic-skill`
// at `weave compile` (cwd = the compiling repo, --output construct/generated/
// vocabulary), mirroring cmd/datatype. The merge primitive (leaf-wins by
// filename) is shared via pkg/layergraph.MergeByName (ARCH-DRY).
package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "vocabulary: want a subcommand (vet, export, check, validate-instance)")
		os.Exit(2)
	}
	var err error
	switch args[0] {
	case "vet":
		err = runVet(args[1:])
	case "export":
		err = runExport(args[1:])
	case "check":
		err = runCheck(args[1:])
	case "validate-instance":
		err = runValidateInstance(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "vocabulary: unknown subcommand %q (want vet, export, check, validate-instance)\n", args[0])
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "vocabulary:", err)
		os.Exit(1)
	}
}
