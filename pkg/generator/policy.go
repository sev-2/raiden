package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type Rls struct {
	CanWrite []string
	CanRead  []string
}

func BuildRlsTag(rlsList objects.Policies, name string, rlsType supabase.RlsType) string {
	var rls Rls

	var readUsingTag, writeCheckTag, writeUsingTag string
	for _, v := range rlsList {
		switch v.Command {
		case objects.PolicyCommandSelect:
			if v.Name == supabase.GetPolicyName(objects.PolicyCommandSelect, strings.ToLower(string(rlsType)), name) {
				rls.CanRead = append(rls.CanRead, v.Roles...)
				if v.Definition != "" {
					readUsingTag = v.Definition
				}
			}
		case objects.PolicyCommandInsert, objects.PolicyCommandUpdate, objects.PolicyCommandDelete:
			if v.Name == supabase.GetPolicyName(objects.PolicyCommandInsert, strings.ToLower(string(rlsType)), name) {
				if len(rls.CanWrite) == 0 {
					rls.CanWrite = append(rls.CanWrite, v.Roles...)
				}

				if len(writeCheckTag) == 0 && v.Check != nil {
					writeCheckTag = *v.Check
				}
			}

			if v.Name == supabase.GetPolicyName(objects.PolicyCommandUpdate, strings.ToLower(string(rlsType)), name) && len(rls.CanWrite) == 0 {
				if len(rls.CanWrite) == 0 {
					rls.CanWrite = append(rls.CanWrite, v.Roles...)
				}

				if len(writeCheckTag) == 0 && v.Check != nil {
					writeCheckTag = *v.Check
				}

				if len(writeUsingTag) == 0 && v.Definition != "" {
					writeUsingTag = v.Definition
				}
			}

			if v.Name == supabase.GetPolicyName(objects.PolicyCommandDelete, strings.ToLower(string(rlsType)), name) && len(rls.CanWrite) == 0 {
				if len(rls.CanWrite) == 0 {
					rls.CanWrite = append(rls.CanWrite, v.Roles...)
				}

				if len(writeUsingTag) == 0 && v.Definition != "" {
					writeUsingTag = v.Definition
				}
			}
		}
	}

	rlsTag := fmt.Sprintf("read:%q write:%q", strings.Join(rls.CanRead, ","), strings.Join(rls.CanWrite, ","))
	if len(readUsingTag) > 0 {
		cleanTag := strings.TrimLeft(strings.TrimRight(readUsingTag, ")"), "(")
		if rlsType == supabase.RlsTypeStorage {
			cleanTag = cleanupRlsTagStorage(name, cleanTag)
		}

		if cleanTag != "" {
			rlsTag = fmt.Sprintf("%s readUsing:%q", rlsTag, cleanTag)
		}
	}

	if len(writeCheckTag) > 0 {
		cleanTag := strings.TrimLeft(strings.TrimRight(writeCheckTag, ")"), "(")
		if rlsType == supabase.RlsTypeStorage {
			cleanTag = cleanupRlsTagStorage(name, cleanTag)
		}

		if cleanTag != "" {
			rlsTag = fmt.Sprintf("%s writeCheck:%q", rlsTag, cleanTag)
		}
	}

	if len(writeUsingTag) > 0 {
		cleanTag := strings.TrimLeft(strings.TrimRight(writeUsingTag, ")"), "(")
		if rlsType == supabase.RlsTypeStorage {
			cleanTag = cleanupRlsTagStorage(name, cleanTag)
		}

		if cleanTag != "" {
			rlsTag = fmt.Sprintf("%s writeUsing:%q", rlsTag, cleanTag)
		}
	}

	return rlsTag
}

func cleanupRlsTagStorage(name, tag string) string {
	// clean storage identifier
	cleanTag := strings.Replace(tag, fmt.Sprintf("bucket_id = '%s'", name), "", 1)
	cleanTag = strings.Replace(cleanTag, "AND", "", 1)
	cleanTag = strings.Replace(cleanTag, "OR", "", 1)
	cleanTag = strings.TrimLeftFunc(cleanTag, unicode.IsSpace)
	return cleanTag
}

var PolicyLogger hclog.Logger = logger.HcLog().Named("generator.policy")

// ----- Define type, variable and constant -----
type GeneratePolicyData struct {
	Imports    []string
	Package    string
	Name       string
	Model      string
	Storage    string
	Roles      string
	Command    string
	Mode       string
	Check      string
	Definition string
}

const (
	PolicyDir      = "internal/policies"
	PolicyTemplate = `package {{ .Package }}

{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)

{{- end }}

type {{ .Name | ToGoIdentifier }} struct {
	raiden.PolicyBase
	{{- if ne .Model "" }}
	model models.{{.Model | ToGoIdentifier}}{{- end}}
	{{- if ne .Storage "" }}
	storage storages.{{.Storage | ToGoIdentifier}}{{- end}}
}

func (p *{{ .Name | ToGoIdentifier }}) Name() string {
	return "{{ .Name }}"
}
{{- if ne .Model "" }}

func (p *{{ .Name | ToGoIdentifier }}) Model() any {
	return &p.model;
}
{{- end}}
{{- if ne .Storage "" }}

func (p *{{ .Name | ToGoIdentifier }}) Storage() any {
	return &p.storage;
}
{{- end}}

func (p *{{ .Name | ToGoIdentifier }}) Roles() []raiden.Role {
	return {{ .Roles }}
}

func (p *{{ .Name | ToGoIdentifier }}) Command() raiden.PolicyCommand {
	return raiden.PolicyCommand{{ .Command }}
}

func (p *{{ .Name | ToGoIdentifier }}) Mode() raiden.PolicyMode {
	return {{ .Mode }}
}

func (p *{{ .Name | ToGoIdentifier }}) Definition() b.Clause {
	return {{ .Definition }}
}

func (p *{{ .Name | ToGoIdentifier }}) Check() b.Clause {
	return {{ .Check }}
}

`
)

func GeneratePolicies(
	basePath string, projectName string, policies []objects.Policy,
	tableMap map[string]objects.Table, storageMap map[string]string,
	roleMap map[string]string, nativeRoleMap map[string]raiden.Role,
	generateFn GenerateFn,
) (err error) {
	folderPath := filepath.Join(basePath, PolicyDir)
	PolicyLogger.Trace("create policies folder if not exist", folderPath)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for _, v := range policies {
		if err := GeneratePolicy(folderPath, projectName, v, tableMap, storageMap, roleMap, nativeRoleMap, generateFn); err != nil {
			return err
		}
	}

	return nil
}

func GeneratePolicy(
	folderPath string, projectName string, policy objects.Policy,
	tableMap map[string]objects.Table, storageMap map[string]string,
	roleMap map[string]string, nativeRoleMap map[string]raiden.Role,
	generateFn GenerateFn,
) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
	}

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", utils.ToSnakeCase(policy.Name), "go"))

	// set imports path
	var imports []string

	raidenPath := fmt.Sprintf("%q", "github.com/sev-2/raiden")
	imports = append(imports, raidenPath)

	builderPath := fmt.Sprintf("b %q", "github.com/sev-2/raiden/pkg/builder")
	imports = append(imports, builderPath)

	roleImportPath := fmt.Sprintf("%s/%s", utils.ToGoModuleName(projectName), RoleDir)
	imports = append(imports, fmt.Sprintf("%q", roleImportPath))

	// convert policy command to appropriate format (PascalCase)
	command := utils.SnakeCaseToPascalCase(strings.ToLower(string(policy.Command)))

	var model, storage, roles, mode string
	checkClause := "b.Clause(\"\")"
	definitionClause := "b.Clause(\"\")"
	if policy.Schema == supabase.DefaultStorageSchema {
		storage = policy.Table
		if _, exist := storageMap[storage]; !exist {
			return fmt.Errorf("generator : bucket %s is not found", policy.Table)
		}
		storagesImportPath := fmt.Sprintf("%s/%s", utils.ToGoModuleName(projectName), StorageDir)
		imports = append(imports, fmt.Sprintf("%q", storagesImportPath))
	} else {
		model = policy.Table
		if model != "" {
			if _, exist := tableMap[model]; !exist {
				return fmt.Errorf("generator : table %s is not found", policy.Table)
			}
		}
		modelsImportPath := fmt.Sprintf("%s/%s", utils.ToGoModuleName(projectName), ModelDir)
		imports = append(imports, fmt.Sprintf("%q", modelsImportPath))
	}

	if len(policy.Roles) > 0 {
		roleArr := []string{}
		isImportNativeRole := false

		for _, r := range policy.Roles {
			isFound := false

			if _, exist := roleMap[r]; exist {
				roleArr = append(roleArr, fmt.Sprintf("&roles.%s{},", utils.SnakeCaseToPascalCase(r)))
				isFound = true
			}

			if _, exist := nativeRoleMap[r]; exist {
				roleArr = append(roleArr, fmt.Sprintf("&native_role.%s{},", utils.SnakeCaseToPascalCase(r)))
				isFound, isImportNativeRole = true, true
			}

			if !isFound {
				return fmt.Errorf("generator : role %s is not found", r)
			}
		}

		if isImportNativeRole {
			nativeRolePath := fmt.Sprintf("native_role %q", "github.com/sev-2/raiden/pkg/postgres/roles")
			imports = append(imports, nativeRolePath)
		}
		roles = fmt.Sprintf("[]raiden.Role{\n        %s\n    }", strings.Join(roleArr, "\n        "))
	} else {
		roles = "[]raiden.Role{}"
	}

	if strings.EqualFold(policy.Action, raiden.ModePermissive.ActionString()) {
		mode = "raiden.ModePermissive"
	} else {
		mode = "raiden.ModeRestrictive"
	}

	qualifier := builder.ClauseQualifier{Schema: policy.Schema, Table: policy.Table}
	receiverName, columnSet := resolvePolicyColumns(model, storage, tableMap)
	if policy.Check != nil && strings.TrimSpace(*policy.Check) != "" {
		checkClause = generateClauseCode(*policy.Check, qualifier, receiverName, columnSet)
	}

	if strings.TrimSpace(policy.Definition) != "" {
		definitionClause = generateClauseCode(policy.Definition, qualifier, receiverName, columnSet)
	}

	// execute the template and write to the file
	data := GeneratePolicyData{
		Package:    "policies",
		Imports:    imports,
		Name:       policy.Name,
		Model:      model,
		Storage:    storage,
		Roles:      roles,
		Mode:       mode,
		Command:    command,
		Check:      checkClause,
		Definition: definitionClause,
	}

	// set input
	input := GenerateInput{
		BindData:     data,
		Template:     PolicyTemplate,
		TemplateName: "policyTemplate",
		OutputPath:   filePath,
		FuncMap:      funcMaps,
	}

	PolicyLogger.Debug("generate policy", "path", input.OutputPath)
	return generateFn(input, nil)
}

func resolvePolicyColumns(modelName, storageName string, tableMap map[string]objects.Table) (string, []objects.Column) {
	if modelName != "" {
		if table, ok := tableMap[modelName]; ok {
			return "p.model", table.Columns
		}
	}
	if storageName != "" {
		if table, ok := tableMap[storageName]; ok {
			return "p.storage", table.Columns
		}
	}
	return "", nil
}

func generateClauseCode(sql string, qualifier builder.ClauseQualifier, receiver string, columns []objects.Column) string {
	_, clauseCode, ok := builder.UnmarshalClause(sql, qualifier)
	if !ok {
		normalized := builder.NormalizeClauseSQL(sql, qualifier)
		return fmt.Sprintf("b.Clause(%q)", normalized)
	}
	if receiver != "" && len(columns) > 0 {
		return injectColumnReferences(clauseCode, receiver, columns)
	}
	return clauseCode
}

func injectColumnReferences(code, receiver string, columns []objects.Column) string {
	if len(columns) == 0 {
		return code
	}
	result := code
	for _, col := range columns {
		quoted := fmt.Sprintf("\"%s\"", col.Name)
		if !strings.Contains(result, quoted) {
			continue
		}
		fieldName := utils.SnakeCaseToPascalCase(col.Name)
		replacement := fmt.Sprintf("b.ColOf(&%s, &%s.%s)", receiver, receiver, fieldName)
		result = strings.ReplaceAll(result, quoted, replacement)
	}
	return result
}
