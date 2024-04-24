package fetch

import (
	"context"

	"layout-example/internal/domain/entity"
	"layout-example/internal/query/entity_name"

	"go.opentelemetry.io/otel"
)

type QueryHandler struct {
	storage entity_name.StorageInterface
}

func NewQueryHandler(storage entity_name.StorageInterface) *QueryHandler {
	return &QueryHandler{
		storage: storage,
	}
}

func (h *QueryHandler) Execute(
	ctx context.Context,
	query *Query,
) (entity.Entities, error) {
	ctx, span := otel.Tracer("").Start(ctx, "QueryHandler.Execute")
	defer span.End()

	return h.storage.Fetch(ctx)
}
