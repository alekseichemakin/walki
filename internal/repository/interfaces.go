package repository

import (
	"context"
	"time"
	"walki/internal/domain"
	"walki/internal/models"
)

type RouteRepository interface {
	Cities(ctx context.Context) ([]string, error)
	ByCity(ctx context.Context, city string) ([]models.RouteVersion, error)
	Details(ctx context.Context, routeID int) (*models.RouteVersion, error)
	VersionByID(ctx context.Context, versionID int) (*models.RouteVersion, error)
}

type OrderRepository interface {
	Create(ctx context.Context, userID, routeID int, amount float64) (versionID int, accessExpiry *time.Time, err error)
	ListByUser(ctx context.Context, userID int) ([]domain.OrderSummary, error)
	UserHasAccess(ctx context.Context, userID, routeID int) (bool, error)
}

type UserRepository interface {
	Upsert(ctx context.Context, u *models.User) error
	ByTelegramID(ctx context.Context, tgID int64) (*models.User, error)
}
