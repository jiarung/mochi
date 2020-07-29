package logging

// Null returns a null logger that does nothing.
func Null() Logger { return (*null)(nil) }

type null struct{}

func (l *null) CloneLogger() Logger                                    { return Null() }
func (l *null) AppendOutput(Output)                                    {}
func (l *null) RangeLabelMap(fn func(key, value string))               {}
func (l *null) GetLabelValue(label string) (value string, found bool)  { return }
func (l *null) SetLabel(label string, value string)                    {}
func (l *null) DeleteLabel(label string)                               {}
func (l *null) ClearLabelMap()                                         {}
func (l *null) Debug(format string, args ...interface{})               {}
func (l *null) Info(format string, args ...interface{})                {}
func (l *null) Notice(format string, args ...interface{})              {}
func (l *null) Warn(format string, args ...interface{})                {}
func (l *null) Error(format string, args ...interface{})               {}
func (l *null) Critical(format string, args ...interface{})            {}
func (l *null) Log(opt *LogOption, format string, args ...interface{}) {}
