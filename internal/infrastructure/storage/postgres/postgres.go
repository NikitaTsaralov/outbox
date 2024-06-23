package postgres

import (
	"context"

	dto2 "github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/storage/postgres/dto"
	"github.com/NikitaTsaralov/transactional-outbox/internal/interfaces"
	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
	"github.com/NikitaTsaralov/utils/utils/trace"
	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"
)

var _ interfaces.OutboxStorage = &Storage{}

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

	dtoEvent := dto2.NewEventFromModel(event)

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).GetContext(
		ctx,
		&id,
		queryCreateEvent,
		dtoEvent.EntityID,
		dtoEvent.IdempotencyKey,
		dtoEvent.Topic,
		dtoEvent.Payload,
		dtoEvent.TraceCarrier,
		dtoEvent.TTL,
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

	dtoBatch := dto2.NewEventBatchFromEvents(events)

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).SelectContext(
		ctx,
		&ids,
		queryBatchCreateEvent,
		dtoBatch.EntityIDs,
		dtoBatch.IdempotencyKeys,
		dtoBatch.Topics,
		dtoBatch.Payloads,
		dtoBatch.TraceCarriers,
		dtoBatch.TTL,
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

	var events dto2.Events

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).SelectContext(ctx, &events, queryFetchUnprocessedEvents, batchSize)
	if err != nil {
		return nil, trace.Wrapf(
			span,
			err,
			"Storage.Postgres.FetchUnprocessedEvents.SelectContext()",
		)
	}

	// pg advisory lock implementation
	events = lo.Filter(events, func(item dto2.Event, _ int) bool {
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

func (s *Storage) DeleteProcessedEventsByTTL(ctx context.Context) error {
	ctx, span := otel.Tracer("").Start(ctx, "Storage.Postgres.DeleteProcessedEventsByTTL")
	defer span.End()

	_, err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).ExecContext(ctx, queryDeleteProcessedEvents)
	if err != nil {
		return trace.Wrapf(
			span,
			err,
			"Storage.Postgres.DeleteProcessedEventsByTTL.ExecContext()",
		)
	}

	return err
}
