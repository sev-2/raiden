package generator

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

var ConfigLogger hclog.Logger = logger.HcLog().Named("generator.config")

// ----- Define type, var and constant -----

const (
	ConfigDir      = "configs"
	ConfigFile     = "app"
	ConfigTemplate = `PROJECT_NAME: {{ .ProjectName }}
PROJECT_ID: {{ .ProjectId }}
DEPLOYMENT_TARGET: {{ .DeploymentTarget }}

MODE: {{ .Mode }}

{{- if eq .Mode "bff"}}
ACCESS_TOKEN: {{ .AccessToken }}
ANON_KEY: {{ .AnonKey }}
SERVICE_KEY: {{ .ServiceKey }}

SUPABASE_API_URL: {{ .SupabaseApiUrl }}
SUPABASE_API_BASE_PATH: {{ .SupabaseApiBasePath }}
SUPABASE_PUBLIC_URL: {{ .SupabasePublicUrl }}
{{- end}}
{{- if eq .Mode "svc"}}
POSTGREST_URL: {{ .PostgRestUrl }}
PG_META_URL: {{ .PgMetaUrl }}
{{- end}}
SERVER_HOST: {{ .ServerHost }}
SERVER_PORT: {{ .ServerPort }}
SERVER_DNS: {{ .ServerDns }}

ENVIRONMENT: development
VERSION: 1.0.0

ALLOWED_TABLES: '{{ .AllowedTables }}'
SCHEDULE_STATUS: '{{ .ScheduleStatus }}'
{{- if ne .GoogleProjectId ""}}
GOOGLE_PROJECT_ID: {{ .GoogleProjectId }}
{{- end }}
{{- if ne .GoogleSaPath ""}}
GOOGLE_SA_PATH: {{ .GoogleSaPath }}
{{- end }}

BREAKER_ENABLE: {{ .BreakerEnable }}

TRACE_ENABLE: {{ .TraceEnable }}
TRACE_COLLECTOR: {{ .TraceCollector}}
TRACE_COLLECTOR_ENDPOINT: {{ .TraceCollectorEndpoint }}

MAX_SERVER_REQUEST_BODY_SIZE: {{ .MaxServerRequestBodySize }}

CORS_ALLOWED_ORIGINS:
CORS_ALLOWED_METHODS:
CORS_ALLOWED_HEADERS:
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
		config.ServerHost = "127.0.0.1"
	}

	if config.ServerPort == "" {
		config.ServerPort = "8002"
	}

	if config.MaxServerRequestBodySize == 0 {
		config.MaxServerRequestBodySize = 8 * 1024 * 1024
	}

	input := GenerateInput{
		BindData:     config,
		Template:     ConfigTemplate,
		TemplateName: "configTemplate",
		OutputPath:   filePath,
	}

	ConfigLogger.Debug("generate config", "path", input.OutputPath)
	return generateFn(input, nil)
}
