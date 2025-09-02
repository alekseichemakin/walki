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

type RouteRunRepository interface {
	FirstPoint(ctx context.Context, versionID int) (*models.RoutePoint, error)
	PointByIndex(ctx context.Context, versionID, idx int) (*models.RoutePoint, error)
	NextIndex(ctx context.Context, versionID, after int) (int, bool, error)
	PrevIndex(ctx context.Context, versionID, before int) (int, bool, error)
	PointMediaIDs(ctx context.Context, pointID int) (photoIDs []int64, audioIDs []int64, err error)
	UpsertProgress(ctx context.Context, userID, routeID, versionID int, idx int) error
	GetProgress(ctx context.Context, userID, versionID int) (*models.RouteProgress, error)
	Finish(ctx context.Context, userID, versionID int) error
	UpdateMessageIDs(ctx context.Context, userID, versionID int, contentMsgID, voiceMsgID *int) error
}

type MediaRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Media, error)

	UpdateStorage(ctx context.Context, mediaID int64, bucket, key, mimeType string, sizeBytes int64) error

	GetTelegramFileID(ctx context.Context, mediaID int64) (fileID string, ok bool, err error)
	UpsertTelegramFileID(ctx context.Context, mediaID int64, fileID, contentType string, chatID *int64) error
}
