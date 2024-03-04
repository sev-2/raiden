package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/build"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/cli/serve"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type RunFlags struct {
	cli.Flags
	Build    build.Flags
	Generate generate.Flags
}

func RunCommand() *cobra.Command {
	var f RunFlags

	cmd := &cobra.Command{
		Use:    "run",
		Short:  "Run app server",
		Long:   "Generate app resource, build binary than serve app",
		PreRun: PreRun(&f.Flags, generate.PreRun),
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

			// 1. generate app resource
			if err := generate.Run(&f.Generate, config, currentDir, false); err != nil {
				logger.Error(err)
				return
			}

			// 2. build app binary
			if err := build.Run(&f.Build, config, currentDir); err != nil {
				logger.Error(err)
				return
			}

			// 3. server application
			if err := serve.Run(config, currentDir); err != nil {
				logger.Error(err)
			}
		},
	}

	f.Build.Bind(cmd)
	f.Generate.Bind(cmd)

	return cmd
}
