package entity

import "github.com/google/uuid"

type Entity struct {
	id    uuid.UUID
	value string
}

func (e *Entity) ID() uuid.UUID {
	return e.id
}

func (e *Entity) Value() string {
	return e.value
}

func NewEntity(opts ...OrderOption) *Entity {
	entity := &Entity{}

	for _, opt := range opts {
		opt(entity)
	}

	return entity

}

type Entities []*Entity
