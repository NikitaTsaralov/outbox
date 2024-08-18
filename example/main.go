package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	transactionalOutbox "github.com/NikitaTsaralov/transactional-outbox"
	"github.com/NikitaTsaralov/utils/logger"
	trmsqlx "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
	"github.com/jmoiron/sqlx"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	topic     = "transactional-outbox"
	timeout   = 10000 // 10s
	batchSize = 100
	eventTTL  = 60000 // 1m
)

func initPostgres() *sqlx.DB {
	db, err := sqlx.Connect(
		"pgx",
		"host=localhost port=5432 user=root dbname=postgres sslmode=disable password=dev",
	)
	if err != nil {
		log.Fatalf("cannot connect to postgres: %v", err)
	}

	return db
}

func initKafka() *kgo.Client {
	client, err := kgo.NewClient(kgo.SeedBrokers([]string{"localhost:29092"}...))
	if err != nil {
		log.Fatalf("cannot init kafka client: %v", err)
	}

	return client
}

func main() {
	// 1. init postgres (jmoiron/sqlx)
	db := initPostgres()
	defer db.Close()

	// 2. init kafka (franz-go)
	broker := initKafka()
	defer broker.Close()

	// 3. prepare context (avito-tech/go-transaction-manager)
	txManager := manager.Must(trmsqlx.NewDefaultFactory(db))
	ctxGetter := trmsqlx.DefaultCtxGetter

	// 4. finally init transactional-outbox
	outbox := transactionalOutbox.NewOutbox(&transactionalOutbox.Config{
		MessageRelay: transactionalOutbox.MessageRelayConfig{
			Timeout:   timeout,
			BatchSize: batchSize,
		},
		GarbageCollector: transactionalOutbox.GarbageCollectorConfig{
			Timeout:   timeout,
			BatchSize: batchSize,
		},
	}, db, broker, txManager, ctxGetter)

	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	// now u can create single event
	_, err := outbox.CreateEventCommandHandler(ctxWithCancel, transactionalOutbox.CreateEventCommand{
		EntityID:       uuid.NewString(),
		IdempotencyKey: uuid.NewString(),
		Payload:        json.RawMessage(`{"a": "b"}`),
		Topic:          topic,
		TTL:            time.Second * 10,
	})
	if err != nil {
		logger.Error("cannot create event: %v", err)
		return
	}

	// or to create multiple events at once
	_, err = outbox.BatchCreateEvents(ctxWithCancel, transactionalOutbox.BatchCreateEventsCommand{
		transactionalOutbox.CreateEventCommand{
			EntityID:       uuid.NewString(),
			IdempotencyKey: uuid.NewString(),
			Payload:        json.RawMessage(`{"c": "d"}`),
			Topic:          topic,
			TTL:            time.Second * 10,
		},
		transactionalOutbox.CreateEventCommand{
			EntityID:       uuid.NewString(),
			IdempotencyKey: uuid.NewString(),
			Payload:        json.RawMessage(`{"e": "f"}`),
			Topic:          topic,
			TTL:            time.Second * 10,
		},
	})
	if err != nil {
		logger.Error("cannot batch create event: %v", err)
		return
	}

	go outbox.RunMessageRelay(ctxWithCancel)     // to start producing messages to your broker
	go outbox.RunGarbageCollector(ctxWithCancel) // if u are concerned about table size just use garbage collector
	time.Sleep(timeout * time.Millisecond)
}
