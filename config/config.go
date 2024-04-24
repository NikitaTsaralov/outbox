package config

import "github.com/NikitaTsaralov/utils/connectors/postgres"

type Config struct {
	Postgres postgres.Config
}
