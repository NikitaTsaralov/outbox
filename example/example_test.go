package example

import (
	"context"
	"testing"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/config"
	"github.com/NikitaTsaralov/transactional-outbox/internal/application"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event/batch_create"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event/create"
	"github.com/NikitaTsaralov/utils/connectors/kafka"
	"github.com/NikitaTsaralov/utils/connectors/kafka/opts/compression"
	"github.com/NikitaTsaralov/utils/connectors/kafka/producer"
	"github.com/NikitaTsaralov/utils/connectors/postgres"
	trmsqlx "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
)

type OutboxTestSuite struct {
	suite.Suite
	service       *application.Service
	txManager     *manager.Manager
	db            *sqlx.DB
	kafkaProducer *producer.Producer
}

func (s *OutboxTestSuite) SetupSuite() {
	cfg := &config.Config{
		Metrics: &config.Metrics{Namespace: "test"},
		MessageRelay: &config.MessageRelay{
			Enabled:   true,
			Timeout:   1000, // 1 s
			BatchSize: 10,
		},
	}

	registry := prometheus.NewRegistry()

	s.db = postgres.New(postgres.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "root",
		Password: "dev",
		DBName:   "postgres",
		SSLMode:  "disable",
		Driver:   "pgx",
		Settings: postgres.Settings{
			MaxOpenConns:    60,
			ConnMaxLifetime: 120,
			MaxIdleConns:    30,
			ConnMaxIdleTime: 20,
		},
	})

	s.kafkaProducer = producer.NewProducer(producer.ProducerConfig{
		CommonConfig: kafka.CommonConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "test",
			SASL:    kafka.SASL{},
			TLS:     kafka.TLS{},
			Metrics: kafka.Metrics{
				Namespace: "test",
			},
			Timeout: kafka.Timeout{
				Dial:               10 * time.Second,
				ConnIdle:           10 * time.Second,
				RequestOverhead:    10 * time.Second,
				Rebalance:          10 * time.Second,
				Retry:              10 * time.Second,
				Session:            10 * time.Second,
				ProduceRequest:     10 * time.Second,
				RecordDelivery:     10 * time.Second,
				TransactionTimeout: 10 * time.Second,
			},
		},
		ProducerPartitioner: 0,
		RequireAcks:         0,
		Compression:         []compression.CompressionType{1},
		RecordRetries:       5,
	}, registry)

	s.txManager = manager.Must(trmsqlx.NewDefaultFactory(s.db))
	ctxGetter := trmsqlx.DefaultCtxGetter

	s.service = application.NewService(cfg, registry, s.db, s.kafkaProducer.Client, ctxGetter, s.txManager)
}

func (s *OutboxTestSuite) TestCreateEvent_Success() {
	err := s.txManager.Do(context.Background(), func(ctx context.Context) error {
		payload := []byte("{\"TestCreateEvent_Success\": \"TestCreateEvent_Success\"}")

		err := s.service.CreateEvent(ctx, &create.Command{
			IdempotencyKey: uuid.NewString(),
			Payload:        payload,
		})
		if err != nil {
			return err
		}

		_, err = s.db.Exec("insert into test (payload, err) values ($1, $2)", payload, false)
		if err != nil {
			return err
		}

		return nil
	})
	s.Require().Nil(err)
}

// TestCreateEvent_Fail checks transaction rollback works like it supposed to
func (s *OutboxTestSuite) TestCreateEvent_Fail() {
	beforeCount := 0

	err := s.db.Get(&beforeCount, `select count(*) from outbox`)
	s.Require().Nil(err)

	err = s.txManager.Do(context.Background(), func(ctx context.Context) error {
		payload := []byte("{\"TestCreateEvent_Fail\": \"TestCreateEvent_Fail\"}")

		err := s.service.CreateEvent(ctx, &create.Command{
			IdempotencyKey: uuid.NewString(),
			Payload:        payload,
		})
		if err != nil {
			return err
		}

		_, err = s.db.Exec(`insert into test (payload, err) values ($1, $2)`, payload, true) // trigger error
		if err != nil {
			return err
		}

		return nil
	})
	s.Require().NotNil(err)

	afterCount := 0

	err = s.db.Get(&afterCount, `select count(*) from outbox`)
	s.Require().Nil(err)
	s.Require().Equal(beforeCount, afterCount) // check no new events were created
}

func (s *OutboxTestSuite) TestBatchCreateEvent_Success() {
	err := s.service.BatchCreateEvent(context.Background(), &batch_create.Command{
		Events: []*batch_create.EventPayload{
			{
				IdempotencyKey: uuid.NewString(),
				Payload:        []byte("{\"TestBatchCreateEvent_Success\": \"1\"}"),
			},
			{
				IdempotencyKey: uuid.NewString(),
				Payload:        []byte("{\"TestBatchCreateEvent_Success\": \"2\"}"),
			},
		},
	})
	s.Require().Nil(err)
}

// TestBatchCreateEvent_Fail checks transaction rollback works like it supposed to
func (s *OutboxTestSuite) TestBatchCreateEvent_Fail() {
	beforeCount := 0

	err := s.db.Get(&beforeCount, `select count(*) from outbox`)
	s.Require().Nil(err)

	err = s.txManager.Do(context.Background(), func(ctx context.Context) error {
		payload := []byte("{\"TestBatchCreateEvent_Fail\": \"TestBatchCreateEvent_Fail\"}")

		err := s.service.BatchCreateEvent(ctx, &batch_create.Command{
			Events: []*batch_create.EventPayload{
				{
					IdempotencyKey: uuid.NewString(),
					Payload:        payload,
				},
				{
					IdempotencyKey: uuid.NewString(),
					Payload:        payload,
				},
			},
		})
		if err != nil {
			return err
		}

		_, err = s.db.Exec(`insert into test (payload, err) values ($1, $2)`, payload, true) // trigger error
		if err != nil {
			return err
		}

		return nil
	})
	s.Require().NotNil(err)

	afterCount := 0

	err = s.db.Get(&afterCount, `select count(*) from outbox`)
	s.Require().Nil(err)
	s.Require().Equal(beforeCount, afterCount) // check no new events were created
}

func (s *OutboxTestSuite) TestMessageRelay() {
	s.service.Start()

	time.Sleep(2 * time.Second)

	s.service.Stop()
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(OutboxTestSuite))
}
