package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type GenerateFlags struct {
	cli.Flags
	Generate generate.Flags
}

func GenerateCommand() *cobra.Command {
	f := GenerateFlags{}

	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate application resource",
		Long:    "Generate route and rpc configuration",
		PreRunE: PreRunE(&f.Flags, generate.PreRun),
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

			return generate.Run(&f.Generate, config, currentDir, false)
		},
	}

	f.Generate.Bind(cmd)
	return cmd
}
