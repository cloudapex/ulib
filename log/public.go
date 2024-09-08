package log

/*******************************************************************************
Copyright:cloud
Author:cloudapex@126.com
Version:1.0
Date:2021-11-22
Description: 日志模块
*******************************************************************************/

// Main Log
var main *logger

// Init
func Init(conf *Config) {
	main = New(conf).Start()
}

// Term
func Term() {
	if main != nil {
		main.Stop()
	}
}

// Filter 设置日志过滤器
func Filter(filter func(msg *LogUnit) bool) { main.AddFilter(filter) }

// GetLevel 获取系统日志当前的过滤等级
func GetLevel() ELogLevel {
	if main != nil {
		return main.GetLevel()
	}
	return ELL_Infos
}

// ====================

// 构建字段型日志处理器<一>
func Field(field string, val interface{}) ILoger {
	return (&LogFields{}).Init(field, val)
}

// 构建字段型日志处理器<二>
func Fields(fields map[string]interface{}) ILoger {
	return &LogFields{fields: fields}
}

// ====================
func Trace(format string, v ...interface{})             { main.Trace(1, nil, format, v...) }
func Tracev(v ...interface{})                           { main.Tracev(1, nil, v...) }
func TraceD(depth int, format string, v ...interface{}) { main.Trace(depth+1, nil, format, v...) }
func TraceDv(depth int, v ...interface{})               { main.Tracev(depth+1, nil, v...) }

func Debug(format string, v ...interface{})             { main.Debug(1, nil, format, v...) }
func Debugv(v ...interface{})                           { main.Debugv(1, nil, v...) }
func DebugD(depth int, format string, v ...interface{}) { main.Debug(depth+1, nil, format, v...) }
func DebugDv(depth int, v ...interface{})               { main.Debugv(depth+1, nil, v...) }

func Info(format string, v ...interface{})             { main.Info(1, nil, format, v...) }
func Infov(v ...interface{})                           { main.Infov(1, nil, v...) }
func InfoD(depth int, format string, v ...interface{}) { main.Info(depth+1, nil, format, v...) }
func InfoDv(depth int, v ...interface{})               { main.Infov(depth+1, nil, v...) }

func Warn(format string, v ...interface{})             { main.Warn(1, nil, format, v...) }
func Warnv(v ...interface{})                           { main.Warnv(1, nil, v...) }
func WarnD(depth int, format string, v ...interface{}) { main.Warn(depth+1, nil, format, v...) }
func WarnDv(depth int, v ...interface{})               { main.Warnv(depth+1, nil, v...) }

func Error(format string, v ...interface{})             { main.Error(1, nil, format, v...) }
func Errorv(v ...interface{})                           { main.Errorv(1, nil, v...) }
func ErrorD(depth int, format string, v ...interface{}) { main.Error(depth+1, nil, format, v...) }
func ErrorDv(depth int, v ...interface{})               { main.Errorv(depth+1, nil, v...) }

func Fatal(format string, v ...interface{})             { main.Fatal(1, nil, format, v...) }
func Fatalv(v ...interface{})                           { main.Fatalv(1, nil, v...) }
func FatalD(depth int, format string, v ...interface{}) { main.Fatal(depth+1, nil, format, v...) }
func FatalDv(depth int, v ...interface{})               { main.Fatalv(depth+1, nil, v...) }
