package outbox

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/utils/logger"
	"github.com/NikitaTsaralov/utils/utils/trace"
	"go.opentelemetry.io/otel"
)

func (o Client) RunMessageRelay(ctx context.Context) {
	for {
		select {
		case <-time.NewTicker(o.cfg.MessageRelay.Timeout * time.Millisecond).C:
			func() {
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
						return err
					}

					if len(events) == 0 {
						return nil
					} // skip

					// 2. send events to the message broker
					sentEventIDs, err := o.broker.PublishEvents(localCtx, events)
					if err != nil {
						// TODO: log sentry
						logger.Error(trace.Wrapf(
							span,
							err,
							"TransactionalOutbox.RunMessageRelay.Tick.broker.PublishEvents()",
						))
					}

					if len(sentEventIDs) == 0 {
						return nil
					} // skip

					// 3. mark events as processed
					err = o.storage.MarkEventsAsProcessed(localCtx, sentEventIDs)
					if err != nil {
						return err
					}

					return nil
				})
				if err != nil {
					// TODO: log sentry
					logger.Error(trace.Wrapf(span, err, "TransactionalOutbox.RunMessageRelay.Tick.txManager.Do()"))
				}
			}()
		case <-ctx.Done():
			return
		}
	}
}
