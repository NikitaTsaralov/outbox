package broker

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
)

type Broker interface {
	PublishEvents(ctx context.Context, events models.Events) ([]int64, error)
}
