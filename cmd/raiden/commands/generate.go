package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type GenerateFlags struct {
	cli.LogFlags
	Generate generate.Flags
}

func GenerateCommand() *cobra.Command {
	f := GenerateFlags{}

	cmd := &cobra.Command{
		Use:    "generate",
		Short:  "Generate application resource",
		Long:   "Generate route and rpc configuration",
		PreRun: PreRun(&f.LogFlags, generate.PreRun),
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)

			// get current directory
			currentDir, errCurDir := utils.GetCurrentDirectory()
			if errCurDir != nil {
				generate.GenerateLogger.Error(errCurDir.Error())
				return
			}

			// load config
			generate.GenerateLogger.Info("Load configuration")
			configFilePath := configure.GetConfigFilePath(currentDir)
			generate.GenerateLogger.Debug("config file information", "path", configFilePath)
			config, err := raiden.LoadConfig(&configFilePath)
			if err != nil {
				generate.GenerateLogger.Error(err.Error())
				return
			}

			if err = generate.Run(&f.Generate, config, currentDir, false); err != nil {
				generate.GenerateLogger.Error(err.Error())
			}
		},
	}

	f.Generate.Bind(cmd)
	return cmd
}
