package service

import (
	"context"
	"fmt"
	"time"
	"walki/internal/repository"
)

type OrderService struct {
	orders repository.OrderRepository
	routes repository.RouteRepository
}

func NewOrderService(o repository.OrderRepository, r repository.RouteRepository) *OrderService {
	return &OrderService{orders: o, routes: r}
}

type PurchaseResult struct {
	VersionID    int
	AccessExpiry *time.Time
	Amount       float64
}

func (s *OrderService) Purchase(ctx context.Context, userID, routeID int) (*PurchaseResult, error) {
	route, err := s.routes.Details(ctx, routeID)
	if err != nil {
		return nil, fmt.Errorf("get route details: %w", err)
	}
	verID, expiry, err := s.orders.Create(ctx, userID, routeID, route.Price)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return &PurchaseResult{VersionID: verID, AccessExpiry: expiry, Amount: route.Price}, nil
}
