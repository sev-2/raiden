package generator

import (
	"fmt"
	"html/template"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var StorageLogger hclog.Logger = logger.HcLog().Named("generator.storage")

type GenerateStorageInput struct {
	Bucket   objects.Bucket
	Policies objects.Policies
}
type GenerateStoragesData struct {
	Imports           []string
	Package           string
	Name              string
	StructName        string
	Receiver          string
	Public            bool
	FileSizeLimit     int
	AvifAutoDetection bool
	AllowedMimeTypes  string
	HasConfigureAcl   bool
	ConfigureAclBody  string
}

const (
	StorageDir      = "internal/storages"
	StorageTemplate = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{- end }}

type {{ .StructName | ToGoIdentifier }} struct {
	raiden.BucketBase

	// Access control
	Acl raiden.Acl
}

func ({{ .Receiver }} *{{ .StructName | ToGoIdentifier }}) Name() string {
	return "{{ .Name }}"
}

{{- if .Public }}
func ({{ .Receiver }} *{{ .StructName | ToGoIdentifier }}) Public() bool {
	return {{ .Public }}
}
{{- end }}
{{- if ne .FileSizeLimit 0}}

func ({{ .Receiver }} *{{ .StructName | ToGoIdentifier }}) FileSizeLimit() int {
	return {{ .FileSizeLimit }} // bytes
}
{{- end }}
{{- if ne .AllowedMimeTypes "" }}
func ({{ .Receiver }} *{{ .StructName | ToGoIdentifier }}) AllowedMimeTypes() []string {
	return {{ .AllowedMimeTypes }}
}
{{- end }}
{{- if .AvifAutoDetection }}

func ({{ .Receiver }} *{{ .StructName | ToGoIdentifier }}) AvifAutoDetection() bool {
	return {{ .AvifAutoDetection }}
}
{{- end }}
{{- if .HasConfigureAcl }}

func ({{ .Receiver }} *{{ .StructName | ToGoIdentifier }}) ConfigureAcl() {
{{ .ConfigureAclBody }}
}
{{- end }}
`
)

func GenerateStorages(basePath string, projectName string, storages []*GenerateStorageInput, roleMap map[string]string, nativeRoleMap map[string]raiden.Role, generateFn GenerateFn) (err error) {
	folderPath := filepath.Join(basePath, StorageDir)
	StorageLogger.Trace("create storages folder", "path", folderPath)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for _, v := range storages {
		if err := GenerateStorage(folderPath, projectName, v, roleMap, nativeRoleMap, generateFn); err != nil {
			return err
		}
	}

	return nil
}

func GenerateStorage(folderPath string, projectName string, storage *GenerateStorageInput, roleMap map[string]string, nativeRoleMap map[string]raiden.Role, generateFn GenerateFn) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
	}

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", utils.ToSnakeCase(storage.Bucket.Name), "go"))

	var objectPolicies objects.Policies
	for i := range storage.Policies {
		p := storage.Policies[i]
		if p.Table == supabase.DefaultObjectTable {
			objectPolicies = append(objectPolicies, p)
		}
	}

	var fileSizeLimit = 0
	if storage.Bucket.FileSizeLimit != nil {
		fileSizeLimit = *storage.Bucket.FileSizeLimit
	}

	var allowedMimeTypes = ""
	if len(storage.Bucket.AllowedMimeTypes) > 0 {
		allowedMimeTypes = GenerateArrayDeclaration(reflect.ValueOf(storage.Bucket.AllowedMimeTypes), false)
	}

	structName := utils.SnakeCaseToPascalCase(storage.Bucket.Name)
	if structName == "" {
		structName = "Storage"
	}
	receiver := strings.ToLower(string(structName[0]))

	aclInfo, err := buildStorageAclInfo(structName, receiver, storage.Bucket, objectPolicies, roleMap, nativeRoleMap)
	if err != nil {
		return err
	}

	imports := []string{fmt.Sprintf("%q", "github.com/sev-2/raiden")}
	if aclInfo.UseBuilder {
		imports = append(imports, `st "github.com/sev-2/raiden/pkg/builder"`)
	}
	moduleName := utils.ToGoModuleName(projectName)
	rolesImportPath := fmt.Sprintf("%s/internal/roles", moduleName)
	if aclInfo.UseRoles {
		imports = append(imports, fmt.Sprintf("roles %q", rolesImportPath))
	}
	if aclInfo.UseNativeRoles {
		imports = append(imports, `native_role "github.com/sev-2/raiden/pkg/postgres/roles"`)
	}
	imports = normalizeImports(imports)

	// execute the template and write to the file
	data := GenerateStoragesData{
		Package:           "storages",
		Imports:           imports,
		Name:              storage.Bucket.Name,
		StructName:        structName,
		Receiver:          receiver,
		Public:            storage.Bucket.Public,
		FileSizeLimit:     fileSizeLimit,
		AvifAutoDetection: storage.Bucket.AvifAutoDetection,
		AllowedMimeTypes:  allowedMimeTypes,
		HasConfigureAcl:   aclInfo.HasConfigure,
		ConfigureAclBody:  aclInfo.Body,
	}

	// set input
	input := GenerateInput{
		BindData:     data,
		Template:     StorageTemplate,
		TemplateName: "storageTemplate",
		OutputPath:   filePath,
		FuncMap:      funcMaps,
	}

	StorageLogger.Debug("generate storages", "path", input.OutputPath)
	return generateFn(input, nil)
}

func buildStorageAclInfo(structName, receiver string, bucket objects.Bucket, policies objects.Policies, roleMap map[string]string, nativeRoleMap map[string]raiden.Role) (aclInfo, error) {
	table := objects.Table{
		Schema:     supabase.DefaultStorageSchema,
		Name:       supabase.DefaultObjectTable,
		RLSEnabled: true,
		RLSForced:  false,
	}
	return buildAclInfo(structName, receiver, table, policies, roleMap, nativeRoleMap, &aclBuildOptions{StorageBucketName: bucket.Name})
}
