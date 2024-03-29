package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	init_cmd "github.com/sev-2/raiden/pkg/cli/init"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type InitFlags struct {
	cli.Flags
	Init init_cmd.Flags
}

func InitCommand() *cobra.Command {
	var f InitFlags

	cmd := &cobra.Command{
		Use:    "init",
		Short:  "Init golang app",
		Long:   "Initialize golang app with install raiden dependency",
		PreRun: PreRun(&f.Flags, init_cmd.PreRun),
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

			if err = init_cmd.Run(&f.Init, currentDir, utils.ToGoModuleName(config.ProjectName)); err != nil {
				logger.Error(err)
			}
		},
	}

	f.Init.Bind(cmd)

	return cmd
}
