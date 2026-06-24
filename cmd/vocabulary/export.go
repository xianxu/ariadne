package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// runVet cue-vets every noun in the DAG-merged set.
func runVet(args []string) error {
	_ = flag.NewFlagSet("vet", flag.ExitOnError).Parse(args)
	paths, err := resolveVocab()
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return fmt.Errorf("no vocabulary found (construct/vocabulary/*.cue)")
	}
	var cue CueRunner = osCue{}
	for _, name := range sortedKeys(paths) {
		if err := cue.Vet(paths[name]); err != nil {
			return err
		}
	}
	return nil
}

// runExport has two modes:
//   - --output <dir>: cue-export every noun → <dir>/<noun>.json, then write a
//     freshness stamp (.source-sha) over the merged source. The materialization
//     mode the `.dynamic-skill` invokes.
//   - --noun <name>: print just that noun's JSON to stdout. The make/embed path
//     (`vocabulary export --noun issue > cmd/sdlc/internal/issue/issue.json`).
func runExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	output := fs.String("output", "", "dir to write every <noun>.json into (+ a freshness stamp)")
	noun := fs.String("noun", "", "print only this noun's JSON to stdout")
	_ = fs.Parse(args)

	paths, err := resolveVocab()
	if err != nil {
		return err
	}
	var cue CueRunner = osCue{}

	if *noun != "" && *output == "" {
		p, ok := paths[*noun]
		if !ok {
			return fmt.Errorf("no vocabulary noun %q (have: %v)", *noun, sortedKeys(paths))
		}
		js, err := cue.Export(p)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(js)
		return err
	}

	if *output == "" {
		return fmt.Errorf("export needs --output <dir> or --noun <name>")
	}
	if err := os.MkdirAll(*output, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", *output, err)
	}
	for _, name := range sortedKeys(paths) {
		js, err := cue.Export(paths[name])
		if err != nil {
			return err
		}
		dst := filepath.Join(*output, name+".json")
		if err := os.WriteFile(dst, js, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dst, err)
		}
	}
	// Render the served skill body (the always-loaded touch-time breadcrumb).
	if err := os.WriteFile(filepath.Join(*output, "SKILL.md"), []byte(renderSkill(sortedKeys(paths))), 0o644); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}
	// Freshness: stamp the materialization with the merged-source digest, so
	// `vocabulary check --output <dir>` can later detect drift vs the source.
	srcs, err := readSources(paths)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*output, stampFile), []byte(hashSources(srcs)+"\n"), 0o644); err != nil {
		return fmt.Errorf("write stamp: %w", err)
	}
	return nil
}

func sortedKeys(m map[string]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
