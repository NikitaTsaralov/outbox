package application

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/domain/broker"
	"github.com/NikitaTsaralov/transactional-outbox/domain/entity"
	"github.com/NikitaTsaralov/transactional-outbox/domain/storage"
	"github.com/NikitaTsaralov/transactional-outbox/infrastructure/broker/kafka"
	"github.com/NikitaTsaralov/transactional-outbox/infrastructure/storage/postgres"

	"github.com/NikitaTsaralov/utils/logger"
	"github.com/NikitaTsaralov/utils/utils/trace"
	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/jmoiron/sqlx"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
)

type Client struct {
	relayTimeout time.Duration // ms
	storage      storage.OutboxRepository
	broker       broker.Broker
	txManager    *manager.Manager
	ctxGetter    *txManager.CtxGetter
}

func NewClient(
	db *sqlx.DB,
	kafkaClient *kgo.Client,
	relayTimeout time.Duration,
	txManager *manager.Manager,
	ctxGetter *txManager.CtxGetter,
) *Client {
	return &Client{
		relayTimeout: relayTimeout,
		storage:      postgres.NewStorage(db, ctxGetter),
		broker:       kafka.NewBroker(kafkaClient),
		txManager:    txManager,
		ctxGetter:    ctxGetter,
	}
}

func (o Client) CreateEvent(ctx context.Context, event entity.CreateEventCommand) (int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "OutboxApp.CreateEvent")
	defer span.End()

	return o.storage.CreateOutboxEvent(ctx, event)
}

func (o Client) BatchCreateEvents(ctx context.Context, events entity.BatchCreateEventCommand) ([]int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "OutboxApp.BatchCreateEvents")
	defer span.End()

	return o.storage.BatchCreateOutboxEvents(ctx, events)
}

func (o Client) RunMessageRelay(ctx context.Context) {
	for {
		select {
		case <-time.NewTicker(o.relayTimeout * time.Millisecond).C:
			func() {
				ctx, span := otel.Tracer("").Start(ctx, "OutboxApp.RunMessageRelay.Tick")

				err := o.txManager.Do(ctx, func(ctx context.Context) error {
					// 1. We fetch unprocessed events
					events, err := o.storage.FetchUnprocessedEvents(ctx)
					if err != nil {
						return err
					}

					if len(events) == 0 {
						return nil
					}

					// 2. We send events to the message broker
					sentEventIDs, err := o.broker.PublishEvents(ctx, events)
					if err != nil {
						// TODO: log sentry
						logger.Error(
							trace.Wrapf(span, err, "OutboxApp.RunMessageRelay.Tick.broker.PublishEvents()"),
						)
					}

					if len(sentEventIDs) == 0 {
						return nil
					}

					// 3. We mark events as processed
					err = o.storage.MarkEventsAsProcessed(ctx, sentEventIDs)
					if err != nil {
						return err
					}

					return nil
				})
				if err != nil {
					// TODO: log sentry
					logger.Error(trace.Wrapf(span, err, "OutboxApp.RunMessageRelay.Tick.txManager.Do()"))
				}
			}()
		case <-ctx.Done():
			return
		}
	}
}
