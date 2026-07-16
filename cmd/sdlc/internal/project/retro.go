package project

import "regexp"

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
