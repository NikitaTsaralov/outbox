package valueobject

import (
	"context"
	"encoding/json"

	"github.com/NikitaTsaralov/utils/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type ID int64

type IDs []ID

func ExtractTraceCarrierToJSON(ctx context.Context) json.RawMessage {
	carrier := make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	traceCarrier, err := json.Marshal(carrier)
	if err != nil {
		logger.Warnf("could not marshal trace carrier, insert only trace id, %v", err.Error())
	}

	return traceCarrier
}
