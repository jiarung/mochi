package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ttacon/chalk"

	"github.com/cobinhood/cobinhood-backend/common/utils"
)

var (
	styleMap = map[Level]chalk.Style{
		Debug:    chalk.ResetColor.NewStyle(),
		Info:     chalk.Green.NewStyle(),
		Notice:   chalk.Cyan.NewStyle(),
		Warn:     chalk.Yellow.NewStyle(),
		Error:    chalk.Red.NewStyle(),
		Critical: chalk.Magenta.NewStyle(),
	}

	timeStyle = chalk.ResetColor.NewStyle().WithTextStyle(chalk.Inverse)
	tagStyle  = chalk.ResetColor.NewStyle().WithBackground(chalk.Blue)
)

const timeFormat = "2006-01-02 15:04:05.000"

// StdoutOpt returns a new stdout option.
func StdoutOpt() *StdoutOption { return &StdoutOption{} }

// StdoutOption defines the option of stdout.
type StdoutOption struct {
	withColor bool
}

// WithColor sets if the output if with color.
func (o *StdoutOption) WithColor(c bool) *StdoutOption {
	o.withColor = c
	return o
}

type stdOutput struct {
	io.Writer
	ctx        context.Context
	cancel     context.CancelFunc
	defaultOpt *StdoutOption
	workerChan *utils.UnlimitedChannel
	closeChan  chan struct{}
}

func newStdOutput() *stdOutput {
	opt := StdoutOpt()
	if !utils.IsCI() {
		opt = opt.WithColor(true)
	}
	o := &stdOutput{
		Writer:     os.Stdout,
		defaultOpt: opt,
		workerChan: utils.NewUnlimitedChannel(),
		closeChan:  make(chan struct{}),
	}
	o.ctx, o.cancel = context.WithCancel(context.Background())
	go o.work()
	return o
}

func (o *stdOutput) Output(
	opt *OutputOpt, level Level, labelMap LabelMap, log string) {
	var b []byte
	defer func() {
		select {
		case o.workerChan.In() <- b:
		case <-o.ctx.Done():
			fmt.Println("Stdout worker channel closed")
		}
	}()

	stdOpt := o.defaultOpt
	if stdOptIn, ok := (*sync.Map)(opt).Load("stdout"); ok {
		stdOpt = stdOptIn.(*StdoutOption)
	}
	tsRaw := time.Now().Format(timeFormat)
	svRaw := fmt.Sprintf("%6s", level.String())
	tagRaw := fmt.Sprintf("%16s", labelMap[LabelTag])
	if !stdOpt.withColor {
		if level <= Error {
			log = fmt.Sprintf("%s: %s", labelMap.debugInfo(false), log)
		}
		log = removeColor(log)
		b = []byte(fmt.Sprintf("%s %s %s %s", tsRaw, svRaw, tagRaw, log))
		return
	}

	if level <= Error {
		log = fmt.Sprintf("%s: %s", labelMap.debugInfo(true), log)
	}
	severitiyStyle := styleMap[level]
	timestamp := timeStyle.Style(tsRaw)
	severity := severitiyStyle.Style(svRaw)
	tag := tagStyle.Style(tagRaw)

	b = []byte(fmt.Sprintf("%s %s %s %s", timestamp, severity, tag, log))
}

func (o *stdOutput) work() {
	defer func() { o.closeChan <- struct{}{} }()
	for {
		select {
		case <-o.ctx.Done():
			o.workerChan.Close()
			<-o.workerChan.Done()
			o.flush()
			return
		case b := <-o.workerChan.Out():
			bytes := b.([]byte)
			if len(bytes) <= 0 {
				continue
			}
			o.Writer.Write(bytes)
		}
	}
}

func (o *stdOutput) flush() {
	for _, bytes := range o.workerChan.Dump() {
		o.Writer.Write(bytes.([]byte))
	}
}

func (o *stdOutput) close() {
	o.cancel()
	<-o.closeChan
}
