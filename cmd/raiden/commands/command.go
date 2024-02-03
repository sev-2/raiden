package commands

import (
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

func PreRunE(f *cli.Flags, callbacks ...func(path string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		f.Bind(cmd)

		// get current directory
		currentDir, errCurDir := utils.GetCurrentDirectory()
		if errCurDir != nil {
			return errCurDir
		}

		for i := range callbacks {
			c := callbacks[i]
			if err := c(currentDir); err != nil {
				return err
			}
		}

		return nil
	}
}
