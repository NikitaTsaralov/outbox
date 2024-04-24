package entity_name

import (
	"context"

	"layout-example/internal/domain/entity"

	"github.com/google/uuid"
)

type StorageInterface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Entity, error)
	Fetch(ctx context.Context) (entity.Entities, error)
}
