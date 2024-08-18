package outbox

import (
	"context"
	"testing"

	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

const (
	timeout   = 1000 // 1s
	eventTTL  = 1000 // 1s
	batchSize = 10
)

type OutboxTestSuite struct {
	suite.Suite
	cfg            *Config
	db             *sqlx.DB
	broker         *kgo.Client
	jaegerExporter *jaeger.Exporter
	traceProvider  *tracesdk.TracerProvider
	txManager      *manager.Manager
	ctxGetter      *txManager.CtxGetter
	outbox         *Outbox
}

func initPostgres() (*sqlx.DB, error) {
	return sqlx.Connect(
		"pgx",
		"host=localhost port=5432 user=root dbname=postgres sslmode=disable password=dev",
	)
}

func initKafka() (*kgo.Client, error) {
	return kgo.NewClient(kgo.SeedBrokers([]string{"localhost:29092"}...))
}

func initJaeger() (*jaeger.Exporter, error) {
	return jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint("localhost:16686"),
		),
	)
}

func initTraceProvider(exp *jaeger.Exporter) *tracesdk.TracerProvider {
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("OutboxTest"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	)

	return tp
}

func (s *OutboxTestSuite) SetupTest() {
	var err error

	s.cfg = &Config{
		MessageRelay: MessageRelayConfig{
			Timeout:   timeout, // 1s
			BatchSize: batchSize,
		},
		GarbageCollector: GarbageCollectorConfig{
			Timeout:   timeout, // 1s
			BatchSize: batchSize,
		},
	}

	s.db, err = initPostgres()
	s.Require().Nil(err)
	s.Require().NotNil(s.db)

	s.broker, err = initKafka()
	s.Require().Nil(err)
	s.Require().NotNil(s.broker)

	s.jaegerExporter, err = initJaeger()
	s.Require().Nil(err)
	s.Require().NotNil(s.jaegerExporter)

	s.traceProvider = initTraceProvider(s.jaegerExporter)

	s.txManager = manager.Must(txManager.NewDefaultFactory(s.db))
	s.ctxGetter = txManager.DefaultCtxGetter

	s.outbox = NewOutbox(s.cfg, s.db, s.broker, s.txManager, s.ctxGetter)

	_, err = s.db.Exec(`delete from outbox;`)
	s.Require().Nil(err)
}

func (s *OutboxTestSuite) TearDownSuite() {
	err := s.db.Close()
	s.Require().Nil(err)

	s.broker.Close()

	err = s.traceProvider.Shutdown(context.Background())
	s.Require().Nil(err)

	err = s.jaegerExporter.Shutdown(context.Background())
	s.Require().Nil(err)
}

func TestOutboxTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(OutboxTestSuite))
}
