package outbox

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/config"
	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/broker/kafka"
	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/storage/postgres"
	interfaces2 "github.com/NikitaTsaralov/transactional-outbox/internal/interfaces"
	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
	"github.com/NikitaTsaralov/utils/logger"
	"github.com/NikitaTsaralov/utils/utils/trace"
	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
)

type Outbox struct {
	cfg       *config.Config
	storage   interfaces2.OutboxStorage
	broker    interfaces2.Broker
	txManager *manager.Manager
	ctxGetter *txManager.CtxGetter
}

func NewOutbox(
	cfg *config.Config,
	db *sqlx.DB,
	kafkaClient *kgo.Client,
	txManager *manager.Manager,
	ctxGetter *txManager.CtxGetter,
) *Outbox {
	return &Outbox{
		cfg:       cfg,
		storage:   postgres.NewStorage(db, ctxGetter),
		broker:    kafka.NewBroker(kafkaClient),
		txManager: txManager,
		ctxGetter: ctxGetter,
	}
}

func (o Outbox) CreateEvent(ctx context.Context, command CreateEventCommand) (int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "TransactionalOutbox.CreateEvent")
	defer span.End()

	return o.storage.CreateOutboxEvent(ctx, models.Event{
		EntityID:       command.EntityID,
		IdempotencyKey: command.IdempotencyKey,
		Topic:          command.Topic,
		Payload:        command.Payload,
		Context:        ctx,
		TTL:            command.TTL,
	})
}

func (o Outbox) BatchCreateEvents(ctx context.Context, command BatchCreateEventCommand) ([]int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "TransactionalOutbox.BatchCreateEvents")
	defer span.End()

	return o.storage.BatchCreateOutboxEvents(ctx, lo.Map(command, func(item CreateEventCommand, _ int) models.Event {
		return models.Event{
			EntityID:       item.EntityID,
			IdempotencyKey: item.IdempotencyKey,
			Topic:          item.Topic,
			Payload:        item.Payload,
			Context:        ctx,
			TTL:            item.TTL,
		}
	}))
}

func (o Outbox) RunMessageRelay(ctx context.Context) {
	for {
		select {
		case <-time.NewTicker(o.cfg.MessageRelay.Timeout * time.Millisecond).C:
			func() {
				localCtx, span := otel.Tracer("").Start(
					context.Background(),
					"TransactionalOutbox.RunMessageRelay.Tick",
				)
				defer span.End()

				// wrap in transaction
				err := o.txManager.Do(localCtx, func(localCtx context.Context) error {
					// 1. fetch unprocessed events
					events, err := o.storage.FetchUnprocessedEvents(localCtx, o.cfg.MessageRelay.BatchSize)
					if err != nil {
						return err
					}

					if len(events) == 0 {
						return nil
					} // skip

					// 2. send events to the message broker
					sentEventIDs, err := o.broker.PublishEvents(localCtx, events)
					if err != nil {
						// TODO: log sentry
						logger.Error(trace.Wrapf(
							span,
							err,
							"TransactionalOutbox.RunMessageRelay.Tick.broker.PublishEvents()",
						))
					}

					if len(sentEventIDs) == 0 {
						return nil
					} // skip

					// 3. mark events as processed
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

func (o Outbox) RunGarbageCollector(ctx context.Context) {
	for {
		select {
		case <-time.NewTicker(o.cfg.GarbageCollector.Timeout * time.Millisecond).C:
			func() {
				localCtx, span := otel.Tracer("").Start(
					context.Background(),
					"TransactionalOutbox.GarbageCollector.Tick",
				)
				defer span.End()

				err := o.storage.DeleteProcessedEventsByTTL(localCtx)
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
