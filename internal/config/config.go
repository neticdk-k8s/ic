package config

import (
	env "github.com/Netflix/go-env"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Logging struct {
		Level     string `env:"LOG_LEVEL,default=info"`
		Formatter string `env:"LOG_FORMATTER,default=json"`
	}
	ServerAPIEndpoint string `env:"SERVER_API_ENDPOINT,default=http://localhost:8086"`
	Debug             bool   `env:"DEBUG,default=false"`
	Extras            env.EnvSet
}

func NewConfig() Config {
	var c Config
	es, err := env.UnmarshalFromEnviron(&c)
	if err != nil {
		log.Fatal().Err(err).Msg("getting environment variables")
	}
	c.Extras = es
	return c
}
