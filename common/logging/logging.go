package logging

import (
	"os"

	"github.com/cobinhood/cobinhood-backend/common/config/misc"
)

var (
	logToStdout      = misc.ServerLogToStdoutBool()
	logToStackdriver = misc.ServerLogToStackdriverBool()

	hostName, logName string
)

// Initialize initalizes the logging package.
func Initialize(logname string) {
	logName = logname
	hostName, _ = os.Hostname()

	if !logToStackdriver {
		return
	}

	Stackdriver().(*stackdriverOutput).refreshLogger(logName)
}

// Finalize finalizes the logger module.
func Finalize() {
	// flush stdout.
	if stdout.IsLoaded() {
		Stdout().(*stdOutput).close()
		stdout.Clear()
	}

	// flush stackdriver out.
	if stackdriverOut.IsLoaded() {
		err := Stackdriver().(*stackdriverOutput).client.Close()
		if err != nil {
			panic(err)
		}
		stackdriverOut.Clear()
	}
}
