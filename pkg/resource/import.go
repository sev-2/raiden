package resource

import (
	"errors"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/postgres/roles"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
)

func Import(flags *Flags, config *raiden.Config) error {
	// // configure supabase adapter
	if config.DeploymentTarget == raiden.DeploymentTargetCloud {
		supabase.ConfigureManagementApi(config.SupabaseApiUrl, config.AccessToken)
	} else {
		supabase.ConfigurationMetaApi(config.SupabaseApiUrl, config.SupabaseApiBaseUrl)
	}

	// load map native role
	mapNativeRole, err := loadMapNativeRole()
	if err != nil {
		return err
	}

	// load supabase resource
	resource, err := Load(flags, config.ProjectId)
	if err != nil {
		return err
	}

	// filter table for with allowed schema
	resource.Tables = filterTableBySchema(resource.Tables, strings.Split(flags.AllowedSchema, ",")...)
	resource.Functions = filterFunctionBySchema(resource.Functions, strings.Split(flags.AllowedSchema, ",")...)
	resource.Roles = filterUserRoleAndBindNativeRole(resource.Roles, mapNativeRole)

	// load app resource
	appTables, appRoles, appRpcFunctions, err := loadAppResource(flags)
	if err != nil {
		return err
	}

	// compare
	if (flags.LoadAll() || flags.ModelsOnly) && len(appTables) > 0 {
		if err := compareTable(resource.Tables, appTables); err != nil {
			return err
		}
	}

	if (flags.LoadAll() || flags.RolesOnly) && len(appRoles) > 0 {
		if err := compareRoles(resource.Roles, appRoles); err != nil {
			return err
		}
	}

	if (flags.LoadAll() || flags.RpcOnly) && len(appRpcFunctions) > 0 {
		if err := compareRpcFunctions(resource.Functions, appRpcFunctions); err != nil {
			return err
		}
	}

	// create import state
	nativeStateRoles, err := createNativeRoleState(mapNativeRole)
	if err != nil {
		return err
	}

	importState := resourceState{
		State: state.State{
			Roles: nativeStateRoles,
		},
	}

	// generate resource
	return generateResource(config, &importState, flags.ProjectPath, resource)
}

func filterTableBySchema(input []supabase.Table, allowedSchema ...string) (output []supabase.Table) {
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

func filterFunctionBySchema(input []supabase.Function, allowedSchema ...string) (output []supabase.Function) {
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

func compareTable(supabaseTable []supabase.Table, appTable []supabase.Table) error {
	diffResult, err := state.CompareTables(supabaseTable, appTable)
	if err != nil {
		return err
	}

	if len(diffResult) > 0 {
		for i := range diffResult {
			d := diffResult[i]
			cli.PrintDiff("table", d.SupabaseResource, d.AppResource, d.Name)
		}
		return errors.New("import tables is canceled, you have conflict table. please fix it first")
	}

	return nil
}

func compareRoles(supabaseRoles []supabase.Role, appRoles []supabase.Role) error {
	diffResult, err := state.CompareRoles(supabaseRoles, appRoles)
	if err != nil {
		return err
	}

	if len(diffResult) > 0 {
		for i := range diffResult {
			d := diffResult[i]
			cli.PrintDiff("role", d.SupabaseResource, d.AppResource, d.Name)
		}
		return errors.New("import roles is canceled, you have conflict role. please fix it first")
	}

	return nil
}

func compareRpcFunctions(supabaseFn []supabase.Function, appFn []supabase.Function) error {
	diffResult, err := state.CompareRpcFunctions(supabaseFn, appFn)
	if err != nil {
		return err
	}

	if len(diffResult) > 0 {
		for i := range diffResult {
			d := diffResult[i]
			cli.PrintDiff("rpc function", d.SupabaseResource, d.AppResource, d.Name)
		}
		return errors.New("import rpc function is canceled, you have conflict rpc function. please fix it first")
	}

	return nil
}

func loadMapNativeRole() (map[string]any, error) {
	mapRole := make(map[string]any)
	for _, r := range roles.NativeRoles {
		role, err := raiden.UnmarshalRole(r)
		if err != nil {
			return nil, err
		}
		mapRole[role.Name] = &role
	}

	return mapRole, nil
}

func filterUserRoleAndBindNativeRole(roles []supabase.Role, mapNativeRole map[string]any) (userRole []supabase.Role) {
	for i := range roles {
		r := roles[i]
		if nr, isExist := mapNativeRole[r.Name]; !isExist {
			userRole = append(userRole, r)
		} else {
			if rl, isRole := nr.(*raiden.Role); isRole {
				rl.ID = r.ID
				rl.ValidUntil = r.ValidUntil
			}
		}
	}
	return
}
