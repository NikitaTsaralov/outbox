package postgres

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/models/event"
	"github.com/NikitaTsaralov/utils/utils/trace"
	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"
)

type Storage struct {
	db        *sqlx.DB
	ctxGetter *txManager.CtxGetter
}

func NewStorage(db *sqlx.DB, ctxGetter *txManager.CtxGetter) *Storage {
	return &Storage{db: db, ctxGetter: ctxGetter}
}

func (s *Storage) CreateOutboxEvent(ctx context.Context, event event.Event) (int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.CreateOutboxEvent")
	defer span.End()

	var id int64

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).GetContext(
		ctx,
		&id,
		queryCreateEvent,
		event.EntityID,
		event.IdempotencyKey,
		event.Topic,
		event.Payload,
		event.TraceCarrier,
	)
	if err != nil {
		return id, trace.Wrapf(
			span,
			err,
			"Storage.Postgres.CreateOutboxEvent.GetContext(entityID: %s)",
			event.EntityID,
		)
	}

	return id, err
}

func (s *Storage) BatchCreateOutboxEvents(
	ctx context.Context,
	events event.Events,
) ([]int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.BatchCreateOutboxEvent")
	defer span.End()

	var ids []int64

	batch := event.NewEventBatchFromEvents(events)

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).SelectContext(
		ctx,
		&ids,
		queryBatchCreateEvent,
		batch.EntityIDs,
		batch.IdempotencyKeys,
		batch.Topics,
		batch.Payloads,
		batch.TraceCarriers,
	)
	if err != nil {
		return ids, trace.Wrapf(
			span,
			err,
			"Storage.Postgres.BatchCreateOutboxEvent.ExecContext(entityIDs: %v)",
			events.EntityIDs(),
		)
	}

	return ids, err
}

func (s *Storage) FetchUnprocessedEvents(ctx context.Context, batchSize int) (event.Events, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.FetchUnprocessedEvents")
	defer span.End()

	var events event.Events

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).SelectContext(ctx, &events, queryFetchUnprocessedEvents, batchSize)
	if err != nil {
		return nil, trace.Wrapf(
			span,
			err,
			"Storage.Postgres.FetchUnprocessedEvents.SelectContext()",
		)
	}

	return lo.Filter(events, func(item event.Event, _ int) bool {
		return item.Available
	}), err
}

func (s *Storage) MarkEventsAsProcessed(ctx context.Context, ids []int64) error {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.MarkEventsAsProcessed")
	defer span.End()

	_, err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).ExecContext(ctx, queryMarkEventsAsProcessed, ids)
	if err != nil {
		return trace.Wrapf(
			span,
			err,
			"Storage.Postgres.MarkEventsAsProcessed.ExecContext(ids: %#v)",
			ids,
		)
	}

	return err
}

func (s *Storage) DeleteProcessedEvents(ctx context.Context, eventTTL time.Duration) error {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.DeleteProcessedEvents")
	defer span.End()

	_, err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).ExecContext(ctx, queryDeleteProcessedEvents, eventTTL)
	if err != nil {
		return trace.Wrapf(
			span,
			err,
			"Storage.Postgres.DeleteProcessedEvents.ExecContext(eventTTL: %d)",
			eventTTL,
		)
	}

	return err
}
