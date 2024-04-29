package config

import (
	"github.com/NikitaTsaralov/utils/connectors/jaeger"
	"github.com/NikitaTsaralov/utils/connectors/postgres"
)

const Path = "config/config.yaml"

type Config struct {
	Postgres postgres.Config
	Jaeger   jaeger.Config
}
