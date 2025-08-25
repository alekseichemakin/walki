package service

import (
	"context"
	"walki/internal/domain"
	"walki/internal/models"
	"walki/internal/repository"
)

type ProfileService struct {
	orders repository.OrderRepository
	routes repository.RouteRepository
}

func NewProfileService(o repository.OrderRepository, r repository.RouteRepository) *ProfileService {
	return &ProfileService{orders: o, routes: r}
}

func (s *ProfileService) MyOrders(ctx context.Context, userID int) ([]domain.OrderSummary, error) {
	return s.orders.ListByUser(ctx, userID)
}
func (s *ProfileService) HasAccess(ctx context.Context, userID, routeID int) (bool, error) {
	return s.orders.UserHasAccess(ctx, userID, routeID)
}
func (s *ProfileService) VersionByID(ctx context.Context, versionID int) (*models.RouteVersion, error) {
	return s.routes.VersionByID(ctx, versionID)
}
