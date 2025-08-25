package service

import (
	"context"
	"walki/internal/models"
	"walki/internal/repository"
)

type RouteService struct {
	repo repository.RouteRepository
}

func NewRouteService(repo repository.RouteRepository) *RouteService {
	return &RouteService{repo: repo}
}

func (s *RouteService) Cities(ctx context.Context) ([]string, error) {
	return s.repo.Cities(ctx)
}

func (s *RouteService) ListByCity(ctx context.Context, city string) ([]models.RouteVersion, error) {
	return s.repo.ByCity(ctx, city)
}

func (s *RouteService) Details(ctx context.Context, routeID int) (*models.RouteVersion, error) {
	return s.repo.Details(ctx, routeID)
}

func (s *RouteService) VersionByID(ctx context.Context, routeVersionID int) (*models.RouteVersion, error) {
	return s.repo.VersionByID(ctx, routeVersionID)
}
