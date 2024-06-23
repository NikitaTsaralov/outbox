package outbox

import (
	"github.com/NikitaTsaralov/transactional-outbox/config"
	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/broker/kafka"
	"github.com/NikitaTsaralov/transactional-outbox/internal/infrastructure/storage/postgres"
	interfaces2 "github.com/NikitaTsaralov/transactional-outbox/internal/interfaces"
	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/jmoiron/sqlx"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Client struct {
	cfg       *config.Config
	storage   interfaces2.OutboxStorage
	broker    interfaces2.Broker
	txManager *manager.Manager
	ctxGetter *txManager.CtxGetter
}

func NewClient(
	cfg *config.Config,
	db *sqlx.DB,
	kafkaClient *kgo.Client,
	txManager *manager.Manager,
	ctxGetter *txManager.CtxGetter,
) *Client {
	return &Client{
		cfg:       cfg,
		storage:   postgres.NewStorage(db, ctxGetter),
		broker:    kafka.NewBroker(kafkaClient),
		txManager: txManager,
		ctxGetter: ctxGetter,
	}
}
