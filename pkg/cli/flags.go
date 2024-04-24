package cli

import (
	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/spf13/cobra"
)

type LogFlags struct {
	DebugMode bool
	TraceMode bool
}

func (f *LogFlags) Bind(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&f.DebugMode, "debug", false, "enable log with debug mode")
	cmd.PersistentFlags().BoolVar(&f.TraceMode, "trace", false, "enable log with trace mode")
}

func (f LogFlags) CheckAndActivateDebug(cmd *cobra.Command) {
	if f.DebugMode {
		logger.HcLog().SetLevel(hclog.Debug)
	}

	if f.TraceMode {
		logger.HcLog().SetLevel(hclog.Trace)
	}
}
