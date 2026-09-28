//go:build unix

package judge

import (
	"os/exec"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
)

func configureReviewProcess(cmd *exec.Cmd) {
	processgroup.Configure(cmd)
}

func terminateReviewProcess(cmd *exec.Cmd, force bool) error {
	return processgroup.Terminate(cmd, force)
}
