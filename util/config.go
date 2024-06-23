package util

import (
	"github.com/spf13/viper"
)

const ConfigName = "config"
const ConfigType = "yaml"

var Configuration Config

type Config struct {
	Server struct {
		Mode       string `mapstructure:"mode"`
		Port       int    `mapstructure:"port"`
		Path       string `mapstructure:"path"`
		APIVersion string `mapstructure:"api_version"`
	} `mapstructure:"server"`
	Artemis struct {
		Host         string `mapstructure:"host"`
		Port         int    `mapstructure:"port"`
		User         string `mapstructure:"user"`
		Password     string `mapstructure:"password"`
		QueueName    string `mapstructure:"queue_name"`
		NumberOfData int    `mapstructure:"number_of_data"`
	} `mapstructure:"artemis"`
	MongoDB struct {
		Host           string `mapstructure:"host"`
		Port           int    `mapstructure:"port"`
		Username       string `mapstructure:"username"`
		Password       string `mapstructure:"password"`
		Database       string `mapstructure:"database"`
		Collection     string `mapstructure:"collection"`
		AuthSource     string `mapstructure:"auth_source"`
		MaxPoolSize    uint64 `mapstructure:"max_pool_size"`
		ConnectTimeout int    `mapstructure:"connect_timeout"`
		NumberOfData   int    `mapstructure:"number_of_data"`
	}
	Logger struct {
		Dir        string `mapstructure:"dir"`
		FileName   string `mapstructure:"file_name"`
		MaxBackups int    `mapstructure:"max_backups"`
		MaxSize    int    `mapstructure:"max_size"`
		MaxAge     int    `mapstructure:"max_age"`
		Compress   bool   `mapstructure:"compress"`
		LocalTime  bool   `mapstructure:"local_time"`
	} `mapstructure:"logger"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(ConfigName)
	viper.SetConfigType(ConfigType)

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	var config Config
	err = viper.Unmarshal(&config)
	Configuration = config
	return
}
