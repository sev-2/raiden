package rpc

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- print diff section -----
func PrintDiffResult(diffResult []CompareDiffResult) error {
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
		return errors.New("canceled import process, you have conflict rpc function. please fix it first")
	}

	return nil
}

func PrintDiff(diffData CompareDiffResult) {
	fileName := utils.ToSnakeCase(diffData.TargetResource.Name)
	printScope := color.New(color.FgHiBlack).PrintfFunc()

	diffMessage, err := GenerateDiffMessage(diffData.Name, diffData.TargetResource.CompleteStatement, diffData.SourceResource.CompleteStatement)
	if err != nil {
		Logger.Error("print diff rpc error", "msg", err.Error())
		return
	}

	printScope("*** Found diff in %s/%s.go ***\n", "/internal/rpc", fileName)
	fmt.Println(diffMessage)
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
			updateData = append(updateData, fmt.Sprintf("- %s", name))
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

// ----- generate message section ------
const DiffTemplate = ` 
{{.Head}}
{{.Body}}
{{.End}}
  `

func GenerateDiffMessage(name string, value, changeValue string) (string, error) {
	printUpdate := color.New(color.FgHiYellow).SprintfFunc()
	printIndent := color.New(color.FgHiBlack).SprintfFunc()
	symbol := printUpdate("~")
	fromIndent := printIndent("from:")
	toIndent := printIndent("to:")

	sHead, sBody, _ := splitFunction(strings.ToLower(changeValue))
	tHead, tBody, tEnd := splitFunction(value)

	var head, body, end string
	if sHead != tHead {
		head = fmt.Sprintf("%s %s %s \n%s %s   %s", symbol, fromIndent, tHead, symbol, toIndent, sHead)
	} else {
		head = tHead
	}

	if sBody != tBody {
		sBodyArr := strings.Split(sBody, ";")
		tBodyArr := strings.Split(tBody, ";")

		sBody = strings.Join(sBodyArr, ";\n")
		tBody = strings.Join(tBodyArr, ";\n")
		body = fmt.Sprintf("\t%s %s %s\t%s %s   %s", symbol, fromIndent, tBody, symbol, toIndent, sBody)
	} else {
		body = tBody
	}

	end = tEnd

	param := map[string]any{
		"Head": head,
		"Body": body,
		"End":  end,
	}

	tmplInstance := template.New("generate diff")

	tmpl, err := tmplInstance.Parse(DiffTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing : %v", err)
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, param); err != nil {
		return "", err
	}

	return buff.String(), nil
}

func splitFunction(query string) (head, body, end string) {
	if strings.Contains(query, "$function$begin") {
		query = strings.ReplaceAll(query, "$function$begin", "$function$ begin")
	}

	if strings.Contains(query, "end$function$") {
		query = strings.ReplaceAll(query, "end$function$", "end $function$")
	}

	splitSql := strings.Split(query, "$function$ begin")
	end = "end $function$"
	if len(splitSql) == 2 {
		head = splitSql[0] + " $function$ begin"
		head = strings.Join(strings.Fields(head), " ")
		body = strings.Replace(splitSql[1], end, "", 1)
		body = strings.Join(strings.Fields(body), " ")
	}

	return head, body, end
}

// ----- diff change -----
const DiffChangeTemplate = `
  {{- if gt (len .NewData) 0}}
  New Rpc
  {{- range .NewData}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .UpdateData) 0}}
  Update Rpc
  {{- range .UpdateData}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .DeleteData) 0}}
  Delete Rpc
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

	tmplInstance := template.New("generate diff change rpc")
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
