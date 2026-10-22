package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

// localCheckEnvironment reads the tested declaration, anchored at the initiating
// checkout. Resolution observes dependencies; it never acquires or changes them.
func localCheckEnvironment(caller, tested string) ([]string, error) {
	caller, err := filepath.Abs(caller)
	if err != nil {
		return nil, err
	}
	rows, err := resolvedDependencyRows(caller, tested, 0)
	if err != nil {
		return nil, err
	}
	return dependencyEnvironment(os.Environ(), caller, rows)
}

// resolvedDependencyRows shares #545's declaration and anchor rules with review
// setup. A positive row limit is checked before resolving any local source.
func resolvedDependencyRows(caller, tested string, limit int) ([]layergraph.Dependency, error) {
	return resolvedDependencyRowsWithAnchors(caller, tested, limit, nil)
}

func resolvedDependencyRowsWithAnchors(caller, tested string, limit int, anchors layergraph.Anchors) ([]layergraph.Dependency, error) {
	caller, err := filepath.Abs(caller)
	if err != nil {
		return nil, err
	}
	content, err := layergraph.ReadDeclaration(filepath.Join(tested, "construct/deps"), layergraph.DeclarationLimit)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	rows, err := layergraph.ParseRows(string(content))
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(rows) > limit {
		return nil, fmt.Errorf("construct/deps exceeds %d direct rows", limit)
	}
	resolver := layergraph.NewResolver(layergraph.OSFS{})
	if anchors != nil {
		resolver.Anchors = anchors
	}
	for i := range rows {
		if rows[i].Kind == "substrate" {
			resolved, err := resolver.Substrate(caller, rows[i].Path)
			if err != nil {
				return nil, err
			}
			rows[i].Path = filepath.Clean(resolved.Path)
		} else {
			rows[i].Path = rows[i].Mount
			if !filepath.IsAbs(rows[i].Path) {
				rows[i].Path = filepath.Join(caller, rows[i].Path)
			}
		}
	}
	return rows, nil
}

// dependencyEnvironment is the pure projection of resolved rows. Reserved
// inherited values are removed even when this declaration exports no dependency.
func dependencyEnvironment(inherited []string, caller string, rows []layergraph.Dependency) ([]string, error) {
	return layergraph.DependencyEnvironment(inherited, caller, rows)
}
