package exec

import (
	"errors"
	"os/exec"

	"github.com/cobinhood/cobinhood-backend/common/logging"
)

var (
	errInvalidCmd = errors.New("invalid command")
)

// Runner defines the command runner func.
type Runner func(logging.Logger, ...string) ([]byte, error)

// Run is the exec implementation
var Run = run

func run(logger logging.Logger, cmds ...string) ([]byte, error) {
	if len(cmds) < 1 {
		return nil, errInvalidCmd
	}
	cmdName := cmds[0]
	cmd := exec.Command(cmdName, cmds[1:]...)
	return cmd.CombinedOutput()
}
