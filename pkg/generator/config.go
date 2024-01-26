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
DEPLOYMENT_TARGET: {{ .DeploymentTarget }}

ACCESS_TOKEN: {{ .AccessToken }}
ANON_KEY: {{ .AnonKey }}
SERVICE_KEY: {{ .ServiceKey }}

SUPABASE_API_URL: {{ .SupabaseApiUrl }}
SUPABASE_API_BASE_PATH: {{ .SupabaseApiBaseUrl }}
SUPABASE_PUBLIC_URL: {{ .SupabasePublicUrl }}

SERVER_HOST: {{ .ServerHost }}
SERVER_PORT: {{ .ServerPort }}

ENVIRONMENT: development
VERSION: 1.0.0

BREAKER_ENABLE: {{ .BreakerEnable }}

TRACE_ENABLE: {{ .TraceEnable }}
TRACE_COLLECTOR: {{ .TraceCollector }}
TRACE_ENDPOINT: {{ .TraceEndpoint }}
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
