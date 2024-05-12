package kafka

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/domain/entity"
	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
)

type Broker struct {
	client *kgo.Client
}

func NewBroker(client *kgo.Client) *Broker {
	return &Broker{client: client}
}

func (b *Broker) PublishEvents(ctx context.Context, events entity.Events) ([]int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Broker.Kafka.PublishEvents")
	defer span.End()

	records := lo.Map(events, func(event entity.Event, _ int) *kgo.Record {
		return &kgo.Record{
			Topic: event.Topic,
			Key:   []byte(event.IdempotencyKey),
			Value: event.Payload,
		}
	})

	produceResults := b.client.ProduceSync(ctx, records...)

	for i := range produceResults {
		if produceResults[i].Err != nil {
			return events.IDs()[:i], produceResults[i].Err
		}
	}

	return events.IDs(), nil
}
