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
{{ if .CloudAccessToken }}CLOUD_ACCESS_TOKEN: {{ .CloudAccessToken }}{{ end }}

ENVIRONMENT=development
VERSION=1.0.0

SUPABASE_API_URL: {{ .SupabaseApiUrl }}
{{ if .SupabaseApiBaseUrl }}SUPABASE_API_BASE_PATH: {{ .SupabaseApiBaseUrl }}{{ end }}
{{ if .SupabaseRestUrl }}SUPABASE_REST_URL: {{ .SupabaseRestUrl }}{{ end }}

SERVER_HOST: {{ .ServerHost }}
SERVER_PORT: {{ .ServerPort }}

{{ if .TraceEnable }}TRACE_ENABLE: {{ .TraceEnable }}{{ end }}
{{ if .TraceCollector }}TRACE_COLLECTOR: {{ .TraceCollector }}{{ end }}
{{ if .TraceEndpoint }}TRACE_ENDPOINT: {{ .TraceEndpoint }}{{ end }}
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
