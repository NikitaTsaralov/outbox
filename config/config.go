package config

import "time"

type Config struct {
	MessageRelay     MessageRelay
	GarbageCollector GarbageCollector
}

type GarbageCollector struct {
	Timeout time.Duration `validate:"required"` // ms
	TTL     time.Duration `validate:"required"` // ms
}

type MessageRelay struct {
	Timeout   time.Duration `validate:"required"` // ms
	BatchSize int           `validate:"required"` // number of messages
}
