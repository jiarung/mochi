package logging

import (
	"cloud.google.com/go/logging"
	"github.com/cobinhood/cobinhood-backend/common/config/misc"
)

// DefaultThresholdLevel returns the default log level.
func DefaultThresholdLevel() Level {
	l := Level(int(misc.ServerLoglevelUint()))
	return l
}

// Level of logger
type Level int

// Log / Severity Levels
const (
	first Level = iota
	Critical
	Error
	Warn
	Notice
	Info
	Debug
	last
)

var (
	levelName = []string{
		"",
		" CRIT",
		"ERROR",
		" WARN",
		" NOTE",
		" INFO",
		"DEBUG",
		"",
	}
	levelSeverity = []logging.Severity{
		-1,
		logging.Critical,
		logging.Error,
		logging.Warning,
		logging.Notice,
		logging.Info,
		logging.Debug,
		-1,
	}
)

// IsValid returns if the l is valid.
func (l Level) IsValid() bool {
	return l < last && l > first
}

// String return the string description of l.
func (l Level) String() string {
	return levelName[l]
}

// Severity return the severity.
func (l Level) Severity() logging.Severity {
	return levelSeverity[l]
}
