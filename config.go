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

type Config struct {
	AccessToken        string           `mapstructure:"ACCESS_TOKEN"`
	AnonKey            string           `mapstructure:"ANON_KEY"`
	BreakerEnable      bool             `mapstructure:"BREAKER_ENABLE"`
	DeploymentTarget   DeploymentTarget `mapstructure:"DEPLOYMENT_TARGET"`
	Environment        string           `mapstructure:"ENVIRONMENT"`
	ProjectName        string           `mapstructure:"PROJECT_NAME"`
	ServiceKey         string           `mapstructure:"SERVICE_KEY"`
	ServerHost         string           `mapstructure:"SERVER_HOST"`
	ServerPort         string           `mapstructure:"SERVER_PORT"`
	SupabaseApiUrl     string           `mapstructure:"SUPABASE_API_URL"`
	SupabaseApiBaseUrl string           `mapstructure:"SUPABASE_API_BASE_PATH"`
	SupabasePublicUrl  string           `mapstructure:"SUPABASE_PUBLIC_URL"`
	TraceEnable        bool             `mapstructure:"TRACE_ENABLE"`
	TraceCollector     string           `mapstructure:"TRACE_COLLECTOR"`
	TraceEndpoint      string           `mapstructure:"TRACE_ENDPOINT"`
	Version            string           `mapstructure:"VERSION"`
}

func LoadConfig(path *string) (*Config, error) {
	filePath := "./configs/app.yaml"

	if path != nil && *path != "" {
		filePath = *path

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
	Infof("success load configuration from : %s", filePath)

	// set default value

	if config.ServerHost == "" {
		config.ServerHost = "127.0.0.1"
	}

	if config.ServerPort != "" {
		config.ServerPort = "8002"
	}

	if config.Version != "" {
		config.Version = "1.0.0"
	}

	if config.Environment == "" {
		config.Environment = "development"
	}

	return &config, nil
}
