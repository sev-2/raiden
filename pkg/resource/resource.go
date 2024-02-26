package resource

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/color"
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
func (f *Flags) All() bool {
	return !f.RpcOnly && !f.RolesOnly && !f.ModelsOnly
}

func (f Flags) CheckAndActivateDebug(cmd *cobra.Command) bool {
	verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")
	if verbose {
		logger.Info("set logger to debug mode")
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

// ----- Handle register models -----
var registeredModels []any

func RegisterModels(list ...any) {
	registeredModels = append(registeredModels, list...)
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

// --- print diff -----
func PrintDiff(resource string, supabaseResource, appResource interface{}, prefix string) {
	spValue := reflect.ValueOf(supabaseResource)
	appValue := reflect.ValueOf(appResource)

	for i := 0; i < spValue.NumField(); i++ {
		fieldSupabase := spValue.Field(i)
		fieldApp := appValue.Field(i)

		if fieldSupabase.Kind() == reflect.Struct {
			PrintDiff(resource, fieldSupabase.Interface(), fieldApp.Interface(), fmt.Sprintf("%s.%s", prefix, fieldSupabase.Type().Field(i).Name))
			continue
		}

		if fieldSupabase.Kind() == reflect.Slice || fieldSupabase.Kind() == reflect.Array {
			for j := 0; j < fieldSupabase.Len(); j++ {
				fieldIdentifier := fmt.Sprintf("%v", j)
				spField := fieldSupabase.Index(j)
				spFieldValue := spField.Interface()
				if spColum, isSpColumn := spFieldValue.(objects.Column); isSpColumn {
					fieldIdentifier = spColum.Name
				}

				if j >= fieldApp.Len() {
					var diffStringArr []string
					for i := 0; i < spField.NumField(); i++ {
						field, fieldValue := spField.Type().Field(i), spField.Field(i)
						diffStringArr = append(diffStringArr, fmt.Sprintf("%s=%v", field.Name, fieldValue.Interface()))
					}
					diffString := strings.Join(diffStringArr, " ")
					PrintDiffDetail(resource, prefix, spValue.Type().Field(i).Name, diffString, "not configure in app")
					continue
				}

				appFieldValue := fieldApp.Index(j).Interface()
				PrintDiff(resource, spFieldValue, appFieldValue, fmt.Sprintf("%s.%s[%s]", prefix, spValue.Type().Field(i).Name, fieldIdentifier))
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
			PrintDiffDetail(resource, prefix, spValue.Type().Field(i).Name, spCompareStr, appCompateStr)
		}
	}
}

func PrintDiffDetail(resource, prefix string, attribute string, spValue, appValue string) {
	print := color.New(color.FgHiBlack).PrintfFunc()
	print("*** Found diff %s in %s.%s \n", resource, prefix, attribute)
	print = color.New(color.FgGreen).PrintfFunc()
	print("// Supabase : %s = %s\n", attribute, spValue)
	print = color.New(color.FgRed).PrintfFunc()
	print("// App : %s = %s \n", attribute, appValue)
	print = color.New(color.FgHiBlack).PrintfFunc()
	print("*** End found diff \n")
}
