package logging

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cobinhood/cobinhood-backend/cache/cacher"
)

// Shared insntances.
var (
	// *multiOutput
	defaultOut = cacher.NewConst(func() interface{} {
		o := multiOutput{}
		if logToStdout {
			o = append(o, Stdout())
		}
		if logToStackdriver {
			o = append(o, Stackdriver())
		}
		if len(o) == 0 {
			fmt.Println("no default logger specified")
		}
		return &o
	})

	// *stdOutput
	stdout = cacher.NewConst(func() interface{} {
		return newStdOutput()
	})

	// *stackdriverOutput
	stackdriverOut = cacher.NewConst(func() interface{} {
		stackdriver, err := newStackdriverOutput(logName)
		if err != nil {
			panic(err)
		}
		return stackdriver
	})
)

// DefaultOutput returns the default output.
func DefaultOutput() Output {
	return (defaultOut.Get()).(*multiOutput)
}

// Stdout returns the stdout output.
func Stdout() Output {
	return (stdout.Get()).(*stdOutput)
}

// Stackdriver returns the stackdriver output.
func Stackdriver() Output {
	return (stackdriverOut.Get()).(*stackdriverOutput)
}

// OutputOpt defines the output option type.
type OutputOpt sync.Map

// Output defines the log output interface.
type Output interface {
	// Output outputs the logs.
	Output(opt *OutputOpt, level Level, labelMap LabelMap, log string)
}

// NewMultiOutput returns a multi output.
func NewMultiOutput(outputs ...Output) Output {
	o := multiOutput(outputs)
	o.expand()
	return &o
}

// multiOutput defines the multiple output.
type multiOutput []Output

func (o *multiOutput) expand() {
	expanded := multiOutput{}
	for _, sub := range *o {
		if mo, ok := sub.(*multiOutput); ok {
			mo.expand()
			expanded = append(expanded, *mo...)
			continue
		}
		expanded = append(expanded, sub)
	}
	*o = expanded
}

// Output outputs the logs.
func (o *multiOutput) Output(
	opt *OutputOpt, level Level, labelMap LabelMap, log string) {
	l := len(*o)
	if l == 0 {
		return
	} else if l == 1 {
		(*o)[0].Output(opt, level, labelMap, log)
		return
	}
	var wg sync.WaitGroup
	for _, out := range *o {
		wg.Add(1)
		go func(o Output) {
			defer wg.Done()
			o.Output(opt, level, labelMap, log)
		}(out)
	}
	wg.Wait()
}

// removeColor returns a new string with color code removed.
func removeColor(s string) string {
	sb := strings.Builder{}
	sb.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' {
			for ; i < len(s) && s[i] != 'm'; i++ {
			}
		} else {
			sb.WriteByte(s[i])
		}
	}
	return sb.String()
}
