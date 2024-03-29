package commands

import (
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

func PreRun(f *cli.Flags, callbacks ...func(path string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		f.Bind(cmd)

		// get current directory
		currentDir, errCurDir := utils.GetCurrentDirectory()
		if errCurDir != nil {
			logger.Error(errCurDir)
			return
		}

		for i := range callbacks {
			c := callbacks[i]
			if err := c(currentDir); err != nil {
				logger.Error(errCurDir)
				return
			}
		}
	}
}
