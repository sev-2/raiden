package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var GeneratorLogger hclog.Logger = logger.HcLog().Named("generator")

// ----- Define type, variable and constant -----
type GenerateInput struct {
	BindData     any
	Template     string
	TemplateName string
	OutputPath   string
	FuncMap      []template.FuncMap
}

type GenerateFn func(input GenerateInput, writer io.Writer) error

// ----- Generate functionality  -----
func DefaultWriter(filePath string) (*os.File, error) {
	file, err := utils.CreateFile(filePath, true)
	if err != nil {
		return nil, fmt.Errorf("failed create file %s : %v", filePath, err)
	}

	return file, nil
}

type FileWriter struct {
	FilePath string
	file     *os.File
}

func (fw *FileWriter) Write(p []byte) (int, error) {
	if fw.file == nil {
		file, err := utils.CreateFile(fw.FilePath, true)
		if err != nil {
			return 0, fmt.Errorf("failed create file %s : %v", fw.FilePath, err)
		}
		fw.file = file
		defer fw.Close()
	}

	formattedCode, err := format.Source(p)
	if err != nil {
		return 0, fmt.Errorf("error format code : %v", err)
	}

	return fw.file.Write(formattedCode)
}

// Close closes the underlying file
func (fw *FileWriter) Close() error {
	if fw.file != nil {
		return fw.file.Close()
	}
	return nil
}

func Generate(input GenerateInput, writer io.Writer) error {
	// set default writer
	if writer == nil {
		file, err := DefaultWriter(input.OutputPath)
		if err != nil {
			return err
		}
		writer = file
	}

	tmplInstance := template.New(input.TemplateName)
	for _, tm := range input.FuncMap {
		tmplInstance.Funcs(tm)
	}

	tmpl, err := tmplInstance.Parse(input.Template)
	if err != nil {
		return fmt.Errorf("error parsing : %v", err)
	}

	var renderedCode bytes.Buffer
	err = tmpl.Execute(&renderedCode, input.BindData)
	if err != nil {
		return fmt.Errorf("error execute template : %v", err)
	}

	_, err = writer.Write(renderedCode.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func CreateInternalFolder(basePath string) (err error) {
	internalFolderPath := filepath.Join(basePath, "internal")
	GeneratorLogger.Trace("create internal folder if not exist", "path", internalFolderPath)
	if exist := utils.IsFolderExists(internalFolderPath); !exist {
		if err := utils.CreateFolder(internalFolderPath); err != nil {
			return err
		}
	}
	return nil
}

func GenerateArrayDeclaration(value reflect.Value, withoutQuote bool) string {
	var arrayValues []string
	for i := 0; i < value.Len(); i++ {
		if withoutQuote {
			arrayValues = append(arrayValues, fmt.Sprintf("%s", value.Index(i).Interface()))
		} else {
			arrayValues = append(arrayValues, fmt.Sprintf("%q", value.Index(i).Interface()))
		}
	}
	return "[]string{" + strings.Join(arrayValues, ", ") + "}"
}

func getStructByBaseName(filePath string, baseStructName string) (r []string, err error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return r, err
	}

	// Traverse the AST to find the struct with the Http attribute
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			if len(st.Fields.List) == 0 {
				continue
			}

			for _, f := range st.Fields.List {
				if se, isSe := f.Type.(*ast.SelectorExpr); isSe && se.Sel.Name == baseStructName {
					r = append(r, typeSpec.Name.Name)
					continue
				}
			}

		}
	}

	return
}

type aclInfo struct {
	Body           string
	HasConfigure   bool
	UseBuilder     bool
	UseRoles       bool
	UseNativeRoles bool
}

type modelRoleRef struct {
	Name        string
	VarName     string
	Constructor string
	IsNative    bool
}

type aclBuildOptions struct {
	StorageBucketName string
}

func normalizeImports(imports []string) []string {
	seen := make(map[string]struct{})
	normalized := make([]string, 0, len(imports))
	for _, imp := range imports {
		imp = strings.TrimSpace(imp)
		if imp == "" {
			continue
		}
		if !strings.Contains(imp, "\"") {
			imp = fmt.Sprintf("\"%s\"", imp)
		}
		if _, ok := seen[imp]; ok {
			continue
		}
		seen[imp] = struct{}{}
		normalized = append(normalized, imp)
	}
	sort.Strings(normalized)
	return normalized
}

func buildAclInfo(structName, receiver string, table objects.Table, policies objects.Policies, roleMap map[string]string, nativeRoleMap map[string]raiden.Role, opts *aclBuildOptions) (aclInfo, error) {
	info := aclInfo{}
	if !table.RLSEnabled && !table.RLSForced && len(policies) == 0 {
		return info, nil
	}

	body := make([]string, 0)

	bucketName := ""
	isStorageScope := false
	if opts != nil {
		bucketName = strings.TrimSpace(opts.StorageBucketName)
		isStorageScope = bucketName != ""
	}

	if table.RLSEnabled && table.RLSForced {
		body = append(body, fmt.Sprintf("\t%s.Acl.Enable().Forced()", receiver))
	} else if table.RLSEnabled {
		body = append(body, fmt.Sprintf("\t%s.Acl.Enable()", receiver))
	} else if table.RLSForced {
		body = append(body, fmt.Sprintf("\t%s.Acl.Forced()", receiver))
	}

	policiesCopy := make(objects.Policies, len(policies))
	copy(policiesCopy, policies)
	sort.SliceStable(policiesCopy, func(i, j int) bool {
		ci := policyCommandOrder(policiesCopy[i].Command)
		cj := policyCommandOrder(policiesCopy[j].Command)
		if ci == cj {
			return strings.ToLower(policiesCopy[i].Name) < strings.ToLower(policiesCopy[j].Name)
		}
		return ci < cj
	})

	roleDecls := make(map[string]*modelRoleRef)
	varNames := make(map[string]bool)
	categoryRules := map[string][]string{}
	qualifier := builder.ClauseQualifier{Schema: table.Schema, Table: table.Name}

	for _, policy := range policiesCopy {
		definition, check := normalizePolicyClausesForModel(policy)
		if isStorageScope {
			definition = builder.StripStorageBucketFilter(definition, bucketName)
			check = builder.StripStorageBucketFilter(check, bucketName)
		}

		definitionTrimmed := strings.TrimSpace(definition)
		definitionCode := ""
		if definitionTrimmed != "" {
			definitionCode = generateClauseCode(definition, qualifier, receiver, table.Columns)
		}
		if isStorageScope {
			if definitionCode == "" {
				definitionCode = "st.True"
			}
			definitionCode = fmt.Sprintf("st.StorageUsingClause(%s, %s)", fmt.Sprintf("%s.Name()", receiver), definitionCode)
			info.UseBuilder = true
		} else if definitionCode != "" {
			info.UseBuilder = true
		}

		checkTrimmed := strings.TrimSpace(check)
		checkCode := ""
		if checkTrimmed != "" {
			checkCode = generateClauseCode(check, qualifier, receiver, table.Columns)
		}
		if isStorageScope && policy.Check != nil {
			if checkCode == "" {
				checkCode = "st.True"
			}
			checkCode = fmt.Sprintf("st.StorageCheckClause(%s, %s)", fmt.Sprintf("%s.Name()", receiver), checkCode)
			info.UseBuilder = true
		} else if checkCode != "" {
			info.UseBuilder = true
		}

		roleArgs, useRoles, useNative, err := resolvePolicyRoles(policy.Roles, roleDecls, varNames, roleMap, nativeRoleMap)
		if err != nil {
			return info, err
		}
		if useRoles {
			info.UseRoles = true
		}
		if useNative {
			info.UseNativeRoles = true
		}

		commandStr := policyCommandToRaiden(policy.Command)
		modeMethod := "WithPermissive"
		if strings.EqualFold(policy.Action, raiden.AclModeRestrictive.ActionString()) {
			modeMethod = "WithRestrictive"
		}

		ruleLine := formatModelRule(policy.Name, roleArgs, commandStr, definitionCode, checkCode, modeMethod)
		category := categorizePolicyCommand(policy.Command)
		categoryRules[category] = append(categoryRules[category], ruleLine)
	}

	if len(roleDecls) > 0 {
		if len(body) > 0 {
			body = append(body, "")
		}
		body = append(body, "\t// related role")
		decls := collectRoleDeclarations(roleDecls)
		for _, decl := range decls {
			body = append(body, fmt.Sprintf("\t%s := %s", decl.VarName, decl.Constructor))
		}
	}

	order := []string{"Read Rule", "Write Rule", "All Action", "Additional Rule"}
	for _, category := range order {
		rules := categoryRules[category]
		if len(rules) == 0 {
			continue
		}
		if len(body) > 0 {
			body = append(body, "")
		}
		body = append(body, fmt.Sprintf("\t// %s", category))
		body = append(body, fmt.Sprintf("\t%s.Acl.Define(", receiver))
		body = append(body, strings.Join(rules, "\n"))
		body = append(body, "\t)")
	}

	info.Body = strings.Join(body, "\n")
	info.HasConfigure = len(body) > 0
	return info, nil
}

func resolvePolicyRoles(roleNames []string, decls map[string]*modelRoleRef, existing map[string]bool, roleMap map[string]string, nativeRoleMap map[string]raiden.Role) ([]string, bool, bool, error) {
	seen := make(map[string]struct{})
	args := make([]string, 0, len(roleNames))
	useRoles := false
	useNative := false
	for _, role := range roleNames {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		// "public" is a PostgreSQL pseudo-role meaning "all roles", not a real role
		if role == "public" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}

		ref, err := ensureRoleRef(role, decls, existing, roleMap, nativeRoleMap)
		if err != nil {
			return nil, false, false, err
		}
		if ref.IsNative {
			useNative = true
		} else {
			useRoles = true
		}
		args = append(args, fmt.Sprintf("%s.Name()", ref.VarName))
	}
	return args, useRoles, useNative, nil
}

func ensureRoleRef(role string, decls map[string]*modelRoleRef, existing map[string]bool, roleMap map[string]string, nativeRoleMap map[string]raiden.Role) (*modelRoleRef, error) {
	if ref, ok := decls[role]; ok {
		return ref, nil
	}

	pascal := utils.SnakeCaseToPascalCase(role)
	if pascal == "" {
		pascal = "Role"
	}
	varName := makeRoleVarName(pascal, existing)

	ref := &modelRoleRef{
		Name:    role,
		VarName: varName,
	}

	if _, ok := nativeRoleMap[role]; ok {
		ref.IsNative = true
		ref.Constructor = fmt.Sprintf("native_role.%s{}", pascal)
		decls[role] = ref
		return ref, nil
	}

	if roleMap != nil {
		if _, ok := roleMap[role]; !ok {
			return nil, fmt.Errorf("generator: role %s is not available", role)
		}
	}

	ref.Constructor = fmt.Sprintf("roles.%s{}", pascal)
	decls[role] = ref
	return ref, nil
}

func makeRoleVarName(pascal string, existing map[string]bool) string {
	if pascal == "" {
		pascal = "Role"
	}
	base := strings.ToLower(pascal[:1]) + pascal[1:]
	candidate := base
	index := 1
	for existing[candidate] {
		index++
		candidate = fmt.Sprintf("%s%d", base, index)
	}
	existing[candidate] = true
	return candidate
}

func collectRoleDeclarations(decls map[string]*modelRoleRef) []*modelRoleRef {
	items := make([]*modelRoleRef, 0, len(decls))
	for _, ref := range decls {
		items = append(items, ref)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].VarName < items[j].VarName
	})
	return items
}

func policyCommandOrder(cmd objects.PolicyCommand) int {
	switch cmd {
	case objects.PolicyCommandSelect:
		return 0
	case objects.PolicyCommandInsert:
		return 1
	case objects.PolicyCommandUpdate:
		return 2
	case objects.PolicyCommandDelete:
		return 3
	case objects.PolicyCommandAll:
		return 4
	default:
		return 5
	}
}

func policyCommandToRaiden(cmd objects.PolicyCommand) string {
	switch cmd {
	case objects.PolicyCommandSelect:
		return "raiden.CommandSelect"
	case objects.PolicyCommandInsert:
		return "raiden.CommandInsert"
	case objects.PolicyCommandUpdate:
		return "raiden.CommandUpdate"
	case objects.PolicyCommandDelete:
		return "raiden.CommandDelete"
	case objects.PolicyCommandAll:
		return "raiden.CommandAll"
	default:
		return fmt.Sprintf("raiden.Command%s", utils.SnakeCaseToPascalCase(strings.ToLower(string(cmd))))
	}
}

func categorizePolicyCommand(cmd objects.PolicyCommand) string {
	switch cmd {
	case objects.PolicyCommandSelect:
		return "Read Rule"
	case objects.PolicyCommandInsert, objects.PolicyCommandUpdate, objects.PolicyCommandDelete:
		return "Write Rule"
	case objects.PolicyCommandAll:
		return "All Action"
	default:
		return "Additional Rule"
	}
}

func formatModelRule(name string, roleArgs []string, command, using, check, mode string) string {
	base := fmt.Sprintf("\t\traiden.Rule(%q)", name)
	if len(roleArgs) > 0 {
		base += fmt.Sprintf(".For(%s)", strings.Join(roleArgs, ", "))
	}
	base += fmt.Sprintf(".To(%s)", command)

	suffix := make([]string, 0, 3)
	if using != "" {
		suffix = append(suffix, fmt.Sprintf("Using(%s)", using))
	}
	if check != "" {
		suffix = append(suffix, fmt.Sprintf("Check(%s)", check))
	}
	suffix = append(suffix, fmt.Sprintf("%s()", mode))

	if len(suffix) > 0 {
		base += ".\n\t\t\t" + strings.Join(suffix, ".\n\t\t\t")
	}

	return base + ","
}

func normalizePolicyClausesForModel(policy objects.Policy) (definition string, check string) {
	definition = strings.TrimSpace(policy.Definition)
	if policy.Check != nil {
		check = strings.TrimSpace(*policy.Check)
	}

	switch policy.Command {
	case objects.PolicyCommandSelect:
		definition = ensureClauseString(definition, true)
		check = ""
	case objects.PolicyCommandInsert:
		definition = ""
		check = ensureClauseString(check, true)
	case objects.PolicyCommandUpdate:
		definition = ensureClauseString(definition, true)
		check = ensureClauseString(check, true)
	case objects.PolicyCommandDelete:
		if strings.TrimSpace(definition) == "" {
			definition = check
		}
		definition = ensureClauseString(definition, true)
		check = ""
	case objects.PolicyCommandAll:
		definition = ensureClauseString(definition, true)
		check = ensureClauseString(check, true)
	default:
		definition = ensureClauseString(definition, false)
		check = strings.TrimSpace(check)
	}

	if strings.EqualFold(definition, "true") {
		definition = "TRUE"
	}
	if strings.EqualFold(check, "true") {
		check = "TRUE"
	}

	return definition, check
}

func ensureClauseString(value string, requireFallback bool) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" && requireFallback {
		return "TRUE"
	}
	return trimmed
}
