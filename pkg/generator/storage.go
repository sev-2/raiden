package generator

import (
	"fmt"
	"html/template"
	"path/filepath"
	"reflect"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type GenerateStoragesData struct {
	Imports           []string
	Package           string
	Name              string
	StructName        string
	Public            bool
	FileSizeLimit     int
	AvifAutoDetection bool
	AllowedMimeTypes  string
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
	
	type {{ .StruckName | ToGoIdentifier }} struct {
		raiden.StorageBase
	}
	
	func (r *{{ .StruckName | ToGoIdentifier }}) Name() string {
		return "{{ .Name }}"
	}

	{{- if not .Public }}
	func (r *{{ .StruckName | ToGoIdentifier }}) Public() bool {
		return {{ .Public }}
	}
	{{- end }}
	{{- if ne .FileSizeLimit 0}}
	
	func (r *{{ .StruckName | ToGoIdentifier }}) FileSizeLimit() int {
		return {{ .FileSizeLimit }} // bytes
	}
	{{- end }}
	{{- if .AvifAutoDetection }}
	func (r *{{ .StruckName | ToGoIdentifier }}) AvifAutoDetection() bool {
		return {{ .AvifAutoDetection }} 
	}
	
	{{- end }}
	{{- if ne .AllowedMimeTypes "" }}

	func (r *{{ .StruckName | ToGoIdentifier }}) AllowedMimeTypes() []string {
		return {{ .AllowedMimeTypes }}
	}
	{{- end }}
	`
)

func GenerateStorages(basePath string, storages []objects.Storage, generateFn GenerateFn) (err error) {
	folderPath := filepath.Join(basePath, StorageDir)
	logger.Debugf("GenerateStorages - create %s folder if not exist", folderPath)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for _, v := range storages {
		if err := GenerateStorage(folderPath, v, generateFn); err != nil {
			return err
		}
	}

	return nil
}

func GenerateStorage(folderPath string, storage objects.Storage, generateFn GenerateFn) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
	}

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", storage.Name, "go"))

	// set imports path
	var imports []string
	raidenPath := fmt.Sprintf("%q", "github.com/sev-2/raiden")
	imports = append(imports, raidenPath)

	var fileSizeLimit = 0
	if storage.FileSizeLimit != nil {
		fileSizeLimit = *storage.FileSizeLimit
	}

	var allowedMimeTypes = ""
	if storage.AllowedMimeTypes != nil && len(storage.AllowedMimeTypes) > 0 {
		allowedMimeTypes = generateArrayDeclaration(reflect.ValueOf(storage.AllowedMimeTypes), false)
	}

	structName := utils.ToSnakeCase(storage.Name)
	// execute the template and write to the file
	data := GenerateStoragesData{
		Package:           "storages",
		Imports:           imports,
		Name:              storage.Name,
		StructName:        structName,
		Public:            storage.Public,
		FileSizeLimit:     fileSizeLimit,
		AvifAutoDetection: storage.AvifAutoDetection,
		AllowedMimeTypes:  allowedMimeTypes,
	}

	// set input
	input := GenerateInput{
		BindData:     data,
		Template:     StorageTemplate,
		TemplateName: "storageTemplate",
		OutputPath:   filePath,
		FuncMap:      funcMaps,
	}

	logger.Debugf("GenerateStorages - generate storages to %s", input.OutputPath)
	return generateFn(input, nil)
}
