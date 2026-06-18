package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// formatList renders the apply-time listing: one line per prototype,
// "<name>\t<description>". This is the matching surface — an agent reads each
// description (the prototype's frontmatter `description:`, "Use when …") to pick
// a type without loading any body. Pure (no IO); the input is already sorted by
// name (mergeTypes), so the output order follows it.
func formatList(protos []TypeProto) string {
	var b strings.Builder
	for _, p := range protos {
		b.WriteString(p.Name)
		b.WriteByte('\t')
		b.WriteString(p.Description)
		b.WriteByte('\n')
	}
	return b.String()
}

// runList resolves the DAG-merged types for cwd's repo and prints formatList to
// out. The thin IO shell over resolveTypes + formatList.
func runList(out io.Writer) error {
	protos, err := resolveTypes()
	if err != nil {
		return err
	}
	fmt.Fprint(out, formatList(protos))
	return nil
}

// runShow resolves the DAG-merged types and prints the leaf-winning prototype
// body (the BodyPath file) for name to out. An unknown name lists the available
// names to errOut and returns a non-zero-exit error, so the agent sees what IS
// available. The resolved body is the most-leafward layer's prototype (the merge
// already picked the winner).
func runShow(name string, out, errOut io.Writer) error {
	protos, err := resolveTypes()
	if err != nil {
		return err
	}
	for _, p := range protos {
		if p.Name == name {
			body, err := os.ReadFile(p.BodyPath)
			if err != nil {
				return fmt.Errorf("read prototype body %s: %w", p.BodyPath, err)
			}
			fmt.Fprint(out, string(body))
			return nil
		}
	}
	// Unknown: list the available names to stderr, then signal non-zero exit.
	names := make([]string, len(protos))
	for i, p := range protos {
		names[i] = p.Name
	}
	sort.Strings(names)
	fmt.Fprintf(errOut, "datatype: unknown type %q. Available: %s\n", name, strings.Join(names, ", "))
	return errUnknownType
}

// errUnknownType is the sentinel runShow returns for an unknown name, so main
// exits non-zero AFTER runShow has already printed the available-names guidance
// (main must not re-print it). Distinguished from a real IO error.
var errUnknownType = fmt.Errorf("unknown datatype")
