package commands

import (
	"path/filepath"

	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/cli/imports"
	init_cmd "github.com/sev-2/raiden/pkg/cli/init"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type StartFlags struct {
	cli.Flags
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
		RunE: func(cmd *cobra.Command, args []string) error {
			debug := f.CheckAndActivateDebug(cmd)

			// preparation
			// - get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				return errCurDir
			}

			// 1. run configure
			promptConfig, err := configure.Run(&f.Configure, currentDir)
			if err != nil {
				return err
			}

			logger.Debug("creating new project folder : ", promptConfig.ProjectName)
			if err = utils.CreateFolder(promptConfig.ProjectName); err != nil {
				return err
			}

			logger.Debug("start generate configuration file")
			projectPath := filepath.Join(currentDir, promptConfig.ProjectName)
			if err = configure.Generate(&promptConfig.Config, projectPath); err != nil {
				return err
			}

			// 2. running generate
			if executeErr := generate.Run(&f.Generate, &promptConfig.Config, projectPath, true); executeErr != nil {
				return executeErr
			}

			// 3. running init
			if executeErr := init_cmd.Run(&f.Init, projectPath, utils.ToGoModuleName(promptConfig.ProjectName)); executeErr != nil {
				return executeErr
			}

			// 4. running import
			return imports.Run(&f.Imports, projectPath, debug)
		},
	}

	f.Configure.Bind(cmd)
	f.Imports.Bind(cmd)
	f.Generate.Bind(cmd)
	f.Init.Bind(cmd)

	return cmd
}
