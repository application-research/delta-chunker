package config

import (
	"github.com/caarlos0/env/v6"
	logging "github.com/ipfs/go-log/v2"
	"github.com/joho/godotenv"
	"github.com/multiformats/go-multiaddr"
)

var (
	log                       = logging.Logger("config")
	defaultTestBootstrapPeers []multiaddr.Multiaddr
)

type DeltaChunkerConfig struct {
	Common struct {
		DBDSN   string `env:"DB_DSN" envDefault:"dcdb.db"`
		Commit  string `env:"COMMIT"`
		Version string `env:"VERSION"`
	}
}

func InitConfig() DeltaChunkerConfig {
	godotenv.Load() // load from environment OR .env file if it exists
	var cfg DeltaChunkerConfig

	if err := env.Parse(&cfg); err != nil {
		log.Fatal("error parsing config: %+v\n", err)
	}

	log.Debug("config parsed successfully")

	return cfg
}
