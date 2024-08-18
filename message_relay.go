package outbox

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/utils/logger"
	"github.com/NikitaTsaralov/utils/utils/trace"
	"go.opentelemetry.io/otel"
)

type MessageRelayConfig struct {
	Timeout   time.Duration `validate:"required"` // ms
	BatchSize int           `validate:"required"` // number of messages
}

func (o Outbox) RunMessageRelay(ctx context.Context) {
	for {
		select {
		case <-time.NewTicker(o.cfg.MessageRelay.Timeout * time.Millisecond).C:
			o.messageRelayTick()
		case <-ctx.Done():
			return
		}
	}
}

func (o Outbox) messageRelayTick() {
	localCtx, span := otel.Tracer("").Start(
		context.Background(),
		"TransactionalOutbox.RunMessageRelay.Tick",
	)
	defer span.End()

	// wrap in transaction
	err := o.txManager.Do(localCtx, func(localCtx context.Context) error {
		// 1. fetch unprocessed events
		events, err := o.storage.FetchUnprocessedEvents(localCtx, o.cfg.MessageRelay.BatchSize)
		if err != nil {
			logger.Errorf("can't fetch unprocessed events: %v", err)
			return err
		}

		if len(events) == 0 {
			return nil
		} // skip

		// 2. send events to the message broker
		sentEventIDs, err := o.broker.PublishEvents(localCtx, events)
		if err != nil {
			logger.Errorf("can't fetch unprocessed events: %v", err)
			return err // message duplication is possible but it need to be handled by the consumer
		}

		if len(sentEventIDs) == 0 {
			logger.Info("no events to send")
			return nil
		} // skip

		// 3. mark events as processed
		err = o.storage.MarkEventsAsProcessed(localCtx, sentEventIDs)
		if err != nil {
			logger.Errorf("can't mark events as processed: %v", err)
			return err
		}

		return nil
	})
	if err != nil {
		// TODO: log sentry
		logger.Errorf(
			"send events transaction aborted: %v",
			trace.Wrapf(span, err, "TransactionalOutbox.MessageRelay.Tick.txManager.Do()"),
		)
	}
}
