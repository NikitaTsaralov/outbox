package interfaces

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/models/event"
)

type OutboxStorage interface {
	CreateOutboxEvent(ctx context.Context, event event.Event) (int64, error)
	BatchCreateOutboxEvents(ctx context.Context, events event.Events) ([]int64, error)
	FetchUnprocessedEvents(ctx context.Context, batchSize int) (event.Events, error)
	MarkEventsAsProcessed(ctx context.Context, ids []int64) error
	DeleteProcessedEvents(ctx context.Context, ttl time.Duration) error
}
