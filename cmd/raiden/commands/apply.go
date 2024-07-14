package commands

import (
	"os"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/apply"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/cli/version"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type ApplyFlags struct {
	cli.LogFlags
	Apply    apply.Flags
	Generate generate.Flags
}

func ApplyCommand() *cobra.Command {
	var f ApplyFlags

	cmd := &cobra.Command{
		Use:    "apply",
		Short:  "Apply resource to supabase",
		Long:   "Apply model, role, rpc and rls to supabase",
		PreRun: PreRun(&f.LogFlags, apply.PreRun),
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)

			// check latest versions
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
				apply.ApplyLogger.Error(errCurDir.Error())
				return
			}

			// load config
			apply.ApplyLogger.Info("load configuration")
			configFilePath := configure.GetConfigFilePath(currentDir)
			apply.ApplyLogger.Debug("config file information", "path", configFilePath)
			config, err := raiden.LoadConfig(&configFilePath)
			if err != nil {
				apply.ApplyLogger.Error(err.Error())
				return
			}

			// 1. generate all resource
			if err = generate.Run(&f.Generate, config, currentDir, false); err != nil {
				apply.ApplyLogger.Error(err.Error())
				return
			}

			// 2. run import
			if err = apply.Run(&f.LogFlags, &f.Apply, currentDir); err != nil {
				apply.ApplyLogger.Error(err.Error())
			}
		},
	}

	f.Apply.Bind(cmd)
	f.Generate.Bind(cmd)

	return cmd
}
