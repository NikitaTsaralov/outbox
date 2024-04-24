package service

import (
	"cqrs-layout-example/internal/command"
	"cqrs-layout-example/internal/query"
)

type OrderService struct {
	Commands *command.EntityCommand
	Queries  *query.EntityQueries
}

func NewOrderService(commands *command.EntityCommand, queries *query.EntityQueries) *OrderService {
	return &OrderService{Commands: commands, Queries: queries}
}
