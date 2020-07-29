package global

import "github.com/jiarung/mochi/types"

const (
	// SystemPhase is phase of system
	SystemPhase = types.SysPhaseProduction
)

var (
	// IsShuttingDown is a global flag indicating server is shutting down.
	IsShuttingDown = false

	// IsReady is a global flag indicating whether the service is ready or not.
	IsReady = false

	// GitCommitHash is the hash of git commit.
	GitCommitHash string
)
