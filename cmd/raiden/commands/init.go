package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	init_cmd "github.com/sev-2/raiden/pkg/cli/init"
	"github.com/sev-2/raiden/pkg/cli/version"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type InitFlags struct {
	cli.LogFlags
	Init init_cmd.Flags
}

func InitCommand() *cobra.Command {
	var f InitFlags

	cmd := &cobra.Command{
		Use:    "init",
		Short:  "Init golang app",
		Long:   "Initialize golang app with install raiden dependency",
		PreRun: PreRun(&f.LogFlags, init_cmd.PreRun),
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)

			// check latest version
			version.Run(appVersion)

			// get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				init_cmd.InitLogger.Error(errCurDir.Error())
				return
			}

			// load config
			init_cmd.InitLogger.Info("Load configuration")
			configFilePath := configure.GetConfigFilePath(currentDir)
			init_cmd.InitLogger.Debug("config file information", "path", configFilePath)
			config, err := raiden.LoadConfig(&configFilePath)
			if err != nil {
				init_cmd.InitLogger.Error(err.Error())
				return
			}

			if err = init_cmd.Run(&f.Init, currentDir, utils.ToGoModuleName(config.ProjectName)); err != nil {
				init_cmd.InitLogger.Error(err.Error())
			}
		},
	}

	f.Init.Bind(cmd)

	return cmd
}
