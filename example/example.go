package main

import (
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/config"
	"github.com/NikitaTsaralov/transactional-outbox/internal/application"
)

func main() {
	cfg := &config.MessageRelay{
		Enabled:   true,
		Timeout:   1000, // 1 s
		BatchSize: 10,
	}

	s := application.NewService(cfg, nil, nil, nil)
	s.Start()

	time.Sleep(3000 * time.Millisecond)
	s.Stop()
}
