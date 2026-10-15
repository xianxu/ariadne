package testfix

import (
	"os"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/agentid"
)

// FixedAgent pins the Agent trailer sdlc stamps on commits (#367), so a test
// sees the same message whichever agent session (or none) runs the suite.
const FixedAgent = "test/test-model/-"

// PinAgent sets FixedAgent for the process; call it from TestMain.
func PinAgent() { _ = os.Setenv(agentid.OverrideEnv, FixedAgent) }
