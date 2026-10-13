// project_invalid: a broken model — one lifecycle edge targets a status outside
// every category. vet_test.sh asserts `cue vet` REJECTS this file; if it ever
// passes, the model's own constraints have stopped biting. Self-contained on
// purpose (own package, verbatim copies of categories/#Status/#Transition):
// a standalone file referencing project.cue's definitions would fail vet with
// "reference not found" — a vacuous pass proving nothing about the enum.
package projectinvalid

import "list"

categories: {
	forming:   ["defined"]
	executing: ["executing", "closing"]
	terminal:  ["closed", "dropped"]
}

#Status: or(list.Concat([categories.forming, categories.executing, categories.terminal]))

#Transition: {
	from:   #Status
	to:     #Status
	event:  string
	guards: [...string]
}

lifecycle: [...#Transition] & [
	{from: "defined", to: "shipped", event: "execute", guards: []}, // "shipped" ∉ #Status → vet must fail HERE
]
