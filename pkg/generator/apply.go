package generator

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

var ApplyLogger hclog.Logger = logger.HcLog().Named("generator.apply")

// ----- Define type, variable and constant -----
type GenerateApplyMainFunctionData struct {
	Package string
	Imports []string
}

const (
	ApplyMainFunctionDirTemplate = "/cmd/apply"
	ApplyMainFunctionTemplate    = `package {{ .Package }}
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
					apply.ApplyLogger.Error(err.Error())
					return
				}
				f.ProjectPath = curDir
			}

			config, err := raiden.LoadConfig(nil)
			if err != nil {
				apply.ApplyLogger.Error(err.Error())
				return
			}

			if err := generate.Run(&f.Generate, config, f.ProjectPath, false); err != nil {
				apply.ApplyLogger.Error(err.Error())
				return
			}

			// register app resource
			bootstrap.RegisterRpc()
			bootstrap.RegisterRoles()
			bootstrap.RegisterModels()
			bootstrap.RegisterStorages()
			
			if err = resource.Apply(&f, config); err != nil {
				apply.ApplyLogger.Error(err.Error())
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

	f.Generate.Bind(cmd)

	cmd.Execute()
}
`
)

// ----- Generate main function -----

func GenerateApplyMainFunction(basePath string, config *raiden.Config, generateFn GenerateFn) error {
	// make sure all folder exist
	cmdFolderPath := filepath.Join(basePath, "cmd")
	ApplyLogger.Trace("create cmd folder if not exist", "path", cmdFolderPath)
	if exist := utils.IsFolderExists(cmdFolderPath); !exist {
		if err := utils.CreateFolder(cmdFolderPath); err != nil {
			return err
		}
	}

	applyMainFunctionPath := filepath.Join(basePath, ApplyMainFunctionDirTemplate)
	ApplyLogger.Trace("create main folder if not exist", "path", applyMainFunctionPath)
	if exist := utils.IsFolderExists(applyMainFunctionPath); !exist {
		if err := utils.CreateFolder(applyMainFunctionPath); err != nil {
			return err
		}
	}

	// set file path
	filePath := filepath.Join(applyMainFunctionPath, "main.go")

	// setup import path
	importPaths := []string{
		fmt.Sprintf("%q", "github.com/sev-2/raiden"),
		fmt.Sprintf("%q", "github.com/sev-2/raiden/pkg/cli/generate"),
		fmt.Sprintf("%q", "github.com/sev-2/raiden/pkg/cli/apply"),
		fmt.Sprintf("%q", "github.com/sev-2/raiden/pkg/resource"),
		fmt.Sprintf("%q", "github.com/sev-2/raiden/pkg/utils"),
		fmt.Sprintf("%q", "github.com/spf13/cobra"),
	}
	rpcImportPath := fmt.Sprintf("\"%s/internal/bootstrap\"", utils.ToGoModuleName(config.ProjectName))
	importPaths = append(importPaths, rpcImportPath)
	data := GenerateApplyMainFunctionData{
		Package: "main",
		Imports: importPaths,
	}

	// setup generate input param
	input := GenerateInput{
		BindData:     data,
		Template:     ApplyMainFunctionTemplate,
		TemplateName: "applyMainFunctionTemplate",
		OutputPath:   filePath,
	}

	ApplyLogger.Debug("generate apply main function", "path", input.OutputPath)
	return generateFn(input, nil)
}
