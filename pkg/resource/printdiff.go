package resource

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type PrintDiffType string

const (
	PrintDiffTypeStorage PrintDiffType = "storage"
	PrintDiffTypeRole    PrintDiffType = "role"
	PrintDiffTypeModel   PrintDiffType = "model"
	PrintDiffTypeRpc     PrintDiffType = "rpc"
)

func PrintDiffResource(diffType PrintDiffType, diffData any) {
	switch diffType {
	case PrintDiffTypeStorage:
		if diffParse, isType := diffData.(CompareDiffResult[objects.Bucket, objects.UpdateBucketParam]); isType {
			PrintDiffStorage(diffParse)
		}
	case PrintDiffTypeRole:
		if diffParse, isType := diffData.(CompareDiffResult[objects.Role, objects.UpdateRoleParam]); isType {
			PrintDiffRole(diffParse)
		}
	}
}

func PrintDiffRole(diffData CompareDiffResult[objects.Role, objects.UpdateRoleParam]) {
	if len(diffData.DiffItems.ChangeItems) == 0 {
		return
	}
	fileName := utils.ToSnakeCase(diffData.TargetResource.Name)
	structName := utils.SnakeCaseToPascalCase(fileName)

	print := color.New(color.FgWhite).SprintfFunc()
	printScope := color.New(color.FgHiBlue).PrintfFunc()
	printLabel := color.New(color.FgHiBlack).SprintfFunc()
	printAdd := color.New(color.FgHiGreen).SprintfFunc()
	printRemove := color.New(color.FgHiRed).SprintfFunc()
	printUpdate := color.New(color.FgHiYellow).SprintfFunc()

	changes := make([]string, 0)
	for _, v := range diffData.DiffItems.ChangeItems {
		switch v {
		case objects.UpdateConnectionLimit:
			if diffData.TargetResource.ConnectionLimit == 0 && diffData.SourceResource.ConnectionLimit > 0 {
				symbol := printAdd("+")
				changeDetail := []string{}
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) ConnectionLimit() int {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", strconv.Itoa(diffData.SourceResource.ConnectionLimit))))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			}

			if diffData.TargetResource.ConnectionLimit > 0 && diffData.SourceResource.ConnectionLimit == 0 {
				symbol := printRemove("-")
				changeDetail := []string{}
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) ConnectionLimit() int {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", strconv.Itoa(diffData.SourceResource.ConnectionLimit))))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			}

			if diffData.TargetResource.ConnectionLimit > 0 && diffData.SourceResource.ConnectionLimit > 0 && diffData.TargetResource.ConnectionLimit != diffData.SourceResource.ConnectionLimit {
				symbol := printUpdate("~")
				changeDetail := []string{printLabel("// diff connection limit before : ")}

				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) ConnectionLimit() int {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", strconv.Itoa(diffData.TargetResource.ConnectionLimit))))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

				changeDetail = append(changeDetail, printLabel("// diff connection limit after : "))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) ConnectionLimit() int {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", strconv.Itoa(diffData.SourceResource.ConnectionLimit))))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			}
		case objects.UpdateRoleName:
			scName := utils.ToSnakeCase(structName)
			if diffData.SourceResource.Name != "" {
				scName = diffData.SourceResource.Name
			}

			tgName := utils.ToSnakeCase(structName)
			if diffData.TargetResource.Name != "" {
				tgName = diffData.TargetResource.Name
			}

			symbol := printUpdate("~")
			changeDetail := []string{printLabel("// diff role name before : ")}
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) Name() string {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", tgName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changeDetail = append(changeDetail, printLabel("// diff role name after : "))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) Name() string {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", scName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changes = append(changes, strings.Join(changeDetail, "\n"))
		// case objects.UpdateRoleIsReplication:
		// case objects.UpdateRoleIsSuperUser:
		case objects.UpdateRoleInheritRole:
			var scInheritRole = "false"
			if diffData.SourceResource.InheritRole {
				scInheritRole = "true"
			}

			var tgInheritRole = "false"
			if diffData.TargetResource.InheritRole {
				tgInheritRole = "true"
			}

			symbol := printUpdate("~")
			changeDetail := []string{printLabel("// diff inherit role before : ")}

			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) InheritRole() bool {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", tgInheritRole)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changeDetail = append(changeDetail, printLabel("// diff inherit role after : "))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) InheritRole() bool  {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", scInheritRole)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changes = append(changes, strings.Join(changeDetail, "\n"))
		case objects.UpdateRoleCanCreateDb:
			var scCanCreateDB = "false"
			if diffData.SourceResource.CanCreateDB {
				scCanCreateDB = "true"
			}

			var tgCanCreateDB = "false"
			if diffData.TargetResource.CanCreateDB {
				tgCanCreateDB = "true"
			}

			symbol := printUpdate("~")
			changeDetail := []string{printLabel("// diff can create db before : ")}

			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) CanCreateDB() bool {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", tgCanCreateDB)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changeDetail = append(changeDetail, printLabel("// diff can create db after : "))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) CanCreateDB() bool  {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", scCanCreateDB)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changes = append(changes, strings.Join(changeDetail, "\n"))
		case objects.UpdateRoleCanCreateRole:
			var scCanCreateRole = "false"
			if diffData.SourceResource.CanCreateRole {
				scCanCreateRole = "true"
			}

			var tgCanCreateRole = "false"
			if diffData.TargetResource.CanCreateRole {
				tgCanCreateRole = "true"
			}

			symbol := printUpdate("~")
			changeDetail := []string{printLabel("// diff can create role before : ")}

			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) CanCreateRole() bool {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", tgCanCreateRole)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changeDetail = append(changeDetail, printLabel("// diff can create role after : "))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) CanCreateRole() bool  {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", scCanCreateRole)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changes = append(changes, strings.Join(changeDetail, "\n"))
		case objects.UpdateRoleCanLogin:
			var scCanLogin = "false"
			if diffData.SourceResource.CanLogin {
				scCanLogin = "true"
			}

			var tgCanLogin = "false"
			if diffData.TargetResource.CanLogin {
				tgCanLogin = "true"
			}

			symbol := printUpdate("~")
			changeDetail := []string{printLabel("// diff can login before : ")}

			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) CanLogin() bool {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", tgCanLogin)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changeDetail = append(changeDetail, printLabel("// diff can login after : "))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) CanLogin() bool  {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", scCanLogin)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changes = append(changes, strings.Join(changeDetail, "\n"))
		case objects.UpdateRoleCanBypassRls:
			var scCanBypassRls = "false"
			if diffData.SourceResource.CanBypassRLS {
				scCanBypassRls = "true"
			}

			var tgCanBypassRls = "false"
			if diffData.TargetResource.CanBypassRLS {
				tgCanBypassRls = "true"
			}

			symbol := printUpdate("~")
			changeDetail := []string{printLabel("// diff can login before : ")}

			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) CanBypassRls() bool {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", tgCanBypassRls)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changeDetail = append(changeDetail, printLabel("// diff can login after : "))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) CanBypassRls() bool  {", structName)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return %s", scCanBypassRls)))
			changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

			changes = append(changes, strings.Join(changeDetail, "\n"))
		// case objects.UpdateRoleConfig:
		case objects.UpdateRoleValidUntil:
			if diffData.TargetResource.ValidUntil == nil && diffData.SourceResource.ValidUntil != nil {
				symbol := printAdd("+")
				changeDetail := []string{}
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) ValidUntil() *objects.SupabaseTime {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  t, err := time.Parse(raiden.DefaultRoleValidUntilLayout, %s)", diffData.SourceResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout))))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  if err != nil {")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("    raiden.Error(err)")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("    return nil")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  }")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return objects.NewSupabaseTime(t)")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			}

			if diffData.TargetResource.ValidUntil != nil && diffData.SourceResource.ValidUntil == nil {
				symbol := printRemove("-")
				changeDetail := []string{}
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) ValidUntil() *objects.SupabaseTime {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  t, err := time.Parse(raiden.DefaultRoleValidUntilLayout, %s)", diffData.SourceResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout))))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  if err != nil {")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("    raiden.Error(err)")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("    return nil")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  }")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return objects.NewSupabaseTime(t)")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			}

			if diffData.TargetResource.ValidUntil != nil && diffData.SourceResource.ValidUntil != nil && diffData.TargetResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout) != diffData.SourceResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout) {
				symbol := printUpdate("~")
				changeDetail := []string{printLabel("// diff valid until before : ")}

				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) ValidUntil() *objects.SupabaseTime {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  t, err := time.Parse(raiden.DefaultRoleValidUntilLayout, %s)", diffData.TargetResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout))))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  if err != nil {")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("    raiden.Error(err)")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("    return nil")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  }")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return objects.NewSupabaseTime(t)")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

				changeDetail = append(changeDetail, printLabel("// diff valid until after : "))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("func (r *%s) ValidUntil() *objects.SupabaseTime {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  t, err := time.Parse(raiden.DefaultRoleValidUntilLayout, %s)", diffData.SourceResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout))))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  if err != nil {")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("    raiden.Error(err)")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("    return nil")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  }")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("  return objects.NewSupabaseTime(t)")))
				changeDetail = append(changeDetail, fmt.Sprintf("%s %s", symbol, print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			}
		}
	}

	printScope("*** Found diff in %s/%s.go ***\n", "/internal/roles", fileName)
	fmt.Println(strings.Join(changes, "\n\n"))
	printScope("\n*** End found diff ***\n")
}

func PrintDiffStorage(diffData CompareDiffResult[objects.Bucket, objects.UpdateBucketParam]) {
	if len(diffData.DiffItems.ChangeItems) == 0 {
		return
	}

	fileName := utils.ToSnakeCase(diffData.TargetResource.Name)
	structName := utils.SnakeCaseToPascalCase(fileName)

	print := color.New(color.FgWhite).SprintfFunc()
	printScope := color.New(color.FgHiBlue).PrintfFunc()
	printLabel := color.New(color.FgHiBlack).SprintfFunc()
	printAdd := color.New(color.FgHiGreen).SprintfFunc()
	printRemove := color.New(color.FgHiRed).SprintfFunc()
	printUpdate := color.New(color.FgHiYellow).SprintfFunc()

	changes := make([]string, 0)
	for _, v := range diffData.DiffItems.ChangeItems {
		switch v {
		case objects.UpdateBucketIsPublic:
			var scIsPublic = "false"
			if diffData.SourceResource.Public {
				scIsPublic = "true"
			}

			var tgIsPublic = "false"
			if diffData.TargetResource.Public {
				tgIsPublic = "true"
			}

			symbol := printUpdate("~")
			changeDetail := []string{}
			changeDetail = append(changeDetail, print("  func (r *%s) Public() bool {", structName))
			changeDetail = append(changeDetail, fmt.Sprintf("    %s %s %s %s", symbol, print("return %s", tgIsPublic), printLabel(">>>"), print(scIsPublic)))
			changeDetail = append(changeDetail, print("  }"))

			changes = append(changes, strings.Join(changeDetail, "\n"))
		case objects.UpdateBucketAllowedMimeTypes:
			if len(diffData.TargetResource.AllowedMimeTypes) == 0 && len(diffData.SourceResource.AllowedMimeTypes) > 0 {
				symbol := printAdd("+")
				changeDetail := []string{}
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("func (r *%s) AllowedMimeTypes() []string {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("  return %s", generator.GenerateArrayDeclaration(reflect.ValueOf(diffData.SourceResource.AllowedMimeTypes), false))))
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			} else if len(diffData.TargetResource.AllowedMimeTypes) > 0 && len(diffData.SourceResource.AllowedMimeTypes) == 0 {
				symbol := printRemove("-")
				changeDetail := []string{}
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("func (r *%s) AllowedMimeTypes() []string {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("  return %s", generator.GenerateArrayDeclaration(reflect.ValueOf(diffData.TargetResource.AllowedMimeTypes), false))))
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			} else {
				symbol := printUpdate("~")
				changeDetail := []string{}
				changeDetail = append(changeDetail, fmt.Sprintf("  %s", print("func (r *%s) AllowedMimeTypes() []string {", structName)))
				changeDetail = append(changeDetail,
					fmt.Sprintf("    %s %s", symbol,
						print(
							"return %s >>> %s",
							generator.GenerateArrayDeclaration(reflect.ValueOf(diffData.TargetResource.AllowedMimeTypes), false),
							generator.GenerateArrayDeclaration(reflect.ValueOf(diffData.SourceResource.AllowedMimeTypes), false),
						),
					),
				)
				changeDetail = append(changeDetail, fmt.Sprintf("  %s", print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			}
		case objects.UpdateBucketFileSizeLimit:
			if diffData.TargetResource.FileSizeLimit == nil && diffData.SourceResource.FileSizeLimit != nil && *diffData.SourceResource.FileSizeLimit > 0 {
				symbol := printAdd("+")
				changeDetail := []string{}
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("func (r *%s) FileSizeLimit() int {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("  return %s", strconv.Itoa(*diffData.SourceResource.FileSizeLimit))))
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			}

			if diffData.TargetResource.FileSizeLimit != nil && *diffData.TargetResource.FileSizeLimit > 0 && diffData.SourceResource.FileSizeLimit == nil {
				symbol := printRemove("-")
				changeDetail := []string{}
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("func (r *%s) FileSizeLimit() int {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("  return %s", strconv.Itoa(*diffData.TargetResource.FileSizeLimit))))
				changeDetail = append(changeDetail, fmt.Sprintf("  %s %s", symbol, print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			}

			if diffData.TargetResource.FileSizeLimit != nil && diffData.SourceResource.FileSizeLimit != nil && *diffData.TargetResource.FileSizeLimit != *diffData.SourceResource.FileSizeLimit {
				symbol := printRemove("-")
				changeDetail := []string{}
				changeDetail = append(changeDetail, fmt.Sprintf("  %s", print("func (r *%s) FileSizeLimit() int {", structName)))
				changeDetail = append(changeDetail, fmt.Sprintf("    %s %s", symbol, print("return %s >>> %s", strconv.Itoa(*diffData.TargetResource.FileSizeLimit), strconv.Itoa(*diffData.SourceResource.FileSizeLimit))))
				changeDetail = append(changeDetail, fmt.Sprintf("  %s", print("}")))

				changes = append(changes, strings.Join(changeDetail, "\n"))
			}
		}
	}

	printScope("*** Found diff in %s/%s.go ***\n", "/internal/storages", fileName)
	fmt.Println(strings.Join(changes, "\n\n"))
	printScope("\n*** End found diff ***\n")
}
