package estimate

import (
	"math"
	"sort"
)

// ForecastCalibration turns an issue estimate into expected focused hours for
// the project forecast (#329): hours = estimate × multiplier, where the point
// multiplier is Σ(build+converge) / Σestimate over the eligible rows — exactly
// estimate ÷ build rate (Σest/Σbuild) × (1 + k), k = Σconverge/Σbuild, the #234
// split aggregated. Low/High are the 25th/75th percentiles of the per-issue
// multipliers; with fewer than MinRangeRows rows they collapse to the point.
// Prior is the class prior for an unestimated issue: the median estimate.
type ForecastCalibration struct {
	N                int
	Partial          int // of N, rows weighted by a partial share (#603)
	Point, Low, High float64
	K                float64
	Prior            float64
	Collapsed        bool
}

// MinRangeRows is the fewest eligible rows a percentile range is drawn from.
const MinRangeRows = 3

// MinRangeWeight is the least calibration weight a row needs to enter a
// percentile range or cast a drift vote (#603): a mostly unattributed row
// still counts in the weighted sums, but cannot widen the range on its own.
const MinRangeWeight = 0.5

// CalibrateForecast reads the ledger rows eligible for an est↔build↔converge
// comparison: a measured split with positive build, a real estimate, not
// excluded from calibration, newest row per issue (the ConvergeFactors
// eligibility plus a usable estimate). A hand-edited negative estimate or
// converge is not a data point: it would yield a negative multiplier. Each row
// counts by its CalibrationWeight (#603): the sums are weighted, a weight-0 row
// is no data point, and only rows of at least MinRangeWeight enter the
// percentiles. ok=false when no row qualifies.
func CalibrateForecast(rows []LedgerRow) (ForecastCalibration, bool) {
	var eligible []LedgerRow
	for _, r := range rows {
		if r.HasSplit() && r.BuildHours > 0 && r.Estimate > 0 && r.ConvergeHours >= 0 && !r.ExcludedFromCalibration() {
			eligible = append(eligible, r)
		}
	}
	eligible = NewestPerIssue(eligible)
	var est, build, conv float64
	var per, ests, weighted []float64
	c := ForecastCalibration{}
	for _, r := range eligible {
		w := r.CalibrationWeight()
		if w <= 0 {
			continue
		}
		c.N++
		if w < 1 {
			c.Partial++
		}
		est += w * r.Estimate
		build += w * r.BuildHours
		conv += w * r.ConvergeHours
		weighted = append(weighted, r.Estimate)
		if w >= MinRangeWeight {
			per = append(per, (r.BuildHours+r.ConvergeHours)/r.Estimate)
			ests = append(ests, r.Estimate)
		}
	}
	if c.N == 0 {
		return ForecastCalibration{}, false
	}
	c.Point, c.K = (build+conv)/est, conv/build
	if len(ests) == 0 {
		ests = weighted // the prior is an estimate's typical size, not a measurement
	}
	c.Prior = quantile(ests, 0.5)
	if len(per) < MinRangeRows {
		c.Low, c.High, c.Collapsed = c.Point, c.Point, true
		return c, true
	}
	c.Low, c.High = quantile(per, 0.25), quantile(per, 0.75)
	// The aggregate point can sit outside the per-issue quartiles when a few
	// large issues dominate the sums; keep low ≤ point ≤ high.
	c.Low, c.High = math.Min(c.Low, c.Point), math.Max(c.High, c.Point)
	return c, true
}

// quantile is the linear-interpolated q-quantile of xs (xs non-empty).
func quantile(xs []float64, q float64) float64 {
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	pos := q * float64(len(s)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	return s[lo] + (s[hi]-s[lo])*(pos-float64(lo))
}
