package raiden

import (
	"path/filepath"

	"github.com/ory/viper"
)

// enum
type (
	SupabaseTargetConfig     string
	WorkloadManagementConfig string
)

const (
	// Supabase target options
	SupabaseTargetConfigCkoud SupabaseTargetConfig = "cloud"
	SupabaseTargetConfigLocal SupabaseTargetConfig = "local"

	// Workload management options
	WorkloadManagementConfigDockerCompose WorkloadManagementConfig = "docker-compose"
	WorkloadManagementConfigKubernetes    WorkloadManagementConfig = "kubernetes"
)

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Controller ControllerConfig `mapstructure:"controller"`
	Model      ModelConfig      `mapstructure:"model"`
	Supabase   SupabaseConfig   `mapstructure:"supabase"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Host    string `mapstructure:"host"`
	Port    string `mapstructure:"port"`
}

type ControllerConfig struct {
	Path string `mapstructure:"path"`
}

type ModelConfig struct {
	Path string `mapstructure:"path"`
}

type SupabaseConfig struct {
	Target             SupabaseTargetConfig     `mapstructure:"target"`
	Url                string                   `mapstructure:"url"`
	WorkloadManagement WorkloadManagementConfig `mapstructure:"workloadManagement"`
	Secret             struct {
		JWT struct {
			Secret     string `mapstructure:"secret"`
			ServiceKey string `mapstructure:"serviceKey"`
			AnonKey    string `mapstructure:"anonKey"`
			Expiry     int    `mapstructure:"expiry"`
		} `mapstructure:"jwt"`
		Basic struct {
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
		} `mapstructure:"basic"`
		Database struct {
			Host     string `mapstructure:"host"`
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			Database string `mapstructure:"database"`
			Port     int    `mapstructure:"port"`
		} `mapstructure:"database"`
	} `mapstructure:"secret"`
	Resources struct {
		Auth struct {
			Enable bool `mapstructure:"enable"`
		} `mapstructure:"auth"`
		Rest struct {
			Enable bool `mapstructure:"enable"`
		} `mapstructure:"rest"`
		Meta struct {
			Enable   bool   `mapstructure:"enable"`
			BasePath string `mapstructure:"basePath"`
		} `mapstructure:"meta"`
		Realtime struct {
			Enable bool `mapstructure:"enable"`
		} `mapstructure:"realtime"`
		Storage struct {
			Enable bool `mapstructure:"enable"`
		} `mapstructure:"storage"`
	} `mapstructure:"resources"`
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
		viper.AddConfigPath("./config")
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
