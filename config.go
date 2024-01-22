package raiden

import (
	"path/filepath"

	"github.com/ory/viper"
)

type DeploymentTarget string

const (
	DeploymentTargetCloud      DeploymentTarget = "cloud"
	DeploymentTargetSelfHosted DeploymentTarget = "self_hosted"
)

type Config struct {
	ProjectName        string           `mapstructure:"PROJECT_NAME"`
	DeploymentTarget   DeploymentTarget `mapstructure:"DEPLOYMENT_TARGET"`
	GoModuleName       string           `mapstructure:"GO_MODULE_NAME"`
	CloudAccessToken   string           `mapstructure:"CLOUD_ACCESS_TOKEN"`
	SupabaseApiUrl     string           `mapstructure:"SUPABASE_API_URL"`
	SupabaseApiBaseUrl string           `mapstructure:"SUPABASE_API_BASE_PATH"`
	SupabaseRestUrl    string           `mapstructure:"SUPABASE_REST_URL"`
	ServerHost         string           `mapstructure:"SERVER_HOST"`
	ServerPort         string           `mapstructure:"SERVER_PORT"`
}

func LoadConfig(path *string) *Config {
	if path != nil {
		folderPath := filepath.Dir(*path)
		file := filepath.Base(*path)

		fileExtension := filepath.Ext(file)[1:]
		fileName := file[:len(file)-len(fileExtension)-1]

		Info("set config file name to ", fileName)
		viper.SetConfigName(fileName)

		Info("set config extension to ", fileExtension)
		viper.SetConfigType(fileExtension)

		Info("set config folder to ", folderPath)
		viper.AddConfigPath(folderPath)

		Info("read configuration from ", *path)
	} else {
		viper.SetConfigName("app")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")
		Info("read configuration from ")
	}

	if err := viper.ReadInConfig(); err != nil {
		Errorf("Error reading config file: %s\n", err)
		return nil
	}

	var config Config
	Info("try marshall configuration")
	if err := viper.Unmarshal(&config); err != nil {
		Errorf("Error unmarshalling config: %s\n", err)
		return nil
	}
	Info("success load configuration")

	return &config
}
