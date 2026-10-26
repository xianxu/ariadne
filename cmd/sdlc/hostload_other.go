//go:build !darwin && !linux

package main

import "errors"

// hostLoadAverage has no reader on this platform; a report says so (#630).
func hostLoadAverage() ([3]float64, error) {
	return [3]float64{}, errors.New("host load average unavailable")
}
