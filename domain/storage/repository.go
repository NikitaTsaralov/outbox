package storage

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/domain/entity"
)

type OutboxRepository interface {
	CreateOutboxEvent(ctx context.Context, command entity.CreateEventCommand) (int64, error)
	BatchCreateOutboxEvents(ctx context.Context, command entity.BatchCreateEventCommand) ([]int64, error)
	FetchUnprocessedEvents(ctx context.Context) (entity.Events, error)
	MarkEventsAsProcessed(ctx context.Context, ids []int64) error
}
