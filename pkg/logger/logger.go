package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

var (
	logErrorColor       = "\033[38;5;9m"
	logPanicColor       = "\033[38;5;124m"
	logFatalColor       = "\033[38;5;1m"
	logBlueColor        = "\033[34m"
	logYellowColor      = "\033[1;33m"
	logFilePathColor    = "\033[38;5;250m"
	logLightOrangeColor = "\033[38;5;11m"
	logDefaultColor     = "\033[0m"
	logMessageColor     = "\033[38;5;255m"
)

type LogLevel string

var (
	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
	LogLevelError LogLevel = "error"

	LogFormatWithFile   = log.Ldate | log.Ltime | log.Lshortfile | log.Lmsgprefix
	LogFormatTime       = log.Ldate | log.Ltime | log.Lmsgprefix
	LogFormatPrefixOnly = log.Lmsgprefix
	LogFormatVerbose    = log.Ldate | log.Ltime | log.Llongfile | log.Lmsgprefix
)

func NewLogger() *Logger {
	return &Logger{
		level:       LogLevelInfo,
		outputLevel: 2,
		printLog:    log.New(os.Stdout, "", LogFormatTime),
		errLog:      log.New(os.Stderr, "", LogFormatTime),
	}
}

func NewLoggerWithFormat(format int) *Logger {
	return &Logger{
		level:       LogLevelInfo,
		outputLevel: 2,
		printLog:    log.New(os.Stdout, "", format),
		errLog:      log.New(os.Stderr, "", format),
	}
}

func StdOut(format int) *log.Logger {
	return log.New(os.Stdout, "", format)
}

type Logger struct {
	level       LogLevel
	outputLevel int
	errLog      *log.Logger
	printLog    *log.Logger
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) SetOutputLevel(outputLevel int) {
	l.outputLevel = outputLevel
}

func (l *Logger) GetLevel() LogLevel {
	return l.level
}

func (l *Logger) Info(v ...any) {
	if l.level == LogLevelError {
		return
	}

	msg := []any{logMessageColor}
	msg = append(msg, v...)
	msg = append(msg, logDefaultColor)
	l.printLog.SetPrefix(logBlueColor + "[Info] " + logFilePathColor)
	l.printLog.Output(l.outputLevel, fmt.Sprint(msg...))
}

func (l *Logger) Infof(format string, v ...any) {
	if l.level == LogLevelError {
		return
	}

	format = logMessageColor + format + logDefaultColor + "\n"
	l.printLog.SetPrefix(logBlueColor + "[Info] " + logFilePathColor)
	l.printLog.Output(l.outputLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Json(data any) {
	jStr, err := json.Marshal(data)
	if err != nil {
		l.Error(err)
		return
	}
	l.Info(string(jStr))
}

func (l *Logger) JsonPretty(data any) {
	jStr, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		l.Error(err)
		return
	}
	l.Info(string(jStr))
}

func (l *Logger) Debug(v ...any) {
	if l.level != LogLevelDebug {
		return
	}

	msg := []any{logMessageColor}
	msg = append(msg, v...)
	msg = append(msg, logDefaultColor)
	l.printLog.SetPrefix(logLightOrangeColor + "[Debug] " + logFilePathColor)
	l.printLog.Output(2, fmt.Sprint(msg...))
}

func (l *Logger) Debugf(format string, v ...any) {
	if l.level != LogLevelDebug {
		return
	}

	format = logMessageColor + format + logDefaultColor + "\n"
	l.printLog.SetPrefix(logLightOrangeColor + "[Debug] " + logFilePathColor)
	l.printLog.Output(2, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(v ...any) {
	msg := []any{logMessageColor}
	msg = append(msg, v...)
	msg = append(msg, logDefaultColor)
	l.errLog.SetPrefix(logErrorColor + "[Error] " + logFilePathColor)
	l.errLog.Output(l.outputLevel, fmt.Sprint(msg...))
}

func (l *Logger) Errorf(format string, v ...any) {
	format = logMessageColor + format + logDefaultColor + "\n"
	l.errLog.SetPrefix(logErrorColor + "[Error] " + logFilePathColor)
	l.errLog.Output(l.outputLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Warning(v ...any) {
	if l.level == LogLevelError {
		return
	}

	msg := []any{logMessageColor}
	msg = append(msg, v...)
	msg = append(msg, logDefaultColor)
	l.printLog.SetPrefix(logYellowColor + "[Warning] " + logFilePathColor)
	l.printLog.Output(l.outputLevel, fmt.Sprint(msg...))
}

func (l *Logger) Warningf(format string, v ...any) {
	if l.level == LogLevelError {
		return
	}

	format = logMessageColor + format + logDefaultColor + "\n"
	l.printLog.SetPrefix(logYellowColor + "[Warning] " + logFilePathColor)
	l.printLog.Output(l.outputLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Panic(v ...any) {
	msg := []any{logMessageColor}
	msg = append(msg, v...)
	msg = append(msg, logDefaultColor, "\n")
	l.errLog.SetPrefix(logPanicColor + "[Panic] " + logFilePathColor)

	s := fmt.Sprint(msg...)
	l.errLog.Output(l.outputLevel+1, s)
	panic(s)
}

func (l *Logger) Panicf(format string, v ...any) {
	l.errLog.SetPrefix(logPanicColor + "[Panic] " + logFilePathColor)
	s := fmt.Sprintf(format, v...)
	l.errLog.Output(l.outputLevel+1, s)
	panic(s)
}

func (l *Logger) Fatal(v ...any) {
	msg := []any{logMessageColor}
	msg = append(msg, v...)
	msg = append(msg, logDefaultColor, "\n")
	l.errLog.SetPrefix(logFatalColor + "[Fatal] " + logFilePathColor)
	l.errLog.Output(l.outputLevel, fmt.Sprint(msg...))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...any) {
	format = logMessageColor + format + logDefaultColor + "\n"
	l.errLog.SetPrefix(logFatalColor + "[Fatal] " + logFilePathColor)
	l.errLog.Output(l.outputLevel, fmt.Sprintf(format, v...))
	os.Exit(1)
}
