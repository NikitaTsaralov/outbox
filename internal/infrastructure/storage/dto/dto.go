package dto

import (
	"layout-example/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type Entity struct {
	ID    uuid.UUID `db:"id"`
	Value string    `db:"value"` // it's just an example
}

func (e *Entity) ToDomain() *entity.Entity {
	return entity.NewEntity(entity.WithID(e.ID), entity.WithValue(e.Value))
}

func NewEntityFromDomain(entity *entity.Entity) *Entity {
	return &Entity{
		ID:    entity.ID(),
		Value: entity.Value(),
	}
}

type Entities []*Entity

func (e Entities) ToDomain() []*entity.Entity {
	return lo.Map(e, func(entity *Entity, _ int) *entity.Entity {
		return entity.ToDomain()
	})
}

func NewEntitiesFromDomain(entities entity.Entities) Entities {
	return lo.Map(entities, func(entity *entity.Entity, _ int) *Entity {
		return NewEntityFromDomain(entity)
	})
}
