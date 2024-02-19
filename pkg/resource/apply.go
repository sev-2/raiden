package resource

import (
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

// TODO : implement all functionality
func Apply(flags *Flags, config *raiden.Config) error {
	// load supabase resource
	resource, err := Load(flags, config)
	if err != nil {
		return err
	}

	// filter table for with allowed schema
	resource.Tables = filterTableBySchema(resource.Tables, strings.Split(flags.AllowedSchema, ",")...)
	resource.Functions = filterFunctionBySchema(resource.Functions, strings.Split(flags.AllowedSchema, ",")...)

	// load app resource
	appTables, appRoles, appRpcFunctions, err := loadAppResource(flags)
	if err != nil {
		return err
	}

	// compare
	if (flags.LoadAll() || flags.ModelsOnly) && len(appTables) > 0 {
		if err := runApplyCompareTable(resource.Tables, appTables); err != nil {
			return err
		}
	}

	if (flags.LoadAll() || flags.RolesOnly) && len(appRoles) > 0 {
		if err := runApplyCompareRoles(resource.Roles, appRoles); err != nil {
			return err
		}
	}

	if (flags.LoadAll() || flags.RpcOnly) && len(appRpcFunctions) > 0 {
		if err := runApplyCompareRpcFunctions(resource.Functions, appRpcFunctions); err != nil {
			return err
		}
	}

	importState := resourceState{
		State: state.State{
			// Roles: nativeStateRoles,
		},
	}

	return applyResource(config, &importState, flags.ProjectPath, resource)
}

func applyResource(config *raiden.Config, importState *resourceState, projectPath string, resource *Resource) error {
	return nil
}

func runApplyCompareTable(supabaseTable []objects.Table, appTable []objects.Table) error {
	return nil
}

func runApplyCompareRoles(supabaseRoles []objects.Role, appRoles []objects.Role) error {
	return nil
}

func runApplyCompareRpcFunctions(supabaseFn []objects.Function, appFn []objects.Function) error {
	return nil
}
