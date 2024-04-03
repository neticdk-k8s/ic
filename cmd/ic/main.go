package main

import (
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/config"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/logging"
	"github.com/neticdk-k8s/k8s-inventory-cli/internal/version"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.NewConfig()
	logging.InitLogger(cfg.Logging.Level, cfg.Logging.Formatter)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Info().Str("version", version.VERSION).Str("commit", version.COMMIT).Msg("starting k8s-inventory-client")
}
