package project

import (
	"regexp"
	"time"
)

var retroDateRE = regexp.MustCompile(`(?m)^### (\d{4}-\d{2}-\d{2}) — retro\b`)

func LatestRetroDate(d *Doc) string {
	matches := retroDateRE.FindAllStringSubmatch(d.SectionBody("Log"), -1)
	latest := ""
	for _, match := range matches {
		if match[1] > latest {
			latest = match[1]
		}
	}
	return latest
}

func RetroStale(d *Doc, today string, maxDays int) bool {
	if !d.HasSection("Log") {
		return false
	}
	latest := LatestRetroDate(d)
	if latest == "" {
		return true
	}
	last, err := time.Parse("2006-01-02", latest)
	if err != nil {
		return true
	}
	now, err := time.Parse("2006-01-02", today)
	if err != nil {
		return false
	}
	return now.Sub(last) > time.Duration(maxDays)*24*time.Hour
}
