package config

import (
	"time"
)

type MessageRelay struct {
	Enabled   bool
	Timeout   time.Duration `validate:"required"` // ms
	BatchSize int           `validate:"required"` // amount
}
