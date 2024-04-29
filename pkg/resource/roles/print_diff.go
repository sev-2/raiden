package roles

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- print diff section -----
type DiffType string

const (
	DiffTypeCreate DiffType = "create"
	DiffTypeUpdate DiffType = "update"
	DiffTypeDelete DiffType = "delete"
)

func PrintDiffResult(diffResult []CompareDiffResult) error {
	if len(diffResult) == 0 {
		return nil
	}

	isConflict := false
	for i := range diffResult {
		d := diffResult[i]
		if d.IsConflict {
			PrintDiff(d)
			if !isConflict {
				isConflict = true
			}
		}
	}

	if isConflict {
		return errors.New("canceled import process, you have conflict in role. please fix it first")
	}

	return nil
}

func PrintDiff(diffData CompareDiffResult) {
	if len(diffData.DiffItems.ChangeItems) == 0 {
		return
	}
	fileName := utils.ToSnakeCase(diffData.TargetResource.Name)
	structName := utils.SnakeCaseToPascalCase(fileName)
	printScope := color.New(color.FgHiBlack).PrintfFunc()

	changes := make([]string, 0)
	for _, v := range diffData.DiffItems.ChangeItems {
		switch v {
		case objects.UpdateConnectionLimit:
			var diffType DiffType
			var value, changedValue string

			if diffData.TargetResource.ConnectionLimit == 0 && diffData.SourceResource.ConnectionLimit > 0 {
				diffType = DiffTypeCreate
				value = strconv.Itoa(diffData.SourceResource.ConnectionLimit)
			}

			if diffData.TargetResource.ConnectionLimit > 0 && diffData.SourceResource.ConnectionLimit == 0 {
				diffType = DiffTypeDelete
				value = strconv.Itoa(diffData.SourceResource.ConnectionLimit)
			}

			if diffData.TargetResource.ConnectionLimit > 0 && diffData.SourceResource.ConnectionLimit > 0 && diffData.TargetResource.ConnectionLimit != diffData.SourceResource.ConnectionLimit {
				diffType = DiffTypeUpdate
				value = strconv.Itoa(diffData.TargetResource.ConnectionLimit)
				changedValue = strconv.Itoa(diffData.SourceResource.ConnectionLimit)
			}

			diffStr, err := GenerateDiffMessage(fileName, diffType, v, value, changedValue)
			if err != nil {
				Logger.Error("print diff roles error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		case objects.UpdateRoleName:
			scName := utils.ToSnakeCase(structName)
			if diffData.SourceResource.Name != "" {
				scName = diffData.SourceResource.Name
			}

			tgName := utils.ToSnakeCase(structName)
			if diffData.TargetResource.Name != "" {
				tgName = diffData.TargetResource.Name
			}

			diffStr, err := GenerateDiffMessage(fileName, DiffTypeUpdate, v, tgName, scName)
			if err != nil {
				Logger.Error("print diff roles error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
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

			diffStr, err := GenerateDiffMessage(fileName, DiffTypeUpdate, v, tgInheritRole, scInheritRole)
			if err != nil {
				Logger.Error("print diff roles error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		case objects.UpdateRoleCanCreateDb:
			var scCanCreateDB = "false"
			if diffData.SourceResource.CanCreateDB {
				scCanCreateDB = "true"
			}

			var tgCanCreateDB = "false"
			if diffData.TargetResource.CanCreateDB {
				tgCanCreateDB = "true"
			}

			diffStr, err := GenerateDiffMessage(fileName, DiffTypeUpdate, v, tgCanCreateDB, scCanCreateDB)
			if err != nil {
				Logger.Error("print diff roles error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		case objects.UpdateRoleCanCreateRole:
			var scCanCreateRole = "false"
			if diffData.SourceResource.CanCreateRole {
				scCanCreateRole = "true"
			}

			var tgCanCreateRole = "false"
			if diffData.TargetResource.CanCreateRole {
				tgCanCreateRole = "true"
			}

			diffStr, err := GenerateDiffMessage(fileName, DiffTypeUpdate, v, tgCanCreateRole, scCanCreateRole)
			if err != nil {
				Logger.Error("print diff roles error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		case objects.UpdateRoleCanLogin:
			var scCanLogin = "false"
			if diffData.SourceResource.CanLogin {
				scCanLogin = "true"
			}

			var tgCanLogin = "false"
			if diffData.TargetResource.CanLogin {
				tgCanLogin = "true"
			}

			diffStr, err := GenerateDiffMessage(fileName, DiffTypeUpdate, v, tgCanLogin, scCanLogin)
			if err != nil {
				Logger.Error("print diff roles error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		case objects.UpdateRoleCanBypassRls:
			var scCanBypassRls = "false"
			if diffData.SourceResource.CanBypassRLS {
				scCanBypassRls = "true"
			}

			var tgCanBypassRls = "false"
			if diffData.TargetResource.CanBypassRLS {
				tgCanBypassRls = "true"
			}

			diffStr, err := GenerateDiffMessage(fileName, DiffTypeUpdate, v, tgCanBypassRls, scCanBypassRls)
			if err != nil {
				Logger.Error("print diff roles error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		// case objects.UpdateRoleConfig:
		case objects.UpdateRoleValidUntil:
			var diffType DiffType
			var value, changedValue string

			if diffData.TargetResource.ValidUntil == nil && diffData.SourceResource.ValidUntil != nil {
				diffType = DiffTypeCreate
				value = diffData.SourceResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)
			}

			if diffData.TargetResource.ValidUntil != nil && diffData.SourceResource.ValidUntil == nil {
				diffType = DiffTypeDelete
				value = diffData.SourceResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)
			}

			if diffData.TargetResource.ValidUntil != nil && diffData.SourceResource.ValidUntil != nil && diffData.TargetResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout) != diffData.SourceResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout) {
				diffType = DiffTypeUpdate
				value = diffData.TargetResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)
				changedValue = diffData.SourceResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)
			}

			diffStr, err := GenerateDiffMessage(fileName, diffType, v, value, changedValue)
			if err != nil {
				Logger.Error("print diff roles error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		}
	}

	printScope("*** Found diff in %s/%s.go ***\n", "/internal/roles", fileName)
	fmt.Println(strings.Join(changes, ""))
	printScope("*** End found diff ***\n")
}

// ----- generate message section ------
const DiffTemplate = ` 
 {{- if or (eq .Type "create") (eq .Type "delete")}}
 {{ .Symbol }} func (r *{{ .Name  | ToGoIdentifier }}) %s {
 %s
 {{ .Symbol }} }
 {{- end}}
 {{- if eq .Type "update"}}
 func (r *{{ .Name | ToGoIdentifier }}) %s {
   %s
 }
{{- end}}
  `

const FuncBodyTemplate = "{{ .Symbol }} return {{ .Value }}"
const FuncBodyUpdateTemplate = "{{ .Symbol }} return {{ .Value }}  >>> {{ .ChangeValue }}"

func buildDiffTemplate(funcDecl string, bodyTemplate string, updateBodyTemplate string) string {
	if bodyTemplate == "" {
		bodyTemplate = FuncBodyTemplate
	}

	if updateBodyTemplate == "" {
		updateBodyTemplate = FuncBodyUpdateTemplate
	}

	return fmt.Sprintf(DiffTemplate, funcDecl, bodyTemplate, funcDecl, updateBodyTemplate)
}

func getDiffSymbol(diffType DiffType) string {
	printAdd := color.New(color.FgHiGreen).SprintfFunc()
	printRemove := color.New(color.FgHiRed).SprintfFunc()
	printUpdate := color.New(color.FgHiYellow).SprintfFunc()

	var symbol string
	switch diffType {
	case DiffTypeCreate:
		symbol = printAdd("+")
	case DiffTypeUpdate:
		symbol = printUpdate("~")
	case DiffTypeDelete:
		symbol = printRemove("-")
	}
	return symbol
}

func GenerateDiffMessage(name string, diffType DiffType, updateType objects.UpdateRoleType, value string, changeValue string) (string, error) {
	param := map[string]any{
		"Name":        name,
		"Type":        diffType,
		"Value":       value,
		"ChangeValue": changeValue,
		"Symbol":      getDiffSymbol(diffType),
	}

	tmplStr := ""
	switch updateType {
	case objects.UpdateConnectionLimit:
		tmplStr = buildDiffTemplate("ConnectionLimit() int", "", "")
	case objects.UpdateRoleName:
		tmplStr = buildDiffTemplate("Name() string", "", "")
	case objects.UpdateRoleIsReplication:
		tmplStr = buildDiffTemplate("IsReplicationRole() bool", "", "")
	case objects.UpdateRoleIsSuperUser:
		tmplStr = buildDiffTemplate("IsSuperuser() bool", "", "")
	case objects.UpdateRoleInheritRole:
		tmplStr = buildDiffTemplate("InheritRole()", "", "")
	case objects.UpdateRoleCanCreateDb:
		tmplStr = buildDiffTemplate("CanCreateDB() bool", "", "")
	case objects.UpdateRoleCanCreateRole:
		tmplStr = buildDiffTemplate("CanCreateRole() bool", "", "")
	case objects.UpdateRoleCanLogin:
		tmplStr = buildDiffTemplate("CanLogin() bool", "", "")
	case objects.UpdateRoleCanBypassRls:
		tmplStr = buildDiffTemplate("CanBypassRls() bool", "", "")
	case objects.UpdateRoleValidUntil:
		tmplStr = buildDiffTemplate(
			"ValidUntil() *objects.SupabaseTime",
			`
  {{ .Symbol }} t, err := time.Parse(raiden.DefaultRoleValidUntilLayout, "{{ .Value }}")
  if err != nil {
  	raiden.Error(err.Error())
  	return nil
  }
  return objects.NewSupabaseTime(t)`,
			`
  {{ .Symbol }} t, err := time.Parse(raiden.DefaultRoleValidUntilLayout, "{{ .Value }}" >>> "{{ .ChangeValue }}")
  if err != nil {
  	raiden.Error(err.Error())
  	return nil
  }
  return objects.NewSupabaseTime(t)`,
		)
	default:
		return "", errors.New("unsupported update type")
	}

	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
	}

	tmplInstance := template.New("generate diff")
	for _, tm := range funcMaps {
		tmplInstance.Funcs(tm)
	}

	tmpl, err := tmplInstance.Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("error parsing : %v", err)
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, param); err != nil {
		return "", err
	}

	return buff.String(), nil
}
