// fleet-policy — the portable repository declaration for fleet concurrency
// admission (ariadne#200). This is an atomic, closed noun: it has no lifecycle.
// Tools consume the concrete metadata export; declarations validate against
// #FleetPolicy.
package fleet_policy

import (
	"list"
	"strings"
)

declarationPath:   ".sdlc/fleet.json"
supportedVersions: [1]
keyKinds:          ["repo", "worktree", "declared-root"]
capacityKinds:     ["bounded", "unbounded"]
actions:           ["reject", "provision-worktree"]

#Version:      or(supportedVersions)
#KeyKind:      or(keyKinds)
#CapacityKind: or(capacityKinds)
#Action:       or(actions)

// A declared-root rule is exactly <safe literal repo-relative prefix>/*. The
// terminal wildcard matches one non-empty path segment; it is not a general
// glob. Empty, wildcard-bearing, dot, and parent-traversal prefix segments fail
// closed.
#DeclaredRootRule: string & =~"^[A-Za-z0-9][A-Za-z0-9._ -]*(?:/[A-Za-z0-9][A-Za-z0-9._ -]*)*/\\*$"

#RepoKey: {
	kind!:  keyKinds[0]
	roots!: []
}

#WorktreeKey: {
	kind!:  keyKinds[1]
	roots!: []
}

#DeclaredRootKey: {
	kind!:  keyKinds[2]
	roots!: [#DeclaredRootRule, ...#DeclaredRootRule]
	// Prefix nesting would make one requested path match multiple rules and
	// introduce undeclared ordering semantics. Compare every pair after removing
	// the terminal wildcard; the resulting list must contain only true values.
	_noOverlap: [
		for i, left in roots
		for j, right in roots
		if i < j {
			!(strings.HasPrefix(left, strings.TrimSuffix(right, "/*")+"/") ||
				strings.HasPrefix(right, strings.TrimSuffix(left, "/*")+"/"))
		},
	] & [...true]
}

#Key: #RepoKey | #WorktreeKey | #DeclaredRootKey

#BoundedCapacity: {
	kind!:  capacityKinds[0]
	limit!: int & >0
}

#UnboundedCapacity: {
	kind!: capacityKinds[1]
}

#Admission: ({
	key!:        #Key
	capacity!:   #BoundedCapacity
	onCapacity!: #Action
} | {
	key!:      #Key
	capacity!: #UnboundedCapacity
})

#FleetPolicy: {
	version!:   #Version
	admission!: #Admission
}

// Compile-time model laws: each public token set is non-empty. These keep the
// concrete export useful to its Go binding instead of allowing vacuous checks.
laws: {
	"versions-nonempty":   list.MinItems(1) & supportedVersions
	"keys-nonempty":       list.MinItems(1) & keyKinds
	"capacities-nonempty": list.MinItems(1) & capacityKinds
	"actions-nonempty":    list.MinItems(1) & actions
}
