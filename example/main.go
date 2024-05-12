package main

import (
	"context"

	transactionalOutbox "github.com/NikitaTsaralov/transactional-outbox/application"
	"github.com/NikitaTsaralov/utils/connectors/kafka/opts/ack_policy"

	"github.com/NikitaTsaralov/utils/connectors/kafka"
	"github.com/NikitaTsaralov/utils/connectors/kafka/producer"
	"github.com/NikitaTsaralov/utils/connectors/postgres"
	"github.com/NikitaTsaralov/utils/logger"
	trmsqlx "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/jmoiron/sqlx"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	postgresDB, kafkaClient := connect()
	defer postgresDB.Close()
	defer kafkaClient.Close()

	client := transactionalOutbox.NewClient(
		postgresDB,
		kafkaClient,
		10000,
		manager.Must(trmsqlx.NewDefaultFactory(postgresDB)),
		trmsqlx.DefaultCtxGetter,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create event
	//id, err := client.CreateEvent(ctx, entity.CreateEventCommand{
	//	EntityID:       uuid.New(),
	//	IdempotencyKey: "test",
	//	Payload:        []byte(`{"test": "test"}`),
	//	Topic:          "test",
	//})
	//if err != nil {
	//	logger.Fatalf("failed to create event: %v", err)
	//}

	// run message relay
	client.RunMessageRelay(ctx)
}

func connect() (*sqlx.DB, *kgo.Client) {
	db, err := postgres.Start(postgres.Config{
		Host:           "localhost",
		Port:           5432,
		User:           "root",
		Password:       "dev",
		DBName:         "postgres",
		SSLModeEnabled: false,
		Driver:         "pgx",
		Settings:       postgres.Settings{},
	})
	if err != nil {
		logger.Fatalf("failed to connect to postgres: %v", err)
	}

	newProducer, err := producer.NewProducer(producer.ProducerConfig{
		CommonConfig: kafka.CommonConfig{
			Brokers: []string{"localhost:29092"},
			Topic:   "test",
		},
		RequireAcks: ack_policy.AllAck,
	}, nil)
	if err != nil {
		logger.Fatalf("failed to create producer: %v", err)
	}

	return db, newProducer.Client
}
