package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// hostLoadAverage is the 1, 5 and 15 minute load (#630), from /proc/loadavg.
func hostLoadAverage() ([3]float64, error) {
	raw, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return [3]float64{}, err
	}
	fields := strings.Fields(string(raw))
	if len(fields) < 3 {
		return [3]float64{}, fmt.Errorf("/proc/loadavg: %q", raw)
	}
	var out [3]float64
	for i := range out {
		if out[i], err = strconv.ParseFloat(fields[i], 64); err != nil {
			return [3]float64{}, fmt.Errorf("/proc/loadavg: %w", err)
		}
	}
	return out, nil
}
