package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/serve"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type ServeFlags struct {
	cli.LogFlags
}

func ServeCommand() *cobra.Command {
	var f ServeFlags

	cmd := &cobra.Command{
		Use:    "serve",
		Short:  "Serve app binary",
		Long:   "Serve build app binary file",
		PreRun: PreRun(&f.LogFlags, serve.PreRun),
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)

			// get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				serve.ServeLogger.Error(errCurDir.Error())
				return
			}

			// load config
			serve.ServeLogger.Info("Load configuration")
			configFilePath := configure.GetConfigFilePath(currentDir)
			serve.ServeLogger.Debug("config file information", "path", configFilePath)
			config, err := raiden.LoadConfig(&configFilePath)
			if err != nil {
				serve.ServeLogger.Error(err.Error())
				return
			}

			if err = serve.Run(config, currentDir); err != nil {
				serve.ServeLogger.Error(err.Error())
			}
		},
	}

	return cmd
}
