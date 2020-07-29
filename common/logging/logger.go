package logging

import (
	"fmt"
	"os"
	"sync"
)

// Opt returns a new LogOption.
func Opt() *LogOption { return &LogOption{0, Info, (*OutputOpt)(&sync.Map{})} }

// DefaultOpt returns a default LogOption.
func DefaultOpt() *LogOption { return Opt() }

// LogOption defines the log option struct.
type LogOption struct {
	stackNum  int
	level     Level
	outputOpt *OutputOpt
}

// StackNum sets the stack number of o.
func (o *LogOption) StackNum(num int) *LogOption {
	o.stackNum = num
	if o.stackNum < 0 {
		o.stackNum = 0
	}
	return o
}

// StackNumDelta sets the stack number of o with delta.
func (o *LogOption) StackNumDelta(d int) *LogOption {
	o.stackNum += d
	if o.stackNum < 0 {
		o.stackNum = 0
	}
	return o
}

// Level sets the level of o.
func (o *LogOption) Level(level Level) *LogOption {
	o.level = level
	return o
}

// OutputOpt sets the output option of o
func (o *LogOption) OutputOpt(key string, opt interface{}) *LogOption {
	(*sync.Map)(o.outputOpt).Store(key, opt)
	return o
}

// Clone returns a clone of o.
func (o *LogOption) Clone() *LogOption {
	newOpt := Opt().
		StackNum(o.stackNum).
		Level(o.level)
	(*sync.Map)(o.outputOpt).Range(func(k, v interface{}) bool {
		newOpt.OutputOpt(k.(string), v)
		return true
	})
	return newOpt
}

// Logger defines the logger interface.
type Logger interface {
	CloneLogger() Logger

	AppendOutput(Output)

	RangeLabelMap(fn func(key, value string))
	GetLabelValue(label string) (value string, found bool)
	SetLabel(label string, value string)
	DeleteLabel(label string)
	ClearLabelMap()

	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Notice(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Critical(format string, args ...interface{})

	Log(opt *LogOption, format string, args ...interface{})
}

// logger defines the logger.
type logger struct {
	sync.RWMutex

	labelMap       LabelMap
	thresholdLevel Level
	output         Output

	defaultOpt *LogOption
}

// LoggerOpt defines the logger options.
type LoggerOpt struct {
	LabelMap       LabelMap
	ThresholdLevel Level
	Output         Output
}

// NewLogger returns a new logger.
func NewLogger(opt ...*LoggerOpt) Logger {
	return NewLoggerTag("", opt...)
}

// NewLoggerTag returns a new logger.
func NewLoggerTag(tag string, opt ...*LoggerOpt) Logger {
	// init new logger instance.
	logger := &logger{
		labelMap:       LabelMap{LabelTag: tag},
		thresholdLevel: DefaultThresholdLevel(),
		output:         DefaultOutput(),
		defaultOpt:     DefaultOpt(),
	}

	// config with opt if specified.
	if len(opt) == 1 && opt[0] != nil {
		if len(opt[0].LabelMap) > 0 {
			for k, v := range opt[0].LabelMap {
				logger.labelMap[k] = v
			}
		}
		if opt[0].ThresholdLevel.IsValid() {
			logger.thresholdLevel = opt[0].ThresholdLevel
		}
		if opt[0].Output != nil {
			logger.output = opt[0].Output
		}
	}

	// check if level is valid.
	if !logger.thresholdLevel.IsValid() {
		panic(fmt.Sprintf("invalid log threshold level (%d, %d), [%d]",
			first, last, logger.thresholdLevel))
	}
	return logger
}

// CloneLogger returns a cloned logger.
func (l *logger) CloneLogger() Logger {
	l.RLock()
	defer l.RUnlock()
	m := LabelMap{}
	for key, value := range l.labelMap {
		m[key] = value
	}
	return &logger{
		labelMap:       m,
		thresholdLevel: l.thresholdLevel,
		output:         l.output,
		defaultOpt:     l.defaultOpt.Clone(),
	}
}

// AppendOutput appends a output.
func (l *logger) AppendOutput(o Output) {
	l.output = NewMultiOutput(l.output, o)
}

// RangeLabelMap ranges labelMap data.
func (l *logger) RangeLabelMap(fn func(key, value string)) {
	l.RLock()
	defer l.RUnlock()
	for key, value := range l.labelMap {
		fn(key, value)
	}
}

// GetLabelValue returns labelMap.
func (l *logger) GetLabelValue(label string) (value string, found bool) {
	l.RLock()
	defer l.RUnlock()
	value, found = l.labelMap[label]
	return
}

// SetLabel returns labelMap.
func (l *logger) SetLabel(label string, value string) {
	l.Lock()
	defer l.Unlock()
	l.labelMap[label] = value
}

// DeleteLabel deletes Label.
func (l *logger) DeleteLabel(label string) {
	l.Lock()
	defer l.Unlock()
	delete(l.labelMap, label)
}

// ClearLabelMap returns labelMap.
func (l *logger) ClearLabelMap() {
	l.Lock()
	defer l.Unlock()
	l.labelMap = LabelMap{}
}

// Debug - logger level of dubug
func (l *logger) Debug(format string, args ...interface{}) {
	l.print(l.defaultOpt.outputOpt, 3, Debug, format, args...)
}

// Info - logger level of info
func (l *logger) Info(format string, args ...interface{}) {
	l.print(l.defaultOpt.outputOpt, 3, Info, format, args...)
}

// Notice - logger level of notice
func (l *logger) Notice(format string, args ...interface{}) {
	l.print(l.defaultOpt.outputOpt, 3, Notice, format, args...)
}

// Warn - logger level of warn
func (l *logger) Warn(format string, args ...interface{}) {
	l.print(l.defaultOpt.outputOpt, 3, Warn, format, args...)
}

// Error - logger level of error
func (l *logger) Error(format string, args ...interface{}) {
	l.print(l.defaultOpt.outputOpt, 3, Error, format, args...)
}

// Critical - logger level of error
func (l *logger) Critical(format string, args ...interface{}) {
	l.print(l.defaultOpt.outputOpt, 3, Critical, format, args...)
}

// Log - logger level of dubug
func (l *logger) Log(opt *LogOption, format string, args ...interface{}) {
	if opt == nil {
		opt = l.defaultOpt
	}
	l.print(opt.outputOpt, opt.stackNum+3, opt.level, format, args...)
}

func (l *logger) print(opt *OutputOpt, numStackFrame int, level Level,
	format string, args ...interface{}) {
	defer func() {
		if level <= Critical {
			Finalize()
			os.Exit(1)
		}
	}()
	if level > l.thresholdLevel {
		return
	}
	l.RLock()
	m := LabelMap{}
	for key, value := range l.labelMap {
		m[key] = value
	}
	l.RUnlock()

	// Setup default tags.
	m["pod"] = hostName
	if t := m[LabelTag]; t == "" {
		m[LabelTag] = hostName
	}

	if level <= Error {
		m.addDebugInfo(numStackFrame)
	}

	l.output.Output(opt, level, m, fmt.Sprintf(format, args...)+"\n")
}
