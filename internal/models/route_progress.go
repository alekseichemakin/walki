package models

import "time"

type RouteProgress struct {
	ID             int
	UserID         int
	OrderID        int
	CurrentPointID int
	Status         string
	StartedAt      time.Time
	CompletedAt    time.Time
}
