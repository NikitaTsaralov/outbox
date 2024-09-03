package postgres

import (
	"context"

	"github.com/NikitaTsaralov/outbox/internal/infrastructure/repository/postgres/dto"
	"github.com/NikitaTsaralov/outbox/internal/models"
	"github.com/NikitaTsaralov/outbox/internal/repository"
	"github.com/NikitaTsaralov/utils/utils/trace"
	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"
)

var _ repository.OutboxStorage = &Storage{}

type Storage struct {
	db        *sqlx.DB
	ctxGetter *txManager.CtxGetter
}

func NewStorage(db *sqlx.DB, ctxGetter *txManager.CtxGetter) *Storage {
	return &Storage{db: db, ctxGetter: ctxGetter}
}

func (s *Storage) CreateOutboxEvent(ctx context.Context, event models.Event) (int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.CreateOutboxEvent")
	defer span.End()

	var id int64

	dtoEvent := dto.NewEventFromModel(event)

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).GetContext(
		ctx,
		&id,
		queryCreateEvent,
		dtoEvent.EntityID,
		dtoEvent.IdempotencyKey,
		dtoEvent.Topic,
		dtoEvent.Key,
		dtoEvent.Payload,
		dtoEvent.TraceCarrier,
		dtoEvent.CreatedAt,
		dtoEvent.ExpiresAt,
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
	events models.Events,
) ([]int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.BatchCreateOutboxEvent")
	defer span.End()

	var ids []int64

	dtoBatch := dto.NewEventBatchFromEvents(events)

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).SelectContext(
		ctx,
		&ids,
		queryBatchCreateEvent,
		dtoBatch.EntityIDs,
		dtoBatch.IdempotencyKeys,
		dtoBatch.Topics,
		dtoBatch.Keys,
		dtoBatch.Payloads,
		dtoBatch.TraceCarriers,
		dtoBatch.CreatedAts,
		dtoBatch.ExpiresAts,
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

func (s *Storage) FetchUnprocessedEvents(ctx context.Context, batchSize int) (models.Events, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.FetchUnprocessedEvents")
	defer span.End()

	var events dto.Events

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).SelectContext(ctx, &events, queryFetchUnprocessedEvents, batchSize)
	if err != nil {
		return nil, trace.Wrapf(
			span,
			err,
			"Storage.Postgres.FetchUnprocessedEvents.SelectContext()",
		)
	}

	// pg advisory lock implementation
	events = lo.Filter(events, func(item dto.Event, _ int) bool {
		return item.Available
	})

	return events.ToModel(), err
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

func (s *Storage) CountExpiredEvents(ctx context.Context) (int, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.CountExpiredEvents")
	defer span.End()

	var count int

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).GetContext(ctx, &count, queryCountExpiredEvents)
	if err != nil {
		return count, trace.Wrapf(
			span,
			err,
			"Storage.Postgres.CountExpiredEvents.GetContext()",
		)
	}

	return count, err
}

func (s *Storage) FetchExpiredEvents(ctx context.Context, batchSize int) (models.Events, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.FetchExpiredEvents")
	defer span.End()

	var events dto.Events

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).SelectContext(ctx, &events, queryFetchExpiredEvents, batchSize)
	if err != nil {
		return nil, trace.Wrapf(
			span,
			err,
			"Storage.Postgres.FetchExpiredEvents.SelectContext(batchSize: %d)",
			batchSize,
		)
	}

	return events.ToModel(), err
}

func (s *Storage) DeleteEventsByIDs(ctx context.Context, ids []int64) error {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.DeleteEventsByIDs")
	defer span.End()

	_, err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).ExecContext(ctx, queryDeleteEventsByIDs, pq.Int64Array(ids))
	if err != nil {
		return trace.Wrapf(
			span,
			err,
			"Storage.Postgres.DeleteEventsByIDs.ExecContext()",
		)
	}

	return nil
}
