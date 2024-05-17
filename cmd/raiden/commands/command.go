package commands

import (
	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

func PreRun(f *cli.LogFlags, callbacks ...func(path string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		f.Bind(cmd)

		// get current directory
		currentDir, errCurDir := utils.GetCurrentDirectory()
		if errCurDir != nil {
			hclog.Default().Error(errCurDir.Error())
			return
		}

		for i := range callbacks {
			c := callbacks[i]
			if err := c(currentDir); err != nil {
				hclog.Default().Error(err.Error())
				return
			}
		}
	}
}
