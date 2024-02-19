package generator

import (
	"fmt"
	"path/filepath"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Define type, var and constant -----

const (
	ConfigDir      = "configs"
	ConfigFile     = "app"
	ConfigTemplate = `PROJECT_NAME: {{ .ProjectName }}
PROJECT_ID: {{ .ProjectId }}
DEPLOYMENT_TARGET: {{ .DeploymentTarget }}

ACCESS_TOKEN: {{ .AccessToken }}
ANON_KEY: {{ .AnonKey }}
SERVICE_KEY: {{ .ServiceKey }}

SUPABASE_API_URL: {{ .SupabaseApiUrl }}
SUPABASE_API_BASE_PATH: {{ .SupabaseApiBasePath }}
SUPABASE_PUBLIC_URL: {{ .SupabasePublicUrl }}

SERVER_HOST: {{ .ServerHost }}
SERVER_PORT: {{ .ServerPort }}

ENVIRONMENT: development
VERSION: 1.0.0

BREAKER_ENABLE: {{ .BreakerEnable }}

TRACE_ENABLE: {{ .TraceEnable }}
TRACE_COLLECTOR: {{ .TraceCollector}}
TRACE_COLLECTOR_ENDPOINT: {{ .TraceCollectorEndpoint }}
`
)

func GenerateConfig(basePath string, config *raiden.Config, generateFn GenerateFn) error {
	// create config folder if not exist
	configPath := filepath.Join(basePath, ConfigDir)
	if exist := utils.IsFolderExists(configPath); !exist {
		if err := utils.CreateFolder(configPath); err != nil {
			return err
		}
	}

	// define file path
	filePath := filepath.Join(configPath, fmt.Sprintf("%s.%s", ConfigFile, "yaml"))

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
		OutputPath:   filePath,
	}

	logger.Debugf("GenerateConfig - generate config to %s", input.OutputPath)
	return generateFn(input, nil)
}
