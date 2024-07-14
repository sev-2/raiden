package commands

import (
	"os"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/build"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/cli/serve"
	"github.com/sev-2/raiden/pkg/cli/version"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type RunFlags struct {
	cli.LogFlags
	Build    build.Flags
	Generate generate.Flags
}

func RunCommand() *cobra.Command {
	var f RunFlags

	cmd := &cobra.Command{
		Use:    "run",
		Short:  "Run app server",
		Long:   "Generate app resource, build binary than serve app",
		PreRun: PreRun(&f.LogFlags, generate.PreRun),
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)

			// check latest version
			isLatest, isUpdate, errVersion := version.Run(appVersion)
			if isUpdate {
				if errVersion != nil {
					version.VersionLogger.Error(errVersion.Error())
				}
				os.Exit(0)
			}

			if isLatest {
				if err := version.RunPackage(appVersion); err != nil {
					version.VersionLogger.Error(err.Error())
					return
				}
			}

			// Preparation
			// - get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				serve.ServeLogger.Error(errCurDir.Error())
				return
			}

			// - load config
			configure.ConfigureLogger.Info("Load configuration")
			configFilePath := configure.GetConfigFilePath(currentDir)
			configure.ConfigureLogger.Debug("config file information", "path", configFilePath)
			config, err := raiden.LoadConfig(&configFilePath)
			if err != nil {
				serve.ServeLogger.Error(err.Error())
				return
			}

			// 1. generate app resource
			if err := generate.Run(&f.Generate, config, currentDir, false); err != nil {
				generate.GenerateLogger.Error(err.Error())
				return
			}

			// 2. build app binary
			if err := build.Run(&f.Build, config, currentDir); err != nil {
				build.BuildLogger.Error(err.Error())
				return
			}

			// 3. server application
			if err := serve.Run(config, currentDir); err != nil {
				serve.ServeLogger.Error(err.Error())
			}
		},
	}

	f.Build.Bind(cmd)
	f.Generate.Bind(cmd)

	return cmd
}
