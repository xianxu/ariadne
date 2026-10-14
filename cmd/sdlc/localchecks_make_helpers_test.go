package main

import "context"

// These aliases keep fixture setup terse; production probes require the caller context.
func makeTestFacts(git func(args ...string) (string, error), rev string) (has bool, weight int, err error) {
	return makeTestFactsContext(context.Background(), git, rev)
}

func probeMakeTestFacts(dir string) (has bool, weight int, err error) {
	return probeMakeTestFactsContext(context.Background(), dir)
}
