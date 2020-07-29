package logging

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ttacon/chalk"

	"github.com/cobinhood/cobinhood-backend/common/global"
)

// LabelMap are Log Labels.
type LabelMap map[string]string

// Enumeration of L.
const (
	LabelTag        = "tag"
	LabelApp        = "app"
	LabelAuthMethod = "auth_method"

	LabelHTTPMethod     = "http_method"
	LabelHTTPURL        = "http_url"
	LabelHTTPRequestTag = "http_request_tag"
	LabelHTTPRequestIP  = "http_request_ip"

	LabelUserID       = "user_id"
	LabelUserDeviceID = "user_device_id"

	LabelTradingPair = "trading_pair"
	LabelOrderID     = "order_id"
	LabelFundingID   = "funding_id"
	LabelCurrencyID  = "currency_id"
)

const (
	labelGitCommit   = "git_commit"
	labelProcessID   = "pid"
	labelGoroutineID = "go_id"
	labelFuncName    = "func_name"
	labelFileName    = "file_name"
	labelLineNumber  = "line_number"
)

var (
	goTagColorFunc         = chalk.ResetColor.NewStyle()
	goIDColorFunc          = chalk.ResetColor.NewStyle()
	processTagColorFunc    = goTagColorFunc
	processIDColorFunc     = goIDColorFunc
	gitTagColorFunc        = goTagColorFunc
	gitCommitHashColorFunc = goIDColorFunc
	funcNameColorFunc      = chalk.Cyan.NewStyle()
	fileColorFunc          = chalk.Magenta.NewStyle()
	lineColorFunc          = chalk.Yellow.NewStyle()
)

// NewLabelMapTag creates a L with tag.
func NewLabelMapTag(tag string) LabelMap {
	return LabelMap{LabelTag: tag}
}

func (l *LabelMap) addDebugInfo(numStackFrame int) {
	(*l)[labelGitCommit] = global.GitCommitHash
	(*l)[labelProcessID] = fmt.Sprintf("%d", os.Getpid())

	buffer := make([]byte, 64)
	buffer = buffer[:runtime.Stack(buffer, false)]
	bufList := bytes.Fields(buffer)
	goroutineID := "-1"
	if len(bufList) >= 2 {
		goroutineID = fmt.Sprintf("%s", string(bufList[1]))
	}
	(*l)[labelGoroutineID] = goroutineID

	funcName := "???"
	pc, file, line, ok := runtime.Caller(numStackFrame)
	if !ok {
		file = "???"
		line = -1
	} else {
		funcName = runtime.FuncForPC(pc).Name()
		file = filepath.Base(file)
	}
	(*l)[labelFuncName] = funcName + "()"
	(*l)[labelFileName] = file
	(*l)[labelLineNumber] = fmt.Sprintf("%d", line)
}

func (l *LabelMap) debugInfo(styled bool) string {
	if !styled {
		return fmt.Sprintf(
			"%s%s:%s%s:%s%s:%s:%s:%s",
			"Git_",
			(*l)[labelGitCommit],
			"PID_",
			(*l)[labelProcessID],
			"GoID_",
			(*l)[labelGoroutineID],
			(*l)[labelFuncName],
			(*l)[labelFileName],
			(*l)[labelLineNumber],
		)
	}
	return fmt.Sprintf(
		"%s%s:%s%s:%s%s:%s:%s:%s",
		gitTagColorFunc.Style("Git_"),
		gitCommitHashColorFunc.Style((*l)[labelGitCommit]),
		processTagColorFunc.Style("PID_"),
		processIDColorFunc.Style((*l)[labelProcessID]),
		goTagColorFunc.Style("GoID_"),
		goIDColorFunc.Style((*l)[labelGoroutineID]),
		funcNameColorFunc.Style((*l)[labelFuncName]),
		fileColorFunc.Style((*l)[labelFileName]),
		lineColorFunc.Style((*l)[labelLineNumber]),
	)
}
