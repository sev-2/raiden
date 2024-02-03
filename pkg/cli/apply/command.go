package apply

import (
	"errors"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/spf13/cobra"
)

type Flags struct {
	RpcOnly    bool
	RolesOnly  bool
	ModelsOnly bool
}

func (f *Flags) Bind(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&f.RpcOnly, "rpc-only", "", false, "apply rpc only")
	cmd.Flags().BoolVarP(&f.RolesOnly, "roles-only", "r", false, "apply roles only")
	cmd.Flags().BoolVarP(&f.ModelsOnly, "models-only", "m", false, "apply models only")
}

func (f *Flags) ApplyAll() bool {
	return !f.RpcOnly && !f.RolesOnly && !f.ModelsOnly
}

func PreRun(projectPath string) error {
	if !configure.IsConfigExist(projectPath) {
		return errors.New("missing config file (./configs/app.yaml), run `raiden configure` first for generate configuration file")
	}

	return nil
}

func Run(flags *Flags, config *raiden.Config, projectPath string) error {
	return nil
}
