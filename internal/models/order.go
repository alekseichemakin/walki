package models

import "time"

type Order struct {
	ID           int
	UserID       int
	RouteID      int
	VersionID    int
	Status       string
	Amount       float64
	CreatedAt    time.Time
	PaidAt       *time.Time
	AccessExpiry *time.Time
}

type UserOrder struct {
	Order
	RouteTitle    string
	RouteCity     string
	RouteLength   float64
	RouteDuration int
}
