package config

import (
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
)

type Config struct {
	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresPort     string `mapstructure:"POSTGRES_PORT"`
	PostgresDB       string `mapstructure:"POSTGRES_DB"`
	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	RedisHost        string `mapstructure:"REDIS_HOST"`
	RedisPort        string `mapstructure:"REDIS_PORT"`
	Port             string `mapstructure:"PORT"`

	WebHookSite WebHookSite `mapstructure:",squash"`
}

type WebHookSite struct {
	URL string `mapstructure:"WEBHOOK_SITE_URL"`
	Key string `mapstructure:"WEBHOOK_SITE_KEY"`
}

func LoadConfig() (Config, error) {
	viper.AddConfigPath("configs")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("POSTGRES_DB", "insider")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("WEBHOOK_SITE_URL", "https://webhook.site/66422dd3-a84d-4099-9bd1-4192b420071c")
	viper.SetDefault("WEBHOOK_SITE_KEY", "INS.me1x9uMcyYGlhKKQVPoc.bO3j9aZwRTOcA2Ywo")

	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("read config: %v failed, use default values", err)
	}

	var conf Config
	err := viper.Unmarshal(&conf)
	if err != nil {
		return conf, fmt.Errorf("unmarshal configuration: %w", err)
	}

	return conf, nil
}
