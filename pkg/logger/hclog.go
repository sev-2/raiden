package logger

import (
	"time"

	"github.com/hashicorp/go-hclog"
)

var hcLogInstance hclog.Logger

func HcLog() hclog.Logger {
	if hcLogInstance == nil {
		hcLogInstance = hclog.New(&hclog.LoggerOptions{
			Level:           hclog.Info,
			Color:           hclog.ForceColor,
			ColorHeaderOnly: true,
			TimeFormat:      time.RFC3339,
		})
	}

	return hcLogInstance
}
