package cli

import (
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/spf13/cobra"
)

type Flags struct {
	Verbose bool
}

func (f *Flags) Bind(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&f.Verbose, "verbose", "v", false, "enable verbose output")
}

func (f Flags) CheckAndActivateDebug(cmd *cobra.Command) {
	verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")
	if verbose {
		logger.SetDebug()
	}
}
