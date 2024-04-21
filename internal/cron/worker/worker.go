package worker

import (
	"context"
	"time"

	"github.com/NikitaTsaralov/transactional-outbox/config"
	"github.com/NikitaTsaralov/transactional-outbox/internal/cron"
	"github.com/NikitaTsaralov/utils/logger"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"go.opentelemetry.io/otel"
)

type Worker struct {
	cfg       *config.MessageRelay
	storage   cron.OutboxStorageInterface
	metrics   cron.MetricsInterface
	publisher cron.OutboxPublisherInterface
	exit      chan struct{}
	ticker    *time.Ticker
	txManager *manager.Manager
}

func NewWorker(
	cfg *config.MessageRelay,
	storage cron.OutboxStorageInterface,
	metrics cron.MetricsInterface,
	publisher cron.OutboxPublisherInterface,
	txManager *manager.Manager,
) *Worker {
	return &Worker{
		cfg:       cfg,
		storage:   storage,
		metrics:   metrics,
		publisher: publisher,
		exit:      make(chan struct{}),
		ticker:    time.NewTicker(cfg.Timeout * time.Millisecond),
		txManager: txManager,
	}
}

func (w *Worker) Start() {
	if w.cfg.Enabled {
		go func() {
			for {
				select {
				case <-w.exit:
					close(w.exit)
					return
				case <-w.ticker.C:
					w.Tick(context.Background())
				}
			}
		}()
	}
}

func (w *Worker) Stop() {
	w.ticker.Stop()
	w.exit <- struct{}{}
	<-w.exit
}

func (w *Worker) Tick(ctx context.Context) {
	ctx, span := otel.Tracer("").Start(ctx, "Worker.Tick")
	defer span.End()

	err := w.txManager.Do(ctx, func(ctx context.Context) error {
		events, err := w.storage.FetchUnprocessedEvents(ctx)
		if err != nil {
			return err
		}

		if len(events) == 0 {
			return nil
		}

		err = w.publisher.PublishEvents(ctx, events)
		if err != nil {
			return err
		}

		err = w.storage.MarkEventsAsProcessed(ctx, events.IDs())
		if err != nil {
			return err
		}

		w.metrics.DecUnprocessedEventsCounter(len(events))

		return nil
	})
	if err != nil {
		logger.Error(err) // TODO: sentry or telegram alert
	}
}
