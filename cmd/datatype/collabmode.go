package main

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/pkg/layergraph"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// collaborationModeDefault reads a prototype's declared default collaboration
// mode (#415): the backticked value the notes of its Frontmatter shape table's
// `collaboration-mode` row start with, one of vocab.CollaborationModes. That
// row is what makes xx-datatype emit the key into every new instance.
func collaborationModeDefault(content string) (string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	_, shape, ok := strings.Cut(content, "\n## Frontmatter shape\n")
	if !ok {
		return "", fmt.Errorf("no Frontmatter shape section")
	}
	shape, _, _ = strings.Cut(shape, "\n## ")
	for _, line := range strings.Split(shape, "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 5 || strings.TrimSpace(cells[1]) != "`collaboration-mode`" {
			continue
		}
		rest, ok := strings.CutPrefix(strings.TrimSpace(cells[3]), "`")
		mode, _, ok2 := strings.Cut(rest, "`")
		if !ok || !ok2 || !slices.Contains(vocab.CollaborationModes, mode) {
			return "", fmt.Errorf("its collaboration-mode row names no default among %s", strings.Join(vocab.CollaborationModes, ", "))
		}
		return mode, nil
	}
	return "", fmt.Errorf("its Frontmatter shape table has no collaboration-mode row")
}

// lintTypes reports every prototype that declares no valid default
// collaboration mode, as one line each.
func lintTypes(fs layergraph.FS, protos []TypeProto) []string {
	var problems []string
	for _, p := range protos {
		content, err := fs.ReadFile(p.BodyPath)
		if err == nil {
			_, err = collaborationModeDefault(string(content))
		}
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s (%s): %v", p.Name, p.BodyPath, err))
		}
	}
	return problems
}

// runLint is `datatype lint`: every DAG-resolved type must declare its
// default collaboration mode. It fails listing the ones that don't.
func runLint(stdout io.Writer) error {
	protos, err := resolveTypes()
	if err != nil {
		return err
	}
	problems := lintTypes(layergraph.OSFS{}, protos)
	for _, p := range problems {
		fmt.Fprintln(stdout, p)
	}
	if len(problems) > 0 {
		return fmt.Errorf("%d of %d types declare no default collaboration mode", len(problems), len(protos))
	}
	fmt.Fprintf(stdout, "%d types, each declares a default collaboration mode\n", len(protos))
	return nil
}
