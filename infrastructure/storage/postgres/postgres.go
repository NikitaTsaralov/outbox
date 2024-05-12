package postgres

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/domain/entity"
	"github.com/NikitaTsaralov/utils/utils/trace"
	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

type Storage struct {
	db        *sqlx.DB
	ctxGetter *txManager.CtxGetter
}

func NewStorage(db *sqlx.DB, ctxGetter *txManager.CtxGetter) *Storage {
	return &Storage{db: db, ctxGetter: ctxGetter}
}

func (s *Storage) CreateOutboxEvent(ctx context.Context, event entity.CreateEventCommand) (int64, error) {
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
	)
	if err != nil {
		return id, trace.Wrapf(
			span,
			err,
			"Storage.Postgres.CreateOutboxEvent.GetContext(entityID: %s)",
			event.EntityID.String(),
		)
	}

	return id, err
}

func (s *Storage) BatchCreateOutboxEvents(
	ctx context.Context,
	events entity.BatchCreateEventCommand,
) ([]int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.BatchCreateOutboxEvent")
	defer span.End()

	var ids []int64

	batch := entity.NewEventBatchFromCommand(events)

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).GetContext(
		ctx,
		&ids,
		queryBatchCreateEvent,
		batch.EntityIDs,
		batch.IdempotencyKeys,
		batch.Payloads,
	)
	if err != nil {
		return ids, trace.Wrapf(
			span,
			err,
			"Storage.Postgres.BatchCreateOutboxEvent.GetContext(entityIDs: %v)",
			events.EntityIDs(),
		)
	}

	return ids, err
}

func (s *Storage) FetchUnprocessedEvents(ctx context.Context) (entity.Events, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.FetchUnprocessedEvents")
	defer span.End()

	var events entity.Events

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).SelectContext(ctx, &events, queryFetchUnprocessedEvents)
	if err != nil {
		return nil, trace.Wrapf(
			span,
			err,
			"Storage.Postgres.FetchUnprocessedEvents.SelectContext()",
		)
	}

	return events, err
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
