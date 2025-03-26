package generator

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

var ImportLogger hclog.Logger = logger.HcLog().Named("generator.import")

// ----- Define type, variable and constant -----
type GenerateImportMainFunctionData struct {
	Package string
	Imports []string
	Mode    raiden.Mode
}

const (
	ImportMainFunctionDirTemplate = "/cmd/import"
	ImportMainFunctionTemplate    = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{- end }}


func main() {
	f := resource.Flags{}

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			f.CheckAndActivateDebug(cmd)
			// load configuration
			if f.ProjectPath == "" {
				curDir, err := utils.GetCurrentDirectory()
				if err != nil {
					imports.ImportLogger.Error(err.Error())
					return
				}
				f.ProjectPath = curDir
			}

			config, err := raiden.LoadConfig(nil)
			if err != nil {
				imports.ImportLogger.Error(err.Error())
				return
			}

			// register app resource
			bootstrap.RegisterModels()
			bootstrap.RegisterTypes()
			{{if eq .Mode "bff"}}
			bootstrap.RegisterRpc()
			bootstrap.RegisterRoles()
			bootstrap.RegisterStorages()
			{{end}}

			if err = generate.Run(&f.Generate, config, f.ProjectPath, false); err != nil {
				imports.ImportLogger.Error(err.Error())
			}

			if err := resource.Import(&f, config); err != nil {
				imports.ImportLogger.Error(err.Error())
			}

			if !f.DryRun {
				imports.ImportLogger.Info("regenerate bootstrap file")
				if err = generate.Run(&f.Generate, config, f.ProjectPath, false); err != nil {
					imports.ImportLogger.Error(err.Error())
					return
				}
				imports.ImportLogger.Info("finish import process")
			}
		},
	}

	f.BindLog(cmd)
	cmd.Flags().StringVarP(&f.ProjectPath, "project-path", "p", "", "set project path")
	cmd.Flags().BoolVarP(&f.RpcOnly, "rpc-only", "", false, "import rpc only")
	cmd.Flags().BoolVarP(&f.RolesOnly, "roles-only", "r", false, "import roles only")
	cmd.Flags().BoolVarP(&f.ModelsOnly, "models-only", "m", false, "import models only")
	cmd.Flags().BoolVarP(&f.StoragesOnly, "storages-only", "", false, "import storages only")
	cmd.Flags().StringVarP(&f.AllowedSchema, "schema", "s", "", "set allowed schema to import, use coma separator for multiple schema")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "run import in simulate mode without actual import resource as code")

	f.Generate.Bind(cmd)

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
`
)

// ----- Generate main function -----

func GenerateImportMainFunction(basePath string, config *raiden.Config, generateFn GenerateFn) error {
	// make sure all folder exist
	cmdFolderPath := filepath.Join(basePath, "cmd")
	ImportLogger.Trace("create cmd folder if not exist", "path", cmdFolderPath)
	if exist := utils.IsFolderExists(cmdFolderPath); !exist {
		if err := utils.CreateFolder(cmdFolderPath); err != nil {
			return err
		}
	}

	importMainFunctionPath := filepath.Join(basePath, ImportMainFunctionDirTemplate)
	ImportLogger.Trace("create import folder folder if not exist", "path", importMainFunctionPath)
	if exist := utils.IsFolderExists(importMainFunctionPath); !exist {
		if err := utils.CreateFolder(importMainFunctionPath); err != nil {
			return err
		}
	}

	// set file path
	filePath := filepath.Join(importMainFunctionPath, "main.go")

	// setup import path
	importPaths := []string{
		fmt.Sprintf("%q", "github.com/sev-2/raiden"),
		fmt.Sprintf("%q", "github.com/sev-2/raiden/pkg/cli/generate"),
		fmt.Sprintf("%q", "github.com/sev-2/raiden/pkg/cli/imports"),
		fmt.Sprintf("%q", "github.com/sev-2/raiden/pkg/resource"),
		fmt.Sprintf("%q", "github.com/sev-2/raiden/pkg/utils"),
		fmt.Sprintf("%q", "github.com/spf13/cobra"),
	}
	rpcImportPath := fmt.Sprintf("\"%s/internal/bootstrap\"", utils.ToGoModuleName(config.ProjectName))
	importPaths = append(importPaths, rpcImportPath)
	data := GenerateImportMainFunctionData{
		Package: "main",
		Imports: importPaths,
		Mode:    config.Mode,
	}

	// setup generate input param
	input := GenerateInput{
		BindData:     data,
		Template:     ImportMainFunctionTemplate,
		TemplateName: "importMainFunctionTemplate",
		OutputPath:   filePath,
	}

	ImportLogger.Debug("generate import main function", "path", input.OutputPath)
	return generateFn(input, nil)
}
