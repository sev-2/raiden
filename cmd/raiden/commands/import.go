package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/imports"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type ImportsFlags struct {
	cli.Flags
	Imports imports.Flags
}

func ImportCommand() *cobra.Command {
	var f ImportsFlags

	cmd := &cobra.Command{
		Use:     "imports",
		Short:   "Import supabase resource",
		Long:    "Fetch supabase resource and generate to file",
		PreRunE: PreRunE(&f.Flags, imports.PreRun),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.CheckAndActivateDebug(cmd)

			// get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				return errCurDir
			}

			// load config
			configFilePath := configure.GetConfigFilePath(currentDir)
			logger.Debug("Load configuration from : ", configFilePath)
			config, err := raiden.LoadConfig(&configFilePath)
			if err != nil {
				return err
			}

			return imports.Run(&f.Imports, config, currentDir)
		},
	}

	f.Imports.Bind(cmd)

	return cmd
}
