package errors

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"

	"github.com/cobinhood/mochi/common/logging"
)

var logger logging.Logger

// Initialize initializes error reporter.
func Initialize(l logging.Logger) {
	logger = l
}

// Catch is used for logging panic call stack. Catch should be called with
// defer.
func Catch() {
	if recovered := recover(); recovered != nil {
		logger.Critical("%v\n%s", recovered, string(debug.Stack()))
	}
}

// CatchWithLogger is a panic handler expected to be deferred.
func CatchWithLogger(logger logging.Logger) {
	if recovered := recover(); recovered != nil {
		format := "\x1b[31m%v\n[Stack Trace]\n%s\x1b[m"
		stack := debug.Stack()
		if logger != nil {
			logger.Error(format, recovered, stack)
		} else {
			fmt.Fprintf(os.Stderr, format, recovered, stack)
		}
	}
}

// RecoverErrorPanic recovers an panic if needed. If panic recovered, it set the
// panic content to pErr and set pRet (return value) to default value.
// Useful when using assertions in types.
func RecoverErrorPanic(pErr *error, pRet interface{}) {
	if r := recover(); r != nil {
		// Catch parsing errors.
		panicErr, ok := r.(error)
		if !ok {
			// Reraise.
			panic(r)
		}
		// Set the error in panic to the content of pErr.
		*pErr = panicErr
		// Set the value inside pRet to zero value.
		ret := reflect.ValueOf(pRet).Elem()
		ret.Set(reflect.New(ret.Type()).Elem())
	}
}

// AssertInt64 parse the string to int64. Panic if any error is encountered.
func AssertInt64(s string) int64 {
	if s == "" {
		return 0
	}
	i, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		panic(err)
	}
	return i
}

// AssertBool parse the string to bool. Panic if any error is encountered.
func AssertBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		panic(err)
	}
	return b
}
