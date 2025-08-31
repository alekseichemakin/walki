package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"walki/internal/models"
)

type RouteRepo struct{ db *pgxpool.Pool }

func NewRouteRepo(db *pgxpool.Pool) *RouteRepo { return &RouteRepo{db: db} }

func (r *RouteRepo) Cities(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT DISTINCT city FROM route_versions ORDER BY city`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *RouteRepo) ByCity(ctx context.Context, city string) ([]models.RouteVersion, error) {
	const q = `
	SELECT id,
		   route_id,
		   version_number,
		   title,
		   description,
		   duration_minutes,
		   length_km,
		   theme,
		   price,
		   city,
		   created_at
	FROM route_versions
	WHERE city = $1
	ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, q, city)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.RouteVersion
	for rows.Next() {
		var r models.RouteVersion
		if err := rows.Scan(
			&r.ID, &r.RouteID, &r.VersionNumber, &r.Title, &r.Description,
			&r.DurationMinutes, &r.LengthKm, &r.Theme, &r.Price, &r.City, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (r *RouteRepo) Details(ctx context.Context, routeID int) (*models.RouteVersion, error) {
	const q = `
	SELECT rv.id, rv.route_id, rv.version_number, rv.title, rv.description,
	       COALESCE(rv.duration_minutes,0), COALESCE(rv.length_km,0),
	       COALESCE(rv.theme,''), COALESCE(rv.price,0), rv.city,
	       rv.created_at,
	       COALESCE(m.url, '') AS cover_image_url
	FROM route_versions rv
	LEFT JOIN route_version_media rvm ON rvm.route_version_id = rv.id
	LEFT JOIN media m ON m.id = rvm.media_id AND m.type = 'image'
	WHERE rv.route_id = $1
	ORDER BY rv.version_number DESC
	LIMIT 1;
	`
	var v models.RouteVersion
	if err := r.db.QueryRow(ctx, q, routeID).Scan(
		&v.ID, &v.RouteID, &v.VersionNumber, &v.Title, &v.Description,
		&v.DurationMinutes, &v.LengthKm, &v.Theme, &v.Price, &v.City,
		&v.CreatedAt, &v.CoverImageURL,
	); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *RouteRepo) VersionByID(ctx context.Context, versionID int) (*models.RouteVersion, error) {
	const q = `
	SELECT rv.id, rv.route_id, rv.version_number, rv.title, rv.description,
	       COALESCE(rv.duration_minutes,0), COALESCE(rv.length_km,0),
	       COALESCE(rv.theme,''), COALESCE(rv.price,0), rv.city,
	       rv.created_at,
	       COALESCE(m.url, '') AS cover_image_url
	FROM route_versions rv
	LEFT JOIN route_version_media rvm ON rvm.route_version_id = rv.id
	LEFT JOIN media m ON m.id = rvm.media_id AND m.type = 'image'
	WHERE rv.id = $1;
	`
	var v models.RouteVersion
	if err := r.db.QueryRow(ctx, q, versionID).Scan(
		&v.ID, &v.RouteID, &v.VersionNumber, &v.Title, &v.Description,
		&v.DurationMinutes, &v.LengthKm, &v.Theme, &v.Price, &v.City,
		&v.CreatedAt, &v.CoverImageURL,
	); err != nil {
		return nil, err
	}
	return &v, nil
}
