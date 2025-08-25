package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
	"walki/internal/models"
)

func (s *Storage) GetRoutesByCity(city string) ([]models.RouteVersion, error) {
	query := s.queries["get_routes_by_city.sql"]
	rows, err := s.db.Query(context.Background(), query, city)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}
	defer rows.Close()

	var routes []models.RouteVersion
	for rows.Next() {
		var r models.RouteVersion
		if err := rows.Scan(
			&r.ID, &r.RouteID, &r.VersionNumber, &r.Title, &r.Description,
			&r.DurationMinutes, &r.LengthKm, &r.Theme, &r.Price, &r.City, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		routes = append(routes, r)
	}

	return routes, nil
}

func (s *Storage) GetRouteDetails(routeID int) (*models.Route, error) {
	query := s.queries["get_route_details.sql"]
	var route models.Route
	var version models.RouteVersion

	err := s.db.QueryRow(context.Background(), query, routeID).Scan(
		&route.ID, &route.Status, &route.IsVisible, &route.CreatedBy,
		&route.CreatedAt, &route.UpdatedAt,
		&version.ID, &version.VersionNumber, &version.Title, &version.Description,
		&version.DurationMinutes, &version.LengthKm, &version.Theme,
		&version.Price, &version.City, &version.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get route details: %w", err)
	}

	// Получаем обложку для версии маршрута
	version.CoverImageURL = s.getRouteCoverImage(version.ID)
	route.Versions = []models.RouteVersion{version}

	return &route, nil
}

// getRouteCoverImage возвращает URL обложки для версии маршрута
func (s *Storage) getRouteCoverImage(routeVersionID int) string {
	var coverImageURL string
	query := s.queries["get_route_cover_image.sql"]

	err := s.db.QueryRow(context.Background(), query, routeVersionID).Scan(&coverImageURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Error getting cover image for route version %d: %v", routeVersionID, err)
		}
		return ""
	}

	return coverImageURL
}

func (s *Storage) GetRouteVersionByVersionID(versionID int) (*models.RouteVersion, error) {
	query := s.queries["get_route_version_by_id.sql"]
	var version models.RouteVersion
	var createdAt time.Time

	err := s.db.QueryRow(context.Background(), query, versionID).Scan(
		&version.ID,
		&version.RouteID,
		&version.VersionNumber,
		&version.Title,
		&version.Description,
		&version.DurationMinutes,
		&version.LengthKm,
		&version.Theme,
		&version.Price,
		&version.City,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("route version with ID %d not found", versionID)
		}
		return nil, fmt.Errorf("failed to get route version: %w", err)
	}

	version.CreatedAt = createdAt
	version.CoverImageURL = s.getRouteCoverImage(versionID)
	return &version, nil
}
