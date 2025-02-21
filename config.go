package raiden

import (
	"path/filepath"

	"github.com/ory/viper"
)

// ----- main configuration functionality -----

type DeploymentTarget string

const (
	DeploymentTargetCloud      DeploymentTarget = "cloud"
	DeploymentTargetSelfHosted DeploymentTarget = "self_hosted"
)

type Mode string

const (
	BffMode Mode = "bff"
	SvcMode Mode = "svc"
)

type TokenType string

const (
	TokenTypeBasic  TokenType = "basic"
	TokenTypeBearer TokenType = "bearer"
)

type Config struct {
	AccessToken              string           `mapstructure:"ACCESS_TOKEN"`
	AnonKey                  string           `mapstructure:"ANON_KEY"`
	AllowedTables            string           `mapstructure:"ALLOWED_TABLES"`
	BreakerEnable            bool             `mapstructure:"BREAKER_ENABLE"`
	CorsAllowedOrigins       string           `mapstructure:"CORS_ALLOWED_ORIGINS"`
	CorsAllowedMethods       string           `mapstructure:"CORS_ALLOWED_METHODS"`
	CorsAllowedHeaders       string           `mapstructure:"CORS_ALLOWED_HEADERS"`
	CorsAllowCredentials     bool             `mapstructure:"CORS_ALLOWED_CREDENTIALS"`
	DeploymentTarget         DeploymentTarget `mapstructure:"DEPLOYMENT_TARGET"`
	Environment              string           `mapstructure:"ENVIRONMENT"`
	GoogleProjectId          string           `mapstructure:"GOOGLE_PROJECT_ID"`
	GoogleSaPath             string           `mapstructure:"GOOGLE_SA_PATH"`
	JwtToken                 string           `mapstructure:"JWT_TOKEN"`
	LogLevel                 string           `mapstructure:"LOG_LEVEL"`
	MaxServerRequestBodySize int              `mapstructure:"MAX_SERVER_REQUEST_BODY_SIZE"`
	Mode                     Mode             `mapstructure:"MODE"`
	PgMetaUrl                string           `mapstructure:"PG_META_URL"`
	PostgRestUrl             string           `mapstructure:"POSTGREST_URL"`
	ProjectId                string           `mapstructure:"PROJECT_ID"`
	ProjectName              string           `mapstructure:"PROJECT_NAME"`
	ServiceKey               string           `mapstructure:"SERVICE_KEY"`
	ServerHost               string           `mapstructure:"SERVER_HOST"`
	ServerPort               string           `mapstructure:"SERVER_PORT"`
	ServerDns                string           `mapstructure:"SERVER_DNS"`
	SupabaseApiUrl           string           `mapstructure:"SUPABASE_API_URL"`
	SupabaseApiBasePath      string           `mapstructure:"SUPABASE_API_BASE_PATH"`
	SupabaseApiToken         string           `mapstructure:"SUPABASE_API_TOKEN"`
	SupabaseApiTokenType     string           `mapstructure:"SUPABASE_API_TOKEN_TYPE"`
	SupabasePublicUrl        string           `mapstructure:"SUPABASE_PUBLIC_URL"`
	ScheduleStatus           ScheduleStatus   `mapstructure:"SCHEDULE_STATUS"`
	TraceEnable              bool             `mapstructure:"TRACE_ENABLE"`
	TraceCollector           string           `mapstructure:"TRACE_COLLECTOR"`
	TraceCollectorEndpoint   string           `mapstructure:"TRACE_COLLECTOR_ENDPOINT"`
	Version                  string           `mapstructure:"VERSION"`
}

// The function `LoadConfig` loads a configuration file based on the provided path or uses default
// values if no path is provided.
func LoadConfig(path *string) (*Config, error) {
	if path != nil && *path != "" {
		folderPath := filepath.Dir(*path)
		file := filepath.Base(*path)

		fileExtension := filepath.Ext(file)[1:]
		fileName := file[:len(file)-len(fileExtension)-1]

		viper.SetConfigName(fileName)
		viper.SetConfigType(fileExtension)
		viper.AddConfigPath(folderPath)
	} else {
		viper.SetConfigName("app")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// set default value
	if config.ServerHost == "" {
		config.ServerHost = "127.0.0.1"
	}

	if config.ServerPort == "" {
		config.ServerPort = "8002"
	}

	if config.Version == "" {
		config.Version = "1.0.0"
	}

	if config.Environment == "" {
		config.Environment = "development"
	}

	if config.Mode == "" {
		config.Mode = BffMode
	}

	if config.ScheduleStatus == "" {
		config.ScheduleStatus = ScheduleStatusOff
	}

	if config.AllowedTables == "" {
		config.AllowedTables = "*"
	}

	if len(config.SupabaseApiBasePath) > 0 && config.SupabaseApiBasePath[0] != '/' {
		config.SupabaseApiBasePath = "/" + config.SupabaseApiBasePath
	}

	if config.MaxServerRequestBodySize == 0 {
		config.MaxServerRequestBodySize = 8 * 1024 * 1024 // Default Max: 8 MB
	}

	return &config, nil
}

func (*Config) GetBool(key string) bool {
	return viper.GetBool(key)
}

func (*Config) GetString(key string) string {
	return viper.GetString(key)
}

func (*Config) GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func (*Config) GetInt(key string) int {
	return viper.GetInt(key)
}

func (*Config) GetIntSlice(key string) []int {
	return viper.GetIntSlice(key)
}

func (*Config) GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}
