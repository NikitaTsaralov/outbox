package entity

import "github.com/google/uuid"

type OrderOption func(entity *Entity)

func WithID(id uuid.UUID) OrderOption {
	return func(entity *Entity) {
		entity.id = id
	}
}

func WithValue(value string) OrderOption {
	return func(entity *Entity) {
		entity.value = value
	}
}
