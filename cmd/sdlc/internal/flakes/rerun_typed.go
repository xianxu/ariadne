package flakes

import (
	"fmt"
	"slices"
	"strings"
)

// RerunPlan selects execution commands for a registered identity set. Go
// commands retain their whole-test selection; typed commands run each exact
// file once. Lua case labels remain in tests for proof and never enter argv.
func RerunPlan(tests []string, unitCommand []string) ([][]string, error) {
	if len(tests) == 0 {
		return nil, fmt.Errorf("no tests to rerun")
	}
	var goTests []string
	units := map[string]bool{}
	for _, test := range tests {
		id, err := ParseIdentity(test)
		if err != nil {
			return nil, err
		}
		if id.Kind() == KindGo {
			goTests = append(goTests, test)
		} else {
			units[id.Unit()] = true
		}
	}
	if len(units) > 0 {
		if len(unitCommand) == 0 {
			return nil, fmt.Errorf("typed rerun requires an execution command")
		}
		for _, arg := range unitCommand {
			if arg == "" || strings.ContainsRune(arg, '\x00') {
				return nil, fmt.Errorf("typed rerun command contains an empty or NUL argument")
			}
		}
	}
	var commands [][]string
	if len(goTests) > 0 {
		var err error
		commands, err = RerunCommands(goTests)
		if err != nil {
			return nil, err
		}
	}
	ordered := make([]string, 0, len(units))
	for unit := range units {
		ordered = append(ordered, unit)
	}
	slices.Sort(ordered)
	for _, unit := range ordered {
		command := append([]string(nil), unitCommand...)
		commands = append(commands, append(command, "--unit", unit))
	}
	return commands, nil
}

// RerunMissingProof lists selected identities lacking authoritative pass proof.
// Typed execution groups by file, but proof requires each original exact case.
func RerunMissingProof(log string, tests []string) []string {
	var goTests, typedTests, missing []string
	for _, test := range tests {
		id, err := ParseIdentity(test)
		if err != nil {
			missing = append(missing, test)
		} else if id.Kind() == KindGo {
			goTests = append(goTests, test)
		} else {
			typedTests = append(typedTests, test)
		}
	}
	if len(goTests) > 0 {
		missing = append(missing, RerunMissingJSON(log, goTests)...)
	}
	if len(typedTests) > 0 {
		missing = append(missing, RequiredUnitProof(ParseUnitEvidence(log), typedTests)...)
	}
	slices.Sort(missing)
	return slices.Compact(missing)
}
