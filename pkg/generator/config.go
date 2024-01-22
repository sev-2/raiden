package generator

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/utils"
)

var configDir = "configs"
var configFile = "app"
var configTemplate = `PROJECT_NAME: {{ .ProjectName }}
GO_MODULE_NAME: {{ .GoModuleName }}
DEPLOYMENT_TARGET: {{ .DeploymentTarget }}
CLOUD_ACCESS_TOKEN: {{ .CloudAccessToken }}

SUPABASE_API_URL: {{ .SupabaseApiUrl }}
SUPABASE_API_BASE_PATH: {{ .SupabaseApiBaseUrl }}
SUPABASE_REST_URL: {{ .SupabaseRestUrl }}

SERVER_HOST: {{ .ServerHost }}
SERVER_PORT: {{ .ServerPort }}
`

func GenerateConfig(config raiden.Config) error {
	folderPath := filepath.Join(config.ProjectName, configDir)
	err := utils.CreateFolder(folderPath)
	if err != nil {
		return err
	}

	tmpl, err := template.New("configTemplate").Parse(configTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template : %v", err)
	}

	// Create or open the output file
	file, err := createFile(getAbsolutePath(folderPath), configFile, "yaml")
	if err != nil {
		return fmt.Errorf("failed create file %s : %v", configFile, err)
	}
	defer file.Close()

	if config.ServerHost == "" {
		config.ServerHost = "127.0.01"
	}

	if config.ServerPort == "" {
		config.ServerPort = "8002"
	}

	// Execute the template and write to the file
	err = tmpl.Execute(file, config)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	return nil
}
