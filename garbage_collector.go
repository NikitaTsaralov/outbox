package outbox

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/utils/logger"
	"github.com/NikitaTsaralov/utils/utils/trace"
	"go.opentelemetry.io/otel"
)

func (o Client) RunGarbageCollector(ctx context.Context) {
	for {
		select {
		case <-time.NewTicker(o.cfg.GarbageCollector.Timeout * time.Millisecond).C:
			func() {
				localCtx, span := otel.Tracer("").Start(
					context.Background(),
					"TransactionalOutbox.GarbageCollector.Tick",
				)
				defer span.End()

				err := o.storage.DeleteProcessedEvents(localCtx)
				if err != nil {
					// TODO: log sentry
					logger.Error(
						trace.Wrapf(span, err, "TransactionalOutbox.GarbageCollector.Tick.txManager.Do()"),
					)
				}
			}()
		case <-ctx.Done():
			return
		}
	}
}
