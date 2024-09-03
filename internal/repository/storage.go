package repository

import (
	"context"

	"github.com/NikitaTsaralov/outbox/internal/models"
)

type OutboxStorage interface {
	CreateOutboxEvent(ctx context.Context, event models.Event) (int64, error)
	BatchCreateOutboxEvents(ctx context.Context, events models.Events) ([]int64, error)

	// 4 message relay

	FetchUnprocessedEvents(ctx context.Context, batchSize int) (models.Events, error)
	MarkEventsAsProcessed(ctx context.Context, ids []int64) error

	// 4 garbage collector

	CountExpiredEvents(ctx context.Context) (int, error)
	FetchExpiredEvents(ctx context.Context, batchSize int) (models.Events, error)
	DeleteEventsByIDs(ctx context.Context, ids []int64) error
}
