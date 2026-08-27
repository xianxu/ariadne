package vocab

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:generate sh -c "vocabulary export --noun fleet-policy > fleetpolicy.json"

//go:embed fleetpolicy.json
var fleetPolicyJSON []byte

// FleetPolicyModel is the read-only contract metadata exported by the
// `fleet-policy` vocabulary noun (ariadne#200). The declaration structure
// remains owned by construct/vocabulary/fleet-policy.cue; runtime loaders use
// these sets instead of restating accepted versions, kinds, or actions.
type FleetPolicyModel struct {
	DeclarationPath   string   `json:"declarationPath"`
	SupportedVersions []int    `json:"supportedVersions"`
	KeyKinds          []string `json:"keyKinds"`
	CapacityKinds     []string `json:"capacityKinds"`
	Actions           []string `json:"actions"`
}

var fleetPolicyModel = mustLoadFleetPolicy()

func mustLoadFleetPolicy() *FleetPolicyModel {
	var m FleetPolicyModel
	if err := json.Unmarshal(fleetPolicyJSON, &m); err != nil {
		panic(fmt.Sprintf("vocab: corrupt embedded fleetpolicy.json (run `make vocab-embed`): %v", err))
	}
	return &m
}

// FleetPolicy returns an isolated snapshot of the embedded fleet-policy
// contract metadata. Callers may inspect or transform it without mutating the
// process-wide vocabulary authority.
func FleetPolicy() *FleetPolicyModel {
	result := *fleetPolicyModel
	result.SupportedVersions = append([]int(nil), fleetPolicyModel.SupportedVersions...)
	result.KeyKinds = append([]string(nil), fleetPolicyModel.KeyKinds...)
	result.CapacityKinds = append([]string(nil), fleetPolicyModel.CapacityKinds...)
	result.Actions = append([]string(nil), fleetPolicyModel.Actions...)
	return &result
}

// SupportsVersion reports whether version is declared by the vocabulary.
func (m *FleetPolicyModel) SupportsVersion(version int) bool {
	for _, supported := range m.SupportedVersions {
		if supported == version {
			return true
		}
	}
	return false
}

// IsKeyKind reports whether kind is a modeled admission-key discriminator.
func (m *FleetPolicyModel) IsKeyKind(kind string) bool { return contains(m.KeyKinds, kind) }

// IsCapacityKind reports whether kind is a modeled capacity discriminator.
func (m *FleetPolicyModel) IsCapacityKind(kind string) bool {
	return contains(m.CapacityKinds, kind)
}

// IsAction reports whether action is a modeled bounded-capacity action.
func (m *FleetPolicyModel) IsAction(action string) bool { return contains(m.Actions, action) }
