package exectest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cobinhood/mochi/common/exec"
	"github.com/cobinhood/mochi/common/logging"
)

type mockRunner struct {
	logger    logging.Logger
	execCount int
	expecteds []Execution
}

// Execution is the mock execution.
type Execution struct {
	ExpectedCmds []string
	MockOut      []byte
	MockErr      error
}

func (r *mockRunner) run(logger logging.Logger, cmds ...string) ([]byte, error) {
	if r.execCount >= len(r.expecteds) {
		err := errors.New("out of expected exec")
		panic(err)
	}

	defer func() {
		r.execCount++
	}()
	execution := r.expecteds[r.execCount]
	if len(execution.ExpectedCmds) != len(cmds) {
		panic(fmt.Errorf("unexpected cmds. \"%s\" != \"%s\"",
			strings.Join(execution.ExpectedCmds, " "), strings.Join(cmds, " ")))
	}
	for i, exp := range execution.ExpectedCmds {
		if exp != cmds[i] {
			panic(fmt.Errorf("unexpected cmds. \"%s\" != \"%s\"",
				strings.Join(execution.ExpectedCmds, " "), strings.Join(cmds, " ")))
		}
	}
	r.logger.Info("Run command: %s", strings.Join(cmds, " "))
	if execution.MockErr != nil {
		r.logger.Error("Mock err: %s", execution.MockErr.Error())
		return nil, execution.MockErr
	}
	r.logger.Info("Mock out: %s", string(execution.MockOut))
	return execution.MockOut, nil
}

// MockRun returns a runner func with predefined expected commands and output.
func MockRun(expecteds []Execution) exec.Runner {
	return (&mockRunner{logging.NewLoggerTag("mock-exec"), 0, expecteds}).run
}
