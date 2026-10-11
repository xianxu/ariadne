package estimate

import "sort"

// ConvergeFactor is the review-and-fix loop's cost relative to build for one
// (model, flow, recipe) group (#234): k = Converge / Build, over N issues. An
// estimate prices build; k is how much convergence that build tends to need.
type ConvergeFactor struct {
	Model, Flow, Recipe string
	N                   int
	Build, Converge     float64 // hours
}

// K is Converge / Build, 0 when there is no build.
func (f ConvergeFactor) K() float64 {
	if f.Build == 0 {
		return 0
	}
	return f.Converge / f.Build
}

// ConvergeFactors groups the newest row of each issue that carries a split
// with positive build (#234), by model, flow and review recipe, sorted by that
// key. Filter, then dedupe: the order NewestPerIssue's contract requires.
func ConvergeFactors(rows []LedgerRow) []ConvergeFactor {
	var eligible []LedgerRow
	for _, r := range rows {
		if r.HasSplit() && r.BuildHours > 0 {
			eligible = append(eligible, r)
		}
	}
	idx := map[[3]string]int{}
	var out []ConvergeFactor
	for _, r := range NewestPerIssue(eligible) {
		key := [3]string{r.Model, r.FlowKind, r.ReviewRecipe}
		i, ok := idx[key]
		if !ok {
			i = len(out)
			idx[key] = i
			out = append(out, ConvergeFactor{Model: r.Model, Flow: r.FlowKind, Recipe: r.ReviewRecipe})
		}
		out[i].N++
		out[i].Build += r.BuildHours
		out[i].Converge += r.ConvergeHours
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Flow != b.Flow {
			return a.Flow < b.Flow
		}
		if a.Model != b.Model {
			return a.Model < b.Model
		}
		return a.Recipe < b.Recipe
	})
	return out
}
