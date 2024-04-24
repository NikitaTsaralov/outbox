package service

import (
	"layout-example/internal/command"
	"layout-example/internal/query"
)

type OrderService struct {
	Commands *command.EntityCommand
	Queries  *query.EntityQueries
}

func NewOrderService(commands *command.EntityCommand, queries *query.EntityQueries) *OrderService {
	return &OrderService{Commands: commands, Queries: queries}
}
