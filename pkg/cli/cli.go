package cli

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"github.com/sev-2/raiden/pkg/supabase"
)

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
					PrintDiffDetail(resource, prefix, spValue.Type().Field(i).Name, diffString, "not configure in app")
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
