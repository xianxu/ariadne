// project_invalid: a broken model — one lifecycle edge targets a status outside
// every category. vet_test.sh asserts `cue vet` REJECTS this file; if it ever
// passes, the model's own constraints have stopped biting. Self-contained on
// purpose (own package, verbatim copies of categories/#Status/#Transition):
// a standalone file referencing project.cue's definitions would fail vet with
// "reference not found" — a vacuous pass proving nothing about the enum.
package projectinvalid

import "list"

categories: {
	forming:   ["ideation", "defined"]
	committed: ["committed"]
	executing: ["executing", "paused"]
	terminal:  ["done", "dropped"]
}

#Status: or(list.Concat([categories.forming, categories.committed, categories.executing, categories.terminal]))

#Transition: {
	from:   #Status
	to:     #Status
	event:  string
	guards: [...string]
}

lifecycle: [...#Transition] & [
	{from: "ideation", to: "shipped", event: "define", guards: []}, // "shipped" ∉ #Status → vet must fail HERE
]
