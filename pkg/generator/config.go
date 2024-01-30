package generator

import (
	"fmt"
	"path/filepath"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Define type, var and constant -----

const (
	ConfigDir      = "configs"
	ConfigFile     = "app"
	ConfigTemplate = `PROJECT_NAME: {{ .ProjectName }}
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
)

func GenerateConfig(config *raiden.Config, generateFn GenerateFn) error {
	// create config folder if not exist
	folderPath := filepath.Join(config.ProjectName, ConfigDir)
	if err := utils.CreateFolder(folderPath); err != nil {
		return err
	}

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", ConfigFile, "yaml"))
	absolutePath, err := utils.GetAbsolutePath(filePath)
	if err != nil {
		return err
	}

	if config.ServerHost == "" {
		config.ServerHost = "127.0.01"
	}

	if config.ServerPort == "" {
		config.ServerPort = "8002"
	}

	input := GenerateInput{
		BindData:     config,
		Template:     ConfigTemplate,
		TemplateName: "configTemplate",
		OutputPath:   absolutePath,
	}

	return generateFn(input)
}
