package outbox

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/config"
	"github.com/NikitaTsaralov/transactional-outbox/infrastructure/broker/kafka"
	"github.com/NikitaTsaralov/transactional-outbox/infrastructure/storage/postgres"
	"github.com/NikitaTsaralov/transactional-outbox/interfaces"
	"github.com/NikitaTsaralov/transactional-outbox/models/event"
	traceCarrier "github.com/NikitaTsaralov/transactional-outbox/pkg/tracecarrier"
	"github.com/NikitaTsaralov/utils/logger"
	"github.com/NikitaTsaralov/utils/utils/trace"
	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
)

type Client struct {
	cfg       *config.Config
	storage   interfaces.OutboxStorage
	broker    interfaces.Broker
	txManager *manager.Manager
	ctxGetter *txManager.CtxGetter
}

func NewClient(
	cfg *config.Config,
	db *sqlx.DB,
	kafkaClient *kgo.Client,
	txManager *manager.Manager,
	ctxGetter *txManager.CtxGetter,
) *Client {
	return &Client{
		cfg:       cfg,
		storage:   postgres.NewStorage(db, ctxGetter),
		broker:    kafka.NewBroker(kafkaClient),
		txManager: txManager,
		ctxGetter: ctxGetter,
	}
}

func (o Client) CreateEvent(ctx context.Context, command CreateEventCommand) (int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "TransactionalOutbox.CreateEvent")
	defer span.End()

	return o.storage.CreateOutboxEvent(ctx, event.Event{
		EntityID:       command.EntityID,
		IdempotencyKey: command.IdempotencyKey,
		Topic:          command.Topic,
		Payload:        command.Payload,
		TraceCarrier:   traceCarrier.NewTraceCarrierFromContext(ctx),
	})
}

func (o Client) BatchCreateEvents(ctx context.Context, command BatchCreateEventCommand) ([]int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "TransactionalOutbox.BatchCreateEvents")
	defer span.End()

	return o.storage.BatchCreateOutboxEvents(ctx, lo.Map(command, func(item CreateEventCommand, _ int) event.Event {
		return event.Event{
			EntityID:       item.EntityID,
			IdempotencyKey: item.IdempotencyKey,
			Topic:          item.Topic,
			Payload:        item.Payload,
			TraceCarrier:   traceCarrier.NewTraceCarrierFromContext(ctx),
		}
	}))
}

func (o Client) RunMessageRelay(ctx context.Context) {
	for {
		select {
		case <-time.NewTicker(o.cfg.MessageRelay.Timeout * time.Millisecond).C:
			func() {
				localCtx, span := otel.Tracer("").Start(
					context.Background(),
					"TransactionalOutbox.RunMessageRelay.Tick",
				)
				defer span.End()

				err := o.txManager.Do(localCtx, func(localCtx context.Context) error {
					// 1. We fetch unprocessed events
					events, err := o.storage.FetchUnprocessedEvents(localCtx, o.cfg.MessageRelay.BatchSize)
					if err != nil {
						return err
					}

					if len(events) == 0 {
						return nil
					} // skip

					// 2. We send events to the message broker
					sentEventIDs, err := o.broker.PublishEvents(localCtx, events)
					if err != nil {
						// TODO: log sentry
						logger.Error(
							trace.Wrapf(span, err, "TransactionalOutbox.RunMessageRelay.Tick.broker.PublishEvents()"),
						)
					}

					if len(sentEventIDs) == 0 {
						return nil
					} // skip

					// 3. We mark events as processed
					err = o.storage.MarkEventsAsProcessed(localCtx, sentEventIDs)
					if err != nil {
						return err
					}

					return nil
				})
				if err != nil {
					// TODO: log sentry
					logger.Error(trace.Wrapf(span, err, "TransactionalOutbox.RunMessageRelay.Tick.txManager.Do()"))
				}
			}()
		case <-ctx.Done():
			return
		}
	}
}

func (o Client) RunGarbageCollector(ctx context.Context) {
	for {
		select {
		case <-time.NewTicker(o.cfg.GarbageCollector.Timeout * time.Millisecond).C:
			func() {
				localCtx, span := otel.Tracer("").Start(
					context.Background(),
					"TransactionalOutbox.GarbageCollector.Tick",
				)
				defer span.End()

				err := o.storage.DeleteProcessedEvents(localCtx, o.cfg.GarbageCollector.TTL)
				if err != nil {
					// TODO: log sentry
					logger.Error(
						trace.Wrapf(span, err, "TransactionalOutbox.GarbageCollector.Tick.txManager.Do()"),
					)
				}
			}()
		case <-ctx.Done():
			return
		}
	}
}
