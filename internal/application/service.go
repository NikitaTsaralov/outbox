package application

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/config"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event/batch_create"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event/create"
	"github.com/NikitaTsaralov/transactional-outbox/internal/cron/worker"
	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/metrics"
	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/publisher/kafka"
	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/storage/postgres"
	trmsqlx "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
)

type Service struct {
	worker  *worker.Worker
	command *command.EventCommand
}

func NewService(
	cfg *config.Config,
	registry *prometheus.Registry,
	dbImpl *sqlx.DB,
	kafkaImpl *kgo.Client,
	ctxGetter *trmsqlx.CtxGetter,
	txManager *manager.Manager,
) *Service {
	outboxStorage := postgres.NewStorage(cfg, dbImpl, ctxGetter)
	outboxPublisher := kafka.NewPublisher(kafkaImpl)
	outboxMetrics := metrics.NewMetrics(cfg.Metrics, registry)

	return &Service{
		worker: worker.NewWorker(
			cfg.MessageRelay,
			outboxStorage,
			outboxMetrics,
			outboxPublisher,
			// here we use our own transaction manager to hide worker inner logic
			manager.Must(trmsqlx.NewDefaultFactory(dbImpl)),
		),
		// while here we use manager from args,
		// that's because command needs to be wrapped in transaction from outside the package
		command: command.NewOutboxCommand(outboxStorage, outboxMetrics, txManager),
	}
}

func (s *Service) CreateEvent(ctx context.Context, command *create.Command) error {
	ctx, span := otel.Tracer("").Start(ctx, "Service.CreateEvent")
	defer span.End()

	err := s.command.Create.Execute(ctx, command)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) BatchCreateEvent(ctx context.Context, command *batch_create.Command) error {
	ctx, span := otel.Tracer("").Start(ctx, "Service.BatchCreateEvent")
	defer span.End()

	err := s.command.BatchCreate.Execute(ctx, command)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Start() {
	s.worker.Start()
}

func (s *Service) Stop() {
	s.worker.Stop()
}
