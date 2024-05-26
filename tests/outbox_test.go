package tests

import (
	"context"
	"encoding/json"
	"sort"
	"testing"
	"time"

	transactionalOutbox "github.com/NikitaTsaralov/transactional-outbox"
	"github.com/NikitaTsaralov/transactional-outbox/config"
	"github.com/NikitaTsaralov/transactional-outbox/models/event"
	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/exporters/jaeger"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

const (
	timeout   = 1000 // 1s
	eventTTL  = 1000 // 1s
	batchSize = 10
)

type OutboxTestSuite struct {
	suite.Suite
	cfg            *config.Config
	db             *sqlx.DB
	broker         *kgo.Client
	jaegerExporter *jaeger.Exporter
	traceProvider  *tracesdk.TracerProvider
	txManager      *manager.Manager
	ctxGetter      *txManager.CtxGetter
	outbox         TransactionalOutbox
}

func (s *OutboxTestSuite) SetupTest() {
	var err error

	s.cfg = &config.Config{
		MessageRelay: config.MessageRelay{
			Timeout:   timeout, // 1s
			BatchSize: batchSize,
		},
		GarbageCollector: config.GarbageCollector{
			Timeout: timeout,  // 1s
			TTL:     eventTTL, // 1s
		},
	}

	s.db, err = initPostgres()
	s.Require().Nil(err)
	s.Require().NotNil(s.db)

	s.broker, err = initKafka()
	s.Require().Nil(err)
	s.Require().NotNil(s.broker)

	s.jaegerExporter, err = initJaeger()
	s.Require().Nil(err)
	s.Require().NotNil(s.jaegerExporter)

	s.traceProvider = initTraceProvider(s.jaegerExporter)

	s.txManager = manager.Must(txManager.NewDefaultFactory(s.db))
	s.ctxGetter = txManager.DefaultCtxGetter

	s.outbox = transactionalOutbox.NewClient(s.cfg, s.db, s.broker, s.txManager, s.ctxGetter)

	_, err = s.db.Exec(queryDeleteAll)
	s.Require().Nil(err)
}

func (s *OutboxTestSuite) TearDownSuite() {
	err := s.db.Close()
	s.Require().Nil(err)

	s.broker.Close()

	err = s.traceProvider.Shutdown(context.Background())
	s.Require().Nil(err)

	err = s.jaegerExporter.Shutdown(context.Background())
	s.Require().Nil(err)
}

func (s *OutboxTestSuite) TestCreateEvent() {
	tests := []struct {
		name string
		data transactionalOutbox.CreateEventCommand
	}{
		{
			name: "test 1",
			data: transactionalOutbox.CreateEventCommand{
				EntityID:       uuid.NewString(),
				IdempotencyKey: uuid.NewString(),
				Payload:        json.RawMessage(`{"1": "2"}`),
				Topic:          "transactional-outbox",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			id, err := s.outbox.CreateEvent(context.Background(), test.data)
			s.Require().Nil(err)
			s.Require().NotNil(id)

			var createdEvent event.Events

			err = s.db.Select(&createdEvent, queryFetchEventsByIDs, pq.Int64Array([]int64{id}))
			s.Require().Nil(err)
			s.Require().Equal(len(createdEvent), 1)

			s.Require().NotNil(createdEvent[0].ID)
			s.Require().Equal(createdEvent[0].ID, id)
			s.Require().Equal(createdEvent[0].EntityID, test.data.EntityID)
			s.Require().Equal(createdEvent[0].IdempotencyKey, test.data.IdempotencyKey)
			s.Require().Equal(createdEvent[0].Payload, test.data.Payload)
			s.Require().Equal(createdEvent[0].Topic, test.data.Topic)
			s.Require().NotNil(createdEvent[0].TraceCarrier)
			s.Require().NotNil(createdEvent[0].CreatedAt)
			s.Require().False(createdEvent[0].SentAt.Valid)
		})
	}
}

func (s *OutboxTestSuite) TestBatchCreateEvent() {
	tests := []struct {
		name string
		data transactionalOutbox.BatchCreateEventCommand
	}{
		{
			name: "test 1",
			data: transactionalOutbox.BatchCreateEventCommand{
				transactionalOutbox.CreateEventCommand{
					EntityID:       uuid.NewString(),
					IdempotencyKey: uuid.NewString(),
					Payload:        json.RawMessage(`{"3": "4"}`),
					Topic:          "transactional-outbox",
				},
				transactionalOutbox.CreateEventCommand{
					EntityID:       uuid.NewString(),
					IdempotencyKey: uuid.NewString(),
					Payload:        json.RawMessage(`{"5": "6"}`),
					Topic:          "transactional-outbox",
				},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			sort.Slice(test.data, func(i, j int) bool {
				return test.data[i].EntityID < test.data[j].EntityID
			})

			ids, err := s.outbox.BatchCreateEvents(context.Background(), test.data)
			s.Require().Nil(err)
			s.Require().NotNil(ids)

			var createdEvents event.Events

			err = s.db.Select(&createdEvents, queryFetchEventsByIDs, pq.Int64Array(ids))
			s.Require().Nil(err)

			s.Require().Equal(len(ids), len(test.data))

			for i := 0; i < len(ids); i++ {
				s.Require().NotNil(createdEvents[i].ID)
				s.Require().Equal(createdEvents[i].ID, ids[i])
				s.Require().Equal(createdEvents[i].EntityID, test.data[i].EntityID)
				s.Require().Equal(createdEvents[i].IdempotencyKey, test.data[i].IdempotencyKey)
				s.Require().Equal(createdEvents[i].Payload, test.data[i].Payload)
				s.Require().Equal(createdEvents[i].Topic, test.data[i].Topic)
				s.Require().NotNil(createdEvents[i].TraceCarrier)
				s.Require().NotNil(createdEvents[i].CreatedAt)
				s.Require().False(createdEvents[i].SentAt.Valid)
			}
		})
	}
}

func (s *OutboxTestSuite) TestRunMessageRelay() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.outbox.RunMessageRelay(ctx)

	time.Sleep(5 * time.Second) // this time is enough

	var createdEvents event.Events

	err := s.db.Select(&createdEvents, queryFetchAll)
	s.Require().Nil(err)

	for _, createdEvent := range createdEvents {
		s.Require().True(createdEvent.SentAt.Valid)
		s.Require().NotNil(createdEvent.SentAt)
	}
}

func (s *OutboxTestSuite) TestRunGarbageCollector() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	time.Sleep(5 * time.Second) // skip relay

	var eventsBefore event.Events

	err := s.db.Select(&eventsBefore, queryFetchAll)
	s.Require().Nil(err)

	eventsToDeleteCount := 0

	for _, eventBefore := range eventsBefore {
		if eventBefore.SentAt.Valid && time.Since(eventBefore.SentAt.Time) > eventTTL {
			eventsToDeleteCount++
		}
	}

	go s.outbox.RunGarbageCollector(ctx)

	time.Sleep(5 * time.Second)

	var createdEvents event.Events

	err = s.db.Select(&createdEvents, queryFetchAll)
	s.Require().Nil(err)
	s.Require().Equal(len(eventsBefore)-eventsToDeleteCount, len(createdEvents))
}

func TestOutboxTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(OutboxTestSuite))
}
