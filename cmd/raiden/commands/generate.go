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
		Use:    "generate",
		Short:  "Generate application resource",
		Long:   "Generate route and rpc configuration",
		PreRun: PreRun(&f.Flags, generate.PreRun),
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)

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

			if err = generate.Run(&f.Generate, config, currentDir, false); err != nil {
				logger.Error(err)
			}
		},
	}

	f.Generate.Bind(cmd)
	return cmd
}
