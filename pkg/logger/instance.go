package logger

var logInstance *Logger

func checkLogInstance() {
	if logInstance == nil {
		logInstance = NewLoggerWithFormat(LogFormatPrefixOnly)
		logInstance.SetOutputLevel(3)
	}
}

func SetLogLevel(level LogLevel) {
	checkLogInstance()
	logInstance.SetLevel(level)
}

func SetFormat(format int) {
	logInstance = NewLoggerWithFormat(format)
}

func Info(v ...any) {
	checkLogInstance()
	logInstance.Info(v...)
}

func Infof(format string, v ...any) {
	checkLogInstance()
	logInstance.Infof(format, v...)
}

func PrintJson(data any, pretty bool) {
	checkLogInstance()
	if pretty {
		logInstance.JsonPretty(data)
	} else {
		logInstance.Json(data)
	}
}

func Debug(v ...any) {
	checkLogInstance()
	logInstance.Debug(v...)
}

func Debugf(format string, v ...any) {
	checkLogInstance()
	logInstance.Debugf(format, v...)
}

func Error(v ...any) {
	checkLogInstance()
	logInstance.Error(v...)
}

func Errorf(format string, v ...any) {
	checkLogInstance()
	logInstance.Errorf(format, v...)
}

func Warning(v ...any) {
	checkLogInstance()
	logInstance.Warning(v...)
}

func Warningf(format string, v ...any) {
	checkLogInstance()
	logInstance.Warningf(format, v...)
}

func Panic(v ...any) {
	checkLogInstance()
	logInstance.Panic(v...)
}

func Panicf(format string, v ...any) {
	checkLogInstance()
	logInstance.Panicf(format, v...)
}

func Fatal(v ...any) {
	checkLogInstance()
	logInstance.Fatal(v...)
}

func Fatalf(format string, v ...any) {
	checkLogInstance()
	logInstance.Fatalf(format, v...)
}
