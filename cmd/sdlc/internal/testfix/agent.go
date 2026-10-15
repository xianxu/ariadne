package testfix

import "os"

// FixedAgent pins the Agent trailer sdlc stamps on commits (#367), so a test
// sees the same message whichever agent session (or none) runs the suite.
const FixedAgent = "test/test-model/-"

// AgentOverrideEnv is agentid.OverrideEnv, named here because testfix is
// imported by agentid's own dependencies (judge); agentid's tests pin the two.
const AgentOverrideEnv = "SDLC_AGENT_ID"

// PinAgent sets FixedAgent for the process; call it from TestMain.
func PinAgent() { _ = os.Setenv(AgentOverrideEnv, FixedAgent) }
