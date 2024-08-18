package outbox

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/utils/logger"
	"github.com/NikitaTsaralov/utils/utils/trace"
	"go.opentelemetry.io/otel"
)

type GarbageCollectorConfig struct {
	Timeout   time.Duration `validate:"required"` // ms
	BatchSize int           `validate:"required"`
	Cooldown  time.Duration `validate:"required"` // time between deletions
}

func (o Outbox) RunGarbageCollector(ctx context.Context) {
	for {
		select {
		case <-time.NewTicker(o.cfg.GarbageCollector.Timeout * time.Millisecond).C:
			o.garbageCollectorTick()
		case <-ctx.Done():
			return
		}
	}
}

func (o Outbox) garbageCollectorTick() {
	ctx, span := otel.Tracer("").Start(
		context.Background(),
		"TransactionalOutbox.GarbageCollectorConfig.Tick",
	)
	defer span.End()

	// count events to delete
	events, err := o.storage.CountExpiredEvents(ctx)
	if err != nil {
		logger.Errorf("can't count expired events: %v", err)
		return
	}

	// iterate over event batches and delete them
	for i := 0; i < events; i += o.cfg.GarbageCollector.BatchSize {
		err := o.txManager.Do(ctx, func(ctx context.Context) error {
			// fetch expired events
			expiredEvents, err := o.storage.FetchExpiredEvents(ctx, o.cfg.GarbageCollector.BatchSize)
			if err != nil {
				logger.Errorf("can't fetch expired events: %v", err)
				return err
			}

			// skip if no events to delete
			if len(expiredEvents) == 0 {
				logger.Info("no expired events to delete")
				return nil
			}

			// delete expired events
			err = o.storage.DeleteEventsByIDs(ctx, expiredEvents.IDs())
			if err != nil {
				logger.Errorf("can't delete expired events: %v", err)
				return err
			}

			return nil
		})
		if err != nil {
			// TODO: log sentry
			logger.Errorf(
				"delete events transaction aborted: %v",
				trace.Wrapf(span, err, "TransactionalOutbox.GarbageCollector.Tick.txManager.Do()"),
			)
		}

		// cool down between deletions to avoid overloading the database
		time.Sleep(o.cfg.GarbageCollector.Cooldown * time.Millisecond)
	}
}
