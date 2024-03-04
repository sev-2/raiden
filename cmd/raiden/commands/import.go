package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/cli/imports"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type ImportsFlags struct {
	cli.Flags
	Imports  imports.Flags
	Generate generate.Flags
}

func ImportCommand() *cobra.Command {
	var f ImportsFlags

	cmd := &cobra.Command{
		Use:    "imports",
		Short:  "Import supabase resource",
		Long:   "Fetch supabase resource and generate to file",
		PreRun: PreRun(&f.Flags, imports.PreRun),
		Run: func(cmd *cobra.Command, args []string) {
			verbose := f.CheckAndActivateDebug(cmd)

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

			// 1. generate all resource
			if err = generate.Run(&f.Generate, config, currentDir, false); err != nil {
				logger.Error(err)
				return
			}

			// 2. run import
			if err = imports.Run(&f.Imports, currentDir, verbose); err != nil {
				logger.Error(err)
				return
			}
		},
	}

	f.Imports.Bind(cmd)
	f.Generate.Bind(cmd)

	return cmd
}
