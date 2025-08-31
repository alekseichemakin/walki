package postgres

import (
	"context"
	"time"
	"walki/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct{ db *pgxpool.Pool }

func NewOrderRepo(db *pgxpool.Pool) *OrderRepo { return &OrderRepo{db: db} }

func (r *OrderRepo) Create(ctx context.Context, userID, routeID int, amount float64) (int, *time.Time, error) {
	// берём актуальную версию маршрута
	const qVersion = `
	  SELECT id
	  FROM route_versions
	  WHERE route_id = $1
	  ORDER BY version_number DESC
	  LIMIT 1;
	`
	var versionID int
	if err := r.db.QueryRow(ctx, qVersion, routeID).Scan(&versionID); err != nil {
		return 0, nil, err
	}

	// создаём заказ как "paid" с доступом 30 дней (если хочешь — поставь NULL = бессрочно)
	expiry := time.Now().Add(30 * 24 * time.Hour)
	const qInsert = `
	  INSERT INTO orders (user_id, version_id, route_id, status, amount, paid_at, access_expiry)
	  VALUES ($1,$2,$3,'paid',$4,NOW(),$5)
	  RETURNING access_expiry;
	`
	if err := r.db.QueryRow(ctx, qInsert, userID, versionID, routeID, amount, expiry).Scan(&expiry); err != nil {
		return 0, nil, err
	}
	return versionID, &expiry, nil
}

func (r *OrderRepo) ListByUser(ctx context.Context, userID int) ([]domain.OrderSummary, error) {
	const q = `
	  SELECT o.route_id, rv.title, rv.city, o.version_id, o.access_expiry
	  FROM orders o
	  JOIN route_versions rv ON rv.id = o.version_id
	  WHERE o.user_id = $1 AND o.status = 'paid'
	  ORDER BY o.created_at DESC;
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.OrderSummary
	for rows.Next() {
		var s domain.OrderSummary
		if err := rows.Scan(&s.RouteID, &s.RouteTitle, &s.RouteCity, &s.VersionID, &s.AccessExpiry); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *OrderRepo) UserHasAccess(ctx context.Context, userID, routeID int) (bool, error) {
	const q = `
	  SELECT EXISTS(
	    SELECT 1 FROM orders
	    WHERE user_id=$1 AND route_id=$2 AND status='paid'
	      AND (access_expiry IS NULL OR access_expiry >= NOW())
	  );
	`
	var ok bool
	if err := r.db.QueryRow(ctx, q, userID, routeID).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}
