package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"walki/internal/models"
)

type RouteRunRepo struct{ db *pgxpool.Pool }

func NewRouteRunRepo(db *pgxpool.Pool) *RouteRunRepo { return &RouteRunRepo{db: db} }

func (r *RouteRunRepo) FirstPoint(ctx context.Context, versionID int) (*models.RoutePoint, error) {
	const q = `SELECT id, version_id, order_index, title, description, latitude, longitude, created_at
               FROM route_points WHERE version_id=$1 ORDER BY order_index ASC LIMIT 1`
	var p models.RoutePoint
	if err := r.db.QueryRow(ctx, q, versionID).
		Scan(&p.ID, &p.VersionID, &p.Idx, &p.Title, &p.Description, &p.Lat, &p.Lon, &p.CreatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *RouteRunRepo) PointByIndex(ctx context.Context, versionID, idx int) (*models.RoutePoint, error) {
	const q = `SELECT id, version_id, order_index, title, description, latitude, longitude, created_at
               FROM route_points WHERE version_id=$1 AND order_index=$2`
	var p models.RoutePoint
	if err := r.db.QueryRow(ctx, q, versionID, idx).
		Scan(&p.ID, &p.VersionID, &p.Idx, &p.Title, &p.Description, &p.Lat, &p.Lon, &p.CreatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *RouteRunRepo) NextIndex(ctx context.Context, versionID, after int) (int, bool, error) {
	const q = `SELECT order_index FROM route_points WHERE version_id=$1 AND order_index>$2 ORDER BY order_index ASC LIMIT 1`
	var idx int
	if err := r.db.QueryRow(ctx, q, versionID, after).Scan(&idx); err != nil {
		// нет следующего
		return 0, false, nil
	}
	return idx, true, nil
}

func (r *RouteRunRepo) PrevIndex(ctx context.Context, versionID, before int) (int, bool, error) {
	const q = `SELECT order_index FROM route_points WHERE version_id=$1 AND order_index<$2 ORDER BY order_index DESC LIMIT 1`
	var idx int
	if err := r.db.QueryRow(ctx, q, versionID, before).Scan(&idx); err != nil {
		return 0, false, nil
	}
	return idx, true, nil
}

func (r *RouteRunRepo) PointMedia(ctx context.Context, pointID int) (photos []string, audios []string, err error) {
	const q = `
      SELECT m.type, m.url
      FROM route_point_media rpm
      JOIN media m ON m.id = rpm.media_id
      WHERE rpm.route_point_id=$1`
	rows, err := r.db.Query(ctx, q, pointID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var typ, url string
		if err := rows.Scan(&typ, &url); err != nil {
			return nil, nil, err
		}
		if typ == "image" {
			photos = append(photos, url)
		} else if typ == "audio" {
			audios = append(audios, url)
		}
	}
	return photos, audios, rows.Err()
}

func (r *RouteRunRepo) UpsertProgress(ctx context.Context, userID, routeID, versionID int, idx int) error {
	const q = `
    INSERT INTO route_progress (user_id, route_id, version_id, current_idx)
    VALUES ($1,$2,$3,$4)
    ON CONFLICT (user_id, version_id) DO UPDATE
      SET current_idx = EXCLUDED.current_idx,
          finished_at = NULL`
	_, err := r.db.Exec(ctx, q, userID, routeID, versionID, idx)
	return err
}

func (r *RouteRunRepo) GetProgress(ctx context.Context, userID, versionID int) (*models.RouteProgress, error) {
	const q = `SELECT id, user_id, route_id, version_id, current_idx, started_at, finished_at,
	                  content_msg_id, voice_msg_id
	           FROM route_progress
	           WHERE user_id=$1 AND version_id=$2`

	var p models.RouteProgress
	if err := r.db.QueryRow(ctx, q, userID, versionID).Scan(
		&p.ID, &p.UserID, &p.RouteID, &p.VersionID, &p.CurrentIdx, &p.StartedAt, &p.FinishedAt,
		&p.ContentMsgID, &p.VoiceMsgID,
	); err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *RouteRunRepo) Finish(ctx context.Context, userID, versionID int) error {
	const q = `UPDATE route_progress SET finished_at=NOW() WHERE user_id=$1 AND version_id=$2`
	ct, err := r.db.Exec(ctx, q, userID, versionID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("progress not found")
	}
	return nil
}

// postgres/route_run_repo.go
func (r *RouteRunRepo) UpdateMessageIDs(
	ctx context.Context,
	userID, versionID int,
	contentMsgID, voiceMsgID *int,
) error {
	set := make([]string, 0, 2)
	args := make([]any, 0, 4)
	i := 1

	if contentMsgID != nil {
		// 0 трактуем как «очистить поле» -> NULL (если ты так договорился с хендлером)
		set = append(set, fmt.Sprintf("content_msg_id = NULLIF($%d, 0)", i))
		args = append(args, *contentMsgID)
		i++
	}
	if voiceMsgID != nil {
		set = append(set, fmt.Sprintf("voice_msg_id = NULLIF($%d, 0)", i))
		args = append(args, *voiceMsgID)
		i++
	}

	if len(set) == 0 {
		// нечего обновлять — это не ошибка
		return nil
	}

	q := fmt.Sprintf(
		`UPDATE route_progress
           SET %s
         WHERE user_id = $%d AND version_id = $%d`,
		strings.Join(set, ", "),
		i, i+1,
	)
	args = append(args, userID, versionID)

	ct, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("progress not found")
	}
	return nil
}
