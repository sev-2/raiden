package imports

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/postgres/roles"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/spf13/cobra"
)

type Flags struct {
	RpcOnly       bool
	RolesOnly     bool
	ModelsOnly    bool
	AllowedSchema string
}

type MapTable map[string]*supabase.Table
type MapRelations map[string][]*generator.Relation
type ManyToManyTable struct {
	Table      string
	Schema     string
	PivotTable string
	PrimaryKey string
	ForeignKey string
}

func (f *Flags) Bind(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&f.RpcOnly, "rpc-only", "", false, "import rpc only")
	cmd.Flags().BoolVarP(&f.RolesOnly, "roles-only", "r", false, "import roles only")
	cmd.Flags().BoolVarP(&f.ModelsOnly, "models-only", "m", false, "import models only")
	cmd.Flags().StringVarP(&f.AllowedSchema, "schema", "s", "", "set allowed schema to import, use coma separator for multiple schema")
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
	// configure supabase adapter
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
	resource.Roles = filterUserRoleAndBindNativeRole(resource.Roles, mapNativeRole)

	// load app resource
	appTables, appRoles, err := loadAppResource(flags)
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

	// create import state
	nativeStateRoles, err := createNativeRoleState(mapNativeRole)
	if err != nil {
		return err
	}

	importState := ImportState{
		State: state.State{
			Roles: nativeStateRoles,
		},
	}

	// generate resource
	return generateResource(config, &importState, projectPath, resource)
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

func printDiff(resource string, supabaseResource, appResource interface{}, prefix string) {
	spValue := reflect.ValueOf(supabaseResource)
	appValue := reflect.ValueOf(appResource)

	for i := 0; i < spValue.NumField(); i++ {
		fieldSupabase := spValue.Field(i)
		fieldApp := appValue.Field(i)

		if fieldSupabase.Kind() == reflect.Struct {
			printDiff(resource, fieldSupabase.Interface(), fieldApp.Interface(), fmt.Sprintf("%s.%s", prefix, fieldSupabase.Type().Field(i).Name))
			continue
		}

		if fieldSupabase.Kind() == reflect.Slice || fieldSupabase.Kind() == reflect.Array {
			for j := 0; j < fieldSupabase.Len(); j++ {
				fieldIdentifier := fmt.Sprintf("%v", j)
				spField := fieldSupabase.Index(j)
				spFieldValue := spField.Interface()
				if spColum, isSpColumn := spFieldValue.(supabase.Column); isSpColumn {
					fieldIdentifier = spColum.Name
				}

				if j >= fieldApp.Len() {
					var diffStringArr []string
					for i := 0; i < spField.NumField(); i++ {
						field, fieldValue := spField.Type().Field(i), spField.Field(i)
						diffStringArr = append(diffStringArr, fmt.Sprintf("%s=%v", field.Name, fieldValue.Interface()))
					}
					diffString := strings.Join(diffStringArr, " ")
					printDiffDetail(resource, prefix, spValue.Type().Field(i).Name, diffString, "not configure in app")
				}

				appFieldValue := fieldApp.Index(j).Interface()
				printDiff(resource, spFieldValue, appFieldValue, fmt.Sprintf("%s.%s[%s]", prefix, spValue.Type().Field(i).Name, fieldIdentifier))
			}
			continue
		}

		fieldSupabaseValue := reflect.ValueOf(fieldSupabase.Interface())
		if fieldSupabaseValue.Kind() == reflect.Pointer {
			fieldSupabaseValue = fieldSupabaseValue.Elem()
		}

		fieldAppValue := reflect.ValueOf(fieldApp.Interface())
		if fieldAppValue.Kind() == reflect.Pointer {
			fieldAppValue = fieldAppValue.Elem()
		}

		spCompareStr, appCompateStr := fmt.Sprintf("%v", fieldSupabaseValue), fmt.Sprintf("%v", fieldAppValue)
		if spCompareStr != appCompateStr {
			printDiffDetail(resource, prefix, spValue.Type().Field(i).Name, spCompareStr, appCompateStr)
		}
	}
}

func printDiffDetail(resource, prefix string, attribute string, spValue, appValue string) {
	print := color.New(color.FgHiBlack).PrintfFunc()
	print("*** Found different %s in %s.%s \n", resource, prefix, attribute)
	print = color.New(color.FgGreen).PrintfFunc()
	print("// Supabase : %s = %s\n", attribute, spValue)
	print = color.New(color.FgRed).PrintfFunc()
	print("// App : %s = %s \n", attribute, appValue)
	print = color.New(color.FgHiBlack).PrintfFunc()
	print("*** End different \n")
}

func compareTable(supabaseTable []supabase.Table, appTable []supabase.Table) error {
	diffResult, err := state.CompareTables(supabaseTable, appTable)
	if err != nil {
		return err
	}

	if len(diffResult) > 0 {
		for i := range diffResult {
			d := diffResult[i]
			printDiff("table", d.SupabaseResource, d.AppResource, d.Name)
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
			printDiff("role", d.SupabaseResource, d.AppResource, d.Name)
		}
		return errors.New("import roles is canceled, you have conflict role. please fix it first")
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
