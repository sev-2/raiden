package commands

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/cli/imports"
	init_cmd "github.com/sev-2/raiden/pkg/cli/init"
	"github.com/sev-2/raiden/pkg/cli/version"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

var StartLogger hclog.Logger = logger.HcLog().Named("start")

type StartFlags struct {
	cli.LogFlags
	Configure configure.Flags
	Generate  generate.Flags
	Imports   imports.Flags
	Init      init_cmd.Flags
}

func StartCommand() *cobra.Command {
	var f StartFlags

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start new app",
		Long:  "Start new project, synchronize resource and scaffold application",
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)

			// check latest version
			_, isUpdate, errVersion := version.Run(appVersion)
			if isUpdate {
				if errVersion != nil {
					version.VersionLogger.Error(errVersion.Error())
				}
				os.Exit(0)
			}

			// preparation
			// - get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				StartLogger.Error(errCurDir.Error())
				return
			}

			// 1. run configure
			promptConfig, err := configure.Run(&f.Configure, currentDir)
			if err != nil {
				StartLogger.Error(err.Error())
				return
			}

			StartLogger.Debug("creating new project folder ", "project-name", promptConfig.ProjectName)
			if err = utils.CreateFolder(promptConfig.ProjectName); err != nil {
				StartLogger.Error(err.Error())
				return
			}

			projectPath := filepath.Join(currentDir, promptConfig.ProjectName)
			if err = configure.Generate(&promptConfig.Config, projectPath); err != nil {
				StartLogger.Error(err.Error())
				return
			}

			// 2. running generate
			if executeErr := generate.Run(&f.Generate, &promptConfig.Config, projectPath, true); executeErr != nil {
				StartLogger.Error(executeErr.Error())
				return
			}

			// 3. running init
			if f.Init.Version == "" {
				f.Init.Version = appVersion
			}

			if executeErr := init_cmd.Run(&f.Init, projectPath, utils.ToGoModuleName(promptConfig.ProjectName)); executeErr != nil {
				StartLogger.Error(executeErr.Error())
				return
			}

			// 4. running import
			if executeErr := imports.Run(&f.LogFlags, &f.Imports, projectPath); executeErr != nil {
				StartLogger.Error(executeErr.Error())
				return
			}
		},
	}

	f.Configure.Bind(cmd)
	f.Imports.Bind(cmd)
	f.Generate.Bind(cmd)
	f.Init.Bind(cmd)

	return cmd
}
