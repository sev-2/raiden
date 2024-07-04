package commands

import (
	"os"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/cli/imports"
	"github.com/sev-2/raiden/pkg/cli/version"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type ImportsFlags struct {
	cli.LogFlags
	Imports  imports.Flags
	Generate generate.Flags
}

func ImportCommand() *cobra.Command {
	var f ImportsFlags

	cmd := &cobra.Command{
		Use:    "imports",
		Short:  "Import supabase resource",
		Long:   "Fetch supabase resource and generate to file",
		PreRun: PreRun(&f.LogFlags, imports.PreRun),
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

			// get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				imports.ImportLogger.Error(errCurDir.Error())
				return
			}

			// load config
			imports.ImportLogger.Info("Load configuration")
			configFilePath := configure.GetConfigFilePath(currentDir)
			imports.ImportLogger.Debug("config file information", "path", configFilePath)
			config, err := raiden.LoadConfig(&configFilePath)
			if err != nil {
				imports.ImportLogger.Error(err.Error())
				return
			}

			// 1. generate all resource
			if err = generate.Run(&f.Generate, config, currentDir, false); err != nil {
				imports.ImportLogger.Error(err.Error())
				return
			}

			// 2. run import
			if err = imports.Run(&f.LogFlags, &f.Imports, currentDir); err != nil {
				imports.ImportLogger.Error(err.Error())
				return
			}
		},
	}

	f.Imports.Bind(cmd)
	f.Generate.Bind(cmd)

	return cmd
}
