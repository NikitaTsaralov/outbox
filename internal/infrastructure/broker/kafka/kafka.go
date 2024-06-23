package kafka

import (
	"context"
	"errors"

	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/broker/kafka/dto"
	"github.com/NikitaTsaralov/transactional-outbox/internal/models"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
)

type Broker struct {
	client *kgo.Client
}

func NewBroker(client *kgo.Client) *Broker {
	return &Broker{client: client}
}

func (b *Broker) PublishEvents(ctx context.Context, events models.Events) ([]int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Broker.Kafka.PublishEvents")
	defer span.End()

	var err error

	ids := make([]int64, len(events))
	produceResults := b.client.ProduceSync(ctx, dto.NewRecordsFromModel(events)...)

	for i, res := range produceResults {
		if res.Err != nil {
			err = errors.Join(res.Err, err) // Accumulate error.

			continue
		}

		ids[i] = events[i].ID
	}

	if err != nil {
		return ids[:len(produceResults)], err
	}

	return ids[:len(produceResults)], nil
}
