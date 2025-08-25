package storage

import (
	"context"
	"fmt"
	"time"
	"walki/internal/models"
)

// CreateOrder создает новый заказ (с заглушкой оплаты)
func (s *Storage) CreateOrder(userID, routeID int, amount float64) (*models.Order, error) {
	// Получаем актуальную версию маршрута
	var versionID int
	err := s.db.QueryRow(context.Background(), s.queries["get_latest_route_version.sql"], routeID).Scan(&versionID)

	if err != nil {
		return nil, fmt.Errorf("failed to get route version: %w", err)
	}

	// Создаем заказ с статусом "paid" (заглушка оплаты)
	now := time.Now()
	accessExpiry := now.Add(30 * 24 * time.Hour) // Доступ на 30 дней

	var order models.Order
	err = s.db.QueryRow(context.Background(), s.queries["create_order.sql"],
		userID, routeID, versionID, amount, now, now, accessExpiry).Scan(
		&order.ID, &order.UserID, &order.RouteID, &order.VersionID, &order.Status,
		&order.Amount, &order.CreatedAt, &order.PaidAt, &order.AccessExpiry,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &order, nil
}

// GetUserOrders возвращает список заказов пользователя
func (s *Storage) GetUserOrders(userId int) ([]models.UserOrder, error) {
	query := s.queries["get_user_orders.sql"]
	rows, err := s.db.Query(context.Background(), query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()

	var orders []models.UserOrder
	for rows.Next() {
		var o models.UserOrder
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.RouteID, &o.VersionID, &o.Status, &o.Amount,
			&o.CreatedAt, &o.PaidAt, &o.AccessExpiry,
			&o.RouteTitle, &o.RouteCity, &o.RouteLength, &o.RouteDuration,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	return orders, nil
}

// HasAccessToRoute проверяет, есть ли у пользователя доступ к маршруту
func (s *Storage) HasAccessToRoute(userID int, routeID int) (bool, error) {
	var hasAccess bool
	err := s.db.QueryRow(context.Background(), s.queries["has_access_to_route.sql"], userID, routeID).Scan(&hasAccess)

	if err != nil {
		return false, fmt.Errorf("failed to check access: %w", err)
	}

	return hasAccess, nil
}
