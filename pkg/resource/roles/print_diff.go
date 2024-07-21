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
	"github.com/sev-2/raiden/pkg/resource/migrator"
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
				value = diffData.TargetResource.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)
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

func GetDiffChangeMessage(items []MigrateItem) string {
	newData := []string{}
	deleteData := []string{}
	updateData := []string{}

	for i := range items {
		item := items[i]

		var name string
		if item.NewData.Name != "" {
			name = item.NewData.Name
		} else if item.OldData.Name != "" {
			name = item.OldData.Name
		}

		switch item.Type {
		case migrator.MigrateTypeCreate:
			newData = append(newData, fmt.Sprintf("- %s", name))
		case migrator.MigrateTypeUpdate:
			diffMessage, err := GenerateDiffChangeUpdateMessage(name, item)
			if err != nil {
				Logger.Error("print change role error", "msg", err.Error())
				continue
			}
			updateData = append(updateData, diffMessage)
		case migrator.MigrateTypeDelete:
			deleteData = append(deleteData, fmt.Sprintf("- %s", name))
		}
	}

	changeMsg, err := GenerateDiffChangeMessage(newData, updateData, deleteData)
	if err != nil {
		Logger.Error("print change role error", "msg", err.Error())
		return ""
	}
	return changeMsg
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

// ----- diff change -----
const DiffChangeTemplate = `
  {{- if gt (len .NewData) 0}}
  New Role
  {{- range .NewData}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .UpdateData) 0}}
  Update Role
  {{- range .UpdateData}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .DeleteData) 0}}
  Delete Role
  {{- range .DeleteData}}
  {{.}}
  {{- end }}
  {{- end -}}
  `

func GenerateDiffChangeMessage(newData []string, updateData []string, deleteData []string) (string, error) {
	param := map[string]any{
		"NewData":    newData,
		"UpdateData": updateData,
		"DeleteData": deleteData,
	}

	tmplInstance := template.New("generate diff change role")
	tmpl, err := tmplInstance.Parse(DiffChangeTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing : %v", err)
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, param); err != nil {
		return "", err
	}

	return buff.String(), nil
}

const DiffChangeUpdateTemplate = `  - Update Role {{ .Name }}
  {{- if gt (len .ChangeItems) 0}}
      Change Configuration
      {{- range .ChangeItems}}
      {{.}}
      {{- end }}
  {{- end -}}
  `

func GenerateDiffChangeUpdateMessage(name string, item MigrateItem) (string, error) {
	diffItems := item.MigrationItems

	var changeMsgArr []string
	for i := range diffItems.ChangeItems {
		c := diffItems.ChangeItems[i]
		switch c {
		case objects.UpdateConnectionLimit:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %v >>> %v", "connection limit", item.OldData.ConnectionLimit, item.NewData.ConnectionLimit))
		case objects.UpdateRoleName:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %v >>> %v", "name", item.OldData.Name, item.NewData.Name))
		case objects.UpdateRoleIsReplication:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %t >>> %t", "is replicate role", item.OldData.IsReplicationRole, item.NewData.IsReplicationRole))
		case objects.UpdateRoleIsSuperUser:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %t >>> %t", "is super user", item.OldData.IsSuperuser, item.NewData.IsSuperuser))
		case objects.UpdateRoleInheritRole:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %t >>> %t", "inherit role", item.OldData.InheritRole, item.NewData.InheritRole))
		case objects.UpdateRoleCanCreateDb:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %t >>> %t", "can create db", item.OldData.CanCreateDB, item.NewData.CanCreateDB))
		case objects.UpdateRoleCanCreateRole:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %t >>> %t", "can create role", item.OldData.CanCreateRole, item.NewData.CanCreateRole))
		case objects.UpdateRoleCanLogin:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %t >>> %t", "can login", item.OldData.CanLogin, item.NewData.CanLogin))
		case objects.UpdateRoleCanBypassRls:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %t >>> %t", "can bypass url", item.OldData.CanBypassRLS, item.NewData.CanBypassRLS))
		case objects.UpdateRoleConfig:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %+v >>> %+v", "can bypass url", item.OldData.Config, item.NewData.Config))
		case objects.UpdateRoleValidUntil:
			var oldValue, newValue = "nil", "nil"
			if item.OldData.ValidUntil != nil {
				oldValue = item.OldData.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)
			}

			if item.NewData.ValidUntil != nil {
				newValue = item.NewData.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)
			}
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "valid until", oldValue, newValue))
		}
	}

	param := map[string]any{
		"Name":        name,
		"ChangeItems": changeMsgArr,
	}

	tmplInstance := template.New("generate diff change update")
	tmpl, err := tmplInstance.Parse(DiffChangeUpdateTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing : %v", err)
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, param); err != nil {
		return "", err
	}

	return buff.String(), nil
}
