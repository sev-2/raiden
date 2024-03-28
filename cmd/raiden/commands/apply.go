package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/apply"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type ApplyFlags struct {
	cli.Flags
	Apply    apply.Flags
	Generate generate.Flags
}

func ApplyCommand() *cobra.Command {
	var f ApplyFlags

	cmd := &cobra.Command{
		Use:    "apply",
		Short:  "Apply resource to supabase",
		Long:   "Apply model, role, rpc and rls to supabase",
		PreRun: PreRun(&f.Flags, apply.PreRun),
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
			if err = apply.Run(&f.Apply, currentDir, verbose); err != nil {
				logger.Error(err)
			}
		},
	}

	f.Apply.Bind(cmd)
	f.Generate.Bind(cmd)

	return cmd
}
