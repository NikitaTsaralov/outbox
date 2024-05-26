# github.com/NikitaTsaralov/transactional-outbox

![cqrs](./assets/ReliablePublication.png)

### Overview

> Transactional outbox is a pattern that ensures that messages are delivered to a message broker in a reliable way.

This project is my own implementation of transaction-outbox pattern for personal use

### Public interface

All necessary functions your application may need to safely publish messages to broker (kafka)

```go
// TransactionalOutbox interface describes public functions u obtain using this package
type TransactionalOutbox interface {
	CreateEvent(ctx context.Context, event transactional_outbox.CreateEventCommand) (int64, error)
	BatchCreateEvents(ctx context.Context, events transactional_outbox.BatchCreateEventCommand) ([]int64, error)
	RunMessageRelay(ctx context.Context)
	RunGarbageCollector(ctx context.Context)
}
```

### Example

To use this package u need to:

```go
package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	transactionalOutbox "github.com/NikitaTsaralov/transactional-outbox"
	"github.com/NikitaTsaralov/transactional-outbox/config"
	trmsqlx "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
	"github.com/jmoiron/sqlx"
	"github.com/twmb/franz-go/pkg/kgo"
)

const topic = "transactional-outbox"

func main() {
	// 1. init postgres (jmoiron/sqlx)
	db, err := sqlx.Connect(
		"pgx",
		"host=localhost port=5432 user=root dbname=postgres sslmode=disable password=dev",
	)
	if err != nil {
		log.Fatalf("cannot connect to postgres: %v", err)
	}

	defer db.Close()

	// 2. init kafka (franz-go)
	broker, err := kgo.NewClient(kgo.SeedBrokers([]string{"localhost:29092"}...))
	if err != nil {
		log.Fatalf("cannot init kafka client: %v", err)
	}

	defer broker.Close()

	// 3. prepare context (avito-tech/go-transaction-manager)
	txManager := manager.Must(trmsqlx.NewDefaultFactory(db))
	ctxGetter := trmsqlx.DefaultCtxGetter

	// 4. finally init transactional-outbox
	outbox := transactionalOutbox.NewClient(&config.Config{
		MessageRelay: config.MessageRelay{
			Timeout:   10000, // 10s
			BatchSize: 100,
		},
		GarbageCollector: config.GarbageCollector{
			Timeout: 1000,      // 10s
			TTL:     1000 * 60, // 60min
		},
	}, db, broker, txManager, ctxGetter)

	// now u can create single event
	eventID, err := outbox.CreateEvent(context.Background(), transactionalOutbox.CreateEventCommand{
		EntityID:       uuid.NewString(),
		IdempotencyKey: uuid.NewString(),
		Payload:        json.RawMessage(`{"a": "b"}`),
		Topic:          topic,
	})
	if err != nil {
		log.Fatalf("cannot create event: %v", err)
	}

	log.Printf("event with id %d created", eventID)

	// or to create multiple events at once
	eventIDs, err := outbox.BatchCreateEvents(context.Background(), transactionalOutbox.BatchCreateEventCommand{
		transactionalOutbox.CreateEventCommand{
			EntityID:       uuid.NewString(),
			IdempotencyKey: uuid.NewString(),
			Payload:        json.RawMessage(`{"c": "d"}`),
			Topic:          topic,
		},
		transactionalOutbox.CreateEventCommand{
			EntityID:       uuid.NewString(),
			IdempotencyKey: uuid.NewString(),
			Payload:        json.RawMessage(`{"e": "f"}`),
			Topic:          topic,
		},
	})
	if err != nil {
		log.Fatalf("cannot batch create event: %v", err)
	}

	log.Printf("events with ids %v created", eventIDs)

	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// to start producing messages to your broker
	go outbox.RunMessageRelay(ctxWithCancel)

	// if u are concerned about table size just use garbage collector
	go outbox.RunGarbageCollector(ctxWithCancel)

	time.Sleep(10 * time.Second)
}
```

You can follow through this example or for more understanding see tests
