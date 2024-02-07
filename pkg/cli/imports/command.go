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
	"github.com/sev-2/raiden/pkg/logger"
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

// The `Bind` function is a method of the `Flags` struct. It takes a `cmd` parameter of type
// `*cobra.Command`, which represents a command in the Cobra library for building command-line
// applications.
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
	if config.DeploymentTarget == raiden.DeploymentTargetCloud {
		supabase.ConfigureManagementApi(config.SupabaseApiUrl, config.AccessToken)
	} else {
		supabase.ConfigurationMetaApi(config.SupabaseApiUrl, config.SupabaseApiBaseUrl)
	}

	// load supabase tables
	resource, err := Load(flags, config.ProjectId)
	if err != nil {
		return err
	}
	resource.Tables = filterTableBySchema(resource.Tables, strings.Split(flags.AllowedSchema, ",")...)

	// load app tables
	appTables, err := loadAppTables()
	if err != nil {
		return err
	}

	// compare table
	diffResult, err := state.CompareTables(resource.Tables, appTables)
	if err != nil {
		return err
	}

	if len(diffResult) > 0 {
		for i := range diffResult {
			d := diffResult[i]
			logger.Debug("[pkg.cli.imports] ", d.Name)
			printDiff(d.SupabaseResource, d.AppResource, d.Name)
		}
		return errors.New("canceled import resource process, you have conflict table. please fix it first")
	}

	// generate resource
	return generateResource(config, projectPath, resource)
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

func loadAppTables() (appTable []supabase.Table, err error) {
	// load app table
	latestState, err := state.Load()
	if err != nil {
		return appTable, err
	}

	return state.ToSupabaseTable(latestState.Tables)
}

func printDiff(supabaseResource, appResource interface{}, prefix string) {
	spValue := reflect.ValueOf(supabaseResource)
	appValue := reflect.ValueOf(appResource)

	for i := 0; i < spValue.NumField(); i++ {
		fieldSupabase := spValue.Field(i)
		fieldApp := appValue.Field(i)

		if fieldSupabase.Kind() == reflect.Struct {
			printDiff(fieldSupabase.Interface(), fieldApp.Interface(), fmt.Sprintf("%s.%s", prefix, fieldSupabase.Type().Field(i).Name))
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
					printDiffDetail(prefix, spValue.Type().Field(i).Name, diffString, "not configure in app")
				}

				appFieldValue := fieldApp.Index(j).Interface()
				// var appFieldValue any
				// if fieldApp.Len()-1 > j {
				// 	fieldApp.Index(j).Interface()
				// }

				// if appFieldValue == nil {
				// 	logger.Info("%+v", spFieldValue)
				// 	printDiffDetail(prefix, spValue.Type().Field(i).Name, "ok", "not configure in app")
				// 	continue
				// }

				printDiff(spFieldValue, appFieldValue, fmt.Sprintf("%s.%s[%s]", prefix, spValue.Type().Field(i).Name, fieldIdentifier))
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
			printDiffDetail(prefix, spValue.Type().Field(i).Name, spCompareStr, appCompateStr)
		}
	}
}

func printDiffDetail(prefix string, attribute string, spValue, appValue string) {
	print := color.New(color.FgHiBlack).PrintfFunc()
	print("*** Found different in %s.%s \n", prefix, attribute)
	print = color.New(color.FgGreen).PrintfFunc()
	print("// Supabase : %s = %s\n", attribute, spValue)
	print = color.New(color.FgRed).PrintfFunc()
	print("// App : %s = %s \n", attribute, appValue)
	print = color.New(color.FgHiBlack).PrintfFunc()
	print("*** End different \n")
}
