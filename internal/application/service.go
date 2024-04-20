package application

import (
	"context"

	"github.com/NikitaTsaralov/transactional-outbox/config"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event/batch_create"
	"github.com/NikitaTsaralov/transactional-outbox/internal/command/event/create"
	"github.com/NikitaTsaralov/transactional-outbox/internal/cron/worker"
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
	cfg *config.MessageRelay,
	registry *prometheus.Registry,
	dbImpl *sqlx.DB,
	kafkaImpl *kgo.Client,
) *Service {
	outboxStorage := postgres.NewStorage(dbImpl, trmsqlx.DefaultCtxGetter)
	outboxPublisher := kafka.NewPublisher(kafkaImpl)

	return &Service{
		worker:  worker.NewWorker(cfg, outboxStorage, outboxPublisher, manager.Must(trmsqlx.NewDefaultFactory(dbImpl))),
		command: command.NewOutboxCommand(outboxStorage, manager.Must(trmsqlx.NewDefaultFactory(dbImpl))),
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
