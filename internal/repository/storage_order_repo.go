package repository

import (
	"context"
	"time"
	"walki/internal/domain"
	"walki/internal/storage"
)

type StorageOrderRepo struct{ s *storage.Storage }

func NewStorageOrderRepo(s *storage.Storage) OrderRepository { return &StorageOrderRepo{s: s} }

func (r *StorageOrderRepo) Create(ctx context.Context, userID, routeID int, amount float64) (int, *time.Time, error) {
	// предполагаем ту же сигнатуру и семантику внутри storage
	order, err := r.s.CreateOrder(userID, routeID, amount)
	if err != nil {
		return 0, nil, err
	}
	return order.VersionID, order.AccessExpiry, nil
}

func (r *StorageOrderRepo) ListByUser(ctx context.Context, userID int) ([]domain.OrderSummary, error) {
	orders, err := r.s.GetUserOrders(userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.OrderSummary, 0, len(orders))
	for _, o := range orders {
		out = append(out, domain.OrderSummary{
			RouteID:      o.RouteID,
			RouteTitle:   o.RouteTitle,
			RouteCity:    o.RouteCity,
			VersionID:    o.VersionID,
			AccessExpiry: o.AccessExpiry,
		})
	}
	return out, nil
}

func (r *StorageOrderRepo) UserHasAccess(ctx context.Context, userID, routeID int) (bool, error) {
	return r.s.HasAccessToRoute(userID, routeID)
}
