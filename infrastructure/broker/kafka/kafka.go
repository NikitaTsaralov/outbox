package kafka

import (
	"context"
	"errors"

	"github.com/NikitaTsaralov/transactional-outbox/models/event"
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

func (b *Broker) PublishEvents(ctx context.Context, events event.Events) ([]int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Broker.Kafka.PublishEvents")
	defer span.End()

	var err error
	ids := make([]int64, len(events))

	records := lo.Map(events, func(event event.Event, _ int) *kgo.Record {
		localCtx, msgSpan := otel.Tracer("").Start(
			event.TraceCarrier.Context(), // restore context from carrier
			"Broker.Kafka.PublishEvents",
		)
		defer msgSpan.End()

		return &kgo.Record{
			Topic:   event.Topic,
			Key:     []byte(event.IdempotencyKey),
			Value:   event.Payload,
			Context: localCtx,
		}
	})

	produceResults := b.client.ProduceSync(ctx, records...)

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
