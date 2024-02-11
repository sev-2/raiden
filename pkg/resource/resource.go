package resource

import (
	"errors"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/spf13/cobra"
)

// Flags is struct to binding options when import and apply is run binart
type Flags struct {
	ProjectPath   string
	RpcOnly       bool
	RolesOnly     bool
	ModelsOnly    bool
	AllowedSchema string
	Verbose       bool
}

// LoadAll is function to check is all resource need to import or apply
func (f *Flags) LoadAll() bool {
	return !f.RpcOnly && !f.RolesOnly && !f.ModelsOnly
}

func (f Flags) CheckAndActivateDebug(cmd *cobra.Command) bool {
	verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")
	if verbose {
		logger.SetDebug()
	}
	return verbose
}

var rpc []raiden.Rpc

func RegisterRpc(list ...raiden.Rpc) {
	rpc = append(rpc, list...)
}

func PreRun(projectPath string) error {
	if !configure.IsConfigExist(projectPath) {
		return errors.New("missing config file (./configs/app.yaml), run `raiden configure` first for generate configuration file")
	}

	return nil
}
