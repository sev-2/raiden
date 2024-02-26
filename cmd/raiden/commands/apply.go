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
		Use:     "apply",
		Short:   "Apply resource to supabase",
		Long:    "Apply model, role, rpc and rls to supabase",
		PreRunE: PreRunE(&f.Flags, apply.PreRun),
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose := f.CheckAndActivateDebug(cmd)

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

			// 1. generate all resource
			if err = generate.Run(&f.Generate, config, currentDir, false); err != nil {
				return err
			}

			// 2. run import
			return apply.Run(&f.Apply, currentDir, verbose)
		},
	}

	f.Apply.Bind(cmd)
	f.Generate.Bind(cmd)

	return cmd
}
