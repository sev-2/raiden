package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/build"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/cli/version"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type BuildFlags struct {
	cli.LogFlags
	Build    build.Flags
	Generate generate.Flags
}

func BuildCommand() *cobra.Command {
	var f BuildFlags

	cmd := &cobra.Command{
		Use:    "build",
		Short:  "Build app binary",
		Long:   "Build app binary base on configuration",
		PreRun: PreRun(&f.LogFlags, build.PreRun),
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)

			// check latest version
			version.Run(appVersion)

			// Preparation
			// - get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				build.BuildLogger.Error(errCurDir.Error())
				return
			}

			// - load config
			build.BuildLogger.Info("Load configuration")
			configFilePath := configure.GetConfigFilePath(currentDir)
			build.BuildLogger.Debug("config file information", "path", configFilePath)
			config, err := raiden.LoadConfig(&configFilePath)
			if err != nil {
				build.BuildLogger.Error(err.Error())
				return
			}

			// 1. generate
			if err := generate.Run(&f.Generate, config, currentDir, false); err != nil {
				build.BuildLogger.Error(err.Error())
				return
			}

			// 2. build app
			if err := build.Run(&f.Build, config, currentDir); err != nil {
				build.BuildLogger.Error(err.Error())
			}
		},
	}

	f.Build.Bind(cmd)
	f.Generate.Bind(cmd)

	return cmd
}
