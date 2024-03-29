package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/build"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type BuildFlags struct {
	cli.Flags
	Build    build.Flags
	Generate generate.Flags
}

func BuildCommand() *cobra.Command {
	var f BuildFlags

	cmd := &cobra.Command{
		Use:    "build",
		Short:  "Build app binary",
		Long:   "Build app binary base on configuration",
		PreRun: PreRun(&f.Flags, build.PreRun),
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)

			// Preparation
			// - get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				logger.Error(errCurDir)
				return
			}

			// - load config
			configFilePath := configure.GetConfigFilePath(currentDir)
			logger.Debug("Load configuration from : ", configFilePath)
			config, err := raiden.LoadConfig(&configFilePath)
			if err != nil {
				logger.Error(err)
				return
			}

			// 1. generate
			if err := generate.Run(&f.Generate, config, currentDir, false); err != nil {
				logger.Error(err)
				return
			}

			// 2. build app
			if err := build.Run(&f.Build, config, currentDir); err != nil {
				logger.Error(err)
			}
		},
	}

	f.Build.Bind(cmd)
	f.Generate.Bind(cmd)

	return cmd
}
