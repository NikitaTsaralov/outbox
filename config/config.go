package config

import (
	"github.com/NikitaTsaralov/utils/connectors/jaeger"
	"github.com/NikitaTsaralov/utils/connectors/postgres"
)

const Path = "config/config.yml"

type Config struct {
	Postgres postgres.Config
	Jaeger   jaeger.Config
}
