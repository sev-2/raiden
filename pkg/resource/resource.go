package resource

import (
	"errors"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/cli/generate"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
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
	Generate      generate.Flags
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

func PreRun(projectPath string) error {
	if !configure.IsConfigExist(projectPath) {
		return errors.New("missing config file (./configs/app.yaml), run `raiden configure` first for generate configuration file")
	}

	return nil
}

// ----- Handle register rpc -----
var registeredRpc []raiden.Rpc

func RegisterRpc(list ...raiden.Rpc) {
	registeredRpc = append(registeredRpc, list...)
}

// ----- Handle register roles -----
var registeredRoles []raiden.Role

func RegisterRole(list ...raiden.Role) {
	registeredRoles = append(registeredRoles, list...)
}

// ----- Filter function -----
func filterTableBySchema(input []objects.Table, allowedSchema ...string) (output []objects.Table) {
	filterSchema := []string{"public"}
	if len(allowedSchema) > 0 && allowedSchema[0] != "" {
		filterSchema = allowedSchema
	}

	mapSchema := map[string]bool{}
	for _, s := range filterSchema {
		mapSchema[s] = true
	}

	for i := range input {
		t := input[i]

		if _, exist := mapSchema[t.Schema]; exist {
			output = append(output, t)
		}
	}

	return
}

func filterFunctionBySchema(input []objects.Function, allowedSchema ...string) (output []objects.Function) {
	filterSchema := []string{"public"}
	if len(allowedSchema) > 0 && allowedSchema[0] != "" {
		filterSchema = allowedSchema
	}

	mapSchema := map[string]bool{}
	for _, s := range filterSchema {
		mapSchema[s] = true
	}

	for i := range input {
		t := input[i]

		if _, exist := mapSchema[t.Schema]; exist {
			output = append(output, t)
		}
	}

	return
}

func filterUserRole(roles []objects.Role, mapNativeRole map[string]raiden.Role) (userRole []objects.Role) {
	for i := range roles {
		r := roles[i]
		if _, isExist := mapNativeRole[r.Name]; !isExist {
			userRole = append(userRole, r)
		}
	}
	return
}

func filterIsNativeRole(mapNativeRole map[string]raiden.Role, supabaseRole []objects.Role) (nativeRoles []state.RoleState) {
	for i := range supabaseRole {
		r := supabaseRole[i]
		if _, isExist := mapNativeRole[r.Name]; !isExist {
			continue
		} else {
			nativeRoles = append(nativeRoles, state.RoleState{
				Role:       r,
				IsNative:   true,
				LastUpdate: time.Now(),
			})
		}
	}

	return
}
