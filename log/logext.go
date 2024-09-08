package log

// 自定义字段日志(带堆栈)
type LogFields struct {
	fields map[string]interface{}
}

func (l *LogFields) Init(field string, val interface{}) ILoger {
	l.fields = map[string]interface{}{}
	return l.Field(field, val)
}
func (l *LogFields) Field(field string, val interface{}) ILoger {
	l.fields[field] = val
	return l
}
func (l *LogFields) Clone() ILoger {
	tmp := map[string]interface{}{}
	for k, v := range l.fields {
		tmp[k] = v
	}
	return &LogFields{fields: tmp}
}
func (l *LogFields) Trace(format string, v ...interface{}) {
	main.Trace(1, l.fields, format, v...)
}
func (l *LogFields) Tracev(v ...interface{}) {
	main.Tracev(1, l.fields, v...)
}
func (l *LogFields) TraceD(depth int, format string, v ...interface{}) {
	main.Trace(depth+1, l.fields, format, v...)
}
func (l *LogFields) TraceDv(depth int, v ...interface{}) {
	main.Tracev(depth+1, l.fields, v...)
}

func (l *LogFields) Debug(format string, v ...interface{}) {
	main.Debug(1, l.fields, format, v...)
}
func (l *LogFields) Debugv(v ...interface{}) {
	main.Debugv(1, l.fields, v...)
}
func (l *LogFields) DebugD(depth int, format string, v ...interface{}) {
	main.Debug(depth+1, l.fields, format, v...)
}
func (l *LogFields) DebugDv(depth int, v ...interface{}) {
	main.Debugv(depth+1, l.fields, v...)
}

func (l *LogFields) Info(format string, v ...interface{}) {
	main.Info(1, l.fields, format, v...)
}
func (l *LogFields) Infov(v ...interface{}) {
	main.Infov(1, l.fields, v...)
}
func (l *LogFields) InfoD(depth int, format string, v ...interface{}) {
	main.Info(depth+1, l.fields, format, v...)
}
func (l *LogFields) InfoDv(depth int, v ...interface{}) {
	main.Infov(depth+1, l.fields, v...)
}

func (l *LogFields) Warn(format string, v ...interface{}) {
	main.Warn(1, l.fields, format, v...)
}
func (l *LogFields) Warnv(v ...interface{}) {
	main.Warnv(1, l.fields, v...)
}
func (l *LogFields) WarnD(depth int, format string, v ...interface{}) {
	main.Warn(depth+1, l.fields, format, v...)
}
func (l *LogFields) WarnDv(depth int, v ...interface{}) {
	main.Warnv(depth+1, l.fields, v...)
}

func (l *LogFields) Error(format string, v ...interface{}) {
	main.Error(1, l.fields, format, v...)
}
func (l *LogFields) Errorv(v ...interface{}) {
	main.Errorv(1, l.fields, v...)
}
func (l *LogFields) ErrorD(depth int, format string, v ...interface{}) {
	main.Error(depth+1, l.fields, format, v...)
}
func (l *LogFields) ErrorDv(depth int, v ...interface{}) {
	main.Errorv(depth+1, l.fields, v...)
}

func (l *LogFields) Fatal(format string, v ...interface{}) {
	main.Fatal(1, l.fields, format, v...)
}
func (l *LogFields) Fatalv(v ...interface{}) {
	main.Fatalv(1, l.fields, v...)
}
func (l *LogFields) FatalD(depth int, format string, v ...interface{}) {
	main.Fatal(depth+1, l.fields, format, v...)
}
func (l *LogFields) FatalDv(depth int, v ...interface{}) {
	main.Fatalv(depth+1, l.fields, v...)
}
