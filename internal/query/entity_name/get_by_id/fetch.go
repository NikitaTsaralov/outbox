package get_by_id

import (
	"context"

	"cqrs-layout-example/internal/domain/entity"
	"cqrs-layout-example/internal/query/entity_name"

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
) (*entity.Entity, error) {
	ctx, span := otel.Tracer("").Start(ctx, "QueryHandler.Execute")
	defer span.End()

	return h.storage.GetByID(ctx, query.id)
}
