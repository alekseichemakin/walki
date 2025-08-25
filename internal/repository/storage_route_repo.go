package repository

import (
	"context"
	"database/sql"
	"walki/internal/models"
	"walki/internal/storage"
)

// StorageRouteRepo — тонкий адаптер: текущий storage => интерфейс RouteRepository.
// Это позволяет не ломать существующий wiring и постепенно выносить SQL в отдельные реализации.
type StorageRouteRepo struct {
	s *storage.Storage
}

func NewStorageRouteRepo(s *storage.Storage) RouteRepository {
	return &StorageRouteRepo{s: s}
}

func (r *StorageRouteRepo) Cities(ctx context.Context) ([]string, error) {
	// Текущий storage не принимает ctx; на первом шаге берём как есть.
	return r.s.GetCities()
}

func (r *StorageRouteRepo) ByCity(ctx context.Context, city string) ([]models.RouteVersion, error) {
	return r.s.GetRoutesByCity(city)
}

func (r *StorageRouteRepo) Details(ctx context.Context, routeID int) (*models.RouteVersion, error) {
	route, err := r.s.GetRouteDetails(routeID)
	if err != nil {
		return nil, err
	}
	if len(route.Versions) == 0 {
		return nil, sql.ErrNoRows
	}
	// Возвращаем актуальную версию (первая в срезе по текущей логике)
	v := route.Versions[0]
	return &v, nil
}

func (r *StorageRouteRepo) VersionByID(ctx context.Context, versionID int) (*models.RouteVersion, error) {
	v, err := r.s.GetRouteVersionByVersionID(versionID)
	if err != nil {
		return nil, err
	}
	return v, nil
}
