package policies

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

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
