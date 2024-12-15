package utils

import "github.com/mxcd/go-config/config"

func LoadConfig() error {
	return config.LoadConfigWithOptions([]config.Value{
		config.String("LOG_LEVEL").NotEmpty().Default("info"),
		config.String("BOT_TOKEN").NotEmpty().Sensitive(),

		config.String("REDIS_HOST").NotEmpty().Default("redis:6379"),
		config.String("REDIS_PASSWORD").Default("").Sensitive(),
		config.Int("REDIS_DB").Default(0),

		config.Int("RSS_INTERVAL").Default(60),
	}, &config.LoadConfigOptions{DotEnvFile: "rss-telegram.env"})
}
