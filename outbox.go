package outbox

import (
	"github.com/NikitaTsaralov/transactional-outbox/internal/broker"
	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/broker/kafka"
	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/repository/postgres"
	"github.com/NikitaTsaralov/transactional-outbox/internal/repository"
	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/jmoiron/sqlx"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Outbox struct {
	cfg       *Config
	storage   repository.OutboxStorage
	broker    broker.Broker
	txManager *manager.Manager
	ctxGetter *txManager.CtxGetter
}

func NewOutbox(
	cfg *Config,
	db *sqlx.DB,
	kafkaClient *kgo.Client,
	txManager *manager.Manager,
	ctxGetter *txManager.CtxGetter,
) *Outbox {
	return &Outbox{
		cfg:       cfg,
		storage:   postgres.NewStorage(db, ctxGetter),
		broker:    kafka.NewBroker(kafkaClient),
		txManager: txManager,
		ctxGetter: ctxGetter,
	}
}
