package raiden

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/logger"
)

var logInstance hclog.Logger

func checkLogInstance() {
	if logInstance == nil {
		logInstance = logger.HcLog()
	}
}

func SetLogLevel(level hclog.Level) {
	checkLogInstance()
	logInstance.SetLevel(level)
}

func GetLogLevel() hclog.Level {
	return logInstance.GetLevel()
}

func Info(message string, v ...any) {
	checkLogInstance()
	logInstance.Info(message, v...)
}

func Debug(message string, v ...any) {
	checkLogInstance()
	logInstance.Debug(message, v...)
}

func Error(message string, v ...any) {
	checkLogInstance()
	logInstance.Error(message, v...)
}

func Warning(message string, v ...any) {
	logInstance.Warn(message, v...)
}

func Panic(message string) {
	checkLogInstance()
	panic(message)
}

func Fatal(message string, v ...any) {
	checkLogInstance()
	logInstance.Error(message, v...)
	os.Exit(1)
}
