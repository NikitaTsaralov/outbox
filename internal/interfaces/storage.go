package interfaces

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
)

type OutboxStorage interface {
	CreateOutboxEvent(ctx context.Context, event models.Event) (int64, error)
	BatchCreateOutboxEvents(ctx context.Context, events models.Events) ([]int64, error)
	FetchUnprocessedEvents(ctx context.Context, batchSize int) (models.Events, error)
	MarkEventsAsProcessed(ctx context.Context, ids []int64) error
	DeleteProcessedEventsByTTL(ctx context.Context) error
}
