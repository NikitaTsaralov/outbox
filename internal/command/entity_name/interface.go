package entity_name

import (
	"context"

	"layout-example/internal/domain/entity"
)

type StorageInterface interface {
	Create(ctx context.Context, entity *entity.Entity) (int64, error)
	Update(ctx context.Context, entity *entity.Entity) (err error)
}
