package postgres

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/entity"
	"github.com/NikitaTsaralov/transactional-outbox/internal/domain/valueobject"
	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/storage/dto"
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
	return &Storage{
		db:        db,
		ctxGetter: ctxGetter,
	}
}

func (s *Storage) CreateEvent(ctx context.Context, event *entity.Event) error {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.CreateEvent")
	defer span.End()

	_, err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).ExecContext(
		ctx,
		queryCreateEvent,
		event.IdempotencyKey,
		event.Payload,
		event.TraceID,
		event.TraceCarrier,
		event.Processed,
		event.CreatedAt,
		event.UpdatedAt,
	)
	if err != nil {
		return trace.Wrapf(span, err, "Storage.CreateEvent.ExecContext(event: %v)", event)
	}

	return nil
}

func (s *Storage) BatchCreateEvents(ctx context.Context, events entity.Events) error {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.CreateEvent")
	defer span.End()

	dtoEventBatch := dto.NewEventBatchFromDomain(events)

	_, err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).ExecContext(
		ctx,
		queryBatchCreateEvent,
		dtoEventBatch.IdempotencyKey,
		dtoEventBatch.Payload,
		dtoEventBatch.TraceID,
		dtoEventBatch.TraceCarrier,
		dtoEventBatch.Processed,
		dtoEventBatch.CreatedAt,
		dtoEventBatch.UpdatedAt,
	)
	if err != nil {
		return trace.Wrapf(span, err, "Storage.CreateEvent.ExecContext(events: %v", events)
	}

	return nil
}

func (s *Storage) FetchUnprocessedEvents(ctx context.Context) (entity.Events, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.FetchUnprocessedEvents")
	defer span.End()

	var dtoEvents dto.Events

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).SelectContext(ctx, &dtoEvents, queryFetchUnprocessedEvents)
	if err != nil {
		return nil, trace.Wrapf(span, err, "Storage.FetchUnprocessedEvents.SelectContext()")
	}

	return dtoEvents.ToDomain(), err
}

func (s *Storage) MarkEventsAsProcessed(ctx context.Context, ids valueobject.IDs) error {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.MarkEventsAsProcessed")
	defer span.End()

	_, err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).ExecContext(ctx, queryMarkEventsAsProcessed, ids)
	if err != nil {
		return trace.Wrapf(span, err, "Storage.MarkEventsAsProcessed.ExecContext(ids: %v)", ids)
	}

	return nil
}
