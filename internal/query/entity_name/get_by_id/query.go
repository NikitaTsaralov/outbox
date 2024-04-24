package get_by_id

import "github.com/google/uuid"

type Query struct {
	id uuid.UUID
}

func NewQuery(id uuid.UUID) *Query {
	return &Query{id: id}
}
