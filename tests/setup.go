package tests

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

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
