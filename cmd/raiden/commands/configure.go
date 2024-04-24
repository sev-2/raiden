package commands

import (
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

// The `type ConfigureFlags struct` is defining a struct type called `ConfigureFlags`. This struct is
// used to hold the flags and options that are used for the `configure` command. It embeds the
// `cli.Flags` struct, which provides common flags and options for the CLI, and also has a field called
// `Configure` of type `configure.Flags`, which holds the specific flags and options for the
// `configure` command.
type ConfigureFlags struct {
	cli.LogFlags
	Configure configure.Flags
}

func ConfigureCommand() *cobra.Command {
	var f ConfigureFlags

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure project",
		Long:  "Configure project and generate config file",
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)

			// get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				configure.ConfigureLogger.Error(errCurDir.Error())
				return
			}

			config, err := configure.Run(&f.Configure, currentDir)
			if err != nil {
				configure.ConfigureLogger.Error(err.Error())
				return
			}

			if err = configure.Generate(&config.Config, currentDir); err != nil {
				configure.ConfigureLogger.Error(err.Error())
			}
		},
	}

	f.Configure.Bind(cmd)

	return cmd
}
