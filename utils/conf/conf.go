package conf

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/pfhds/live-stream-ai/models"
	"github.com/spf13/viper"
)

func New() (*models.Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")

	var config models.Config
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Printf("Configuration file changed")
	})

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
