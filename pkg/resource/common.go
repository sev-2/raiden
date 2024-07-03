package resource

import (
	"errors"
	"time"

	"github.com/hashicorp/go-hclog"
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
	StoragesOnly  bool
	AllowedSchema string
	DebugMode     bool
	TraceMode     bool
	Generate      generate.Flags
	DryRun        bool
}

// LoadAll is function to check is all resource need to import or apply
func (f *Flags) All() bool {
	return !f.RpcOnly && !f.RolesOnly && !f.ModelsOnly && !f.StoragesOnly
}

func (f *Flags) BindLog(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&f.DebugMode, "debug", false, "enable log with debug mode")
	cmd.PersistentFlags().BoolVar(&f.TraceMode, "trace", false, "enable log with trace mode")
}

func (f Flags) CheckAndActivateDebug(cmd *cobra.Command) {
	if f.DebugMode {
		logger.HcLog().SetLevel(hclog.Debug)
	}

	if f.TraceMode {
		logger.HcLog().SetLevel(hclog.Trace)
	}
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

// ----- Handle register models -----
var RegisteredModels []any

func RegisterModels(list ...any) {
	RegisteredModels = append(RegisteredModels, list...)
}

// ----- Handle register storages -----
var registeredStorages []raiden.Bucket

func RegisterStorages(list ...raiden.Bucket) {
	registeredStorages = append(registeredStorages, list...)
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

// ---- extract resource -----
func extractAppResource(f *Flags, latestState *state.State) (
	extractedTable state.ExtractTableResult, extractedRole state.ExtractRoleResult,
	extractedRpc state.ExtractRpcResult, extractedStorage state.ExtractStorageResult,
	err error,
) {
	if latestState == nil {
		return
	}

	if f.StoragesOnly || f.ModelsOnly {
		ImportLogger.Debug("Start extract role")
		extractedRole, err = state.ExtractRole(latestState.Roles, registeredRoles, false)
		if err != nil {
			return
		}
		ImportLogger.Debug("Finish extract role")
	}

	if f.All() || f.ModelsOnly {
		ImportLogger.Debug("Start extract table")
		extractedTable, err = state.ExtractTable(latestState.Tables, RegisteredModels)
		if err != nil {
			return
		}
		ImportLogger.Debug("Finish extract table")
	}

	if f.All() || f.RolesOnly {
		ImportLogger.Debug("Start extract role")
		extractedRole, err = state.ExtractRole(latestState.Roles, registeredRoles, false)
		if err != nil {
			return
		}
		ImportLogger.Debug("Finish extract role")
	}

	if f.All() || f.RpcOnly {
		ImportLogger.Debug("Start extract rpc")
		extractedRpc, err = state.ExtractRpc(latestState.Rpc, registeredRpc)
		if err != nil {
			return
		}
		ImportLogger.Debug("FInish extract rpc")
	}

	if f.All() || f.StoragesOnly {
		ImportLogger.Debug("Start extract storage")
		extractedStorage, err = state.ExtractStorage(latestState.Storage, registeredStorages)
		if err != nil {
			return
		}
		ImportLogger.Debug("Finish extract storage")
	}

	return
}
