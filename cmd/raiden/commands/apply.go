package commands

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/apply"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type ApplyFlags struct {
	cli.Flags
	Apply apply.Flags
}

func ApplyCommand() *cobra.Command {
	var f ApplyFlags

	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Apply resource to supabase",
		Long:    "Apply model, role, rpc and rls to supabase",
		PreRunE: PreRunE(&f.Flags, apply.PreRun),
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

			return apply.Run(&f.Apply, config, currentDir)
		},
	}

	f.Apply.Bind(cmd)

	return cmd
}
