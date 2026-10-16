package tlrules

import (
	_ "embed"
	"strings"
)

//go:embed planning.md
var planning string

// ComposeHelp inserts the shared planning contract into a command's help.
func ComposeHelp(help string) string {
	return strings.ReplaceAll(help, "{{PLANNING}}", strings.TrimSpace(planning))
}
