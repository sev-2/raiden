package imports

import (
	"errors"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/spf13/cobra"
)

// The `Flags` type represents a set of boolean flags used for command line options.
// @property {bool} RpcOnly - A boolean flag indicating whether to import only supabase RPC function.
// @property {bool} RolesOnly - A boolean flag indicating whether to import roles only.
// @property {bool} ModelsOnly - A boolean flag indicating whether to import models only.
type Flags struct {
	RpcOnly    bool
	RolesOnly  bool
	ModelsOnly bool
}

// The `Bind` function is a method of the `Flags` struct. It takes a `cmd` parameter of type
// `*cobra.Command`, which represents a command in the Cobra library for building command-line
// applications.
func (f *Flags) Bind(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&f.RpcOnly, "rpc-only", "", false, "import rpc only")
	cmd.Flags().BoolVarP(&f.RolesOnly, "roles-only", "r", false, "import roles only")
	cmd.Flags().BoolVarP(&f.ModelsOnly, "models-only", "m", false, "import models only")
}

func (f *Flags) LoadAll() bool {
	return !f.RpcOnly && !f.RolesOnly && !f.ModelsOnly
}

func PreRun(projectPath string) error {
	if !configure.IsConfigExist(projectPath) {
		return errors.New("missing config file (./configs/app.yaml), run `raiden configure` first for generate configuration file")
	}

	return nil
}

func Run(flags *Flags, config *raiden.Config, projectPath string) error {
	resource, err := Load(flags, config.ProjectId)
	if err != nil {
		return err
	}

	return generateResource(config, projectPath, resource)
}
