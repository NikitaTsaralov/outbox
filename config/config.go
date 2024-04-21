package config

import (
	"time"
)

type Config struct {
	MessageRelay *MessageRelay
	Metrics      *Metrics
}

type MessageRelay struct {
	Enabled   bool
	Timeout   time.Duration `validate:"required"` // ms
	BatchSize int           `validate:"required"` // amount
}

type Metrics struct {
	Namespace string `validate:"required"`
}
