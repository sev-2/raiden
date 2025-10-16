package policies

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- print diff section -----
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
		return errors.New("canceled import process, you have conflict in policy. please fix it first")
	}

	return nil
}

func PrintDiff(diffData CompareDiffResult) {
	if len(diffData.DiffItems.ChangeItems) == 0 {
		return
	}
	fileName := utils.ToSnakeCase(diffData.TargetResource.Name)
	printScope := color.New(color.FgHiBlack).PrintfFunc()

	changes := make([]string, 0)
	policyName := diffData.SourceResource.Name
	if policyName == "" {
		policyName = diffData.TargetResource.Name
	}

	for _, v := range diffData.DiffItems.ChangeItems {
		switch v {
		case objects.UpdatePolicyName:
			changeMsg := fmt.Sprintf("- name: %s >>> %s", diffData.TargetResource.Name, diffData.SourceResource.Name)
			changes = append(changes, changeMsg)
		case objects.UpdatePolicyDefinition:
			oldDef := utils.ConvertAllToString(diffData.TargetResource.Definition)
			if len(oldDef) == 0 {
				oldDef = "unset"
			}

			newDef := utils.ConvertAllToString(diffData.SourceResource.Definition)
			if len(newDef) == 0 {
				newDef = "unset"
			}

			if oldDef == newDef {
				continue
			}

			changeMsg := fmt.Sprintf("- definition: %s >>> %s", oldDef, newDef)
			changes = append(changes, changeMsg)
		case objects.UpdatePolicyCheck:
			oldCheck := utils.ConvertAllToString(diffData.TargetResource.Check)
			if len(oldCheck) == 0 {
				oldCheck = "unset"
			}

			newCheck := utils.ConvertAllToString(diffData.SourceResource.Check)
			if len(newCheck) == 0 {
				newCheck = "unset"
			}

			if oldCheck == newCheck {
				continue
			}

			changeMsg := fmt.Sprintf("- check: %s >>> %s", oldCheck, newCheck)
			changes = append(changes, changeMsg)
		case objects.UpdatePolicyRoles:
			// Convert roles to string representation for comparison
			oldRoles := strings.Join(diffData.TargetResource.Roles, ", ")
			newRoles := strings.Join(diffData.SourceResource.Roles, ", ")

			if oldRoles == newRoles {
				continue
			}

			changeMsg := fmt.Sprintf("- roles: [%s] >>> [%s]", oldRoles, newRoles)
			changes = append(changes, changeMsg)
		}
	}

	if len(changes) == 0 {
		return
	}

	printScope("*** Found diff in %s/%s.go ***\n", "/internal/policies", fileName)
	fmt.Println(strings.Join(changes, "\n"))
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
				Logger.Error("print change policy error", "msg", err.Error())
				continue
			}
			updateData = append(updateData, diffMessage)
		case migrator.MigrateTypeDelete:
			deleteData = append(deleteData, fmt.Sprintf("- %s", name))
		}
	}

	changeMsg, err := GenerateDiffChangeMessage(newData, updateData, deleteData)
	if err != nil {
		Logger.Error("print change policy error", "msg", err.Error())
		return ""
	}
	return changeMsg
}

// ----- diff change -----
const DiffChangeTemplate = `
  {{- if gt (len .NewData) 0}}
  New Policy
  {{- range .NewData}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .UpdateData) 0}}
  Update Policy
  {{- range .UpdateData}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .DeleteData) 0}}
  Delete Policy
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

	tmplInstance := template.New("generate diff change policy")
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

const DiffChangeUpdateTemplate = `  - Update Policy {{ .Name }}
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
		case objects.UpdatePolicyName:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "name", item.OldData.Name, item.NewData.Name))
		case objects.UpdatePolicyDefinition:
			oldDef := utils.ConvertAllToString(item.OldData.Definition)
			if len(oldDef) == 0 {
				oldDef = "unset"
			}

			newDef := utils.ConvertAllToString(item.NewData.Definition)
			if len(newDef) == 0 {
				newDef = "unset"
			}

			if oldDef == newDef {
				continue
			}

			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "definition", oldDef, newDef))
		case objects.UpdatePolicyCheck:
			oldCheck := utils.ConvertAllToString(item.OldData.Check)
			if len(oldCheck) == 0 {
				oldCheck = "unset"
			}

			newCheck := utils.ConvertAllToString(item.NewData.Check)
			if len(newCheck) == 0 {
				newCheck = "unset"
			}

			if oldCheck == newCheck {
				continue
			}

			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "check", oldCheck, newCheck))
		case objects.UpdatePolicyRoles:
			oldRoles := strings.Join(item.OldData.Roles, ",")
			if len(oldRoles) == 0 {
				oldRoles = "unset"
			}

			newRoles := strings.Join(item.NewData.Roles, ",")
			if len(newRoles) == 0 {
				newRoles = "unset"
			}

			if oldRoles != newRoles {
				changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "roles", oldRoles, newRoles))
			}
		case objects.UpdatePolicySchema:
			if !strings.EqualFold(item.OldData.Schema, item.NewData.Schema) {
				changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "schema", item.OldData.Schema, item.NewData.Schema))
			}
		case objects.UpdatePolicyTable:
			if !strings.EqualFold(item.OldData.Table, item.NewData.Table) {
				changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "table", item.OldData.Table, item.NewData.Table))
			}
		case objects.UpdatePolicyAction:
			if !strings.EqualFold(item.OldData.Action, item.NewData.Action) {
				changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "action", item.OldData.Action, item.NewData.Action))
			}
		case objects.UpdatePolicyCommand:
			if !strings.EqualFold(string(item.OldData.Command), string(item.NewData.Command)) {
				changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "command", item.OldData.Command, item.NewData.Command))
			}
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
