package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/serve"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type ServeFlags struct {
	cli.Flags
}

func ServeCommand() *cobra.Command {
	var f ServeFlags

	cmd := &cobra.Command{
		Use:    "serve",
		Short:  "Serve app binary",
		Long:   "Serve build app binary file",
		PreRun: PreRun(&f.Flags, serve.PreRun),
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)

			// get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				logger.Error(errCurDir)
				return
			}

			// load config
			configFilePath := configure.GetConfigFilePath(currentDir)
			logger.Debug("Load configuration from : ", configFilePath)
			config, err := raiden.LoadConfig(&configFilePath)
			if err != nil {
				logger.Error(err)
				return
			}

			if err = serve.Run(config, currentDir); err != nil {
				logger.Error(err)
			}
		},
	}

	return cmd
}
